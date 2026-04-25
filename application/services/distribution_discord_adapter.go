package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/config"
)

var discordWebhookPattern = regexp.MustCompile(`^https://discord(?:app)?\.com/api/webhooks/([^/]+)/([^/?#]+)`)

type DiscordAdapter struct {
	client    *http.Client
	username  string
	avatarURL string
}

type DiscordWebhookMetadata struct {
	ID        string `json:"id"`
	Type      int    `json:"type"`
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	Name      string `json:"name"`
	Token     string `json:"token"`
}

type DiscordSendRequest struct {
	WebhookURL  string
	Title       string
	Body        string
	MediaURL    string
	ContentType string
}

type DiscordSendResponse struct {
	ID        string                 `json:"id"`
	ChannelID string                 `json:"channel_id"`
	GuildID   string                 `json:"guild_id"`
	Content   string                 `json:"content"`
	WebhookID string                 `json:"webhook_id"`
	Raw       map[string]interface{} `json:"raw"`
}

func NewDiscordAdapter(cfg *config.Config) *DiscordAdapter {
	return &DiscordAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		username:  cfg.Distribution.DiscordUsername,
		avatarURL: cfg.Distribution.DiscordAvatarURL,
	}
}

func (a *DiscordAdapter) ValidateWebhook(ctx context.Context, webhookURL string) (*DiscordWebhookMetadata, error) {
	if strings.TrimSpace(webhookURL) == "" {
		return nil, errors.New("Discord webhook URL 不能为空")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, webhookURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create discord webhook request failed: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, &DistributionAdapterError{
			Message:   fmt.Sprintf("validate discord webhook failed: %v", err),
			Retriable: true,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read discord webhook response failed: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, &DistributionAdapterError{
			StatusCode:  resp.StatusCode,
			Message:     fmt.Sprintf("validate discord webhook failed: %s", resp.Status),
			Body:        string(body),
			Retriable:   resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError,
			NeedsRebind: resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
		}
	}

	var metadata DiscordWebhookMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("unmarshal discord webhook metadata failed: %w", err)
	}

	if metadata.ID == "" {
		metadata.ID, _ = ExtractDiscordWebhookID(webhookURL)
	}

	return &metadata, nil
}

func (a *DiscordAdapter) Send(ctx context.Context, req DiscordSendRequest) (*DiscordSendResponse, error) {
	webhookURL := strings.TrimSpace(req.WebhookURL)
	if webhookURL == "" {
		return nil, errors.New("Discord webhook URL 不能为空")
	}

	content := buildDiscordContent(req.Title, req.Body, req.MediaURL, req.ContentType)
	payload := map[string]interface{}{
		"content": content,
		"allowed_mentions": map[string]interface{}{
			"parse": []string{},
		},
	}
	if a.username != "" {
		payload["username"] = a.username
	}
	if a.avatarURL != "" {
		payload["avatar_url"] = a.avatarURL
	}

	if req.ContentType == "image" && req.MediaURL != "" {
		payload["embeds"] = []map[string]interface{}{
			{
				"title":       truncateString(req.Title, 256),
				"description": truncateString(sanitizeDiscordMentions(strings.TrimSpace(req.Body)), 4096),
				"image": map[string]interface{}{
					"url": req.MediaURL,
				},
			},
		}
	}

	sendURL := webhookURL
	if strings.Contains(sendURL, "?") {
		sendURL += "&wait=true"
	} else {
		sendURL += "?wait=true"
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal discord payload failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, sendURL, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("create discord send request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, &DistributionAdapterError{
			Message:   fmt.Sprintf("send discord message failed: %v", err),
			Retriable: true,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read discord send response failed: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, &DistributionAdapterError{
			StatusCode:  resp.StatusCode,
			Message:     fmt.Sprintf("send discord message failed: %s", resp.Status),
			Body:        string(body),
			Retriable:   resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError,
			NeedsRebind: resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
		}
	}

	result := &DiscordSendResponse{}
	if err := json.Unmarshal(body, &result.Raw); err != nil {
		return nil, fmt.Errorf("unmarshal discord send response failed: %w", err)
	}
	if value, _ := result.Raw["id"].(string); value != "" {
		result.ID = value
	}
	if value, _ := result.Raw["channel_id"].(string); value != "" {
		result.ChannelID = value
	}
	if value, _ := result.Raw["guild_id"].(string); value != "" {
		result.GuildID = value
	}
	if value, _ := result.Raw["content"].(string); value != "" {
		result.Content = value
	}
	if webhookID, _ := ExtractDiscordWebhookID(webhookURL); webhookID != "" {
		result.WebhookID = webhookID
	}

	return result, nil
}

func ExtractDiscordWebhookID(webhookURL string) (string, string) {
	matches := discordWebhookPattern.FindStringSubmatch(strings.TrimSpace(webhookURL))
	if len(matches) != 3 {
		return "", ""
	}
	return matches[1], matches[2]
}

func buildDiscordContent(title, body, mediaURL, contentType string) string {
	parts := make([]string, 0, 3)
	if trimmed := sanitizeDiscordMentions(strings.TrimSpace(title)); trimmed != "" {
		parts = append(parts, trimmed)
	}
	if trimmed := sanitizeDiscordMentions(strings.TrimSpace(body)); trimmed != "" {
		parts = append(parts, trimmed)
	}
	if mediaURL != "" && contentType == "video" {
		parts = append(parts, mediaURL)
	}
	if mediaURL != "" && contentType == "image" && len(parts) == 0 {
		parts = append(parts, mediaURL)
	}

	return truncateString(strings.Join(parts, "\n\n"), 1900)
}

func sanitizeDiscordMentions(input string) string {
	replacer := strings.NewReplacer(
		"@everyone", "@\u200beveryone",
		"@here", "@\u200bhere",
		"<@", "<@\u200b",
	)
	return replacer.Replace(input)
}

func truncateString(input string, limit int) string {
	if len(input) <= limit || limit <= 0 {
		return input
	}
	return input[:limit]
}

func normalizeWebhookURL(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	return parsed.String()
}

func distributionSecretKey() string {
	return strings.TrimSpace(os.Getenv("DISTRIBUTION_SECRET_KEY"))
}
