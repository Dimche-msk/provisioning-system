package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"provisioning-system/internal/config"

	"github.com/google/uuid"
)

type SessionInfo struct {
	Expiration   time.Time
	LastActivity time.Time
}

type AuthHandler struct {
	Config   *config.SystemConfig
	sessions map[string]*SessionInfo // Token -> SessionInfo
	mu       sync.RWMutex
}

func NewAuthHandler(cfg *config.SystemConfig) *AuthHandler {
	return &AuthHandler{
		Config:   cfg,
		sessions: make(map[string]*SessionInfo),
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == h.Config.Auth.AdminUser && req.Password == h.Config.Auth.AdminPassword {
		// Генерация случайного токена
		token := uuid.New().String()
		expiration := time.Now().Add(24 * time.Hour)

		// Сохранение сессии
		h.mu.Lock()
		h.sessions[token] = &SessionInfo{
			Expiration:   expiration,
			LastActivity: time.Now(),
		}
		h.mu.Unlock()

		// Устанавливаем куку
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Path:     "/",
			Expires:  expiration,
			HttpOnly: true,
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok", "message": "Logged in"}`))
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Удаляем сессию из памяти
		h.mu.Lock()
		delete(h.sessions, cookie.Value)
		h.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})
	w.Write([]byte(`{"status": "ok", "message": "Logged out"}`))
}

func (h *AuthHandler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	isValid, remainingSeconds := h.isValidSession(r)
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            "ok",
		"remaining_seconds": remainingSeconds,
	})
}

// Вспомогательная функция проверки сессии
func (h *AuthHandler) isValidSession(r *http.Request) (bool, int) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false, 0
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	session, exists := h.sessions[cookie.Value]
	if !exists {
		return false, 0
	}

	now := time.Now()

	// Check hard expiration
	if now.After(session.Expiration) {
		delete(h.sessions, cookie.Value)
		return false, 0
	}

	// Check inactivity timeout
	autoLogoffTime := h.Config.Auth.AutoLogoffTime
	remainingSeconds := 0

	if autoLogoffTime > 0 {
		timeoutDuration := time.Duration(autoLogoffTime) * time.Minute
		elapsed := now.Sub(session.LastActivity)
		if elapsed > timeoutDuration {
			delete(h.sessions, cookie.Value)
			return false, 0
		}
		remainingSeconds = int((timeoutDuration - elapsed).Seconds())
	}

	// Update last activity time (except check-auth endpoint GET requests to prevent keeping session alive via background checking)
	cleanPath := strings.TrimSuffix(r.URL.Path, "/")
	if !strings.HasSuffix(cleanPath, "/check-auth") || r.Method == http.MethodPost {
		session.LastActivity = now
		if autoLogoffTime > 0 {
			remainingSeconds = autoLogoffTime * 60
		}
	}

	return true, remainingSeconds
}

// Middleware для защиты роутов
func (h *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isValid, _ := h.isValidSession(r)
		if !isValid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
