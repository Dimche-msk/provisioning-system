package models

import (
	"time"

	"gorm.io/gorm"
)

type PhoneLine struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PhoneID         uint   `gorm:"index" json:"phone_id"`
	AccountName     string `json:"account_name"`              // Line name (e.g., "1", "line.1")
	PhoneNumber     string `gorm:"index" json:"phone_number"` // Can be non-numeric
	Domain          string `json:"domain"`                    // Can differ from phone domain
	CallerID        string `json:"caller_id"`
	AccountSettings string `json:"account_settings"` // JSON
	Description     string `json:"description"`
}
