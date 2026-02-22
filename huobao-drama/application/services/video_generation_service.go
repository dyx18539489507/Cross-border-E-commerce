package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
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
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/video"
	"gorm.io/gorm"
)

type VideoGenerationService struct {
	db              *gorm.DB
	transferService *ResourceTransferService
	log             *logger.Logger
	localStorage    *storage.LocalStorage
	aiService       *AIService
	ffmpeg          *ffmpeg.FFmpeg
}

func NewVideoGenerationService(db *gorm.DB, transferService *ResourceTransferService, localStorage *storage.LocalStorage, aiService *AIService, log *logger.Logger) *VideoGenerationService {
	service := &VideoGenerationService{
		db:              db,
		localStorage:    localStorage,
		transferService: transferService,
		aiService:       aiService,
		log:             log,
		ffmpeg:          ffmpeg.NewFFmpeg(log),
	}

	go service.RecoverPendingTasks()
	go func() {
		service.RecoverCompletedVideoLocalLinks()
		service.NormalizeCompletedVideoPlaybackURLs()
	}()

	return service
}

func firstVideoDeviceID(deviceIDs []string) string {
	if len(deviceIDs) == 0 {
		return ""
	}
	return strings.TrimSpace(deviceIDs[0])
}

func (s *VideoGenerationService) applyVideoDeviceScope(query *gorm.DB, deviceID string) *gorm.DB {
	if deviceID == "" {
		return query
	}
	dramaSubQuery := s.db.Model(&models.Drama{}).Select("id").Where("device_id = ?", deviceID)
	return query.Where("drama_id IN (?)", dramaSubQuery)
}

type GenerateVideoRequest struct {
	StoryboardID *uint  `json:"storyboard_id"`
	DramaID      string `json:"drama_id" binding:"required"`
	ImageGenID   *uint  `json:"image_gen_id"`

	// 参考图模式：single, first_last, multiple, none
	ReferenceMode string `json:"reference_mode"`

	// 单图模式
	ImageURL string `json:"image_url"`

	// 首尾帧模式
	FirstFrameURL *string `json:"first_frame_url"`
	LastFrameURL  *string `json:"last_frame_url"`

	// 多图模式
	ReferenceImageURLs []string `json:"reference_image_urls"`

	Prompt       string  `json:"prompt" binding:"required,min=5,max=2000"`
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Duration     *int    `json:"duration"`
	FPS          *int    `json:"fps"`
	AspectRatio  *string `json:"aspect_ratio"`
	Style        *string `json:"style"`
	MotionLevel  *int    `json:"motion_level"`
	CameraMotion *string `json:"camera_motion"`
	Seed         *int64  `json:"seed"`
}

