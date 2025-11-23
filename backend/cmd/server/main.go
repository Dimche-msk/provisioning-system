package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	"provisioning-system/internal/api"
	"provisioning-system/internal/broadcaster"
	"provisioning-system/internal/config"
	"provisioning-system/internal/db"
	"provisioning-system/internal/provisioner"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	// 1. Парсинг аргументов CLI
	configDir := flag.String("config-dir", ".", "Directory containing provisioning-system.yaml and vendors/")
	flag.Parse()

	// 2. Загрузка конфигурации
	cfg, err := config.LoadConfig(*configDir)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	port := ":" + cfg.Server.Port
	fmt.Printf("Starting Provisioning Server on port %s...\n", port)
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
		// Используем http.StripPrefix чтобы убрать /config/ из пути
		configFs := http.FileServer(http.Dir(configsDir))

		// Оборачиваем в логгер
		loggingHandler := deviceLogger.Middleware(http.StripPrefix("/config/", configFs))

		r.PathPrefix("/config/").Handler(loggingHandler)
		fmt.Printf("Serving generated configs at /config/ from %s\n", configsDir)
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

	// 7. Инициализация API Handlers
	sysHandler := api.NewSystemHandler(*configDir, &cfg, provManager, database)
	phoneHandler := api.NewPhoneHandler(database, provManager)
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
	protected.HandleFunc("/domains", sysHandler.GetDomains).Methods("GET")
	protected.HandleFunc("/deploy", sysHandler.Deploy).Methods("POST")

	protected.HandleFunc("/phones", phoneHandler.CreatePhone).Methods("POST")
	protected.HandleFunc("/phones", phoneHandler.GetPhones).Methods("GET")
	protected.HandleFunc("/phones/{id}", phoneHandler.UpdatePhone).Methods("PUT")

	protected.HandleFunc("/vendors", phoneHandler.GetVendors).Methods("GET")
	protected.HandleFunc("/models", phoneHandler.GetModels).Methods("GET")

	// Debug API (SSE)
	protected.HandleFunc("/debug/logs", debugHandler.StreamLogs).Methods("GET")

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

		if path == "/" {
			fileServer.ServeHTTP(w, r)
			return
		}

		f, err := dist.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			// SPA Fallback
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
