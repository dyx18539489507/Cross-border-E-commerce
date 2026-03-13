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

type AIService struct {
	db  *gorm.DB
	log *logger.Logger
	cfg *config.Config
}

func NewAIService(db *gorm.DB, log *logger.Logger, cfgs ...*config.Config) *AIService {
	var cfg *config.Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}

	return &AIService{
		db:  db,
		log: log,
		cfg: cfg,
	}
}

func firstAIDeviceID(deviceIDs []string) string {
	if len(deviceIDs) == 0 {
		return ""
	}
	return strings.TrimSpace(deviceIDs[0])
}

type CreateAIConfigRequest struct {
	ServiceType   string            `json:"service_type" binding:"required,oneof=text image video"`
	Name          string            `json:"name" binding:"required,min=1,max=100"`
	Provider      string            `json:"provider" binding:"required"`
	BaseURL       string            `json:"base_url" binding:"required,url"`
	APIKey        string            `json:"api_key" binding:"required"`
	Model         models.ModelField `json:"model" binding:"required"`
	Endpoint      string            `json:"endpoint"`
	QueryEndpoint string            `json:"query_endpoint"`
	Priority      int               `json:"priority"`
	IsDefault     bool              `json:"is_default"`
	Settings      string            `json:"settings"`
}

type UpdateAIConfigRequest struct {
	Name          string             `json:"name" binding:"omitempty,min=1,max=100"`
	Provider      string             `json:"provider"`
	BaseURL       string             `json:"base_url" binding:"omitempty,url"`
	APIKey        string             `json:"api_key"`
	Model         *models.ModelField `json:"model"`
	Endpoint      string             `json:"endpoint"`
	QueryEndpoint string             `json:"query_endpoint"`
	Priority      *int               `json:"priority"`
	IsDefault     bool               `json:"is_default"`
	IsActive      bool               `json:"is_active"`
	Settings      string             `json:"settings"`
}

type TestConnectionRequest struct {
	BaseURL  string            `json:"base_url" binding:"required,url"`
	APIKey   string            `json:"api_key" binding:"required"`
	Model    models.ModelField `json:"model" binding:"required"`
	Provider string            `json:"provider"`
	Endpoint string            `json:"endpoint"`
}

func (s *AIService) CreateConfig(req *CreateAIConfigRequest, deviceIDs ...string) (*models.AIServiceConfig, error) {
	deviceID := firstAIDeviceID(deviceIDs)
	// 根据 provider 和 service_type 自动设置 endpoint
	endpoint := req.Endpoint
	queryEndpoint := req.QueryEndpoint

	if endpoint == "" {
		switch req.Provider {
		case "gemini", "google":
			if req.ServiceType == "text" {
				endpoint = "/v1beta/models/{model}:generateContent"
			} else if req.ServiceType == "image" {
				endpoint = "/v1beta/models/{model}:generateContent"
			}
		case "openai":
			if req.ServiceType == "text" {
				endpoint = "/chat/completions"
			} else if req.ServiceType == "image" {
				endpoint = "/images/generations"
			} else if req.ServiceType == "video" {
				endpoint = "/videos"
				if queryEndpoint == "" {
					queryEndpoint = "/videos/{taskId}"
				}
			}
		case "chatfire":
			if req.ServiceType == "text" {
				endpoint = "/chat/completions"
			} else if req.ServiceType == "image" {
				endpoint = "/images/generations"
			} else if req.ServiceType == "video" {
				endpoint = "/video/generations"
				if queryEndpoint == "" {
					queryEndpoint = "/video/task/{taskId}"
				}
			}
		case "doubao", "volcengine", "volces":
			if req.ServiceType == "video" {
				endpoint = "/contents/generations/tasks"
				if queryEndpoint == "" {
					queryEndpoint = "/generations/tasks/{taskId}"
				}
			}
		default:
			// 默认使用 OpenAI 格式
			if req.ServiceType == "text" {
				endpoint = "/chat/completions"
			} else if req.ServiceType == "image" {
				endpoint = "/images/generations"
			}
		}
	}

	config := &models.AIServiceConfig{
		DeviceID:      deviceID,
		ServiceType:   req.ServiceType,
		Name:          req.Name,
		Provider:      req.Provider,
		BaseURL:       req.BaseURL,
		APIKey:        req.APIKey,
		Model:         req.Model,
		Endpoint:      endpoint,
		QueryEndpoint: queryEndpoint,
		Priority:      req.Priority,
		IsDefault:     req.IsDefault,
		IsActive:      true,
		Settings:      req.Settings,
	}

	if err := s.db.Create(config).Error; err != nil {
		s.log.Errorw("Failed to create AI config", "error", err)
		return nil, err
	}

	s.log.Infow("AI config created", "config_id", config.ID, "provider", req.Provider, "endpoint", endpoint)
	return config, nil
}

