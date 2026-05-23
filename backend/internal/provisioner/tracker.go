package provisioner

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"provisioning-system/internal/logger"
	"provisioning-system/internal/models"

	"gorm.io/gorm"
)

type TrackerEntry struct {
	PhoneID   uint
	CurrentIP string
}

type PhoneTracker struct {
	mu      sync.RWMutex
	entries map[string]*TrackerEntry
	db      *gorm.DB
	manager *Manager
}

func NewPhoneTracker(manager *Manager) *PhoneTracker {
	return &PhoneTracker{
		entries: make(map[string]*TrackerEntry),
		manager: manager,
	}
}

func (t *PhoneTracker) Initialize(db *gorm.DB) error {
	t.mu.Lock()
	t.db = db
	t.entries = make(map[string]*TrackerEntry)
	t.mu.Unlock()

	return t.Rebuild()
}

func (t *PhoneTracker) Rebuild() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.db == nil {
		return fmt.Errorf("database not initialized in tracker")
	}

	// 1. Find domains with IP tracking enabled
	trackedDomains := make(map[string]bool)
	for _, domain := range t.manager.Config.Domains {
		effective := t.manager.Config.GetEffectiveDomainConfig(domain.Name)
		if effective.TrackPhoneIPs {
			trackedDomains[domain.Name] = true
		}
	}

	if len(trackedDomains) == 0 {
		// Clear entries and exit if tracking is not enabled on any domain
		t.entries = make(map[string]*TrackerEntry)
		return nil
	}

	// 2. Query all phones from database
	var phones []models.Phone
	if err := t.db.Find(&phones).Error; err != nil {
		return fmt.Errorf("failed to fetch phones for tracker rebuild: %w", err)
	}

	// 3. Build the map
	newEntries := make(map[string]*TrackerEntry)
	for _, phone := range phones {
		if trackedDomains[phone.Domain] {
			filename := t.getPhoneFilename(phone)
			if filename != "" {
				ip := phone.IPAddress
				newEntries[filename] = &TrackerEntry{
					PhoneID:   phone.ID,
					CurrentIP: ip,
				}
			}
		}
	}

	t.entries = newEntries
	logger.Info("Phone IP Tracker rebuilt successfully. Tracking %d configuration files.", len(t.entries))
	return nil
}

func (t *PhoneTracker) getPhoneFilename(phone models.Phone) string {
	path, err := t.manager.GetPhoneConfigPath("", phone)
	if err != nil || path == "" {
		return ""
	}
	return strings.ToLower(filepath.Base(path))
}

func (t *PhoneTracker) UpdateIP(filename string, clientIP string) {
	if clientIP == "" {
		return
	}

	filenameLower := strings.ToLower(filename)

	t.mu.RLock()
	entry, exists := t.entries[filenameLower]
	t.mu.RUnlock()

	if !exists {
		return
	}

	if entry.CurrentIP == clientIP {
		// IP has not changed
		return
	}

	// IP has changed! Update database and in-memory map
	t.mu.Lock()
	// Re-verify under write lock
	entry, exists = t.entries[filenameLower]
	if !exists || entry.CurrentIP == clientIP {
		t.mu.Unlock()
		return
	}

	oldIP := entry.CurrentIP
	entry.CurrentIP = clientIP
	phoneID := entry.PhoneID
	t.mu.Unlock()

	// Update database
	logger.Info("Phone IP change detected for file %s (ID %d). Old IP: %s, New IP: %s. Updating DB.", filename, phoneID, oldIP, clientIP)
	if err := t.db.Model(&models.Phone{}).Where("id = ?", phoneID).Update("ip_address", clientIP).Error; err != nil {
		logger.Error("Failed to update phone IP in database: %v", err)
	}
}

func (t *PhoneTracker) RegisterPhone(phone models.Phone) {
	effective := t.manager.Config.GetEffectiveDomainConfig(phone.Domain)
	if !effective.TrackPhoneIPs {
		return
	}

	filename := t.getPhoneFilename(phone)
	if filename == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	ip := phone.IPAddress

	t.entries[filename] = &TrackerEntry{
		PhoneID:   phone.ID,
		CurrentIP: ip,
	}
	logger.Info("Registered phone %s in IP tracker", filename)
}

func (t *PhoneTracker) UnregisterPhone(phone models.Phone) {
	filename := t.getPhoneFilename(phone)
	if filename == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.entries, filename)
	logger.Info("Unregistered phone %s from IP tracker", filename)
}
