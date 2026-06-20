/**
 * 模块说明：丝路 Agent 业务编排服务。
 * 业务场景：把跨境商品信息、目标市场、商品图片和用户追问交给 DeepSeek/GLM/视觉模型，生成可落地的出海营销方案。
 * 核心职责：输入抽取、视觉理解、文本方案生成、流式摘要、任务链事件、追问增量更新和本地兜底。
 */
package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
)

var ErrSilkroadAgentConfigMissing = errors.New("silkroad agent model config missing")

type SilkroadAgentInput struct {
	RequestID         string   `json:"requestId" form:"requestId"`
	ProductName       string   `json:"productName" form:"productName"`
	Category          string   `json:"category" form:"category"`
	TargetMarket      string   `json:"targetMarket" form:"targetMarket"`
	TargetPlatform    string   `json:"targetPlatform" form:"targetPlatform"`
	TargetAudience    string   `json:"targetAudience" form:"targetAudience"`
	CoreSellingPoints []string `json:"coreSellingPoints" form:"coreSellingPoints"`
	MaterialSpec      string   `json:"materialSpec" form:"materialSpec"`
	UsageScenario     string   `json:"usageScenario" form:"usageScenario"`
	RawPrompt         string   `json:"rawPrompt" form:"rawPrompt"`
	ImageDataURL      string   `json:"imageDataUrl" form:"imageDataUrl"`
}

type SilkroadAgentAnalyzeInput struct {
	UserInput         string   `json:"userInput"`
	Scene             string   `json:"scene"`
	ProductName       string   `json:"productName"`
	Category          string   `json:"category"`
	TargetMarket      string   `json:"targetMarket"`
	TargetPlatform    string   `json:"targetPlatform"`
	TargetAudience    string   `json:"targetAudience"`
	CoreSellingPoints []string `json:"coreSellingPoints"`
	MaterialSpec      string   `json:"materialSpec"`
	UsageScenario     string   `json:"usageScenario"`
	ImageDataURL      string   `json:"imageDataUrl"`
}

type SilkroadAgentAnalyzeEvent struct {
	Type string
	Data interface{}
}

type SilkroadAgentFollowUpInput struct {
	Question string                       `json:"question"`
	Context  SilkroadAgentFollowUpContext `json:"context"`
}

type SilkroadAgentFollowUpContext struct {
	ProductName        string `json:"productName"`
	Category           string `json:"category"`
	TargetMarket       string `json:"targetMarket"`
	Platform           string `json:"platform"`
	Audience           string `json:"audience"`
	SellingPoints      string `json:"sellingPoints"`
	MaterialSpec       string `json:"materialSpec"`
	UsageScenario      string `json:"usageScenario"`
	ImageUnderstanding string `json:"imageUnderstanding"`
	RawPrompt          string `json:"rawPrompt"`
	ComplianceResult   string `json:"complianceResult"`
	ContentStrategy    string `json:"contentStrategy"`
	DigitalHumanPlan   string `json:"digitalHumanPlan"`
	PromotionAdvice    string `json:"promotionAdvice"`
}

type SilkroadAgentFollowUpResult struct {
	Summary         string                       `json:"summary"`
	Intent          string                       `json:"intent,omitempty"`
	AffectedModules []string                     `json:"affectedModules"`
	UpdatedFields   map[string]interface{}       `json:"updatedFields,omitempty"`
	MissingFields   []string                     `json:"missingFields,omitempty"`
	Cards           []SilkroadAgentFollowUpCard  `json:"cards,omitempty"`
	Details         SilkroadAgentFollowUpDetails `json:"details"`
}

type SilkroadAgentFollowUpCard struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type SilkroadAgentFollowUpDetails struct {
	Compliance      string `json:"compliance"`
	ContentStyle    string `json:"contentStyle"`
	VideoExpression string `json:"videoExpression"`
	Promotion       string `json:"promotion"`
}

type MobileRecognizedInfo struct {
	Product       string `json:"product"`
	Category      string `json:"category"`
	Market        string `json:"market"`
	Platform      string `json:"platform"`
	Audience      string `json:"audience"`
	SellingPoints string `json:"sellingPoints"`
}

type SilkroadAgentResult struct {
	RecognizedInfo RecognizedInfo `json:"recognizedInfo"`
	Overview       AgentOverview  `json:"overview"`
	Compliance     CompliancePlan `json:"compliance"`
	Localization   Localization   `json:"localization"`
	Script         VideoScript    `json:"script"`
	DigitalHuman   DigitalHuman   `json:"digitalHuman"`
	Promotion      PromotionPlan  `json:"promotion"`
	AgentMessage   AgentMessage   `json:"agentMessage"`
	ErrorMessage   string         `json:"errorMessage,omitempty"`
	IsMock         bool           `json:"isMock,omitempty"`
	Model          string         `json:"model,omitempty"`
}

type RecognizedInfo struct {
	ProductName        string   `json:"productName"`
	Category           string   `json:"category"`
	TargetMarket       string   `json:"targetMarket"`
	TargetPlatform     string   `json:"targetPlatform"`
	TargetAudience     string   `json:"targetAudience"`
	CoreSellingPoints  []string `json:"coreSellingPoints"`
	ImageUnderstanding string   `json:"imageUnderstanding"`
}

type AgentOverview struct {
	ComplianceRiskLevel     string `json:"complianceRiskLevel"`
	MarketStrategy          string `json:"marketStrategy"`
	RecommendedVideoStyle   string `json:"recommendedVideoStyle"`
	RecommendedDigitalHuman string `json:"recommendedDigitalHuman"`
}

type CompliancePlan struct {
	Title                string                  `json:"title"`
	Summary              string                  `json:"summary"`
	RiskTags             []string                `json:"riskTags"`
	MissingInfo          []string                `json:"missingInfo"`
	Suggestions          []string                `json:"suggestions"`
	ForbiddenExpressions []string                `json:"forbiddenExpressions"`
	SaferExpressions     []string                `json:"saferExpressions"`
	Level                string                  `json:"level,omitempty"`
	Score                int                     `json:"score,omitempty"`
	RiskReasons          []string                `json:"riskReasons,omitempty"`
	MatchedRules         []models.ComplianceRule `json:"matchedRules,omitempty"`
	Disclaimer           string                  `json:"disclaimer,omitempty"`
}

type Localization struct {
	Direction        string   `json:"direction"`
	Reason           string   `json:"reason"`
	Keywords         []string `json:"keywords"`
	Tone             string   `json:"tone"`
	SceneSuggestions []string `json:"sceneSuggestions"`
}

type VideoScript struct {
	Title      string           `json:"title"`
	Duration   string           `json:"duration"`
	Opening    ScriptSegment    `json:"opening"`
	Middle     ScriptSegment    `json:"middle"`
	Ending     ScriptSegment    `json:"ending"`
	Storyboard []StoryboardShot `json:"storyboard"`
}

type ScriptSegment struct {
	Time    string `json:"time"`
	Content string `json:"content"`
}

type StoryboardShot struct {
	Shot      string `json:"shot"`
	Visual    string `json:"visual"`
	Voiceover string `json:"voiceover"`
	Subtitle  string `json:"subtitle"`
}

type DigitalHuman struct {
	Persona        string `json:"persona"`
	Tone           string `json:"tone"`
	VideoRatio     string `json:"videoRatio"`
	SubtitleAdvice string `json:"subtitleAdvice"`
	VisualStyle    string `json:"visualStyle"`
	ShootingStyle  string `json:"shootingStyle"`
}

type PromotionPlan struct {
	Platforms          []string `json:"platforms"`
	ContentTags        []string `json:"contentTags"`
	FocusMetrics       []string `json:"focusMetrics"`
	OptimizationAdvice string   `json:"optimizationAdvice"`
}

type AgentMessage struct {
	Summary           string   `json:"summary"`
	MissingInfoNotice string   `json:"missingInfoNotice"`
	QuickActions      []string `json:"quickActions"`
}

type SilkroadAgentService struct {
	cfg        *config.Config
	log        *logger.Logger
	httpClient *http.Client
}