func (s *AIService) GetConfig(configID uint, deviceIDs ...string) (*models.AIServiceConfig, error) {
	deviceID := firstAIDeviceID(deviceIDs)
	var config models.AIServiceConfig
	err := s.scopedConfigQuery().
		Where("id = ?", configID).
		Scopes(withScopedDeviceFallback(deviceID)).
		First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("config not found")
		}
		return nil, err
	}
	return &config, nil
}

func (s *AIService) ListConfigs(serviceType string, deviceIDs ...string) ([]models.AIServiceConfig, error) {
	deviceID := firstAIDeviceID(deviceIDs)
	var configs []models.AIServiceConfig
	query := s.scopedConfigQuery().Scopes(withScopedDeviceFallback(deviceID))

	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}

	err := query.Order("priority DESC, created_at DESC").Find(&configs).Error
	if err != nil {
		s.log.Errorw("Failed to list AI configs", "error", err)
		return nil, err
	}

	return configs, nil
}

func (s *AIService) UpdateConfig(configID uint, req *UpdateAIConfigRequest, deviceIDs ...string) (*models.AIServiceConfig, error) {
	deviceID := firstAIDeviceID(deviceIDs)
	var config models.AIServiceConfig
	if err := s.scopedConfigQuery().
		Where("id = ?", configID).
		Scopes(withScopedDeviceFallback(deviceID)).
		First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("config not found")
		}
		return nil, err
	}

	tx := s.db.Begin()

	// 不再需要is_default独占逻辑

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Provider != "" {
		updates["provider"] = req.Provider
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Model != nil && len(*req.Model) > 0 {
		updates["model"] = *req.Model
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}

	// 如果提供了 provider，根据 provider 和 service_type 自动设置 endpoint
	if req.Provider != "" && req.Endpoint == "" {
		provider := req.Provider
		serviceType := config.ServiceType

		switch provider {
		case "gemini", "google":
			if serviceType == "text" || serviceType == "image" {
				updates["endpoint"] = "/v1beta/models/{model}:generateContent"
			}
		case "openai":
			if serviceType == "text" {
				updates["endpoint"] = "/chat/completions"
			} else if serviceType == "image" {
				updates["endpoint"] = "/images/generations"
			} else if serviceType == "video" {
				updates["endpoint"] = "/videos"
				updates["query_endpoint"] = "/videos/{taskId}"
			}
		case "chatfire":
			if serviceType == "text" {
				updates["endpoint"] = "/chat/completions"
			} else if serviceType == "image" {
				updates["endpoint"] = "/images/generations"
			} else if serviceType == "video" {
				updates["endpoint"] = "/video/generations"
				updates["query_endpoint"] = "/video/task/{taskId}"
			}
		}
	} else if req.Endpoint != "" {
		updates["endpoint"] = req.Endpoint
	}

	// 允许清空query_endpoint，所以不检查是否为空
	updates["query_endpoint"] = req.QueryEndpoint
	if req.Settings != "" {
		updates["settings"] = req.Settings
	}
	updates["is_default"] = req.IsDefault
	updates["is_active"] = req.IsActive

	if err := tx.Model(&config).Updates(updates).Error; err != nil {
		tx.Rollback()
		s.log.Errorw("Failed to update AI config", "error", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.log.Infow("AI config updated", "config_id", configID)
	return &config, nil
}

func (s *AIService) DeleteConfig(configID uint, deviceIDs ...string) error {
	deviceID := firstAIDeviceID(deviceIDs)
	result := s.scopedConfigQuery().
		Where("id = ?", configID).
		Scopes(withScopedDeviceFallback(deviceID)).
		Delete(&models.AIServiceConfig{})

	if result.Error != nil {
		s.log.Errorw("Failed to delete AI config", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("config not found")
	}

	s.log.Infow("AI config deleted", "config_id", configID)
	return nil
}

func (s *AIService) TestConnection(req *TestConnectionRequest) error {
	s.log.Infow("TestConnection called", "baseURL", req.BaseURL, "provider", req.Provider, "endpoint", req.Endpoint, "modelCount", len(req.Model))

	// 使用第一个模型进行测试
	model := ""
	if len(req.Model) > 0 {
		model = req.Model[0]
	}
	s.log.Infow("Using model for test", "model", model, "provider", req.Provider)

	// 根据 provider 参数选择客户端
	var client ai.AIClient
	var endpoint string

	switch req.Provider {
	case "gemini", "google":
		// Gemini
		s.log.Infow("Using Gemini client", "baseURL", req.BaseURL)
		endpoint = "/v1beta/models/{model}:generateContent"
		client = ai.NewGeminiClient(req.BaseURL, req.APIKey, model, endpoint)
	case "openai", "chatfire":
		// OpenAI 格式（包括 chatfire 等）
		s.log.Infow("Using OpenAI-compatible client", "baseURL", req.BaseURL, "provider", req.Provider)
		endpoint = req.Endpoint
		if endpoint == "" {
			endpoint = "/chat/completions"
		}
		client = ai.NewOpenAIClient(req.BaseURL, req.APIKey, model, endpoint)
	default:
		// 默认使用 OpenAI 格式
		s.log.Infow("Using default OpenAI-compatible client", "baseURL", req.BaseURL)
		endpoint = req.Endpoint
		if endpoint == "" {
			endpoint = "/chat/completions"
		}
		client = ai.NewOpenAIClient(req.BaseURL, req.APIKey, model, endpoint)
	}

	s.log.Infow("Calling TestConnection on client", "endpoint", endpoint)
	err := client.TestConnection()
	if err != nil {
		s.log.Errorw("TestConnection failed", "error", err)
	} else {
		s.log.Infow("TestConnection succeeded")
	}
	return err
}

func (s *AIService) GetDefaultConfig(serviceType string, deviceIDs ...string) (*models.AIServiceConfig, error) {
	deviceID := firstAIDeviceID(deviceIDs)
	var config models.AIServiceConfig
	err := s.scopedConfigQuery().
		Where("service_type = ? AND is_active = ?", serviceType, true).
		Scopes(withScopedDeviceFallback(deviceID)).
		First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active config found")
		}
		return nil, err
	}

	return &config, nil
}

// GetConfigForModel 根据服务类型和模型名称获取优先级最高的激活配置
func (s *AIService) GetConfigForModel(serviceType string, modelName string, deviceIDs ...string) (*models.AIServiceConfig, error) {
	deviceID := firstAIDeviceID(deviceIDs)
	var configs []models.AIServiceConfig
	err := s.scopedConfigQuery().
		Where("service_type = ? AND is_active = ?", serviceType, true).
		Scopes(withScopedDeviceFallback(deviceID)).
		Find(&configs).Error

	if err != nil {
		return nil, err
	}

	// 查找包含指定模型的配置
	for _, config := range configs {
		for _, model := range config.Model {
			if model == modelName {
				return &config, nil
			}
		}
	}

	return nil, errors.New("no active config found for model: " + modelName)
}

func (s *AIService) scopedConfigQuery() *gorm.DB {
	return s.db.Model(&models.AIServiceConfig{}).
		Order("CASE WHEN device_id = '' THEN 1 ELSE 0 END ASC").
		Order("priority DESC").
		Order("created_at DESC")
}

func withScopedDeviceFallback(deviceID string) func(*gorm.DB) *gorm.DB {
	deviceID = strings.TrimSpace(deviceID)
	return func(db *gorm.DB) *gorm.DB {
		if deviceID == "" {
			return db
		}
		return db.Where("(device_id = ? OR device_id = '')", deviceID)
	}
}

func (s *AIService) GetAIClient(serviceType string, deviceIDs ...string) (ai.AIClient, error) {
	config, err := s.GetDefaultConfig(serviceType, deviceIDs...)
	if err != nil {
		if serviceType == "text" && isNoActiveConfigError(err) {
			client, fallbackErr := s.newComplianceTextClient("")
			if fallbackErr == nil {
				return client, nil
			}
			if s.log != nil {
				s.log.Warnw("Failed to use compliance text fallback", "error", fallbackErr)
			}
		}
		return nil, err
	}

	// 使用第一个模型
	model := ""
	if len(config.Model) > 0 {
		model = config.Model[0]
	}

	// 使用数据库配置中的 endpoint，如果为空则根据 provider 设置默认值
	endpoint := config.Endpoint
	if endpoint == "" {
		switch config.Provider {
		case "gemini", "google":
			endpoint = "/v1beta/models/{model}:generateContent"
		default:
			endpoint = "/chat/completions"
		}
	}

	// 根据 provider 创建对应的客户端
	switch config.Provider {
	case "gemini", "google":
		return ai.NewGeminiClient(config.BaseURL, config.APIKey, model, endpoint), nil
	default:
		// openai, chatfire 等其他厂商都使用 OpenAI 格式
		return ai.NewOpenAIClient(config.BaseURL, config.APIKey, model, endpoint), nil
	}
}

// GetAIClientForModel 根据服务类型和模型名称获取对应的AI客户端
func (s *AIService) GetAIClientForModel(serviceType string, modelName string, deviceIDs ...string) (ai.AIClient, error) {
	config, err := s.GetConfigForModel(serviceType, modelName, deviceIDs...)
	if err != nil {
		if serviceType == "text" && isNoActiveConfigError(err) {
			client, fallbackErr := s.newComplianceTextClient(modelName)
			if fallbackErr == nil {
				return client, nil
			}
			if s.log != nil {
				s.log.Warnw("Failed to use compliance text fallback for model", "model", modelName, "error", fallbackErr)
			}
		}
		return nil, err
	}

	// 使用数据库配置中的 endpoint，如果为空则根据 provider 设置默认值
	endpoint := config.Endpoint
	if endpoint == "" {
		switch config.Provider {
		case "gemini", "google":
			endpoint = "/v1beta/models/{model}:generateContent"
		default:
			endpoint = "/chat/completions"
		}
	}

	// 根据 provider 创建对应的客户端
	switch config.Provider {
	case "gemini", "google":
		return ai.NewGeminiClient(config.BaseURL, config.APIKey, modelName, endpoint), nil
	default:
		// openai, chatfire 等其他厂商都使用 OpenAI 格式
		return ai.NewOpenAIClient(config.BaseURL, config.APIKey, modelName, endpoint), nil
	}
}

func (s *AIService) GenerateText(prompt string, systemPrompt string, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	client, err := s.GetAIClient("text")
	if err != nil {
		return "", normalizeTextGenerationError(fmt.Errorf("failed to get AI client: %w", err))
	}

	text, genErr := client.GenerateText(prompt, systemPrompt, options...)
	if genErr == nil {
		return text, nil
	}

	if isTextServiceFallbackError(genErr) {
		if s.log != nil {
			s.log.Warnw("Default text model failed, trying compliance fallback", "error", genErr)
		}
		fallbackClient, fallbackErr := s.newComplianceTextClient("")
		if fallbackErr == nil {
			fallbackText, retryErr := fallbackClient.GenerateText(prompt, systemPrompt, options...)
			if retryErr == nil && strings.TrimSpace(fallbackText) != "" {
				return fallbackText, nil
			}
			if retryErr != nil {
				genErr = retryErr
			}
		} else if s.log != nil {
			s.log.Warnw("Compliance fallback unavailable for text generation", "error", fallbackErr)
		}
	}

	return "", normalizeTextGenerationError(genErr)
}

func (s *AIService) GenerateTextForModel(modelName string, prompt string, systemPrompt string, options ...func(*ai.ChatCompletionRequest)) (string, error) {
	requestedModel := strings.TrimSpace(modelName)
	if requestedModel == "" {
		return s.GenerateText(prompt, systemPrompt, options...)
	}

	var lastErr error

	// 显式指定模型时优先尝试合规通道，避免默认文本配置被停用时仍然命中不可用模型。
	if fallbackClient, err := s.newComplianceTextClient(requestedModel); err == nil {
		text, genErr := fallbackClient.GenerateText(prompt, systemPrompt, options...)
		if genErr == nil && strings.TrimSpace(text) != "" {
			return text, nil
		}
		lastErr = genErr
		if genErr != nil && !isTextServiceFallbackError(genErr) {
			return "", normalizeTextGenerationError(genErr)
		}
	} else {
		lastErr = err
	}

	client, err := s.GetAIClientForModel("text", requestedModel)
	if err != nil {
		if lastErr == nil {
			lastErr = err
		}
		return "", normalizeTextGenerationError(fmt.Errorf("failed to get AI client: %w", lastErr))
	}

	text, genErr := client.GenerateText(prompt, systemPrompt, options...)
	if genErr != nil {
		if isTextServiceFallbackError(genErr) && lastErr != nil {
			return "", normalizeTextGenerationError(lastErr)
		}
		return "", normalizeTextGenerationError(genErr)
	}

	return text, nil
}

func (s *AIService) newComplianceTextClient(modelName string) (ai.AIClient, error) {
	if s.cfg == nil || !s.cfg.Compliance.Enabled {
		return nil, errors.New("compliance text fallback is disabled")
	}

	baseURL := strings.TrimSpace(s.cfg.Compliance.BaseURL)
	apiKey := strings.TrimSpace(s.cfg.Compliance.APIKey)
	endpoint := strings.TrimSpace(s.cfg.Compliance.Endpoint)
	model := strings.TrimSpace(modelName)
	if model == "" {
		model = strings.TrimSpace(s.cfg.Compliance.Model)
	}
	if endpoint == "" {
		endpoint = "/chat/completions"
	}

	if baseURL == "" || apiKey == "" || model == "" {
		return nil, errors.New("compliance text fallback is not configured")
	}

	if s.log != nil {
		s.log.Infow("Using compliance text fallback", "model", model, "base_url", baseURL)
	}

	return ai.NewOpenAIClient(baseURL, apiKey, model, endpoint), nil
}

func isNoActiveConfigError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no active config found")
}

