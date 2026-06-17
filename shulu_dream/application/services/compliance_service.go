package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/drama-generator/backend/pkg/ai"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/utils"
)

type ComplianceRiskLevel string

const (
	ComplianceRiskGreen  ComplianceRiskLevel = "green"
	ComplianceRiskYellow ComplianceRiskLevel = "yellow"
	ComplianceRiskOrange ComplianceRiskLevel = "orange"
	ComplianceRiskRed    ComplianceRiskLevel = "red"
)

type ComplianceRequest struct {
	Title                  string   `json:"title"`
	Description            string   `json:"description"`
	TargetCountry          []string `json:"target_country"`
	MaterialComposition    string   `json:"material_composition,omitempty"`
	MarketingSellingPoints string   `json:"marketing_selling_points,omitempty"`
}

type ComplianceResult struct {
	Score                    int                 `json:"score"`
	Level                    ComplianceRiskLevel `json:"level"`
	LevelLabel               string              `json:"level_label"`
	Summary                  string              `json:"summary"`
	NonCompliancePoints      []string            `json:"non_compliance_points"`
	RectificationSuggestions []string            `json:"rectification_suggestions"`
	SuggestedCategories      []string            `json:"suggested_categories"`
}

type complianceRawResult struct {
	Score                    int      `json:"score"`
	Summary                  string   `json:"summary"`
	NonCompliancePoints      []string `json:"non_compliance_points"`
	RectificationSuggestions []string `json:"rectification_suggestions"`
	SuggestedCategories      []string `json:"suggested_categories"`
}

type ComplianceService struct {
	log             *logger.Logger
	enabled         bool
	baseURL         string
	apiKey          string
	endpoint        string
	candidateModels []string
}

func NewComplianceService(cfg config.ComplianceConfig, log *logger.Logger) *ComplianceService {
	svc := &ComplianceService{
		log:      log,
		enabled:  cfg.Enabled,
		baseURL:  cfg.BaseURL,
		apiKey:   cfg.APIKey,
		endpoint: cfg.Endpoint,
	}

	svc.candidateModels = buildCandidateModels(cfg.Model)

	return svc
}

