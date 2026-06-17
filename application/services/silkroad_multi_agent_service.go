package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
)

type PlanningAgentOutput struct {
	TaskChain             []string `json:"task_chain"`
	ExecutionSteps        []string `json:"execution_steps"`
	MissingInformation    []string `json:"missing_information"`
	FinalOutputPlan       []string `json:"final_output_plan"`
	RecommendedWorkspaces []string `json:"recommended_workspaces"`
}

type ProductAgentOutput struct {
	ProductName        string            `json:"product_name"`
	Category           string            `json:"category"`
	CoreSellingPoints  []string          `json:"core_selling_points"`
	UsageScenarios     []string          `json:"usage_scenarios"`
	TargetUsers        []string          `json:"target_users"`
	SensitivePoints    []string          `json:"sensitive_points"`
	Attributes         map[string]string `json:"attributes"`
	ImageUnderstanding string            `json:"image_understanding"`
}

type ComplianceAgentOutput struct {
	Level                string                  `json:"level"`
	Score                int                     `json:"score"`
	Summary              string                  `json:"summary"`
	RiskReasons          []string                `json:"risk_reasons"`
	MatchedRules         []models.ComplianceRule `json:"matched_rules"`
	ForbiddenExpressions []string                `json:"forbidden_expressions"`
	SaferExpressions     []string                `json:"safer_expressions"`
	Suggestions          []string                `json:"suggestions"`
	Disclaimer           string                  `json:"disclaimer"`
	ShouldContinue       bool                    `json:"should_continue"`
}

type LocalizationAgentOutput struct {
	LanguageStyle          string   `json:"language_style"`
	CulturalAdaptation     []string `json:"cultural_adaptation"`
	VisualStyleAdvice      []string `json:"visual_style_advice"`
	VoiceoverStyleAdvice   []string `json:"voiceover_style_advice"`
	LocalConsumerConcerns  []string `json:"local_consumer_concerns"`
	LocalizedSellingPoints []string `json:"localized_selling_points"`
	Tone                   string   `json:"tone"`
}

type ContentAgentOutput struct {
	MarketingTitles    []string      `json:"marketing_titles"`
	SellingPointCopy   []string      `json:"selling_point_copy"`
	ShortVideoScript   VideoScript   `json:"short_video_script"`
	DigitalHumanScript string        `json:"digital_human_script"`
	DigitalHumanPlan   DigitalHuman  `json:"digital_human_plan"`
	PromotionAdvice    PromotionPlan `json:"promotion_advice"`
	ContentWarnings    []string      `json:"content_warnings"`
}

type SilkroadAgentWorkflowResult struct {
	SessionID      uint                 `json:"session_id,omitempty"`
	Result         SilkroadAgentResult  `json:"result"`
	Traces         []models.AgentTrace  `json:"traces"`
	Critic         *models.CriticResult `json:"critic,omitempty"`
	Revised        bool                 `json:"revised"`
	WorkflowStatus string               `json:"workflow_status"`
}

type workflowStageContext struct {
	Input              SilkroadAgentInput       `json:"input"`
	ImageUnderstanding string                   `json:"image_understanding,omitempty"`
	Planning           *PlanningAgentOutput     `json:"planning,omitempty"`
	Product            *ProductAgentOutput      `json:"product,omitempty"`
	ComplianceRules    []models.ComplianceRule  `json:"compliance_rules,omitempty"`
	Compliance         *ComplianceAgentOutput   `json:"compliance,omitempty"`
	Localization       *LocalizationAgentOutput `json:"localization,omitempty"`
	Content            *ContentAgentOutput      `json:"content,omitempty"`
	Critic             *models.CriticResult     `json:"critic,omitempty"`
	RevisionAdvice     []string                 `json:"revision_advice,omitempty"`
}

