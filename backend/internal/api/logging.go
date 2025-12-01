package api

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"provisioning-system/internal/broadcaster"
	"provisioning-system/internal/config"
)

type DeviceLogger struct {
	Config      *config.SystemConfig
	Broadcaster *broadcaster.Broadcaster
	mu          sync.Mutex
}

func NewDeviceLogger(cfg *config.SystemConfig, b *broadcaster.Broadcaster) *DeviceLogger {
	return &DeviceLogger{
		Config:      cfg,
		Broadcaster: b,
	}
}

// responseWriterWrapper captures the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (l *DeviceLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if logging is enabled
		logLevel := l.Config.Server.LogDeviceAccess
		if logLevel == "" || logLevel == "none" {
			next.ServeHTTP(w, r)
			return
		}

		wrapper := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(wrapper, r)

		// Determine if we should log based on level
		shouldLog := false
		if logLevel == "full" || logLevel == "access" {
			shouldLog = true
		} else if logLevel == "error" && wrapper.statusCode >= 400 {
			shouldLog = true
		}

		if shouldLog {
			// Extract IP (handle X-Forwarded-For if behind proxy, but simple RemoteAddr for now)
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			event := broadcaster.LogEvent{
				Time:          start,
				SourceIP:      ip,
				StatusCode:    wrapper.statusCode,
				RequestedFile: r.URL.Path,
				Method:        r.Method,
				UserAgent:     r.UserAgent(),
			}

			// Broadcast
			if l.Broadcaster != nil {
				l.Broadcaster.Broadcast(event)
			}

			// Log to file
			if l.Config.Server.LogFilePath != "" {
				l.logToFile(event)
			}
		}
	})
}

func (l *DeviceLogger) logToFile(event broadcaster.LogEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.Config.Server.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer f.Close()

	logLine := fmt.Sprintf("%s | %s | %d | %s | %s | %s | %s\n",
		event.Time.Format(time.RFC3339),
		event.SourceIP,
		event.StatusCode,
		event.Method,
		event.RequestedFile,
		event.UserAgent,
		event.Message,
	)

	if _, err := f.WriteString(logLine); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

func (l *DeviceLogger) LogCustom(r *http.Request, statusCode int, message string) {
	logLevel := l.Config.Server.LogDeviceAccess
	if logLevel == "" || logLevel == "none" {
		return
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	event := broadcaster.LogEvent{
		Time:          time.Now(),
		SourceIP:      ip,
		StatusCode:    statusCode,
		RequestedFile: r.URL.Path,
		Method:        r.Method,
		UserAgent:     r.UserAgent(),
		Message:       message,
	}

	if l.Broadcaster != nil {
		l.Broadcaster.Broadcast(event)
	}

	if l.Config.Server.LogFilePath != "" {
		l.logToFile(event)
	}
}
