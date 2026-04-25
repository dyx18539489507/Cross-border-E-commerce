package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DistributionTargetType string

const (
	DistributionTargetTypeUploadPostProfile DistributionTargetType = "upload_post_profile"
	DistributionTargetTypePinterestBoard    DistributionTargetType = "pinterest_board"
	DistributionTargetTypeRedditSubreddit   DistributionTargetType = "reddit_subreddit"
	DistributionTargetTypeDiscordWebhook    DistributionTargetType = "discord_webhook"
)

type DistributionTargetStatus string

const (
	DistributionTargetStatusPending     DistributionTargetStatus = "pending"
	DistributionTargetStatusActive      DistributionTargetStatus = "active"
	DistributionTargetStatusNeedsRebind DistributionTargetStatus = "needs_rebind"
	DistributionTargetStatusDisabled    DistributionTargetStatus = "disabled"
)

type DistributionTarget struct {
	ID              uint                     `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID        string                   `gorm:"type:varchar(128);not null;default:'';index;uniqueIndex:idx_distribution_target_identity" json:"-"`
	Platform        DistributionPlatform     `gorm:"type:varchar(32);not null;index;uniqueIndex:idx_distribution_target_identity" json:"platform"`
	TargetType      DistributionTargetType   `gorm:"type:varchar(40);not null;index;uniqueIndex:idx_distribution_target_identity" json:"target_type"`
	Identifier      string                   `gorm:"type:varchar(200);not null;uniqueIndex:idx_distribution_target_identity" json:"identifier"`
	Name            *string                  `gorm:"type:varchar(200)" json:"name,omitempty"`
	Status          DistributionTargetStatus `gorm:"type:varchar(24);not null;default:'active';index" json:"status"`
	IsDefault       bool                     `gorm:"not null;default:false;index" json:"is_default"`
	Config          datatypes.JSON           `gorm:"type:json" json:"config,omitempty"`
	SecretEncrypted *string                  `gorm:"type:text" json:"-"`
	LastValidatedAt *time.Time               `json:"last_validated_at,omitempty"`
	LastSyncAt      *time.Time               `json:"last_sync_at,omitempty"`
	CreatedAt       time.Time                `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time                `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt           `gorm:"index" json:"-"`
}

func (d *DistributionTarget) TableName() string {
	return "distribution_targets"
}
