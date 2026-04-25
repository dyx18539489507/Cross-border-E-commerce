package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	maxDistributionAttempts = 3
)

type DistributionService struct {
	db                  *gorm.DB
	cfg                 *config.Config
	log                 *logger.Logger
	uploadPost          *UploadPostAdapter
	discord             *DiscordAdapter
	historyLookbackPage int
}

type CreateDistributionRequest struct {
	SourceType        string                      `json:"sourceType"`
	SourceRef         string                      `json:"sourceRef"`
	ContentType       string                      `json:"contentType"`
	Title             string                      `json:"title"`
	Body              string                      `json:"body"`
	MediaURL          string                      `json:"mediaUrl"`
	MediaRef          string                      `json:"mediaRef"`
	SelectedPlatforms []string                    `json:"selectedPlatforms"`
	PlatformOptions   DistributionPlatformOptions `json:"platformOptions"`
	PublishMode       string                      `json:"publishMode"`
	ScheduledAt       *time.Time                  `json:"scheduledAt"`
}

type DistributionPlatformOptions struct {
	Reddit    DistributionRedditOptions    `json:"reddit"`
	Pinterest DistributionPinterestOptions `json:"pinterest"`
	Discord   DistributionDiscordOptions   `json:"discord"`
}

type DistributionRedditOptions struct {
	Subreddit    string `json:"subreddit"`
	FlairID      string `json:"flairId"`
	FirstComment string `json:"firstComment"`
}

type DistributionPinterestOptions struct {
	BoardID string `json:"boardId"`
}

type DistributionDiscordOptions struct {
	TargetID *uint `json:"targetId"`
}

type UpsertRedditTargetRequest struct {
	Subreddit string `json:"subreddit"`
	FlairID   string `json:"flairId"`
}

type UpsertDiscordTargetRequest struct {
	WebhookURL string `json:"webhookUrl"`
	Name       string `json:"name"`
	IsDefault  bool   `json:"isDefault"`
}

type DistributionTargetsView struct {
	UploadPostProfile *models.UploadPostProfile   `json:"uploadPostProfile,omitempty"`
	Targets           []models.DistributionTarget `json:"targets"`
}

type distributionResolvedMedia struct {
	URL        string
	LocalPath  string
	FileName   string
	SourceType models.DistributionSourceType
	SourceRef  *string
}

type distributionExecutionPayload struct {
	Title            string                 `json:"title,omitempty"`
	Body             string                 `json:"body,omitempty"`
	MediaURL         string                 `json:"media_url,omitempty"`
	MediaPath        string                 `json:"media_path,omitempty"`
	MediaFileName    string                 `json:"media_file_name,omitempty"`
	Subreddit        string                 `json:"subreddit,omitempty"`
	FlairID          string                 `json:"flair_id,omitempty"`
	FirstComment     string                 `json:"first_comment,omitempty"`
	BoardID          string                 `json:"board_id,omitempty"`
	DiscordTargetID  *uint                  `json:"discord_target_id,omitempty"`
	TargetIdentifier string                 `json:"target_identifier,omitempty"`
	TargetName       string                 `json:"target_name,omitempty"`
	TargetConfig     map[string]interface{} `json:"target_config,omitempty"`
}

func NewDistributionService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *DistributionService {
	return &DistributionService{
		db:                  db,
		cfg:                 cfg,
		log:                 log,
		uploadPost:          NewUploadPostAdapter(cfg),
		discord:             NewDiscordAdapter(cfg),
		historyLookbackPage: cfg.Distribution.HistoryLookbackPages,
	}
}

func (s *DistributionService) ListTargets(deviceID string) (*DistributionTargetsView, error) {
	var profile models.UploadPostProfile
	var profilePtr *models.UploadPostProfile
	if err := s.db.Where("device_id = ?", deviceID).First(&profile).Error; err == nil {
		profileCopy := profile
		profilePtr = &profileCopy
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("query upload-post profile failed: %w", err)
	}

	var targets []models.DistributionTarget
	if err := s.db.
		Where("device_id = ?", deviceID).
		Order("platform ASC, target_type ASC, is_default DESC, updated_at DESC").
		Find(&targets).Error; err != nil {
		return nil, fmt.Errorf("query distribution targets failed: %w", err)
	}

	return &DistributionTargetsView{
		UploadPostProfile: profilePtr,
		Targets:           targets,
	}, nil
}

func (s *DistributionService) EnsureUploadPostProfile(ctx context.Context, deviceID string) (*models.UploadPostProfile, error) {
	if !s.uploadPost.IsConfigured() {
		return nil, fmt.Errorf("UPLOAD_POST_API_KEY 未配置")
	}

	profile, err := s.ensureLocalUploadPostProfile(deviceID)
	if err != nil {
		return nil, err
	}

	remoteProfile, err := s.uploadPost.EnsureUserProfile(ctx, profile.Username)
	if err != nil {
		s.markUploadPostProfileError(profile.ID)
		return nil, err
	}

	if err := s.applyUploadPostProfileSnapshot(deviceID, profile, remoteProfile); err != nil {
		return nil, err
	}

	return s.getUploadPostProfile(deviceID)
}

func (s *DistributionService) SyncUploadPostProfile(ctx context.Context, deviceID string) (*models.UploadPostProfile, error) {
	if !s.uploadPost.IsConfigured() {
		return nil, fmt.Errorf("UPLOAD_POST_API_KEY 未配置")
	}

	profile, err := s.ensureLocalUploadPostProfile(deviceID)
	if err != nil {
		return nil, err
	}

	remoteProfile, err := s.uploadPost.GetUserProfile(ctx, profile.Username)
	if err != nil {
		s.markUploadPostProfileError(profile.ID)
		return nil, err
	}

	if err := s.applyUploadPostProfileSnapshot(deviceID, profile, remoteProfile); err != nil {
		return nil, err
	}

	return s.getUploadPostProfile(deviceID)
}

func (s *DistributionService) GenerateUploadPostConnectLink(ctx context.Context, deviceID string) (*models.UploadPostProfile, string, error) {
	profile, err := s.EnsureUploadPostProfile(ctx, deviceID)
	if err != nil {
		return nil, "", err
	}

	resp, err := s.uploadPost.GenerateConnectURL(ctx, UploadPostGenerateLinkRequest{
		Username:           profile.Username,
		RedirectURL:        strings.TrimSpace(s.cfg.Distribution.UploadPostRedirectURL),
		LogoImage:          strings.TrimSpace(s.cfg.Distribution.UploadPostLogoImage),
		ConnectTitle:       strings.TrimSpace(s.cfg.Distribution.UploadPostConnectTitle),
		ConnectDescription: strings.TrimSpace(s.cfg.Distribution.UploadPostConnectDesc),
	})
	if err != nil {
		return nil, "", err
	}

	return profile, resp.AccessURL, nil
}

func (s *DistributionService) ListPinterestBoards(ctx context.Context, deviceID string) ([]models.DistributionTarget, error) {
	profile, err := s.SyncUploadPostProfile(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if !hasConnectedUploadPostPlatform(profile, models.DistributionPlatformPinterest) {
		return nil, fmt.Errorf("请先连接 Pinterest")
	}

	boards, err := s.uploadPost.GetPinterestBoards(ctx, profile.Username)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.DistributionTarget{}).
			Where("device_id = ? AND platform = ? AND target_type = ?", deviceID, models.DistributionPlatformPinterest, models.DistributionTargetTypePinterestBoard).
			Update("status", models.DistributionTargetStatusDisabled).Error; err != nil {
			return err
		}

		var existingDefault models.DistributionTarget
		hasDefault := tx.Where("device_id = ? AND platform = ? AND target_type = ? AND is_default = ?", deviceID, models.DistributionPlatformPinterest, models.DistributionTargetTypePinterestBoard, true).
			First(&existingDefault).Error == nil

		for index, board := range boards {
			configJSON := mustJSON(map[string]interface{}{
				"id":              board.ID,
				"name":            board.Name,
				"description":     board.Description,
				"privacy":         board.Privacy,
				"url":             board.URL,
				"image_cover_url": board.CoverImage,
				"pin_count":       board.PinCount,
				"follower_count":  board.FollowerCount,
			})

			displayName := strings.TrimSpace(board.Name)
			target := models.DistributionTarget{
				DeviceID:        deviceID,
				Platform:        models.DistributionPlatformPinterest,
				TargetType:      models.DistributionTargetTypePinterestBoard,
				Identifier:      board.ID,
				Name:            stringPtr(displayName),
				Status:          models.DistributionTargetStatusActive,
				IsDefault:       !hasDefault && index == 0,
				Config:          configJSON,
				LastValidatedAt: &now,
				LastSyncAt:      &now,
			}

			var existing models.DistributionTarget
			err := tx.Where("device_id = ? AND platform = ? AND target_type = ? AND identifier = ?",
				deviceID,
				models.DistributionPlatformPinterest,
				models.DistributionTargetTypePinterestBoard,
				board.ID,
			).First(&existing).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				if err := tx.Create(&target).Error; err != nil {
					return err
				}
			case err != nil:
				return err
			default:
				updates := map[string]interface{}{
					"name":              target.Name,
					"status":            models.DistributionTargetStatusActive,
					"config":            configJSON,
					"last_validated_at": &now,
					"last_sync_at":      &now,
				}
				if !hasDefault && index == 0 {
					updates["is_default"] = true
				}
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("sync pinterest boards failed: %w", err)
	}

	var targets []models.DistributionTarget
	if err := s.db.
		Where("device_id = ? AND platform = ? AND target_type = ? AND status != ?",
			deviceID,
			models.DistributionPlatformPinterest,
			models.DistributionTargetTypePinterestBoard,
			models.DistributionTargetStatusDisabled,
		).
		Order("is_default DESC, name ASC").
		Find(&targets).Error; err != nil {
		return nil, err
	}

	return targets, nil
}

