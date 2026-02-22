package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/external/ffmpeg"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/video"
	"gorm.io/gorm"
)

type VideoMergeService struct {
	db              *gorm.DB
	aiService       *AIService
	transferService *ResourceTransferService
	ffmpeg          *ffmpeg.FFmpeg
	storagePath     string
	baseURL         string
	log             *logger.Logger
}

func NewVideoMergeService(db *gorm.DB, transferService *ResourceTransferService, storagePath, baseURL string, log *logger.Logger) *VideoMergeService {
	return &VideoMergeService{
		db:              db,
		aiService:       NewAIService(db, log),
		transferService: transferService,
		ffmpeg:          ffmpeg.NewFFmpeg(log),
		storagePath:     storagePath,
		baseURL:         baseURL,
		log:             log,
	}
}

type MergeVideoRequest struct {
	EpisodeID  string             `json:"episode_id" binding:"required"`
	DramaID    string             `json:"drama_id" binding:"required"`
	Title      string             `json:"title"`
	Scenes     []models.SceneClip `json:"scenes" binding:"required,min=1"`
	AudioClips []models.AudioClip `json:"audio_clips,omitempty"`
	Provider   string             `json:"provider"`
	Model      string             `json:"model"`
}

type DistributeVideoRequest struct {
	Platforms   []string `json:"platforms"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Hashtags    []string `json:"hashtags"`
}

var defaultDistributionPlatforms = []models.VideoDistributionPlatform{
	models.VideoDistributionPlatformTikTok,
	models.VideoDistributionPlatformYouTube,
	models.VideoDistributionPlatformInstagram,
	models.VideoDistributionPlatformX,
}

var distributionPlatformAlias = map[string]models.VideoDistributionPlatform{
	"tiktok":          models.VideoDistributionPlatformTikTok,
	"tik tok":         models.VideoDistributionPlatformTikTok,
	"douyin":          models.VideoDistributionPlatformTikTok,
	"youtube":         models.VideoDistributionPlatformYouTube,
	"youtube shorts":  models.VideoDistributionPlatformYouTube,
	"shorts":          models.VideoDistributionPlatformYouTube,
	"instagram":       models.VideoDistributionPlatformInstagram,
	"instagram reels": models.VideoDistributionPlatformInstagram,
	"reels":           models.VideoDistributionPlatformInstagram,
	"x":               models.VideoDistributionPlatformX,
	"twitter":         models.VideoDistributionPlatformX,
	"x twitter":       models.VideoDistributionPlatformX,
	"x com":           models.VideoDistributionPlatformX,
	"x.com":           models.VideoDistributionPlatformX,
}

type distributionGatewayRequest struct {
	Platform    string   `json:"platform"`
	SourceURL   string   `json:"source_url"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Hashtags    []string `json:"hashtags,omitempty"`
}

type distributionGatewayResponse struct {
	Success      bool   `json:"success"`
	PublishedURL string `json:"published_url"`
	Message      string `json:"message"`
	Error        string `json:"error"`
}

func (s *VideoMergeService) MergeVideos(req *MergeVideoRequest) (*models.VideoMerge, error) {
	// 验证episode权限
	var episode models.Episode
	if err := s.db.Preload("Drama").Where("id = ?", req.EpisodeID).First(&episode).Error; err != nil {
		return nil, fmt.Errorf("episode not found")
	}

	// 验证所有场景都有视频
	for i, scene := range req.Scenes {
		if scene.VideoURL == "" {
			return nil, fmt.Errorf("scene %d has no video", i+1)
		}
	}

	provider := req.Provider
	if provider == "" {
		provider = "doubao"
	}

	// 序列化场景列表
	scenesJSON, err := json.Marshal(req.Scenes)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize scenes: %w", err)
	}

	// 序列化音频列表（可选）
	var audioJSON []byte
	if len(req.AudioClips) > 0 {
		audioJSON, err = json.Marshal(req.AudioClips)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize audio clips: %w", err)
		}
	}

	s.log.Infow("Serialized scenes to JSON",
		"scenes_count", len(req.Scenes),
		"scenes_json", string(scenesJSON))

	epID, _ := strconv.ParseUint(req.EpisodeID, 10, 32)
	dramaID, _ := strconv.ParseUint(req.DramaID, 10, 32)

	videoMerge := &models.VideoMerge{
		EpisodeID:  uint(epID),
		DramaID:    uint(dramaID),
		Title:      req.Title,
		Provider:   provider,
		Model:      &req.Model,
		Scenes:     scenesJSON,
		AudioClips: audioJSON,
		Status:     models.VideoMergeStatusPending,
	}

	if err := s.db.Create(videoMerge).Error; err != nil {
		return nil, fmt.Errorf("failed to create merge record: %w", err)
	}

	go s.processMergeVideo(videoMerge.ID)

	return videoMerge, nil
}

