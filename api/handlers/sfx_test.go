package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func TestGenerateSFX(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			LocalPath: tmpDir,
			BaseURL:   "http://localhost:8080/static",
		},
	}
	log := logger.NewLogger(false)
	handler := NewSFXHandler(cfg, log)

	router := gin.New()
	router.POST("/api/v1/sfx/generate", handler.GenerateSFX)

	reqBody := []byte(`{"prompt":"爆炸","count":3}`)
	req := httptest.NewRequest("POST", "/api/v1/sfx/generate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var payload struct {
		Items []SFXItem `json:"items"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(payload.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(payload.Items))
	}

	for _, item := range payload.Items {
		if !strings.Contains(item.URL, "/sfx/") {
			t.Fatalf("unexpected url: %s", item.URL)
		}
		fileName := strings.TrimPrefix(item.URL, cfg.Storage.BaseURL+"/sfx/")
		if fileName == item.URL {
			t.Fatalf("failed to parse filename from url: %s", item.URL)
		}
		filePath := filepath.Join(tmpDir, "sfx", fileName)
		if _, err := os.Stat(filePath); err != nil {
			t.Fatalf("expected file to exist: %v", err)
		}
	}
}

func TestListSFX(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Storage: config.StorageConfig{
			LocalPath: tmpDir,
			BaseURL:   "http://localhost:8080/static",
		},
	}
	log := logger.NewLogger(false)
	handler := NewSFXHandler(cfg, log)

	router := gin.New()
	router.GET("/api/v1/sfx", handler.ListSFX)

	req := httptest.NewRequest("GET", "/api/v1/sfx?category=转场&limit=3", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var payload struct {
		Items []SFXItem `json:"items"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(payload.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(payload.Items))
	}
}
