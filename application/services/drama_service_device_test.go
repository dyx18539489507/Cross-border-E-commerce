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
