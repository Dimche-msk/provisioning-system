package models

import (
	"time"

	"gorm.io/gorm"
)

type Phone struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Domain                string      `gorm:"index" json:"domain"`
	Vendor                string      `json:"vendor"`
	ModelID               string      `json:"model_id"` // ID модели телефона (например, "yealink-t46s")
	ExpansionModulesCount int         `json:"expansion_modules_count"`
	ExpansionModuleModel  string      `json:"expansion_module_model"` // Модель модуля расширения (например, "M680")
	Type                  string      `json:"type"`                   // "phone" or "gateway"
	MacAddress            *string     `gorm:"uniqueIndex" json:"mac_address"`
	PhoneNumber           *string     `gorm:"uniqueIndex" json:"phone_number"` // Used for search
	IPAddress             string      `json:"ip_address"`
	Description           string      `json:"description"`
	Lines                 []PhoneLine `gorm:"foreignKey:PhoneID" json:"lines"`
}
