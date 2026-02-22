package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type VideoDistributionPlatform string

const (
	VideoDistributionPlatformTikTok    VideoDistributionPlatform = "tiktok"
	VideoDistributionPlatformYouTube   VideoDistributionPlatform = "youtube"
	VideoDistributionPlatformInstagram VideoDistributionPlatform = "instagram"
	VideoDistributionPlatformX         VideoDistributionPlatform = "x"
)

type VideoDistributionStatus string

const (
	VideoDistributionStatusPending    VideoDistributionStatus = "pending"
	VideoDistributionStatusProcessing VideoDistributionStatus = "processing"
	VideoDistributionStatusPublished  VideoDistributionStatus = "published"
	VideoDistributionStatusFailed     VideoDistributionStatus = "failed"
)

type VideoDistribution struct {
	ID           uint                    `gorm:"primaryKey;autoIncrement" json:"id"`
	MergeID      uint                    `gorm:"not null;index" json:"merge_id"`
	EpisodeID    uint                    `gorm:"not null;index" json:"episode_id"`
	DramaID      uint                    `gorm:"not null;index" json:"drama_id"`
	Platform     string                  `gorm:"type:varchar(32);not null;index" json:"platform"`
	Title        string                  `gorm:"type:varchar(200)" json:"title"`
	Description  string                  `gorm:"type:text" json:"description"`
	Hashtags     datatypes.JSON          `gorm:"type:json" json:"hashtags,omitempty"`
	SourceURL    string                  `gorm:"type:varchar(500);not null" json:"source_url"`
	Status       VideoDistributionStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Message      *string                 `gorm:"type:text" json:"message,omitempty"`
	PublishedURL *string                 `gorm:"type:varchar(500)" json:"published_url,omitempty"`
	ErrorMsg     *string                 `gorm:"type:text" json:"error_msg,omitempty"`
	StartedAt    *time.Time              `json:"started_at,omitempty"`
	CompletedAt  *time.Time              `json:"completed_at,omitempty"`
	CreatedAt    time.Time               `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time               `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt          `gorm:"index" json:"-"`

	Merge VideoMerge `gorm:"foreignKey:MergeID" json:"merge,omitempty"`
}

func (v *VideoDistribution) TableName() string {
	return "video_distributions"
}
