package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
)

const ComplianceAssistantDisclaimer = "本结果仅用于跨境电商营销合规辅助，不构成法律意见；实际上架与投放前建议结合目标国家法规、平台政策和专业合规意见进行复核。"

type ComplianceRuleService struct {
	log      *logger.Logger
	rulePath string
	once     sync.Once
	rules    []models.ComplianceRule
	loadErr  error
}

func NewComplianceRuleService(log *logger.Logger) *ComplianceRuleService {
	return &ComplianceRuleService{
		log:      log,
		rulePath: filepath.Join("data", "compliance_rules.json"),
	}
}

func (s *ComplianceRuleService) LoadRules() ([]models.ComplianceRule, error) {
	s.once.Do(func() {
		raw, err := os.ReadFile(s.rulePath)
		if err != nil {
			alt := filepath.Join("..", "data", "compliance_rules.json")
			raw, err = os.ReadFile(alt)
		}
		if err != nil {
			s.loadErr = err
			if s.log != nil {
				s.log.Warnw("failed to load compliance rules", "error", err)
			}
			return
		}
		if err := json.Unmarshal(raw, &s.rules); err != nil {
			s.loadErr = err
			if s.log != nil {
				s.log.Warnw("failed to parse compliance rules", "error", err)
			}
			return
		}
	})
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return append([]models.ComplianceRule{}, s.rules...), nil
}

func (s *ComplianceRuleService) SearchRules(country, platform, category string, keywords []string) []models.ComplianceRule {
	rules, err := s.LoadRules()
	if err != nil || len(rules) == 0 {
		return []models.ComplianceRule{}
	}

	targetCountry := normalizeComplianceRuleText(country)
	targetPlatform := normalizeComplianceRuleText(platform)
	targetCategory := normalizeComplianceCategory(category)
	normalizedKeywords := normalizeComplianceKeywords(keywords)

	type scoredRule struct {
		rule  models.ComplianceRule
		score int
	}
	scored := make([]scoredRule, 0, len(rules))
	for _, rule := range rules {
		score := 0
		ruleCountry := normalizeComplianceRuleText(rule.Country)
		rulePlatform := normalizeComplianceRuleText(rule.Platform)
		ruleCategory := normalizeComplianceCategory(rule.Category)

		if targetCountry != "" && ruleCountry == targetCountry {
			score += 40
		} else if ruleCountry == "通用" {
			score += 12
		} else if targetCountry == "" {
			score += 4
		} else {
			continue
		}

		if targetPlatform != "" && rulePlatform == targetPlatform {
			score += 28
		} else if rulePlatform == "通用平台" || rulePlatform == "通用" {
			score += 10
		} else if targetPlatform == "" {
			score += 4
		} else {
			continue
		}

		if targetCategory != "" && ruleCategory == targetCategory {
			score += 24
		} else if ruleCategory == "通用类目" || ruleCategory == "通用" {
			score += 8
		} else if categoryKeywordOverlap(targetCategory, ruleCategory) {
			score += 14
		} else if targetCategory == "" {
			score += 4
		} else {
			continue
		}

		ruleText := normalizeComplianceRuleText(strings.Join([]string{
			rule.RiskType,
			rule.RuleText,
			strings.Join(rule.ForbiddenExpressions, " "),
			strings.Join(rule.SaferExpressions, " "),
		}, " "))
		for _, keyword := range normalizedKeywords {
			if keyword == "" {
				continue
			}
			if strings.Contains(ruleText, keyword) {
				score += 8
			}
			for _, expr := range rule.ForbiddenExpressions {
				if strings.Contains(normalizeComplianceRuleText(expr), keyword) || strings.Contains(keyword, normalizeComplianceRuleText(expr)) {
					score += 12
				}
			}
		}

		if score > 0 {
			scored = append(scored, scoredRule{rule: rule, score: score})
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].rule.ID < scored[j].rule.ID
		}
		return scored[i].score > scored[j].score
	})

	limit := 5
	if len(scored) < limit {
		limit = len(scored)
	}
	result := make([]models.ComplianceRule, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, scored[i].rule)
	}
	return result
}

