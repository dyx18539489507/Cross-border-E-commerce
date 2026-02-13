package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
)

func TestSpeechTTSServiceSynthesize(t *testing.T) {
	type submitObservedRequest struct {
		Namespace string `json:"namespace"`
		ReqParams struct {
			Text    string `json:"text"`
			Speaker string `json:"speaker"`
			Audio   struct {
				Format     string `json:"format"`
				SampleRate int    `json:"sample_rate"`
			} `json:"audio_params"`
		} `json:"req_params"`
	}

	var capturedSubmit submitObservedRequest
	var capturedHeaders map[string]string
	queryCallCount := 0

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/submit":
			capturedHeaders = map[string]string{
				"x-api-app-id":      r.Header.Get("X-Api-App-Id"),
				"x-api-access-key":  r.Header.Get("X-Api-Access-Key"),
				"x-api-resource-id": r.Header.Get("X-Api-Resource-Id"),
				"x-api-request-id":  r.Header.Get("X-Api-Request-Id"),
			}
			if err := json.NewDecoder(r.Body).Decode(&capturedSubmit); err != nil {
				t.Fatalf("decode submit request failed: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    20000000,
				"message": "ok",
				"data": map[string]any{
					"task_id":     "task-123",
					"task_status": 1,
				},
			})
		case "/query":
			queryCallCount++
			if queryCallCount == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"code":    20000000,
					"message": "ok",
					"data": map[string]any{
						"task_id":     "task-123",
						"task_status": 1,
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":    20000000,
				"message": "ok",
				"data": map[string]any{
					"task_id":     "task-123",
					"task_status": 2,
					"audio_url":   server.URL + "/audio.mp3",
				},
			})
		case "/audio.mp3":
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write([]byte("audio-bytes"))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Volcengine: config.VolcengineConfig{
			Speech: config.VolcengineSpeechConfig{
				AppID:          "test-app-id",
				Token:          "test-token",
				SubmitEndpoint: server.URL + "/submit",
				QueryEndpoint:  server.URL + "/query",
				ResourceID:     "volc.service_type.10029",
				Namespace:      "BidirectionalTTS",
			},
		},
	}

	service := NewSpeechTTSService(cfg, logger.NewLogger(false))
	audio, contentType, err := service.Synthesize(context.Background(), "你好，数字人", "zh_male_livelybro_mars_bigtts")
	if err != nil {
		t.Fatalf("synthesize failed: %v", err)
	}
	if contentType != "audio/mpeg" {
		t.Fatalf("unexpected content type: %s", contentType)
	}
	if string(audio) != "audio-bytes" {
		t.Fatalf("unexpected audio bytes: %q", string(audio))
	}
	if queryCallCount < 2 {
		t.Fatalf("expected at least 2 query calls, got %d", queryCallCount)
	}

	if capturedSubmit.Namespace != "BidirectionalTTS" {
		t.Fatalf("unexpected namespace: %s", capturedSubmit.Namespace)
	}
	if capturedSubmit.ReqParams.Text != "你好，数字人" {
		t.Fatalf("unexpected text: %s", capturedSubmit.ReqParams.Text)
	}
	if capturedSubmit.ReqParams.Speaker != "zh_male_livelybro_mars_bigtts" {
		t.Fatalf("unexpected speaker: %s", capturedSubmit.ReqParams.Speaker)
	}
	if capturedSubmit.ReqParams.Audio.Format != "mp3" {
		t.Fatalf("unexpected format: %s", capturedSubmit.ReqParams.Audio.Format)
	}
	if capturedSubmit.ReqParams.Audio.SampleRate != 24000 {
		t.Fatalf("unexpected sample rate: %d", capturedSubmit.ReqParams.Audio.SampleRate)
	}

	if capturedHeaders["x-api-app-id"] != "test-app-id" {
		t.Fatalf("unexpected X-Api-App-Id: %s", capturedHeaders["x-api-app-id"])
	}
	if capturedHeaders["x-api-access-key"] != "test-token" {
		t.Fatalf("unexpected X-Api-Access-Key: %s", capturedHeaders["x-api-access-key"])
	}
	if capturedHeaders["x-api-resource-id"] != "volc.service_type.10029" {
		t.Fatalf("unexpected X-Api-Resource-Id: %s", capturedHeaders["x-api-resource-id"])
	}
	if strings.TrimSpace(capturedHeaders["x-api-request-id"]) == "" {
		t.Fatalf("expected non-empty X-Api-Request-Id")
	}
}

func TestSpeechTTSServiceSynthesizeFailedCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    55000000,
			"message": "requested resource not granted",
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		Volcengine: config.VolcengineConfig{
			Speech: config.VolcengineSpeechConfig{
				AppID:          "test-app-id",
				Token:          "test-token",
				SubmitEndpoint: server.URL,
				QueryEndpoint:  server.URL,
				ResourceID:     "volc.service_type.10029",
			},
		},
	}

	service := NewSpeechTTSService(cfg, logger.NewLogger(false))
	_, _, err := service.Synthesize(context.Background(), "测试", "zh_male_livelybro_mars_bigtts")
	if err == nil {
		t.Fatalf("expected error for failed tts code")
	}
	if !strings.Contains(err.Error(), "requested resource not granted") {
		t.Fatalf("unexpected error: %v", err)
	}
}