func (s *DistributionService) SetDefaultTarget(deviceID string, targetID uint) (*models.DistributionTarget, error) {
	var target models.DistributionTarget
	if err := s.db.Where("device_id = ?", deviceID).First(&target, targetID).Error; err != nil {
		return nil, fmt.Errorf("target not found")
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.DistributionTarget{}).
			Where("device_id = ? AND platform = ? AND target_type = ?", deviceID, target.Platform, target.TargetType).
			Update("is_default", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&target).Updates(map[string]interface{}{
			"is_default": true,
			"status":     models.DistributionTargetStatusActive,
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.getDistributionTarget(deviceID, targetID)
}

func (s *DistributionService) SaveRedditDefaultTarget(deviceID string, req UpsertRedditTargetRequest) (*models.DistributionTarget, error) {
	subreddit := normalizeSubreddit(req.Subreddit)
	if subreddit == "" {
		return nil, fmt.Errorf("subreddit 不能为空")
	}

	now := time.Now()
	configJSON := mustJSON(map[string]interface{}{
		"subreddit": subreddit,
		"flair_id":  strings.TrimSpace(req.FlairID),
	})

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.DistributionTarget{}).
			Where("device_id = ? AND platform = ? AND target_type = ?", deviceID, models.DistributionPlatformReddit, models.DistributionTargetTypeRedditSubreddit).
			Update("is_default", false).Error; err != nil {
			return err
		}

		var existing models.DistributionTarget
		err := tx.Where("device_id = ? AND platform = ? AND target_type = ? AND identifier = ?",
			deviceID,
			models.DistributionPlatformReddit,
			models.DistributionTargetTypeRedditSubreddit,
			subreddit,
		).First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			target := models.DistributionTarget{
				DeviceID:        deviceID,
				Platform:        models.DistributionPlatformReddit,
				TargetType:      models.DistributionTargetTypeRedditSubreddit,
				Identifier:      subreddit,
				Name:            stringPtr("r/" + subreddit),
				Status:          models.DistributionTargetStatusActive,
				IsDefault:       true,
				Config:          configJSON,
				LastValidatedAt: &now,
				LastSyncAt:      &now,
			}
			return tx.Create(&target).Error
		case err != nil:
			return err
		default:
			return tx.Model(&existing).Updates(map[string]interface{}{
				"name":              stringPtr("r/" + subreddit),
				"status":            models.DistributionTargetStatusActive,
				"is_default":        true,
				"config":            configJSON,
				"last_validated_at": &now,
				"last_sync_at":      &now,
			}).Error
		}
	}); err != nil {
		return nil, fmt.Errorf("save reddit target failed: %w", err)
	}

	var target models.DistributionTarget
	if err := s.db.
		Where("device_id = ? AND platform = ? AND target_type = ? AND identifier = ?",
			deviceID,
			models.DistributionPlatformReddit,
			models.DistributionTargetTypeRedditSubreddit,
			subreddit,
		).
		First(&target).Error; err != nil {
		return nil, err
	}

	return &target, nil
}

func (s *DistributionService) UpsertDiscordTarget(ctx context.Context, deviceID string, req UpsertDiscordTargetRequest) (*models.DistributionTarget, error) {
	webhookURL := normalizeWebhookURL(req.WebhookURL)
	if webhookURL == "" {
		return nil, fmt.Errorf("Webhook URL 不能为空")
	}

	metadata, err := s.discord.ValidateWebhook(ctx, webhookURL)
	if err != nil {
		return nil, err
	}

	webhookID, _ := ExtractDiscordWebhookID(webhookURL)
	if webhookID == "" {
		webhookID = metadata.ID
	}
	if webhookID == "" {
		return nil, fmt.Errorf("无法识别 Discord webhook")
	}

	encryptedWebhook, err := encryptDistributionSecret(webhookURL)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	displayName := strings.TrimSpace(req.Name)
	if displayName == "" {
		displayName = strings.TrimSpace(metadata.Name)
	}
	if displayName == "" {
		displayName = "Discord Webhook"
	}

	configJSON := mustJSON(map[string]interface{}{
		"webhook_id": webhookID,
		"guild_id":   metadata.GuildID,
		"channel_id": metadata.ChannelID,
		"name":       metadata.Name,
	})

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if req.IsDefault {
			if err := tx.Model(&models.DistributionTarget{}).
				Where("device_id = ? AND platform = ? AND target_type = ?", deviceID, models.DistributionPlatformDiscord, models.DistributionTargetTypeDiscordWebhook).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}

		var existing models.DistributionTarget
		err := tx.Where("device_id = ? AND platform = ? AND target_type = ? AND identifier = ?",
			deviceID,
			models.DistributionPlatformDiscord,
			models.DistributionTargetTypeDiscordWebhook,
			webhookID,
		).First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			target := models.DistributionTarget{
				DeviceID:        deviceID,
				Platform:        models.DistributionPlatformDiscord,
				TargetType:      models.DistributionTargetTypeDiscordWebhook,
				Identifier:      webhookID,
				Name:            stringPtr(displayName),
				Status:          models.DistributionTargetStatusActive,
				IsDefault:       req.IsDefault,
				Config:          configJSON,
				SecretEncrypted: stringPtr(encryptedWebhook),
				LastValidatedAt: &now,
				LastSyncAt:      &now,
			}
			if !req.IsDefault {
				target.IsDefault = s.countTargets(deviceID, models.DistributionPlatformDiscord, models.DistributionTargetTypeDiscordWebhook) == 0
			}
			return tx.Create(&target).Error
		case err != nil:
			return err
		default:
			updates := map[string]interface{}{
				"name":              stringPtr(displayName),
				"status":            models.DistributionTargetStatusActive,
				"config":            configJSON,
				"secret_encrypted":  stringPtr(encryptedWebhook),
				"last_validated_at": &now,
				"last_sync_at":      &now,
			}
			if req.IsDefault {
				updates["is_default"] = true
			}
			return tx.Model(&existing).Updates(updates).Error
		}
	}); err != nil {
		return nil, fmt.Errorf("save discord target failed: %w", err)
	}

	var target models.DistributionTarget
	if err := s.db.
		Where("device_id = ? AND platform = ? AND target_type = ? AND identifier = ?",
			deviceID,
			models.DistributionPlatformDiscord,
			models.DistributionTargetTypeDiscordWebhook,
			webhookID,
		).
		First(&target).Error; err != nil {
		return nil, err
	}

	return &target, nil
}

func (s *DistributionService) DeleteTarget(deviceID string, targetID uint) error {
	return s.db.Where("device_id = ?", deviceID).Delete(&models.DistributionTarget{}, targetID).Error
}

func (s *DistributionService) ensureLocalUploadPostProfile(deviceID string) (*models.UploadPostProfile, error) {
	var profile models.UploadPostProfile
	err := s.db.Where("device_id = ?", deviceID).First(&profile).Error
	switch {
	case err == nil:
		return &profile, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, fmt.Errorf("query upload-post profile failed: %w", err)
	}

	username := uploadPostUsernameForDevice(deviceID)
	profile = models.UploadPostProfile{
		DeviceID: deviceID,
		Username: username,
		Status:   models.UploadPostProfileStatusPending,
	}
	if err := s.db.Create(&profile).Error; err != nil {
		return nil, fmt.Errorf("create upload-post profile failed: %w", err)
	}

	return &profile, nil
}

