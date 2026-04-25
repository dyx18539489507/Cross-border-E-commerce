package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DistributionResultStatus string

const (
	DistributionResultStatusPending    DistributionResultStatus = "pending"
	DistributionResultStatusScheduled  DistributionResultStatus = "scheduled"
	DistributionResultStatusProcessing DistributionResultStatus = "processing"
	DistributionResultStatusSuccess    DistributionResultStatus = "success"
	DistributionResultStatusFailed     DistributionResultStatus = "failed"
)

type DistributionResult struct {
	ID                uint                    `gorm:"primaryKey;autoIncrement" json:"id"`
	JobID             uint                    `gorm:"not null;index" json:"job_id"`
	DeviceID          string                  `gorm:"type:varchar(128);not null;default:'';index" json:"-"`
	Platform          DistributionPlatform    `gorm:"type:varchar(32);not null;index" json:"platform"`
	TargetID          *uint                   `gorm:"index" json:"target_id,omitempty"`
	ContentType       DistributionContentType `gorm:"type:varchar(16);not null;index" json:"content_type"`
	Status            DistributionResultStatus `gorm:"type:varchar(24);not null;default:'pending';index" json:"status"`
	TargetSnapshot    datatypes.JSON          `gorm:"type:json" json:"target_snapshot,omitempty"`
	RequestSnapshot   datatypes.JSON          `gorm:"type:json" json:"request_snapshot,omitempty"`
	ResponseSnapshot  datatypes.JSON          `gorm:"type:json" json:"response_snapshot,omitempty"`
	ExternalRequestID *string                 `gorm:"type:varchar(120);index" json:"request_id,omitempty"`
	ExternalJobID     *string                 `gorm:"type:varchar(120);index" json:"job_id_external,omitempty"`
	ExternalMessageID *string                 `gorm:"type:varchar(120)" json:"message_id,omitempty"`
	PublishedURL      *string                 `gorm:"type:text" json:"published_url,omitempty"`
	ErrorMsg          *string                 `gorm:"type:text" json:"error_msg,omitempty"`
	AttemptCount      int                     `gorm:"not null;default:0" json:"attempt_count"`
	NextRetryAt       *time.Time              `gorm:"index" json:"next_retry_at,omitempty"`
	StartedAt         *time.Time              `json:"started_at,omitempty"`
	CompletedAt       *time.Time              `json:"completed_at,omitempty"`
	CreatedAt         time.Time               `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time               `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt          `gorm:"index" json:"-"`

	Job    DistributionJob    `gorm:"foreignKey:JobID" json:"job,omitempty"`
	Target *DistributionTarget `gorm:"foreignKey:TargetID" json:"target,omitempty"`
}

func (d *DistributionResult) TableName() string {
	return "distribution_results"
}
