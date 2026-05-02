package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SilkroadAgentHandler struct {
	service *services.SilkroadAgentService
	log     *logger.Logger
}

func NewSilkroadAgentHandler(cfg *config.Config, log *logger.Logger) *SilkroadAgentHandler {
	return &SilkroadAgentHandler{
		service: services.NewSilkroadAgentService(cfg, log),
		log:     log,
	}
}

func (h *SilkroadAgentHandler) Generate(c *gin.Context) {
	input, err := bindSilkroadAgentInput(c)
	if err != nil {
		response.BadRequest(c, "无效的 Agent 请求参数")
		return
	}

	result, err := h.service.Generate(input)
	if err != nil {
		if errors.Is(err, services.ErrSilkroadAgentConfigMissing) {
			response.Error(c, http.StatusBadRequest, "GLM_CONFIG_MISSING", "GLM_API_KEY 或模型环境变量未配置，无法调用丝路 Agent。")
			return
		}
		if h.log != nil {
			h.log.Warnw("silkroad agent generation failed", "error", err)
		}
		response.InternalError(c, "丝路 Agent 生成失败，请稍后重试。")
		return
	}

	response.Success(c, result)
}

func (h *SilkroadAgentHandler) Analyze(c *gin.Context) {
	var input services.SilkroadAgentAnalyzeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "无效的 Agent 分析请求参数")
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.InternalError(c, "当前服务不支持流式输出")
		return
	}

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	emit := func(event services.SilkroadAgentAnalyzeEvent) error {
		payload, err := json.Marshal(event.Data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event.Type); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	if err := h.service.StreamMobileTransitionAnalysis(c.Request.Context(), input, emit); err != nil {
		if h.log != nil {
			h.log.Warnw("silkroad agent mobile analysis stream failed", "error", err)
		}
		_ = emit(services.SilkroadAgentAnalyzeEvent{
			Type: "error",
			Data: map[string]string{"message": "网络波动，已切换为本地演示流程"},
		})
	}
}

func bindSilkroadAgentInput(c *gin.Context) (services.SilkroadAgentInput, error) {
	var input services.SilkroadAgentInput
	contentType := c.ContentType()
	if contentType == "multipart/form-data" || contentType == "application/x-www-form-urlencoded" {
		input.ProductName = c.PostForm("productName")
		input.Category = c.PostForm("category")
		input.TargetMarket = c.PostForm("targetMarket")
		input.TargetPlatform = c.PostForm("targetPlatform")
		input.TargetAudience = c.PostForm("targetAudience")
		input.MaterialSpec = c.PostForm("materialSpec")
		input.UsageScenario = c.PostForm("usageScenario")
		input.RawPrompt = c.PostForm("rawPrompt")
		input.ImageDataURL = c.PostForm("imageDataUrl")
		input.CoreSellingPoints = c.PostFormArray("coreSellingPoints")
		if len(input.CoreSellingPoints) == 0 {
			input.CoreSellingPoints = []string{c.PostForm("coreSellingPoints")}
		}
		return input, nil
	}

	err := c.ShouldBindJSON(&input)
	return input, err
}
