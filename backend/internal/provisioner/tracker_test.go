package provisioner

import (
	"testing"

	"provisioning-system/internal/config"
	"provisioning-system/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestPhoneTracker(t *testing.T) {
	// 1. Setup sqlite in-memory DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := db.AutoMigrate(&models.Phone{}); err != nil {
		t.Fatalf("Failed to auto-migrate: %v", err)
	}

	// 2. Setup Config with domains
	cfg := &config.SystemConfig{
		Domains: []config.DomainSettings{
			{
				Name:          "domain-tracked",
				TrackPhoneIPs: true,
			},
			{
				Name:          "domain-untracked",
				TrackPhoneIPs: false,
			},
		},
	}

	// 3. Create provisioner.Manager
	m := NewManager(cfg)
	// Add mock vendor
	m.Vendors = []VendorConfig{
		{
			ID:              "mitel",
			PhoneConfigFile: "{{account.mac_address|lower}}.cfg",
		},
	}

	// 4. Create Phones in DB
	mac1 := "001122334455"
	phoneTracked := models.Phone{
		Domain:     "domain-tracked",
		Vendor:     "mitel",
		MacAddress: &mac1,
		IPAddress:  "192.168.1.10",
	}
	if err := db.Create(&phoneTracked).Error; err != nil {
		t.Fatalf("Failed to create phone: %v", err)
	}

	mac2 := "554433221100"
	phoneUntracked := models.Phone{
		Domain:     "domain-untracked",
		Vendor:     "mitel",
		MacAddress: &mac2,
		IPAddress:  "192.168.1.20",
	}
	if err := db.Create(&phoneUntracked).Error; err != nil {
		t.Fatalf("Failed to create phone: %v", err)
	}

	// 5. Initialize tracker
	if err := m.InitTracker(db); err != nil {
		t.Fatalf("Failed to init tracker: %v", err)
	}

	// Check if only phoneTracked is tracked
	m.Tracker.mu.RLock()
	trackedCount := len(m.Tracker.entries)
	m.Tracker.mu.RUnlock()

	if trackedCount != 1 {
		t.Errorf("Expected 1 tracked phone, got %d", trackedCount)
	}

	// 6. Test UpdateIP
	// Different IP
	m.Tracker.UpdateIP("001122334455.cfg", "192.168.1.100")

	// Read from DB
	var dbPhone models.Phone
	if err := db.First(&dbPhone, phoneTracked.ID).Error; err != nil {
		t.Fatalf("Failed to fetch phone from DB: %v", err)
	}
	if dbPhone.IPAddress != "192.168.1.100" {
		t.Errorf("Expected IPAddress to be updated to 192.168.1.100, got %s", dbPhone.IPAddress)
	}

	// Test UpdateIP with untracked config file -> should do nothing
	m.Tracker.UpdateIP("554433221100.cfg", "192.168.1.200")
	var dbPhone2 models.Phone
	if err := db.First(&dbPhone2, phoneUntracked.ID).Error; err != nil {
		t.Fatalf("Failed to fetch phone from DB: %v", err)
	}
	if dbPhone2.IPAddress != "192.168.1.20" {
		t.Errorf("Expected untracked phone IPAddress to remain unchanged, got %s", dbPhone2.IPAddress)
	}
}
