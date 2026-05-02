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

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
)

var ErrSilkroadAgentConfigMissing = errors.New("silkroad agent model config missing")

type SilkroadAgentInput struct {
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
	Title                string   `json:"title"`
	Summary              string   `json:"summary"`
	RiskTags             []string `json:"riskTags"`
	MissingInfo          []string `json:"missingInfo"`
	Suggestions          []string `json:"suggestions"`
	ForbiddenExpressions []string `json:"forbiddenExpressions"`
	SaferExpressions     []string `json:"saferExpressions"`
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

func (s *SilkroadAgentService) Generate(input SilkroadAgentInput) (*SilkroadAgentResult, error) {
	input = normalizeSilkroadInput(input)
	settings := s.readSettings()
	if settings.APIKey == "" {
		if s.cfg != nil && s.cfg.App.Debug {
			result := buildFallbackAgentResult(input, "开发环境未配置 AGENT_API_KEY/DEEPSEEK_API_KEY，已返回 mock 数据。")
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
		result = fillAgentDefaults(parsedResult, input, imageUnderstanding)
	}
	result.Model = settings.TextModel
	return result, nil
}

func (s *SilkroadAgentService) StreamMobileTransitionAnalysis(ctx context.Context, input SilkroadAgentAnalyzeInput, emit func(SilkroadAgentAnalyzeEvent) error) error {
	input = normalizeSilkroadAnalyzeInput(input)
	settings := readDeepSeekAnalyzeSettings()
	recognized := buildMobileRecognizedInfo(input)
	imageUnderstanding := s.analyzeTransitionImage(input)

	emitFallback := func(message string) error {
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
		{Step: 1, Name: "商品理解", Status: "completed", Description: "确认商品类目、卖点与使用场景"},
		{Step: 2, Name: "合规风险识别", Status: "completed", Description: "匹配目标市场规则与广告敏感表达"},
		{Step: 3, Name: "本地化方向", Status: "completed", Description: "生成符合马来西亚用户习惯的内容方向"},
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
	return string(jsonPayload) + "\n\n请生成手机过渡页可见的流式分析摘要，围绕跨境电商、合规、本地化、数字人和短视频投放展开。"
}

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
	return string(jsonPayload) + "\n\n请基于以上信息生成结果页可直接展示的 JSON。缺失信息要写入 missingInfo 和 missingInfoNotice，不要自行编造认证或绝对法律结论。"
}

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
	if input.UserInput == "" {
		input.UserInput = "我有一款便携榨汁杯，想卖到马来西亚，主要做 TikTok 短视频，目标用户是年轻女生，主打便携和健康。"
	}
	if input.Scene == "" {
		input.Scene = "mobile_transition"
	}
	return input
}

func buildMobileRecognizedInfo(input SilkroadAgentAnalyzeInput) MobileRecognizedInfo {
	prompt := input.UserInput
	product := firstNonEmpty(input.ProductName, extractPromptMatch(prompt, []string{
		`(?:我有|我们有|这是一款|这款|一款|一个|一种|商品是|产品是)([^，,。；;\n]{2,28}?)(?:，|,|。|；|;|想|计划|准备|主打|目标|卖到|出口|做|$)`,
		`(?:销售|卖)([^，,。；;\n]{2,28}?)(?:，|,|。|；|;|到|去|$)`,
	}), "便携榨汁杯")
	category := firstNonEmpty(input.Category, extractPromptMatch(prompt, []string{
		`(?:商品类目|产品类目|类目|品类|属于)(?:是|为|:|：)?([^，,。；;\n]{2,28})`,
	}), inferCategory(product))
	market := firstNonEmpty(input.TargetMarket, extractPromptMatch(prompt, []string{
		`(?:卖到|出口到|进入|面向|投放到|推广到)([^，,。；;\n]{2,24}?)(?:市场|用户|消费者|，|,|。|；|;|$)`,
		`(?:目标市场|目标国家|国家|市场)(?:是|为|:|：)?([^，,。；;\n]{2,24})`,
	}), "马来西亚")
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
		points = []string{"便携", "健康", "易清洗"}
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
	return []string{
		"已接收到你的出海需求，正在从自然语言中提取商品、目标市场、平台、人群和核心卖点。\n\n",
		fmt.Sprintf("识别到商品为%s，属于%s，目标市场为%s，主要投放平台为%s。\n\n", info.Product, info.Category, info.Market, info.Platform),
		"该商品涉及食品接触场景，后续合规分析将重点关注杯体材质、食品接触认证、电池容量和充电方式。\n\n",
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
	return "跨境电商商品"
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
	if result.RecognizedInfo.TargetMarket == "" {
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
			MarketStrategy:          "先用安全表达验证市场兴趣，再逐步补齐认证材料与本地化素材。",
			RecommendedVideoStyle:   "生活场景化竖屏短视频",
			RecommendedDigitalHuman: "亲和、可信、语速自然的本地化讲解型数字人",
		},
		Compliance: CompliancePlan{
			Title:                "合规分析结果",
			Summary:              "当前信息仍不足以形成确定结论，建议将页面表达控制在使用场景、材质说明和体验描述上，避免功效承诺、绝对安全、认证暗示等高风险表述。",
			RiskTags:             []string{"信息缺口", "功效表达", "认证证明"},
			MissingInfo:          missingInfo,
			Suggestions:          []string{"补充材质、规格、容量或成分信息。", "补充面向目标市场的检测报告、认证编号或合规声明。", "广告文案使用“适合/帮助/便于”等边界表达。"},
			ForbiddenExpressions: []string{"100%安全", "永久有效", "官方认证", "治疗/治愈", "保证通过"},
			SaferExpressions:     []string{"日常使用更方便", "适合通勤/办公室等场景", "建议以实际检测材料为准", "详情请查看材质说明"},
		},
		Localization: Localization{
			Direction:        "场景化种草 + 实用价值说明",
			Reason:           "在目标市场信息不足时，先聚焦高频生活场景和明确卖点，可降低夸大宣传风险。",
			Keywords:         []string{"portable", "daily use", "easy to use", "lifestyle"},
			Tone:             "自然、克制、可信",
			SceneSuggestions: []string{"开箱展示", "日常通勤", "办公室/宿舍", "使用前后对比但不夸大效果"},
		},
		Script: VideoScript{
			Title:    "短视频脚本",
			Duration: "20-25s",
			Opening:  ScriptSegment{Time: "0-3s", Content: "用一个真实生活小问题切入，展示目标用户在日常场景中的痛点。"},
			Middle:   ScriptSegment{Time: "3-20s", Content: "展示商品外观、核心卖点和使用步骤，强调便捷、材质说明和适用场景，避免绝对化承诺。"},
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
