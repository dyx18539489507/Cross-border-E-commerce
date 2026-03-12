package services

import (
	"path/filepath"
	"testing"

	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestAIService(t *testing.T) (*AIService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "ai-service-test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.AIServiceConfig{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	return NewAIService(db, logger.NewLogger(false)), db
}

func TestGetDefaultConfigPrefersDeviceConfigAndFallsBackToGlobal(t *testing.T) {
	svc, db := newTestAIService(t)

	globalConfig := models.AIServiceConfig{
		ServiceType: "image",
		Provider:    "volcengine",
		Name:        "global image",
		BaseURL:     "https://example.com",
		APIKey:      "global-key",
		Model:       models.ModelField{"global-model"},
		IsActive:    true,
		Priority:    10,
	}
	if err := db.Create(&globalConfig).Error; err != nil {
		t.Fatalf("create global config: %v", err)
	}

	deviceConfig := models.AIServiceConfig{
		DeviceID:    "dev_known",
		ServiceType: "image",
		Provider:    "openai",
		Name:        "device image",
		BaseURL:     "https://device.example.com",
		APIKey:      "device-key",
		Model:       models.ModelField{"device-model"},
		IsActive:    true,
		Priority:    20,
	}
	if err := db.Create(&deviceConfig).Error; err != nil {
		t.Fatalf("create device config: %v", err)
	}

	got, err := svc.GetDefaultConfig("image", "dev_known")
	if err != nil {
		t.Fatalf("GetDefaultConfig for known device failed: %v", err)
	}
	if got.ID != deviceConfig.ID {
		t.Fatalf("expected device config %d, got %d", deviceConfig.ID, got.ID)
	}

	got, err = svc.GetDefaultConfig("image", "dev_missing")
	if err != nil {
		t.Fatalf("GetDefaultConfig with global fallback failed: %v", err)
	}
	if got.ID != globalConfig.ID {
		t.Fatalf("expected global fallback config %d, got %d", globalConfig.ID, got.ID)
	}
}

func TestGetConfigForModelFallsBackToGlobalConfig(t *testing.T) {
	svc, db := newTestAIService(t)

	globalConfig := models.AIServiceConfig{
		ServiceType: "image",
		Provider:    "volcengine",
		Name:        "global image",
		BaseURL:     "https://example.com",
		APIKey:      "global-key",
		Model:       models.ModelField{"global-model"},
		IsActive:    true,
		Priority:    10,
	}
	if err := db.Create(&globalConfig).Error; err != nil {
		t.Fatalf("create global config: %v", err)
	}

	deviceConfig := models.AIServiceConfig{
		DeviceID:    "dev_known",
		ServiceType: "image",
		Provider:    "openai",
		Name:        "device image",
		BaseURL:     "https://device.example.com",
		APIKey:      "device-key",
		Model:       models.ModelField{"device-model"},
		IsActive:    true,
		Priority:    20,
	}
	if err := db.Create(&deviceConfig).Error; err != nil {
		t.Fatalf("create device config: %v", err)
	}

	got, err := svc.GetConfigForModel("image", "device-model", "dev_known")
	if err != nil {
		t.Fatalf("GetConfigForModel for device model failed: %v", err)
	}
	if got.ID != deviceConfig.ID {
		t.Fatalf("expected device config %d, got %d", deviceConfig.ID, got.ID)
	}

	got, err = svc.GetConfigForModel("image", "global-model", "dev_missing")
	if err != nil {
		t.Fatalf("GetConfigForModel with global fallback failed: %v", err)
	}
	if got.ID != globalConfig.ID {
		t.Fatalf("expected global fallback config %d, got %d", globalConfig.ID, got.ID)
	}
}

func TestListConfigsIncludesLegacyGlobalConfigsForDevice(t *testing.T) {
	svc, db := newTestAIService(t)

	globalConfig := models.AIServiceConfig{
		ServiceType: "image",
		Provider:    "volcengine",
		Name:        "global image",
		BaseURL:     "https://example.com",
		APIKey:      "global-key",
		Model:       models.ModelField{"global-model"},
		IsActive:    true,
		Priority:    10,
	}
	if err := db.Create(&globalConfig).Error; err != nil {
		t.Fatalf("create global config: %v", err)
	}

	configs, err := svc.ListConfigs("image", "dev_missing")
	if err != nil {
		t.Fatalf("ListConfigs failed: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].ID != globalConfig.ID {
		t.Fatalf("expected global config %d, got %d", globalConfig.ID, configs[0].ID)
	}
}