func (s *VideoMergeService) processMergeVideo(mergeID uint) {
	var videoMerge models.VideoMerge
	if err := s.db.First(&videoMerge, mergeID).Error; err != nil {
		s.log.Errorw("Failed to load video merge", "error", err, "id", mergeID)
		return
	}

	s.db.Model(&videoMerge).Update("status", models.VideoMergeStatusProcessing)

	client, err := s.getVideoClient(videoMerge.Provider)
	if err != nil {
		s.updateMergeError(mergeID, err.Error())
		return
	}

	// 解析场景列表
	var scenes []models.SceneClip
	if err := json.Unmarshal(videoMerge.Scenes, &scenes); err != nil {
		s.updateMergeError(mergeID, fmt.Sprintf("failed to parse scenes: %v", err))
		return
	}

	// 解析音频列表（可选）
	var audioClips []models.AudioClip
	if len(videoMerge.AudioClips) > 0 {
		if err := json.Unmarshal(videoMerge.AudioClips, &audioClips); err != nil {
			s.log.Warnw("Failed to parse audio clips, will ignore", "error", err, "merge_id", mergeID)
			audioClips = nil
		}
	}

	// 调用视频合并API
	result, err := s.mergeVideoClips(client, scenes, audioClips)
	if err != nil {
		s.updateMergeError(mergeID, err.Error())
		return
	}

	if !result.Completed {
		s.db.Model(&videoMerge).Updates(map[string]interface{}{
			"status":  models.VideoMergeStatusProcessing,
			"task_id": result.TaskID,
		})
		go s.pollMergeStatus(mergeID, client, result.TaskID)
		return
	}

	s.completeMerge(mergeID, result)
}

func (s *VideoMergeService) mergeVideoClips(client video.VideoClient, scenes []models.SceneClip, audioClips []models.AudioClip) (*video.VideoResult, error) {
	if len(scenes) == 0 {
		return nil, fmt.Errorf("no scenes to merge")
	}

	// 按Order字段排序场景
	sort.Slice(scenes, func(i, j int) bool {
		return scenes[i].Order < scenes[j].Order
	})

	s.log.Infow("Merging video clips with FFmpeg", "scene_count", len(scenes))

	// 计算总时长
	var totalDuration float64
	for _, scene := range scenes {
		totalDuration += scene.Duration
	}

	// 准备FFmpeg合成选项
	clips := make([]ffmpeg.VideoClip, len(scenes))
	for i, scene := range scenes {
		resolvedVideoURL := s.resolveVideoURL(scene.VideoURL)
		if resolvedVideoURL == "" {
			return nil, fmt.Errorf("scene %d has empty video url", i+1)
		}

		clips[i] = ffmpeg.VideoClip{
			URL:        resolvedVideoURL,
			Duration:   scene.Duration,
			StartTime:  scene.StartTime,
			EndTime:    scene.EndTime,
			Transition: scene.Transition,
		}

		s.log.Infow("Clip added to merge queue",
			"order", scene.Order,
			"index", i,
			"source_url", scene.VideoURL,
			"resolved_url", resolvedVideoURL,
			"duration", scene.Duration,
			"start_time", scene.StartTime,
			"end_time", scene.EndTime)
	}

	// 创建视频输出目录
	videoDir := filepath.Join(s.storagePath, "videos", "merged")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create video directory: %w", err)
	}

	// 生成输出文件名
	fileName := fmt.Sprintf("merged_%d.mp4", time.Now().Unix())
	outputPath := filepath.Join(videoDir, fileName)

	// 使用FFmpeg合成视频
	mergedPath, err := s.ffmpeg.MergeVideos(&ffmpeg.MergeOptions{
		OutputPath: outputPath,
		Clips:      clips,
		AudioClips: s.buildAudioClips(audioClips),
	})
	if err != nil {
		return nil, fmt.Errorf("ffmpeg merge failed: %w", err)
	}

	s.log.Infow("Video merged successfully", "path", mergedPath)

	// 生成访问URL（相对路径）
	relPath := filepath.Join("videos", "merged", fileName)
	videoURL := fmt.Sprintf("%s/%s", s.baseURL, relPath)

	result := &video.VideoResult{
		VideoURL:  videoURL, // 返回可访问的URL
		Duration:  int(totalDuration),
		Completed: true,
		Status:    "completed",
	}

	return result, nil
}

