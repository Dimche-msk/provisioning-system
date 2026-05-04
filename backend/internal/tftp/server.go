package tftp

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"provisioning-system/internal/config"
	"provisioning-system/internal/devicelogger"

	"github.com/pin/tftp/v3"
)

type Server struct {
	ConfigDir    string
	Config       *config.SystemConfig
	DeviceLogger *devicelogger.DeviceLogger
	isRunning    bool
	lastError    error
}

func NewServer(configDir string, cfg *config.SystemConfig, dl *devicelogger.DeviceLogger) *Server {
	return &Server{
		ConfigDir:    configDir,
		Config:       cfg,
		DeviceLogger: dl,
	}
}

func (s *Server) Start() error {
	if !s.Config.Server.TFTPServer {
		return nil
	}

	addr := ":" + s.Config.Server.TFTPPort
	server := tftp.NewServer(s.readHandler, nil)
	server.SetTimeout(5 * time.Second)

	fmt.Printf("Starting TFTP Server on %s\n", addr)
	s.isRunning = true
	err := server.ListenAndServe(addr)
	if err != nil {
		s.isRunning = false
		s.lastError = err
		return fmt.Errorf("TFTP server failed to start on %s: %w", addr, err)
	}
	return nil
}

func (s *Server) Status() bool {
	return s.isRunning
}

func (s *Server) GetLastError() string {
	if s.lastError != nil {
		return s.lastError.Error()
	}
	return ""
}

func (s *Server) readHandler(filename string, rf io.ReaderFrom) error {
	// Simple path validation to prevent traversal
	cleanPath := filepath.Clean(filename)
	if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		// Even if it starts with /, we treat it as relative to our domains
		cleanPath = strings.TrimLeft(cleanPath, "/\\")
	}

	remoteAddr := ""
	if r, ok := rf.(interface{ RemoteAddr() net.Addr }); ok {
		remoteAddr = r.RemoteAddr().String()
	}
	clientIP, _, _ := net.SplitHostPort(remoteAddr)
	if clientIP == "" {
		clientIP = remoteAddr
	}

	configsDir := filepath.Join(s.ConfigDir, "temp_configs")
	
	// Fallback logic similar to HTTP server in main.go
	var foundPath string
	var foundDomain string

	// 1. Check default domain (first in config)
	if len(s.Config.Domains) > 0 {
		defaultDomain := s.Config.Domains[0].Name
		configPath := filepath.Join(configsDir, defaultDomain, cleanPath)
		if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
			foundPath = configPath
			foundDomain = defaultDomain
		}
	}

	// 2. Search in all other domains if not found in default
	if foundPath == "" {
		entries, err := os.ReadDir(configsDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					domain := entry.Name()
					configPath := filepath.Join(configsDir, domain, cleanPath)
					if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
						foundPath = configPath
						foundDomain = domain
						break
					}
				}
			}
		}
	}

	if foundPath == "" {
		s.DeviceLogger.LogAccess(clientIP, 404, "TFTP", "/"+cleanPath, "TFTP Client", "File not found")
		return fmt.Errorf("file not found")
	}

	file, err := os.Open(foundPath)
	if err != nil {
		s.DeviceLogger.LogAccess(clientIP, 500, "TFTP", "/"+cleanPath, "TFTP Client", fmt.Sprintf("Error opening file: %v", err))
		return err
	}
	defer file.Close()

	s.DeviceLogger.LogAccess(clientIP, 200, "TFTP", "/"+cleanPath, "TFTP Client", fmt.Sprintf("Served from domain: %s", foundDomain))
	
	_, err = rf.ReadFrom(file)
	return err
}