func (s *ComplianceRuleService) MatchRulesByProduct(productInfo interface{}, targetMarket, platform string) []models.ComplianceRule {
	payload, _ := json.Marshal(productInfo)
	keywords := extractComplianceKeywords(string(payload))
	category := inferComplianceCategoryFromText(string(payload))
	return s.SearchRules(targetMarket, platform, category, keywords)
}

func normalizeComplianceKeywords(keywords []string) []string {
	out := make([]string, 0, len(keywords))
	seen := map[string]struct{}{}
	for _, keyword := range keywords {
		for _, part := range strings.FieldsFunc(keyword, func(r rune) bool {
			return r == ',' || r == '，' || r == ';' || r == '；' || r == '/' || r == '、' || r == '\n'
		}) {
			normalized := normalizeComplianceRuleText(part)
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			out = append(out, normalized)
		}
	}
	return out
}

func extractComplianceKeywords(text string) []string {
	normalized := strings.ToLower(text)
	candidates := []string{
		"电池", "锂", "battery", "充电", "无线", "电子", "usb",
		"美妆", "护肤", "祛斑", "美白", "治疗", "medical", "skin",
		"婴儿", "儿童", "baby", "kids", "母婴",
		"服饰", "面料", "尺码", "塑形", "fashion",
		"家居", "厨房", "承重", "安装", "home",
		"食品", "配料", "过敏原", "supplement", "food",
		"认证", "检测", "官方", "安全", "功效",
	}
	out := make([]string, 0, len(candidates))
	for _, item := range candidates {
		if strings.Contains(normalized, strings.ToLower(item)) {
			out = append(out, item)
		}
	}
	return out
}

func inferComplianceCategoryFromText(text string) string {
	value := normalizeComplianceRuleText(text)
	switch {
	case strings.Contains(value, "电子") || strings.Contains(value, "电池") || strings.Contains(value, "充电") || strings.Contains(value, "usb"):
		return "电子产品"
	case strings.Contains(value, "美妆") || strings.Contains(value, "护肤") || strings.Contains(value, "个护") || strings.Contains(value, "skin"):
		return "美妆个护"
	case strings.Contains(value, "婴儿") || strings.Contains(value, "儿童") || strings.Contains(value, "母婴") || strings.Contains(value, "baby"):
		return "母婴用品"
	case strings.Contains(value, "服饰") || strings.Contains(value, "衣") || strings.Contains(value, "鞋") || strings.Contains(value, "面料") || strings.Contains(value, "fashion"):
		return "服饰"
	case strings.Contains(value, "家居") || strings.Contains(value, "厨房") || strings.Contains(value, "收纳") || strings.Contains(value, "home"):
		return "家居用品"
	default:
		return "通用类目"
	}
}

func normalizeComplianceCategory(category string) string {
	category = normalizeComplianceRuleText(category)
	if category == "" {
		return ""
	}
	return inferComplianceCategoryFromText(category)
}

func categoryKeywordOverlap(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	return strings.Contains(left, right) || strings.Contains(right, left)
}

func normalizeComplianceRuleText(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "United States", "美国")
	value = strings.ReplaceAll(value, "USA", "美国")
	value = strings.ReplaceAll(value, "US", "美国")
	value = strings.ReplaceAll(value, "United Kingdom", "英国")
	value = strings.ReplaceAll(value, "UK", "英国")
	value = strings.ReplaceAll(value, "Malaysia", "马来西亚")
	value = strings.ReplaceAll(value, "Singapore", "新加坡")
	value = strings.ReplaceAll(value, "Saudi Arabia", "沙特")
	value = strings.ReplaceAll(value, "Saudi", "沙特")
	value = strings.ReplaceAll(value, "general", "通用")
	return strings.ToLower(value)
}
