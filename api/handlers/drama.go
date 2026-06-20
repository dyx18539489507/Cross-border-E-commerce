/**
 * 模块说明：项目创建与合规预检 HTTP 处理器。
 * 业务场景：数字丝路商品录入完成后，需要通过合规预检 token 约束创建流程；文件中其余旧短剧接口不在本次注释范围。
 * 核心职责：本次注释只覆盖商品创建、合规预检和相关通知，不扩展旧短剧业务说明。
 */
package handlers

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DramaHandler struct {
	db                  *gorm.DB
	dramaService        *services.DramaService
	videoMergeService   *services.VideoMergeService
	notificationService *services.NotificationService
	log                 *logger.Logger
}

func isProjectAPIRequest(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/v1/projects")
}

func marketingProjectResponse(project *models.Drama) gin.H {
	if project == nil {
		return gin.H{}
	}
	targetMarkets := make([]string, 0)
	for _, market := range strings.Split(project.TargetCountry, ",") {
		if market = strings.TrimSpace(market); market != "" {
			targetMarkets = append(targetMarkets, market)
		}
	}
	return gin.H{
		"id":                     project.ID,
		"project_name":           project.Title,
		"product_name":           project.Title,
		"product_description":    project.Description,
		"target_markets":         targetMarkets,
		"material_composition":   project.MaterialComposition,
		"product_selling_points": project.MarketingSellingPoints,
		"marketing_style":        project.Genre,
		"compliance_score":       project.ComplianceScore,
		"compliance_status":      project.ComplianceLevel,
		"compliance_report":      project.ComplianceReport,
		"status":                 project.Status,
		"thumbnail":              project.Thumbnail,
		"metadata":               project.Metadata,
		"content_versions":       project.Episodes,
		"digital_presenters":     project.Characters,
		"marketing_scenes":       project.Scenes,
		"created_at":             project.CreatedAt,
		"updated_at":             project.UpdatedAt,
	}
}

/**
 * 功能：初始化项目处理器并注入合规服务。
 * 参数：db/cfg/log/transferService/notificationService 为数据库、配置、日志、资源转移与通知依赖。
 * 返回：包含数字丝路合规预检能力的 DramaHandler。
 */
func NewDramaHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, transferService *services.ResourceTransferService, notificationService *services.NotificationService) *DramaHandler {
	complianceService := services.NewComplianceService(cfg.Compliance, log)
	return &DramaHandler{
		db:                  db,
		dramaService:        services.NewDramaService(db, log, complianceService),
		videoMergeService:   services.NewVideoMergeService(db, transferService, cfg.Storage.LocalPath, cfg.Storage.BaseURL, log),
		notificationService: notificationService,
		log:                 log,
	}
}

/**
 * 功能：创建数字丝路商品项目。
 * 参数：c 为 Gin 请求上下文，Body 中包含商品录入适配出的创建请求和可选 compliance_token。
 * 返回：创建后的项目与合规结果；红色风险或预检失效时返回业务错误。
 */
func (h *DramaHandler) CreateDrama(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	var req services.CreateDramaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	drama, compliance, err := h.dramaService.CreateDrama(&req, deviceID)
	if err != nil {
		if errors.Is(err, services.ErrTargetCountryRequired) {
			response.BadRequest(c, "请选择目标国家")
			return
		}
		if errors.Is(err, services.ErrCompliancePrecheckInvalid) {
			// 商品内容变化后必须重新预检，避免用户拿旧 token 绕过新的合规风险。
			response.Error(c, 400, "COMPLIANCE_PRECHECK_REQUIRED", "合规预检已失效或内容已变化，请重新校验后再继续")
			return
		}
		if errors.Is(err, services.ErrComplianceRiskForbidden) {
			response.ErrorWithDetails(
				c,
				400,
				"COMPLIANCE_BLOCKED",
				"风险评级为红色（>=80），禁止创建，请先根据整改建议完善商品信息后重试",
				gin.H{"compliance": compliance},
			)
			return
		}
		response.InternalError(c, "创建失败")
		return
	}

	h.notify(deviceID, services.CreateNotificationInput{
		Type:    "product_created",
		Title:   "商品项目已创建",
		Content: "「" + drama.Title + "」已保存并进入营销内容生产流程。",
		Path:    "/projects/" + strconv.FormatUint(uint64(drama.ID), 10),
		Metadata: map[string]interface{}{
			"dramaId": drama.ID,
			"level":   compliance.Level,
			"score":   compliance.Score,
		},
	})

	if isProjectAPIRequest(c) {
		response.Created(c, gin.H{"project": marketingProjectResponse(drama), "compliance": compliance})
		return
	}
	response.Created(c, gin.H{"drama": drama, "compliance": compliance})
}