func (s *VideoGenerationService) GenerateVideo(request *GenerateVideoRequest, deviceIDs ...string) (*models.VideoGeneration, error) {
	deviceID := firstVideoDeviceID(deviceIDs)
	if request.StoryboardID != nil {
		var storyboard models.Storyboard
		storyboardQuery := s.db.Preload("Episode").Where("storyboards.id = ?", *request.StoryboardID)
		if deviceID != "" {
			storyboardQuery = storyboardQuery.Joins("JOIN episodes ON episodes.id = storyboards.episode_id").
				Joins("JOIN dramas ON dramas.id = episodes.drama_id").
				Where("dramas.device_id = ?", deviceID)
		}
		if err := storyboardQuery.First(&storyboard).Error; err != nil {
			return nil, fmt.Errorf("storyboard not found")
		}
		if fmt.Sprintf("%d", storyboard.Episode.DramaID) != request.DramaID {
			return nil, fmt.Errorf("storyboard does not belong to drama")
		}
	}

	if request.ImageGenID != nil {
		var imageGen models.ImageGeneration
		imageQuery := s.db.Model(&models.ImageGeneration{}).Where("id = ?", *request.ImageGenID)
		imageQuery = s.applyVideoDeviceScope(imageQuery, deviceID)
		if err := imageQuery.First(&imageGen).Error; err != nil {
			return nil, fmt.Errorf("image generation not found")
		}
		if request.ImageURL == "" {
			if imageGen.MinioURL != nil && *imageGen.MinioURL != "" {
				request.ImageURL = *imageGen.MinioURL
			} else if imageGen.ImageURL != nil && *imageGen.ImageURL != "" {
				request.ImageURL = *imageGen.ImageURL
			} else if imageGen.LocalPath != nil && *imageGen.LocalPath != "" {
				request.ImageURL = *imageGen.LocalPath
			}
		}
		if request.Prompt == "" && imageGen.Prompt != "" {
			request.Prompt = imageGen.Prompt
		}
	}

	dramaQuery := s.db.Where("id = ?", request.DramaID)
	if deviceID != "" {
		dramaQuery = dramaQuery.Where("device_id = ?", deviceID)
	}
	var drama models.Drama
	if err := dramaQuery.First(&drama).Error; err != nil {
		return nil, fmt.Errorf("drama not found")
	}
	if request.ImageURL != "" {
		request.ImageURL = s.normalizeMediaURL(request.ImageURL)
	}

	provider := strings.TrimSpace(request.Provider)
	if s.aiService != nil {
		if request.Model != "" {
			if cfg, err := s.aiService.GetConfigForModel("video", request.Model, drama.DeviceID); err == nil {
				if p := strings.TrimSpace(cfg.Provider); p != "" {
					provider = p
				}
			}
		}
		if provider == "" {
			if cfg, err := s.aiService.GetDefaultConfig("video", drama.DeviceID); err == nil {
				if p := strings.TrimSpace(cfg.Provider); p != "" {
					provider = p
				}
			}
		}
	}
	if provider == "" {
		provider = "doubao"
	}

	dramaID, _ := strconv.ParseUint(request.DramaID, 10, 32)

	videoGen := &models.VideoGeneration{
		StoryboardID: request.StoryboardID,
		DramaID:      uint(dramaID),
		ImageGenID:   request.ImageGenID,
		Provider:     provider,
		Prompt:       request.Prompt,
		Model:        request.Model,
		Duration:     request.Duration,
		FPS:          request.FPS,
		AspectRatio:  request.AspectRatio,
		Style:        request.Style,
		MotionLevel:  request.MotionLevel,
		CameraMotion: request.CameraMotion,
		Seed:         request.Seed,
		Status:       models.VideoStatusPending,
	}

	// 根据参考图模式处理不同的参数
	if request.ReferenceMode != "" {
		videoGen.ReferenceMode = &request.ReferenceMode
	}

	switch request.ReferenceMode {
	case "single":
		// 单图模式
		if request.ImageURL != "" {
			videoGen.ImageURL = &request.ImageURL
		}
	case "first_last":
		// 首尾帧模式
		if request.FirstFrameURL != nil {
			videoGen.FirstFrameURL = request.FirstFrameURL
		}
		if request.LastFrameURL != nil {
			videoGen.LastFrameURL = request.LastFrameURL
		}
	case "multiple":
		// 多图模式
		if len(request.ReferenceImageURLs) > 0 {
			referenceImagesJSON, err := json.Marshal(request.ReferenceImageURLs)
			if err == nil {
				referenceImagesStr := string(referenceImagesJSON)
				videoGen.ReferenceImageURLs = &referenceImagesStr
			}
		}
	case "none":
		// 无参考图，纯文本生成
	default:
		// 向后兼容：如果没有指定模式，根据提供的参数自动判断
		if request.ImageURL != "" {
			videoGen.ImageURL = &request.ImageURL
			mode := "single"
			videoGen.ReferenceMode = &mode
		} else if request.FirstFrameURL != nil || request.LastFrameURL != nil {
			videoGen.FirstFrameURL = request.FirstFrameURL
			videoGen.LastFrameURL = request.LastFrameURL
			mode := "first_last"
			videoGen.ReferenceMode = &mode
		} else if len(request.ReferenceImageURLs) > 0 {
			referenceImagesJSON, err := json.Marshal(request.ReferenceImageURLs)
			if err == nil {
				referenceImagesStr := string(referenceImagesJSON)
				videoGen.ReferenceImageURLs = &referenceImagesStr
				mode := "multiple"
				videoGen.ReferenceMode = &mode
			}
		}
	}

	if err := s.db.Create(videoGen).Error; err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	go s.ProcessVideoGeneration(videoGen.ID)

	return videoGen, nil
}

