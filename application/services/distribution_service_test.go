package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDistributionServiceTestEnv(t *testing.T) (*DistributionService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(
		&models.UploadPostProfile{},
		&models.DistributionTarget{},
		&models.DistributionJob{},
		&models.DistributionResult{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	cfg := &config.Config{
		Storage: config.StorageConfig{
			LocalPath: t.TempDir(),
			BaseURL:   "http://localhost:5678/static",
		},
		Distribution: config.DistributionConfig{
			UploadPostBaseURL:        "http://localhost.invalid",
			UploadPostConnectTitle:   "Connect",
			UploadPostConnectDesc:    "Connect your accounts",
			StatusPollIntervalSecond: 1,
			HistoryLookbackPages:     1,
			DiscordUsername:          "Drama Generator",
		},
	}

	log := logger.NewLogger(false)
	t.Cleanup(func() {
		_ = log.Sync()
	})

	return NewDistributionService(db, cfg, log), db
}

func TestDistributionService_CreateDistributionRequiresDiscordTarget(t *testing.T) {
	service, _ := newDistributionServiceTestEnv(t)

	_, err := service.CreateDistribution(context.Background(), "dev_test", &CreateDistributionRequest{
		ContentType:       "text",
		Title:             "hello",
		SelectedPlatforms: []string{"discord"},
	})
	if err == nil {
		t.Fatal("expected create distribution to fail without discord target")
	}
	if !strings.Contains(err.Error(), "Discord") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDistributionService_CreateDiscordDistributionSuccess(t *testing.T) {
	t.Setenv("DISTRIBUTION_SECRET_KEY", "test-distribution-secret")

	service, _ := newDistributionServiceTestEnv(t)

	var (
		mu          sync.Mutex
		lastPayload map[string]interface{}
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/webhooks/test-id/test-token"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         "test-id",
				"guild_id":   "guild-1",
				"channel_id": "channel-1",
				"name":       "Demo Hook",
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/webhooks/test-id/test-token"):
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&lastPayload); err != nil {
				t.Fatalf("decode discord payload failed: %v", err)
			}
			mu.Lock()
			defer mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         "message-1",
				"guild_id":   "guild-1",
				"channel_id": "channel-1",
				"content":    lastPayload["content"],
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service.discord.client = server.Client()

	target, err := service.UpsertDiscordTarget(context.Background(), "dev_test", UpsertDiscordTargetRequest{
		WebhookURL: server.URL + "/webhooks/test-id/test-token",
		IsDefault:  true,
	})
	if err != nil {
		t.Fatalf("upsert discord target failed: %v", err)
	}
	if target.Identifier != "test-id" {
		t.Fatalf("unexpected target identifier: %s", target.Identifier)
	}

	job, err := service.CreateDistribution(context.Background(), "dev_test", &CreateDistributionRequest{
		ContentType:       "text",
		Title:             "Launch",
		Body:              "notify @everyone now",
		SelectedPlatforms: []string{"discord"},
	})
	if err != nil {
		t.Fatalf("create distribution failed: %v", err)
	}

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		loaded, loadErr := service.GetDistributionJob("dev_test", job.ID)
		if loadErr != nil {
			t.Fatalf("get distribution job failed: %v", loadErr)
		}
		if loaded.Status == models.DistributionJobStatusCompleted {
			if len(loaded.Results) != 1 {
				t.Fatalf("unexpected results count: %d", len(loaded.Results))
			}
			result := loaded.Results[0]
			if result.Status != models.DistributionResultStatusSuccess {
				t.Fatalf("unexpected result status: %s", result.Status)
			}
			if result.ExternalMessageID == nil || *result.ExternalMessageID != "message-1" {
				t.Fatalf("unexpected message id: %+v", result.ExternalMessageID)
			}
			if result.PublishedURL == nil || !strings.Contains(*result.PublishedURL, "discord.com/channels/guild-1/channel-1/message-1") {
				t.Fatalf("unexpected published url: %+v", result.PublishedURL)
			}

			mu.Lock()
			defer mu.Unlock()
			content, _ := lastPayload["content"].(string)
			if !strings.Contains(content, "@\u200beveryone") {
				t.Fatalf("expected sanitized discord content, got %q", content)
			}
			allowedMentions, _ := lastPayload["allowed_mentions"].(map[string]interface{})
			if allowedMentions == nil {
				t.Fatalf("expected allowed_mentions in payload")
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatal("distribution job did not complete in time")
}

func TestUploadPostAdapterGenerateConnectURL(t *testing.T) {
	t.Setenv("UPLOAD_POST_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/uploadposts/users/generate-jwt" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Apikey test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_url": "https://upload-post.example/connect/abc",
			"duration":   "48h",
		})
	}))
	defer server.Close()

	cfg := &config.Config{
		Distribution: config.DistributionConfig{
			UploadPostBaseURL:     server.URL,
			UploadPostConnectTitle: "Connect",
			UploadPostConnectDesc:  "Connect your accounts",
		},
	}

	adapter := NewUploadPostAdapter(cfg)
	resp, err := adapter.GenerateConnectURL(context.Background(), UploadPostGenerateLinkRequest{
		Username:           "demo-user",
		ConnectTitle:       "Connect",
		ConnectDescription: "Connect your accounts",
	})
	if err != nil {
		t.Fatalf("generate connect url failed: %v", err)
	}
	if resp.AccessURL != "https://upload-post.example/connect/abc" {
		t.Fatalf("unexpected access url: %s", resp.AccessURL)
	}
}