func (s *SilkroadAgentService) GenerateWithWorkflow(input SilkroadAgentInput) (*SilkroadAgentWorkflowResult, error) {
	input = extractSilkroadInput(input)
	settings := s.readSettings()
	canUseModel := settings.APIKey != "" && settings.TextModel != ""
	modelName := settings.TextModel
	if !canUseModel {
		modelName = "local-workflow-fallback"
	}

	imageUnderstanding := s.workflowImageUnderstanding(input)
	traces := make([]models.AgentTrace, 0, 8)
	usedFallback := !canUseModel

	planning := s.runPlanningAgent(settings, canUseModel, input, imageUnderstanding, &traces, &usedFallback)
	product := s.runProductAgent(settings, canUseModel, input, imageUnderstanding, planning, &traces, &usedFallback)

	ruleService := NewComplianceRuleService(s.log)
	keywords := workflowComplianceKeywords(input, product)
	matchedRules := ruleService.SearchRules(input.TargetMarket, input.TargetPlatform, product.Category, keywords)
	compliance := s.runComplianceAgent(settings, canUseModel, input, product, matchedRules, &traces, &usedFallback)
	localization := s.runLocalizationAgent(settings, canUseModel, input, product, compliance, &traces, &usedFallback)
	content := s.runContentAgent(settings, canUseModel, input, product, compliance, localization, nil, &traces, &usedFallback)

	critic := s.runCriticAgent(settings, canUseModel, input, product, compliance, localization, content, &traces, &usedFallback)
	revised := false
	if critic != nil && critic.NeedRevise {
		revised = true
		content = s.runContentAgent(settings, canUseModel, input, product, compliance, localization, critic.RevisionAdvice, &traces, &usedFallback)
		critic = s.runCriticAgent(settings, canUseModel, input, product, compliance, localization, content, &traces, &usedFallback)
	}

	result := buildWorkflowAgentResult(input, imageUnderstanding, product, compliance, localization, content)
	result.Model = modelName
	result.IsMock = usedFallback
	if usedFallback && result.ErrorMessage == "" {
		result.ErrorMessage = "部分 Agent 阶段使用本地规则兜底，建议配置模型后重新生成。"
	}

	status := "completed"
	if usedFallback {
		status = "completed_with_fallback"
	}
	return &SilkroadAgentWorkflowResult{
		Result:         *result,
		Traces:         traces,
		Critic:         critic,
		Revised:        revised,
		WorkflowStatus: status,
	}, nil
}

func (s *SilkroadAgentService) workflowImageUnderstanding(input SilkroadAgentInput) string {
	if !hasUsableAgentImage(input.ImageDataURL) {
		return ""
	}
	visionSettings := s.readVisionSettings()
	if visionSettings.APIKey == "" || visionSettings.VisionModel == "" {
		return "已收到商品图片；当前未配置视觉模型，工作流主要依据文本信息分析。"
	}
	text, err := s.callVision(visionSettings, visionSettings.VisionModel, input)
	if err != nil {
		if s.log != nil {
			s.log.Warnw("silkroad workflow vision analysis failed", "error", err)
		}
		return "图片分析暂不可用，工作流已主要依据文本信息生成方案。"
	}
	return strings.TrimSpace(text)
}

func (s *SilkroadAgentService) runPlanningAgent(settings silkroadAgentSettings, canUseModel bool, input SilkroadAgentInput, imageUnderstanding string, traces *[]models.AgentTrace, usedFallback *bool) *PlanningAgentOutput {
	fallback := localPlanningAgent(input, imageUnderstanding)
	payload := workflowStageContext{Input: input, ImageUnderstanding: imageUnderstanding}
	output := fallback
	trace := startWorkflowTrace("PlanningAgent", "planning", payload)
	status := "completed"
	var stageErr error
	if canUseModel {
		stageErr = s.callWorkflowStageJSON(settings, "PlanningAgent", planningAgentPrompt(), payload, &output)
		if stageErr != nil {
			*usedFallback = true
			status = "fallback"
			output = fallback
		}
	} else {
		status = "fallback"
	}
	finishWorkflowTrace(&trace, output, status, stageErr)
	*traces = append(*traces, trace)
	return &output
}

func (s *SilkroadAgentService) runProductAgent(settings silkroadAgentSettings, canUseModel bool, input SilkroadAgentInput, imageUnderstanding string, planning *PlanningAgentOutput, traces *[]models.AgentTrace, usedFallback *bool) *ProductAgentOutput {
	fallback := localProductAgent(input, imageUnderstanding)
	payload := workflowStageContext{Input: input, ImageUnderstanding: imageUnderstanding, Planning: planning}
	output := fallback
	trace := startWorkflowTrace("ProductAgent", "product_understanding", payload)
	status := "completed"
	var stageErr error
	if canUseModel {
		stageErr = s.callWorkflowStageJSON(settings, "ProductAgent", productAgentPrompt(), payload, &output)
		if stageErr != nil {
			*usedFallback = true
			status = "fallback"
			output = fallback
		}
	} else {
		status = "fallback"
	}
	output = fillProductAgentOutput(output, fallback)
	finishWorkflowTrace(&trace, output, status, stageErr)
	*traces = append(*traces, trace)
	return &output
}

func (s *SilkroadAgentService) runComplianceAgent(settings silkroadAgentSettings, canUseModel bool, input SilkroadAgentInput, product *ProductAgentOutput, rules []models.ComplianceRule, traces *[]models.AgentTrace, usedFallback *bool) *ComplianceAgentOutput {
	fallback := localComplianceAgent(input, product, rules)
	payload := workflowStageContext{Input: input, Product: product, ComplianceRules: rules}
	output := fallback
	trace := startWorkflowTrace("ComplianceAgent", "compliance_rag", payload)
	status := "completed"
	var stageErr error
	if canUseModel {
		stageErr = s.callWorkflowStageJSON(settings, "ComplianceAgent", complianceAgentPrompt(), payload, &output)
		if stageErr != nil {
			*usedFallback = true
			status = "fallback"
			output = fallback
		}
	} else {
		status = "fallback"
	}
	output = fillComplianceAgentOutput(output, fallback, rules)
	finishWorkflowTrace(&trace, output, status, stageErr)
	*traces = append(*traces, trace)
	return &output
}