/**
 * 功能：执行数字丝路商品合规预检。
 * 参数：c 为 Gin 请求上下文，Body 中包含商品标题、描述、目标国家、材质和营销卖点。
 * 返回：合规评分、风险等级、整改建议和短期 compliance_token。
 */
func (h *DramaHandler) CheckCompliance(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	var req services.CreateDramaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	compliance, complianceToken, err := h.dramaService.EvaluateCompliance(&req, deviceID)
	if err != nil {
		if errors.Is(err, services.ErrTargetCountryRequired) {
			response.BadRequest(c, "请选择目标国家")
			return
		}
		response.InternalError(c, "合规校验失败")
		return
	}

	h.notify(deviceID, services.CreateNotificationInput{
		Type:    "compliance_completed",
		Title:   "合规检测完成",
		Content: "商品合规评分 " + strconv.Itoa(compliance.Score) + "，风险等级：" + compliance.LevelLabel + "。",
		Path:    "/compliance",
		Metadata: map[string]interface{}{
			"level": compliance.Level,
			"score": compliance.Score,
		},
	})

	response.Success(c, gin.H{
		"compliance":       compliance,
		"compliance_token": complianceToken,
	})
}

func (h *DramaHandler) GetDrama(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	dramaID := c.Param("id")

	drama, err := h.dramaService.GetDrama(dramaID, deviceID)
	if err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "获取失败")
		return
	}

	if isProjectAPIRequest(c) {
		response.Success(c, marketingProjectResponse(drama))
		return
	}
	response.Success(c, drama)
}

func (h *DramaHandler) ListDramas(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	var query services.DramaListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	dramas, total, err := h.dramaService.ListDramas(&query, deviceID)
	if err != nil {
		response.InternalError(c, "获取列表失败")
		return
	}

	if isProjectAPIRequest(c) {
		projects := make([]gin.H, 0, len(dramas))
		for index := range dramas {
			projects = append(projects, marketingProjectResponse(&dramas[index]))
		}
		response.SuccessWithPagination(c, projects, total, query.Page, query.PageSize)
		return
	}
	response.SuccessWithPagination(c, dramas, total, query.Page, query.PageSize)
}

func (h *DramaHandler) UpdateDrama(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	dramaID := c.Param("id")

	var req services.UpdateDramaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	drama, err := h.dramaService.UpdateDrama(dramaID, &req, deviceID)
	if err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "更新失败")
		return
	}

	if isProjectAPIRequest(c) {
		response.Success(c, marketingProjectResponse(drama))
		return
	}
	response.Success(c, drama)
}

func (h *DramaHandler) DeleteDrama(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	dramaID := c.Param("id")

	if err := h.dramaService.DeleteDrama(dramaID, deviceID); err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "删除失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *DramaHandler) GetDramaStats(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	stats, err := h.dramaService.GetDramaStats(deviceID)
	if err != nil {
		response.InternalError(c, "获取统计失败")
		return
	}

	response.Success(c, stats)
}

func (h *DramaHandler) GetProjectScript(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	projectID := c.Param("id")

	var drama models.Drama
	if err := scopeProjectQuery(h.db.Where("id = ?", projectID), deviceID).
		Preload("Episodes", func(db *gorm.DB) *gorm.DB {
			return db.Order("episode_number ASC")
		}).
		First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "获取营销脚本失败")
		return
	}

	response.Success(c, gin.H{
		"project_id":       drama.ID,
		"project_name":     drama.Title,
		"content_versions": drama.Episodes,
	})
}

