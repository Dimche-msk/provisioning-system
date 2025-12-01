package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type PhoneLine struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PhoneID               uint   `gorm:"index" json:"phone_id"`
	Type                  string `json:"type"`                    // "line", "hard_key", "soft_key"
	Number                int    `json:"number"`                  // Sequential number
	ExpansionModuleNumber int    `json:"expansion_module_number"` // 0 if main device
	KeyNumber             int    `json:"key_number"`              // Number within module/keys
	AdditionalInfo        string `json:"additional_info"`         // JSON string
}

func (l PhoneLine) GetAdditionalInfoMap() map[string]interface{} {
	var m map[string]interface{}
	if l.AdditionalInfo == "" {
		return make(map[string]interface{})
	}
	if err := json.Unmarshal([]byte(l.AdditionalInfo), &m); err != nil {
		return make(map[string]interface{})
	}
	return m
}
