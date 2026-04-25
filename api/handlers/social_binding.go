package handlers

import (
	"errors"
	"strings"

	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SocialBindingHandler struct {
	db  *gorm.DB
	log *logger.Logger
}

type upsertSocialBindingRequest struct {
	AccountIdentifier string `json:"account_identifier" binding:"required,max=120"`
	DisplayName       string `json:"display_name" binding:"omitempty,max=120"`
}

func NewSocialBindingHandler(db *gorm.DB, log *logger.Logger) *SocialBindingHandler {
	return &SocialBindingHandler{
		db:  db,
		log: log,
	}
}

func (h *SocialBindingHandler) ListBindings(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	var bindings []models.SocialAccountBinding
	if err := h.db.
		Where("device_id = ?", deviceID).
		Order("platform ASC").
		Find(&bindings).Error; err != nil {
		h.log.Errorw("Failed to list social bindings", "error", err, "device_id", deviceID)
		response.InternalError(c, "获取绑定信息失败")
		return
	}

	response.Success(c, gin.H{"bindings": bindings})
}

func (h *SocialBindingHandler) UpsertBinding(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	platform := normalizeSocialBindingPlatform(c.Param("platform"))
	if platform == "" {
		response.BadRequest(c, "不支持的平台")
		return
	}

	var req upsertSocialBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的绑定信息")
		return
	}

	accountIdentifier := strings.TrimSpace(req.AccountIdentifier)
	if accountIdentifier == "" {
		response.BadRequest(c, "请输入账号标识")
		return
	}

	var displayName *string
	if trimmed := strings.TrimSpace(req.DisplayName); trimmed != "" {
		displayName = &trimmed
	}

	var binding models.SocialAccountBinding
	err := h.db.Where("device_id = ? AND platform = ?", deviceID, platform).First(&binding).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		binding = models.SocialAccountBinding{
			DeviceID:          deviceID,
			Platform:          platform,
			AccountIdentifier: accountIdentifier,
			DisplayName:       displayName,
		}
		if err := h.db.Create(&binding).Error; err != nil {
			h.log.Errorw("Failed to create social binding", "error", err, "device_id", deviceID, "platform", platform)
			response.InternalError(c, "绑定失败")
			return
		}
	case err != nil:
		h.log.Errorw("Failed to query social binding", "error", err, "device_id", deviceID, "platform", platform)
		response.InternalError(c, "绑定失败")
		return
	default:
		binding.AccountIdentifier = accountIdentifier
		binding.DisplayName = displayName
		if err := h.db.Save(&binding).Error; err != nil {
			h.log.Errorw("Failed to update social binding", "error", err, "device_id", deviceID, "platform", platform)
			response.InternalError(c, "绑定失败")
			return
		}
	}

	response.Success(c, gin.H{
		"message": "绑定成功",
		"binding": binding,
	})
}

func normalizeSocialBindingPlatform(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(models.VideoDistributionPlatformDiscord):
		return string(models.VideoDistributionPlatformDiscord)
	case string(models.VideoDistributionPlatformReddit):
		return string(models.VideoDistributionPlatformReddit)
	case string(models.VideoDistributionPlatformPinterest):
		return string(models.VideoDistributionPlatformPinterest)
	default:
		return ""
	}
}