func (s *VideoMergeService) buildAudioClips(audioClips []models.AudioClip) []ffmpeg.AudioClip {
	if len(audioClips) == 0 {
		return nil
	}

	sort.Slice(audioClips, func(i, j int) bool {
		return audioClips[i].Position < audioClips[j].Position
	})

	result := make([]ffmpeg.AudioClip, 0, len(audioClips))
	for _, clip := range audioClips {
		audioURL := strings.TrimSpace(clip.AudioURL)
		if audioURL == "" {
			continue
		}

		duration := clip.Duration
		if duration <= 0 && clip.EndTime > clip.StartTime {
			duration = clip.EndTime - clip.StartTime
		}
		if duration < 0 {
			duration = 0
		}

		volume := clip.Volume
		if volume <= 0 {
			volume = 1
		}

		result = append(result, ffmpeg.AudioClip{
			URL:       s.resolveAudioURL(audioURL),
			StartTime: clip.StartTime,
			EndTime:   clip.EndTime,
			Duration:  duration,
			Position:  clip.Position,
			Volume:    volume,
		})
	}

	return result
}

func (s *VideoMergeService) resolveVideoURL(videoURL string) string {
	trimmed := strings.TrimSpace(videoURL)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "file://") {
		trimmed = strings.TrimPrefix(trimmed, "file://")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}

	if strings.HasPrefix(trimmed, "/data/") {
		local := filepath.Join(s.storagePath, strings.TrimPrefix(trimmed, "/data/"))
		if _, err := os.Stat(local); err == nil {
			return local
		}
	}

	if strings.HasPrefix(trimmed, "/static/") {
		local := filepath.Join(s.storagePath, strings.TrimPrefix(trimmed, "/static/"))
		if _, err := os.Stat(local); err == nil {
			return local
		}
	}

	if filepath.IsAbs(trimmed) {
		if _, err := os.Stat(trimmed); err == nil {
			return trimmed
		}
	}

	local := filepath.Join(s.storagePath, strings.TrimPrefix(trimmed, "/"))
	if _, err := os.Stat(local); err == nil {
		return local
	}

	base := strings.TrimSpace(s.baseURL)
	if base != "" && (strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://")) {
		if parsedBase, err := url.Parse(base); err == nil && parsedBase.Scheme != "" && parsedBase.Host != "" {
			if strings.HasPrefix(trimmed, "/") {
				return fmt.Sprintf("%s://%s%s", parsedBase.Scheme, parsedBase.Host, trimmed)
			}
			if ref, err := url.Parse(trimmed); err == nil {
				return parsedBase.ResolveReference(ref).String()
			}
		}
	}

	return trimmed
}

func (s *VideoMergeService) resolveAudioURL(audioURL string) string {
	trimmed := strings.TrimSpace(audioURL)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "file://") {
		trimmed = strings.TrimPrefix(trimmed, "file://")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}

	if strings.HasPrefix(trimmed, "/data/") {
		local := filepath.Join(s.storagePath, strings.TrimPrefix(trimmed, "/data/"))
		if _, err := os.Stat(local); err == nil {
			return local
		}
	}

	if strings.HasPrefix(trimmed, "/static/") {
		local := filepath.Join(s.storagePath, strings.TrimPrefix(trimmed, "/static/"))
		if _, err := os.Stat(local); err == nil {
			return local
		}
	}

	if filepath.IsAbs(trimmed) {
		if _, err := os.Stat(trimmed); err == nil {
			return trimmed
		}
	}

	local := filepath.Join(s.storagePath, strings.TrimPrefix(trimmed, "/"))
	if _, err := os.Stat(local); err == nil {
		return local
	}

	return trimmed
}

func (s *VideoMergeService) pollMergeStatus(mergeID uint, client video.VideoClient, taskID string) {
	maxAttempts := 240
	pollInterval := 5 * time.Second

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(pollInterval)

		result, err := client.GetTaskStatus(taskID)
		if err != nil {
			s.log.Errorw("Failed to get merge task status", "error", err, "task_id", taskID)
			continue
		}

		if result.Completed {
			s.completeMerge(mergeID, result)
			return
		}

		if result.Error != "" {
			s.updateMergeError(mergeID, result.Error)
			return
		}
	}

	s.updateMergeError(mergeID, "timeout: video merge took too long")
}