func (s *DistributionService) getUploadPostProfile(deviceID string) (*models.UploadPostProfile, error) {
	var profile models.UploadPostProfile
	if err := s.db.Where("device_id = ?", deviceID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *DistributionService) markUploadPostProfileError(profileID uint) {
	_ = s.db.Model(&models.UploadPostProfile{}).
		Where("id = ?", profileID).
		Update("status", models.UploadPostProfileStatusError).Error
}

func (s *DistributionService) applyUploadPostProfileSnapshot(deviceID string, profile *models.UploadPostProfile, remoteProfile *UploadPostProfileResponse) error {
	connectedPlatforms, needsRebind := extractUploadPostPlatformState(remoteProfile)
	now := time.Now()
	status := models.UploadPostProfileStatusPending
	if len(connectedPlatforms) > 0 {
		status = models.UploadPostProfileStatusActive
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.UploadPostProfile{}).
			Where("id = ?", profile.ID).
			Updates(map[string]interface{}{
				"status":              status,
				"connected_platforms": mustJSON(connectedPlatforms),
				"profile_snapshot":    mustJSON(remoteProfile),
				"last_sync_at":        &now,
			}).Error; err != nil {
			return err
		}

		for _, platform := range []models.DistributionPlatform{
			models.DistributionPlatformPinterest,
			models.DistributionPlatformReddit,
		} {
			targetStatus := models.DistributionTargetStatusNeedsRebind
			if containsString(connectedPlatforms, string(platform)) && !needsRebind[string(platform)] {
				targetStatus = models.DistributionTargetStatusActive
			}
			if err := tx.Model(&models.DistributionTarget{}).
				Where("device_id = ? AND platform = ?", deviceID, platform).
				Update("status", targetStatus).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("save upload-post profile snapshot failed: %w", err)
	}

	return nil
}

func (s *DistributionService) getDistributionTarget(deviceID string, targetID uint) (*models.DistributionTarget, error) {
	var target models.DistributionTarget
	if err := s.db.Where("device_id = ?", deviceID).First(&target, targetID).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *DistributionService) countTargets(deviceID string, platform models.DistributionPlatform, targetType models.DistributionTargetType) int64 {
	var count int64
	_ = s.db.Model(&models.DistributionTarget{}).
		Where("device_id = ? AND platform = ? AND target_type = ?", deviceID, platform, targetType).
		Count(&count).Error
	return count
}

func uploadPostUsernameForDevice(deviceID string) string {
	sum := sha1.Sum([]byte(deviceID))
	return "drama_" + hex.EncodeToString(sum[:12])
}

func extractUploadPostPlatformState(profile *UploadPostProfileResponse) ([]string, map[string]bool) {
	connected := make([]string, 0)
	needsRebind := make(map[string]bool)
	if profile == nil {
		return connected, needsRebind
	}

	for platform, value := range profile.SocialAccounts {
		if value == nil {
			continue
		}

		connected = append(connected, platform)
		if payload, ok := value.(map[string]interface{}); ok {
			if reauth, ok := payload["reauth_required"].(bool); ok && reauth {
				needsRebind[platform] = true
			}
		}
	}

	return connected, needsRebind
}

func hasConnectedUploadPostPlatform(profile *models.UploadPostProfile, platform models.DistributionPlatform) bool {
	if profile == nil || len(profile.ConnectedPlatforms) == 0 {
		return false
	}

	var items []string
	if err := json.Unmarshal(profile.ConnectedPlatforms, &items); err != nil {
		return false
	}

	return containsString(items, string(platform))
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}

func normalizeSubreddit(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	trimmed = strings.TrimPrefix(trimmed, "r/")
	return trimmed
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func mustJSON(value interface{}) datatypes.JSON {
	if value == nil {
		return datatypes.JSON([]byte("null"))
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte("null"))
	}
	return datatypes.JSON(raw)
}

func parseUintID(value string) (uint, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func resolveFileName(pathOrURL string) string {
	if pathOrURL == "" {
		return ""
	}

	if parsedURL, err := url.Parse(pathOrURL); err == nil && parsedURL.Path != "" {
		base := filepath.Base(parsedURL.Path)
		if base != "." && base != "/" {
			return base
		}
	}

	base := filepath.Base(pathOrURL)
	if base == "." || base == "/" {
		return ""
	}

	return base
}

func (s *DistributionService) CreateDistribution(ctx context.Context, deviceID string, req *CreateDistributionRequest) (*models.DistributionJob, error) {
	if req == nil {
		return nil, fmt.Errorf("请求不能为空")
	}

	contentType, err := normalizeContentType(req.ContentType)
	if err != nil {
		return nil, err
	}

	publishMode, err := normalizePublishMode(req.PublishMode)
	if err != nil {
		return nil, err
	}
	if publishMode == models.DistributionPublishModeSchedule {
		if req.ScheduledAt == nil {
			return nil, fmt.Errorf("scheduledAt 不能为空")
		}
		if req.ScheduledAt.Before(time.Now().Add(10 * time.Second)) {
			return nil, fmt.Errorf("scheduledAt 必须晚于当前时间")
		}
	}

	platforms, err := normalizeDistributionPlatformInputs(req.SelectedPlatforms)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	body := strings.TrimSpace(req.Body)
	if contentType == models.DistributionContentTypeText {
		if title == "" && body == "" {
			return nil, fmt.Errorf("文本分发至少需要标题或正文")
		}
		if containsPlatform(platforms, models.DistributionPlatformReddit) && title == "" {
			return nil, fmt.Errorf("Reddit text-only 分发需要 title")
		}
	}

	media, err := s.resolveDistributionMedia(req, contentType)
	if err != nil {
		return nil, err
	}

	sourceType := normalizeSourceType(req.SourceType)
	sourceRef := strings.TrimSpace(req.SourceRef)
	if media.SourceType != "" && sourceType == models.DistributionSourceTypeManual {
		sourceType = media.SourceType
	}
	if media.SourceRef != nil && sourceRef == "" {
		sourceRef = *media.SourceRef
	}

	var profile *models.UploadPostProfile
	if includesUploadPostPlatform(platforms) {
		profile, err = s.SyncUploadPostProfile(ctx, deviceID)
		if err != nil {
			return nil, err
		}
	}

	type preparedResult struct {
		platform       models.DistributionPlatform
		targetID       *uint
		targetSnapshot datatypes.JSON
		requestPayload datatypes.JSON
		initialStatus  models.DistributionResultStatus
	}

	prepared := make([]preparedResult, 0, len(platforms))
	for _, platform := range platforms {
		targetID, targetSnapshot, requestPayload, initialStatus, err := s.prepareDistributionResult(ctx, deviceID, profile, platform, contentType, publishMode, req, &media, title, body)
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, preparedResult{
			platform:       platform,
			targetID:       targetID,
			targetSnapshot: targetSnapshot,
			requestPayload: requestPayload,
			initialStatus:  initialStatus,
		})
	}

	selectedPlatformsJSON := mustJSON(platforms)
	platformOptionsJSON := mustJSON(req.PlatformOptions)
	requestSnapshot := mustJSON(map[string]interface{}{
		"content_type":       contentType,
		"title":              title,
		"body":               body,
		"media_url":          media.URL,
		"media_ref":          strings.TrimSpace(req.MediaRef),
		"selected_platforms": platforms,
		"platform_options":   req.PlatformOptions,
		"publish_mode":       publishMode,
		"scheduled_at":       req.ScheduledAt,
	})

	status := models.DistributionJobStatusPending
	if publishMode == models.DistributionPublishModeSchedule {
		status = models.DistributionJobStatusScheduled
	}

	job := models.DistributionJob{
		DeviceID:          deviceID,
		SourceType:        sourceType,
		ContentType:       contentType,
		Title:             stringPtr(title),
		Body:              stringPtr(body),
		MediaURL:          stringPtr(firstNonEmpty(media.URL, req.MediaURL)),
		SelectedPlatforms: selectedPlatformsJSON,
		PlatformOptions:   platformOptionsJSON,
		PublishMode:       publishMode,
		ScheduledAt:       req.ScheduledAt,
		Status:            status,
		RequestSnapshot:   requestSnapshot,
	}
	if sourceRef != "" {
		job.SourceRef = stringPtr(sourceRef)
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&job).Error; err != nil {
			return err
		}

		for _, item := range prepared {
			result := models.DistributionResult{
				JobID:           job.ID,
				DeviceID:        deviceID,
				Platform:        item.platform,
				TargetID:        item.targetID,
				ContentType:     contentType,
				Status:          item.initialStatus,
				TargetSnapshot:  item.targetSnapshot,
				RequestSnapshot: item.requestPayload,
			}
			if err := tx.Create(&result).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("create distribution job failed: %w", err)
	}

	go s.processJobAsync(job.ID)

	return s.GetDistributionJob(deviceID, job.ID)
}

func (s *DistributionService) ListDistributionJobs(deviceID string, page, pageSize int) ([]models.DistributionJob, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	query := s.db.Model(&models.DistributionJob{}).Where("device_id = ?", deviceID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count distribution jobs failed: %w", err)
	}

	var jobs []models.DistributionJob
	if err := s.db.
		Where("device_id = ?", deviceID).
		Preload("Results", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Preload("Target")
		}).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("list distribution jobs failed: %w", err)
	}

	return jobs, total, nil
}