func NewSilkroadAgentService(cfg *config.Config, log *logger.Logger) *SilkroadAgentService {
	return &SilkroadAgentService{
		cfg: cfg,
		log: log,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

/**
 * 功能：对外暴露 Agent 输入抽取能力。
 * 参数：input 为用户提交的商品、市场、平台、图片等原始字段。
 * 返回：补齐可推断字段后的 SilkroadAgentInput。
 */
func (s *SilkroadAgentService) ExtractInput(input SilkroadAgentInput) SilkroadAgentInput {
	return extractSilkroadInput(input)
}

/**
 * 功能：生成丝路 Agent 完整结果页方案。
 * 参数：input 为商品、图片、目标市场、平台、人群、卖点和原始描述。
 * 返回：SilkroadAgentResult；配置缺失或模型调用失败时返回错误，开发环境可返回本地兜底结果。
 */
func (s *SilkroadAgentService) Generate(input SilkroadAgentInput) (*SilkroadAgentResult, error) {
	input = extractSilkroadInput(input)
	settings := s.readSettings()
	if settings.APIKey == "" {
		if s.cfg != nil && s.cfg.App.Debug {
			result := buildFallbackAgentResult(input, "开发环境未配置 AGENT_API_KEY/DEEPSEEK_API_KEY，已返回本地兜底结果。")
			result.IsMock = true
			result.Model = settings.TextModel
			return result, nil
		}
		return nil, ErrSilkroadAgentConfigMissing
	}
	if settings.TextModel == "" {
		return nil, fmt.Errorf("%w: AGENT_TEXT_MODEL or DEEPSEEK_MODEL is required", ErrSilkroadAgentConfigMissing)
	}

	imageUnderstanding := ""
	if hasUsableAgentImage(input.ImageDataURL) {
		// 视觉模型只负责理解图片中的外观、包装和潜在风险线索，不直接给出最终合规结论。
		visionSettings := s.readVisionSettings()
		if visionSettings.APIKey == "" || visionSettings.VisionModel == "" {
			imageUnderstanding = "已收到商品图片；当前模型未配置视觉能力，已主要依据文本信息生成方案。"
		} else {
			text, err := s.callVision(visionSettings, visionSettings.VisionModel, input)
			if err != nil {
				if s.log != nil {
					s.log.Warnw("silkroad agent vision analysis failed", "error", err)
				}
				imageUnderstanding = "图片分析暂不可用，已主要依据文本信息生成方案。"
			} else {
				imageUnderstanding = strings.TrimSpace(text)
			}
		}
	}

	raw, err := s.callText(settings, input, imageUnderstanding)
	if err != nil {
		return nil, err
	}

	parsedResult, parseErr := parseAgentResult(raw)
	var result *SilkroadAgentResult
	if parseErr != nil {
		if s.log != nil {
			s.log.Warnw("silkroad agent result parse failed", "error", parseErr)
		}
		result = buildFallbackAgentResult(input, "模型返回内容未能解析为完整 JSON，已使用系统兜底结构。")
		result.ErrorMessage = "模型返回内容未能解析为完整 JSON，已使用系统兜底结构。"
	} else {
		// 文本模型负责把商品、图片理解和目标市场合并为结果页 JSON；缺口字段在 fill 阶段统一兜底。
		result = fillAgentDefaults(parsedResult, input, imageUnderstanding)
	}
	result.Model = settings.TextModel
	return result, nil
}

/**
 * 功能：为过渡页生成流式分析摘要和任务状态事件。
 * 参数：ctx 控制客户端断开；input 为过渡页上下文；emit 用于输出 SSE 事件。
 * 返回：错误；调用方会转换为前端可识别的 error/fallback_notice 事件。
 */
func (s *SilkroadAgentService) StreamMobileTransitionAnalysis(ctx context.Context, input SilkroadAgentAnalyzeInput, emit func(SilkroadAgentAnalyzeEvent) error) error {
	input = normalizeSilkroadAnalyzeInput(input)
	settings := readDeepSeekAnalyzeSettings()
	recognized := buildMobileRecognizedInfo(input)
	imageUnderstanding := s.analyzeTransitionImage(input)

	emitFallback := func(message string) error {
		// 过渡页优先保证用户看到稳定流程，模型不可用时用本地规则输出摘要和任务链，不阻塞最终结果页兜底。
		if message != "" {
			if err := emit(SilkroadAgentAnalyzeEvent{Type: "fallback_notice", Data: map[string]string{"message": message}}); err != nil {
				return err
			}
		}
		for _, text := range fallbackMobileAnalysisParagraphs(recognized) {
			if err := emit(SilkroadAgentAnalyzeEvent{Type: "analysis_summary_delta", Data: map[string]string{"text": text}}); err != nil {
				return err
			}
			if err := waitOrDone(ctx, 260*time.Millisecond); err != nil {
				return err
			}
		}
		return s.emitMobileAnalysisTail(ctx, recognized, emit)
	}

	if settings.APIKey == "" || settings.TextModel == "" {
		return emitFallback("网络波动，已切换为本地演示流程")
	}

	deltaCount := 0
	req := llmChatRequest{
		Model: settings.TextModel,
		Messages: []llmMessage{
			{Role: "system", Content: silkroadMobileTransitionSystemPrompt()},
			{Role: "user", Content: buildMobileTransitionPrompt(input, recognized, imageUnderstanding)},
		},
		Temperature: 0.25,
		MaxTokens:   760,
		Stream:      true,
	}
	err := s.streamChat(ctx, settings, req, func(delta string) error {
		delta = stripReasoningSensitiveText(delta)
		if strings.TrimSpace(delta) == "" {
			return nil
		}
		deltaCount++
		return emit(SilkroadAgentAnalyzeEvent{Type: "analysis_summary_delta", Data: map[string]string{"text": delta}})
	})
	if err != nil {
		if s.log != nil {
			s.log.Warnw("silkroad mobile transition stream failed", "error", err)
		}
		if deltaCount == 0 {
			return emitFallback("网络波动，已切换为本地演示流程")
		}
		if emitErr := emit(SilkroadAgentAnalyzeEvent{Type: "fallback_notice", Data: map[string]string{"message": "网络波动，已继续执行本地任务链"}}); emitErr != nil {
			return emitErr
		}
	}

	return s.emitMobileAnalysisTail(ctx, recognized, emit)
}

/**
 * 功能：根据用户追问生成增量优化建议。
 * 参数：ctx 控制流式请求生命周期；input 包含追问文本和当前方案上下文；emit 输出 result/error。
 * 返回：错误；模型不可用时返回本地追问结果，避免结果页对话中断。
 */
func (s *SilkroadAgentService) StreamFollowUp(ctx context.Context, input SilkroadAgentFollowUpInput, emit func(SilkroadAgentAnalyzeEvent) error) error {
	input = normalizeSilkroadFollowUpInput(input)
	if input.Question == "" {
		return errors.New("follow-up question is empty")
	}

	emitLocalResult := func(reason string) error {
		if s.log != nil && reason != "" {
			s.log.Warnw("silkroad agent follow-up switched to local fallback", "reason", reason)
		}
		return emit(SilkroadAgentAnalyzeEvent{Type: "result", Data: buildLocalFollowUpResult(input)})
	}

	settings := readDeepSeekFollowUpSettings()
	if settings.APIKey == "" {
		return emitLocalResult("DEEPSEEK_API_KEY is empty")
	}

	var raw strings.Builder
	req := llmChatRequest{
		Model: settings.TextModel,
		Messages: []llmMessage{
			{Role: "system", Content: silkroadFollowUpSystemPrompt()},
			{Role: "user", Content: buildFollowUpPrompt(input)},
		},
		Temperature:    0.25,
		MaxTokens:      1300,
		Stream:         true,
		ResponseFormat: &llmResponseFormat{Type: "json_object"},
	}

	if err := s.streamChat(ctx, settings, req, func(delta string) error {
		cleaned := stripReasoningSensitiveText(delta)
		// 追问要求 JSON 结果，服务端先拼完整响应再解析，避免前端拿到半段不可用结构。
		raw.WriteString(cleaned)
		return nil
	}); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return emitLocalResult(err.Error())
	}

	if strings.TrimSpace(raw.String()) == "" {
		return emitLocalResult("empty model response")
	}

	result := parseFollowUpResult(raw.String(), input)
	return emit(SilkroadAgentAnalyzeEvent{Type: "result", Data: result})
}

/**
 * 功能：在过渡页阶段分析商品图片。
 * 参数：input 包含用户上传图片和文本上下文。
 * 返回：图片理解摘要；未配置视觉模型时返回安全兜底说明。
 */
func (s *SilkroadAgentService) analyzeTransitionImage(input SilkroadAgentAnalyzeInput) string {
	if !hasUsableAgentImage(input.ImageDataURL) {
		return ""
	}

	visionSettings := s.readVisionSettings()
	if visionSettings.APIKey == "" || visionSettings.VisionModel == "" {
		return "用户已上传商品图片；当前未配置视觉模型，过渡页先依据文字需求进行分析。"
	}

	text, err := s.callVision(visionSettings, visionSettings.VisionModel, SilkroadAgentInput{
		ProductName:       input.ProductName,
		Category:          input.Category,
		TargetMarket:      input.TargetMarket,
		TargetPlatform:    input.TargetPlatform,
		TargetAudience:    input.TargetAudience,
		CoreSellingPoints: input.CoreSellingPoints,
		MaterialSpec:      input.MaterialSpec,
		UsageScenario:     input.UsageScenario,
		RawPrompt:         input.UserInput,
		ImageDataURL:      input.ImageDataURL,
	})
	if err != nil {
		if s.log != nil {
			s.log.Warnw("silkroad transition image analysis failed", "error", err)
		}
		return "用户已上传商品图片；图片识别暂不可用，过渡页先依据文字需求进行分析。"
	}
	return strings.TrimSpace(text)
}

/**
 * 功能：发送过渡页分析尾部事件。
 * 参数：ctx 控制等待过程；recognized 为已识别商品信息；emit 输出 SSE。
 * 返回：错误；依次输出识别信息、分析完成、六个任务状态和 all_done。
 */
func (s *SilkroadAgentService) emitMobileAnalysisTail(ctx context.Context, recognized MobileRecognizedInfo, emit func(SilkroadAgentAnalyzeEvent) error) error {
	if err := emit(SilkroadAgentAnalyzeEvent{Type: "recognized_info", Data: recognized}); err != nil {
		return err
	}
	if err := waitOrDone(ctx, 260*time.Millisecond); err != nil {
		return err
	}
	if err := emit(SilkroadAgentAnalyzeEvent{Type: "analysis_done", Data: map[string]string{"message": "分析摘要已完成，正在编排 Agent 任务链……"}}); err != nil {
		return err
	}

	tasks := []struct {
		Step        int    `json:"step"`
		Name        string `json:"name"`
		Status      string `json:"status"`
		Description string `json:"description"`
	}{
		// 任务链顺序对应前端展示：先确定商品与合规边界，再进入内容、本地化、数字人和投放。
		{Step: 1, Name: "商品理解", Status: "completed", Description: "确认商品类目、卖点与使用场景"},
		{Step: 2, Name: "合规风险识别", Status: "completed", Description: "匹配目标市场规则与广告敏感表达"},
		{Step: 3, Name: "本地化方向", Status: "completed", Description: taskLocalizationDescription(recognized.Market)},
		{Step: 4, Name: "短视频脚本", Status: "completed", Description: "生成开头、中段、结尾三段式脚本"},
		{Step: 5, Name: "数字人方案", Status: "completed", Description: "推荐数字人形象、口播语气与字幕语言"},
		{Step: 6, Name: "投放优化", Status: "completed", Description: "规划平台、内容方向与关键指标"},
	}
	for _, task := range tasks {
		if err := waitOrDone(ctx, 480*time.Millisecond); err != nil {
			return err
		}
		if err := emit(SilkroadAgentAnalyzeEvent{Type: "task_status", Data: task}); err != nil {
			return err
		}
	}
	return emit(SilkroadAgentAnalyzeEvent{Type: "all_done", Data: map[string]string{"message": "所有任务节点已完成"}})
}

type silkroadAgentSettings struct {
	APIKey      string
	BaseURL     string
	TextModel   string
	VisionModel string
}

func readDeepSeekAnalyzeSettings() silkroadAgentSettings {
	return silkroadAgentSettings{
		APIKey:      firstEnv("DEEPSEEK_API_KEY", "AGENT_API_KEY", "SILKROAD_AGENT_API_KEY"),
		BaseURL:     firstEnvWithDefault("https://api.deepseek.com", "DEEPSEEK_BASE_URL", "AGENT_BASE_URL", "SILKROAD_AGENT_BASE_URL"),
		TextModel:   firstEnvWithDefault("deepseek-v4-flash", "DEEPSEEK_MODEL", "AGENT_ANALYZE_MODEL", "SILKROAD_AGENT_ANALYZE_MODEL"),
		VisionModel: "",
	}
}

func readDeepSeekFollowUpSettings() silkroadAgentSettings {
	return silkroadAgentSettings{
		APIKey:      strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		BaseURL:     "https://api.deepseek.com",
		TextModel:   firstEnvWithDefault("deepseek-v4-flash", "DEEPSEEK_FOLLOW_UP_MODEL", "DEEPSEEK_MODEL"),
		VisionModel: "",
	}
}

/**
 * 功能：读取完整 Agent 方案生成的模型配置。
 * 参数：无；内部读取环境变量和默认 provider。
 * 返回：文本模型、可选视觉模型、BaseURL 和 API Key；DeepSeek 负责文本方案，GLM/方舟可承担视觉理解。
 */
func (s *SilkroadAgentService) readSettings() silkroadAgentSettings {
	apiKey := firstEnv("AGENT_API_KEY", "SILKROAD_AGENT_API_KEY", "DEEPSEEK_API_KEY", "GLM_API_KEY", "ARK_API_KEY")
	provider := strings.ToLower(firstEnv("AGENT_PROVIDER", "SILKROAD_AGENT_PROVIDER"))
	if provider == "" && os.Getenv("DEEPSEEK_API_KEY") != "" {
		provider = "deepseek"
	}
	defaultBaseURL := "https://ark.cn-beijing.volces.com/api/v3"
	if provider == "deepseek" {
		defaultBaseURL = "https://api.deepseek.com"
	}
	textModel := firstEnv("AGENT_TEXT_MODEL", "SILKROAD_AGENT_TEXT_MODEL", "DEEPSEEK_MODEL", "DEEPSEEK_TEXT_MODEL", "GLM_TEXT_MODEL", "ARK_MODEL")
	if textModel == "" && provider == "deepseek" {
		textModel = "deepseek-v4-pro"
	}

	return silkroadAgentSettings{
		APIKey:      apiKey,
		BaseURL:     firstEnvWithDefault(defaultBaseURL, "AGENT_BASE_URL", "SILKROAD_AGENT_BASE_URL", "DEEPSEEK_BASE_URL", "GLM_BASE_URL", "ARK_BASE_URL"),
		TextModel:   textModel,
		VisionModel: firstEnv("AGENT_VISION_MODEL", "SILKROAD_AGENT_VISION_MODEL", "GLM_VISION_MODEL", "ARK_VISION_MODEL"),
	}
}

/**
 * 功能：读取商品图片理解的视觉模型配置。
 * 参数：无；内部按 AGENT/GLM/ARK/DEEPSEEK 变量优先级解析。
 * 返回：视觉模型调用配置；缺失时上层会降级为文本分析。
 */
func (s *SilkroadAgentService) readVisionSettings() silkroadAgentSettings {
	model := firstEnv("AGENT_VISION_MODEL", "SILKROAD_AGENT_VISION_MODEL")
	provider := strings.ToLower(firstEnv("AGENT_VISION_PROVIDER", "SILKROAD_AGENT_VISION_PROVIDER"))

	if model == "" {
		if glmModel := firstEnv("GLM_VISION_MODEL"); glmModel != "" {
			model = glmModel
			if provider == "" {
				provider = "glm"
			}
		} else if arkModel := firstEnv("ARK_VISION_MODEL"); arkModel != "" {
			model = arkModel
			if provider == "" {
				provider = "ark"
			}
		}
	}

	defaultBaseURL := "https://ark.cn-beijing.volces.com/api/v3"
	if provider == "deepseek" {
		defaultBaseURL = "https://api.deepseek.com"
	} else if provider == "glm" {
		defaultBaseURL = firstEnvWithDefault(defaultBaseURL, "GLM_BASE_URL", "ARK_BASE_URL")
	}

	apiKey := firstEnv("AGENT_VISION_API_KEY", "SILKROAD_AGENT_VISION_API_KEY")
	if apiKey == "" {
		switch provider {
		case "deepseek":
			apiKey = firstEnv("DEEPSEEK_API_KEY", "AGENT_API_KEY", "SILKROAD_AGENT_API_KEY")
		case "glm":
			apiKey = firstEnv("GLM_API_KEY", "ARK_API_KEY")
		case "ark":
			apiKey = firstEnv("ARK_API_KEY", "GLM_API_KEY")
		default:
			apiKey = firstEnv("GLM_API_KEY", "ARK_API_KEY", "AGENT_API_KEY", "SILKROAD_AGENT_API_KEY", "DEEPSEEK_API_KEY")
		}
	}

	baseURL := firstEnv("AGENT_VISION_BASE_URL", "SILKROAD_AGENT_VISION_BASE_URL")
	if baseURL == "" {
		switch provider {
		case "deepseek":
			baseURL = firstEnvWithDefault(defaultBaseURL, "DEEPSEEK_BASE_URL", "AGENT_BASE_URL", "SILKROAD_AGENT_BASE_URL")
		case "glm":
			baseURL = firstEnvWithDefault(defaultBaseURL, "GLM_BASE_URL", "ARK_BASE_URL")
		case "ark":
			baseURL = firstEnvWithDefault(defaultBaseURL, "ARK_BASE_URL", "GLM_BASE_URL")
		default:
			baseURL = firstEnvWithDefault(defaultBaseURL, "AGENT_BASE_URL", "SILKROAD_AGENT_BASE_URL", "GLM_BASE_URL", "ARK_BASE_URL", "DEEPSEEK_BASE_URL")
		}
	}

	return silkroadAgentSettings{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		TextModel:   "",
		VisionModel: model,
	}
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func firstEnvWithDefault(fallback string, keys ...string) string {
	if value := firstEnv(keys...); value != "" {
		return value
	}
	return fallback
}

func hasUsableAgentImage(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	lowerValue := strings.ToLower(value)
	return strings.HasPrefix(lowerValue, "data:image/") ||
		strings.HasPrefix(lowerValue, "http://") ||
		strings.HasPrefix(lowerValue, "https://")
}

/**
 * 功能：调用视觉模型理解商品图片。
 * 参数：settings 为视觉模型配置；model 为视觉模型名；input 包含图片 URL/Data URL 和商品上下文。
 * 返回：图片理解文本，包含外观、包装、用途和潜在宣传风险线索。
 */
func (s *SilkroadAgentService) callVision(settings silkroadAgentSettings, model string, input SilkroadAgentInput) (string, error) {
	if !hasUsableAgentImage(input.ImageDataURL) {
		return "", errors.New("vision image is empty")
	}

	userText := "请分析这张跨境电商商品图片，提取商品外观、包装文字、可能用途、材质线索、适合场景和潜在合规/宣传风险。只输出简洁中文段落。"
	req := llmChatRequest{
		Model: model,
		Messages: []llmMessage{
			{Role: "system", Content: "你是谨慎的跨境电商商品图片理解助手。不要做绝对法律结论。"},
			{
				Role: "user",
				Content: []llmContentPart{
					{Type: "text", Text: userText},
					{Type: "image_url", ImageURL: &llmImageURL{URL: input.ImageDataURL}},
				},
			},
		},
		Temperature: 0.2,
		MaxTokens:   900,
	}
	return s.sendChat(settings, req)
}

/**
 * 功能：调用文本模型生成完整 Agent JSON。
 * 参数：settings 为文本模型配置；input 为结构化商品上下文；imageUnderstanding 为视觉模型摘要。
 * 返回：模型原始 JSON 文本。
 */
func (s *SilkroadAgentService) callText(settings silkroadAgentSettings, input SilkroadAgentInput, imageUnderstanding string) (string, error) {
	req := llmChatRequest{
		Model: settings.TextModel,
		Messages: []llmMessage{
			{Role: "system", Content: silkroadAgentSystemPrompt()},
			{Role: "user", Content: buildAgentUserPrompt(input, imageUnderstanding)},
		},
		Temperature:    0.35,
		MaxTokens:      3200,
		ResponseFormat: &llmResponseFormat{Type: "json_object"},
	}
	return s.sendChat(settings, req)
}

type llmChatRequest struct {
	Model          string             `json:"model"`
	Messages       []llmMessage       `json:"messages"`
	Temperature    float64            `json:"temperature,omitempty"`
	MaxTokens      int                `json:"max_tokens,omitempty"`
	Stream         bool               `json:"stream"`
	ResponseFormat *llmResponseFormat `json:"response_format,omitempty"`
}

type llmResponseFormat struct {
	Type string `json:"type"`
}

type llmMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type llmContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *llmImageURL `json:"image_url,omitempty"`
}

type llmImageURL struct {
	URL string `json:"url"`
}

type llmChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (s *SilkroadAgentService) sendChat(settings silkroadAgentSettings, payload llmChatRequest) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal agent model request failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, chatCompletionURL(settings.BaseURL), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create agent model request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+settings.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("agent model request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", fmt.Errorf("read agent model response failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiResp llmChatResponse
		if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Error != nil && apiResp.Error.Message != "" {
			return "", fmt.Errorf("agent model api error: %s", apiResp.Error.Message)
		}
		return "", fmt.Errorf("agent model api error: status %d", resp.StatusCode)
	}

	var chatResp llmChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parse agent model response failed: %w", err)
	}
	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return "", fmt.Errorf("agent model api error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", errors.New("agent model response has no choices")
	}
	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	if content == "" {
		return "", errors.New("agent model response content is empty")
	}
	return content, nil
}