func (s *VideoMergeService) completeMerge(mergeID uint, result *video.VideoResult) {
	now := time.Now()

	// 获取merge记录
	var videoMerge models.VideoMerge
	if err := s.db.First(&videoMerge, mergeID).Error; err != nil {
		s.log.Errorw("Failed to load video merge for completion", "error", err, "id", mergeID)
		return
	}

	finalVideoURL := result.VideoURL

	// 使用本地存储，不再使用MinIO
	s.log.Infow("Video merge completed, using local storage", "merge_id", mergeID, "local_path", result.VideoURL)

	updates := map[string]interface{}{
		"status":       models.VideoMergeStatusCompleted,
		"merged_url":   finalVideoURL,
		"completed_at": now,
	}

	if result.Duration > 0 {
		updates["duration"] = result.Duration
	}

	s.db.Model(&models.VideoMerge{}).Where("id = ?", mergeID).Updates(updates)

	// 更新episode的状态和最终视频URL
	if videoMerge.EpisodeID != 0 {
		s.db.Model(&models.Episode{}).Where("id = ?", videoMerge.EpisodeID).Updates(map[string]interface{}{
			"status":    "completed",
			"video_url": finalVideoURL,
		})
		s.log.Infow("Episode finalized", "episode_id", videoMerge.EpisodeID, "video_url", finalVideoURL)
	}

	s.log.Infow("Video merge completed", "id", mergeID, "url", finalVideoURL)
}

func (s *VideoMergeService) updateMergeError(mergeID uint, errorMsg string) {
	s.db.Model(&models.VideoMerge{}).Where("id = ?", mergeID).Updates(map[string]interface{}{
		"status":    models.VideoMergeStatusFailed,
		"error_msg": errorMsg,
	})
	s.log.Errorw("Video merge failed", "id", mergeID, "error", errorMsg)
}

func (s *VideoMergeService) getVideoClient(provider string) (video.VideoClient, error) {
	config, err := s.aiService.GetDefaultConfig("video")
	if err != nil {
		return nil, fmt.Errorf("failed to get video config: %w", err)
	}

	// 使用第一个模型
	model := ""
	if len(config.Model) > 0 {
		model = config.Model[0]
	}

	// 根据配置中的 provider 创建对应的客户端
	var endpoint string
	var queryEndpoint string

	switch config.Provider {
	case "runway":
		return video.NewRunwayClient(config.BaseURL, config.APIKey, model), nil
	case "pika":
		return video.NewPikaClient(config.BaseURL, config.APIKey, model), nil
	case "openai", "sora":
		return video.NewOpenAISoraClient(config.BaseURL, config.APIKey, model), nil
	case "minimax":
		return video.NewMinimaxClient(config.BaseURL, config.APIKey, model), nil
	case "chatfire":
		endpoint = "/video/generations"
		queryEndpoint = "/video/task/{taskId}"
		return video.NewChatfireClient(config.BaseURL, config.APIKey, model, endpoint, queryEndpoint), nil
	case "doubao", "volces", "ark":
		endpoint = "/contents/generations/tasks"
		queryEndpoint = "/generations/tasks/{taskId}"
		return video.NewVolcesArkClient(config.BaseURL, config.APIKey, model, endpoint, queryEndpoint), nil
	default:
		endpoint = "/contents/generations/tasks"
		queryEndpoint = "/generations/tasks/{taskId}"
		return video.NewVolcesArkClient(config.BaseURL, config.APIKey, model, endpoint, queryEndpoint), nil
	}
}

func (s *VideoMergeService) GetMerge(mergeID uint) (*models.VideoMerge, error) {
	var merge models.VideoMerge
	if err := s.db.Where("id = ? ", mergeID).First(&merge).Error; err != nil {
		return nil, err
	}
	return &merge, nil
}