func (s *DistributionService) GetDistributionJob(deviceID string, jobID uint) (*models.DistributionJob, error) {
	var job models.DistributionJob
	if err := s.db.
		Where("device_id = ?", deviceID).
		Preload("Results", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Preload("Target")
		}).
		First(&job, jobID).Error; err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *DistributionService) RetryDistribution(ctx context.Context, deviceID string, jobID uint) (*models.DistributionJob, error) {
	job, err := s.GetDistributionJob(deviceID, jobID)
	if err != nil {
		return nil, err
	}

	resultIDs := make([]uint, 0)
	for _, result := range job.Results {
		if result.Status == models.DistributionResultStatusFailed {
			resultIDs = append(resultIDs, result.ID)
		}
	}
	if len(resultIDs) == 0 {
		return job, nil
	}

	if err := s.db.Model(&models.DistributionResult{}).
		Where("device_id = ? AND id IN ?", deviceID, resultIDs).
		Updates(map[string]interface{}{
			"status":              models.DistributionResultStatusPending,
			"error_msg":           nil,
			"response_snapshot":   datatypes.JSON([]byte("null")),
			"external_request_id": nil,
			"external_job_id":     nil,
			"external_message_id": nil,
			"published_url":       nil,
			"completed_at":        nil,
			"next_retry_at":       nil,
		}).Error; err != nil {
		return nil, fmt.Errorf("reset distribution results failed: %w", err)
	}

	if err := s.db.Model(&models.DistributionJob{}).Where("device_id = ? AND id = ?", deviceID, jobID).
		Updates(map[string]interface{}{
			"status":       models.DistributionJobStatusPending,
			"error_msg":    nil,
			"completed_at": nil,
		}).Error; err != nil {
		return nil, fmt.Errorf("reset distribution job failed: %w", err)
	}

	go s.processJobAsync(jobID)

	return s.GetDistributionJob(deviceID, jobID)
}

func (s *DistributionService) ProcessPendingWork(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = 50
	}

	var results []models.DistributionResult
	if err := s.db.
		Preload("Job").
		Preload("Target").
		Where("status IN ?", []models.DistributionResultStatus{
			models.DistributionResultStatusPending,
			models.DistributionResultStatusProcessing,
			models.DistributionResultStatusScheduled,
		}).
		Order("updated_at ASC").
		Limit(limit).
		Find(&results).Error; err != nil {
		return fmt.Errorf("query pending distribution results failed: %w", err)
	}

	for _, result := range results {
		resultCopy := result
		if resultCopy.NextRetryAt != nil && resultCopy.NextRetryAt.After(time.Now()) {
			continue
		}
		if err := s.processResult(ctx, &resultCopy.Job, &resultCopy); err != nil {
			s.log.Errorw("Process distribution result failed", "error", err, "result_id", resultCopy.ID)
		}
	}

	return nil
}

func (s *DistributionService) processJobAsync(jobID uint) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	var job models.DistributionJob
	if err := s.db.
		Preload("Results", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Preload("Target")
		}).
		First(&job, jobID).Error; err != nil {
		s.log.Errorw("Load distribution job failed", "error", err, "job_id", jobID)
		return
	}

	for _, result := range job.Results {
		resultCopy := result
		if err := s.processResult(ctx, &job, &resultCopy); err != nil {
			s.log.Errorw("Process distribution result failed", "error", err, "job_id", jobID, "result_id", resultCopy.ID, "platform", resultCopy.Platform)
		}
	}

	if err := s.refreshJobStatus(jobID); err != nil {
		s.log.Errorw("Refresh distribution job status failed", "error", err, "job_id", jobID)
	}
}

func (s *DistributionService) findPinterestBoardTarget(deviceID, boardID string) (*models.DistributionTarget, error) {
	query := s.db.Where("device_id = ? AND platform = ? AND target_type = ?",
		deviceID,
		models.DistributionPlatformPinterest,
		models.DistributionTargetTypePinterestBoard,
	)
	if strings.TrimSpace(boardID) != "" {
		query = query.Where("identifier = ?", strings.TrimSpace(boardID))
	} else {
		query = query.Where("is_default = ?", true)
	}

	var target models.DistributionTarget
	if err := query.First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("请先选择 Pinterest board")
		}
		return nil, err
	}
	return &target, nil
}

func (s *DistributionService) findRedditTarget(deviceID, subreddit string) (*models.DistributionTarget, error) {
	query := s.db.Where("device_id = ? AND platform = ? AND target_type = ?",
		deviceID,
		models.DistributionPlatformReddit,
		models.DistributionTargetTypeRedditSubreddit,
	)
	if subreddit != "" {
		query = query.Where("identifier = ?", subreddit)
	} else {
		query = query.Where("is_default = ?", true)
	}

	var target models.DistributionTarget
	if err := query.First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if subreddit != "" {
				return nil, nil
			}
			return nil, fmt.Errorf("请先配置默认 subreddit 或在发布时填写 subreddit")
		}
		return nil, err
	}
	return &target, nil
}

func (s *DistributionService) findDiscordTarget(deviceID string, targetID *uint) (*models.DistributionTarget, error) {
	query := s.db.Where("device_id = ? AND platform = ? AND target_type = ?",
		deviceID,
		models.DistributionPlatformDiscord,
		models.DistributionTargetTypeDiscordWebhook,
	)
	if targetID != nil {
		query = query.Where("id = ?", *targetID)
	} else {
		query = query.Where("is_default = ?", true)
	}

	var target models.DistributionTarget
	if err := query.First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("请先配置 Discord webhook 目标")
		}
		return nil, err
	}
	return &target, nil
}

func (s *DistributionService) resolveDistributionMedia(req *CreateDistributionRequest, contentType models.DistributionContentType) (distributionResolvedMedia, error) {
	if contentType == models.DistributionContentTypeText {
		return distributionResolvedMedia{}, nil
	}

	mediaURL := strings.TrimSpace(req.MediaURL)
	if mediaURL != "" {
		resolved := distributionResolvedMedia{
			URL:       mediaURL,
			LocalPath: s.resolveLocalStoragePath(mediaURL),
			FileName:  resolveFileName(mediaURL),
		}
		if resolved.LocalPath != "" && resolved.FileName == "" {
			resolved.FileName = filepath.Base(resolved.LocalPath)
		}
		return resolved, nil
	}

	sourceType := normalizeSourceType(req.SourceType)
	sourceRef := strings.TrimSpace(req.SourceRef)
	if strings.TrimSpace(req.MediaRef) != "" {
		if parsedType, parsedRef := parseMediaRef(req.MediaRef); parsedType != "" {
			sourceType = parsedType
			sourceRef = parsedRef
		}
	}

	if sourceRef == "" {
		return distributionResolvedMedia{}, fmt.Errorf("媒体资源不能为空")
	}

	switch sourceType {
	case models.DistributionSourceTypeImageGen:
		var image models.ImageGeneration
		id, err := parseUintID(sourceRef)
		if err != nil {
			return distributionResolvedMedia{}, fmt.Errorf("无效的 image_generation ID")
		}
		if err := s.db.First(&image, id).Error; err != nil {
			return distributionResolvedMedia{}, err
		}
		return distributionResolvedMedia{
			URL:        firstNonEmpty(derefString(image.MinioURL), derefString(image.ImageURL)),
			LocalPath:  derefString(image.LocalPath),
			FileName:   resolveFileName(firstNonEmpty(derefString(image.LocalPath), derefString(image.MinioURL), derefString(image.ImageURL))),
			SourceType: models.DistributionSourceTypeImageGen,
			SourceRef:  stringPtr(sourceRef),
		}, nil
	case models.DistributionSourceTypeVideoGen:
		var video models.VideoGeneration
		id, err := parseUintID(sourceRef)
		if err != nil {
			return distributionResolvedMedia{}, fmt.Errorf("无效的 video_generation ID")
		}
		if err := s.db.First(&video, id).Error; err != nil {
			return distributionResolvedMedia{}, err
		}
		return distributionResolvedMedia{
			URL:        firstNonEmpty(derefString(video.MinioURL), derefString(video.VideoURL)),
			LocalPath:  derefString(video.LocalPath),
			FileName:   resolveFileName(firstNonEmpty(derefString(video.LocalPath), derefString(video.MinioURL), derefString(video.VideoURL))),
			SourceType: models.DistributionSourceTypeVideoGen,
			SourceRef:  stringPtr(sourceRef),
		}, nil
	case models.DistributionSourceTypeVideoMerge:
		var merge models.VideoMerge
		id, err := parseUintID(sourceRef)
		if err != nil {
			return distributionResolvedMedia{}, fmt.Errorf("无效的 video_merge ID")
		}
		if err := s.db.First(&merge, id).Error; err != nil {
			return distributionResolvedMedia{}, err
		}
		mergedURL := derefString(merge.MergedURL)
		return distributionResolvedMedia{
			URL:        mergedURL,
			LocalPath:  s.resolveLocalStoragePath(mergedURL),
			FileName:   resolveFileName(mergedURL),
			SourceType: models.DistributionSourceTypeVideoMerge,
			SourceRef:  stringPtr(sourceRef),
		}, nil
	case models.DistributionSourceTypeAsset:
		var asset models.Asset
		id, err := parseUintID(sourceRef)
		if err != nil {
			return distributionResolvedMedia{}, fmt.Errorf("无效的 asset ID")
		}
		if err := s.db.First(&asset, id).Error; err != nil {
			return distributionResolvedMedia{}, err
		}
		return distributionResolvedMedia{
			URL:        asset.URL,
			LocalPath:  derefString(asset.LocalPath),
			FileName:   resolveFileName(firstNonEmpty(derefString(asset.LocalPath), asset.URL)),
			SourceType: models.DistributionSourceTypeAsset,
			SourceRef:  stringPtr(sourceRef),
		}, nil
	default:
		return distributionResolvedMedia{}, fmt.Errorf("不支持的媒体引用类型")
	}
}

