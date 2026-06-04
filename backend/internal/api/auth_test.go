package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"provisioning-system/internal/config"
)

func TestIsValidSession_HardExpiration(t *testing.T) {
	cfg := &config.SystemConfig{}
	cfg.Auth.AutoLogoffTime = 0 // Disabled
	h := NewAuthHandler(cfg)

	// Add an expired session
	token := "expired-token"
	h.sessions[token] = &SessionInfo{
		Expiration:   time.Now().Add(-1 * time.Hour),
		LastActivity: time.Now().Add(-1 * time.Hour),
	}

	req := httptest.NewRequest("GET", "/api/phones", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})

	isValid, _ := h.isValidSession(req)
	if isValid {
		t.Error("Expected session to be invalid due to hard expiration, but got valid")
	}

	// Verify it was deleted
	h.mu.RLock()
	_, exists := h.sessions[token]
	h.mu.RUnlock()
	if exists {
		t.Error("Expected expired session to be deleted from sessions map")
	}
}

func TestIsValidSession_InactivityTimeout(t *testing.T) {
	cfg := &config.SystemConfig{}
	cfg.Auth.AutoLogoffTime = 5 // 5 minutes inactivity timeout
	h := NewAuthHandler(cfg)

	// Add an inactive session (inactive for 6 minutes)
	token := "inactive-token"
	h.sessions[token] = &SessionInfo{
		Expiration:   time.Now().Add(24 * time.Hour),
		LastActivity: time.Now().Add(-6 * time.Minute),
	}

	req := httptest.NewRequest("GET", "/api/phones", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})

	isValid, _ := h.isValidSession(req)
	if isValid {
		t.Error("Expected session to be invalid due to inactivity timeout, but got valid")
	}

	// Verify it was deleted
	h.mu.RLock()
	_, exists := h.sessions[token]
	h.mu.RUnlock()
	if exists {
		t.Error("Expected inactive session to be deleted from sessions map")
	}
}

func TestIsValidSession_CheckAuthDoesNotExtend(t *testing.T) {
	cfg := &config.SystemConfig{}
	cfg.Auth.AutoLogoffTime = 5
	h := NewAuthHandler(cfg)

	token := "valid-token"
	lastActivity := time.Now().Add(-2 * time.Minute)
	h.sessions[token] = &SessionInfo{
		Expiration:   time.Now().Add(24 * time.Hour),
		LastActivity: lastActivity,
	}

	// Test /api/check-auth
	req := httptest.NewRequest("GET", "/api/check-auth", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})

	isValid, remaining := h.isValidSession(req)
	if !isValid {
		t.Error("Expected session to be valid")
	}

	// Expected remaining: (5m - 2m) = 3m = 180 seconds. Allow slight variation due to time.Now().
	if remaining < 178 || remaining > 180 {
		t.Errorf("Expected remaining seconds to be around 180, got %d", remaining)
	}

	// Verify LastActivity did NOT change
	h.mu.RLock()
	session := h.sessions[token]
	h.mu.RUnlock()
	if !session.LastActivity.Equal(lastActivity) {
		t.Errorf("Expected LastActivity to remain %v, but got %v", lastActivity, session.LastActivity)
	}

	// Test /check-auth (alternative path)
	reqAlt := httptest.NewRequest("GET", "/check-auth", nil)
	reqAlt.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})

	isValid, _ = h.isValidSession(reqAlt)
	if !isValid {
		t.Error("Expected session to be valid")
	}

	// Verify LastActivity still did NOT change
	h.mu.RLock()
	session = h.sessions[token]
	h.mu.RUnlock()
	if !session.LastActivity.Equal(lastActivity) {
		t.Errorf("Expected LastActivity to remain %v, but got %v", lastActivity, session.LastActivity)
	}
}

func TestIsValidSession_ActiveRequestExtends(t *testing.T) {
	cfg := &config.SystemConfig{}
	cfg.Auth.AutoLogoffTime = 5
	h := NewAuthHandler(cfg)

	token := "valid-token"
	lastActivity := time.Now().Add(-2 * time.Minute)
	h.sessions[token] = &SessionInfo{
		Expiration:   time.Now().Add(24 * time.Hour),
		LastActivity: lastActivity,
	}

	req := httptest.NewRequest("GET", "/api/phones", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})

	isValid, remaining := h.isValidSession(req)
	if !isValid {
		t.Error("Expected session to be valid")
	}

	// Since the request is active, remaining seconds should be reset to full 5m = 300s
	if remaining != 300 {
		t.Errorf("Expected remaining seconds to be reset to 300, got %d", remaining)
	}

	// Verify LastActivity WAS extended (updated to recent time)
	h.mu.RLock()
	session := h.sessions[token]
	h.mu.RUnlock()
	if session.LastActivity.Equal(lastActivity) {
		t.Error("Expected LastActivity to be updated, but it remained the same")
	}
	if time.Since(session.LastActivity) > time.Second {
		t.Errorf("Expected LastActivity to be updated to now, got %v", session.LastActivity)
	}
}

func TestIsValidSession_PostCheckAuthResetsTimer(t *testing.T) {
	cfg := &config.SystemConfig{}
	cfg.Auth.AutoLogoffTime = 5
	h := NewAuthHandler(cfg)

	token := "valid-token"
	lastActivity := time.Now().Add(-2 * time.Minute)
	h.sessions[token] = &SessionInfo{
		Expiration:   time.Now().Add(24 * time.Hour),
		LastActivity: lastActivity,
	}

	// Post request to check-auth
	req := httptest.NewRequest("POST", "/api/check-auth", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: token,
	})

	isValid, remaining := h.isValidSession(req)
	if !isValid {
		t.Error("Expected session to be valid")
	}

	// Since request is POST, remaining seconds should be reset to full 300s
	if remaining != 300 {
		t.Errorf("Expected remaining seconds to be reset to 300, got %d", remaining)
	}

	// Verify LastActivity WAS extended (updated to recent time)
	h.mu.RLock()
	session := h.sessions[token]
	h.mu.RUnlock()
	if session.LastActivity.Equal(lastActivity) {
		t.Error("Expected LastActivity to be updated due to POST request, but it remained the same")
	}
}