func (s *SilkroadAgentService) streamChat(ctx context.Context, settings silkroadAgentSettings, payload llmChatRequest, onDelta func(string) error) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal agent stream request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatCompletionURL(settings.BaseURL), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create agent stream request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+settings.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("agent stream request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		var apiResp llmChatResponse
		if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Error != nil && apiResp.Error.Message != "" {
			return fmt.Errorf("agent stream api error: %s", apiResp.Error.Message)
		}
		return fmt.Errorf("agent stream api error: status %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("read agent stream failed: %w", err)
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				return nil
			}
			var chunk llmChatStreamChunk
			if unmarshalErr := json.Unmarshal([]byte(data), &chunk); unmarshalErr == nil {
				for _, choice := range chunk.Choices {
					if choice.Delta.Content != "" {
						// 只向前端透出 content，reasoning_content 等内部推理字段不会进入用户可见摘要。
						if deltaErr := onDelta(choice.Delta.Content); deltaErr != nil {
							return deltaErr
						}
					}
				}
			}
		}

		if errors.Is(err, io.EOF) {
			return nil
		}
	}
}

type llmChatStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content,omitempty"`
		} `json:"delta"`
	} `json:"choices"`
}

func chatCompletionURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL
	}
	return baseURL + "/chat/completions"
}

func silkroadAgentSystemPrompt() string {
	return `你是“丝路 Agent”，一个面向跨境电商中小卖家和外贸工厂的 AI 准入合规与智能营销助手。
你需要根据商品信息、目标国家、目标平台和商品图片，生成一份适合展示在网页结果页的出海营销方案。
你必须谨慎处理合规内容，不要给出绝对法律结论，只输出“风险提示、可能限制、建议补充材料、建议表达边界”。
你必须避免夸大功效、绝对化宣传、医疗化表述、虚假认证表述。
如果用户没有明确写出目标国家或目标市场，targetMarket 必须保持为空字符串，并把“目标市场”写入 missingInfo，不要猜测国家。
你必须输出严格 JSON，不要输出 Markdown，不要输出解释性废话。

输出 JSON 结构必须符合：
{
  "recognizedInfo": {
    "productName": "",
    "category": "",
    "targetMarket": "",
    "targetPlatform": "",
    "targetAudience": "",
    "coreSellingPoints": [],
    "imageUnderstanding": ""
  },
  "overview": {
    "complianceRiskLevel": "低风险/中风险/高风险",
    "marketStrategy": "",
    "recommendedVideoStyle": "",
    "recommendedDigitalHuman": ""
  },
  "compliance": {
    "title": "合规分析结果",
    "summary": "",
    "riskTags": [],
    "missingInfo": [],
    "suggestions": [],
    "forbiddenExpressions": [],
    "saferExpressions": []
  },
  "localization": {
    "direction": "",
    "reason": "",
    "keywords": [],
    "tone": "",
    "sceneSuggestions": []
  },
  "script": {
    "title": "短视频脚本",
    "duration": "20-25s",
    "opening": { "time": "0-3s", "content": "" },
    "middle": { "time": "3-20s", "content": "" },
    "ending": { "time": "20-25s", "content": "" },
    "storyboard": [
      { "shot": "", "visual": "", "voiceover": "", "subtitle": "" }
    ]
  },
  "digitalHuman": {
    "persona": "",
    "tone": "",
    "videoRatio": "",
    "subtitleAdvice": "",
    "visualStyle": "",
    "shootingStyle": ""
  },
  "promotion": {
    "platforms": [],
    "contentTags": [],
    "focusMetrics": [],
    "optimizationAdvice": ""
  },
  "agentMessage": {
    "summary": "",
    "missingInfoNotice": "",
    "quickActions": []
  }
}`
}

