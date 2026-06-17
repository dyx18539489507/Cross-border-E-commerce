package services

import (
	"fmt"

	"github.com/drama-generator/backend/domain/models"
)

// DeleteStoryboard 删除分镜记录
func (s *StoryboardService) DeleteStoryboard(storyboardID string, deviceIDs ...string) error {
	deviceID := firstStoryboardDeviceID(deviceIDs)
	query := s.db.Model(&models.Storyboard{}).Where("storyboards.id = ?", storyboardID)
	if deviceID != "" {
		query = query.Joins("JOIN episodes ON episodes.id = storyboards.episode_id").
			Joins("JOIN dramas ON dramas.id = episodes.drama_id").
			Where("dramas.device_id = ?", deviceID)
	}

	result := query.Delete(&models.Storyboard{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete storyboard: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("storyboard not found")
	}
	return nil
}
