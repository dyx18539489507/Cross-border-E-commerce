package models

import "time"

type AgentTrace struct {
	AgentName  string      `json:"agent_name"`
	Stage      string      `json:"stage"`
	Input      interface{} `json:"input,omitempty" gorm:"-"`
	Output     interface{} `json:"output,omitempty" gorm:"-"`
	Status     string      `json:"status"`
	Error      string      `json:"error,omitempty"`
	StartedAt  time.Time   `json:"started_at"`
	EndedAt    time.Time   `json:"ended_at"`
	DurationMs int64       `json:"duration_ms"`
}

type ComplianceRule struct {
	ID                   string   `json:"id"`
	Country              string   `json:"country"`
	Platform             string   `json:"platform"`
	Category             string   `json:"category"`
	RiskType             string   `json:"risk_type"`
	RuleText             string   `json:"rule_text"`
	ForbiddenExpressions []string `json:"forbidden_expressions"`
	SaferExpressions     []string `json:"safer_expressions"`
	SourceURL            string   `json:"source_url"`
	UpdatedAt            string   `json:"updated_at"`
}

type CriticResult struct {
	CompletenessScore int      `json:"completeness_score"`
	ComplianceScore   int      `json:"compliance_score"`
	LocalizationScore int      `json:"localization_score"`
	MarketingScore    int      `json:"marketing_score"`
	OverallScore      int      `json:"overall_score"`
	Problems          []string `json:"problems"`
	RevisionAdvice    []string `json:"revision_advice"`
	NeedRevise        bool     `json:"need_revise"`
}