func (s *SilkroadAgentService) runLocalizationAgent(settings silkroadAgentSettings, canUseModel bool, input SilkroadAgentInput, product *ProductAgentOutput, compliance *ComplianceAgentOutput, traces *[]models.AgentTrace, usedFallback *bool) *LocalizationAgentOutput {
	fallback := localLocalizationAgent(input, product, compliance)
	payload := workflowStageContext{Input: input, Product: product, Compliance: compliance}
	output := fallback
	trace := startWorkflowTrace("LocalizationAgent", "localization", payload)
	status := "completed"
	var stageErr error
	if canUseModel {
		stageErr = s.callWorkflowStageJSON(settings, "LocalizationAgent", localizationAgentPrompt(), payload, &output)
		if stageErr != nil {
			*usedFallback = true
			status = "fallback"
			output = fallback
		}
	} else {
		status = "fallback"
	}
	output = fillLocalizationAgentOutput(output, fallback)
	finishWorkflowTrace(&trace, output, status, stageErr)
	*traces = append(*traces, trace)
	return &output
}

func (s *SilkroadAgentService) runContentAgent(settings silkroadAgentSettings, canUseModel bool, input SilkroadAgentInput, product *ProductAgentOutput, compliance *ComplianceAgentOutput, localization *LocalizationAgentOutput, revisionAdvice []string, traces *[]models.AgentTrace, usedFallback *bool) *ContentAgentOutput {
	fallback := localContentAgent(input, product, compliance, localization, revisionAdvice)
	stage := "content_generation"
	if len(revisionAdvice) > 0 {
		stage = "content_revision"
	}
	payload := workflowStageContext{Input: input, Product: product, Compliance: compliance, Localization: localization, RevisionAdvice: revisionAdvice}
	output := fallback
	trace := startWorkflowTrace("ContentAgent", stage, payload)
	status := "completed"
	var stageErr error
	if canUseModel {
		stageErr = s.callWorkflowStageJSON(settings, "ContentAgent", contentAgentPrompt(len(revisionAdvice) > 0), payload, &output)
		if stageErr != nil {
			*usedFallback = true
			status = "fallback"
			output = fallback
		}
	} else {
		status = "fallback"
	}
	output = fillContentAgentOutput(output, fallback)
	finishWorkflowTrace(&trace, output, status, stageErr)
	*traces = append(*traces, trace)
	return &output
}

func (s *SilkroadAgentService) runCriticAgent(settings silkroadAgentSettings, canUseModel bool, input SilkroadAgentInput, product *ProductAgentOutput, compliance *ComplianceAgentOutput, localization *LocalizationAgentOutput, content *ContentAgentOutput, traces *[]models.AgentTrace, usedFallback *bool) *models.CriticResult {
	fallback := localCriticAgent(product, compliance, localization, content)
	payload := workflowStageContext{Input: input, Product: product, Compliance: compliance, Localization: localization, Content: content}
	output := fallback
	trace := startWorkflowTrace("CriticAgent", "critic_review", payload)
	status := "completed"
	var stageErr error
	if canUseModel {
		stageErr = s.callWorkflowStageJSON(settings, "CriticAgent", criticAgentPrompt(), payload, &output)
		if stageErr != nil {
			*usedFallback = true
			status = "fallback"
			output = fallback
		}
	} else {
		status = "fallback"
	}
	output = fillCriticResult(output)
	finishWorkflowTrace(&trace, output, status, stageErr)
	*traces = append(*traces, trace)
	return &output
}

func (s *SilkroadAgentService) callWorkflowStageJSON(settings silkroadAgentSettings, agentName, systemPrompt string, payload interface{}, output interface{}) error {
	userPayload, _ := json.MarshalIndent(payload, "", "  ")
	raw, err := s.sendChat(settings, llmChatRequest{
		Model: settings.TextModel,
		Messages: []llmMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: string(userPayload)},
		},
		Temperature:    0.25,
		MaxTokens:      1800,
		ResponseFormat: &llmResponseFormat{Type: "json_object"},
	})
	if err != nil {
		return err
	}
	if err := parseWorkflowJSON(raw, output); err != nil {
		return fmt.Errorf("%s returned invalid json: %w", agentName, err)
	}
	return nil
}

func parseWorkflowJSON(raw string, output interface{}) error {
	jsonText := extractFirstJSONObject(raw)
	if jsonText == "" {
		return errors.New("no json object found")
	}
	return json.Unmarshal([]byte(jsonText), output)
}

func startWorkflowTrace(agentName, stage string, input interface{}) models.AgentTrace {
	now := time.Now()
	return models.AgentTrace{
		AgentName: agentName,
		Stage:     stage,
		Input:     input,
		Status:    "running",
		StartedAt: now,
	}
}

