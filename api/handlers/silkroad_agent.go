/**
 * 模块说明：丝路 Agent HTTP 入口。
 * 业务场景：前端把用户自然语言、商品图片、目标市场和追问发送到后端，由 Agent 服务生成跨境营销方案。
 * 核心职责：绑定请求参数、转发给 Agent 服务、输出 SSE 流式事件，并在完成时写入设备通知。
 */
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SilkroadAgentHandler struct {
	service             *services.SilkroadAgentService
	historyService      *services.SilkroadAgentHistoryService
	projectService      *services.SilkroadAgentProjectService
	notificationService *services.NotificationService
	log                 *logger.Logger
}

/**
 * 功能：创建丝路 Agent 控制器。
 * 参数：cfg 为模型与运行配置；log 为日志实例；notificationService 用于写入前端通知中心。
 * 返回：可注册到路由的 SilkroadAgentHandler。
 */
func NewSilkroadAgentHandler(cfg *config.Config, log *logger.Logger, notificationService *services.NotificationService, historyService *services.SilkroadAgentHistoryService, projectService *services.SilkroadAgentProjectService) *SilkroadAgentHandler {
	return &SilkroadAgentHandler{
		service:             services.NewSilkroadAgentService(cfg, log),
		historyService:      historyService,
		projectService:      projectService,
		notificationService: notificationService,
		log:                 log,
	}
}

/**
 * 功能：生成完整的丝路 Agent 方案。
 * 参数：c 为 Gin 请求上下文，承载 JSON 或表单格式的 Agent 输入。
 * 返回：HTTP JSON 响应，成功时包含合规、本地化、脚本、数字人和投放建议。
 */
func (h *SilkroadAgentHandler) Generate(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	input, err := bindSilkroadAgentInput(c)
	if err != nil {
		response.BadRequest(c, "无效的 Agent 请求参数")
		return
	}

	result, err := h.service.Generate(input)
	if err != nil {
		if errors.Is(err, services.ErrSilkroadAgentConfigMissing) {
			// 模型配置缺失属于可修复的业务配置问题，返回 400 让前端提示管理员配置，而不是伪装成系统故障。
			response.Error(c, http.StatusBadRequest, "GLM_CONFIG_MISSING", "GLM_API_KEY 或模型环境变量未配置，无法调用丝路 Agent。")
			return
		}
		if h.log != nil {
			h.log.Warnw("silkroad agent generation failed", "error", err)
		}
		response.InternalError(c, "丝路 Agent 生成失败，请稍后重试。")
		return
	}

	historyItem := h.saveHistory(deviceID, input, result)
	notificationPath := "/agent/result"
	metadata := map[string]interface{}{
		"model":  result.Model,
		"market": result.RecognizedInfo.TargetMarket,
	}
	if historyItem != nil {
		notificationPath = fmt.Sprintf("/agent/result?historyId=%d", historyItem.ID)
		metadata["history_id"] = historyItem.ID
		metadata["request_id"] = historyItem.RequestID
	}

	h.notify(deviceID, services.CreateNotificationInput{
		Type:     "agent_completed",
		Title:    "丝路 Agent 方案已生成",
		Content:  "「" + result.RecognizedInfo.ProductName + "」的出海营销方案已完成。",
		Path:     notificationPath,
		Metadata: metadata,
	})

	response.Success(c, result)
}

/**
 * 功能：执行分阶段多 Agent 工作流并返回 Trace、Critic 评分和最终方案。
 * 参数：c 为 Gin 请求上下文，Body 与原 generate 保持一致。
 * 返回：SilkroadAgentWorkflowResult；同时保存历史记录，便于结果页和论文实验复现阶段输出。
 */
func (h *SilkroadAgentHandler) GenerateWorkflow(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	input, err := bindSilkroadAgentInput(c)
	if err != nil {
		response.BadRequest(c, "无效的 Agent 工作流请求参数")
		return
	}

	workflow, err := h.service.GenerateWithWorkflow(input)
	if err != nil {
		if errors.Is(err, services.ErrSilkroadAgentConfigMissing) {
			response.Error(c, http.StatusBadRequest, "AGENT_CONFIG_MISSING", "Agent 模型环境变量未配置，无法调用多 Agent 工作流。")
			return
		}
		if h.log != nil {
			h.log.Warnw("silkroad workflow generation failed", "error", err)
		}
		response.InternalError(c, "丝路 Agent 工作流生成失败，请稍后重试。")
		return
	}

	historyItem := h.saveWorkflowHistory(deviceID, input, workflow)
	notificationPath := "/agent/result"
	metadata := map[string]interface{}{
		"model":           workflow.Result.Model,
		"market":          workflow.Result.RecognizedInfo.TargetMarket,
		"workflow_status": workflow.WorkflowStatus,
		"revised":         workflow.Revised,
	}
	if historyItem != nil {
		workflow.SessionID = historyItem.ID
		notificationPath = fmt.Sprintf("/agent/result?historyId=%d", historyItem.ID)
		metadata["history_id"] = historyItem.ID
		metadata["request_id"] = historyItem.RequestID
	}

	h.notify(deviceID, services.CreateNotificationInput{
		Type:     "agent_workflow_completed",
		Title:    "多 Agent 营销方案已生成",
		Content:  "「" + workflow.Result.RecognizedInfo.ProductName + "」的跨境营销工作流已完成。",
		Path:     notificationPath,
		Metadata: metadata,
	})

	response.Success(c, workflow)
}

