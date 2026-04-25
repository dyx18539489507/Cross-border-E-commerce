package services

import (
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDramaServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Drama{}, &models.Episode{}, &models.Character{}, &models.Scene{}, &models.Storyboard{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	return db
}

func TestDramaServiceListDramasIsScopedByDevice(t *testing.T) {
	db := newDramaServiceTestDB(t)
	svc := NewDramaService(db, logger.NewLogger(false), nil)

	dramaA := &models.Drama{DeviceID: "dev_browser_a_123456", Title: "A", Status: "draft", TargetCountry: "US", ComplianceLevel: "green"}
	dramaB := &models.Drama{DeviceID: "dev_browser_b_123456", Title: "B", Status: "draft", TargetCountry: "JP", ComplianceLevel: "green"}
	if err := db.Create(dramaA).Error; err != nil {
		t.Fatalf("create dramaA: %v", err)
	}
	if err := db.Create(dramaB).Error; err != nil {
		t.Fatalf("create dramaB: %v", err)
	}

	query := &DramaListQuery{Page: 1, PageSize: 20}
	items, total, err := svc.ListDramas(query, dramaA.DeviceID)
	if err != nil {
		t.Fatalf("list dramas: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(items) != 1 || items[0].ID != dramaA.ID {
		t.Fatalf("expected only dramaA, got %+v", items)
	}
}

func TestDramaServiceClaimsLegacyDramasForFirstDevice(t *testing.T) {
	db := newDramaServiceTestDB(t)
	svc := NewDramaService(db, logger.NewLogger(false), nil)

	legacy := &models.Drama{Title: "legacy", Status: "draft", TargetCountry: "US", ComplianceLevel: "green"}
	if err := db.Create(legacy).Error; err != nil {
		t.Fatalf("create legacy drama: %v", err)
	}

	query := &DramaListQuery{Page: 1, PageSize: 20}
	items, total, err := svc.ListDramas(query, "dev_new_browser_123456")
	if err != nil {
		t.Fatalf("list dramas: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected claimed legacy drama, got total=%d len=%d", total, len(items))
	}

	var refreshed models.Drama
	if err := db.First(&refreshed, legacy.ID).Error; err != nil {
		t.Fatalf("reload legacy drama: %v", err)
	}
	if refreshed.DeviceID != "dev_new_browser_123456" {
		t.Fatalf("expected claimed device_id, got %q", refreshed.DeviceID)
	}
}

func TestDramaServiceCreatesWithPrecheckedComplianceToken(t *testing.T) {
	db := newDramaServiceTestDB(t)
	svc := NewDramaService(db, logger.NewLogger(false), nil)

	req := &CreateDramaRequest{
		Title:                  "磁吸充电宝",
		Description:            "兼容磁吸协议的无线充",
		TargetCountry:          []string{"US", "JP"},
		MaterialComposition:    "ABS+锂电池",
		MarketingSellingPoints: "磁吸快充",
	}
	input, err := prepareCreateDramaInput(req)
	if err != nil {
		t.Fatalf("prepare create drama input: %v", err)
	}

	deviceID := "dev_cache_browser_123456"
	expected := &ComplianceResult{
		Score:                    75,
		Level:                    ComplianceRiskOrange,
		LevelLabel:               "高",
		Summary:                  "缓存命中的橙色风险结果",
		NonCompliancePoints:      []string{"疑似侵权表述"},
		RectificationSuggestions: []string{"修改营销文案"},
		SuggestedCategories:      []string{"电子产品"},
	}
	cacheKey := buildComplianceCacheKey(input, deviceID)
	svc.setCachedComplianceResult(cacheKey, expected)
	token := svc.issueComplianceToken(cacheKey, deviceID, expected)
	req.ComplianceToken = token

	compliance, issuedToken, err := svc.EvaluateCompliance(&CreateDramaRequest{
		Title:                  req.Title,
		Description:            req.Description,
		TargetCountry:          req.TargetCountry,
		MaterialComposition:    req.MaterialComposition,
		MarketingSellingPoints: req.MarketingSellingPoints,
	}, deviceID)
	if err != nil {
		t.Fatalf("evaluate compliance: %v", err)
	}
	if compliance.Score != expected.Score || compliance.Level != expected.Level {
		t.Fatalf("expected cached compliance %+v, got %+v", expected, compliance)
	}
	if issuedToken == "" {
		t.Fatal("expected non-empty compliance token")
	}

	drama, complianceFromCreate, err := svc.CreateDrama(req, deviceID)
	if err != nil {
		t.Fatalf("create drama: %v", err)
	}
	if complianceFromCreate.Score != expected.Score || complianceFromCreate.Level != expected.Level {
		t.Fatalf("expected create to reuse cached compliance %+v, got %+v", expected, complianceFromCreate)
	}
	if drama.ComplianceScore != expected.Score {
		t.Fatalf("expected stored compliance score %d, got %d", expected.Score, drama.ComplianceScore)
	}
}