func (s *VideoGenerationService) ProcessVideoGeneration(videoGenID uint) {
	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, videoGenID).Error; err != nil {
		s.log.Errorw("Failed to load video generation", "error", err, "id", videoGenID)
		return
	}

	s.db.Model(&videoGen).Update("status", models.VideoStatusProcessing)

	client, err := s.getVideoClient(videoGen.Provider, videoGen.Model, videoGen.DramaID)
	if err != nil {
		s.log.Errorw("Failed to get video client", "error", err, "provider", videoGen.Provider, "model", videoGen.Model)
		s.updateVideoGenError(videoGenID, err.Error())
		return
	}

	s.log.Infow("Starting video generation", "id", videoGenID, "prompt", videoGen.Prompt, "provider", videoGen.Provider)

	var opts []video.VideoOption
	if videoGen.Model != "" {
		opts = append(opts, video.WithModel(videoGen.Model))
	}
	if videoGen.Duration != nil {
		opts = append(opts, video.WithDuration(*videoGen.Duration))
	}
	if videoGen.FPS != nil {
		opts = append(opts, video.WithFPS(*videoGen.FPS))
	}
	if videoGen.AspectRatio != nil {
		opts = append(opts, video.WithAspectRatio(*videoGen.AspectRatio))
	}
	if videoGen.Style != nil {
		opts = append(opts, video.WithStyle(*videoGen.Style))
	}
	if videoGen.MotionLevel != nil {
		opts = append(opts, video.WithMotionLevel(*videoGen.MotionLevel))
	}
	if videoGen.CameraMotion != nil {
		opts = append(opts, video.WithCameraMotion(*videoGen.CameraMotion))
	}
	if videoGen.Seed != nil {
		opts = append(opts, video.WithSeed(*videoGen.Seed))
	}

	// 根据参考图模式添加相应的选项
	if videoGen.ReferenceMode != nil {
		switch *videoGen.ReferenceMode {
		case "first_last":
			// 首尾帧模式
			if videoGen.FirstFrameURL != nil {
				opts = append(opts, video.WithFirstFrame(s.resolveMediaInputURLForProvider(*videoGen.FirstFrameURL, videoGen.Provider)))
			}
			if videoGen.LastFrameURL != nil {
				opts = append(opts, video.WithLastFrame(s.resolveMediaInputURLForProvider(*videoGen.LastFrameURL, videoGen.Provider)))
			}
		case "multiple":
			// 多图模式
			if videoGen.ReferenceImageURLs != nil {
				var imageURLs []string
				if err := json.Unmarshal([]byte(*videoGen.ReferenceImageURLs), &imageURLs); err == nil {
					resolved := make([]string, 0, len(imageURLs))
					for _, url := range imageURLs {
						resolved = append(resolved, s.resolveMediaInputURLForProvider(url, videoGen.Provider))
					}
					opts = append(opts, video.WithReferenceImages(resolved))
				}
			}
		}
	}

	// 构造imageURL参数（单图模式使用，其他模式传空字符串）
	imageURL := ""
	if videoGen.ImageURL != nil {
		imageURL = s.resolveMediaInputURLForProvider(*videoGen.ImageURL, videoGen.Provider)
	}

	result, err := client.GenerateVideo(imageURL, videoGen.Prompt, opts...)
	if err != nil {
		s.log.Errorw("Video generation API call failed", "error", err, "id", videoGenID)
		s.updateVideoGenError(videoGenID, err.Error())
		return
	}

	if result.TaskID != "" {
		s.db.Model(&videoGen).Updates(map[string]interface{}{
			"task_id": result.TaskID,
			"status":  models.VideoStatusProcessing,
		})
		go s.pollTaskStatus(videoGenID, result.TaskID, videoGen.Provider, videoGen.Model)
		return
	}

	if result.VideoURL != "" {
		s.completeVideoGeneration(videoGenID, result.VideoURL, &result.Duration, &result.Width, &result.Height, nil)
		return
	}

	s.updateVideoGenError(videoGenID, "no task ID or video URL returned")
}

func (s *VideoGenerationService) pollTaskStatus(videoGenID uint, taskID string, provider string, model string) {
	var videoGen models.VideoGeneration
	if err := s.db.Select("drama_id").Where("id = ?", videoGenID).First(&videoGen).Error; err != nil {
		s.log.Errorw("Failed to load video generation for polling", "error", err, "id", videoGenID)
		s.updateVideoGenError(videoGenID, "failed to load generation record")
		return
	}

	client, err := s.getVideoClient(provider, model, videoGen.DramaID)
	if err != nil {
		s.log.Errorw("Failed to get video client for polling", "error", err)
		s.updateVideoGenError(videoGenID, "failed to get video client")
		return
	}

	maxAttempts := 300
	interval := 10 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(interval)

		var videoGen models.VideoGeneration
		if err := s.db.First(&videoGen, videoGenID).Error; err != nil {
			s.log.Errorw("Failed to load video generation", "error", err, "id", videoGenID)
			return
		}

		if videoGen.Status != models.VideoStatusProcessing {
			s.log.Infow("Video generation status changed, stopping poll", "id", videoGenID, "status", videoGen.Status)
			return
		}

		result, err := client.GetTaskStatus(taskID)
		if err != nil {
			s.log.Errorw("Failed to get task status", "error", err, "task_id", taskID)
			continue
		}

		if result.Completed {
			if result.VideoURL != "" {
				s.completeVideoGeneration(videoGenID, result.VideoURL, &result.Duration, &result.Width, &result.Height, nil)
				return
			}
			s.updateVideoGenError(videoGenID, "task completed but no video URL")
			return
		}

		if result.Error != "" {
			s.updateVideoGenError(videoGenID, result.Error)
			return
		}

		s.log.Infow("Video generation in progress", "id", videoGenID, "attempt", attempt+1)
	}

	s.updateVideoGenError(videoGenID, "polling timeout")
}

