package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateProjectFromAgentRequest struct {
	Result   *SilkroadAgentResult         `json:"result"`
	Workflow *SilkroadAgentWorkflowResult `json:"workflow"`
}

type CreateProjectFromAgentResponse struct {
	ProjectID   uint                   `json:"project_id"`
	EpisodeID   uint                   `json:"episode_id,omitempty"`
	Path        string                 `json:"path"`
	Summary     string                 `json:"summary"`
	CreatedFrom string                 `json:"created_from"`
	Project     *models.Drama          `json:"project,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type SilkroadAgentProjectService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewSilkroadAgentProjectService(db *gorm.DB, log *logger.Logger) *SilkroadAgentProjectService {
	return &SilkroadAgentProjectService{db: db, log: log}
}

func (s *SilkroadAgentProjectService) CreateFromAgentResult(deviceID string, result *SilkroadAgentResult, workflow *SilkroadAgentWorkflowResult) (*CreateProjectFromAgentResponse, error) {
	if result == nil && workflow != nil {
		result = &workflow.Result
	}
	if result == nil {
		return nil, fmt.Errorf("agent result is required")
	}

	now := time.Now()
	title := firstNonBlank(result.RecognizedInfo.ProductName, result.Script.Title, "Agent 营销项目")
	description := buildAgentProjectDescription(result)
	complianceJSON, _ := json.Marshal(map[string]interface{}{
		"source":        "silkroad_agent",
		"risk_level":    result.Overview.ComplianceRiskLevel,
		"summary":       result.Compliance.Summary,
		"risk_reasons":  result.Compliance.RiskReasons,
		"risk_tags":     result.Compliance.RiskTags,
		"suggestions":   result.Compliance.Suggestions,
		"matched_rules": result.Compliance.MatchedRules,
		"disclaimer":    firstNonBlank(result.Compliance.Disclaimer, ComplianceAssistantDisclaimer),
	})
	metadataJSON, _ := json.Marshal(buildAgentProjectMetadata(result, workflow))
	tagsJSON, _ := json.Marshal(uniqueStrings(append([]string{
		result.RecognizedInfo.Category,
		result.RecognizedInfo.TargetPlatform,
		result.Overview.ComplianceRiskLevel,
	}, result.Promotion.ContentTags...)))

	var createdDrama models.Drama
	var createdEpisode models.Episode
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		drama := models.Drama{
			DeviceID:               strings.TrimSpace(deviceID),
			Title:                  title,
			TargetCountry:          normalizeAgentProjectCountry(result.RecognizedInfo.TargetMarket),
			ComplianceScore:        agentComplianceScore(result),
			ComplianceLevel:        agentComplianceLevel(result),
			ComplianceReport:       datatypes.JSON(complianceJSON),
			Genre:                  ptrStringIfNotBlank(result.RecognizedInfo.Category),
			Style:                  "marketing",
			TotalEpisodes:          1,
			Status:                 "draft",
			Tags:                   datatypes.JSON(tagsJSON),
			Metadata:               datatypes.JSON(metadataJSON),
			CreatedAt:              now,
			UpdatedAt:              now,
			Description:            ptrStringIfNotBlank(description),
			MaterialComposition:    ptrStringIfNotBlank(agentMaterialFromResult(result)),
			MarketingSellingPoints: ptrStringIfNotBlank(strings.Join(result.RecognizedInfo.CoreSellingPoints, "、")),
		}
		if err := tx.Create(&drama).Error; err != nil {
			return err
		}
		createdDrama = drama

		scriptContent := formatAgentScriptForWorkspace(result)
		episode := models.Episode{
			DramaID:       drama.ID,
			EpisodeNum:    1,
			Title:         firstNonBlank(result.Script.Title, "跨境营销脚本"),
			ScriptContent: ptrStringIfNotBlank(scriptContent),
			Description:   ptrStringIfNotBlank(result.AgentMessage.Summary),
			Duration:      1,
			Status:        "draft",
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Create(&episode).Error; err != nil {
			return err
		}
		createdEpisode = episode

		character := models.Character{
			DramaID:     drama.ID,
			Name:        "数字人营销表达",
			Role:        ptrStringIfNotBlank("跨境营销讲解者"),
			Description: ptrStringIfNotBlank(result.DigitalHuman.Persona),
			Personality: ptrStringIfNotBlank(result.DigitalHuman.Tone),
			Appearance:  ptrStringIfNotBlank(result.DigitalHuman.VisualStyle),
			VoiceStyle:  ptrStringIfNotBlank(result.DigitalHuman.SubtitleAdvice),
			SortOrder:   1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Create(&character).Error; err != nil {
			return err
		}

		if err := createAgentProjectSceneAndStoryboards(tx, drama.ID, episode.ID, result, now); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if s.log != nil {
			s.log.Warnw("failed to create marketing project from agent", "error", err)
		}
		return nil, err
	}

	path := fmt.Sprintf("/projects/%d", createdDrama.ID)
	return &CreateProjectFromAgentResponse{
		ProjectID:   createdDrama.ID,
		EpisodeID:   createdEpisode.ID,
		Path:        path,
		Summary:     fmt.Sprintf("已创建「%s」营销项目，并带入商品、合规、脚本、数字人与投放建议。", title),
		CreatedFrom: "silkroad_agent",
		Project:     &createdDrama,
		Metadata: map[string]interface{}{
			"target_market": result.RecognizedInfo.TargetMarket,
			"platform":      result.RecognizedInfo.TargetPlatform,
			"risk_level":    result.Overview.ComplianceRiskLevel,
		},
	}, nil
}

func buildAgentProjectDescription(result *SilkroadAgentResult) string {
	parts := []string{
		result.RecognizedInfo.ImageUnderstanding,
		result.Overview.MarketStrategy,
		result.Localization.Direction,
		result.Compliance.Summary,
	}
	return strings.Join(cleanStringList(parts), "\n")
}

func buildAgentProjectMetadata(result *SilkroadAgentResult, workflow *SilkroadAgentWorkflowResult) map[string]interface{} {
	metadata := map[string]interface{}{
		"source":                "silkroad_agent",
		"product_name":          result.RecognizedInfo.ProductName,
		"target_market":         result.RecognizedInfo.TargetMarket,
		"target_platform":       result.RecognizedInfo.TargetPlatform,
		"category":              result.RecognizedInfo.Category,
		"localization":          result.Localization,
		"script":                result.Script,
		"digital_human":         result.DigitalHuman,
		"promotion":             result.Promotion,
		"agent_message":         result.AgentMessage,
		"matched_rules":         result.Compliance.MatchedRules,
		"compliance_disclaimer": firstNonBlank(result.Compliance.Disclaimer, ComplianceAssistantDisclaimer),
	}
	if workflow != nil {
		metadata["workflow_status"] = workflow.WorkflowStatus
		metadata["workflow_revised"] = workflow.Revised
		metadata["critic"] = workflow.Critic
		metadata["agent_traces"] = workflow.Traces
	}
	return metadata
}

func formatAgentScriptForWorkspace(result *SilkroadAgentResult) string {
	lines := []string{
		fmt.Sprintf("标题：%s", firstNonBlank(result.Script.Title, result.RecognizedInfo.ProductName)),
		fmt.Sprintf("时长：%s", firstNonBlank(result.Script.Duration, "20-25s")),
		fmt.Sprintf("[%s] %s", firstNonBlank(result.Script.Opening.Time, "0-3s"), result.Script.Opening.Content),
		fmt.Sprintf("[%s] %s", firstNonBlank(result.Script.Middle.Time, "3-20s"), result.Script.Middle.Content),
		fmt.Sprintf("[%s] %s", firstNonBlank(result.Script.Ending.Time, "20-25s"), result.Script.Ending.Content),
		"",
		"数字人口播：",
		result.DigitalHuman.Persona + "；" + result.DigitalHuman.Tone,
	}
	return strings.Join(lines, "\n")
}

func createAgentProjectSceneAndStoryboards(tx *gorm.DB, dramaID, episodeID uint, result *SilkroadAgentResult, now time.Time) error {
	scenePrompt := firstNonBlank(strings.Join(result.Localization.SceneSuggestions, "；"), result.Localization.Direction, "跨境营销商品展示场景")
	scene := models.Scene{
		DramaID:         dramaID,
		EpisodeID:       &episodeID,
		Location:        "跨境营销内容场景",
		Time:            "day",
		Prompt:          scenePrompt,
		StoryboardCount: maxInt(1, len(result.Script.Storyboard)),
		Status:          "pending",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := tx.Create(&scene).Error; err != nil {
		return err
	}

	shots := result.Script.Storyboard
	if len(shots) == 0 {
		shots = []StoryboardShot{
			{Shot: "镜头 1", Visual: "商品与目标用户场景同框", Voiceover: result.Script.Opening.Content, Subtitle: "开头钩子"},
			{Shot: "镜头 2", Visual: "商品细节与使用步骤", Voiceover: result.Script.Middle.Content, Subtitle: "核心卖点"},
			{Shot: "镜头 3", Visual: "行动引导与详情页提示", Voiceover: result.Script.Ending.Content, Subtitle: "温和 CTA"},
		}
	}
	for index, shot := range shots {
		title := firstNonBlank(shot.Shot, fmt.Sprintf("镜头 %d", index+1))
		storyboard := models.Storyboard{
			EpisodeID:        episodeID,
			SceneID:          &scene.ID,
			StoryboardNumber: index + 1,
			Title:            ptrStringIfNotBlank(title),
			Location:         ptrStringIfNotBlank("跨境营销内容场景"),
			Time:             ptrStringIfNotBlank("day"),
			ShotType:         ptrStringIfNotBlank("产品展示"),
			Action:           ptrStringIfNotBlank(shot.Visual),
			Dialogue:         ptrStringIfNotBlank(shot.Voiceover),
			Description:      ptrStringIfNotBlank(shot.Subtitle),
			ImagePrompt:      ptrStringIfNotBlank(shot.Visual),
			VideoPrompt:      ptrStringIfNotBlank(shot.Voiceover),
			Duration:         5,
			Status:           "pending",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := tx.Create(&storyboard).Error; err != nil {
			return err
		}
	}
	return nil
}

func normalizeAgentProjectCountry(value string) string {
	value = cleanupTargetMarket(value)
	if isUnknownTargetMarket(value) {
		return ""
	}
	return value
}

func agentComplianceScore(result *SilkroadAgentResult) int {
	if result.Compliance.Score > 0 {
		return clampScore(result.Compliance.Score)
	}
	risk := result.Overview.ComplianceRiskLevel
	switch {
	case strings.Contains(risk, "高"):
		return 70
	case strings.Contains(risk, "中"):
		return 48
	default:
		return 22
	}
}

func agentComplianceLevel(result *SilkroadAgentResult) string {
	score := agentComplianceScore(result)
	level, _ := riskLevelByScore(score)
	return string(level)
}

func agentMaterialFromResult(result *SilkroadAgentResult) string {
	return firstNonBlank(result.RecognizedInfo.ImageUnderstanding, strings.Join(result.Compliance.MissingInfo, "、"))
}

func ptrStringIfNotBlank(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
