package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DistributionPlatform string

const (
	DistributionPlatformDiscord   DistributionPlatform = "discord"
	DistributionPlatformReddit    DistributionPlatform = "reddit"
	DistributionPlatformPinterest DistributionPlatform = "pinterest"
)

type DistributionContentType string

const (
	DistributionContentTypeText  DistributionContentType = "text"
	DistributionContentTypeImage DistributionContentType = "image"
	DistributionContentTypeVideo DistributionContentType = "video"
)

type DistributionPublishMode string

const (
	DistributionPublishModeImmediate DistributionPublishMode = "immediate"
	DistributionPublishModeSchedule  DistributionPublishMode = "schedule"
)

type DistributionJobStatus string

const (
	DistributionJobStatusPending         DistributionJobStatus = "pending"
	DistributionJobStatusScheduled       DistributionJobStatus = "scheduled"
	DistributionJobStatusProcessing      DistributionJobStatus = "processing"
	DistributionJobStatusCompleted       DistributionJobStatus = "completed"
	DistributionJobStatusPartiallyFailed DistributionJobStatus = "partially_failed"
	DistributionJobStatusFailed          DistributionJobStatus = "failed"
)

type DistributionSourceType string

const (
	DistributionSourceTypeManual         DistributionSourceType = "manual"
	DistributionSourceTypeVideoMerge     DistributionSourceType = "video_merge"
	DistributionSourceTypeImageGen       DistributionSourceType = "image_generation"
	DistributionSourceTypeVideoGen       DistributionSourceType = "video_generation"
	DistributionSourceTypeAsset          DistributionSourceType = "asset"
	DistributionSourceTypeStoryboard     DistributionSourceType = "storyboard"
	DistributionSourceTypeEpisode        DistributionSourceType = "episode"
	DistributionSourceTypeUnknown        DistributionSourceType = "unknown"
)

type DistributionJob struct {
	ID                uint                    `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID          string                  `gorm:"type:varchar(128);not null;default:'';index" json:"-"`
	SourceType        DistributionSourceType  `gorm:"type:varchar(32);not null;default:'manual';index" json:"source_type"`
	SourceRef         *string                 `gorm:"type:varchar(120);index" json:"source_ref,omitempty"`
	ContentType       DistributionContentType `gorm:"type:varchar(16);not null;index" json:"content_type"`
	Title             *string                 `gorm:"type:varchar(200)" json:"title,omitempty"`
	Body              *string                 `gorm:"type:text" json:"body,omitempty"`
	MediaURL          *string                 `gorm:"type:text" json:"media_url,omitempty"`
	SelectedPlatforms datatypes.JSON          `gorm:"type:json" json:"selected_platforms,omitempty"`
	PlatformOptions   datatypes.JSON          `gorm:"type:json" json:"platform_options,omitempty"`
	PublishMode       DistributionPublishMode `gorm:"type:varchar(16);not null;default:'immediate'" json:"publish_mode"`
	ScheduledAt       *time.Time              `gorm:"index" json:"scheduled_at,omitempty"`
	Status            DistributionJobStatus   `gorm:"type:varchar(24);not null;default:'pending';index" json:"status"`
	RequestSnapshot   datatypes.JSON          `gorm:"type:json" json:"request_snapshot,omitempty"`
	ErrorMsg          *string                 `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt         time.Time               `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time               `gorm:"not null;autoUpdateTime" json:"updated_at"`
	CompletedAt       *time.Time              `json:"completed_at,omitempty"`
	DeletedAt         gorm.DeletedAt          `gorm:"index" json:"-"`

	Results []DistributionResult `gorm:"foreignKey:JobID" json:"results,omitempty"`
}

func (d *DistributionJob) TableName() string {
	return "distribution_jobs"
}