func finishWorkflowTrace(trace *models.AgentTrace, output interface{}, status string, err error) {
	trace.EndedAt = time.Now()
	trace.DurationMs = trace.EndedAt.Sub(trace.StartedAt).Milliseconds()
	trace.Output = output
	trace.Status = status
	if err != nil {
		trace.Error = err.Error()
	}
}

func planningAgentPrompt() string {
	return `你是 PlanningAgent，负责把跨境电商营销任务拆成可执行任务链。
只输出 JSON，不要 Markdown。字段：task_chain, execution_steps, missing_information, final_output_plan, recommended_workspaces。`
}

func productAgentPrompt() string {
	return `你是 ProductAgent，负责理解商品本身。
只输出 JSON，不要 Markdown。字段：product_name, category, core_selling_points, usage_scenarios, target_users, sensitive_points, attributes, image_understanding。不要做最终合规结论。`
}

func complianceAgentPrompt() string {
	return `你是 ComplianceAgent，负责基于输入商品和 compliance_rules 做合规辅助判断。
必须把命中的规则作为依据，不要只自由发挥。只输出 JSON，不要 Markdown。
字段：level(低风险/中风险/高风险), score(0-100), summary, risk_reasons, matched_rules, forbidden_expressions, safer_expressions, suggestions, disclaimer, should_continue。
disclaimer 必须包含：本结果仅用于跨境电商营销合规辅助，不构成法律意见。`
}

func localizationAgentPrompt() string {
	return `你是 LocalizationAgent，负责目标市场本地化策略。
只输出 JSON，不要 Markdown。字段：language_style, cultural_adaptation, visual_style_advice, voiceover_style_advice, local_consumer_concerns, localized_selling_points, tone。`
}

func contentAgentPrompt(revision bool) string {
	if revision {
		return `你是 ContentAgent，负责根据 CriticAgent 修改建议生成修订版跨境营销内容。
只输出 JSON，不要 Markdown。字段：marketing_titles, selling_point_copy, short_video_script, digital_human_script, digital_human_plan, promotion_advice, content_warnings。必须规避合规建议中的高风险表达。`
	}
	return `你是 ContentAgent，负责生成跨境营销标题、卖点文案、短视频脚本、数字人口播稿和投放建议。
只输出 JSON，不要 Markdown。字段：marketing_titles, selling_point_copy, short_video_script, digital_human_script, digital_human_plan, promotion_advice, content_warnings。必须规避合规建议中的高风险表达。`
}

func criticAgentPrompt() string {
	return `你是 CriticAgent，负责评审 ContentAgent 输出。
只输出 JSON，不要 Markdown。字段：completeness_score, compliance_score, localization_score, marketing_score, overall_score, problems, revision_advice, need_revise。
评分范围 1-5。overall_score < 4 或 compliance_score < 4 或存在明显合规问题时 need_revise=true。`
}

func localPlanningAgent(input SilkroadAgentInput, imageUnderstanding string) PlanningAgentOutput {
	missing := []string{}
	if input.TargetMarket == "" {
		missing = append(missing, "目标国家/地区")
	}
	if input.MaterialSpec == "" {
		missing = append(missing, "材质/成分/规格")
	}
	if imageUnderstanding == "" && !hasUsableAgentImage(input.ImageDataURL) {
		missing = append(missing, "商品图片或包装信息")
	}
	return PlanningAgentOutput{
		TaskChain:             []string{"商品理解", "合规知识检索", "合规分析", "本地化策略", "营销内容生成", "Critic 评审", "工作台沉淀"},
		ExecutionSteps:        []string{"抽取商品字段", "匹配合规规则", "生成风险边界", "规划本地化卖点", "生成脚本和数字人口播", "根据评分决定是否修订"},
		MissingInformation:    missing,
		FinalOutputPlan:       []string{"商品结构化结果", "合规风险与命中规则", "本地化策略", "短视频脚本", "数字人口播", "投放建议"},
		RecommendedWorkspaces: []string{"/compliance", "/workspace/script", "/workspace/content", "/workspace/timeline"},
	}
}

func localProductAgent(input SilkroadAgentInput, imageUnderstanding string) ProductAgentOutput {
	return ProductAgentOutput{
		ProductName:       firstNonBlank(input.ProductName, "待分析商品"),
		Category:          firstNonBlank(input.Category, inferCategory(input.ProductName)),
		CoreSellingPoints: append([]string{}, input.CoreSellingPoints...),
		UsageScenarios:    cleanStringList([]string{input.UsageScenario}),
		TargetUsers:       cleanStringList([]string{input.TargetAudience}),
		SensitivePoints:   workflowSensitivePoints(input),
		Attributes: map[string]string{
			"material_spec":   input.MaterialSpec,
			"target_market":   input.TargetMarket,
			"target_platform": input.TargetPlatform,
		},
		ImageUnderstanding: firstNonBlank(imageUnderstanding, "当前主要依据文本信息理解商品。"),
	}
}

