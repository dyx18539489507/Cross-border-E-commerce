package services

import (
	"fmt"

	"github.com/drama-generator/backend/domain/models"
)

// DeleteStoryboard 删除分镜记录
func (s *StoryboardService) DeleteStoryboard(storyboardID string) error {
	result := s.db.Delete(&models.Storyboard{}, storyboardID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete storyboard: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("storyboard not found")
	}
	return nil
}