func isTextServiceFallbackError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "reached the set inference limit") ||
		strings.Contains(msg, "model service has been paused") ||
		strings.Contains(msg, "safe experience mode") ||
		strings.Contains(msg, "modelnotopen") ||
		strings.Contains(msg, "invalidendpointormodel") ||
		strings.Contains(msg, "does not exist or you do not have access") ||
		strings.Contains(msg, "no active config found") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "service unavailable") ||
		strings.Contains(msg, "temporarily unavailable") ||
		strings.Contains(msg, "overloaded") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "context deadline exceeded")
}

func normalizeTextGenerationError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "reached the set inference limit"),
		strings.Contains(lower, "model service has been paused"),
		strings.Contains(lower, "safe experience mode"):
		return fmt.Errorf("当前默认文本模型额度已耗尽或服务已暂停，请在“设置 > AI服务配置”切换到可用文本模型，或在供应商控制台关闭 Safe Experience Mode 后重试")
	case strings.Contains(lower, "modelnotopen"),
		strings.Contains(lower, "invalidendpointormodel"),
		strings.Contains(lower, "does not exist or you do not have access"),
		strings.Contains(lower, "no active config found"),
		strings.Contains(lower, "failed to get ai client"):
		return fmt.Errorf("当前没有可用的文本模型，请在“设置 > AI服务配置”中启用可用的文本模型后重试")
	case strings.Contains(lower, "rate limit"),
		strings.Contains(lower, "too many requests"):
		return fmt.Errorf("文本模型请求过于频繁，请稍后重试")
	}

	return err
}
