package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	CustomVoiceStatusProcessing = "processing"
	CustomVoiceStatusCompleted  = "completed"
	CustomVoiceStatusFailed     = "failed"
)

// CustomVoice 自定义音色（声音复刻）
type CustomVoice struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string         `gorm:"type:varchar(100);not null" json:"name"`
	Provider       string         `gorm:"type:varchar(30);not null;default:'volcengine'" json:"provider"`
	SpeakerID      string         `gorm:"type:varchar(120);not null;index" json:"speaker_id"`
	VoiceType      string         `gorm:"type:varchar(120);not null;index" json:"voice_type"`
	ResourceID     string         `gorm:"type:varchar(120)" json:"resource_id"`
	SourceAudioURL string         `gorm:"type:varchar(500);not null" json:"source_audio_url"`
	TrialURL       string         `gorm:"type:varchar(500)" json:"trial_url"`
	Status         string         `gorm:"type:varchar(20);not null;default:'processing';index" json:"status"`
	LastError      *string        `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt      time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *CustomVoice) TableName() string {
	return "custom_voices"
}

func (c *CustomVoice) PublicID() string {
	return "custom-" + fmtUint(c.ID)
}

func fmtUint(value uint) string {
	if value == 0 {
		return "0"
	}

	digits := [20]byte{}
	index := len(digits)
	n := value
	for n > 0 {
		index--
		digits[index] = byte('0' + (n % 10))
		n /= 10
	}
	return string(digits[index:])
}