func (s *DistributionService) resolveLocalStoragePath(mediaURL string) string {
	trimmed := strings.TrimSpace(mediaURL)
	if trimmed == "" {
		return ""
	}
	if filepath.IsAbs(trimmed) {
		return trimmed
	}

	if strings.HasPrefix(trimmed, "/static/") {
		return filepath.Join(s.cfg.Storage.LocalPath, strings.TrimPrefix(trimmed, "/static/"))
	}

	baseURL := strings.TrimRight(strings.TrimSpace(s.cfg.Storage.BaseURL), "/")
	if baseURL != "" && strings.HasPrefix(trimmed, baseURL+"/") {
		relative := strings.TrimPrefix(trimmed, baseURL+"/")
		return filepath.Join(s.cfg.Storage.LocalPath, relative)
	}

	return ""
}

func (s *DistributionService) loadResultTarget(result *models.DistributionResult) (*models.DistributionTarget, error) {
	if result.Target != nil {
		return result.Target, nil
	}
	if result.TargetID == nil {
		return nil, fmt.Errorf("分发目标不存在")
	}

	var target models.DistributionTarget
	if err := s.db.First(&target, *result.TargetID).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *DistributionService) handleSubmissionError(job *models.DistributionJob, result *models.DistributionResult, err error) error {
	var adapterErr *DistributionAdapterError
	retriable := false
	needsRebind := false
	errorMessage := err.Error()
	if errors.As(err, &adapterErr) {
		retriable = adapterErr.Retriable
		needsRebind = adapterErr.NeedsRebind
		errorMessage = firstNonEmpty(adapterErr.Body, adapterErr.Message, errorMessage)
	}

	if needsRebind {
		s.markTargetNeedsRebind(result)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"error_msg": stringPtr(errorMessage),
	}
	if retriable && result.AttemptCount < maxDistributionAttempts {
		nextRetry := now.Add(retryDelay(result.AttemptCount))
		updates["status"] = models.DistributionResultStatusPending
		updates["next_retry_at"] = &nextRetry
		updates["completed_at"] = nil
	} else {
		updates["status"] = models.DistributionResultStatusFailed
		updates["completed_at"] = &now
	}

	if err := s.updateResultStatus(result.ID, updates); err != nil {
		return err
	}

	return s.refreshJobStatus(job.ID)
}

func (s *DistributionService) scheduleSoftRefresh(resultID uint) error {
	nextRetry := time.Now().Add(30 * time.Second)
	return s.updateResultStatus(resultID, map[string]interface{}{
		"next_retry_at": &nextRetry,
	})
}

func (s *DistributionService) updateResultStatus(resultID uint, updates map[string]interface{}) error {
	return s.db.Model(&models.DistributionResult{}).Where("id = ?", resultID).Updates(updates).Error
}

func (s *DistributionService) findUploadPostHistory(ctx context.Context, result *models.DistributionResult) (*UploadPostHistoryItem, error) {
	pages := s.historyLookbackPage
	if pages <= 0 {
		pages = 3
	}

	for page := 1; page <= pages; page++ {
		resp, err := s.uploadPost.GetUploadHistory(ctx, page, 50)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.History {
			if !strings.EqualFold(item.Platform, string(result.Platform)) {
				continue
			}
			if result.ExternalRequestID != nil && item.RequestID != "" && item.RequestID == *result.ExternalRequestID {
				itemCopy := item
				return &itemCopy, nil
			}
			if result.ExternalJobID != nil && item.JobID != "" && item.JobID == *result.ExternalJobID {
				itemCopy := item
				return &itemCopy, nil
			}
		}
	}

	return nil, gorm.ErrRecordNotFound
}

func matchUploadPostStatusResult(items []UploadPostStatusItem, platform string) *UploadPostStatusItem {
	for _, item := range items {
		if strings.EqualFold(item.Platform, platform) {
			itemCopy := item
			return &itemCopy
		}
	}
	return nil
}

func parseExecutionPayload(raw datatypes.JSON) (distributionExecutionPayload, error) {
	var payload distributionExecutionPayload
	if len(raw) == 0 {
		return payload, nil
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, fmt.Errorf("parse request snapshot failed: %w", err)
	}
	return payload, nil
}

func normalizeContentType(value string) (models.DistributionContentType, error) {
	switch models.DistributionContentType(strings.ToLower(strings.TrimSpace(value))) {
	case models.DistributionContentTypeText:
		return models.DistributionContentTypeText, nil
	case models.DistributionContentTypeImage:
		return models.DistributionContentTypeImage, nil
	case models.DistributionContentTypeVideo:
		return models.DistributionContentTypeVideo, nil
	default:
		return "", fmt.Errorf("不支持的 contentType")
	}
}

func normalizePublishMode(value string) (models.DistributionPublishMode, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return models.DistributionPublishModeImmediate, nil
	}
	switch models.DistributionPublishMode(trimmed) {
	case models.DistributionPublishModeImmediate:
		return models.DistributionPublishModeImmediate, nil
	case models.DistributionPublishModeSchedule:
		return models.DistributionPublishModeSchedule, nil
	default:
		return "", fmt.Errorf("不支持的 publishMode")
	}
}

func includesUploadPostPlatform(platforms []models.DistributionPlatform) bool {
	return containsPlatform(platforms, models.DistributionPlatformPinterest) || containsPlatform(platforms, models.DistributionPlatformReddit)
}

func containsPlatform(platforms []models.DistributionPlatform, target models.DistributionPlatform) bool {
	for _, platform := range platforms {
		if platform == target {
			return true
		}
	}
	return false
}

func normalizeSourceType(value string) models.DistributionSourceType {
	switch models.DistributionSourceType(strings.TrimSpace(strings.ToLower(value))) {
	case models.DistributionSourceTypeVideoMerge:
		return models.DistributionSourceTypeVideoMerge
	case models.DistributionSourceTypeImageGen:
		return models.DistributionSourceTypeImageGen
	case models.DistributionSourceTypeVideoGen:
		return models.DistributionSourceTypeVideoGen
	case models.DistributionSourceTypeAsset:
		return models.DistributionSourceTypeAsset
	default:
		return models.DistributionSourceTypeManual
	}
}

func parseMediaRef(value string) (models.DistributionSourceType, string) {
	parts := strings.SplitN(strings.TrimSpace(value), ":", 2)
	if len(parts) != 2 {
		return models.DistributionSourceTypeManual, ""
	}
	return normalizeSourceType(parts[0]), strings.TrimSpace(parts[1])
}

func retryDelay(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 30 * time.Second
	case 2:
		return 2 * time.Minute
	default:
		return 10 * time.Minute
	}
}

func boardTargetSnapshot(target *models.DistributionTarget) map[string]interface{} {
	return map[string]interface{}{
		"id":         target.ID,
		"identifier": target.Identifier,
		"name":       derefString(target.Name),
		"config":     jsonToMap(target.Config),
	}
}

