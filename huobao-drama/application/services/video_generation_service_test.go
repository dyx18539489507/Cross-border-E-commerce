package services

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestTryLocalToDataURLSupportsConfiguredBaseURLHost(t *testing.T) {
	storageDir := filepath.Join(t.TempDir(), "storage")
	imageDir := filepath.Join(storageDir, "images")
	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		t.Fatalf("create image dir: %v", err)
	}
	imagePath := filepath.Join(imageDir, "demo.jpeg")
	if err := os.WriteFile(imagePath, []byte("fake-jpeg-content"), 0o644); err != nil {
		t.Fatalf("write image file: %v", err)
	}

	localStorage, err := storage.NewLocalStorage(storageDir, "https://beige-mice-kiss.loca.lt/static")
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	svc := &VideoGenerationService{
		localStorage: localStorage,
		log:          logger.NewLogger(false),
	}

	dataURL, err := svc.tryLocalToDataURL("https://beige-mice-kiss.loca.lt/static/images/demo.jpeg")
	if err != nil {
		t.Fatalf("tryLocalToDataURL failed: %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/jpeg;base64,") {
		t.Fatalf("unexpected data url prefix: %q", dataURL)
	}
}

func TestIsLocalMediaURLRecognizesConfiguredBaseURL(t *testing.T) {
	localStorage, err := storage.NewLocalStorage(filepath.Join(t.TempDir(), "storage"), "https://beige-mice-kiss.loca.lt/static")
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	svc := &VideoGenerationService{
		localStorage: localStorage,
		log:          logger.NewLogger(false),
	}

	if !svc.isLocalMediaURL("https://beige-mice-kiss.loca.lt/static/images/demo.jpeg") {
		t.Fatalf("expected URL under configured storage base_url to be treated as local")
	}
	if svc.isLocalMediaURL("https://example.com/static/images/demo.jpeg") {
		t.Fatalf("expected unrelated static URL not to be treated as local")
	}
}

func TestGenerateVideoFromImageUsesDefaultConfigProvider(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "video-provider-test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Drama{}, &models.ImageGeneration{}, &models.VideoGeneration{}, &models.AIServiceConfig{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	drama := models.Drama{Title: "provider test drama", Status: "draft", Style: "realistic"}
	if err := db.Create(&drama).Error; err != nil {
		t.Fatalf("create drama: %v", err)
	}

	imageURL := "/static/images/demo.jpeg"
	imageGen := models.ImageGeneration{
		DramaID:  drama.ID,
		Provider: "openai",
		Prompt:   "a valid prompt for video generation",
		Status:   models.ImageStatusCompleted,
		ImageURL: &imageURL,
	}
	if err := db.Create(&imageGen).Error; err != nil {
		t.Fatalf("create image generation: %v", err)
	}

	cfg := models.AIServiceConfig{
		ServiceType: "video",
		Provider:    "custom-video",
		Name:        "custom default video",
		BaseURL:     "https://example.com",
		APIKey:      "test-key",
		Model:       models.ModelField{"custom-model"},
		IsActive:    true,
		Priority:    100,
	}
	if err := db.Create(&cfg).Error; err != nil {
		t.Fatalf("create ai config: %v", err)
	}

	log := logger.NewLogger(false)
	svc := NewVideoGenerationService(db, nil, nil, NewAIService(db, log), log)

	videoGen, err := svc.GenerateVideoFromImage(imageGen.ID)
	if err != nil {
		t.Fatalf("GenerateVideoFromImage failed: %v", err)
	}
	if videoGen.Provider != "custom-video" {
		t.Fatalf("expected provider to come from default AI config, got %q", videoGen.Provider)
	}
}

func TestResolveMediaInputURLForProviderUsesDataURLForVolcesLocalImage(t *testing.T) {
	storageDir := filepath.Join(t.TempDir(), "storage")
	imageDir := filepath.Join(storageDir, "images")
	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		t.Fatalf("create image dir: %v", err)
	}
	imagePath := filepath.Join(imageDir, "demo.jpeg")
	if err := os.WriteFile(imagePath, []byte("fake-jpeg-content"), 0o644); err != nil {
		t.Fatalf("write image file: %v", err)
	}

	localStorage, err := storage.NewLocalStorage(storageDir, "https://beige-mice-kiss.loca.lt/static")
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	svc := &VideoGenerationService{
		localStorage: localStorage,
		log:          logger.NewLogger(false),
	}

	resolved := svc.resolveMediaInputURLForProvider("https://beige-mice-kiss.loca.lt/static/images/demo.jpeg", "volces")
	if !strings.HasPrefix(resolved, "data:image/jpeg;base64,") {
		t.Fatalf("expected data url for volces local image, got %q", resolved)
	}
}
