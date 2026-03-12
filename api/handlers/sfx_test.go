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
		Items   []SFXItem `json:"items"`
		Page    int       `json:"page"`
		Limit   int       `json:"limit"`
		Total   int       `json:"total"`
		HasMore bool      `json:"has_more"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(payload.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(payload.Items))
	}
	if payload.Page != 1 {
		t.Fatalf("expected page 1, got %d", payload.Page)
	}
	if payload.Limit != 3 {
		t.Fatalf("expected limit 3, got %d", payload.Limit)
	}
	if payload.Total <= len(payload.Items) {
		t.Fatalf("expected total to be larger than current page size, got total=%d items=%d", payload.Total, len(payload.Items))
	}
	if !payload.HasMore {
		t.Fatal("expected has_more to be true on the first page")
	}
	if payload.Items[0].Category != "转场" {
		t.Fatalf("expected display category to be 转场, got %s", payload.Items[0].Category)
	}
	if payload.Items[0].Rank != 1 {
		t.Fatalf("expected first item rank 1, got %d", payload.Items[0].Rank)
	}

	nextReq := httptest.NewRequest("GET", "/api/v1/sfx?category=转场&limit=3&page=2", nil)
	nextResp := httptest.NewRecorder()
	router.ServeHTTP(nextResp, nextReq)

	if nextResp.Code != 200 {
		t.Fatalf("expected page 2 status 200, got %d", nextResp.Code)
	}

	var nextPayload struct {
		Items []SFXItem `json:"items"`
		Page  int       `json:"page"`
	}
	if err := json.Unmarshal(nextResp.Body.Bytes(), &nextPayload); err != nil {
		t.Fatalf("failed to parse page 2 response: %v", err)
	}
	if len(nextPayload.Items) != 3 {
		t.Fatalf("expected 3 items on page 2, got %d", len(nextPayload.Items))
	}
	if nextPayload.Page != 2 {
		t.Fatalf("expected page 2, got %d", nextPayload.Page)
	}
	if nextPayload.Items[0].ID == payload.Items[0].ID {
		t.Fatal("expected page 2 to return a different first item")
	}
	if nextPayload.Items[0].Rank != 4 {
		t.Fatalf("expected page 2 first rank 4, got %d", nextPayload.Items[0].Rank)
	}
}
