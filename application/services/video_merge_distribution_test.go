package services

import (
	"strings"
	"testing"
	"time"

	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDistributionTestService(t *testing.T) (*VideoMergeService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db failed: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := db.AutoMigrate(&models.Drama{}, &models.Episode{}, &models.VideoMerge{}, &models.VideoDistribution{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	log := logger.NewLogger(false)
	t.Cleanup(func() {
		_ = log.Sync()
	})

	service := &VideoMergeService{
		db:  db,
		log: log,
	}

	return service, db
}

func seedCompletedMerge(t *testing.T, db *gorm.DB, status models.VideoMergeStatus) models.VideoMerge {
	t.Helper()

	drama := models.Drama{
		Title:         "测试短剧",
		TargetCountry: "US",
		Status:        "draft",
	}
	if err := db.Create(&drama).Error; err != nil {
		t.Fatalf("create drama failed: %v", err)
	}

	episode := models.Episode{
		DramaID:    drama.ID,
		EpisodeNum: 1,
		Title:      "第一集",
		Status:     "completed",
	}
	if err := db.Create(&episode).Error; err != nil {
		t.Fatalf("create episode failed: %v", err)
	}

	model := "doubao-seedance"
	mergedURL := "https://cdn.example.com/video/merged.mp4"
	merge := models.VideoMerge{
		EpisodeID: episode.ID,
		DramaID:   drama.ID,
		Title:     "测试合成视频",
		Provider:  "doubao",
		Model:     &model,
		Status:    status,
		Scenes:    []byte("[]"),
		MergedURL: &mergedURL,
	}
	if err := db.Create(&merge).Error; err != nil {
		t.Fatalf("create merge failed: %v", err)
	}

	return merge
}

func TestDistributeVideo_Success(t *testing.T) {
	service, db := newDistributionTestService(t)
	merge := seedCompletedMerge(t, db, models.VideoMergeStatusCompleted)

	records, err := service.DistributeVideo(merge.ID, &DistributeVideoRequest{
		Platforms:   []string{"tiktok", "youtube"},
		Title:       "一键分发测试",
		Description: "测试描述",
		Hashtags:    []string{"#出海", "短剧", "出海"},
	})
	if err != nil {
		t.Fatalf("DistributeVideo failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("unexpected distribution records: got=%d want=2", len(records))
	}

	deadline := time.Now().Add(6 * time.Second)
	for {
		distributions, listErr := service.ListDistributions(merge.ID)
		if listErr != nil {
			t.Fatalf("ListDistributions failed: %v", listErr)
		}

		if len(distributions) == 2 {
			allPublished := true
			platforms := map[string]bool{}
			for _, item := range distributions {
				platforms[item.Platform] = true
				if item.Status != models.VideoDistributionStatusPublished || item.PublishedURL == nil || *item.PublishedURL == "" {
					allPublished = false
					break
				}
			}

			if allPublished {
				if !platforms["tiktok"] || !platforms["youtube"] {
					t.Fatalf("unexpected platform set: %+v", platforms)
				}
				return
			}
		}

		if time.Now().After(deadline) {
			t.Fatalf("distributions not published in time")
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestDistributeVideo_RequiresCompletedMerge(t *testing.T) {
	service, db := newDistributionTestService(t)
	merge := seedCompletedMerge(t, db, models.VideoMergeStatusProcessing)

	if _, err := service.DistributeVideo(merge.ID, &DistributeVideoRequest{
		Platforms: []string{"tiktok"},
	}); err == nil {
		t.Fatalf("expected error for non-completed merge")
	}
}

func TestDistributeVideo_XPlatform(t *testing.T) {
	service, db := newDistributionTestService(t)
	merge := seedCompletedMerge(t, db, models.VideoMergeStatusCompleted)

	records, err := service.DistributeVideo(merge.ID, &DistributeVideoRequest{
		Platforms: []string{"x"},
		Title:     "X发布测试",
		Hashtags:  []string{"跨境", "短剧"},
	})
	if err != nil {
		t.Fatalf("DistributeVideo failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("unexpected distribution records: got=%d want=1", len(records))
	}

	deadline := time.Now().Add(6 * time.Second)
	for {
		distributions, listErr := service.ListDistributions(merge.ID)
		if listErr != nil {
			t.Fatalf("ListDistributions failed: %v", listErr)
		}

		if len(distributions) == 1 {
			item := distributions[0]
			if item.Status == models.VideoDistributionStatusPublished && item.PublishedURL != nil {
				if item.Platform != "x" {
					t.Fatalf("unexpected platform: %s", item.Platform)
				}
				if !strings.Contains(*item.PublishedURL, "x.com/intent/post") {
					t.Fatalf("unexpected published url: %s", *item.PublishedURL)
				}
				return
			}
		}

		if time.Now().After(deadline) {
			t.Fatalf("x distribution not published in time")
		}
		time.Sleep(200 * time.Millisecond)
	}
}
