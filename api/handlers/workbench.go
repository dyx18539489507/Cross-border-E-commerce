package handlers

import (
	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type WorkbenchHandler struct {
	service *services.WorkbenchService
	log     *logger.Logger
}

func NewWorkbenchHandler(service *services.WorkbenchService, log *logger.Logger) *WorkbenchHandler {
	return &WorkbenchHandler{
		service: service,
		log:     log,
	}
}

func (h *WorkbenchHandler) Summary(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	summary, err := h.service.Summary(deviceID)
	if err != nil {
		if h.log != nil {
			h.log.Warnw("failed to get workbench summary", "error", err, "device_id", deviceID)
		}
		response.InternalError(c, "获取工作台数据失败")
		return
	}

	response.Success(c, summary)
}
