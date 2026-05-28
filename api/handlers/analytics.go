package handlers

import (
	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	services2 "github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service *services2.AnalyticsService
	log     *logger.Logger
}

func NewAnalyticsHandler(service *services2.AnalyticsService, log *logger.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: service, log: log}
}

func (h *AnalyticsHandler) Summary(c *gin.Context) {
	summary, err := h.service.Summary(middlewares2.GetDeviceID(c))
	if err != nil {
		if h.log != nil {
			h.log.Warnw("failed to get analytics summary", "error", err)
		}
		response.InternalError(c, "获取数据分析失败")
		return
	}

	response.Success(c, summary)
}
