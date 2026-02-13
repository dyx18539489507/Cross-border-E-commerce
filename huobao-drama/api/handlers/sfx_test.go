package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func TestSfxListMergeAndSort(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		query := r.URL.Query()
		if strings.Contains(path, "/search/text/") {
			if query.Get("token") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"missing token"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"id": 1,
						"name": "Freesound Bomb",
						"description": "爆炸音效",
						"tags": ["爆炸", "cinematic"],
						"duration": 2.2,
						"username": "fs-user",
						"num_downloads": 1200,
						"avg_rating": 4.6,
						"num_ratings": 36,
						"previews": {
							"preview-hq-mp3": "https://cdn.freesound.test/preview1.mp3"
						},
						"images": {
							"waveform_m": "https://cdn.freesound.test/wave1.png"
						}
					},
					{
						"id": 2,
						"name": "Freesound Door",
						"duration": 1.1,
						"username": "fs-user-2",
						"num_downloads": 800,
						"avg_rating": 4.0,
						"num_ratings": 18,
						"previews": {
							"preview-hq-mp3": "https://cdn.freesound.test/preview2.mp3"
						}
					}
				]
			}`))
			return
		}

		if strings.Contains(path, "/api/audio/") {
			if query.Get("key") == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`[ERROR 400] missing key`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"hits": [
					{
						"id": 100,
						"name": "Pixabay Explosion",
						"tags": "explosion, boom",
						"duration": 3.5,
						"downloads": 3400,
						"likes": 140,
						"user": "pix-user",
						"userImageURL": "https://cdn.pixabay.test/u1.png",
						"audio": {
							"mp3": "https://cdn.pixabay.test/audio1.mp3"
						}
					},
					{
						"id": 101,
						"tags": "door knock",
						"duration": 1.4,
						"downloads": 900,
						"likes": 12,
						"user": "pix-user-2",
						"audio": {
							"mp3": "https://cdn.pixabay.test/audio2.mp3"
						}
					}
				]
			}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer upstream.Close()

	cfg := &config.Config{
		SFX: config.SFXConfig{
			DefaultLimit:   20,
			RequestTimeout: 10,
			Freesound: config.FreesoundConfig{
				APIKey:  "freesound-test-key",
				BaseURL: upstream.URL,
			},
			Pixabay: config.PixabaySFXConfig{
				APIKey:  "pixabay-test-key",
				BaseURL: upstream.URL,
			},
		},
	}

	handler := NewSfxHandler(cfg, logger.NewLogger(false))
	router := gin.New()
	router.GET("/api/v1/sfx", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sfx?keywords=explosion&limit=4", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Items []SFXItem `json:"items"`
		Total int       `json:"total"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if payload.Total == 0 || len(payload.Items) == 0 {
		t.Fatalf("expected non-empty items")
	}
	if len(payload.Items) > 4 {
		t.Fatalf("expected <=4 items, got %d", len(payload.Items))
	}

	if payload.Items[0].Source != "pixabay" {
		t.Fatalf("expected first item from pixabay by heat, got %s", payload.Items[0].Source)
	}
	if payload.Items[0].Rank != 1 {
		t.Fatalf("expected first item rank=1, got %d", payload.Items[0].Rank)
	}
}

func TestSfxListFallsBackWhenPixabayUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search/text/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"id": 11,
						"name": "Freesound Wind",
						"duration": 4.2,
						"username": "fs-user",
						"num_downloads": 120,
						"avg_rating": 4.2,
						"num_ratings": 5,
						"previews": {
							"preview-hq-mp3": "https://cdn.freesound.test/wind.mp3"
						}
					}
				]
			}`))
			return
		}
		if strings.Contains(r.URL.Path, "/api/audio/") || strings.Contains(r.URL.Path, "/api/sounds/") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("[ERROR 403] Access denied."))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()

	cfg := &config.Config{
		SFX: config.SFXConfig{
			DefaultLimit:   20,
			RequestTimeout: 10,
			Freesound:      config.FreesoundConfig{APIKey: "freesound-test-key", BaseURL: upstream.URL},
			Pixabay:        config.PixabaySFXConfig{APIKey: "pixabay-test-key", BaseURL: upstream.URL},
		},
	}

	handler := NewSfxHandler(cfg, logger.NewLogger(false))
	router := gin.New()
	router.GET("/api/v1/sfx", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sfx?limit=20", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Items    []SFXItem `json:"items"`
		Warnings []string  `json:"warnings"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(payload.Items) == 0 {
		t.Fatalf("expected freesound fallback items")
	}
	if payload.Items[0].Source != "freesound" {
		t.Fatalf("expected freesound source, got %s", payload.Items[0].Source)
	}
	if len(payload.Warnings) == 0 {
		t.Fatalf("expected pixabay warning when upstream denied")
	}
	if !containsAny(payload.Warnings, "pixabay") {
		t.Fatalf("expected warnings to mention pixabay, got %#v", payload.Warnings)
	}
}

func containsAny(values []string, needle string) bool {
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func TestFirstTagOrFallback(t *testing.T) {
	if got := firstTagOrFallback("door,knock", "fallback"); got != "door" {
		t.Fatalf("unexpected first tag: %s", got)
	}
	if got := firstTagOrFallback("", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestFetchPixabayRequestURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var receivedURL *url.URL
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hits": []}`))
	}))
	defer upstream.Close()

	handler := NewSfxHandler(&config.Config{
		SFX: config.SFXConfig{RequestTimeout: 10},
	}, logger.NewLogger(false))

	_, err := handler.fetchPixabayByEndpoint(context.Background(), upstream.URL+"/api/audio/", "k", "open door", 12)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receivedURL == nil {
		t.Fatalf("expected request captured")
	}
	if receivedURL.Query().Get("order") != "popular" {
		t.Fatalf("expected order=popular")
	}
}