func redditTargetSnapshot(target *models.DistributionTarget, subreddit, flairID string) map[string]interface{} {
	snapshot := map[string]interface{}{
		"subreddit": subreddit,
		"flair_id":  flairID,
	}
	if target != nil {
		snapshot["id"] = target.ID
		snapshot["identifier"] = target.Identifier
		snapshot["name"] = derefString(target.Name)
		snapshot["config"] = jsonToMap(target.Config)
	}
	return snapshot
}

func discordTargetSnapshot(target *models.DistributionTarget) map[string]interface{} {
	return map[string]interface{}{
		"id":         target.ID,
		"identifier": target.Identifier,
		"name":       derefString(target.Name),
		"config":     jsonToMap(target.Config),
	}
}

func jsonToMap(raw datatypes.JSON) map[string]interface{} {
	if len(raw) == 0 {
		return map[string]interface{}{}
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]interface{}{}
	}
	return payload
}

func readTargetConfigString(raw datatypes.JSON, key string) string {
	return strings.TrimSpace(fmt.Sprint(jsonToMap(raw)[key]))
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func timePtrValue(value *time.Time, fallback time.Time) *time.Time {
	if value != nil {
		return value
	}
	return &fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func nilJSON() datatypes.JSON {
	return datatypes.JSON([]byte("null"))
}

func uintPtr(value uint) *uint {
	return &value
}

func uintPtrValue(value uint) *uint {
	return &value
}

func discordMessageURL(guildID, channelID, messageID string) string {
	if guildID == "" || channelID == "" || messageID == "" {
		return ""
	}
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)
}

func (s *DistributionService) prepareDistributionResult(
	ctx context.Context,
	deviceID string,
	profile *models.UploadPostProfile,
	platform models.DistributionPlatform,
	contentType models.DistributionContentType,
	publishMode models.DistributionPublishMode,
	req *CreateDistributionRequest,
	media *distributionResolvedMedia,
	title string,
	body string,
) (*uint, datatypes.JSON, datatypes.JSON, models.DistributionResultStatus, error) {
	targetSnapshot := mustJSON(map[string]interface{}{
		"platform": platform,
	})
	initialStatus := models.DistributionResultStatusPending
	if publishMode == models.DistributionPublishModeSchedule {
		initialStatus = models.DistributionResultStatusScheduled
	}

	payload := distributionExecutionPayload{
		Title:         title,
		Body:          body,
		MediaFileName: "",
	}
	if media != nil {
		payload.MediaURL = media.URL
		payload.MediaPath = media.LocalPath
		payload.MediaFileName = media.FileName
	}

	switch platform {
	case models.DistributionPlatformPinterest:
		if contentType == models.DistributionContentTypeText {
			return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("Pinterest 不支持 text-only 分发")
		}
		if profile == nil || !hasConnectedUploadPostPlatform(profile, models.DistributionPlatformPinterest) {
			return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("请先连接 Pinterest")
		}

		boardID := strings.TrimSpace(req.PlatformOptions.Pinterest.BoardID)
		var target *models.DistributionTarget
		var err error
		if boardID == "" {
			target, err = s.getDefaultTarget(deviceID, models.DistributionPlatformPinterest, models.DistributionTargetTypePinterestBoard)
			if err != nil {
				return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("请先选择默认 Pinterest board")
			}
			boardID = target.Identifier
		} else {
			target, _ = s.findTargetByIdentifier(deviceID, models.DistributionPlatformPinterest, models.DistributionTargetTypePinterestBoard, boardID)
		}
		if boardID == "" {
			return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("Pinterest board 不能为空")
		}

		payload.BoardID = boardID
		if target != nil {
			payload.TargetIdentifier = target.Identifier
			payload.TargetName = derefString(target.Name)
			payload.TargetConfig = jsonMap(target.Config)
			targetSnapshot = mustJSON(map[string]interface{}{
				"id":         target.ID,
				"identifier": target.Identifier,
				"name":       target.Name,
				"config":     jsonMap(target.Config),
			})
			return &target.ID, targetSnapshot, mustJSON(payload), initialStatus, nil
		}
		targetSnapshot = mustJSON(map[string]interface{}{
			"identifier": boardID,
		})
		return nil, targetSnapshot, mustJSON(payload), initialStatus, nil

	case models.DistributionPlatformReddit:
		if profile == nil || !hasConnectedUploadPostPlatform(profile, models.DistributionPlatformReddit) {
			return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("请先连接 Reddit")
		}
		subreddit := normalizeSubreddit(req.PlatformOptions.Reddit.Subreddit)
		var target *models.DistributionTarget
		var err error
		if subreddit == "" {
			target, err = s.getDefaultTarget(deviceID, models.DistributionPlatformReddit, models.DistributionTargetTypeRedditSubreddit)
			if err != nil {
				return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("请先配置默认 subreddit")
			}
			subreddit = target.Identifier
		} else {
			target, _ = s.findTargetByIdentifier(deviceID, models.DistributionPlatformReddit, models.DistributionTargetTypeRedditSubreddit, subreddit)
		}
		if subreddit == "" {
			return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("subreddit 不能为空")
		}
		if title == "" {
			return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("Reddit 分发需要 title")
		}

		payload.Subreddit = subreddit
		payload.FlairID = strings.TrimSpace(req.PlatformOptions.Reddit.FlairID)
		payload.FirstComment = strings.TrimSpace(req.PlatformOptions.Reddit.FirstComment)
		if payload.FlairID == "" && target != nil {
			if flairID, ok := jsonMap(target.Config)["flair_id"].(string); ok {
				payload.FlairID = flairID
			}
		}

		if target != nil {
			payload.TargetIdentifier = target.Identifier
			payload.TargetName = derefString(target.Name)
			payload.TargetConfig = jsonMap(target.Config)
			targetSnapshot = mustJSON(map[string]interface{}{
				"id":         target.ID,
				"identifier": target.Identifier,
				"name":       target.Name,
				"config":     jsonMap(target.Config),
			})
			return &target.ID, targetSnapshot, mustJSON(payload), initialStatus, nil
		}

		targetSnapshot = mustJSON(map[string]interface{}{
			"identifier": subreddit,
			"flair_id":   payload.FlairID,
		})
		return nil, targetSnapshot, mustJSON(payload), initialStatus, nil

	case models.DistributionPlatformDiscord:
		targetID := req.PlatformOptions.Discord.TargetID
		var target *models.DistributionTarget
		var err error
		if targetID != nil {
			target, err = s.getDistributionTarget(deviceID, *targetID)
		} else {
			target, err = s.getDefaultTarget(deviceID, models.DistributionPlatformDiscord, models.DistributionTargetTypeDiscordWebhook)
		}
		if err != nil || target == nil {
			return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("请先连接 Discord 频道目标")
		}

		payload.DiscordTargetID = &target.ID
		payload.TargetIdentifier = target.Identifier
		payload.TargetName = derefString(target.Name)
		payload.TargetConfig = jsonMap(target.Config)
		targetSnapshot = mustJSON(map[string]interface{}{
			"id":         target.ID,
			"identifier": target.Identifier,
			"name":       target.Name,
			"config":     jsonMap(target.Config),
		})
		return &target.ID, targetSnapshot, mustJSON(payload), initialStatus, nil

	default:
		return nil, nilJSON(), nilJSON(), initialStatus, fmt.Errorf("暂不支持的平台: %s", platform)
	}
}

func (s *DistributionService) processResult(ctx context.Context, job *models.DistributionJob, result *models.DistributionResult) error {
	if job == nil || result == nil {
		return nil
	}
	if result.Status == models.DistributionResultStatusSuccess {
		return nil
	}
	if result.NextRetryAt != nil && result.NextRetryAt.After(time.Now()) {
		return nil
	}

	if isUploadPostPlatform(result.Platform) && (hasValue(result.ExternalRequestID) || hasValue(result.ExternalJobID)) {
		return s.refreshUploadPostResult(ctx, job, result)
	}

	if result.Platform == models.DistributionPlatformDiscord &&
		job.PublishMode == models.DistributionPublishModeSchedule &&
		job.ScheduledAt != nil &&
		job.ScheduledAt.After(time.Now()) {
		return s.updateDistributionResult(result.ID, map[string]interface{}{
			"status": models.DistributionResultStatusScheduled,
		})
	}

	if err := s.updateDistributionResult(result.ID, map[string]interface{}{
		"status":        models.DistributionResultStatusProcessing,
		"started_at":    time.Now(),
		"attempt_count": gorm.Expr("attempt_count + 1"),
		"next_retry_at": nil,
	}); err != nil {
		return err
	}
	result.AttemptCount++

	switch result.Platform {
	case models.DistributionPlatformDiscord:
		return s.dispatchDiscordResult(ctx, job, result)
	case models.DistributionPlatformPinterest, models.DistributionPlatformReddit:
		return s.dispatchUploadPostResult(ctx, job, result)
	default:
		return s.markResultFailed(result, fmt.Errorf("unsupported platform: %s", result.Platform), false, false)
	}
}

func (s *DistributionService) dispatchUploadPostResult(ctx context.Context, job *models.DistributionJob, result *models.DistributionResult) error {
	var payload distributionExecutionPayload
	if err := json.Unmarshal(result.RequestSnapshot, &payload); err != nil {
		return s.markResultFailed(result, fmt.Errorf("解析分发请求失败: %w", err), false, false)
	}

	profile, err := s.getUploadPostProfile(result.DeviceID)
	if err != nil {
		return s.markResultFailed(result, fmt.Errorf("Upload-Post profile 不存在"), false, true)
	}

	req := UploadPostPublishRequest{
		Username:           profile.Username,
		Platform:           result.Platform,
		ContentType:        job.ContentType,
		Title:              payload.Title,
		Body:               payload.Body,
		MediaURL:           payload.MediaURL,
		MediaPath:          payload.MediaPath,
		MediaFileName:      payload.MediaFileName,
		AsyncUpload:        job.PublishMode == models.DistributionPublishModeImmediate,
		RedditSubreddit:    payload.Subreddit,
		RedditFlairID:      payload.FlairID,
		RedditFirstComment: payload.FirstComment,
		PinterestBoardID:   payload.BoardID,
	}
	if job.PublishMode == models.DistributionPublishModeSchedule {
		req.AsyncUpload = false
		req.ScheduledAt = job.ScheduledAt
	}

	resp, err := s.uploadPost.Upload(ctx, req)
	if err != nil {
		return s.markResultFailed(result, err, true, needsRebind(err))
	}

	updatePayload := map[string]interface{}{
		"response_snapshot": mustJSON(resp.Raw),
		"error_msg":         nil,
	}

	switch {
	case resp.RequestID != "":
		updatePayload["status"] = models.DistributionResultStatusProcessing
		updatePayload["external_request_id"] = resp.RequestID
	case resp.JobID != "":
		updatePayload["status"] = models.DistributionResultStatusScheduled
		updatePayload["external_job_id"] = resp.JobID
	default:
		platformResult, ok := resp.ResultByName[string(result.Platform)]
		if !ok {
			updatePayload["status"] = models.DistributionResultStatusProcessing
		} else if platformResult.Success {
			now := time.Now()
			updatePayload["status"] = models.DistributionResultStatusSuccess
			updatePayload["published_url"] = stringPtr(platformResult.URL)
			updatePayload["completed_at"] = &now
		} else {
			failErr := fmt.Errorf(firstNonEmpty(platformResult.Error, "平台分发失败"))
			return s.markResultFailed(result, failErr, false, false)
		}
	}

	if err := s.updateDistributionResult(result.ID, updatePayload); err != nil {
		return err
	}

	return s.refreshJobStatus(job.ID)
}

func (s *DistributionService) dispatchDiscordResult(ctx context.Context, job *models.DistributionJob, result *models.DistributionResult) error {
	if result.TargetID == nil {
		return s.markResultFailed(result, fmt.Errorf("Discord target 缺失"), false, true)
	}

	target, err := s.getDistributionTarget(result.DeviceID, *result.TargetID)
	if err != nil {
		return s.markResultFailed(result, fmt.Errorf("Discord target 不存在"), false, true)
	}
	if target.SecretEncrypted == nil || *target.SecretEncrypted == "" {
		return s.markResultFailed(result, fmt.Errorf("Discord webhook 缺失"), false, true)
	}

	webhookURL, err := decryptDistributionSecret(*target.SecretEncrypted)
	if err != nil {
		return s.markResultFailed(result, fmt.Errorf("Discord webhook 解密失败"), false, true)
	}

	var payload distributionExecutionPayload
	if err := json.Unmarshal(result.RequestSnapshot, &payload); err != nil {
		return s.markResultFailed(result, fmt.Errorf("解析分发请求失败: %w", err), false, false)
	}

	resp, err := s.discord.Send(ctx, DiscordSendRequest{
		WebhookURL:  webhookURL,
		Title:       payload.Title,
		Body:        payload.Body,
		MediaURL:    payload.MediaURL,
		ContentType: string(job.ContentType),
	})
	if err != nil {
		return s.markResultFailed(result, err, true, needsRebind(err))
	}

	now := time.Now()
	publishedURL := buildDiscordMessageURL(resp.GuildID, resp.ChannelID, resp.ID)
	if err := s.updateDistributionResult(result.ID, map[string]interface{}{
		"status":              models.DistributionResultStatusSuccess,
		"response_snapshot":   mustJSON(resp.Raw),
		"external_message_id": stringPtr(resp.ID),
		"published_url":       stringPtr(publishedURL),
		"completed_at":        &now,
		"error_msg":           nil,
	}); err != nil {
		return err
	}

	if err := s.updateDistributionTargetStatus(target.ID, models.DistributionTargetStatusActive); err != nil {
		s.log.Warnw("Update discord target status failed", "error", err, "target_id", target.ID)
	}

	return s.refreshJobStatus(job.ID)
}

func (s *DistributionService) refreshUploadPostResult(ctx context.Context, job *models.DistributionJob, result *models.DistributionResult) error {
	statusResp, err := s.uploadPost.GetUploadStatus(ctx, derefString(result.ExternalRequestID), derefString(result.ExternalJobID))
	if err != nil {
		if needsRebind(err) {
			_ = s.markTargetNeedsRebind(result)
		}
		return s.rescheduleRefresh(result, err)
	}

	updates := map[string]interface{}{
		"response_snapshot": mustJSON(statusResp),
	}

	switch statusResp.Status {
	case "pending":
		if hasValue(result.ExternalJobID) {
			updates["status"] = models.DistributionResultStatusScheduled
		} else {
			updates["status"] = models.DistributionResultStatusProcessing
		}
		if err := s.updateDistributionResult(result.ID, updates); err != nil {
			return err
		}
		return s.refreshJobStatus(job.ID)
	case "in_progress":
		updates["status"] = models.DistributionResultStatusProcessing
		if err := s.updateDistributionResult(result.ID, updates); err != nil {
			return err
		}
		return s.refreshJobStatus(job.ID)
	case "completed":
		statusItem := matchUploadPostStatusItem(statusResp.Results, result.Platform)
		historyItem, _ := s.lookupUploadPostHistory(ctx, result)
		if statusItem != nil && statusItem.Success {
			now := time.Now()
			if historyItem != nil && historyItem.PostURL != "" {
				updates["published_url"] = stringPtr(historyItem.PostURL)
			}
			if historyItem != nil && historyItem.PlatformPostID != "" {
				updates["external_message_id"] = stringPtr(historyItem.PlatformPostID)
			}
			updates["status"] = models.DistributionResultStatusSuccess
			updates["completed_at"] = &now
			updates["error_msg"] = nil
			if err := s.updateDistributionResult(result.ID, updates); err != nil {
				return err
			}
			return s.refreshJobStatus(job.ID)
		}

		if historyItem != nil && historyItem.Success {
			now := time.Now()
			updates["status"] = models.DistributionResultStatusSuccess
			updates["published_url"] = stringPtr(historyItem.PostURL)
			updates["external_message_id"] = stringPtr(historyItem.PlatformPostID)
			updates["completed_at"] = &now
			updates["error_msg"] = nil
			if err := s.updateDistributionResult(result.ID, updates); err != nil {
				return err
			}
			return s.refreshJobStatus(job.ID)
		}

		message := "平台分发失败"
		if statusItem != nil && statusItem.Message != "" {
			message = statusItem.Message
		}
		if historyItem != nil && historyItem.ErrorMessage != "" {
			message = historyItem.ErrorMessage
		}
		return s.markCompletedFailure(job.ID, result, message)
	default:
		return nil
	}
}

func (s *DistributionService) lookupUploadPostHistory(ctx context.Context, result *models.DistributionResult) (*UploadPostHistoryItem, error) {
	lookback := s.historyLookbackPage
	if lookback <= 0 {
		lookback = 3
	}

	for page := 1; page <= lookback; page++ {
		history, err := s.uploadPost.GetUploadHistory(ctx, page, 50)
		if err != nil {
			return nil, err
		}
		for _, item := range history.History {
			if item.Platform != string(result.Platform) {
				continue
			}
			if hasValue(result.ExternalRequestID) && item.RequestID == derefString(result.ExternalRequestID) {
				itemCopy := item
				return &itemCopy, nil
			}
			if hasValue(result.ExternalJobID) && item.JobID == derefString(result.ExternalJobID) {
				itemCopy := item
				return &itemCopy, nil
			}
		}
	}

	return nil, nil
}

func (s *DistributionService) refreshJobStatus(jobID uint) error {
	var results []models.DistributionResult
	if err := s.db.Where("job_id = ?", jobID).Find(&results).Error; err != nil {
		return err
	}
	if len(results) == 0 {
		return nil
	}

	var successCount, failedCount, processingCount, scheduledCount, pendingCount int
	var firstError string
	for _, item := range results {
		switch item.Status {
		case models.DistributionResultStatusSuccess:
			successCount++
		case models.DistributionResultStatusFailed:
			failedCount++
			if firstError == "" && item.ErrorMsg != nil {
				firstError = *item.ErrorMsg
			}
		case models.DistributionResultStatusProcessing:
			processingCount++
		case models.DistributionResultStatusScheduled:
			scheduledCount++
		default:
			pendingCount++
		}
	}

	total := len(results)
	jobStatus := models.DistributionJobStatusPending
	var completedAt interface{}
	switch {
	case successCount == total:
		jobStatus = models.DistributionJobStatusCompleted
		now := time.Now()
		completedAt = &now
	case failedCount == total:
		jobStatus = models.DistributionJobStatusFailed
		now := time.Now()
		completedAt = &now
	case successCount+failedCount == total && successCount > 0 && failedCount > 0:
		jobStatus = models.DistributionJobStatusPartiallyFailed
		now := time.Now()
		completedAt = &now
	case processingCount > 0:
		jobStatus = models.DistributionJobStatusProcessing
	case scheduledCount > 0 && pendingCount == 0:
		jobStatus = models.DistributionJobStatusScheduled
	case pendingCount > 0 || scheduledCount > 0:
		jobStatus = models.DistributionJobStatusPending
	}

	updates := map[string]interface{}{
		"status":       jobStatus,
		"completed_at": completedAt,
	}
	if firstError != "" {
		updates["error_msg"] = firstError
	}
	if completedAt == nil {
		updates["completed_at"] = nil
	}

	return s.db.Model(&models.DistributionJob{}).Where("id = ?", jobID).Updates(updates).Error
}

func (s *DistributionService) markResultFailed(result *models.DistributionResult, err error, allowRetry bool, rebind bool) error {
	if rebind {
		_ = s.markTargetNeedsRebind(result)
	}

	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "平台分发失败"
	}

	if allowRetry && result.AttemptCount < maxDistributionAttempts && isRetriableDistributionError(err) {
		nextRetry := time.Now().Add(retryDelay(result.AttemptCount))
		return s.updateDistributionResult(result.ID, map[string]interface{}{
			"status":        models.DistributionResultStatusPending,
			"error_msg":     stringPtr(message),
			"next_retry_at": &nextRetry,
		})
	}

	now := time.Now()
	if err := s.updateDistributionResult(result.ID, map[string]interface{}{
		"status":        models.DistributionResultStatusFailed,
		"error_msg":     stringPtr(message),
		"completed_at":  &now,
		"next_retry_at": nil,
	}); err != nil {
		return err
	}

	return s.refreshJobStatus(result.JobID)
}