func (s *VideoGenerationService) completeVideoGeneration(videoGenID uint, videoURL string, duration *int, width *int, height *int, firstFrameURL *string) {
	sourceVideoURL := strings.TrimSpace(videoURL)
	playbackVideoURL := sourceVideoURL
	var localVideoURL string
	var durationProbePath string

	// 下载远端视频到本地存储，并优先使用本地地址写库（避免上游签名URL过期）
	if s.localStorage != nil && sourceVideoURL != "" &&
		(strings.HasPrefix(sourceVideoURL, "http://") || strings.HasPrefix(sourceVideoURL, "https://")) {
		downloadedURL, err := s.localStorage.DownloadFromURL(sourceVideoURL, "videos")
		if err != nil {
			s.log.Warnw("Failed to download video to local storage",
				"error", err,
				"id", videoGenID,
				"original_url", sourceVideoURL)
		} else if downloadedURL != "" {
			localVideoURL = downloadedURL
			if stablePlaybackURL := s.toStablePlaybackURL(downloadedURL); stablePlaybackURL != "" {
				playbackVideoURL = stablePlaybackURL
			} else {
				playbackVideoURL = downloadedURL
			}
			durationProbePath = s.resolveLocalStorageFilePath(downloadedURL)
			s.log.Infow("Video downloaded to local storage",
				"id", videoGenID,
				"source_url", sourceVideoURL,
				"local_url", localVideoURL,
				"playback_url", playbackVideoURL)
		}
	}

	if durationProbePath == "" {
		durationProbePath = s.resolveLocalStorageFilePath(playbackVideoURL)
	}
	if durationProbePath == "" {
		durationProbePath = playbackVideoURL
	}

	// 优先使用本地文件探测真实时长
	if durationProbePath != "" && s.ffmpeg != nil {
		if probedDuration, err := s.ffmpeg.GetVideoDuration(durationProbePath); err == nil {
			// 转换为整数秒（向上取整）
			durationInt := int(probedDuration + 0.5)
			duration = &durationInt
			s.log.Infow("Probed video duration",
				"id", videoGenID,
				"duration_seconds", durationInt,
				"duration_float", probedDuration)
		} else {
			s.log.Warnw("Failed to probe video duration, using provided duration",
				"error", err,
				"id", videoGenID,
				"probe_path", durationProbePath)
		}
	}

	// 下载首帧图片到本地存储（仅用于缓存，不更新数据库）
	if firstFrameURL != nil && *firstFrameURL != "" && s.localStorage != nil {
		_, err := s.localStorage.DownloadFromURL(*firstFrameURL, "video_frames")
		if err != nil {
			s.log.Warnw("Failed to download first frame to local storage",
				"error", err,
				"id", videoGenID,
				"original_url", *firstFrameURL)
		} else {
			s.log.Infow("First frame downloaded to local storage for caching",
				"id", videoGenID,
				"original_url", *firstFrameURL)
		}
	}

	// 数据库优先保存可持久播放地址
	updates := map[string]interface{}{
		"status":    models.VideoStatusCompleted,
		"video_url": playbackVideoURL,
	}
	if localVideoURL != "" {
		updates["local_path"] = playbackVideoURL
	}
	if duration != nil {
		updates["duration"] = *duration
	}
	if width != nil {
		updates["width"] = *width
	}
	if height != nil {
		updates["height"] = *height
	}
	if firstFrameURL != nil {
		updates["first_frame_url"] = *firstFrameURL
	}

	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGenID).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to update video generation", "error", err, "id", videoGenID)
		return
	}

	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, videoGenID).Error; err == nil {
		if videoGen.StoryboardID != nil {
			// 更新 Storyboard 的 video_url 和 duration
			storyboardUpdates := map[string]interface{}{
				"video_url": playbackVideoURL,
			}
			if duration != nil {
				storyboardUpdates["duration"] = *duration
			}
			if err := s.db.Model(&models.Storyboard{}).Where("id = ?", *videoGen.StoryboardID).Updates(storyboardUpdates).Error; err != nil {
				s.log.Warnw("Failed to update storyboard", "storyboard_id", *videoGen.StoryboardID, "error", err)
			} else {
				s.log.Infow("Updated storyboard with video info", "storyboard_id", *videoGen.StoryboardID, "duration", duration)
			}
		}
	}

	s.log.Infow("Video generation completed",
		"id", videoGenID,
		"source_url", sourceVideoURL,
		"playback_url", playbackVideoURL,
		"duration", duration)
}

func (s *VideoGenerationService) resolveLocalStorageFilePath(raw string) string {
	if s.localStorage == nil {
		return ""
	}

	rel := s.localStaticRelativePath(raw)
	if rel == "" {
		return ""
	}

	localPath := filepath.Join(s.localStorage.BasePath(), filepath.FromSlash(rel))
	if info, err := os.Stat(localPath); err != nil || info.IsDir() {
		return ""
	}

	return localPath
}

func (s *VideoGenerationService) toStablePlaybackURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "/static/") {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "static/") {
		return "/" + trimmed
	}

	if s.localStorage == nil {
		return trimmed
	}

	baseURL := strings.TrimSuffix(s.localStorage.BaseURL(), "/")
	if baseURL == "" {
		return trimmed
	}

	baseParsed, err := url.Parse(baseURL)
	if err != nil {
		return trimmed
	}

	prefixPath := strings.TrimSuffix(baseParsed.Path, "/")
	if prefixPath == "" {
		prefixPath = "/static"
	}
	if !strings.HasPrefix(prefixPath, "/") {
		prefixPath = "/" + prefixPath
	}

	prefixURL := baseURL + "/"
	if strings.HasPrefix(trimmed, prefixURL) {
		rel := strings.TrimPrefix(trimmed, prefixURL)
		rel = strings.TrimPrefix(rel, "/")
		if rel != "" {
			return prefixPath + "/" + rel
		}
	}

	return trimmed
}

