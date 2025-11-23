package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"provisioning-system/internal/config"

	"github.com/google/uuid"
)

type AuthHandler struct {
	Config   *config.SystemConfig
	sessions map[string]time.Time // Token -> Expiration
	mu       sync.RWMutex
}

func NewAuthHandler(cfg *config.SystemConfig) *AuthHandler {
	return &AuthHandler{
		Config:   cfg,
		sessions: make(map[string]time.Time),
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
		h.sessions[token] = expiration
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
	if !h.isValidSession(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

// Вспомогательная функция проверки сессии
func (h *AuthHandler) isValidSession(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}

	h.mu.RLock()
	expiration, exists := h.sessions[cookie.Value]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiration) {
		// Сессия истекла, удаляем (ленивая очистка)
		h.mu.Lock()
		delete(h.sessions, cookie.Value)
		h.mu.Unlock()
		return false
	}

	return true
}

// Middleware для защиты роутов
func (h *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.isValidSession(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