func (s *ComplianceService) Evaluate(req ComplianceRequest) (*ComplianceResult, error) {
	sanitized := sanitizeComplianceRequest(req)
	if !s.enabled {
		return s.ruleBasedFallback(sanitized, "合规服务已禁用，使用规则引擎初筛"), nil
	}
	if s.apiKey == "" || s.baseURL == "" {
		return s.ruleBasedFallback(sanitized, "未配置合规 API Key，使用规则引擎初筛"), nil
	}

	payload, _ := json.MarshalIndent(sanitized, "", "  ")
	messages := []ai.ChatMessage{
		{
			Role: "system",
			Content: `你是跨境电商合规审查专家，请基于目标国家法规和常见平台禁限售规则进行评估。

输出要求：
1. 必须只输出一个 JSON 对象，不要输出 Markdown。
2. 字段必须完整：
{
  "score": 0-100 的整数,
  "summary": "一句话总结",
  "non_compliance_points": ["违规点1", "违规点2"],
  "rectification_suggestions": ["整改建议1", "整改建议2"],
  "suggested_categories": ["建议类目1", "建议类目2"]
}
3. 评分规则：
- 0-29: 低风险（绿色）
- 30-59: 中风险（黄色）
- 60-79: 高风险（橙色）
- >=80: 禁止（红色）
4. 如果涉及禁售、严重违规或高概率侵权，请给出 >=80 分。`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("请评估以下商品项目：\n%s", string(payload)),
		},
	}

	resp, modelName, err := s.chatCompletionWithFallbackModels(messages)
	if err != nil {
		s.log.Warnw("Compliance AI check failed, fallback to rule-based", "error", err, "models", s.candidateModels)
		return s.ruleBasedFallback(sanitized, "AI 合规服务调用失败，已切换规则引擎初筛"), nil
	}
	s.log.Infow("Compliance AI check succeeded", "model", modelName)
	if len(resp.Choices) == 0 {
		s.log.Warn("Compliance AI check returned empty choices, fallback to rule-based")
		return s.ruleBasedFallback(sanitized, "AI 未返回有效内容，已切换规则引擎初筛"), nil
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	var raw complianceRawResult
	if err := utils.SafeParseAIJSON(content, &raw); err != nil {
		s.log.Warnw("Compliance AI JSON parse failed, fallback to rule-based", "error", err)
		return s.ruleBasedFallback(sanitized, "AI 返回格式异常，已切换规则引擎初筛"), nil
	}

	result := &ComplianceResult{
		Score:                    clampScore(raw.Score),
		Summary:                  strings.TrimSpace(raw.Summary),
		NonCompliancePoints:      normalizeList(raw.NonCompliancePoints),
		RectificationSuggestions: normalizeList(raw.RectificationSuggestions),
		SuggestedCategories:      normalizeList(raw.SuggestedCategories),
	}
	result.Level, result.LevelLabel = riskLevelByScore(result.Score)
	if result.Summary == "" {
		result.Summary = fmt.Sprintf("已完成合规评估，当前风险等级：%s", result.LevelLabel)
	}

	return result, nil
}

func sanitizeComplianceRequest(req ComplianceRequest) ComplianceRequest {
	return ComplianceRequest{
		Title:                  strings.TrimSpace(req.Title),
		Description:            strings.TrimSpace(req.Description),
		TargetCountry:          normalizeCountryCodes(req.TargetCountry),
		MaterialComposition:    strings.TrimSpace(req.MaterialComposition),
		MarketingSellingPoints: strings.TrimSpace(req.MarketingSellingPoints),
	}
}

func (s *ComplianceService) ruleBasedFallback(req ComplianceRequest, reason string) *ComplianceResult {
	score := 18
	nonCompliancePoints := make([]string, 0)
	rectificationSuggestions := []string{
		reason,
		"建议补充产品检测报告、标签信息和目标国准入材料后再复核。",
	}
	suggestedCategories := []string{"家居用品", "服饰配件", "3C配件"}

	fullText := strings.ToLower(strings.Join([]string{
		req.Title,
		req.Description,
		req.MaterialComposition,
		req.MarketingSellingPoints,
	}, " "))

	type riskRule struct {
		pattern       *regexp.Regexp
		scoreDelta    int
		issue         string
		rectification string
	}

	rules := []riskRule{
		{
			pattern:       regexp.MustCompile(`枪|炮|弹|刀|武器|gun|weapon|knife`),
			scoreDelta:    85,
			issue:         "存在疑似武器/危险品描述，属于高风险禁限售项。",
			rectification: "删除危险品相关描述，确认是否为平台禁售商品。",
		},
		{
			pattern:       regexp.MustCompile(`药|医疗|治疗|处方|medical|drug|medicine`),
			scoreDelta:    38,
			issue:         "涉及医疗/药品功效表述，需医疗器械或药品准入资质。",
			rectification: "删除治疗功效承诺，补充合法资质并按医疗品类申报。",
		},
		{
			pattern:       regexp.MustCompile(`食品|保健|口服|food|supplement`),
			scoreDelta:    28,
			issue:         "涉及食品/保健品，目标国通常要求成分和标签合规。",
			rectification: "补充配料表、营养标签和进口准入证明。",
		},
		{
			pattern:       regexp.MustCompile(`电池|锂|battery|lithium`),
			scoreDelta:    24,
			issue:         "涉及电池类商品，可能涉及运输及认证限制。",
			rectification: "补充 UN38.3/MSDS/运输标签等文件。",
		},
		{
			pattern:       regexp.MustCompile(`儿童|婴儿|baby|kids|child`),
			scoreDelta:    20,
			issue:         "面向儿童的商品通常存在更严格安全与标签要求。",
			rectification: "补充儿童产品安全认证及年龄标签说明。",
		},
	}

	for _, rule := range rules {
		if rule.pattern.MatchString(fullText) {
			nonCompliancePoints = append(nonCompliancePoints, rule.issue)
			rectificationSuggestions = append(rectificationSuggestions, rule.rectification)
			if rule.scoreDelta >= 80 {
				score = maxInt(score, rule.scoreDelta)
			} else {
				score += rule.scoreDelta
			}
		}
	}

	highRegulationMarkets := map[string]struct{}{
		"US": {},
		"DE": {},
		"FR": {},
		"GB": {},
	}
	for _, country := range req.TargetCountry {
		if _, ok := highRegulationMarkets[country]; ok {
			score += 8
			rectificationSuggestions = append(rectificationSuggestions, "欧美市场监管较严格，建议补充英文标签、原产地和合规声明。")
			break
		}
	}

	score = clampScore(score)
	level, levelLabel := riskLevelByScore(score)

	summary := fmt.Sprintf("已完成初步合规评估（规则引擎），当前风险等级：%s。", levelLabel)
	if len(nonCompliancePoints) == 0 {
		nonCompliancePoints = append(nonCompliancePoints, "未识别到显著禁限售关键词，仍建议人工复核。")
	}

	return &ComplianceResult{
		Score:                    score,
		Level:                    level,
		LevelLabel:               levelLabel,
		Summary:                  summary,
		NonCompliancePoints:      nonCompliancePoints,
		RectificationSuggestions: uniqueStrings(rectificationSuggestions),
		SuggestedCategories:      suggestedCategories,
	}
}

func riskLevelByScore(score int) (ComplianceRiskLevel, string) {
	score = clampScore(score)
	switch {
	case score >= 80:
		return ComplianceRiskRed, "禁止"
	case score >= 60:
		return ComplianceRiskOrange, "高"
	case score >= 30:
		return ComplianceRiskYellow, "中"
	default:
		return ComplianceRiskGreen, "低"
	}
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func normalizeList(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{}
	}
	return result
}

func uniqueStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func buildCandidateModels(primary string) []string {
	models := []string{
		strings.TrimSpace(primary),
		"deepseek-v3-2-251201",
		"deepseek-v3-250324",
		"deepseek-v3-1-terminus",
		"deepseek-v3-1-250821",
		"deepseek-v3-241226",
		"deepseek-r1-250120",
	}

	seen := make(map[string]struct{}, len(models))
	result := make([]string, 0, len(models))
	for _, model := range models {
		if model == "" {
			continue
		}
		if _, exists := seen[model]; exists {
			continue
		}
		seen[model] = struct{}{}
		result = append(result, model)
	}
	return result
}

func normalizeCountryCodes(countries []string) []string {
	if len(countries) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(countries))
	out := make([]string, 0, len(countries))
	for _, country := range countries {
		code := strings.ToUpper(strings.TrimSpace(country))
		if len(code) < 2 {
			continue
		}
		if len(code) > 2 {
			code = code[:2]
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	return out
}

func (s *ComplianceService) chatCompletionWithFallbackModels(messages []ai.ChatMessage) (*ai.ChatCompletionResponse, string, error) {
	var lastErr error
	models := s.candidateModels
	if len(models) == 0 {
		models = buildCandidateModels("")
	}

	for _, model := range models {
		client := ai.NewOpenAIClient(s.baseURL, s.apiKey, model, s.endpoint)
		resp, err := client.ChatCompletion(messages, ai.WithTemperature(0.1), ai.WithMaxTokens(900))
		if err == nil {
			return resp, model, nil
		}

		lastErr = err
		if !isModelAccessError(err) {
			return nil, model, err
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no model available")
	}
	return nil, "", lastErr
}

func isModelAccessError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "ModelNotOpen") ||
		strings.Contains(msg, "InvalidEndpointOrModel") ||
		strings.Contains(msg, "does not exist or you do not have access")
}
