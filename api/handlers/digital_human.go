package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DigitalHumanHandler struct {
	service             *services.DigitalHumanService
	uploadService       *services.UploadService
	speechTTSService    *services.SpeechTTSService
	voiceLibraryService *services.VoiceLibraryService
	log                 *logger.Logger
}

func NewDigitalHumanHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) (*DigitalHumanHandler, error) {
	uploadService, err := services.NewUploadService(cfg, log)
	if err != nil {
		return nil, err
	}

	var voiceLibraryService *services.VoiceLibraryService
	voiceLibraryService, err = services.NewVoiceLibraryService(cfg, db, log)
	if err != nil {
		log.Warnw("Failed to initialize voice library service for digital human", "error", err)
	}

	return &DigitalHumanHandler{
		service:             services.NewDigitalHumanService(cfg, log),
		uploadService:       uploadService,
		speechTTSService:    services.NewSpeechTTSService(cfg, log),
		voiceLibraryService: voiceLibraryService,
		log:                 log,
	}, nil
}

func (h *DigitalHumanHandler) Generate(c *gin.Context) {
	imageFile, imageHeader, err := c.Request.FormFile("image")
	if err != nil {
		response.BadRequest(c, "请上传图片")
		return
	}
	defer imageFile.Close()

	// `audio` is optional: users can either upload audio or select a voice.
	audioFile, audioHeader, audioErr := c.Request.FormFile("audio")
	if audioErr == nil {
		defer audioFile.Close()
	}

	speechText := strings.TrimSpace(c.PostForm("speech_text"))
	motionText := strings.TrimSpace(c.PostForm("motion_text"))
	voiceID := strings.TrimSpace(c.PostForm("voice_id"))
	voiceType := strings.TrimSpace(c.PostForm("voice_type"))
	audioURL := strings.TrimSpace(c.PostForm("audio_url"))
	ttsResourceID := ""
	if speechText == "" {
		response.BadRequest(c, "请填写说话内容")
		return
	}

	imageContentType := imageHeader.Header.Get("Content-Type")
	if imageContentType == "" {
		imageContentType = "application/octet-stream"
	}

	allowedImageTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedImageTypes[imageContentType] {
		response.BadRequest(c, "只支持图片格式 (jpg, png, gif, webp)")
		return
	}
	if imageHeader.Size > 10*1024*1024 {
		response.BadRequest(c, "图片大小不能超过10MB")
		return
	}

	// Read the image once so we can:
	// 1) store it locally for later access/debug
	// 2) send a data URL to Volcengine to avoid requiring a public reachable URL
	//    (tunnels like localtunnel can be unreachable from some upstream networks)
	imageBytes, err := io.ReadAll(io.LimitReader(imageFile, 10*1024*1024+1))
	if err != nil {
		h.log.Errorw("Failed to read image", "error", err)
		response.InternalError(c, "读取图片失败")
		return
	}
	if len(imageBytes) > 10*1024*1024 {
		response.BadRequest(c, "图片大小不能超过10MB")
		return
	}
	imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)

	var audioContentType string
	if audioErr == nil {
		audioContentType = audioHeader.Header.Get("Content-Type")
		if audioContentType == "" {
			audioContentType = "application/octet-stream"
		}
		if !isAllowedAudioUpload(audioContentType, audioHeader.Filename) {
			response.BadRequest(c, "只支持常见音频格式 (mp3, wav, m4a, ogg, aac)")
			return
		}
		if audioHeader.Size > 20*1024*1024 {
			response.BadRequest(c, "音频大小不能超过20MB")
			return
		}
	} else if audioURL != "" {
		if !strings.HasPrefix(audioURL, "http://") && !strings.HasPrefix(audioURL, "https://") {
			response.BadRequest(c, "audio_url 必须是 http/https URL")
			return
		}
	}

	if audioErr != nil && audioURL == "" && voiceID != "" && h.voiceLibraryService != nil {
		customVoice, isCustom, resolveErr := h.voiceLibraryService.ResolveCustomVoiceByPublicID(c.Request.Context(), voiceID)
		if resolveErr != nil {
			h.log.Errorw("Failed to resolve custom voice", "voice_id", voiceID, "error", resolveErr)
			if isCustom {
				response.BadRequest(c, "自定义音色不存在或不可用，请重新上传后重试")
				return
			}
		} else if isCustom && customVoice != nil {
			if customVoice.Status != models.CustomVoiceStatusCompleted {
				response.BadRequest(c, "该自定义音色仍在训练中，请稍后重试")
				return
			}
			voiceType = strings.TrimSpace(customVoice.VoiceType)
			if voiceType == "" {
				voiceType = strings.TrimSpace(customVoice.SpeakerID)
			}
			ttsResourceID = strings.TrimSpace(customVoice.ResourceID)
		}
	}

	if audioErr != nil && audioURL == "" && voiceType == "" {
		response.BadRequest(c, "请先选择音色或上传音频")
		return
	}

	if audioErr != nil && audioURL == "" && voiceType != "" && !h.speechTTSService.IsConfigured() {
		response.Error(c, 400, "DIGITAL_HUMAN_TTS_NOT_ENABLED", "当前未配置文本转语音能力，请先上传音频，或联系管理员配置后重试")
		return
	}

	uploadedImageURL, err := h.uploadService.UploadFile(bytes.NewReader(imageBytes), imageHeader.Filename, imageContentType, "digital-human/images")
	if err != nil {
		h.log.Errorw("Failed to upload image", "error", err)
		response.InternalError(c, "图片上传失败")
		return
	}

	if audioErr == nil {
		uploadedAudioURL, err := h.uploadService.UploadFile(audioFile, audioHeader.Filename, audioContentType, "digital-human/audios")
		if err != nil {
			h.log.Errorw("Failed to upload audio", "error", err)
			response.InternalError(c, "音频上传失败")
			return
		}
		audioURL = uploadedAudioURL
		// 上传音频时，优先音频驱动，避免和文本驱动参数冲突。
		speechText = ""
		voiceType = ""
	} else if audioURL != "" {
		// 外部 audio_url 同样优先音频驱动。
		speechText = ""
		voiceType = ""
	}

	if audioURL == "" && speechText != "" && voiceType != "" && h.speechTTSService.IsConfigured() {
		ttsAudioURL, ttsErr := h.speechTTSService.SynthesizeToURLWithResource(c.Request.Context(), speechText, voiceType, ttsResourceID)
		if ttsErr != nil {
			h.log.Errorw("Failed to synthesize speech from text", "error", ttsErr)
			ttsErrMsg := ttsErr.Error()
			switch {
			case strings.Contains(ttsErrMsg, "resource ID is mismatched with speaker related resource"):
				response.BadRequest(c, "当前音色不支持文本合成，请更换音色后重试")
				return
			case strings.Contains(ttsErrMsg, "requested resource not granted") ||
				strings.Contains(ttsErrMsg, "load grant: requested grant not found") ||
				strings.Contains(ttsErrMsg, "invalid auth token") ||
				strings.Contains(ttsErrMsg, "missing Authorization header"):
				response.Error(c, 400, "DIGITAL_HUMAN_TTS_NOT_ENABLED", "当前火山语音账号未开通TTS资源或配置无效，请上传音频后生成，或联系管理员开通后再试")
				return
			default:
				response.InternalError(c, fmt.Sprintf("文本转音频失败: %v", ttsErr))
				return
			}
		}

		audioURL = strings.TrimSpace(ttsAudioURL)
		// 优先使用“文本->音频”结果，避免依赖数字人接口的内置文本配音开通状态。
		speechText = ""
		voiceType = ""
	}

	result, err := h.service.Generate(c.Request.Context(), &services.DigitalHumanRequest{
		ImageURL:    uploadedImageURL,
		ImageBase64: imageBase64,
		AudioURL:    audioURL,
		VoiceType:   voiceType,
		SpeechText:  speechText,
		MotionText:  motionText,
	})
	if err != nil {
		h.log.Errorw("Failed to generate digital human video", "error", err)
		msg := err.Error()
		switch {
		case strings.Contains(msg, "audio_url or speech_text is required"):
			response.BadRequest(c, "请先选择音色或上传音频")
		case strings.Contains(msg, "voice_type is required when speech_text is provided"):
			response.BadRequest(c, "请先选择音色或上传音频")
		case speechText != "" && (strings.Contains(msg, "\"code\":50215") || strings.Contains(msg, "Input invalid for this service")):
			response.Error(c, 400, "DIGITAL_HUMAN_TTS_NOT_ENABLED", "当前模型仅支持音频驱动，暂不支持直接文本配音。请上传音频后重试")
		case strings.Contains(msg, "Invalid parameter: AppID") || strings.Contains(msg, "UnauthorizedRequest.AppID"):
			response.Error(c, 400, "DIGITAL_HUMAN_TTS_NOT_ENABLED", "当前火山账号未开通文本配音能力，请先上传音频，或联系管理员开通后再试")
		case strings.Contains(msg, "\"code\":50218") || strings.Contains(msg, "Resource exists risk"):
			// Volcengine content risk control.
			response.BadRequest(c, "内容安全审核未通过，请更换角色图片/文案/音色后重试")
		case strings.Contains(msg, "\"code\":50514") || strings.Contains(msg, "Pre Audio Risk Not Pass"):
			response.BadRequest(c, "音频内容安全审核未通过，请更换说话内容或音色后重试")
		case strings.Contains(msg, "\"code\":50430") || strings.Contains(msg, "API Concurrent Limit"):
			// Volcengine concurrency limit.
			response.Error(c, 429, "TOO_MANY_REQUESTS", "请求过于频繁，请稍后重试")
		default:
			response.InternalError(c, msg)
		}
		return
	}

	response.Success(c, result)
}

func isAllowedAudioUpload(contentType, filename string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if strings.HasPrefix(ct, "audio/") || ct == "video/mp4" {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}

	byExt := strings.ToLower(mime.TypeByExtension(ext))
	if strings.HasPrefix(byExt, "audio/") || byExt == "video/mp4" {
		return true
	}

	switch ext {
	case ".mp3", ".wav", ".m4a", ".ogg", ".aac", ".flac", ".mp4":
		return true
	default:
		return false
	}
}
