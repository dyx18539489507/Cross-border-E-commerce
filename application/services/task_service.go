package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewTaskService(db *gorm.DB, log *logger.Logger) *TaskService {
	return &TaskService{
		db:  db,
		log: log,
	}
}

func firstTaskDeviceID(deviceIDs []string) string {
	if len(deviceIDs) == 0 {
		return ""
	}
	return deviceIDs[0]
}

// CreateTask 创建新任务
func (s *TaskService) CreateTask(taskType, resourceID string, deviceIDs ...string) (*models.AsyncTask, error) {
	deviceID := firstTaskDeviceID(deviceIDs)
	task := &models.AsyncTask{
		ID:         uuid.New().String(),
		DeviceID:   deviceID,
		Type:       taskType,
		Status:     "pending",
		Progress:   0,
		ResourceID: resourceID,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// UpdateTaskStatus 更新任务状态
func (s *TaskService) UpdateTaskStatus(taskID, status string, progress int, message string) error {
	updates := map[string]interface{}{
		"status":     status,
		"progress":   progress,
		"message":    message,
		"updated_at": time.Now(),
	}

	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	return s.db.Model(&models.AsyncTask{}).
		Where("id = ?", taskID).
		Updates(updates).Error
}

// UpdateTaskError 更新任务错误
func (s *TaskService) UpdateTaskError(taskID string, err error) error {
	now := time.Now()
	return s.db.Model(&models.AsyncTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":       "failed",
			"error":        err.Error(),
			"progress":     0,
			"completed_at": &now,
			"updated_at":   time.Now(),
		}).Error
}

// UpdateTaskResult 更新任务结果
func (s *TaskService) UpdateTaskResult(taskID string, result interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	now := time.Now()
	return s.db.Model(&models.AsyncTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":       "completed",
			"progress":     100,
			"result":       string(resultJSON),
			"completed_at": &now,
			"updated_at":   time.Now(),
		}).Error
}

// GetTask 获取任务信息
func (s *TaskService) GetTask(taskID string, deviceIDs ...string) (*models.AsyncTask, error) {
	deviceID := firstTaskDeviceID(deviceIDs)
	var task models.AsyncTask
	query := s.db.Where("id = ?", taskID)
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if err := query.First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTasksByResource 获取资源相关的所有任务
func (s *TaskService) GetTasksByResource(resourceID string, deviceIDs ...string) ([]*models.AsyncTask, error) {
	deviceID := firstTaskDeviceID(deviceIDs)
	var tasks []*models.AsyncTask
	query := s.db.Where("resource_id = ?", resourceID)
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if err := query.
		Order("created_at DESC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}
