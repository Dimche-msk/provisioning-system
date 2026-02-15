package license

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"provisioning-system/internal/logger"
)

type Tier string

const (
	Free Tier = "Free"
	Pro  Tier = "Pro"
	VIP  Tier = "VIP"
)

type LicenseInfo struct {
	Tier         Tier      `json:"tier"`
	CustomerID   string    `json:"customer_id"`
	IssuedTo     string    `json:"issued_to"`
	ValidFrom    time.Time `json:"valid_from"`
	Expiry       time.Time `json:"expiry"`
	SupportLevel string    `json:"support_level"`
	LicenseKey   string    `json:"license_key"`
}

type Manager struct {
	mu          sync.RWMutex
	licensePath string
	current     *LicenseInfo
}

func NewManager(configDir string) *Manager {
	m := &Manager{
		licensePath: filepath.Join(configDir, "license.key"),
	}
	m.Reload()
	return m
}

func (m *Manager) Reload() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Default to Free
	m.current = &LicenseInfo{
		Tier:         Free,
		SupportLevel: "Community",
	}

	if _, err := os.Stat(m.licensePath); os.IsNotExist(err) {
		logger.Debug("No license key found, running in Free mode")
		return
	}

	data, err := os.ReadFile(m.licensePath)
	if err != nil {
		logger.Error("Failed to read license key: %v", err)
		return
	}

	var info LicenseInfo
	if err := json.Unmarshal(data, &info); err != nil {
		logger.Error("Invalid license key format: %v", err)
		return
	}

	// Basic check: is it expired?
	if !info.Expiry.IsZero() && time.Now().After(info.Expiry) {
		logger.Warn("License key is expired")
		// We still load it so the UI can show "Expired" status if needed
		// But in a real system we might force Free tier here
	}

	m.current = &info
	logger.Info("License loaded: Tier=%s, Customer=%s, IssuedTo=%s", m.current.Tier, m.current.CustomerID, m.current.IssuedTo)
}

func (m *Manager) GetStatus() LicenseInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.current
}

func (m *Manager) IsPro() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current.Tier == Pro || m.current.Tier == VIP
}

func (m *Manager) SupportLevel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current.SupportLevel
}