func workflowSensitivePoints(input SilkroadAgentInput) []string {
	text := strings.ToLower(strings.Join([]string{input.ProductName, input.Category, input.MaterialSpec, input.RawPrompt, strings.Join(input.CoreSellingPoints, " ")}, " "))
	points := []string{}
	if hasAny(text, []string{"治疗", "medical", "药", "功效", "美白", "祛斑"}) {
		points = append(points, "可能涉及医疗化或功效宣称")
	}
	if hasAny(text, []string{"婴儿", "儿童", "baby", "kids"}) {
		points = append(points, "可能涉及儿童产品安全与年龄标签")
	}
	if hasAny(text, []string{"电池", "battery", "锂", "充电"}) {
		points = append(points, "可能涉及电池运输和电气安全")
	}
	if len(points) == 0 {
		points = append(points, "需避免绝对化宣传和未经核实认证表述")
	}
	return points
}

func localComplianceAgent(input SilkroadAgentInput, product *ProductAgentOutput, rules []models.ComplianceRule) ComplianceAgentOutput {
	score := 22
	riskReasons := []string{}
	for _, rule := range rules {
		riskReasons = append(riskReasons, fmt.Sprintf("命中规则 %s：%s", rule.ID, rule.RuleText))
		score += 8
		if strings.Contains(rule.RiskType, "儿童") || strings.Contains(rule.RiskType, "医疗") || strings.Contains(rule.RiskType, "功效") {
			score += 8
		}
	}
	for _, point := range product.SensitivePoints {
		riskReasons = append(riskReasons, point)
		score += 6
	}
	score = clampScore(score)
	level := "低风险"
	switch {
	case score >= 70:
		level = "高风险"
	case score >= 40:
		level = "中风险"
	}
	forbidden, safer := expressionsFromRules(rules)
	if len(forbidden) == 0 {
		forbidden = []string{"100%安全", "治疗/治愈", "官方认证", "保证通过"}
	}
	if len(safer) == 0 {
		safer = []string{"适合日常使用", "建议查看材质说明", "以实际检测材料为准"}
	}
	if len(riskReasons) == 0 {
		riskReasons = []string{"未命中显著高风险规则，仍建议人工复核目标市场平台政策。"}
	}
	return ComplianceAgentOutput{
		Level:                level,
		Score:                score,
		Summary:              fmt.Sprintf("已结合本地合规知识库完成辅助评估，当前为%s。", level),
		RiskReasons:          uniqueStrings(riskReasons),
		MatchedRules:         rules,
		ForbiddenExpressions: forbidden,
		SaferExpressions:     safer,
		Suggestions:          []string{"补充材质、标签、认证或检测材料。", "广告文案避免绝对化、医疗化和未经证实认证表达。", "上架前结合目标国家法规和平台政策复核。"},
		Disclaimer:           ComplianceAssistantDisclaimer,
		ShouldContinue:       score < 80,
	}
}

func expressionsFromRules(rules []models.ComplianceRule) ([]string, []string) {
	forbidden := []string{}
	safer := []string{}
	for _, rule := range rules {
		forbidden = append(forbidden, rule.ForbiddenExpressions...)
		safer = append(safer, rule.SaferExpressions...)
	}
	return uniqueStrings(forbidden), uniqueStrings(safer)
}

func localLocalizationAgent(input SilkroadAgentInput, product *ProductAgentOutput, compliance *ComplianceAgentOutput) LocalizationAgentOutput {
	market := firstNonBlank(input.TargetMarket, "目标市场")
	points := product.CoreSellingPoints
	if len(points) == 0 {
		points = inferSellingPoints(product.ProductName, product.Category)
	}
	return LocalizationAgentOutput{
		LanguageStyle:          fmt.Sprintf("%s消费者易理解的自然口语表达", market),
		CulturalAdaptation:     []string{"用真实生活场景承接卖点", "避免过度夸张和敏感文化符号", "保留材质、规格和使用限制信息"},
		VisualStyleAdvice:      []string{"产品特写", "开箱展示", "日常使用场景", "字幕强调安全表达"},
		VoiceoverStyleAdvice:   []string{"语速自然", "少用绝对化词汇", "用体验描述替代功效承诺"},
		LocalConsumerConcerns:  []string{"价格与耐用性", "材质与安全", "物流与售后", "真实使用反馈"},
		LocalizedSellingPoints: points,
		Tone:                   "自然、可信、克制",
	}
}

