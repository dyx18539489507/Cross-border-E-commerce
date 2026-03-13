package services

import (
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAIServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.AIServiceConfig{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	return db
}

func TestAIServiceListConfigsIsScopedByDevice(t *testing.T) {
	db := newAIServiceTestDB(t)
	svc := NewAIService(db, logger.NewLogger(false))

	deviceA := "dev_ai_browser_a_123456"
	deviceB := "dev_ai_browser_b_123456"

	configA, err := svc.CreateConfig(&CreateAIConfigRequest{
		ServiceType: "text",
		Name:        "Config A",
		Provider:    "openai",
		BaseURL:     "https://api.example.com",
		APIKey:      "key-a",
		Model:       models.ModelField{"model-a"},
		Priority:    10,
	}, deviceA)
	if err != nil {
		t.Fatalf("create configA: %v", err)
	}

	if _, err := svc.CreateConfig(&CreateAIConfigRequest{
		ServiceType: "text",
		Name:        "Config B",
		Provider:    "openai",
		BaseURL:     "https://api.example.com",
		APIKey:      "key-b",
		Model:       models.ModelField{"model-b"},
		Priority:    20,
	}, deviceB); err != nil {
		t.Fatalf("create configB: %v", err)
	}

	items, err := svc.ListConfigs("", deviceA)
	if err != nil {
		t.Fatalf("list configs: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 config, got %d", len(items))
	}
	if items[0].ID != configA.ID {
		t.Fatalf("expected only configA, got %+v", items)
	}
}

func TestAIServiceGetConfigRejectsOtherDevice(t *testing.T) {
	db := newAIServiceTestDB(t)
	svc := NewAIService(db, logger.NewLogger(false))

	config, err := svc.CreateConfig(&CreateAIConfigRequest{
		ServiceType: "text",
		Name:        "Config A",
		Provider:    "openai",
		BaseURL:     "https://api.example.com",
		APIKey:      "key-a",
		Model:       models.ModelField{"model-a"},
	}, "dev_ai_browser_a_654321")
	if err != nil {
		t.Fatalf("create config: %v", err)
	}

	if _, err := svc.GetConfig(config.ID, "dev_ai_browser_b_654321"); err == nil || err.Error() != "config not found" {
		t.Fatalf("expected config not found, got %v", err)
	}
}