func silkroadMobileTransitionSystemPrompt() string {
	return `你是「丝路 Agent」，服务于“数字丝路——跨境电商 AI 准入合规与智能营销引擎”。
你的任务是基于用户的自然语言输入，生成面向用户可见的分析摘要，并提取结构化商品信息。
你需要围绕跨境电商出海场景进行分析，重点关注商品理解、商品类目识别、目标市场判断、平台场景识别、目标用户识别、核心卖点提取、合规风险初步判断、本地化内容方向、数字人营销方向和投放优化方向。
如果用户没有明确写出目标国家或目标市场，只说明“目标市场暂未明确”，不要猜测具体国家。
只输出面向用户可见的分析摘要，不要输出原始思维链，不要展示内部推理过程，不要使用“思维链”“chain-of-thought”“reasoning_content”等词。
文案要专业、简洁、适合移动端逐步展示。请输出 3 到 4 个短段落，直接写自然语言，不要 Markdown，不要 JSON，不要列表。`
}

func buildMobileTransitionPrompt(input SilkroadAgentAnalyzeInput, recognized MobileRecognizedInfo, imageUnderstanding string) string {
	payload := map[string]interface{}{
		"用户输入":   input.UserInput,
		"已初步识别":  recognized,
		"页面场景":   input.Scene,
		"图片理解结果": imageUnderstanding,
	}
	jsonPayload, _ := json.MarshalIndent(payload, "", "  ")
	return string(jsonPayload) + "\n\n请生成手机过渡页可见的流式分析摘要，围绕跨境电商、合规、本地化、数字人和短视频投放展开。用户未明确目标市场时，不要自行补全国家。"
}

/**
 * 功能：构造完整方案生成提示词。
 * 参数：input 为用户商品上下文；imageUnderstanding 为视觉模型提取出的商品图片线索。
 * 返回：要求模型输出严格 JSON 的用户提示词，缺失信息会被要求写入 missingInfo。
 */
func buildAgentUserPrompt(input SilkroadAgentInput, imageUnderstanding string) string {
	payload := map[string]interface{}{
		"商品名称":     input.ProductName,
		"商品类目":     input.Category,
		"目标市场/国家":  input.TargetMarket,
		"目标平台":     input.TargetPlatform,
		"目标人群":     input.TargetAudience,
		"核心卖点":     input.CoreSellingPoints,
		"材质/成分/规格": input.MaterialSpec,
		"使用场景":     input.UsageScenario,
		"用户原始描述":   input.RawPrompt,
		"图片理解结果":   imageUnderstanding,
	}
	jsonPayload, _ := json.MarshalIndent(payload, "", "  ")
	return string(jsonPayload) + "\n\n请基于以上信息生成结果页可直接展示的 JSON。缺失信息要写入 missingInfo 和 missingInfoNotice，不要自行编造认证、目标市场或绝对法律结论；如果目标市场/国家为空，recognizedInfo.targetMarket 必须为空字符串。"
}

func silkroadFollowUpSystemPrompt() string {
	return `你是“丝路 Agent”的追问编排器，专注于跨境电商商品理解、准入合规和智能营销方案优化。
你需要像首页 Agent 输入一样理解用户追加的一句话：可能是换商品、换市场、换平台、补充材质/成分/使用场景/图片信息、调整内容语气、重做脚本、询问合规或投放。
先判断用户意图，再基于已有方案做增量更新；不要为了凑模块输出无关内容。
如果用户表达不完整，例如只说“换个产品”但没有新产品名称/类目/卖点，intent 必须为 ask_clarification，并把需要补充的信息写入 missingFields。
不要输出冗长解释，不要输出原始思维链。
必须只输出合法 JSON，不要输出 Markdown，不要输出额外解释。

intent 只能从以下值中选择：
change_product, change_market, change_platform, change_audience, add_product_info, add_material_info, add_usage_scenario, add_image_info, adjust_content_tone, optimize_script, optimize_compliance, optimize_promotion, ask_clarification, general_question

updatedFields 只写用户这次追问中可以明确提取并应用到商品资料的字段。可用字段包括：
productName, category, targetMarket, targetPlatform, targetAudience, coreSellingPoints, materialSpec, usageScenario, marketingGoal, budgetPreference, complianceHints, localizationHints, description

cards 是给前端展示的动态建议模块，数量完全由用户追问和受影响范围决定；不要为了凑数量输出无关内容，也不要因为固定上限删除确实相关的模块。每张卡必须直接服务于本次追问。可用 type 包括：
product, market, platform, audience, selling_point, material, scenario, compliance, localization, script, digital_human, promotion, clarification

JSON 结构必须符合：
{
  "intent": "change_market",
  "summary": "已理解用户本次追问，并基于当前商品方案完成增量调整。",
  "affectedModules": ["商品资料", "合规风险", "本地化内容"],
  "updatedFields": {
    "targetMarket": "印尼"
  },
  "missingFields": [],
  "cards": [
    {
      "type": "localization",
      "title": "本地化调整",
      "content": "说明本次追问导致的内容语气、场景或表达变化。"
    }
  ],
  "details": {
    "compliance": "兼容旧版前端的合规提醒，可与 cards 内容一致或为空。",
    "contentStyle": "兼容旧版前端的内容风格建议，可与 cards 内容一致或为空。",
    "videoExpression": "兼容旧版前端的视频表达建议，可与 cards 内容一致或为空。",
    "promotion": "兼容旧版前端的投放建议，可与 cards 内容一致或为空。"
  }
}`
}

func buildFollowUpPrompt(input SilkroadAgentFollowUpInput) string {
	payload := map[string]interface{}{
		"用户追问": input.Question,
		"当前方案上下文": map[string]string{
			"当前商品名称":  input.Context.ProductName,
			"当前商品类目":  input.Context.Category,
			"当前目标市场":  input.Context.TargetMarket,
			"当前平台":    input.Context.Platform,
			"当前目标用户":  input.Context.Audience,
			"当前卖点":    input.Context.SellingPoints,
			"当前材质成分":  input.Context.MaterialSpec,
			"当前使用场景":  input.Context.UsageScenario,
			"图片理解信息":  input.Context.ImageUnderstanding,
			"原始用户输入":  input.Context.RawPrompt,
			"当前合规结论":  input.Context.ComplianceResult,
			"当前内容策略":  input.Context.ContentStrategy,
			"当前数字人方案": input.Context.DigitalHumanPlan,
			"当前投放建议":  input.Context.PromotionAdvice,
		},
	}
	jsonPayload, _ := json.MarshalIndent(payload, "", "  ")
	return string(jsonPayload) + "\n\n请基于现有方案做增量优化。先判断用户意图，再返回动态 cards 和可沉淀到商品资料的 updatedFields。只返回 JSON，不要重新从零生成完整方案，不要输出思维链。"
}

/**
 * 功能：解析完整 Agent 方案 JSON。
 * 参数：raw 为文本模型返回内容，可能包含额外文本。
 * 返回：SilkroadAgentResult；找不到合法 JSON 时返回错误并交由上层兜底。
 */
func parseAgentResult(raw string) (SilkroadAgentResult, error) {
	var result SilkroadAgentResult
	jsonText := extractFirstJSONObject(raw)
	if jsonText == "" {
		return result, errors.New("no json object found in glm response")
	}
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return result, err
	}
	return result, nil
}

/**
 * 功能：解析追问结果。
 * 参数：raw 为模型返回内容；input 为原始追问，用于解析失败时推断兜底模块。
 * 返回：SilkroadAgentFollowUpResult，保证前端至少拿到 summary、cards 和 details。
 */
func parseFollowUpResult(raw string, input SilkroadAgentFollowUpInput) SilkroadAgentFollowUpResult {
	raw = stripReasoningSensitiveText(raw)
	jsonText := extractFirstJSONObject(raw)
	if jsonText != "" {
		var result SilkroadAgentFollowUpResult
		if err := json.Unmarshal([]byte(jsonText), &result); err == nil {
			return fillFollowUpDefaults(result, raw, input)
		}
	}
	return fillFollowUpDefaults(SilkroadAgentFollowUpResult{Summary: summarizeFollowUpText(raw)}, raw, input)
}

/**
 * 功能：生成本地追问兜底结果。
 * 参数：input 为用户追问和当前方案上下文。
 * 返回：基于关键词推断的增量建议，避免模型不可用时结果页对话彻底失败。
 */
func buildLocalFollowUpResult(input SilkroadAgentFollowUpInput) SilkroadAgentFollowUpResult {
	product := firstNonBlank(input.Context.ProductName, "当前商品")
	intent := inferFollowUpIntent(input.Question)
	platform := firstNonBlank(inferPlatformFromPrompt(input.Question), input.Context.Platform, "TikTok")
	targetMarket := firstNonBlank(inferTargetMarketFromPrompt(input.Question), input.Context.TargetMarket)
	if isUnknownTargetMarket(targetMarket) {
		targetMarket = "目标市场"
	}
	marketPhrase := cleanupTargetMarket(targetMarket)
	if marketPhrase == "" || isUnknownTargetMarket(marketPhrase) {
		marketPhrase = "目标市场"
	}
	missingFields := []string{}
	if intent == "ask_clarification" {
		missingFields = []string{"新商品名称", "商品类目", "核心卖点或使用场景"}
	}
	updatedFields := map[string]interface{}{}
	if inferredMarket := inferTargetMarketFromPrompt(input.Question); inferredMarket != "" {
		updatedFields["targetMarket"] = inferredMarket
	}
	if inferredPlatform := inferPlatformFromPrompt(input.Question); inferredPlatform != "" {
		updatedFields["targetPlatform"] = inferredPlatform
	}
	if material := inferMaterialFromPrompt(input.Question); material != "" {
		updatedFields["materialSpec"] = material
	}
	if scenario := inferUsageScenarioFromPrompt(input.Question); scenario != "" {
		updatedFields["usageScenario"] = scenario
	}
	cards := buildLocalFollowUpCards(intent, product, marketPhrase, platform, missingFields)

	result := SilkroadAgentFollowUpResult{
		Intent:          intent,
		Summary:         buildLocalFollowUpSummary(intent, product, marketPhrase, platform),
		AffectedModules: inferFollowUpModules(input.Question + " " + marketPhrase + " " + platform),
		UpdatedFields:   updatedFields,
		MissingFields:   missingFields,
		Cards:           cards,
		Details: SilkroadAgentFollowUpDetails{
			Compliance:      fmt.Sprintf("切换到%s时需重新核对当地准入、标签和广告表达边界，避免治疗、减肥、绝对化功效等高风险说法。", marketPhrase),
			ContentStyle:    "语气可更年轻、口语化，保留真实体验和生活化表达，减少夸张承诺。",
			VideoExpression: "前 3 秒突出可见卖点和高频场景，中段展示包装、使用或试吃反馈，结尾引导查看详情。",
			Promotion:       fmt.Sprintf("优先在%s测试短视频素材，结合完播率、点击率和评论问题继续迭代内容。", platform),
		},
	}
	return fillFollowUpDefaults(result, input.Question+" "+marketPhrase, input)
}