// SaveProjectScript stores marketing content versions while the legacy episode table
// remains the persistence layer. The API contract intentionally exposes only project
// and content-version vocabulary to new clients.
func (h *DramaHandler) SaveProjectScript(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	projectID := c.Param("id")
	if _, ok := h.findProject(c, projectID, deviceID); !ok {
		return
	}

	var req struct {
		ContentVersions []struct {
			VersionNumber int     `json:"version_number"`
			Title         string  `json:"title"`
			ScriptContent *string `json:"script_content"`
			Description   *string `json:"description"`
			Duration      int     `json:"duration"`
			Status        string  `json:"status"`
			Shots         []struct {
				ShotNumber         int    `json:"shot_number"`
				Visual             string `json:"visual"`
				SellingPoint       string `json:"selling_point"`
				Voiceover          string `json:"voiceover"`
				Subtitle           string `json:"subtitle"`
				DigitalHumanAction string `json:"digital_human_action"`
				Source             string `json:"source"`
				MarketAdaptation   string `json:"market_adaptation"`
				Duration           int    `json:"duration"`
			} `json:"shots"`
		} `json:"content_versions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	episodes := make([]models.Episode, 0, len(req.ContentVersions))
	for index, version := range req.ContentVersions {
		number := version.VersionNumber
		if number <= 0 {
			number = index + 1
		}
		episodes = append(episodes, models.Episode{
			EpisodeNum:    number,
			Title:         version.Title,
			ScriptContent: version.ScriptContent,
			Description:   version.Description,
			Duration:      version.Duration,
			Status:        version.Status,
		})
	}
	if err := h.dramaService.SaveEpisodes(projectID, &services.SaveEpisodesRequest{Episodes: episodes}, deviceID); err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "保存营销脚本失败")
		return
	}

	for versionIndex, version := range req.ContentVersions {
		versionNumber := version.VersionNumber
		if versionNumber <= 0 {
			versionNumber = versionIndex + 1
		}
		var episode models.Episode
		if err := h.db.Where("drama_id = ? AND episode_number = ?", projectID, versionNumber).First(&episode).Error; err != nil {
			response.InternalError(c, "保存营销分镜失败")
			return
		}
		for shotIndex, shot := range version.Shots {
			shotNumber := shot.ShotNumber
			if shotNumber <= 0 {
				shotNumber = shotIndex + 1
			}
			duration := shot.Duration
			if duration <= 0 {
				duration = 5
			}
			updates := models.Storyboard{
				EpisodeID:        episode.ID,
				StoryboardNumber: shotNumber,
				Title:            stringPointerIfNotBlank(shot.Subtitle),
				Description:      stringPointerIfNotBlank(shot.Visual),
				Result:           stringPointerIfNotBlank(shot.SellingPoint),
				Dialogue:         stringPointerIfNotBlank(shot.Voiceover),
				Action:           stringPointerIfNotBlank(shot.DigitalHumanAction),
				ImagePrompt:      stringPointerIfNotBlank(shot.Source),
				Atmosphere:       stringPointerIfNotBlank(shot.MarketAdaptation),
				Duration:         duration,
				Status:           "pending",
			}
			var existing models.Storyboard
			err := h.db.Where("episode_id = ? AND storyboard_number = ?", episode.ID, shotNumber).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := h.db.Create(&updates).Error; err != nil {
					response.InternalError(c, "保存营销分镜失败")
					return
				}
			} else if err != nil || h.db.Model(&existing).Updates(&updates).Error != nil {
				response.InternalError(c, "保存营销分镜失败")
				return
			}
		}
	}

	var project models.Drama
	if err := scopeProjectQuery(h.db.Where("id = ?", projectID), deviceID).
		Preload("Episodes", func(db *gorm.DB) *gorm.DB { return db.Order("episode_number ASC") }).
		Preload("Episodes.Storyboards", func(db *gorm.DB) *gorm.DB { return db.Order("storyboard_number ASC") }).
		First(&project).Error; err != nil {
		response.InternalError(c, "读取营销脚本失败")
		return
	}
	response.Success(c, gin.H{
		"project_id":       project.ID,
		"project_name":     project.Title,
		"content_versions": project.Episodes,
	})
}

func stringPointerIfNotBlank(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func (h *DramaHandler) GetProjectTimeline(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	projectID := c.Param("id")
	if _, ok := h.findProject(c, projectID, deviceID); !ok {
		return
	}

	var timeline models.Timeline
	err := h.db.
		Preload("Tracks", func(db *gorm.DB) *gorm.DB { return db.Order("`order` ASC") }).
		Preload("Tracks.Clips", func(db *gorm.DB) *gorm.DB { return db.Order("start_time ASC") }).
		Where("drama_id = ?", projectID).
		Order("updated_at DESC").
		First(&timeline).Error
	if err == nil {
		response.Success(c, gin.H{"project_id": projectID, "timeline": timeline})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		response.InternalError(c, "获取时间线失败")
		return
	}

	var drama models.Drama
	if err := scopeProjectQuery(h.db.Where("id = ?", projectID), deviceID).First(&drama).Error; err != nil {
		response.NotFound(c, "营销项目不存在")
		return
	}
	metadata := map[string]interface{}{}
	if len(drama.Metadata) > 0 {
		_ = json.Unmarshal(drama.Metadata, &metadata)
	}
	response.Success(c, gin.H{
		"project_id": projectID,
		"timeline":   metadata["timeline"],
	})
}

func (h *DramaHandler) SaveProjectTimeline(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	projectID := c.Param("id")
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var drama models.Drama
	if err := scopeProjectQuery(h.db.Where("id = ?", projectID), deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "保存时间线失败")
		return
	}

	metadata := map[string]interface{}{}
	if len(drama.Metadata) > 0 {
		_ = json.Unmarshal(drama.Metadata, &metadata)
	}
	metadata["timeline"] = payload
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		response.BadRequest(c, "时间线数据格式错误")
		return
	}
	if err := h.db.Model(&drama).Update("metadata", metadataJSON).Error; err != nil {
		response.InternalError(c, "保存时间线失败")
		return
	}
	response.Success(c, gin.H{"project_id": projectID, "timeline": payload})
}

func (h *DramaHandler) GetProjectAssets(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	projectID := c.Param("id")
	if _, ok := h.findProject(c, projectID, deviceID); !ok {
		return
	}

	var assets []models.Asset
	if err := h.db.Where("drama_id = ?", projectID).Order("created_at DESC").Find(&assets).Error; err != nil {
		response.InternalError(c, "获取项目素材失败")
		return
	}
	response.Success(c, gin.H{"project_id": projectID, "items": assets})
}

func (h *DramaHandler) GetProjectTasks(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	projectID := c.Param("id")
	if _, ok := h.findProject(c, projectID, deviceID); !ok {
		return
	}

	var images []models.ImageGeneration
	var videos []models.VideoGeneration
	var tasks []models.AsyncTask
	_ = h.db.Where("drama_id = ?", projectID).Order("created_at DESC").Find(&images).Error
	_ = h.db.Where("drama_id = ?", projectID).Order("created_at DESC").Find(&videos).Error
	taskQuery := h.db.Where("resource_id = ?", projectID)
	if deviceID != "" {
		taskQuery = taskQuery.Where("device_id = ?", deviceID)
	}
	_ = taskQuery.Order("created_at DESC").Find(&tasks).Error
	response.Success(c, gin.H{
		"project_id": projectID,
		"images":     images,
		"videos":     videos,
		"tasks":      tasks,
	})
}

func (h *DramaHandler) findProject(c *gin.Context, projectID string, deviceID string) (*models.Drama, bool) {
	var drama models.Drama
	if err := scopeProjectQuery(h.db.Where("id = ?", projectID), deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "营销项目不存在")
			return nil, false
		}
		response.InternalError(c, "获取营销项目失败")
		return nil, false
	}
	return &drama, true
}

func scopeProjectQuery(query *gorm.DB, deviceID string) *gorm.DB {
	if deviceID == "" {
		return query
	}
	return query.Where("device_id = ?", deviceID)
}

func (h *DramaHandler) notify(deviceID string, input services.CreateNotificationInput) {
	if h.notificationService == nil {
		return
	}
	if _, err := h.notificationService.Create(deviceID, input); err != nil && h.log != nil {
		h.log.Warnw("failed to create drama notification", "error", err, "device_id", deviceID, "type", input.Type)
	}
}

func (h *DramaHandler) SaveOutline(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	dramaID := c.Param("id")

	var req services.SaveOutlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.dramaService.SaveOutline(dramaID, &req, deviceID); err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "保存失败")
		return
	}

	response.Success(c, gin.H{"message": "保存成功"})
}

func (h *DramaHandler) GetCharacters(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	dramaID := c.Param("id")
	episodeID := c.Query("episode_id") // 可选：如果提供则只返回该章节的角色

	var episodeIDPtr *string
	if episodeID != "" {
		episodeIDPtr = &episodeID
	}

	characters, err := h.dramaService.GetCharacters(dramaID, episodeIDPtr, deviceID)
	if err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		if err.Error() == "episode not found" {
			response.NotFound(c, "营销内容版本不存在")
			return
		}
		response.InternalError(c, "获取数字人形象失败")
		return
	}

	response.Success(c, characters)
}

func (h *DramaHandler) SaveCharacters(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	dramaID := c.Param("id")

	var req services.SaveCharactersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.dramaService.SaveCharacters(dramaID, &req, deviceID); err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "保存失败")
		return
	}

	response.Success(c, gin.H{"message": "保存成功"})
}

func (h *DramaHandler) SaveEpisodes(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	dramaID := c.Param("id")

	var req services.SaveEpisodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.dramaService.SaveEpisodes(dramaID, &req, deviceID); err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "保存失败")
		return
	}

	response.Success(c, gin.H{"message": "保存成功"})
}

func (h *DramaHandler) SaveProgress(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	dramaID := c.Param("id")

	var req services.SaveProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.dramaService.SaveProgress(dramaID, &req, deviceID); err != nil {
		if err.Error() == "drama not found" {
			response.NotFound(c, "营销项目不存在")
			return
		}
		response.InternalError(c, "保存失败")
		return
	}

	response.Success(c, gin.H{"message": "保存成功"})
}

// FinalizeEpisode 完成集数制作（触发视频合成）
func (h *DramaHandler) FinalizeEpisode(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	episodeID := c.Param("episode_id")
	if episodeID == "" {
		response.BadRequest(c, "episode_id不能为空")
		return
	}

	// 尝试读取时间线数据（可选）
	var timelineData *services.FinalizeEpisodeRequest
	if err := c.ShouldBindJSON(&timelineData); err != nil {
		// 如果没有请求体或解析失败，使用nil（将使用默认场景顺序）
		h.log.Warnw("No timeline data provided, will use default scene order", "error", err)
		timelineData = nil
	} else if timelineData != nil {
		h.log.Infow("Received timeline data", "clips_count", len(timelineData.Clips), "episode_id", episodeID)
	}

	// 触发视频合成任务
	result, err := h.videoMergeService.FinalizeEpisode(episodeID, timelineData, deviceID)
	if err != nil {
		h.log.Errorw("Failed to finalize episode", "error", err, "episode_id", episodeID)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// DownloadEpisodeVideo 下载剧集视频
func (h *DramaHandler) DownloadEpisodeVideo(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)

	episodeID := c.Param("episode_id")
	if episodeID == "" {
		response.BadRequest(c, "episode_id不能为空")
		return
	}

	// 查询episode
	var episode models.Episode
	if err := h.db.Preload("Drama").
		Joins("JOIN dramas ON dramas.id = episodes.drama_id").
		Where("episodes.id = ? AND dramas.device_id = ?", episodeID, deviceID).
		First(&episode).Error; err != nil {
		response.NotFound(c, "剧集不存在")
		return
	}

	// 检查是否有视频
	if episode.VideoURL == nil || *episode.VideoURL == "" {
		response.BadRequest(c, "该剧集还没有生成视频")
		return
	}

	// 返回视频URL，让前端重定向下载
	c.JSON(200, gin.H{
		"video_url":      *episode.VideoURL,
		"title":          episode.Title,
		"episode_number": episode.EpisodeNum,
	})
}
