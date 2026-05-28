package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SilkroadAgentSession struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID       string         `gorm:"type:varchar(128);not null;default:'';index" json:"-"`
	RequestID      string         `gorm:"type:varchar(80);not null;default:'';index" json:"request_id"`
	ProductName    string         `gorm:"type:varchar(200);not null;default:''" json:"product_name"`
	Category       string         `gorm:"type:varchar(120);not null;default:''" json:"category"`
	TargetMarket   string         `gorm:"type:varchar(120);not null;default:''" json:"target_market"`
	TargetPlatform string         `gorm:"type:varchar(120);not null;default:''" json:"target_platform"`
	TargetAudience string         `gorm:"type:varchar(200);not null;default:''" json:"target_audience"`
	RawPrompt      string         `gorm:"type:text" json:"raw_prompt"`
	InputSnapshot  datatypes.JSON `gorm:"type:json" json:"input_snapshot"`
	ResultSnapshot datatypes.JSON `gorm:"type:json" json:"result_snapshot"`
	Status         string         `gorm:"type:varchar(24);not null;default:'completed';index" json:"status"`
	Model          string         `gorm:"type:varchar(120);not null;default:''" json:"model"`
	CreatedAt      time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SilkroadAgentSession) TableName() string {
	return "silkroad_agent_sessions"
}