func (s *VideoGenerationService) localStaticRelativePath(raw string) string {
	if s.localStorage == nil {
		return ""
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	baseURL := strings.TrimSuffix(strings.TrimSpace(s.localStorage.BaseURL()), "/")
	if baseURL != "" && strings.HasPrefix(trimmed, baseURL) {
		rel := strings.TrimPrefix(trimmed, baseURL)
		rel = strings.TrimPrefix(rel, "/")
		if strings.HasPrefix(rel, "static/") {
			rel = strings.TrimPrefix(rel, "static/")
		}
		return strings.TrimPrefix(rel, "/")
	}

	switch {
	case strings.HasPrefix(trimmed, "/static/"):
		return strings.TrimPrefix(trimmed, "/static/")
	case strings.HasPrefix(trimmed, "static/"):
		return strings.TrimPrefix(trimmed, "static/")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return ""
	}
	if strings.HasPrefix(parsed.Path, "/static/") && s.isLoopbackHost(parsed.Hostname()) {
		return strings.TrimPrefix(parsed.Path, "/static/")
	}

	return ""
}

func (s *VideoGenerationService) isLoopbackHost(host string) bool {
	normalized := strings.ToLower(strings.TrimSpace(host))
	return normalized == "localhost" ||
		normalized == "127.0.0.1" ||
		normalized == "0.0.0.0" ||
		normalized == "host.docker.internal"
}

func (s *VideoGenerationService) updateVideoGenError(videoGenID uint, errorMsg string) {
	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGenID).Updates(map[string]interface{}{
		"status":    models.VideoStatusFailed,
		"error_msg": errorMsg,
	}).Error; err != nil {
		s.log.Errorw("Failed to update video generation error", "error", err, "id", videoGenID)
	}
}

func (s *VideoGenerationService) getVideoClient(provider string, modelName string, dramaID uint) (video.VideoClient, error) {
	// 根据模型名称获取AI配置
	var config *models.AIServiceConfig
	var err error
	deviceID := ""
	if dramaID != 0 {
		var drama models.Drama
		if e := s.db.Select("device_id").Where("id = ?", dramaID).First(&drama).Error; e == nil {
			deviceID = drama.DeviceID
		}
	}

	if modelName != "" {
		config, err = s.aiService.GetConfigForModel("video", modelName, deviceID)
		if err != nil {
			s.log.Warnw("Failed to get config for model, using default", "model", modelName, "error", err)
			config, err = s.aiService.GetDefaultConfig("video", deviceID)
			if err != nil {
				return nil, fmt.Errorf("no video AI config found: %w", err)
			}
		}
	} else {
		config, err = s.aiService.GetDefaultConfig("video", deviceID)
		if err != nil {
			return nil, fmt.Errorf("no video AI config found: %w", err)
		}
	}

	// 使用配置中的信息创建客户端
	baseURL := config.BaseURL
	apiKey := config.APIKey
	model := modelName
	if model == "" && len(config.Model) > 0 {
		model = config.Model[0]
	}

	// 根据配置中的 provider 创建对应的客户端
	switch config.Provider {
	case "chatfire":
		return video.NewChatfireClient(baseURL, apiKey, model, "/video/generations", "/video/task/{taskId}"), nil
	case "doubao", "volcengine", "volces":
		return video.NewVolcesArkClient(baseURL, apiKey, model, "/contents/generations/tasks", "/contents/generations/tasks/{taskId}"), nil
	case "openai":
		// OpenAI Sora 使用 /v1/videos 端点
		return video.NewOpenAISoraClient(baseURL, apiKey, model), nil
	case "runway":
		return video.NewRunwayClient(baseURL, apiKey, model), nil
	case "pika":
		return video.NewPikaClient(baseURL, apiKey, model), nil
	case "minimax":
		return video.NewMinimaxClient(baseURL, apiKey, model), nil
	default:
		return nil, fmt.Errorf("unsupported video provider: %s (requested: %s)", config.Provider, provider)
	}
}

func (s *VideoGenerationService) normalizeMediaURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	if s.localStorage == nil {
		return trimmed
	}
	base := strings.TrimSuffix(s.localStorage.BaseURL(), "/")
	if base == "" {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "/static/") {
		if strings.HasSuffix(base, "/static") {
			return base + strings.TrimPrefix(trimmed, "/static")
		}
		return base + trimmed
	}
	if strings.HasPrefix(trimmed, "static/") {
		if strings.HasSuffix(base, "/static") {
			return base + strings.TrimPrefix(trimmed, "static")
		}
		return base + "/" + trimmed
	}
	if strings.HasPrefix(trimmed, "/") {
		return base + trimmed
	}
	return base + "/" + trimmed
}

