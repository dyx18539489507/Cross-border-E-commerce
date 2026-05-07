package services

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const defaultNotificationLimit = 20

type CreateNotificationInput struct {
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Path     string                 `json:"path"`
	Metadata map[string]interface{} `json:"metadata"`
}

type NotificationEvent struct {
	Type         string               `json:"type"`
	UnreadCount  int64                `json:"unreadCount"`
	Notification *models.Notification `json:"notification,omitempty"`
}

type notificationSubscriber struct {
	id       uint64
	deviceID string
	ch       chan NotificationEvent
}

type NotificationHub struct {
	mu          sync.RWMutex
	nextID      uint64
	subscribers map[string]map[uint64]*notificationSubscriber
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		subscribers: make(map[string]map[uint64]*notificationSubscriber),
	}
}

func (h *NotificationHub) Subscribe(deviceID string) (<-chan NotificationEvent, func()) {
	deviceID = strings.TrimSpace(deviceID)
	ch := make(chan NotificationEvent, 10)

	h.mu.Lock()
	h.nextID++
	id := h.nextID
	if h.subscribers[deviceID] == nil {
		h.subscribers[deviceID] = make(map[uint64]*notificationSubscriber)
	}
	h.subscribers[deviceID][id] = &notificationSubscriber{
		id:       id,
		deviceID: deviceID,
		ch:       ch,
	}
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		if bucket, ok := h.subscribers[deviceID]; ok {
			delete(bucket, id)
			if len(bucket) == 0 {
				delete(h.subscribers, deviceID)
			}
		}
		h.mu.Unlock()
		close(ch)
	}

	return ch, cancel
}

func (h *NotificationHub) Broadcast(deviceID string, event NotificationEvent) {
	deviceID = strings.TrimSpace(deviceID)

	h.mu.RLock()
	bucket := h.subscribers[deviceID]
	targets := make([]*notificationSubscriber, 0, len(bucket))
	for _, subscriber := range bucket {
		targets = append(targets, subscriber)
	}
	h.mu.RUnlock()

	for _, subscriber := range targets {
		select {
		case subscriber.ch <- event:
		default:
		}
	}
}

type NotificationService struct {
	db  *gorm.DB
	log *logger.Logger
	hub *NotificationHub
}

func NewNotificationService(db *gorm.DB, log *logger.Logger) *NotificationService {
	return &NotificationService{
		db:  db,
		log: log,
		hub: NewNotificationHub(),
	}
}

func (s *NotificationService) List(deviceID string, limit int) ([]models.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultNotificationLimit
	}

	var notifications []models.Notification
	err := s.scope(deviceID).
		Where("dismissed_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (s *NotificationService) UnreadCount(deviceID string) (int64, error) {
	var count int64
	err := s.scope(deviceID).
		Where("read_at IS NULL AND dismissed_at IS NULL").
		Count(&count).Error
	return count, err
}

func (s *NotificationService) Create(deviceID string, input CreateNotificationInput) (*models.Notification, error) {
	deviceID = strings.TrimSpace(deviceID)
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" {
		title = "系统通知"
	}
	if content == "" {
		content = "有新的业务动态需要查看。"
	}

	notificationType := strings.TrimSpace(input.Type)
	if notificationType == "" {
		notificationType = "system"
	}

	metadata := datatypes.JSON([]byte("{}"))
	if len(input.Metadata) > 0 {
		if raw, err := json.Marshal(input.Metadata); err == nil {
			metadata = datatypes.JSON(raw)
		}
	}

	notification := &models.Notification{
		DeviceID: deviceID,
		Type:     notificationType,
		Title:    title,
		Content:  content,
		Path:     strings.TrimSpace(input.Path),
		Metadata: metadata,
	}

	if err := s.db.Create(notification).Error; err != nil {
		if s.log != nil {
			s.log.Warnw("failed to create notification", "error", err, "device_id", deviceID, "type", notificationType)
		}
		return nil, err
	}

	s.broadcastWithCount(deviceID, "created", notification)
	return notification, nil
}

func (s *NotificationService) MarkRead(deviceID string, id uint) (*models.Notification, error) {
	now := time.Now()
	var notification models.Notification
	err := s.scope(deviceID).Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}

	if notification.ReadAt == nil {
		notification.ReadAt = &now
		if err := s.db.Save(&notification).Error; err != nil {
			return nil, err
		}
	}

	s.broadcastWithCount(deviceID, "updated", &notification)
	return &notification, nil
}

func (s *NotificationService) MarkAllRead(deviceID string) error {
	now := time.Now()
	if err := s.scope(deviceID).
		Where("read_at IS NULL AND dismissed_at IS NULL").
		Update("read_at", now).Error; err != nil {
		return err
	}

	s.broadcastWithCount(deviceID, "read_all", nil)
	return nil
}

func (s *NotificationService) Dismiss(deviceID string, id uint) error {
	now := time.Now()
	result := s.scope(deviceID).
		Where("id = ?", id).
		Update("dismissed_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	s.broadcastWithCount(deviceID, "dismissed", nil)
	return nil
}

func (s *NotificationService) Subscribe(deviceID string) (<-chan NotificationEvent, func()) {
	return s.hub.Subscribe(deviceID)
}

func (s *NotificationService) scope(deviceID string) *gorm.DB {
	query := s.db.Model(&models.Notification{})
	if strings.TrimSpace(deviceID) != "" {
		query = query.Where("device_id = ?", strings.TrimSpace(deviceID))
	}
	return query
}

func (s *NotificationService) broadcastWithCount(deviceID string, eventType string, notification *models.Notification) {
	count, err := s.UnreadCount(deviceID)
	if err != nil && s.log != nil {
		s.log.Warnw("failed to count unread notifications", "error", err, "device_id", deviceID)
	}
	s.hub.Broadcast(deviceID, NotificationEvent{
		Type:         eventType,
		UnreadCount:  count,
		Notification: notification,
	})
}