func (s *DistributionService) markCompletedFailure(jobID uint, result *models.DistributionResult, message string) error {
	now := time.Now()
	if err := s.updateDistributionResult(result.ID, map[string]interface{}{
		"status":        models.DistributionResultStatusFailed,
		"error_msg":     stringPtr(firstNonEmpty(message, "平台分发失败")),
		"completed_at":  &now,
		"next_retry_at": nil,
	}); err != nil {
		return err
	}
	return s.refreshJobStatus(jobID)
}

func (s *DistributionService) rescheduleRefresh(result *models.DistributionResult, err error) error {
	nextRetry := time.Now().Add(30 * time.Second)
	return s.updateDistributionResult(result.ID, map[string]interface{}{
		"next_retry_at": &nextRetry,
		"error_msg":     stringPtr(err.Error()),
	})
}

func (s *DistributionService) updateDistributionResult(resultID uint, updates map[string]interface{}) error {
	return s.db.Model(&models.DistributionResult{}).Where("id = ?", resultID).Updates(updates).Error
}

func (s *DistributionService) updateDistributionTargetStatus(targetID uint, status models.DistributionTargetStatus) error {
	return s.db.Model(&models.DistributionTarget{}).Where("id = ?", targetID).Update("status", status).Error
}

func (s *DistributionService) markTargetNeedsRebind(result *models.DistributionResult) error {
	if result.TargetID != nil {
		return s.updateDistributionTargetStatus(*result.TargetID, models.DistributionTargetStatusNeedsRebind)
	}
	if isUploadPostPlatform(result.Platform) {
		_ = s.db.Model(&models.UploadPostProfile{}).
			Where("device_id = ?", result.DeviceID).
			Update("status", models.UploadPostProfileStatusError).Error
		return s.db.Model(&models.DistributionTarget{}).
			Where("device_id = ? AND platform = ?", result.DeviceID, result.Platform).
			Update("status", models.DistributionTargetStatusNeedsRebind).Error
	}
	return nil
}

