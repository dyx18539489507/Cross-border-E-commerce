package handlers

import (
	"errors"
	"strconv"
	"strings"

	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	services2 "github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DistributionHandler struct {
	service *services2.DistributionService
	log     *logger.Logger
}

func NewDistributionHandler(service *services2.DistributionService, log *logger.Logger) *DistributionHandler {
	return &DistributionHandler{
		service: service,
		log:     log,
	}
}

func (h *DistributionHandler) ListTargets(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	targets, err := h.service.ListTargets(deviceID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"targets": targets})
}

func (h *DistributionHandler) EnsureUploadPostProfile(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	profile, err := h.service.EnsureUploadPostProfile(c.Request.Context(), deviceID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"profile": profile})
}

func (h *DistributionHandler) GenerateUploadPostConnectLink(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	profile, accessURL, err := h.service.GenerateUploadPostConnectLink(c.Request.Context(), deviceID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{
		"profile":    profile,
		"access_url": accessURL,
	})
}

func (h *DistributionHandler) SyncUploadPostProfile(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	profile, err := h.service.SyncUploadPostProfile(c.Request.Context(), deviceID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"profile": profile})
}

func (h *DistributionHandler) ListPinterestBoards(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	boards, err := h.service.ListPinterestBoards(c.Request.Context(), deviceID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"boards": boards})
}

func (h *DistributionHandler) SetDefaultTarget(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 target ID")
		return
	}

	target, err := h.service.SetDefaultTarget(deviceID, uint(targetID))
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"target": target})
}

func (h *DistributionHandler) SaveRedditDefaultTarget(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	var req services2.UpsertRedditTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的 subreddit 配置")
		return
	}

	target, err := h.service.SaveRedditDefaultTarget(deviceID, req)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"target": target})
}

func (h *DistributionHandler) UpsertDiscordTarget(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	var req services2.UpsertDiscordTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的 Discord 配置")
		return
	}

	target, err := h.service.UpsertDiscordTarget(c.Request.Context(), deviceID, req)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"target": target})
}

func (h *DistributionHandler) DeleteTarget(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 target ID")
		return
	}

	if err := h.service.DeleteTarget(deviceID, uint(targetID)); err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "target 已删除"})
}

func (h *DistributionHandler) CreateDistribution(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	var req services2.CreateDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的分发请求")
		return
	}

	job, err := h.service.CreateDistribution(c.Request.Context(), deviceID, &req)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"job": job})
}

func (h *DistributionHandler) ListDistributionJobs(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	jobs, total, err := h.service.ListDistributionJobs(deviceID, page, pageSize)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{
		"jobs":      jobs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *DistributionHandler) GetDistributionJob(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	jobID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 job ID")
		return
	}

	job, err := h.service.GetDistributionJob(deviceID, uint(jobID))
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"job": job})
}

func (h *DistributionHandler) RetryDistribution(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	jobID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 job ID")
		return
	}

	job, err := h.service.RetryDistribution(c.Request.Context(), deviceID, uint(jobID))
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"job": job})
}

func (h *DistributionHandler) respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.NotFound(c, "资源不存在")
	case isDistributionBadRequest(err):
		response.BadRequest(c, err.Error())
	default:
		h.log.Errorw("Distribution request failed", "error", err)
		response.InternalError(c, err.Error())
	}
}

func isDistributionBadRequest(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	for _, marker := range []string{
		"不支持",
		"不能为空",
		"无效",
		"请先",
		"需要",
		"暂不支持",
	} {
		if strings.Contains(message, marker) {
			return true
		}
	}

	return false
}
