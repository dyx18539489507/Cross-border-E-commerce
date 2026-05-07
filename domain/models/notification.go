package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Notification struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID    string         `gorm:"type:varchar(128);not null;default:'';index" json:"-"`
	Type        string         `gorm:"type:varchar(50);not null;default:'system';index" json:"type"`
	Title       string         `gorm:"type:varchar(120);not null" json:"title"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Path        string         `gorm:"type:varchar(500);not null;default:''" json:"path,omitempty"`
	Metadata    datatypes.JSON `gorm:"type:json" json:"metadata,omitempty"`
	ReadAt      *time.Time     `gorm:"index" json:"read_at,omitempty"`
	DismissedAt *time.Time     `gorm:"index" json:"dismissed_at,omitempty"`
	CreatedAt   time.Time      `gorm:"not null;autoCreateTime;index" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}
