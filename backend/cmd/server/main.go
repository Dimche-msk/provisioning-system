package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	"provisioning-system/internal/api"
	"provisioning-system/internal/backup"
	"provisioning-system/internal/broadcaster"
	"provisioning-system/internal/config"
	"provisioning-system/internal/db"
	"provisioning-system/internal/logger" // This is the custom logger package
	"provisioning-system/internal/provisioner"
	"provisioning-system/internal/version"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	// 1. Парсинг аргументов CLI
	configDir := flag.String("config-dir", ".", "Directory containing provisioning-system.yaml and vendors/")
	logLevel := flag.String("log-level", "ERROR", "Log level (DEBUG, INFO, WARN, ERROR)")
	flag.Parse()

	// 2. Инициализация логгера
	logger.SetLevel(*logLevel)

	// 2. Загрузка конфигурации
	cfg, err := config.LoadConfig(*configDir)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	port := ":" + cfg.Server.Port
	fmt.Printf("Starting Provisioning Server version %s on port %s...\n", version.Version, port)
	fmt.Printf("Using config dir: %s\n", *configDir)

	// 3. Инициализация компонентов
	b := broadcaster.New()
	deviceLogger := api.NewDeviceLogger(cfg, b)
	authHandler := api.NewAuthHandler(cfg)

	// 4. Настройка роутинга
	r := mux.NewRouter()

	// Раздача сгенерированных конфигов (если включено)
	if cfg.Server.ServeConfigs {
		configsDir := filepath.Join(*configDir, "temp_configs")
		configFs := http.FileServer(http.Dir(configsDir))

		// Оборачиваем в логгер
		loggingHandler := deviceLogger.Middleware(http.StripPrefix("/config/", configFs))

		r.PathPrefix("/config/").Handler(loggingHandler)
		fmt.Printf("Serving generated configs at http://.../ from %s\n", configsDir)
	}

	// 5. Инициализация Provisioner Manager
	provManager := provisioner.NewManager(cfg)
	if err := provManager.LoadVendors(filepath.Join(*configDir, "vendors")); err != nil {
		log.Printf("Warning: Failed to load vendors: %v", err)
	}
	if err := provManager.LoadModels(); err != nil {
		log.Printf("Warning: Failed to load models: %v", err)
	}

	// 6. Инициализация Database
	database, err := db.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 7. Инициализация Backup Manager
	backupManager := backup.NewManager(cfg, database)

	// 8. Инициализация API Handlers
	sysHandler := api.NewSystemHandler(*configDir, &cfg, provManager, database, backupManager)
	phoneHandler := api.NewPhoneHandler(*configDir, database, provManager)
	debugHandler := api.NewDebugHandler(b)

	// API Routes
	apiRouter := r.PathPrefix("/api").Subrouter()

	// Public API
	apiRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	apiRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	apiRouter.HandleFunc("/check-auth", authHandler.CheckAuth).Methods("GET")

	// Protected API
	// Middleware wrapper
	protected := apiRouter.PathPrefix("/").Subrouter()
	protected.Use(func(next http.Handler) http.Handler {
		return authHandler.Middleware(next)
	})

	protected.HandleFunc("/system/reload", sysHandler.Reload).Methods("POST")
	protected.HandleFunc("/system/apply", sysHandler.ApplyConfig).Methods("POST")
	protected.HandleFunc("/domains", sysHandler.GetDomains).Methods("GET")
	protected.HandleFunc("/deploy", sysHandler.Deploy).Methods("POST")

	protected.HandleFunc("/system/backup", sysHandler.CreateBackup).Methods("POST")
	protected.HandleFunc("/system/backups", sysHandler.ListBackups).Methods("GET")
	protected.HandleFunc("/system/restore", sysHandler.RestoreBackup).Methods("POST")
	protected.HandleFunc("/system/stats", sysHandler.GetSystemStats).Methods("GET")

	protected.HandleFunc("/phones", phoneHandler.CreatePhone).Methods("POST")
	protected.HandleFunc("/phones", phoneHandler.GetPhones).Methods("GET")
	protected.HandleFunc("/phones/{id}", phoneHandler.UpdatePhone).Methods("PUT")
	protected.HandleFunc("/phones/{id}", phoneHandler.DeletePhone).Methods("DELETE")

	protected.HandleFunc("/vendors", phoneHandler.GetVendors).Methods("GET")
	protected.HandleFunc("/models", phoneHandler.GetModels).Methods("GET")

	// Debug API (SSE)
	protected.HandleFunc("/debug/logs", debugHandler.StreamLogs).Methods("GET")

	// Serve Vendor Static Files (Images, etc.)
	vendorsDir := filepath.Join(*configDir, "vendors")
	r.PathPrefix("/api/vendors-static/").Handler(http.StripPrefix("/api/vendors-static/", http.FileServer(http.Dir(vendorsDir))))

	// Статика (Фронтенд) - SPA Fallback
	dist, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	fileServer := http.FileServer(http.FS(dist))

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// API requests should not be handled here
		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// 1. Check if it matches a static file (Frontend)
		f, err := dist.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		if path == "/" {
			fileServer.ServeHTTP(w, r)
			return
		}

		// 2. Universal Config Serving (Fallback for configs)
		if cfg.Server.ServeConfigs {
			configsDir := filepath.Join(*configDir, "temp_configs")

			// 2a. Check default domain (first in config)
			var defaultDomain string
			if len(cfg.Domains) > 0 {
				defaultDomain = cfg.Domains[0].Name
				configPath := filepath.Join(configsDir, defaultDomain, path)
				if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
					deviceLogger.LogCustom(r, http.StatusOK, fmt.Sprintf("Served from default domain: %s", defaultDomain))
					http.ServeFile(w, r, configPath)
					return
				}
			}

			// 2b. Search in all other domains
			entries, err := os.ReadDir(configsDir)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						domain := entry.Name()
						// Skip default domain as we already checked it
						if domain == defaultDomain {
							continue
						}

						configPath := filepath.Join(configsDir, domain, path)
						if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
							deviceLogger.LogCustom(r, http.StatusOK, fmt.Sprintf("Served from domain: %s", domain))
							http.ServeFile(w, r, configPath)
							return
						}
					}
				}
			}
		}

		// 3. Check if it looks like a file (has extension)
		// If it looks like a file and wasn't found above, return 404 and log it
		if filepath.Ext(path) != "" {
			deviceLogger.LogCustom(r, http.StatusNotFound, "File not found")
			http.NotFound(w, r)
			return
		}

		// 4. SPA Fallback
		// If not found anywhere else, serve index.html
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