func (s *VideoGenerationService) resolveMediaInputURLForProvider(raw string, provider string) string {
	normalized := s.normalizeMediaURL(raw)
	if normalized == "" || strings.HasPrefix(normalized, "data:") {
		return normalized
	}

	providerKey := strings.ToLower(strings.TrimSpace(provider))
	isLocal := s.isLocalMediaURL(normalized)

	// 本地静态资源优先转 data URL，避免依赖临时公网域名或第三方中转
	if isLocal && (providerKey == "openai" || providerKey == "doubao" || providerKey == "volces" || providerKey == "volcengine" || providerKey == "chatfire") {
		dataURL, err := s.tryLocalToDataURL(normalized)
		if err == nil && dataURL != "" {
			return dataURL
		}
	}

	// 对 doubao/volces/chatfire 保留公网 URL 中转兜底
	if isLocal && (providerKey == "doubao" || providerKey == "volces" || providerKey == "volcengine" || providerKey == "chatfire") {
		if publicURL, err := s.uploadPublicMedia(normalized); err == nil && publicURL != "" {
			return publicURL
		}
	}

	if providerKey == "openai" && (strings.HasPrefix(normalized, "http://") || strings.HasPrefix(normalized, "https://")) {
		remoteDataURL, fetchErr := s.fetchURLToDataURL(normalized)
		if fetchErr == nil && remoteDataURL != "" {
			return remoteDataURL
		}
	}
	return normalized
}

