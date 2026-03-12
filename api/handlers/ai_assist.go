package handlers

import (
	"errors"
	"strings"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AIAssistHandler struct {
	assistService *services.AIAssistService
	log           *logger.Logger
}

func NewAIAssistHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AIAssistHandler {
	return &AIAssistHandler{
		assistService: services.NewAIAssistService(db, cfg, log),
		log:           log,
	}
}

func (h *AIAssistHandler) GenerateEpisodeScript(c *gin.Context) {
	var req services.GenerateAssistScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.assistService.GenerateEpisodeScript(&req)
	if err != nil {
		h.log.Errorw("AI assist script generation failed", "error", err, "drama_id", req.DramaID)
		switch {
		case errors.Is(err, services.ErrAssistDramaNotFound):
			response.NotFound(c, "项目不存在")
		case errors.Is(err, services.ErrAssistDeepSeekUnavailable):
			response.BadRequest(c, "DeepSeek v3.2 不可用，请在“设置 > AI服务配置”中启用 DeepSeek 模型/endpoint，或配置 COMPLIANCE_API_KEY（兼容 DEEPSEEK_API_KEY）")
		case strings.Contains(err.Error(), "prompt is required"):
			response.BadRequest(c, "请输入AI创作需求")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, result)
}