func localContentAgent(input SilkroadAgentInput, product *ProductAgentOutput, compliance *ComplianceAgentOutput, localization *LocalizationAgentOutput, revisionAdvice []string) ContentAgentOutput {
	productName := firstNonBlank(product.ProductName, input.ProductName, "这款商品")
	points := product.CoreSellingPoints
	if len(points) == 0 {
		points = inferSellingPoints(productName, product.Category)
	}
	prefix := ""
	if len(revisionAdvice) > 0 {
		prefix = "修订版："
	}
	return ContentAgentOutput{
		MarketingTitles: []string{
			fmt.Sprintf("%s%s的日常场景实用指南", prefix, productName),
			fmt.Sprintf("%s用真实体验了解%s", prefix, productName),
		},
		SellingPointCopy: []string{
			fmt.Sprintf("围绕%s，展示可见卖点和适用场景。", strings.Join(points, "、")),
			"表达以材质、使用方式和真实体验为主，避免高风险功效承诺。",
		},
		ShortVideoScript: VideoScript{
			Title:    fmt.Sprintf("%s%s短视频脚本", prefix, productName),
			Duration: "20-25s",
			Opening:  ScriptSegment{Time: "0-3s", Content: fmt.Sprintf("用目标用户的日常小困扰切入，快速露出%s。", productName)},
			Middle:   ScriptSegment{Time: "3-20s", Content: fmt.Sprintf("展示%s、使用步骤和%s场景，字幕保留材质/规格提醒。", strings.Join(points, "、"), firstNonBlank(input.UsageScenario, "真实使用"))},
			Ending:   ScriptSegment{Time: "20-25s", Content: "用温和 CTA 引导查看详情、认证材料和用户评价。"},
			Storyboard: []StoryboardShot{
				{Shot: "镜头 1", Visual: "商品与目标用户场景同框", Voiceover: fmt.Sprintf("如果你也在找一款更顺手的%s，可以先看这个场景。", productName), Subtitle: "Start with real use"},
				{Shot: "镜头 2", Visual: "产品细节、材质或包装标签特写", Voiceover: "重点看材质、规格和日常使用方式。", Subtitle: "Check details before purchase"},
				{Shot: "镜头 3", Visual: "自然使用反馈和详情页引导", Voiceover: "想了解更多，建议查看详情和适用说明。", Subtitle: "See details"},
			},
		},
		DigitalHumanScript: fmt.Sprintf("这款%s更适合用真实场景介绍。我们先看材质和规格，再看它在日常里的使用方式。购买前也建议查看详情页的说明。", productName),
		DigitalHumanPlan: DigitalHuman{
			Persona:        "本地生活方式讲解型数字人",
			Tone:           localization.Tone,
			VideoRatio:     "9:16",
			SubtitleAdvice: "目标市场语言字幕 + 卖点短词 + 合规提醒",
			VisualStyle:    strings.Join(localization.VisualStyleAdvice, " / "),
			ShootingStyle:  "产品近景 + 场景演示 + 字幕重点提示",
		},
		PromotionAdvice: PromotionPlan{
			Platforms:          []string{firstNonBlank(input.TargetPlatform, "TikTok Shop")},
			ContentTags:        []string{"场景种草", "开箱演示", "产品细节", "本地化字幕"},
			FocusMetrics:       []string{"完播率", "点击率", "收藏率", "评论问题"},
			OptimizationAdvice: "先用不同开头做小流量测试，再根据评论集中补充材质、尺寸、认证或使用限制信息。",
		},
		ContentWarnings: compliance.ForbiddenExpressions,
	}
}

func localCriticAgent(product *ProductAgentOutput, compliance *ComplianceAgentOutput, localization *LocalizationAgentOutput, content *ContentAgentOutput) models.CriticResult {
	problems := []string{}
	advice := []string{}
	completeness := 5
	if product.ProductName == "" || content.ShortVideoScript.Opening.Content == "" || content.DigitalHumanScript == "" {
		completeness = 3
		problems = append(problems, "内容结构仍有缺口，需补齐商品、脚本或数字人口播。")
		advice = append(advice, "补齐营销标题、短视频三段式脚本、数字人口播和投放建议。")
	}
	complianceScore := 5
	if compliance.Score >= 60 {
		complianceScore = 3
		problems = append(problems, "合规风险偏高，需要进一步弱化功效和认证表达。")
		advice = append(advice, "删除绝对化、医疗化、未经核实认证和保证性说法。")
	}
	localizationScore := 4
	if len(localization.LocalizedSellingPoints) == 0 {
		localizationScore = 3
		advice = append(advice, "补充目标市场消费者关注点和本地化卖点。")
	}
	marketingScore := 4
	if len(content.MarketingTitles) == 0 || len(content.SellingPointCopy) == 0 {
		marketingScore = 3
		advice = append(advice, "加强前三秒钩子和场景化卖点表达。")
	}
	overall := (completeness + complianceScore + localizationScore + marketingScore) / 4
	needRevise := overall < 4 || complianceScore < 4 || len(problems) > 0 && complianceScore < 5
	if len(problems) == 0 {
		problems = []string{"未发现明显结构性问题，建议上架前人工复核合规依据。"}
	}
	if len(advice) == 0 {
		advice = []string{"保留当前结构，补充目标市场真实素材和平台政策复核。"}
	}
	return models.CriticResult{
		CompletenessScore: completeness,
		ComplianceScore:   complianceScore,
		LocalizationScore: localizationScore,
		MarketingScore:    marketingScore,
		OverallScore:      overall,
		Problems:          uniqueStrings(problems),
		RevisionAdvice:    uniqueStrings(advice),
		NeedRevise:        needRevise,
	}
}

