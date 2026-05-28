/**
 * 模块说明：数字丝路分发接口控制器。
 * 业务场景：前端需要管理社媒授权目标，并把生成的图文/视频提交到多平台异步分发。
 * 核心职责：按设备身份隔离目标与任务，绑定请求参数，并把业务错误转成前端可理解的响应。
 */
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

/**
 * 功能：查询当前设备的分发账号和目标。
 * 参数：c 为 Gin 请求上下文，设备 ID 从中间件读取。
 * 返回：targets 包含 Upload-Post profile、Pinterest board、Reddit subreddit 和 Discord webhook。
 */
func (h *DistributionHandler) ListTargets(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	targets, err := h.service.ListTargets(deviceID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"targets": targets})
}

/**
 * 功能：创建或确认 Upload-Post profile。
 * 参数：c 为 Gin 请求上下文。
 * 返回：profile；Pinterest/Reddit 授权链接依赖它建立外部账号归属。
 */
func (h *DistributionHandler) EnsureUploadPostProfile(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	profile, err := h.service.EnsureUploadPostProfile(c.Request.Context(), deviceID)
	if err != nil {
		h.respondError(c, err)
		return
	}

	response.Success(c, gin.H{"profile": profile})
}

/**
 * 功能：生成第三方平台授权连接链接。
 * 参数：c 为 Gin 请求上下文。
 * 返回：profile 与 access_url，前端打开 access_url 完成 Pinterest/Reddit 授权。
 */
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

/**
 * 功能：创建一次异步分发任务。
 * 参数：c 为 Gin 请求上下文，Body 包含内容、媒体、平台、目标和发布时间。
 * 返回：job，包含每个平台结果的初始状态。
 */
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

/**
 * 功能：统一分发业务错误响应。
 * 参数：c 为 Gin 请求上下文；err 为服务层错误。
 * 返回：无返回值；可修复的配置/参数问题返回 400，未知错误返回 500。
 */
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
