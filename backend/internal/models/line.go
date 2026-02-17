package models

import (
	"encoding/json"
	"time"
)

type PhoneLine struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	PhoneID        uint   `gorm:"index;uniqueIndex:idx_phone_type_key_panel" json:"phone_id"`
	Type           string `gorm:"uniqueIndex:idx_phone_type_key_panel" json:"type"` // "Line", "Free", etc.
	KeyNumber      *int   `gorm:"uniqueIndex:idx_phone_type_key_panel" json:"key_number"`
	PanelNumber    *int   `gorm:"uniqueIndex:idx_phone_type_key_panel" json:"panel_number"` // 0 if main device
	AccountNumber  int    `json:"account_number"`                                           // Association with SIP account
	AdditionalInfo string `json:"additional_info"`                                          // JSON string
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