func fillProductAgentOutput(output, fallback ProductAgentOutput) ProductAgentOutput {
	output.ProductName = firstNonBlank(output.ProductName, fallback.ProductName)
	output.Category = firstNonBlank(output.Category, fallback.Category)
	if len(output.CoreSellingPoints) == 0 {
		output.CoreSellingPoints = fallback.CoreSellingPoints
	}
	if len(output.UsageScenarios) == 0 {
		output.UsageScenarios = fallback.UsageScenarios
	}
	if len(output.TargetUsers) == 0 {
		output.TargetUsers = fallback.TargetUsers
	}
	if len(output.SensitivePoints) == 0 {
		output.SensitivePoints = fallback.SensitivePoints
	}
	if len(output.Attributes) == 0 {
		output.Attributes = fallback.Attributes
	}
	output.ImageUnderstanding = firstNonBlank(output.ImageUnderstanding, fallback.ImageUnderstanding)
	return output
}

func fillComplianceAgentOutput(output, fallback ComplianceAgentOutput, rules []models.ComplianceRule) ComplianceAgentOutput {
	output.Level = firstNonBlank(output.Level, fallback.Level)
	if output.Score == 0 {
		output.Score = fallback.Score
	}
	output.Score = clampScore(output.Score)
	output.Summary = firstNonBlank(output.Summary, fallback.Summary)
	if len(output.RiskReasons) == 0 {
		output.RiskReasons = fallback.RiskReasons
	}
	if len(output.MatchedRules) == 0 {
		output.MatchedRules = rules
	}
	if len(output.ForbiddenExpressions) == 0 {
		output.ForbiddenExpressions = fallback.ForbiddenExpressions
	}
	if len(output.SaferExpressions) == 0 {
		output.SaferExpressions = fallback.SaferExpressions
	}
	if len(output.Suggestions) == 0 {
		output.Suggestions = fallback.Suggestions
	}
	if output.Disclaimer == "" {
		output.Disclaimer = ComplianceAssistantDisclaimer
	}
	if !output.ShouldContinue {
		output.ShouldContinue = output.Score < 80
	}
	return output
}

func fillLocalizationAgentOutput(output, fallback LocalizationAgentOutput) LocalizationAgentOutput {
	output.LanguageStyle = firstNonBlank(output.LanguageStyle, fallback.LanguageStyle)
	if len(output.CulturalAdaptation) == 0 {
		output.CulturalAdaptation = fallback.CulturalAdaptation
	}
	if len(output.VisualStyleAdvice) == 0 {
		output.VisualStyleAdvice = fallback.VisualStyleAdvice
	}
	if len(output.VoiceoverStyleAdvice) == 0 {
		output.VoiceoverStyleAdvice = fallback.VoiceoverStyleAdvice
	}
	if len(output.LocalConsumerConcerns) == 0 {
		output.LocalConsumerConcerns = fallback.LocalConsumerConcerns
	}
	if len(output.LocalizedSellingPoints) == 0 {
		output.LocalizedSellingPoints = fallback.LocalizedSellingPoints
	}
	output.Tone = firstNonBlank(output.Tone, fallback.Tone)
	return output
}

func fillContentAgentOutput(output, fallback ContentAgentOutput) ContentAgentOutput {
	if len(output.MarketingTitles) == 0 {
		output.MarketingTitles = fallback.MarketingTitles
	}
	if len(output.SellingPointCopy) == 0 {
		output.SellingPointCopy = fallback.SellingPointCopy
	}
	if output.ShortVideoScript.Opening.Content == "" {
		output.ShortVideoScript = fallback.ShortVideoScript
	}
	if output.DigitalHumanScript == "" {
		output.DigitalHumanScript = fallback.DigitalHumanScript
	}
	if output.DigitalHumanPlan.Persona == "" {
		output.DigitalHumanPlan = fallback.DigitalHumanPlan
	}
	if len(output.PromotionAdvice.Platforms) == 0 {
		output.PromotionAdvice = fallback.PromotionAdvice
	}
	if len(output.ContentWarnings) == 0 {
		output.ContentWarnings = fallback.ContentWarnings
	}
	return output
}