func (s *DistributionService) getDefaultTarget(deviceID string, platform models.DistributionPlatform, targetType models.DistributionTargetType) (*models.DistributionTarget, error) {
	var target models.DistributionTarget
	if err := s.db.
		Where("device_id = ? AND platform = ? AND target_type = ? AND is_default = ?", deviceID, platform, targetType, true).
		First(&target).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *DistributionService) findTargetByIdentifier(deviceID string, platform models.DistributionPlatform, targetType models.DistributionTargetType, identifier string) (*models.DistributionTarget, error) {
	var target models.DistributionTarget
	if err := s.db.
		Where("device_id = ? AND platform = ? AND target_type = ? AND identifier = ?", deviceID, platform, targetType, identifier).
		First(&target).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *DistributionService) localPathToStaticURL(localPath string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(s.cfg.Storage.BaseURL), "/")
	if baseURL == "" {
		return ""
	}
	storagePath := strings.TrimRight(strings.TrimSpace(s.cfg.Storage.LocalPath), string(filepath.Separator))
	if storagePath == "" {
		return ""
	}
	relativePath := strings.TrimPrefix(strings.TrimPrefix(localPath, storagePath), string(filepath.Separator))
	if relativePath == "" {
		return ""
	}
	return baseURL + "/" + filepath.ToSlash(relativePath)
}

func normalizeDistributionPlatformInputs(raw []string) ([]models.DistributionPlatform, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("selectedPlatforms 不能为空")
	}

	seen := make(map[models.DistributionPlatform]struct{})
	result := make([]models.DistributionPlatform, 0, len(raw))
	for _, item := range raw {
		var platform models.DistributionPlatform
		switch strings.ToLower(strings.TrimSpace(item)) {
		case string(models.DistributionPlatformPinterest):
			platform = models.DistributionPlatformPinterest
		case string(models.DistributionPlatformReddit):
			platform = models.DistributionPlatformReddit
		case string(models.DistributionPlatformDiscord):
			platform = models.DistributionPlatformDiscord
		default:
			return nil, fmt.Errorf("不支持的平台: %s", item)
		}
		if _, exists := seen[platform]; exists {
			continue
		}
		seen[platform] = struct{}{}
		result = append(result, platform)
	}

	return result, nil
}

func isUploadPostPlatform(platform models.DistributionPlatform) bool {
	return platform == models.DistributionPlatformPinterest || platform == models.DistributionPlatformReddit
}

func needsRebind(err error) bool {
	var adapterErr *DistributionAdapterError
	return errors.As(err, &adapterErr) && adapterErr.NeedsRebind
}

func isRetriableDistributionError(err error) bool {
	var adapterErr *DistributionAdapterError
	if errors.As(err, &adapterErr) {
		return adapterErr.Retriable
	}
	return false
}

func matchUploadPostStatusItem(items []UploadPostStatusItem, platform models.DistributionPlatform) *UploadPostStatusItem {
	for _, item := range items {
		if item.Platform == string(platform) {
			itemCopy := item
			return &itemCopy
		}
	}
	if len(items) == 1 {
		itemCopy := items[0]
		return &itemCopy
	}
	return nil
}

func buildDiscordMessageURL(guildID, channelID, messageID string) string {
	if guildID == "" || channelID == "" || messageID == "" {
		return ""
	}
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)
}

func hasValue(value *string) bool {
	return value != nil && strings.TrimSpace(*value) != ""
}

func jsonMap(raw datatypes.JSON) map[string]interface{} {
	if len(raw) == 0 {
		return map[string]interface{}{}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil || result == nil {
		return map[string]interface{}{}
	}

	return result
}
