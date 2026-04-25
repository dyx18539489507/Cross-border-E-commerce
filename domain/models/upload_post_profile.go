package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UploadPostProfileStatus string

const (
	UploadPostProfileStatusPending UploadPostProfileStatus = "pending"
	UploadPostProfileStatusActive  UploadPostProfileStatus = "active"
	UploadPostProfileStatusError   UploadPostProfileStatus = "error"
)

type UploadPostProfile struct {
	ID                 uint                    `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID           string                  `gorm:"type:varchar(128);not null;default:'';uniqueIndex" json:"-"`
	Username           string                  `gorm:"type:varchar(120);not null;uniqueIndex" json:"username"`
	Status             UploadPostProfileStatus `gorm:"type:varchar(24);not null;default:'pending'" json:"status"`
	ConnectedPlatforms datatypes.JSON          `gorm:"type:json" json:"connected_platforms,omitempty"`
	ProfileSnapshot    datatypes.JSON          `gorm:"type:json" json:"profile_snapshot,omitempty"`
	LastSyncAt         *time.Time              `json:"last_sync_at,omitempty"`
	CreatedAt          time.Time               `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time               `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (u *UploadPostProfile) TableName() string {
	return "upload_post_profiles"
}