func (s *VideoGenerationService) tryLocalToDataURL(url string) (string, error) {
	if s.localStorage == nil {
		return "", fmt.Errorf("local storage not available")
	}

	rel := s.localStaticRelativePath(url)
	if rel == "" {
		return "", fmt.Errorf("not a local static url")
	}

	localPath := filepath.Join(s.localStorage.BasePath(), filepath.FromSlash(rel))
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", err
	}
	mimeType := mime.TypeByExtension(filepath.Ext(localPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

func (s *VideoGenerationService) fetchURLToDataURL(url string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch status %d", resp.StatusCode)
	}

	const maxSize = 10 << 20
	reader := io.LimitReader(resp.Body, maxSize+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	if len(data) > maxSize {
		return "", fmt.Errorf("image too large")
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(url))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded), nil
}

func (s *VideoGenerationService) isLocalMediaURL(raw string) bool {
	if s.localStaticRelativePath(raw) != "" {
		return true
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	return s.isLoopbackHost(u.Hostname())
}

func (s *VideoGenerationService) uploadPublicMedia(raw string) (string, error) {
	if s.localStorage == nil {
		return "", fmt.Errorf("local storage not available")
	}

	rel := s.localStaticRelativePath(raw)
	if rel == "" {
		return "", fmt.Errorf("invalid local url")
	}
	localPath := filepath.Join(s.localStorage.BasePath(), filepath.FromSlash(rel))
	info, err := os.Stat(localPath)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("local file not found")
	}
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	const maxSize = 10 << 20
	if info.Size() > maxSize {
		return "", fmt.Errorf("file too large")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("reqtype", "fileupload"); err != nil {
		return "", err
	}
	part, err := writer.CreateFormFile("fileToUpload", filepath.Base(localPath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://catbox.moe/user/api.php", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "Mozilla/5.0")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload status %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	publicURL := strings.TrimSpace(string(respBody))
	if publicURL == "" || !strings.HasPrefix(publicURL, "http") {
		return "", fmt.Errorf("upload invalid response")
	}
	return publicURL, nil
}

func (s *VideoGenerationService) RecoverPendingTasks() {
	var pendingVideos []models.VideoGeneration
	if err := s.db.Where("status = ? AND task_id != ''", models.VideoStatusProcessing).Find(&pendingVideos).Error; err != nil {
		s.log.Errorw("Failed to load pending video tasks", "error", err)
		return
	}

	s.log.Infow("Recovering pending video generation tasks", "count", len(pendingVideos))

	for _, videoGen := range pendingVideos {
		go s.pollTaskStatus(videoGen.ID, *videoGen.TaskID, videoGen.Provider, videoGen.Model)
	}
}

func (s *VideoGenerationService) RecoverCompletedVideoLocalLinks() {
	if s.localStorage == nil {
		return
	}

	var completedVideos []models.VideoGeneration
	if err := s.db.
		Where("status = ?", models.VideoStatusCompleted).
		Where("video_url IS NOT NULL AND video_url <> ''").
		Where("local_path IS NULL OR local_path = ''").
		Order("updated_at asc").
		Find(&completedVideos).Error; err != nil {
		s.log.Warnw("Failed to load completed videos for local link recovery", "error", err)
		return
	}

	if len(completedVideos) == 0 {
		return
	}

	type cachedVideoFile struct {
		name     string
		localURL string
		ts       time.Time
		used     bool
	}

	cacheFiles := make([]cachedVideoFile, 0)
	videosDir := filepath.Join(s.localStorage.BasePath(), "videos")
	entries, err := os.ReadDir(videosDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			timestamp, ok := parseVideoCacheTimestamp(entry.Name())
			if !ok {
				continue
			}
			cacheFiles = append(cacheFiles, cachedVideoFile{
				name:     entry.Name(),
				localURL: s.localStorage.GetURL(filepath.ToSlash(filepath.Join("videos", entry.Name()))),
				ts:       timestamp,
			})
		}
		sort.Slice(cacheFiles, func(i, j int) bool {
			return cacheFiles[i].ts.Before(cacheFiles[j].ts)
		})
	}

	const maxMatchDistance = 2 * time.Minute
	var recoveredCount int

	for _, videoGen := range completedVideos {
		if videoGen.VideoURL == nil {
			continue
		}

		sourceURL := strings.TrimSpace(*videoGen.VideoURL)
		if sourceURL == "" {
			continue
		}

		// 历史缓存兜底：按完成时间匹配本地缓存文件
		referenceTime := videoGen.UpdatedAt
		if videoGen.CompletedAt != nil {
			referenceTime = *videoGen.CompletedAt
		}

		bestIdx := -1
		bestDiff := maxMatchDistance + time.Second
		for idx := range cacheFiles {
			if cacheFiles[idx].used {
				continue
			}
			delta := cacheFiles[idx].ts.Sub(referenceTime)
			if delta < 0 {
				delta = -delta
			}
			if delta <= maxMatchDistance && delta < bestDiff {
				bestDiff = delta
				bestIdx = idx
			}
		}

		if bestIdx >= 0 {
			s.rebindVideoToLocalURL(videoGen, cacheFiles[bestIdx].localURL)
			cacheFiles[bestIdx].used = true
			recoveredCount++
			continue
		}

		// 缓存未匹配到时再尝试重下远程URL（URL仍有效时可恢复）
		if strings.HasPrefix(sourceURL, "http://") || strings.HasPrefix(sourceURL, "https://") {
			if downloadedURL, err := s.localStorage.DownloadFromURL(sourceURL, "videos"); err == nil && downloadedURL != "" {
				s.rebindVideoToLocalURL(videoGen, downloadedURL)
				recoveredCount++
			}
		}
	}

	if recoveredCount > 0 {
		s.log.Infow("Recovered completed videos with local playback URLs",
			"total_candidates", len(completedVideos),
			"recovered", recoveredCount)
	}
}

func (s *VideoGenerationService) NormalizeCompletedVideoPlaybackURLs() {
	if s.localStorage == nil {
		return
	}

	var completedVideos []models.VideoGeneration
	if err := s.db.
		Where("status = ?", models.VideoStatusCompleted).
		Where("(video_url IS NOT NULL AND video_url <> '') OR (local_path IS NOT NULL AND local_path <> '')").
		Find(&completedVideos).Error; err != nil {
		s.log.Warnw("Failed to load completed videos for playback url normalization", "error", err)
		return
	}

	if len(completedVideos) == 0 {
		return
	}

	var normalizedCount int
	for _, videoGen := range completedVideos {
		source := ""
		if videoGen.LocalPath != nil && strings.TrimSpace(*videoGen.LocalPath) != "" {
			source = *videoGen.LocalPath
		} else if videoGen.VideoURL != nil {
			source = *videoGen.VideoURL
		}

		stableURL := s.toStablePlaybackURL(source)
		if !strings.HasPrefix(stableURL, "/static/") {
			continue
		}

		needUpdate := videoGen.VideoURL == nil || strings.TrimSpace(*videoGen.VideoURL) != stableURL ||
			videoGen.LocalPath == nil || strings.TrimSpace(*videoGen.LocalPath) != stableURL
		if !needUpdate {
			continue
		}

		updates := map[string]interface{}{
			"video_url":  stableURL,
			"local_path": stableURL,
		}
		if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGen.ID).Updates(updates).Error; err != nil {
			s.log.Warnw("Failed to normalize completed video playback url",
				"id", videoGen.ID,
				"target_url", stableURL,
				"error", err)
			continue
		}

		if videoGen.StoryboardID != nil {
			if err := s.db.Model(&models.Storyboard{}).Where("id = ?", *videoGen.StoryboardID).Update("video_url", stableURL).Error; err != nil {
				s.log.Warnw("Failed to normalize storyboard video url",
					"storyboard_id", *videoGen.StoryboardID,
					"target_url", stableURL,
					"error", err)
			}
		}

		normalizedCount++
	}

	if normalizedCount > 0 {
		s.log.Infow("Normalized completed video playback URLs",
			"total_candidates", len(completedVideos),
			"normalized", normalizedCount)
	}
}

func (s *VideoGenerationService) rebindVideoToLocalURL(videoGen models.VideoGeneration, localURL string) {
	playbackURL := s.toStablePlaybackURL(localURL)
	if playbackURL == "" {
		playbackURL = localURL
	}

	updates := map[string]interface{}{
		"video_url":  playbackURL,
		"local_path": playbackURL,
	}
	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGen.ID).Updates(updates).Error; err != nil {
		s.log.Warnw("Failed to update video generation local link",
			"id", videoGen.ID,
			"local_url", localURL,
			"playback_url", playbackURL,
			"error", err)
		return
	}

	if videoGen.StoryboardID != nil {
		if err := s.db.Model(&models.Storyboard{}).Where("id = ?", *videoGen.StoryboardID).Update("video_url", playbackURL).Error; err != nil {
			s.log.Warnw("Failed to sync storyboard local video url",
				"storyboard_id", *videoGen.StoryboardID,
				"playback_url", playbackURL,
				"error", err)
		}
	}
}

func parseVideoCacheTimestamp(fileName string) (time.Time, bool) {
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	if len(base) < len("20060102_150405_000") {
		return time.Time{}, false
	}
	timestampPart := base[:len("20060102_150405_000")]
	parsed, err := time.ParseInLocation("20060102_150405_000", timestampPart, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func (s *VideoGenerationService) GetVideoGeneration(id uint, deviceIDs ...string) (*models.VideoGeneration, error) {
	deviceID := firstVideoDeviceID(deviceIDs)
	var videoGen models.VideoGeneration
	query := s.db.Model(&models.VideoGeneration{}).Where("id = ?", id)
	query = s.applyVideoDeviceScope(query, deviceID)
	if err := query.First(&videoGen).Error; err != nil {
		return nil, err
	}
	return &videoGen, nil
}

func (s *VideoGenerationService) ListVideoGenerations(dramaID *uint, storyboardID *uint, status string, limit int, offset int, deviceIDs ...string) ([]*models.VideoGeneration, int64, error) {
	deviceID := firstVideoDeviceID(deviceIDs)
	var videos []*models.VideoGeneration
	var total int64

	query := s.db.Model(&models.VideoGeneration{})
	query = s.applyVideoDeviceScope(query, deviceID)

	if dramaID != nil {
		query = query.Where("drama_id = ?", *dramaID)
	}
	if storyboardID != nil {
		query = query.Where("storyboard_id = ?", *storyboardID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

func (s *VideoGenerationService) GenerateVideoFromImage(imageGenID uint, deviceIDs ...string) (*models.VideoGeneration, error) {
	deviceID := firstVideoDeviceID(deviceIDs)
	var imageGen models.ImageGeneration
	query := s.db.Model(&models.ImageGeneration{}).Where("id = ?", imageGenID)
	query = s.applyVideoDeviceScope(query, deviceID)
	if err := query.First(&imageGen).Error; err != nil {
		return nil, fmt.Errorf("image generation not found")
	}

	if imageGen.Status != models.ImageStatusCompleted || imageGen.ImageURL == nil {
		return nil, fmt.Errorf("image is not ready")
	}

	// 获取关联的Storyboard以获取时长
	var duration *int
	if imageGen.StoryboardID != nil {
		var storyboard models.Storyboard
		if err := s.db.Where("id = ?", *imageGen.StoryboardID).First(&storyboard).Error; err == nil {
			duration = &storyboard.Duration
			s.log.Infow("Using storyboard duration for video generation",
				"storyboard_id", *imageGen.StoryboardID,
				"duration", storyboard.Duration)
		}
	}

	req := &GenerateVideoRequest{
		DramaID:      fmt.Sprintf("%d", imageGen.DramaID),
		StoryboardID: imageGen.StoryboardID,
		ImageGenID:   &imageGenID,
		ImageURL:     *imageGen.ImageURL,
		Prompt:       imageGen.Prompt,
		Duration:     duration,
	}

	return s.GenerateVideo(req, deviceID)
}

func (s *VideoGenerationService) BatchGenerateVideosForEpisode(episodeID string, deviceIDs ...string) ([]*models.VideoGeneration, error) {
	deviceID := firstVideoDeviceID(deviceIDs)
	var episode models.Episode
	episodeQuery := s.db.Preload("Storyboards").Where("episodes.id = ?", episodeID)
	if deviceID != "" {
		episodeQuery = episodeQuery.Joins("JOIN dramas ON dramas.id = episodes.drama_id").
			Where("dramas.device_id = ?", deviceID)
	}
	if err := episodeQuery.First(&episode).Error; err != nil {
		return nil, fmt.Errorf("episode not found")
	}

	var results []*models.VideoGeneration
	for _, storyboard := range episode.Storyboards {
		if storyboard.ImagePrompt == nil {
			continue
		}

		var imageGen models.ImageGeneration
		if err := s.db.Where("storyboard_id = ? AND status = ?", storyboard.ID, models.ImageStatusCompleted).
			Order("created_at DESC").First(&imageGen).Error; err != nil {
			s.log.Warnw("No completed image for storyboard", "storyboard_id", storyboard.ID)
			continue
		}

		videoGen, err := s.GenerateVideoFromImage(imageGen.ID, deviceID)
		if err != nil {
			s.log.Errorw("Failed to generate video", "storyboard_id", storyboard.ID, "error", err)
			continue
		}

		results = append(results, videoGen)
	}

	return results, nil
}

func (s *VideoGenerationService) DeleteVideoGeneration(id uint, deviceIDs ...string) error {
	deviceID := firstVideoDeviceID(deviceIDs)
	query := s.db.Model(&models.VideoGeneration{}).Where("id = ?", id)
	query = s.applyVideoDeviceScope(query, deviceID)
	return query.Delete(&models.VideoGeneration{}).Error
}