/**
 * 功能：输出过渡页所需的 Agent 流式分析。
 * 参数：c 为 Gin 请求上下文，Body 中包含页面场景、商品字段、目标市场和图片信息。
 * 返回：text/event-stream；事件包括摘要增量、识别信息、任务状态和完成信号。
 */
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
		// 前端按 event 名称分发状态，因此这里统一把业务事件编码成 SSE 块并立即 Flush。
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

/**
 * 功能：处理结果页上的 Agent 追问。
 * 参数：c 为 Gin 请求上下文，包含用户追问和当前方案上下文。
 * 返回：text/event-stream；成功返回 result 事件，失败返回 error 事件供前端展示重试入口。
 */
func (h *SilkroadAgentHandler) FollowUp(c *gin.Context) {
	var input services.SilkroadAgentFollowUpInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "无效的 Agent 追问请求参数")
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

	if err := h.service.StreamFollowUp(c.Request.Context(), input, emit); err != nil {
		if h.log != nil {
			h.log.Warnw("silkroad agent follow-up stream failed", "error", err)
		}
		_ = emit(services.SilkroadAgentAnalyzeEvent{
			Type: "error",
			Data: map[string]string{"message": "Agent 暂时无法继续分析，请稍后重试。"},
		})
	}
}

/**
 * 功能：做一次轻量 Agent 输入识别。
 * 参数：c 为 Gin 请求上下文，支持 JSON、form-data 或表单编码。
 * 返回：归一化后的 Agent 输入，用于前端预填或调试识别结果。
 */
func (h *SilkroadAgentHandler) Extract(c *gin.Context) {
	input, err := bindSilkroadAgentInput(c)
	if err != nil {
		response.BadRequest(c, "无效的 Agent 识别请求参数")
		return
	}

	extracted := h.service.ExtractInput(input)
	response.Success(c, extracted)
}

/**
 * 功能：查询当前浏览器设备的 Agent 历史方案。
 * 参数：c 为 Gin 请求上下文，limit 控制返回条数。
 * 返回：items 列表，包含输入快照、结果快照和目标市场摘要。
 */
func (h *SilkroadAgentHandler) ListHistory(c *gin.Context) {
	if h.historyService == nil {
		response.Success(c, gin.H{"items": []services.SilkroadAgentHistoryItem{}})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, err := h.historyService.List(middlewares2.GetDeviceID(c), limit)
	if err != nil {
		if h.log != nil {
			h.log.Warnw("failed to list silkroad agent history", "error", err)
		}
		response.InternalError(c, "获取 Agent 历史失败")
		return
	}

	response.Success(c, gin.H{"items": items})
}

/**
 * 功能：读取单条 Agent 历史方案。
 * 参数：c 为 Gin 请求上下文，id 为历史记录 ID。
 * 返回：item 包含当时的输入、结果和模型信息，可用于恢复结果页。
 */
func (h *SilkroadAgentHandler) GetHistory(c *gin.Context) {
	if h.historyService == nil {
		response.NotFound(c, "Agent 历史不存在")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 Agent 历史 ID")
		return
	}

	item, err := h.historyService.Get(middlewares2.GetDeviceID(c), uint(id))
	if err != nil {
		response.NotFound(c, "Agent 历史不存在")
		return
	}

	response.Success(c, gin.H{"item": item})
}

/**
 * 功能：从 Agent 历史记录一键创建营销项目。
 * 参数：id 为历史记录 ID；Body 可选传入 result/workflow 作为历史缺失时的兜底。
 * 返回：项目 ID、跳转路径和创建摘要。
 */
func (h *SilkroadAgentHandler) CreateProjectFromHistory(c *gin.Context) {
	if h.projectService == nil {
		response.InternalError(c, "营销项目创建服务未初始化")
		return
	}

	deviceID := middlewares2.GetDeviceID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的 Agent 历史 ID")
		return
	}

	var req services.CreateProjectFromAgentRequest
	_ = c.ShouldBindJSON(&req)

	if h.historyService != nil {
		if item, err := h.historyService.Get(deviceID, uint(id)); err == nil {
			if item.Workflow != nil {
				req.Workflow = item.Workflow
				req.Result = &item.Workflow.Result
			} else if item.Result != nil {
				req.Result = item.Result
			}
		}
	}

	h.createProjectFromAgentPayload(c, deviceID, req)
}

