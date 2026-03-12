package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/ai"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

var (
	ErrAssistDramaNotFound       = errors.New("drama not found")
	ErrAssistDeepSeekUnavailable = errors.New("deepseek v3.2 not available")
)

type GenerateAssistScriptRequest struct {
	DramaID       string `json:"drama_id" binding:"required"`
	EpisodeNumber int    `json:"episode_number"`
	Prompt        string `json:"prompt" binding:"required"`
	Model         string `json:"model"`
}

type GenerateAssistScriptResult struct {
	Content string `json:"content"`
	Model   string `json:"model"`
}

type AIAssistService struct {
	db         *gorm.DB
	aiService  *AIService
	log        *logger.Logger
	cfg        *config.Config
	promptI18n *PromptI18n
}

func NewAIAssistService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AIAssistService {
	return &AIAssistService{
		db:         db,
		aiService:  NewAIService(db, log, cfg),
		log:        log,
		cfg:        cfg,
		promptI18n: NewPromptI18n(cfg),
	}
}

func (s *AIAssistService) GenerateEpisodeScript(req *GenerateAssistScriptRequest) (*GenerateAssistScriptResult, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	var drama models.Drama
	if err := s.db.Where("id = ?", req.DramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssistDramaNotFound
		}
		return nil, err
	}

	episodeNumber := req.EpisodeNumber
	if episodeNumber <= 0 {
		episodeNumber = 1
	}

	systemPrompt := s.buildSystemPrompt()
	userPrompt := s.buildUserPrompt(&drama, episodeNumber, prompt)

	content, model, err := s.generateWithDeepSeek(systemPrompt, userPrompt, req.Model)
	if err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("empty response content")
	}

	return &GenerateAssistScriptResult{
		Content: content,
		Model:   model,
	}, nil
}

func (s *AIAssistService) buildSystemPrompt() string {
	if s.promptI18n.IsEnglish() {
		return `You are an expert short-drama script assistant.

Write a complete episode script in plain text based on user requirements.

Requirements:
1. Output plain text only. Do not output markdown headings, code blocks, or JSON.
2. Keep Chinese names/settings consistent if user input includes Chinese context.
3. Include scene transitions, action descriptions, and character dialogue.
4. Strong pacing: setup -> conflict escalation -> turning point -> closure.
5. Ensure the content is directly usable for later character/scene extraction.`
	}

	return `你是专业短剧编剧助手。

请根据用户需求，产出可直接用于后续“提取角色和场景”的章节剧本正文。

要求：
1. 只输出纯文本，不要输出 Markdown 标题、代码块或 JSON。
2. 内容结构清晰，包含场景切换、人物动作、对白与情绪推进。
3. 节奏要完整：铺垫 -> 冲突升级 -> 关键转折 -> 收束。
4. 人设、地点、事件前后逻辑一致，避免明显跳戏。
5. 语言自然，画面感强。`
}

func (s *AIAssistService) buildUserPrompt(drama *models.Drama, episodeNumber int, prompt string) string {
	title := strings.TrimSpace(drama.Title)
	if title == "" {
		title = "未命名项目"
	}

	description := ""
	if drama.Description != nil {
		description = strings.TrimSpace(*drama.Description)
	}
	if description == "" {
		description = "无"
	}

	if s.promptI18n.IsEnglish() {
		return fmt.Sprintf("Project title: %s\nProject description: %s\nEpisode: %d\nUser requirement: %s\n\nPlease generate the final episode script now.",
			title, description, episodeNumber, prompt)
	}

	return fmt.Sprintf("项目标题：%s\n项目描述：%s\n章节：第%d章\n用户需求：%s\n\n请直接输出本章最终剧本正文。",
		title, description, episodeNumber, prompt)
}

func (s *AIAssistService) generateWithDeepSeek(systemPrompt string, userPrompt string, requestedModel string) (string, string, error) {
	activeDeepSeekModels := s.listActiveDeepSeekTextModels()
	models := buildAssistCandidateModels(requestedModel, s.cfg.Compliance.Model, activeDeepSeekModels...)
	var lastErr error

	baseURL := strings.TrimSpace(s.cfg.Compliance.BaseURL)
	apiKey := strings.TrimSpace(s.cfg.Compliance.APIKey)
	endpoint := strings.TrimSpace(s.cfg.Compliance.Endpoint)
	if endpoint == "" {
		endpoint = "/chat/completions"
	}

	// 优先使用环境配置的 DeepSeek 直连（火山方舟 OpenAI 兼容接口）
	if baseURL != "" && apiKey != "" {
		for _, model := range models {
			client := ai.NewOpenAIClient(baseURL, apiKey, model, endpoint)
			text, err := client.GenerateText(
				userPrompt,
				systemPrompt,
				ai.WithTemperature(0.8),
				ai.WithMaxTokens(2200),
			)
			if err == nil && strings.TrimSpace(text) != "" {
				return text, model, nil
			}
			lastErr = err
			if err != nil && !isAssistModelAccessError(err) {
				return "", "", err
			}
		}
	}

	// 其次尝试数据库中配置的文本服务（必须匹配 DeepSeek 模型）
	for _, model := range models {
		client, err := s.aiService.GetAIClientForModel("text", model)
		if err != nil {
			lastErr = err
			continue
		}

		text, err := client.GenerateText(
			userPrompt,
			systemPrompt,
			ai.WithTemperature(0.8),
			ai.WithMaxTokens(2200),
		)
		if err == nil && strings.TrimSpace(text) != "" {
			return text, model, nil
		}
		lastErr = err
		if err != nil && !isAssistModelAccessError(err) {
			return "", "", err
		}
	}

	if lastErr == nil {
		lastErr = errors.New("no available deepseek model")
	}

	return "", "", fmt.Errorf("%w: %v", ErrAssistDeepSeekUnavailable, lastErr)
}

func (s *AIAssistService) listActiveDeepSeekTextModels() []string {
	configs, err := s.aiService.ListConfigs("text")
	if err != nil {
		if s.log != nil {
			s.log.Warnw("failed to list text AI configs for assist model fallback", "error", err)
		}
		return []string{}
	}

	models := make([]string, 0)
	for _, cfg := range configs {
		if !cfg.IsActive {
			continue
		}
		for _, model := range cfg.Model {
			trimmed := strings.TrimSpace(model)
			if trimmed == "" {
				continue
			}
			if isDeepSeekLikeModel(trimmed) {
				models = append(models, trimmed)
			}
		}
	}

	return models
}

func isDeepSeekLikeModel(model string) bool {
	value := strings.ToLower(strings.TrimSpace(model))
	if value == "" {
		return false
	}
	return strings.Contains(value, "deepseek") || strings.HasPrefix(value, "ep-")
}

func buildAssistCandidateModels(requestedModel string, configModel string, extraModels ...string) []string {
	models := []string{
		strings.TrimSpace(requestedModel),
		strings.TrimSpace(configModel),
	}
	models = append(models, extraModels...)
	models = append(models,
		"deepseek-v3-2-251201",
		"deepseek-v3-250324",
		"deepseek-v3-1-terminus",
		"deepseek-v3-1-250821",
		"deepseek-v3-241226",
		"deepseek-v3-2",
		"deepseek-v3.2",
	)

	seen := make(map[string]struct{}, len(models))
	result := make([]string, 0, len(models))
	for _, model := range models {
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		result = append(result, model)
	}
	return result
}

func isAssistModelAccessError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "ModelNotOpen") ||
		strings.Contains(msg, "InvalidEndpointOrModel") ||
		strings.Contains(msg, "does not exist or you do not have access") ||
		strings.Contains(msg, "no active config found for model")
}