func inferFollowUpIntent(question string) string {
	text := strings.ToLower(strings.TrimSpace(question))
	if text == "" {
		return "general_question"
	}
	if hasAny(text, []string{"换个产品", "换一个产品", "换商品", "换一个商品", "改个产品", "改一个产品"}) &&
		!hasAny(text, []string{"换成", "改成", "换为", "改为"}) {
		return "ask_clarification"
	}
	if hasAny(text, []string{"换成", "改成", "换为", "改为", "改卖", "换卖"}) &&
		hasAny(text, []string{"产品", "商品", "水杯", "杯", "包", "鞋", "灯", "玩具", "服", "美妆", "食品"}) {
		return "change_product"
	}
	if inferTargetMarketFromPrompt(question) != "" || hasAny(text, []string{"市场", "国家", "地区", "本地化"}) {
		return "change_market"
	}
	if inferPlatformFromPrompt(question) != "" || hasAny(text, []string{"平台", "渠道"}) {
		return "change_platform"
	}
	if hasAny(text, []string{"人群", "用户", "受众", "学生", "妈妈", "女性", "男性", "年轻", "儿童", "亲子"}) {
		return "change_audience"
	}
	if hasAny(text, []string{"材质", "材料", "成分", "规格", "认证"}) {
		return "add_material_info"
	}
	if hasAny(text, []string{"场景", "使用", "户外", "厨房", "通勤", "办公室", "宿舍"}) {
		return "add_usage_scenario"
	}
	if hasAny(text, []string{"图片", "图", "照片", "包装"}) {
		return "add_image_info"
	}
	if hasAny(text, []string{"语气", "年轻", "口语", "风格", "活泼", "高级", "专业"}) {
		return "adjust_content_tone"
	}
	if hasAny(text, []string{"脚本", "分镜", "镜头", "开头", "字幕", "口播", "视频"}) {
		return "optimize_script"
	}
	if hasAny(text, []string{"合规", "风险", "违规", "准入", "禁", "广告法"}) {
		return "optimize_compliance"
	}
	if hasAny(text, []string{"投放", "预算", "点击", "转化", "完播", "推广"}) {
		return "optimize_promotion"
	}
	return "general_question"
}

func inferPlatformFromPrompt(value string) string {
	platforms := []string{"TikTok", "Instagram Reels", "Instagram", "YouTube Shorts", "YouTube", "Shopee", "Lazada", "Amazon", "Temu", "eBay", "Facebook"}
	lowerValue := strings.ToLower(value)
	for _, platform := range platforms {
		if strings.Contains(lowerValue, strings.ToLower(platform)) {
			return platform
		}
	}
	return ""
}

func inferMaterialFromPrompt(value string) string {
	return cleanupExtractedChinese(extractPromptMatch(value, []string{
		`(?:材质|材料|成分|面料)(?:是|为|:|：)?([^，,。；;\n]{1,40})`,
	}))
}

func inferUsageScenarioFromPrompt(value string) string {
	return cleanupExtractedChinese(extractPromptMatch(value, []string{
		`(?:场景|使用场景|适合|用于)(?:是|为|:|：)?([^，,。；;\n]{2,48})`,
	}))
}

func buildLocalFollowUpSummary(intent string, product string, market string, platform string) string {
	switch intent {
	case "ask_clarification":
		return "我还需要新的商品名称、类目或卖点，才能把当前方案切换到新产品。"
	case "change_market":
		return fmt.Sprintf("已将本次追问理解为目标市场调整，后续方案会围绕%s重新校准。", market)
	case "change_platform":
		return fmt.Sprintf("已将本次追问理解为平台调整，内容形式和投放建议会优先适配%s。", platform)
	case "change_product":
		return "已将本次追问理解为商品变更，需要同步重看商品信息、合规边界和内容表达。"
	case "add_material_info":
		return "已记录新的材质/成分信息，后续合规判断会优先基于这部分资料。"
	case "adjust_content_tone":
		return "已将本次追问理解为内容语气调整，会优先更新口播、字幕和场景表达。"
	case "optimize_script":
		return "已将本次追问理解为脚本优化，会聚焦开头钩子、镜头节奏和口播表达。"
	case "optimize_promotion":
		return "已将本次追问理解为投放优化，会聚焦平台组合、预算测试和关键指标。"
	default:
		return fmt.Sprintf("已基于%s和当前方案，整理出适合继续优化的增量建议。", product)
	}
}

func buildLocalFollowUpCards(intent string, product string, market string, platform string, missingFields []string) []SilkroadAgentFollowUpCard {
	if intent == "ask_clarification" {
		return []SilkroadAgentFollowUpCard{{
			Type:    "clarification",
			Title:   "需要补充信息",
			Content: "请补充新商品名称、类目、核心卖点或使用场景，Agent 才能把当前方案切换到新产品。",
		}}
	}
	cards := []SilkroadAgentFollowUpCard{}
	addCard := func(cardType string, title string, content string) {
		cards = append(cards, SilkroadAgentFollowUpCard{Type: cardType, Title: title, Content: content})
	}
	switch intent {
	case "change_market":
		addCard("market", "市场策略更新", fmt.Sprintf("后续商品定位、表达语言和内容场景优先围绕%s重新组织。", market))
		addCard("compliance", "合规重新校准", fmt.Sprintf("进入%s前需重新核对当地准入、标签和广告表达限制，避免绝对化功效承诺。", market))
		addCard("promotion", "投放测试建议", fmt.Sprintf("可先在%s测试本地化短视频素材，再用完播率、点击率和评论问题迭代。", platform))
	case "change_platform":
		addCard("platform", "平台适配", fmt.Sprintf("内容节奏、标题钩子和素材规格优先适配%s的推荐机制。", platform))
		addCard("script", "视频表达调整", "开头 3 秒突出可见卖点，中段用真实场景承接，结尾引导查看详情或评论互动。")
	case "change_product":
		addCard("product", "商品资料重置", "新产品会影响类目、卖点、目标人群和使用场景，建议补齐基础资料后再继续沉淀。")
		addCard("compliance", "合规边界重看", "商品变化后原合规结论不能直接复用，需要重新核对材质、认证和广告敏感表达。")
		addCard("script", "内容方案重写", "脚本钩子、场景和数字人口播都应围绕新产品的可视卖点重新生成。")
	case "add_material_info":
		addCard("material", "材质信息已记录", "材质/成分会直接影响准入、标签、禁限售和广告表达边界。")
		addCard("compliance", "合规判断更新", "后续合规分析会优先核对该材质是否涉及认证、儿童接触、食品接触或敏感功效表达。")
	case "adjust_content_tone":
		addCard("localization", "内容语气更新", "口播和字幕可以更贴近日常表达，减少硬广感和夸张承诺。")
		addCard("script", "脚本表达调整", "开头用生活化痛点承接，中段用具体场景展示卖点，结尾保留轻量行动引导。")
	case "optimize_script":
		addCard("script", "脚本优化重点", "优先优化开头钩子、镜头节奏、口播密度和字幕可读性。")
	case "optimize_compliance":
		addCard("compliance", "合规优化重点", "减少绝对化、医疗化、认证暗示和未经证实的功效承诺。")
	case "optimize_promotion":
		addCard("promotion", "投放优化重点", fmt.Sprintf("围绕%s做小预算 A/B 测试，观察完播率、点击率、收藏率和评论问题。", platform))
	default:
		addCard("product", "增量理解", fmt.Sprintf("本次追问会作为%s的补充上下文，继续影响后续合规、内容和投放方案。", product))
	}
	if len(missingFields) > 0 && len(cards) < 4 {
		addCard("clarification", "仍需补充", fmt.Sprintf("建议继续补充：%s。", strings.Join(missingFields, "、")))
	}
	return cards
}

func hasAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func extractFirstJSONObject(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	start := strings.Index(text, "{")
	if start < 0 {
		return ""
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}

func normalizeSilkroadInput(input SilkroadAgentInput) SilkroadAgentInput {
	input.ProductName = strings.TrimSpace(input.ProductName)
	input.Category = strings.TrimSpace(input.Category)
	input.TargetMarket = strings.TrimSpace(input.TargetMarket)
	input.TargetPlatform = strings.TrimSpace(input.TargetPlatform)
	input.TargetAudience = strings.TrimSpace(input.TargetAudience)
	input.MaterialSpec = strings.TrimSpace(input.MaterialSpec)
	input.UsageScenario = strings.TrimSpace(input.UsageScenario)
	input.RawPrompt = strings.TrimSpace(input.RawPrompt)
	input.ImageDataURL = strings.TrimSpace(input.ImageDataURL)
	input.CoreSellingPoints = cleanStringList(input.CoreSellingPoints)
	if shouldDropImplicitTargetMarket(input.RawPrompt, input.TargetMarket) {
		input.TargetMarket = ""
	}
	if input.ProductName == "" {
		input.ProductName = inferProductFromPrompt(input.RawPrompt)
	}
	if input.Category == "" && input.ProductName != "" {
		input.Category = inferCategory(input.ProductName)
	}
	return input
}

func extractSilkroadInput(input SilkroadAgentInput) SilkroadAgentInput {
	input = normalizeSilkroadInput(input)
	prompt := input.RawPrompt

	if input.ProductName == "" {
		input.ProductName = inferProductFromPrompt(prompt)
	}
	if input.Category == "" {
		input.Category = extractPromptMatch(prompt, []string{
			`(?:商品类目|产品类目|类目|品类|属于)(?:是|为|:|：)?([^，,。；;\n]{2,28})`,
		})
	}
	if input.Category == "" && input.ProductName != "" {
		input.Category = inferCategory(input.ProductName)
	}
	if input.TargetMarket == "" {
		input.TargetMarket = inferTargetMarketFromPrompt(prompt)
	}
	if input.TargetPlatform == "" {
		input.TargetPlatform = extractKnownPlatform(prompt)
	}
	if input.TargetAudience == "" {
		input.TargetAudience = extractPromptMatch(prompt, []string{
			`(?:目标用户|目标人群|受众|面向用户)(?:是|为|:|：)?([^，,。；;\n]{2,44})`,
			`(?:用户是|人群是)([^，,。；;\n]{2,44})`,
		})
	}
	if input.MaterialSpec == "" {
		input.MaterialSpec = extractPromptMatch(prompt, []string{
			`(?:材质|成分|容量|规格|尺寸|型号)(?:是|为|:|：)?([^。；;\n]{2,52})`,
		})
	}
	if input.UsageScenario == "" {
		input.UsageScenario = extractPromptMatch(prompt, []string{
			`(?:使用场景|应用场景|场景)(?:是|为|:|：)?([^。；;\n]{2,64})`,
		})
	}
	if len(input.CoreSellingPoints) == 0 {
		input.CoreSellingPoints = cleanStringList([]string{extractPromptMatch(prompt, []string{
			`(?:核心卖点|卖点|主打|突出)(?:是|为|:|：)?([^。；;\n]{2,80})`,
		})})
	}
	if len(input.CoreSellingPoints) == 0 && input.ProductName != "" {
		input.CoreSellingPoints = inferSellingPoints(input.ProductName, input.Category)
	}

	input.ProductName = cleanupProductName(input.ProductName)
	input.Category = cleanupExtractedChinese(input.Category)
	input.TargetMarket = cleanupExtractedChinese(input.TargetMarket)
	input.TargetPlatform = strings.TrimSpace(input.TargetPlatform)
	input.TargetAudience = cleanupExtractedChinese(input.TargetAudience)
	input.MaterialSpec = cleanupExtractedChinese(input.MaterialSpec)
	input.UsageScenario = cleanupExtractedChinese(input.UsageScenario)
	input.CoreSellingPoints = cleanStringList(input.CoreSellingPoints)
	return input
}

func normalizeSilkroadAnalyzeInput(input SilkroadAgentAnalyzeInput) SilkroadAgentAnalyzeInput {
	input.UserInput = strings.TrimSpace(input.UserInput)
	input.Scene = strings.TrimSpace(input.Scene)
	input.ProductName = strings.TrimSpace(input.ProductName)
	input.Category = strings.TrimSpace(input.Category)
	input.TargetMarket = strings.TrimSpace(input.TargetMarket)
	input.TargetPlatform = strings.TrimSpace(input.TargetPlatform)
	input.TargetAudience = strings.TrimSpace(input.TargetAudience)
	input.MaterialSpec = strings.TrimSpace(input.MaterialSpec)
	input.UsageScenario = strings.TrimSpace(input.UsageScenario)
	input.ImageDataURL = strings.TrimSpace(input.ImageDataURL)
	input.CoreSellingPoints = cleanStringList(input.CoreSellingPoints)
	if shouldDropImplicitTargetMarket(input.UserInput, input.TargetMarket) {
		input.TargetMarket = ""
	}
	if input.ProductName == "" {
		input.ProductName = inferProductFromPrompt(input.UserInput)
	}
	if input.Category == "" && input.ProductName != "" {
		input.Category = inferCategory(input.ProductName)
	}
	if input.UserInput == "" {
		input.UserInput = "请补充商品信息。"
	}
	if input.Scene == "" {
		input.Scene = "mobile_transition"
	}
	return input
}

func normalizeSilkroadFollowUpInput(input SilkroadAgentFollowUpInput) SilkroadAgentFollowUpInput {
	input.Question = strings.TrimSpace(input.Question)
	input.Context.ProductName = strings.TrimSpace(input.Context.ProductName)
	input.Context.TargetMarket = strings.TrimSpace(input.Context.TargetMarket)
	input.Context.Platform = strings.TrimSpace(input.Context.Platform)
	input.Context.Audience = strings.TrimSpace(input.Context.Audience)
	input.Context.SellingPoints = strings.TrimSpace(input.Context.SellingPoints)
	input.Context.ComplianceResult = strings.TrimSpace(input.Context.ComplianceResult)
	input.Context.ContentStrategy = strings.TrimSpace(input.Context.ContentStrategy)
	input.Context.DigitalHumanPlan = strings.TrimSpace(input.Context.DigitalHumanPlan)
	input.Context.PromotionAdvice = strings.TrimSpace(input.Context.PromotionAdvice)
	return input
}

func buildMobileRecognizedInfo(input SilkroadAgentAnalyzeInput) MobileRecognizedInfo {
	prompt := input.UserInput
	product := firstNonEmpty(input.ProductName, inferProductFromPrompt(prompt), "待分析商品")
	category := firstNonEmpty(input.Category, extractPromptMatch(prompt, []string{
		`(?:商品类目|产品类目|类目|品类|属于)(?:是|为|:|：)?([^，,。；;\n]{2,28})`,
	}), inferCategory(product))
	market := firstNonEmpty(input.TargetMarket, inferTargetMarketFromPrompt(prompt), "目标市场待补充")
	platform := firstNonEmpty(input.TargetPlatform, extractKnownPlatform(prompt), "TikTok")
	audience := firstNonEmpty(input.TargetAudience, extractPromptMatch(prompt, []string{
		`(?:目标用户|目标人群|受众|面向用户)(?:是|为|:|：)?([^，,。；;\n]{2,44})`,
		`(?:用户是|人群是)([^，,。；;\n]{2,44})`,
	}), "年轻女性 / 学生 / 办公室")
	points := input.CoreSellingPoints
	if len(points) == 0 {
		points = cleanStringList([]string{extractPromptMatch(prompt, []string{
			`(?:核心卖点|卖点|主打|突出)(?:是|为|:|：)?([^。；;\n]{2,80})`,
		})})
	}
	if len(points) == 0 {
		points = inferSellingPoints(product, category)
	}

	return MobileRecognizedInfo{
		Product:       cleanupExtractedChinese(product),
		Category:      cleanupExtractedChinese(category),
		Market:        cleanupExtractedChinese(market),
		Platform:      strings.TrimSpace(platform),
		Audience:      cleanupExtractedChinese(audience),
		SellingPoints: strings.Join(points, " / "),
	}
}

func fallbackMobileAnalysisParagraphs(info MobileRecognizedInfo) []string {
	riskFocus := "该商品后续合规分析将重点关注目标市场准入要求、必要认证、禁限售规则和广告敏感表达。\n\n"
	if isFoodProduct(info.Product, info.Category) {
		riskFocus = "该商品涉及食品销售场景，后续合规分析将重点关注配料表、过敏原提示、保质期、产地标识、食品准入与平台广告表达边界。\n\n"
	} else if strings.Contains(info.Category, "食品接触") || strings.Contains(info.Product, "杯") {
		riskFocus = "该商品涉及食品接触场景，后续合规分析将重点关注材质说明、食品接触认证、电池参数和充电安全描述。\n\n"
	}

	return []string{
		"已接收到你的出海需求，正在从自然语言中提取商品、目标市场、平台、人群和核心卖点。\n\n",
		fmt.Sprintf("识别到商品为%s，属于%s，%s，主要投放平台为%s。\n\n", info.Product, info.Category, describeTargetMarketForAnalysis(info.Market), info.Platform),
		riskFocus,
		"接下来将基于合规边界，生成本地化营销方向、短视频脚本、数字人方案和投放建议。",
	}
}

func extractPromptMatch(value string, patterns []string) string {
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		match := re.FindStringSubmatch(value)
		if len(match) > 1 && strings.TrimSpace(match[1]) != "" {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func extractKnownPlatform(value string) string {
	platforms := []string{"TikTok", "Instagram Reels", "Instagram", "YouTube Shorts", "YouTube", "Shopee", "Lazada", "Amazon", "Temu", "eBay", "Facebook", "小红书", "抖音"}
	lowerValue := strings.ToLower(value)
	for _, platform := range platforms {
		if strings.Contains(lowerValue, strings.ToLower(platform)) {
			return platform
		}
	}
	return ""
}

func inferCategory(product string) string {
	if strings.Contains(product, "榨汁杯") || strings.Contains(product, "杯") {
		return "小家电 / 食品接触用品"
	}
	if isFoodProduct(product, "") {
		return "食品饮料 / 即食食品"
	}
	return "跨境电商商品"
}

func inferProductFromPrompt(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	product := extractPromptMatch(value, []string{
		`(?:商品准入分析|商品|产品|品名|商品名称|产品名称)\s*[:：]\s*([^，,。；;\n]{1,28})`,
		`^(?:帮我|请|麻烦)?(?:分析一下|分析|看看|测一下|识别一下|识别)?\s*([^，,。；;\n]{1,28}?)(?:卖到|卖|出口到|进入|面向|投放到|推广到|上架|发布到|做|去|给|给到|$)`,
		`(?:我有|我们有|这是一款|这款|一款|一个|一种|商品是|产品是)([^，,。；;\n]{2,28}?)(?:，|,|。|；|;|想|计划|准备|主打|目标|卖到|出口|做|$)`,
		`(?:销售|卖)([^，,。；;\n]{2,28}?)(?:，|,|。|；|;|到|去|$)`,
	})
	if product != "" {
		return cleanupProductName(product)
	}
	for _, line := range strings.Split(value, "\n") {
		line = cleanupProductName(line)
		if line != "" && len([]rune(line)) <= 18 {
			return line
		}
	}
	return ""
}

func inferTargetMarketFromPrompt(value string) string {
	market := extractPromptMatch(value, []string{
		`(?:卖到|卖|出口到|进入|面向|投放到|推广到|去|给到)([^，,。；;\n]{1,24}?)(?:市场|用户|消费者|，|,|。|；|;|$)`,
		`(?:目标市场|目标国家|国家|市场)(?:是|为|:|：)?([^，,。；;\n]{2,24})`,
	})
	if market != "" {
		return cleanupTargetMarket(market)
	}

	markets := []string{
		"美国", "英国", "加拿大", "澳大利亚", "德国", "法国", "意大利", "西班牙", "荷兰", "日本", "韩国",
		"马来西亚", "新加坡", "泰国", "越南", "印尼", "印度尼西亚", "菲律宾", "印度", "墨西哥", "巴西",
		"沙特", "阿联酋", "中东", "欧洲", "东南亚", "北美",
	}
	lowerValue := strings.ToLower(value)
	for _, item := range markets {
		if strings.Contains(lowerValue, strings.ToLower(item)) {
			return item
		}
	}
	return ""
}

func cleanupTargetMarket(value string) string {
	value = cleanupExtractedChinese(value)
	for _, platform := range []string{"TikTok", "Instagram Reels", "Instagram", "YouTube Shorts", "YouTube", "Shopee", "Lazada", "Amazon", "Temu", "eBay", "Facebook", "小红书", "抖音"} {
		value = strings.ReplaceAll(value, platform, "")
		value = strings.ReplaceAll(value, strings.ToLower(platform), "")
	}
	value = strings.TrimSuffix(value, "市场")
	return cleanupExtractedChinese(value)
}

func isUnknownTargetMarket(value string) bool {
	value = cleanupExtractedChinese(value)
	if value == "" {
		return true
	}
	return strings.Contains(value, "待补充") || strings.Contains(value, "待识别") || strings.Contains(value, "未明确")
}

func shouldDropImplicitTargetMarket(prompt string, market string) bool {
	prompt = strings.TrimSpace(prompt)
	market = cleanupTargetMarket(market)
	if prompt == "" || market == "" || isUnknownTargetMarket(market) {
		return false
	}
	return inferTargetMarketFromPrompt(prompt) == "" && !strings.Contains(prompt, market)
}

func describeTargetMarketForAnalysis(market string) string {
	if isUnknownTargetMarket(market) {
		return "目标市场暂未明确"
	}
	return fmt.Sprintf("目标市场为%s", cleanupTargetMarket(market))
}

func taskLocalizationDescription(market string) string {
	if isUnknownTargetMarket(market) {
		return "根据已明确的目标市场生成内容方向"
	}
	return fmt.Sprintf("生成符合%s用户习惯的内容方向", cleanupTargetMarket(market))
}

func inferSellingPoints(product string, category string) []string {
	if isFoodProduct(product, category) {
		return []string{"口味", "便捷", "场景化"}
	}
	if strings.Contains(product, "榨汁杯") || strings.Contains(product, "杯") {
		return []string{"便携", "健康", "易清洗"}
	}
	return []string{"实用", "本地化", "易展示"}
}

func cleanupProductName(value string) string {
	value = cleanupExtractedChinese(value)
	value = strings.TrimPrefix(value, "一下")
	tagPattern := regexp.MustCompile(`(?:生成本地化脚本|数字人口播方案|投放优化建议|商品准入分析)\s*[:：]?`)
	parts := tagPattern.Split(value, 2)
	if len(parts) > 0 {
		value = parts[0]
	}
	value = strings.TrimPrefix(value, "分析")
	value = strings.TrimPrefix(value, "识别")
	value = strings.TrimPrefix(value, "检测")
	return cleanupExtractedChinese(value)
}

func isFoodProduct(product string, category string) bool {
	value := product + " " + category
	foodKeywords := []string{"食品", "饮料", "零食", "即食", "餐", "鸡", "鸭", "肉", "鱼", "虾", "糕", "饼", "糖", "茶", "咖啡", "炸", "烤", "卤", "吃"}
	for _, keyword := range foodKeywords {
		if strings.Contains(value, keyword) {
			return true
		}
	}
	return false
}

func cleanupExtractedChinese(value string) string {
	return strings.TrimSpace(strings.Trim(value, "，,。；; "))
}

func stripReasoningSensitiveText(value string) string {
	blocked := []string{"思维链", "chain-of-thought", "reasoning_content", "内部推理", "隐藏推理"}
	cleaned := value
	for _, word := range blocked {
		cleaned = strings.ReplaceAll(cleaned, word, "")
	}
	return cleaned
}

func fillFollowUpDefaults(result SilkroadAgentFollowUpResult, raw string, input SilkroadAgentFollowUpInput) SilkroadAgentFollowUpResult {
	result.Intent = strings.TrimSpace(result.Intent)
	if result.Intent == "" {
		result.Intent = inferFollowUpIntent(input.Question)
	}
	if strings.TrimSpace(result.Summary) == "" {
		result.Summary = summarizeFollowUpText(raw)
	}
	if strings.TrimSpace(result.Summary) == "" {
		result.Summary = "已基于当前商品和原方案，整理出适合继续优化的补充建议。"
	}
	result.Summary = limitRunes(cleanupFollowUpText(result.Summary), 120)

	result.AffectedModules = cleanStringList(result.AffectedModules)
	if len(result.AffectedModules) == 0 {
		result.AffectedModules = inferFollowUpModules(raw + " " + input.Question)
	}
	if len(result.AffectedModules) > 6 {
		result.AffectedModules = result.AffectedModules[:6]
	}

	result.MissingFields = cleanStringList(result.MissingFields)
	result.UpdatedFields = cleanFollowUpUpdatedFields(result.UpdatedFields)
	result.Cards = cleanFollowUpCards(result.Cards)
	if len(result.Cards) == 0 {
		result.Cards = followUpDetailsToCards(result.Details, result.AffectedModules)
	}
	if len(result.Cards) == 0 && result.Intent == "ask_clarification" {
		missing := result.MissingFields
		if len(missing) == 0 {
			missing = []string{"新商品名称", "商品类目", "核心卖点或使用场景"}
			result.MissingFields = missing
		}
		result.Cards = []SilkroadAgentFollowUpCard{{
			Type:    "clarification",
			Title:   "需要补充信息",
			Content: fmt.Sprintf("请补充%s，Agent 才能继续更新当前方案。", strings.Join(missing, "、")),
		}}
	}

	if strings.TrimSpace(result.Details.Compliance) == "" {
		result.Details.Compliance = "继续避免绝对化、医疗化和未经证实的认证表达，并以目标市场实际准入材料为准。"
	}
	if strings.TrimSpace(result.Details.ContentStyle) == "" {
		result.Details.ContentStyle = "内容语气应贴近目标人群日常表达，保留真实体验感，减少夸张承诺。"
	}
	if strings.TrimSpace(result.Details.VideoExpression) == "" {
		result.Details.VideoExpression = "前 3 秒突出核心场景和可见卖点，中段用真实使用画面承接，不做过度功效演绎。"
	}
	if strings.TrimSpace(result.Details.Promotion) == "" {
		result.Details.Promotion = "优先测试当前主平台短视频内容，结合完播率、点击率和评论问题继续迭代素材。"
	}
	result.Details.Compliance = limitRunes(cleanupFollowUpText(result.Details.Compliance), 96)
	result.Details.ContentStyle = limitRunes(cleanupFollowUpText(result.Details.ContentStyle), 96)
	result.Details.VideoExpression = limitRunes(cleanupFollowUpText(result.Details.VideoExpression), 96)
	result.Details.Promotion = limitRunes(cleanupFollowUpText(result.Details.Promotion), 96)
	return result
}

func cleanFollowUpUpdatedFields(fields map[string]interface{}) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	allowed := map[string]bool{
		"productName":       true,
		"category":          true,
		"targetMarket":      true,
		"targetPlatform":    true,
		"targetAudience":    true,
		"coreSellingPoints": true,
		"materialSpec":      true,
		"usageScenario":     true,
		"marketingGoal":     true,
		"budgetPreference":  true,
		"complianceHints":   true,
		"localizationHints": true,
		"description":       true,
	}
	cleaned := map[string]interface{}{}
	for key, value := range fields {
		if !allowed[key] {
			continue
		}
		switch typed := value.(type) {
		case string:
			if text := cleanupFollowUpText(typed); text != "" {
				cleaned[key] = limitRunes(text, 160)
			}
		case []interface{}:
			items := make([]string, 0, len(typed))
			for _, item := range typed {
				if text := cleanupFollowUpText(fmt.Sprint(item)); text != "" {
					items = append(items, limitRunes(text, 80))
				}
			}
			if len(items) > 0 {
				cleaned[key] = items
			}
		case []string:
			items := cleanStringList(typed)
			if len(items) > 0 {
				cleaned[key] = items
			}
		default:
			if text := cleanupFollowUpText(fmt.Sprint(value)); text != "" {
				cleaned[key] = limitRunes(text, 160)
			}
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

func cleanFollowUpCards(cards []SilkroadAgentFollowUpCard) []SilkroadAgentFollowUpCard {
	if len(cards) == 0 {
		return nil
	}
	allowedTypes := map[string]bool{
		"product": true, "market": true, "platform": true, "audience": true, "selling_point": true,
		"material": true, "scenario": true, "compliance": true, "localization": true, "script": true,
		"digital_human": true, "promotion": true, "clarification": true,
	}
	out := make([]SilkroadAgentFollowUpCard, 0, len(cards))
	seen := map[string]bool{}
	for _, card := range cards {
		cardType := strings.TrimSpace(card.Type)
		if cardType == "" || !allowedTypes[cardType] {
			cardType = "product"
		}
		title := limitRunes(cleanupFollowUpText(card.Title), 20)
		content := limitRunes(cleanupFollowUpText(card.Content), 120)
		if title == "" || content == "" {
			continue
		}
		key := cardType + "|" + title
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, SilkroadAgentFollowUpCard{Type: cardType, Title: title, Content: content})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func followUpDetailsToCards(details SilkroadAgentFollowUpDetails, modules []string) []SilkroadAgentFollowUpCard {
	moduleText := strings.Join(modules, " ")
	candidates := []SilkroadAgentFollowUpCard{}
	if strings.TrimSpace(details.Compliance) != "" && strings.Contains(moduleText, "合规") {
		candidates = append(candidates, SilkroadAgentFollowUpCard{Type: "compliance", Title: "合规提醒", Content: details.Compliance})
	}
	if strings.TrimSpace(details.ContentStyle) != "" && (strings.Contains(moduleText, "市场") || strings.Contains(moduleText, "内容") || strings.Contains(moduleText, "本地化")) {
		candidates = append(candidates, SilkroadAgentFollowUpCard{Type: "localization", Title: "内容调整", Content: details.ContentStyle})
	}
	if strings.TrimSpace(details.VideoExpression) != "" && (strings.Contains(moduleText, "视频") || strings.Contains(moduleText, "脚本") || strings.Contains(moduleText, "表达")) {
		candidates = append(candidates, SilkroadAgentFollowUpCard{Type: "script", Title: "视频表达", Content: details.VideoExpression})
	}
	if strings.TrimSpace(details.Promotion) != "" && strings.Contains(moduleText, "投放") {
		candidates = append(candidates, SilkroadAgentFollowUpCard{Type: "promotion", Title: "投放建议", Content: details.Promotion})
	}
	if len(candidates) == 0 {
		if strings.TrimSpace(details.ContentStyle) != "" {
			candidates = append(candidates, SilkroadAgentFollowUpCard{Type: "localization", Title: "内容调整", Content: details.ContentStyle})
		}
		if strings.TrimSpace(details.Compliance) != "" {
			candidates = append(candidates, SilkroadAgentFollowUpCard{Type: "compliance", Title: "合规提醒", Content: details.Compliance})
		}
	}
	return cleanFollowUpCards(candidates)
}

func inferFollowUpModules(text string) []string {
	type moduleRule struct {
		name     string
		keywords []string
	}
	rules := []moduleRule{
		{name: "市场策略", keywords: []string{"市场", "国家", "印尼", "印度尼西亚", "马来西亚", "东南亚", "中东", "目标用户"}},
		{name: "内容风格", keywords: []string{"语气", "年轻", "口语", "风格", "内容", "本地化", "Z 世代", "Z世代"}},
		{name: "投放建议", keywords: []string{"投放", "平台", "TikTok", "Instagram", "Reels", "完播率", "点击率"}},
		{name: "合规风险", keywords: []string{"合规", "风险", "认证", "准入", "减肥", "治疗", "绝对化", "广告"}},
	}
	modules := make([]string, 0, len(rules))
	for _, rule := range rules {
		for _, keyword := range rule.keywords {
			if strings.Contains(strings.ToLower(text), strings.ToLower(keyword)) {
				modules = append(modules, rule.name)
				break
			}
		}
	}
	if len(modules) == 0 {
		return []string{"市场策略", "内容风格", "投放建议", "合规风险"}
	}
	return modules
}

func summarizeFollowUpText(raw string) string {
	text := cleanupFollowUpText(raw)
	if text == "" {
		return ""
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '。' || r == '！' || r == '!' || r == '\n'
	})
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			return limitRunes(part, 120)
		}
	}
	return limitRunes(text, 120)
}

func cleanupFollowUpText(value string) string {
	value = stripReasoningSensitiveText(value)
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.Join(strings.Fields(value), " ")
	return value
}

func limitRunes(value string, max int) string {
	runes := []rune(strings.TrimSpace(value))
	if max <= 0 || len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max]) + "…"
}

func waitOrDone(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func cleanStringList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n'
		}) {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func fillAgentDefaults(result SilkroadAgentResult, input SilkroadAgentInput, imageUnderstanding string) *SilkroadAgentResult {
	fallback := buildFallbackAgentResult(input, "")
	if result.RecognizedInfo.ProductName == "" {
		result.RecognizedInfo.ProductName = fallback.RecognizedInfo.ProductName
	}
	if result.RecognizedInfo.Category == "" {
		result.RecognizedInfo.Category = fallback.RecognizedInfo.Category
	}
	if input.TargetMarket == "" && inferTargetMarketFromPrompt(input.RawPrompt) == "" {
		result.RecognizedInfo.TargetMarket = fallback.RecognizedInfo.TargetMarket
	} else if result.RecognizedInfo.TargetMarket == "" {
		result.RecognizedInfo.TargetMarket = fallback.RecognizedInfo.TargetMarket
	}
	if result.RecognizedInfo.TargetPlatform == "" {
		result.RecognizedInfo.TargetPlatform = fallback.RecognizedInfo.TargetPlatform
	}
	if result.RecognizedInfo.TargetAudience == "" {
		result.RecognizedInfo.TargetAudience = fallback.RecognizedInfo.TargetAudience
	}
	if len(result.RecognizedInfo.CoreSellingPoints) == 0 {
		result.RecognizedInfo.CoreSellingPoints = fallback.RecognizedInfo.CoreSellingPoints
	}
	if result.RecognizedInfo.ImageUnderstanding == "" {
		if imageUnderstanding != "" {
			result.RecognizedInfo.ImageUnderstanding = imageUnderstanding
		} else {
			result.RecognizedInfo.ImageUnderstanding = fallback.RecognizedInfo.ImageUnderstanding
		}
	}
	if result.Overview.ComplianceRiskLevel == "" {
		result.Overview.ComplianceRiskLevel = fallback.Overview.ComplianceRiskLevel
	}
	if result.Overview.MarketStrategy == "" {
		result.Overview.MarketStrategy = fallback.Overview.MarketStrategy
	}
	if result.Overview.RecommendedVideoStyle == "" {
		result.Overview.RecommendedVideoStyle = fallback.Overview.RecommendedVideoStyle
	}
	if result.Overview.RecommendedDigitalHuman == "" {
		result.Overview.RecommendedDigitalHuman = fallback.Overview.RecommendedDigitalHuman
	}
	if result.Compliance.Title == "" {
		result.Compliance.Title = "合规分析结果"
	}
	if result.Compliance.Summary == "" {
		result.Compliance.Summary = fallback.Compliance.Summary
	}
	if len(result.Compliance.RiskTags) == 0 {
		result.Compliance.RiskTags = fallback.Compliance.RiskTags
	}
	if len(result.Compliance.MissingInfo) == 0 {
		result.Compliance.MissingInfo = fallback.Compliance.MissingInfo
	}
	if len(result.Compliance.Suggestions) == 0 {
		result.Compliance.Suggestions = fallback.Compliance.Suggestions
	}
	if len(result.Compliance.ForbiddenExpressions) == 0 {
		result.Compliance.ForbiddenExpressions = fallback.Compliance.ForbiddenExpressions
	}
	if len(result.Compliance.SaferExpressions) == 0 {
		result.Compliance.SaferExpressions = fallback.Compliance.SaferExpressions
	}
	if result.Localization.Direction == "" {
		result.Localization.Direction = fallback.Localization.Direction
	}
	if result.Localization.Reason == "" {
		result.Localization.Reason = fallback.Localization.Reason
	}
	if len(result.Localization.Keywords) == 0 {
		result.Localization.Keywords = fallback.Localization.Keywords
	}
	if result.Localization.Tone == "" {
		result.Localization.Tone = fallback.Localization.Tone
	}
	if len(result.Localization.SceneSuggestions) == 0 {
		result.Localization.SceneSuggestions = fallback.Localization.SceneSuggestions
	}
	if result.Script.Title == "" {
		result.Script.Title = "短视频脚本"
	}
	if result.Script.Duration == "" {
		result.Script.Duration = "20-25s"
	}
	if result.Script.Opening.Content == "" {
		result.Script.Opening = fallback.Script.Opening
	}
	if result.Script.Middle.Content == "" {
		result.Script.Middle = fallback.Script.Middle
	}
	if result.Script.Ending.Content == "" {
		result.Script.Ending = fallback.Script.Ending
	}
	if len(result.Script.Storyboard) == 0 {
		result.Script.Storyboard = fallback.Script.Storyboard
	}
	if result.DigitalHuman.Persona == "" {
		result.DigitalHuman.Persona = fallback.DigitalHuman.Persona
	}
	if result.DigitalHuman.Tone == "" {
		result.DigitalHuman.Tone = fallback.DigitalHuman.Tone
	}
	if result.DigitalHuman.VideoRatio == "" {
		result.DigitalHuman.VideoRatio = fallback.DigitalHuman.VideoRatio
	}
	if result.DigitalHuman.SubtitleAdvice == "" {
		result.DigitalHuman.SubtitleAdvice = fallback.DigitalHuman.SubtitleAdvice
	}
	if result.DigitalHuman.VisualStyle == "" {
		result.DigitalHuman.VisualStyle = fallback.DigitalHuman.VisualStyle
	}
	if result.DigitalHuman.ShootingStyle == "" {
		result.DigitalHuman.ShootingStyle = fallback.DigitalHuman.ShootingStyle
	}
	if len(result.Promotion.Platforms) == 0 {
		result.Promotion.Platforms = fallback.Promotion.Platforms
	}
	if len(result.Promotion.ContentTags) == 0 {
		result.Promotion.ContentTags = fallback.Promotion.ContentTags
	}
	if len(result.Promotion.FocusMetrics) == 0 {
		result.Promotion.FocusMetrics = fallback.Promotion.FocusMetrics
	}
	if result.Promotion.OptimizationAdvice == "" {
		result.Promotion.OptimizationAdvice = fallback.Promotion.OptimizationAdvice
	}
	if result.AgentMessage.Summary == "" {
		result.AgentMessage.Summary = fallback.AgentMessage.Summary
	}
	if result.AgentMessage.MissingInfoNotice == "" {
		result.AgentMessage.MissingInfoNotice = fallback.AgentMessage.MissingInfoNotice
	}
	if len(result.AgentMessage.QuickActions) == 0 {
		result.AgentMessage.QuickActions = fallback.AgentMessage.QuickActions
	}
	return &result
}

func buildFallbackAgentResult(input SilkroadAgentInput, errorMessage string) *SilkroadAgentResult {
	productName := firstNonBlank(input.ProductName, "待分析商品")
	category := firstNonBlank(input.Category, "未填写类目")
	targetMarket := firstNonBlank(input.TargetMarket, "目标市场待补充")
	targetPlatform := firstNonBlank(input.TargetPlatform, "目标平台待补充")
	targetAudience := firstNonBlank(input.TargetAudience, "目标人群待补充")
	points := input.CoreSellingPoints
	if len(points) == 0 {
		points = []string{"核心卖点待补充"}
	}
	imageUnderstanding := "未上传商品图片，当前方案主要依据文本信息生成。"
	if hasUsableAgentImage(input.ImageDataURL) {
		imageUnderstanding = "已收到商品图片；若视觉模型不可用，建议人工补充包装文字、材质和认证信息。"
	}
	missingInfo := []string{"材质/成分/规格", "认证或检测报告", "目标售价区间"}
	if input.MaterialSpec != "" {
		missingInfo = []string{"认证或检测报告", "目标售价区间"}
	}
	isFood := isFoodProduct(productName, category)
	if isFood {
		missingInfo = []string{"配料表", "过敏原提示", "保质期", "产地/生产信息", "目标市场食品准入材料"}
	}
	localizationReason := "在目标市场信息不足时，先聚焦高频生活场景和明确卖点，可降低夸大宣传风险。"
	localizationKeywords := []string{"portable", "daily use", "easy to use", "lifestyle"}
	sceneSuggestions := []string{"开箱展示", "日常通勤", "办公室/宿舍", "使用前后对比但不夸大效果"}
	if isFood {
		localizationReason = "目标市场用户更容易被真实试吃、风味描述、价格场景和清晰食品标签信息吸引。"
		localizationKeywords = []string{"taste test", "ready to eat", "snack time", "food review"}
		sceneSuggestions = []string{"开箱试吃", "朋友聚餐", "夜宵场景", "便利餐食"}
	}

	return &SilkroadAgentResult{
		RecognizedInfo: RecognizedInfo{
			ProductName:        productName,
			Category:           category,
			TargetMarket:       targetMarket,
			TargetPlatform:     targetPlatform,
			TargetAudience:     targetAudience,
			CoreSellingPoints:  points,
			ImageUnderstanding: imageUnderstanding,
		},
		Overview: AgentOverview{
			ComplianceRiskLevel:     "中风险",
			MarketStrategy:          ternaryString(isFood, "先补齐食品标签与准入信息，再生成本地化素材。", "先用安全表达验证市场兴趣，再逐步补齐认证材料与本地化素材。"),
			RecommendedVideoStyle:   "生活场景化竖屏短视频",
			RecommendedDigitalHuman: "亲和、可信、语速自然的本地化讲解型数字人",
		},
		Compliance: CompliancePlan{
			Title:                "合规分析结果",
			Summary:              ternaryString(isFood, "当前信息仍不足以形成确定结论，建议补充配料表、过敏原、保质期、产地与食品准入材料，并避免健康功效或绝对化口味承诺。", "当前信息仍不足以形成确定结论，建议将页面表达控制在使用场景、材质说明和体验描述上，避免功效承诺、绝对安全、认证暗示等高风险表述。"),
			RiskTags:             ternaryStringSlice(isFood, []string{"食品标签", "准入材料", "功效表达"}, []string{"信息缺口", "功效表达", "认证证明"}),
			MissingInfo:          missingInfo,
			Suggestions:          []string{"补充材质、规格、容量或成分信息。", "补充面向目标市场的检测报告、认证编号或合规声明。", "广告文案使用“适合/帮助/便于”等边界表达。"},
			ForbiddenExpressions: []string{"100%安全", "永久有效", "官方认证", "治疗/治愈", "保证通过"},
			SaferExpressions:     []string{"日常使用更方便", "适合通勤/办公室等场景", "建议以实际检测材料为准", "详情请查看材质说明"},
		},
		Localization: Localization{
			Direction:        "场景化种草 + 实用价值说明",
			Reason:           localizationReason,
			Keywords:         localizationKeywords,
			Tone:             "自然、克制、可信",
			SceneSuggestions: sceneSuggestions,
		},
		Script: VideoScript{
			Title:    "短视频脚本",
			Duration: "20-25s",
			Opening:  ScriptSegment{Time: "0-3s", Content: "用一个真实生活小问题切入，展示目标用户在日常场景中的痛点。"},
			Middle:   ScriptSegment{Time: "3-20s", Content: ternaryString(isFood, "展示包装、色泽、口感反馈和食用场景，强调风味与便利性，避免绝对化或健康功效承诺。", "展示商品外观、核心卖点和使用步骤，强调便捷、材质说明和适用场景，避免绝对化承诺。")},
			Ending:   ScriptSegment{Time: "20-25s", Content: "用温和 CTA 收尾，引导查看详情、收藏或评论提问。"},
			Storyboard: []StoryboardShot{
				{Shot: "镜头 1", Visual: "商品与使用场景同框", Voiceover: "日常遇到这个小麻烦吗？", Subtitle: "Make daily routines easier"},
				{Shot: "镜头 2", Visual: "近景展示规格、材质或核心结构", Voiceover: "它的设计更适合随手使用。", Subtitle: "Portable and easy to use"},
				{Shot: "镜头 3", Visual: "用户完成一次使用并露出自然反馈", Voiceover: "想了解更多，可以看看详情。", Subtitle: "See details before purchase"},
			},
		},
		DigitalHuman: DigitalHuman{
			Persona:        "亲和型本地生活方式讲解者",
			Tone:           "自然、可靠、少夸张",
			VideoRatio:     "9:16",
			SubtitleAdvice: "目标市场语言字幕 + 关键卖点英文短词",
			VisualStyle:    "明亮、干净、生活方式感",
			ShootingStyle:  "手持近景 + 产品特写 + 场景演示",
		},
		Promotion: PromotionPlan{
			Platforms:          []string{targetPlatform},
			ContentTags:        []string{"场景种草", "开箱演示", "痛点解决", "本地化字幕"},
			FocusMetrics:       []string{"完播率", "点击率", "收藏率", "评论问题数量"},
			OptimizationAdvice: "先用 3 条不同开头测试用户痛点表达，保留合规边界清晰且完播率更高的版本。",
		},
		AgentMessage: AgentMessage{
			Summary:           "我已基于当前信息生成出海营销方案。当前结论偏审慎，适合先用于内容方向验证。",
			MissingInfoNotice: "建议继续补充材质/成分/规格、认证或检测材料、目标售价和竞品链接，以提高合规判断和脚本准确度。",
			QuickActions:      []string{"补充认证材料", "生成英文版脚本", "切换目标市场", "降低风险表达", "进入视频剪辑"},
		},
		ErrorMessage: errorMessage,
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func ternaryString(condition bool, yes string, no string) string {
	if condition {
		return yes
	}
	return no
}

func ternaryStringSlice(condition bool, yes []string, no []string) []string {
	if condition {
		return yes
	}
	return no
}