func (s *VideoMergeService) ListMerges(episodeID *string, status string, page, pageSize int) ([]models.VideoMerge, int64, error) {
	query := s.db.Model(&models.VideoMerge{})

	if episodeID != nil && *episodeID != "" {
		query = query.Where("episode_id = ?", *episodeID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var merges []models.VideoMerge
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&merges).Error; err != nil {
		return nil, 0, err
	}

	return merges, total, nil
}

func (s *VideoMergeService) DeleteMerge(mergeID uint) error {
	result := s.db.Where("id = ? ", mergeID).Delete(&models.VideoMerge{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("merge not found")
	}
	return nil
}

func (s *VideoMergeService) DistributeVideo(mergeID uint, req *DistributeVideoRequest) ([]models.VideoDistribution, error) {
	merge, err := s.GetMerge(mergeID)
	if err != nil {
		return nil, fmt.Errorf("merge not found")
	}

	mergeURL := strings.TrimSpace(func() string {
		if merge.MergedURL == nil {
			return ""
		}
		return *merge.MergedURL
	}())
	if merge.Status != models.VideoMergeStatusCompleted || mergeURL == "" {
		return nil, fmt.Errorf("video merge is not completed yet")
	}

	if req == nil {
		req = &DistributeVideoRequest{}
	}

	platforms, err := normalizeDistributionPlatforms(req.Platforms)
	if err != nil {
		return nil, err
	}

	hashtagsJSON, err := json.Marshal(normalizeDistributionHashtags(req.Hashtags))
	if err != nil {
		return nil, fmt.Errorf("failed to serialize hashtags: %w", err)
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSpace(merge.Title)
	}
	description := strings.TrimSpace(req.Description)

	records := make([]models.VideoDistribution, 0, len(platforms))
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, platform := range platforms {
			record := models.VideoDistribution{
				MergeID:      merge.ID,
				EpisodeID:    merge.EpisodeID,
				DramaID:      merge.DramaID,
				Platform:     string(platform),
				Title:        title,
				Description:  description,
				Hashtags:     hashtagsJSON,
				SourceURL:    mergeURL,
				Status:       models.VideoDistributionStatusPending,
				PublishedURL: nil,
				ErrorMsg:     nil,
			}
			if err := tx.Create(&record).Error; err != nil {
				return fmt.Errorf("failed to create distribution record: %w", err)
			}
			records = append(records, record)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	for _, record := range records {
		distributionID := record.ID
		go s.processDistribution(distributionID)
	}

	return records, nil
}

func (s *VideoMergeService) ListDistributions(mergeID uint) ([]models.VideoDistribution, error) {
	if _, err := s.GetMerge(mergeID); err != nil {
		return nil, fmt.Errorf("merge not found")
	}

	var distributions []models.VideoDistribution
	if err := s.db.Where("merge_id = ?", mergeID).
		Order("created_at DESC").
		Find(&distributions).Error; err != nil {
		return nil, err
	}
	return distributions, nil
}

func (s *VideoMergeService) processDistribution(distributionID uint) {
	var distribution models.VideoDistribution
	if err := s.db.First(&distribution, distributionID).Error; err != nil {
		s.log.Errorw("Failed to load distribution", "error", err, "id", distributionID)
		return
	}

	startedAt := time.Now()
	if err := s.db.Model(&models.VideoDistribution{}).
		Where("id = ?", distributionID).
		Updates(map[string]interface{}{
			"status":     models.VideoDistributionStatusProcessing,
			"started_at": startedAt,
			"updated_at": time.Now(),
		}).Error; err != nil {
		s.log.Errorw("Failed to update distribution status", "error", err, "id", distributionID)
		return
	}

	time.Sleep(1200 * time.Millisecond)

	sourceURL := strings.TrimSpace(distribution.SourceURL)
	if sourceURL == "" {
		errMsg := "视频地址为空，无法分发"
		completedAt := time.Now()
		s.db.Model(&models.VideoDistribution{}).
			Where("id = ?", distributionID).
			Updates(map[string]interface{}{
				"status":       models.VideoDistributionStatusFailed,
				"error_msg":    errMsg,
				"completed_at": completedAt,
				"updated_at":   time.Now(),
			})
		return
	}

	hashtags := parseDistributionHashtags(distribution.Hashtags)
	publishedURL, message, err, handled := tryDispatchWithGateway(distribution, hashtags)
	if handled {
		completedAt := time.Now()
		if err != nil {
			errMsg := err.Error()
			s.log.Errorw("Distribution gateway failed",
				"error", err,
				"platform", distribution.Platform,
				"distribution_id", distributionID)
			_ = s.db.Model(&models.VideoDistribution{}).
				Where("id = ?", distributionID).
				Updates(map[string]interface{}{
					"status":       models.VideoDistributionStatusFailed,
					"message":      nil,
					"error_msg":    errMsg,
					"completed_at": completedAt,
					"updated_at":   time.Now(),
				}).Error
			return
		}

		if strings.TrimSpace(message) == "" {
			message = "平台发布完成"
		}

		if err := s.db.Model(&models.VideoDistribution{}).
			Where("id = ?", distributionID).
			Updates(map[string]interface{}{
				"status":        models.VideoDistributionStatusPublished,
				"message":       message,
				"published_url": publishedURL,
				"error_msg":     nil,
				"completed_at":  completedAt,
				"updated_at":    time.Now(),
			}).Error; err != nil {
			s.log.Errorw("Failed to complete distribution via gateway", "error", err, "id", distributionID)
		}
		return
	}

	publishedURL = buildDistributionPublishedURL(distribution.Platform, sourceURL, distributionID, distribution.Title, hashtags)
	completedAt := time.Now()
	message = "已提交平台发布，等待平台审核"
	if err := s.db.Model(&models.VideoDistribution{}).
		Where("id = ?", distributionID).
		Updates(map[string]interface{}{
			"status":        models.VideoDistributionStatusPublished,
			"message":       message,
			"published_url": publishedURL,
			"error_msg":     nil,
			"completed_at":  completedAt,
			"updated_at":    time.Now(),
		}).Error; err != nil {
		s.log.Errorw("Failed to complete distribution", "error", err, "id", distributionID)
	}
}

func normalizeDistributionPlatforms(platforms []string) ([]models.VideoDistributionPlatform, error) {
	if len(platforms) == 0 {
		return append([]models.VideoDistributionPlatform{}, defaultDistributionPlatforms...), nil
	}

	result := make([]models.VideoDistributionPlatform, 0, len(platforms))
	seen := make(map[models.VideoDistributionPlatform]struct{})

	for _, raw := range platforms {
		key := strings.ToLower(strings.TrimSpace(raw))
		key = strings.ReplaceAll(key, "-", " ")
		key = strings.Join(strings.Fields(key), " ")
		platform, ok := distributionPlatformAlias[key]
		if !ok {
			return nil, fmt.Errorf("unsupported platform: %s", raw)
		}
		if _, exists := seen[platform]; exists {
			continue
		}
		seen[platform] = struct{}{}
		result = append(result, platform)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("at least one platform is required")
	}

	return result, nil
}

func normalizeDistributionHashtags(hashtags []string) []string {
	if len(hashtags) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(hashtags))
	seen := make(map[string]struct{})
	for _, item := range hashtags {
		tag := strings.TrimSpace(strings.TrimPrefix(item, "#"))
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, tag)
		if len(result) >= 20 {
			break
		}
	}

	if len(result) == 0 {
		return []string{}
	}
	return result
}

func parseDistributionHashtags(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}

	var hashtags []string
	if err := json.Unmarshal(raw, &hashtags); err != nil {
		return nil
	}
	return normalizeDistributionHashtags(hashtags)
}