func fillCriticResult(result models.CriticResult) models.CriticResult {
	result.CompletenessScore = clampCriticScore(result.CompletenessScore)
	result.ComplianceScore = clampCriticScore(result.ComplianceScore)
	result.LocalizationScore = clampCriticScore(result.LocalizationScore)
	result.MarketingScore = clampCriticScore(result.MarketingScore)
	if result.OverallScore == 0 {
		result.OverallScore = (result.CompletenessScore + result.ComplianceScore + result.LocalizationScore + result.MarketingScore) / 4
	}
	result.OverallScore = clampCriticScore(result.OverallScore)
	result.Problems = cleanStringList(result.Problems)
	result.RevisionAdvice = cleanStringList(result.RevisionAdvice)
	if len(result.Problems) == 0 {
		result.Problems = []string{"未发现明显结构性问题，建议人工复核合规依据。"}
	}
	if len(result.RevisionAdvice) == 0 {
		result.RevisionAdvice = []string{"保留当前结构，补充目标市场真实素材和平台政策复核。"}
	}
	if result.OverallScore < 4 || result.ComplianceScore < 4 {
		result.NeedRevise = true
	}
	return result
}

func clampCriticScore(score int) int {
	if score < 1 {
		return 1
	}
	if score > 5 {
		return 5
	}
	return score
}

func buildWorkflowAgentResult(input SilkroadAgentInput, imageUnderstanding string, product *ProductAgentOutput, compliance *ComplianceAgentOutput, localization *LocalizationAgentOutput, content *ContentAgentOutput) *SilkroadAgentResult {
	result := &SilkroadAgentResult{
		RecognizedInfo: RecognizedInfo{
			ProductName:        firstNonBlank(product.ProductName, input.ProductName, "待分析商品"),
			Category:           firstNonBlank(product.Category, input.Category, "跨境电商商品"),
			TargetMarket:       input.TargetMarket,
			TargetPlatform:     firstNonBlank(input.TargetPlatform, "目标平台待补充"),
			TargetAudience:     firstNonBlank(strings.Join(product.TargetUsers, " / "), input.TargetAudience, "目标人群待补充"),
			CoreSellingPoints:  product.CoreSellingPoints,
			ImageUnderstanding: firstNonBlank(product.ImageUnderstanding, imageUnderstanding, "当前主要依据文本信息生成。"),
		},
		Overview: AgentOverview{
			ComplianceRiskLevel:     compliance.Level,
			MarketStrategy:          strings.Join(localization.CulturalAdaptation, "；"),
			RecommendedVideoStyle:   firstNonBlank(strings.Join(localization.VisualStyleAdvice, " / "), "生活场景化竖屏短视频"),
			RecommendedDigitalHuman: firstNonBlank(content.DigitalHumanPlan.Persona, "本地生活方式讲解型数字人"),
		},
		Compliance: CompliancePlan{
			Title:                "合规分析结果",
			Summary:              compliance.Summary,
			RiskTags:             compliance.RiskReasons,
			MissingInfo:          workflowMissingInfo(input),
			Suggestions:          compliance.Suggestions,
			ForbiddenExpressions: compliance.ForbiddenExpressions,
			SaferExpressions:     compliance.SaferExpressions,
			Level:                compliance.Level,
			Score:                compliance.Score,
			RiskReasons:          compliance.RiskReasons,
			MatchedRules:         compliance.MatchedRules,
			Disclaimer:           compliance.Disclaimer,
		},
		Localization: Localization{
			Direction:        localization.LanguageStyle,
			Reason:           strings.Join(localization.CulturalAdaptation, "；"),
			Keywords:         localization.LocalizedSellingPoints,
			Tone:             localization.Tone,
			SceneSuggestions: localization.VisualStyleAdvice,
		},
		Script:       content.ShortVideoScript,
		DigitalHuman: content.DigitalHumanPlan,
		Promotion:    content.PromotionAdvice,
		AgentMessage: AgentMessage{
			Summary:           firstNonBlank(firstWorkflowString(content.SellingPointCopy), "已完成多 Agent 工作流方案生成。"),
			MissingInfoNotice: strings.Join(workflowMissingInfo(input), "、"),
			QuickActions:      []string{"一键创建营销项目", "进入合规分析", "进入脚本工作台", "生成数字人口播", "进入剪辑时间线"},
		},
	}
	return fillAgentDefaults(*result, input, imageUnderstanding)
}

func firstWorkflowString(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func workflowMissingInfo(input SilkroadAgentInput) []string {
	missing := []string{}
	if input.TargetMarket == "" {
		missing = append(missing, "目标国家/地区")
	}
	if input.MaterialSpec == "" {
		missing = append(missing, "材质/成分/规格")
	}
	if input.UsageScenario == "" {
		missing = append(missing, "使用场景")
	}
	if len(missing) == 0 {
		missing = append(missing, "认证或检测材料", "目标售价区间")
	}
	return missing
}

func workflowComplianceKeywords(input SilkroadAgentInput, product *ProductAgentOutput) []string {
	values := []string{
		input.ProductName,
		input.Category,
		input.MaterialSpec,
		input.UsageScenario,
		input.RawPrompt,
		product.ProductName,
		product.Category,
		strings.Join(product.CoreSellingPoints, " "),
		strings.Join(product.SensitivePoints, " "),
	}
	values = append(values, input.CoreSellingPoints...)
	return append(values, extractComplianceKeywords(strings.Join(values, " "))...)
}
