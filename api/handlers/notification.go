package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	service *services.NotificationService
	log     *logger.Logger
}

func NewNotificationHandler(service *services.NotificationService, log *logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		log:     log,
	}
}

func (h *NotificationHandler) List(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	notifications, err := h.service.List(deviceID, limit)
	if err != nil {
		response.InternalError(c, "获取通知失败")
		return
	}

	unreadCount, err := h.service.UnreadCount(deviceID)
	if err != nil {
		response.InternalError(c, "获取未读数量失败")
		return
	}

	response.Success(c, gin.H{
		"items":       notifications,
		"unreadCount": unreadCount,
	})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	count, err := h.service.UnreadCount(deviceID)
	if err != nil {
		response.InternalError(c, "获取未读数量失败")
		return
	}

	response.Success(c, gin.H{"unreadCount": count})
}

func (h *NotificationHandler) Create(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	var input services.CreateNotificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "无效的通知参数")
		return
	}

	notification, err := h.service.Create(deviceID, input)
	if err != nil {
		response.InternalError(c, "创建通知失败")
		return
	}

	response.Created(c, notification)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	id, err := parseNotificationID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的通知 ID")
		return
	}

	notification, err := h.service.MarkRead(deviceID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "通知不存在")
			return
		}
		response.InternalError(c, "更新通知失败")
		return
	}

	response.Success(c, notification)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	if err := h.service.MarkAllRead(deviceID); err != nil {
		response.InternalError(c, "更新通知失败")
		return
	}

	response.Success(c, gin.H{"message": "ok"})
}

func (h *NotificationHandler) Dismiss(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	id, err := parseNotificationID(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的通知 ID")
		return
	}

	if err := h.service.Dismiss(deviceID, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "通知不存在")
			return
		}
		response.InternalError(c, "删除通知失败")
		return
	}

	response.Success(c, gin.H{"message": "ok"})
}

func (h *NotificationHandler) Stream(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.InternalError(c, "当前服务不支持实时通知")
		return
	}

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	events, cancel := h.service.Subscribe(deviceID)
	defer cancel()

	count, err := h.service.UnreadCount(deviceID)
	if err == nil {
		_ = writeNotificationSSE(c, flusher, "snapshot", services.NotificationEvent{
			Type:        "snapshot",
			UnreadCount: count,
		})
	}

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeNotificationSSE(c, flusher, "notification", event); err != nil {
				if h.log != nil {
					h.log.Warnw("notification sse write failed", "error", err, "device_id", deviceID)
				}
				return
			}
		case <-ticker.C:
			if _, err := fmt.Fprint(c.Writer, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func parseNotificationID(raw string) (uint, error) {
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func writeNotificationSSE(c *gin.Context, flusher http.Flusher, eventName string, event services.NotificationEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.Writer, "event: %s\n", eventName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}