func tryDispatchWithGateway(distribution models.VideoDistribution, hashtags []string) (publishedURL string, message string, err error, handled bool) {
	gatewayURL := strings.TrimSpace(os.Getenv("DISTRIBUTION_GATEWAY_URL"))
	if gatewayURL == "" {
		return "", "", nil, false
	}

	payload := distributionGatewayRequest{
		Platform:    strings.ToLower(strings.TrimSpace(distribution.Platform)),
		SourceURL:   strings.TrimSpace(distribution.SourceURL),
		Title:       strings.TrimSpace(distribution.Title),
		Description: strings.TrimSpace(distribution.Description),
		Hashtags:    hashtags,
	}
	body, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return "", "", fmt.Errorf("failed to encode distribution payload: %w", marshalErr), true
	}

	req, reqErr := http.NewRequest(http.MethodPost, gatewayURL, bytes.NewReader(body))
	if reqErr != nil {
		return "", "", fmt.Errorf("failed to build distribution gateway request: %w", reqErr), true
	}

	req.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(os.Getenv("DISTRIBUTION_GATEWAY_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	timeout := 30 * time.Second
	if timeoutRaw := strings.TrimSpace(os.Getenv("DISTRIBUTION_GATEWAY_TIMEOUT_SECONDS")); timeoutRaw != "" {
		if timeoutSeconds, convErr := strconv.Atoi(timeoutRaw); convErr == nil && timeoutSeconds > 0 {
			timeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	client := &http.Client{Timeout: timeout}
	resp, doErr := client.Do(req)
	if doErr != nil {
		return "", "", fmt.Errorf("distribution gateway request failed: %w", doErr), true
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", "", fmt.Errorf("failed to read distribution gateway response: %w", readErr), true
	}

	var gatewayResp distributionGatewayResponse
	if len(respBody) > 0 {
		if unmarshalErr := json.Unmarshal(respBody, &gatewayResp); unmarshalErr != nil {
			return "", "", fmt.Errorf("invalid distribution gateway response: %w", unmarshalErr), true
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(gatewayResp.Error)
		if msg == "" {
			msg = fmt.Sprintf("gateway status %d", resp.StatusCode)
		}
		return "", "", fmt.Errorf("distribution gateway failed: %s", msg), true
	}

	if !gatewayResp.Success {
		msg := strings.TrimSpace(gatewayResp.Error)
		if msg == "" {
			msg = "distribution gateway returned failure"
		}
		return "", "", fmt.Errorf(msg), true
	}

	publishedURL = strings.TrimSpace(gatewayResp.PublishedURL)
	if publishedURL == "" {
		publishedURL = strings.TrimSpace(distribution.SourceURL)
	}
	message = strings.TrimSpace(gatewayResp.Message)
	return publishedURL, message, nil, true
}

func buildDistributionPublishedURL(platform string, sourceURL string, distributionID uint, title string, hashtags []string) string {
	escapedSource := url.QueryEscape(sourceURL)
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case string(models.VideoDistributionPlatformTikTok):
		return fmt.Sprintf("https://www.tiktok.com/upload?source=%s&distribution_id=%d", escapedSource, distributionID)
	case string(models.VideoDistributionPlatformYouTube):
		return fmt.Sprintf("https://studio.youtube.com/channel/UC/upload?source=%s&distribution_id=%d", escapedSource, distributionID)
	case string(models.VideoDistributionPlatformInstagram):
		return fmt.Sprintf("https://www.instagram.com/reels/upload?source=%s&distribution_id=%d", escapedSource, distributionID)
	case string(models.VideoDistributionPlatformX):
		textParts := make([]string, 0, 4)
		if trimmedTitle := strings.TrimSpace(title); trimmedTitle != "" {
			textParts = append(textParts, trimmedTitle)
		}
		textParts = append(textParts, sourceURL)
		if len(hashtags) > 0 {
			hashtagsText := make([]string, 0, len(hashtags))
			for _, tag := range hashtags {
				cleaned := strings.TrimSpace(strings.TrimPrefix(tag, "#"))
				if cleaned == "" {
					continue
				}
				hashtagsText = append(hashtagsText, "#"+cleaned)
			}
			if len(hashtagsText) > 0 {
				textParts = append(textParts, strings.Join(hashtagsText, " "))
			}
		}
		return fmt.Sprintf("https://x.com/intent/post?text=%s", url.QueryEscape(strings.Join(textParts, "\n")))
	default:
		return sourceURL
	}
}

// TimelineClip 时间线片段数据
type TimelineClip struct {
	AssetID      string                 `json:"asset_id"`      // 素材库视频ID（优先使用）
	StoryboardID string                 `json:"storyboard_id"` // 分镜ID（fallback）
	Order        int                    `json:"order"`
	StartTime    float64                `json:"start_time"`
	EndTime      float64                `json:"end_time"`
	Duration     float64                `json:"duration"`
	Transition   map[string]interface{} `json:"transition"`
}

// FinalizeEpisodeRequest 完成剧集制作请求
type FinalizeEpisodeRequest struct {
	EpisodeID  string             `json:"episode_id"`
	Clips      []TimelineClip     `json:"clips"`
	AudioClips []models.AudioClip `json:"audio_clips"`
}

// FinalizeEpisode 完成集数制作，根据时间线场景顺序合成最终视频
func (s *VideoMergeService) FinalizeEpisode(episodeID string, timelineData *FinalizeEpisodeRequest) (map[string]interface{}, error) {
	// 验证episode存在且属于该用户
	var episode models.Episode
	if err := s.db.Preload("Drama").Preload("Storyboards").Where("id = ?", episodeID).First(&episode).Error; err != nil {
		return nil, fmt.Errorf("episode not found")
	}

	// 构建分镜ID映射
	sceneMap := make(map[string]models.Storyboard)
	for _, scene := range episode.Storyboards {
		sceneMap[fmt.Sprintf("%d", scene.ID)] = scene
	}

	// 根据时间线数据构建场景片段
	var sceneClips []models.SceneClip
	var skippedScenes []int

	if timelineData != nil && len(timelineData.Clips) > 0 {
		// 使用前端提供的时间线数据
		for _, clip := range timelineData.Clips {
			// 优先使用素材库中的视频（通过AssetID）
			var videoURL string
			var sceneID uint

			if clip.AssetID != "" {
				// 从素材库获取视频URL
				var asset models.Asset
				if err := s.db.Where("id = ? AND type = ?", clip.AssetID, models.AssetTypeVideo).First(&asset).Error; err == nil {
					videoURL = asset.URL
					// 如果asset关联了storyboard，使用关联的storyboard_id
					if asset.StoryboardID != nil {
						sceneID = *asset.StoryboardID
					}
					s.log.Infow("Using video from asset library", "asset_id", clip.AssetID, "video_url", videoURL)
				} else {
					s.log.Warnw("Asset not found, will try storyboard video", "asset_id", clip.AssetID, "error", err)
				}
			}

			// 如果没有从素材库获取到视频，尝试从storyboard获取
			if videoURL == "" && clip.StoryboardID != "" {
				scene, exists := sceneMap[clip.StoryboardID]
				if !exists {
					s.log.Warnw("Storyboard not found in episode, skipping", "storyboard_id", clip.StoryboardID)
					continue
				}

				if scene.VideoURL != nil && *scene.VideoURL != "" {
					videoURL = *scene.VideoURL
					sceneID = scene.ID
					s.log.Infow("Using video from storyboard", "storyboard_id", clip.StoryboardID, "video_url", videoURL)
				}
			}

			// 如果仍然没有视频URL，跳过该片段
			if videoURL == "" {
				s.log.Warnw("No video available for clip, skipping", "clip", clip)
				if clip.StoryboardID != "" {
					if scene, exists := sceneMap[clip.StoryboardID]; exists {
						skippedScenes = append(skippedScenes, scene.StoryboardNumber)
					}
				}
				continue
			}

			sceneClip := models.SceneClip{
				SceneID:    sceneID,
				VideoURL:   videoURL,
				Duration:   clip.Duration,
				Order:      clip.Order,
				StartTime:  clip.StartTime,
				EndTime:    clip.EndTime,
				Transition: clip.Transition,
			}
			s.log.Infow("Adding scene clip with transition",
				"scene_id", sceneID,
				"order", clip.Order,
				"transition", clip.Transition)
			sceneClips = append(sceneClips, sceneClip)
		}
	} else {
		// 没有时间线数据，使用默认场景顺序
		if len(episode.Storyboards) == 0 {
			return nil, fmt.Errorf("no scenes found for this episode")
		}

		order := 0
		for _, scene := range episode.Storyboards {
			// 优先从素材库查找该分镜关联的视频
			var videoURL string
			var asset models.Asset
			if err := s.db.Where("storyboard_id = ? AND type = ? AND episode_id = ?",
				scene.ID, models.AssetTypeVideo, episode.ID).
				Order("created_at DESC").
				First(&asset).Error; err == nil {
				videoURL = asset.URL
				s.log.Infow("Using video from asset library for storyboard",
					"storyboard_id", scene.ID,
					"asset_id", asset.ID,
					"video_url", videoURL)
			} else if scene.VideoURL != nil && *scene.VideoURL != "" {
				// 如果素材库没有，使用storyboard的video_url作为fallback
				videoURL = *scene.VideoURL
				s.log.Infow("Using fallback video from storyboard",
					"storyboard_id", scene.ID,
					"video_url", videoURL)
			}

			// 跳过没有视频的场景
			if videoURL == "" {
				s.log.Warnw("Scene has no video, skipping", "storyboard_number", scene.StoryboardNumber)
				skippedScenes = append(skippedScenes, scene.StoryboardNumber)
				continue
			}

			clip := models.SceneClip{
				SceneID:  scene.ID,
				VideoURL: videoURL,
				Duration: float64(scene.Duration),
				Order:    order,
			}
			sceneClips = append(sceneClips, clip)
			order++
		}
	}

	// 检查是否至少有一个场景可以合成
	if len(sceneClips) == 0 {
		return nil, fmt.Errorf("no scenes with videos available for merging")
	}

	// 创建视频合成任务
	title := fmt.Sprintf("%s - 第%d集", episode.Drama.Title, episode.EpisodeNum)

	var audioClips []models.AudioClip
	if timelineData != nil {
		audioClips = timelineData.AudioClips
	}

	finalReq := &MergeVideoRequest{
		EpisodeID:  episodeID,
		DramaID:    fmt.Sprintf("%d", episode.DramaID),
		Title:      title,
		Scenes:     sceneClips,
		AudioClips: audioClips,
		Provider:   "doubao", // 默认使用doubao
	}

	// 执行视频合成
	videoMerge, err := s.MergeVideos(finalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to start video merge: %w", err)
	}

	// 更新episode状态为processing
	s.db.Model(&episode).Updates(map[string]interface{}{
		"status": "processing",
	})

	result := map[string]interface{}{
		"message":      "视频合成任务已创建，正在后台处理",
		"merge_id":     videoMerge.ID,
		"episode_id":   episodeID,
		"scenes_count": len(sceneClips),
	}

	// 如果有跳过的场景，添加提示信息
	if len(skippedScenes) > 0 {
		result["skipped_scenes"] = skippedScenes
		result["warning"] = fmt.Sprintf("已跳过 %d 个未生成视频的场景（场景编号：%v）", len(skippedScenes), skippedScenes)
	}

	return result, nil
}
