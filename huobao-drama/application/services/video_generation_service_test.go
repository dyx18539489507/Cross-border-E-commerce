package services

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCompleteVideoGenerationPersistsLocalPlaybackURL(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "video-gen-test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Drama{}, &models.Episode{}, &models.Storyboard{}, &models.ImageGeneration{}, &models.VideoGeneration{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	drama := models.Drama{Title: "test drama", Status: "draft", Style: "realistic"}
	if err := db.Create(&drama).Error; err != nil {
		t.Fatalf("create drama: %v", err)
	}

	episode := models.Episode{DramaID: drama.ID, EpisodeNum: 1, Title: "ep1", Status: "draft"}
	if err := db.Create(&episode).Error; err != nil {
		t.Fatalf("create episode: %v", err)
	}

	storyboard := models.Storyboard{EpisodeID: episode.ID, StoryboardNumber: 1, Status: "pending"}
	if err := db.Create(&storyboard).Error; err != nil {
		t.Fatalf("create storyboard: %v", err)
	}

	videoGen := models.VideoGeneration{
		StoryboardID: &storyboard.ID,
		DramaID:      drama.ID,
		Provider:     "doubao",
		Prompt:       "test prompt",
		Status:       models.VideoStatusProcessing,
	}
	if err := db.Create(&videoGen).Error; err != nil {
		t.Fatalf("create video generation: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write([]byte("fake-mp4-content"))
	}))
	defer server.Close()

	localStorage, err := storage.NewLocalStorage(filepath.Join(t.TempDir(), "storage"), "http://localhost:5678/static")
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	svc := &VideoGenerationService{
		db:           db,
		localStorage: localStorage,
		log:          logger.NewLogger(false),
		ffmpeg:       nil,
	}

	duration := 8
	svc.completeVideoGeneration(videoGen.ID, server.URL+"/video.mp4", &duration, nil, nil, nil)

	var updated models.VideoGeneration
	if err := db.First(&updated, videoGen.ID).Error; err != nil {
		t.Fatalf("load updated video generation: %v", err)
	}

	if updated.VideoURL == nil || *updated.VideoURL == "" {
		t.Fatalf("expected persisted video_url, got nil")
	}
	if updated.LocalPath == nil || *updated.LocalPath == "" {
		t.Fatalf("expected persisted local_path, got nil")
	}
	if *updated.VideoURL != *updated.LocalPath {
		t.Fatalf("expected video_url and local_path to match, got video_url=%q local_path=%q", *updated.VideoURL, *updated.LocalPath)
	}
	if got := *updated.VideoURL; len(got) < len("/static/videos/") || got[:len("/static/videos/")] != "/static/videos/" {
		t.Fatalf("expected local static playback url path, got %q", got)
	}

	var updatedStoryboard models.Storyboard
	if err := db.First(&updatedStoryboard, storyboard.ID).Error; err != nil {
		t.Fatalf("load updated storyboard: %v", err)
	}
	if updatedStoryboard.VideoURL == nil || *updatedStoryboard.VideoURL == "" {
		t.Fatalf("expected storyboard.video_url to be synced")
	}
	if *updatedStoryboard.VideoURL != *updated.VideoURL {
		t.Fatalf("expected storyboard.video_url=%q, got %q", *updated.VideoURL, *updatedStoryboard.VideoURL)
	}
}
