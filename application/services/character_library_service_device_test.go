package services

import (
	"strconv"
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCharacterLibraryServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.CharacterLibrary{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	return db
}

func TestCharacterLibraryListItemsIsScopedByDevice(t *testing.T) {
	db := newCharacterLibraryServiceTestDB(t)
	svc := NewCharacterLibraryService(db, logger.NewLogger(false))

	deviceA := "dev_library_browser_a_123456"
	deviceB := "dev_library_browser_b_123456"

	itemA, err := svc.CreateLibraryItem(&CreateLibraryItemRequest{
		Name:       "Hero A",
		ImageURL:   "https://example.com/a.png",
		SourceType: "manual",
	}, deviceA)
	if err != nil {
		t.Fatalf("create itemA: %v", err)
	}

	if _, err := svc.CreateLibraryItem(&CreateLibraryItemRequest{
		Name:       "Hero B",
		ImageURL:   "https://example.com/b.png",
		SourceType: "manual",
	}, deviceB); err != nil {
		t.Fatalf("create itemB: %v", err)
	}

	items, total, err := svc.ListLibraryItems(&CharacterLibraryQuery{Page: 1, PageSize: 20}, deviceA)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(items) != 1 || items[0].ID != itemA.ID {
		t.Fatalf("expected only itemA, got %+v", items)
	}
}

func TestCharacterLibraryGetItemRejectsOtherDevice(t *testing.T) {
	db := newCharacterLibraryServiceTestDB(t)
	svc := NewCharacterLibraryService(db, logger.NewLogger(false))

	item, err := svc.CreateLibraryItem(&CreateLibraryItemRequest{
		Name:       "Hero A",
		ImageURL:   "https://example.com/a.png",
		SourceType: "manual",
	}, "dev_library_browser_a_654321")
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	if _, err := svc.GetLibraryItem(strconv.FormatUint(uint64(item.ID), 10), "dev_library_browser_b_654321"); err == nil || err.Error() != "library item not found" {
		t.Fatalf("expected library item not found, got %v", err)
	}
}