/**
 * 功能：直接使用当前页面提交的 Agent 结果创建营销项目。
 * 参数：Body 包含 result 或 workflow。
 * 返回：项目 ID、跳转路径和创建摘要。
 */
func (h *SilkroadAgentHandler) CreateProject(c *gin.Context) {
	if h.projectService == nil {
		response.InternalError(c, "营销项目创建服务未初始化")
		return
	}

	var req services.CreateProjectFromAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的 Agent 项目创建参数")
		return
	}
	h.createProjectFromAgentPayload(c, middlewares2.GetDeviceID(c), req)
}

func (h *SilkroadAgentHandler) createProjectFromAgentPayload(c *gin.Context, deviceID string, req services.CreateProjectFromAgentRequest) {
	project, err := h.projectService.CreateFromAgentResult(deviceID, req.Result, req.Workflow)
	if err != nil {
		if h.log != nil {
			h.log.Warnw("failed to create project from agent", "error", err, "device_id", deviceID)
		}
		response.BadRequest(c, "无法从当前 Agent 结果创建营销项目")
		return
	}

	h.notify(deviceID, services.CreateNotificationInput{
		Type:    "agent_project_created",
		Title:   "营销项目已创建",
		Content: project.Summary,
		Path:    project.Path,
		Metadata: map[string]interface{}{
			"project_id": project.ProjectID,
			"episode_id": project.EpisodeID,
			"source":     project.CreatedFrom,
		},
	})
	response.Created(c, project)
}

/**
 * 功能：绑定丝路 Agent 输入。
 * 参数：c 为 Gin 请求上下文。
 * 返回：SilkroadAgentInput 和错误；兼容首页 JSON 请求与未来图片上传表单请求。
 */
func bindSilkroadAgentInput(c *gin.Context) (services.SilkroadAgentInput, error) {
	var input services.SilkroadAgentInput
	contentType := c.ContentType()
	if contentType == "multipart/form-data" || contentType == "application/x-www-form-urlencoded" {
		// 表单分支保留给后续真实上传场景；当前前端主要传 JSON 和 imageDataUrl。
		input.ProductName = c.PostForm("productName")
		input.Category = c.PostForm("category")
		input.TargetMarket = c.PostForm("targetMarket")
		input.TargetPlatform = c.PostForm("targetPlatform")
		input.TargetAudience = c.PostForm("targetAudience")
		input.MaterialSpec = c.PostForm("materialSpec")
		input.UsageScenario = c.PostForm("usageScenario")
		input.RawPrompt = c.PostForm("rawPrompt")
		input.ImageDataURL = c.PostForm("imageDataUrl")
		input.RequestID = c.PostForm("requestId")
		input.CoreSellingPoints = c.PostFormArray("coreSellingPoints")
		if len(input.CoreSellingPoints) == 0 {
			input.CoreSellingPoints = []string{c.PostForm("coreSellingPoints")}
		}
		return input, nil
	}

	err := c.ShouldBindJSON(&input)
	return input, err
}

/**
 * 功能：为当前设备写入 Agent 业务通知。
 * 参数：deviceID 为匿名设备身份；input 为通知标题、内容、路径和元数据。
 * 返回：无返回值；通知失败只记录日志，不影响 Agent 主流程返回。
 */
func (h *SilkroadAgentHandler) notify(deviceID string, input services.CreateNotificationInput) {
	if h.notificationService == nil {
		return
	}
	if input.Content == "「」的出海营销方案已完成。" {
		input.Content = "新的出海营销方案已完成。"
	}
	if _, err := h.notificationService.Create(deviceID, input); err != nil && h.log != nil {
		h.log.Warnw("failed to create silkroad agent notification", "error", err, "device_id", deviceID, "type", input.Type)
	}
}

func (h *SilkroadAgentHandler) saveHistory(deviceID string, input services.SilkroadAgentInput, result *services.SilkroadAgentResult) *services.SilkroadAgentHistoryItem {
	if h.historyService == nil {
		return nil
	}
	item, err := h.historyService.Save(deviceID, input, result)
	if err != nil && h.log != nil {
		// 历史记录是体验增强能力，保存失败不应阻断 Agent 主结果返回。
		h.log.Warnw("failed to save silkroad agent history", "error", err, "device_id", deviceID, "request_id", input.RequestID)
	}
	if err != nil {
		return nil
	}
	return item
}

func (h *SilkroadAgentHandler) saveWorkflowHistory(deviceID string, input services.SilkroadAgentInput, workflow *services.SilkroadAgentWorkflowResult) *services.SilkroadAgentHistoryItem {
	if h.historyService == nil {
		return nil
	}
	item, err := h.historyService.SaveWorkflow(deviceID, input, workflow)
	if err != nil && h.log != nil {
		h.log.Warnw("failed to save silkroad workflow history", "error", err, "device_id", deviceID, "request_id", input.RequestID)
	}
	if err != nil {
		return nil
	}
	return item
}
