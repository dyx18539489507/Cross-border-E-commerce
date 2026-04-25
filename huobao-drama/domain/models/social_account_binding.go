package models

import (
	"time"

	"gorm.io/gorm"
)

type SocialAccountBinding struct {
	ID                uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID          string         `gorm:"type:varchar(128);not null;default:'';uniqueIndex:idx_social_binding_device_platform" json:"-"`
	Platform          string         `gorm:"type:varchar(32);not null;uniqueIndex:idx_social_binding_device_platform" json:"platform"`
	AccountIdentifier string         `gorm:"type:varchar(120);not null" json:"account_identifier"`
	DisplayName       *string        `gorm:"type:varchar(120)" json:"display_name,omitempty"`
	CreatedAt         time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *SocialAccountBinding) TableName() string {
	return "social_account_bindings"
}
