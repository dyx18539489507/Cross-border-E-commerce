package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"path/filepath"
	"strings"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type DigitalHumanHandler struct {
	service          *services.DigitalHumanService
	uploadService    *services.UploadService
	speechTTSService *services.SpeechTTSService
	log              *logger.Logger
}

func NewDigitalHumanHandler(cfg *config.Config, log *logger.Logger) (*DigitalHumanHandler, error) {
	uploadService, err := services.NewUploadService(cfg, log)
	if err != nil {
		return nil, err
	}

	return &DigitalHumanHandler{
		service:          services.NewDigitalHumanService(cfg, log),
		uploadService:    uploadService,
		speechTTSService: services.NewSpeechTTSService(cfg, log),
		log:              log,
	}, nil
}

func (h *DigitalHumanHandler) Generate(c *gin.Context) {
	imageFile, imageHeader, err := c.Request.FormFile("image")
	if err != nil {
		response.BadRequest(c, "请上传图片")
		return
	}
	defer imageFile.Close()

	// `audio` is optional: the frontend can either upload audio or select a voice for TTS.
	audioFile, audioHeader, audioErr := c.Request.FormFile("audio")
	if audioErr == nil {
		defer audioFile.Close()
	}

	speechText := strings.TrimSpace(c.PostForm("speech_text"))
	motionText := strings.TrimSpace(c.PostForm("motion_text"))
	voiceID := strings.TrimSpace(c.PostForm("voice_id"))
	voiceType := strings.TrimSpace(c.PostForm("voice_type"))
	ttsResourceID := strings.TrimSpace(c.PostForm("resource_id"))
	audioURL := strings.TrimSpace(c.PostForm("audio_url"))

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

	normalizedImageBytes, normalizedFileName, normalizedContentType, normalizeErr := normalizeDigitalHumanImage(imageHeader.Filename, imageBytes, imageContentType)
	if normalizeErr != nil {
		h.log.Warnw("Failed to normalize digital human image, fallback to original", "error", normalizeErr, "filename", imageHeader.Filename)
		normalizedImageBytes = imageBytes
		normalizedFileName = imageHeader.Filename
		normalizedContentType = imageContentType
	}
	imageBase64 := base64.StdEncoding.EncodeToString(normalizedImageBytes)

	var audioContentType string
	if audioErr == nil {
		audioContentType = audioHeader.Header.Get("Content-Type")
		if audioContentType == "" {
			audioContentType = "application/octet-stream"
		}
		if !(strings.HasPrefix(audioContentType, "audio/") || audioContentType == "video/mp4") {
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
	} else {
		if speechText == "" {
			response.BadRequest(c, "请上传音频，或填写说话内容并选择音色")
			return
		}
		if voiceType == "" {
			response.BadRequest(c, "填写说话内容时请先选择音色")
			return
		}
		if !h.speechTTSService.IsConfigured() {
			response.Error(c, 400, "DIGITAL_HUMAN_TTS_NOT_ENABLED", "当前未配置文本转语音能力，请先上传音频，或联系管理员配置后重试")
			return
		}
	}

	uploadedImageURL, err := h.uploadService.UploadFile(bytes.NewReader(normalizedImageBytes), normalizedFileName, normalizedContentType, "digital-human/images")
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
		speechText = ""
		voiceType = ""
	} else if audioURL != "" {
		speechText = ""
		voiceType = ""
	} else {
		ttsAudioURL, ttsErr := h.speechTTSService.SynthesizeToURLWithResource(c.Request.Context(), speechText, voiceType, ttsResourceID)
		if ttsErr != nil {
			h.log.Errorw("Failed to synthesize speech from text", "error", ttsErr)
			ttsErrMsg := ttsErr.Error()
			switch {
			case strings.Contains(ttsErrMsg, "resource ID is mismatched with speaker related resource"):
				response.BadRequest(c, "当前音色不支持文本合成，请更换音色后重试")
			case voiceID != "" && strings.Contains(ttsErrMsg, "resource_id is required"):
				response.BadRequest(c, "当前音色缺少资源配置，请刷新音色列表后重试")
			case strings.Contains(ttsErrMsg, "requested resource not granted") ||
				strings.Contains(ttsErrMsg, "load grant: requested grant not found") ||
				strings.Contains(ttsErrMsg, "invalid auth token") ||
				strings.Contains(ttsErrMsg, "missing Authorization header"):
				response.Error(c, 400, "DIGITAL_HUMAN_TTS_NOT_ENABLED", "当前火山语音账号未开通TTS资源或配置无效，请上传音频后生成，或联系管理员开通后再试")
			default:
				response.InternalError(c, fmt.Sprintf("文本转音频失败: %v", ttsErr))
			}
			return
		}
		audioURL = strings.TrimSpace(ttsAudioURL)
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
		case strings.Contains(msg, "subject not recognized"):
			response.BadRequest(c, "未检测到清晰人物主体，请上传单人、正脸、主体完整的角色图后重试")
		case strings.Contains(msg, "subject recognition submit failed") || strings.Contains(msg, "subject recognition failed"):
			response.Error(c, 502, "DIGITAL_HUMAN_SUBJECT_RECOGNITION_FAILED", "角色主体识别失败，请稍后重试；若仍失败，请更换清晰单人角色图")
		case strings.Contains(msg, "audio_url or speech_text is required"):
			response.BadRequest(c, "请上传音频，或填写说话内容并选择音色")
		case strings.Contains(msg, "voice_type is required when speech_text is provided"):
			response.BadRequest(c, "填写说话内容时请先选择音色")
		case speechText != "" && (strings.Contains(msg, "\"code\":50215") || strings.Contains(msg, "Input invalid for this service")):
			response.Error(c, 400, "DIGITAL_HUMAN_TTS_NOT_ENABLED", "当前模型仅支持音频驱动，暂不支持直接文本配音。请上传音频后重试")
		case strings.Contains(msg, "Invalid parameter: AppID") || strings.Contains(msg, "UnauthorizedRequest.AppID"):
			response.Error(c, 400, "DIGITAL_HUMAN_TTS_NOT_ENABLED", "当前火山账号未开通文本配音能力，请先上传音频，或联系管理员开通后再试")
		case strings.Contains(msg, "\"code\":50218") || strings.Contains(msg, "Resource exists risk"):
			// Volcengine content risk control.
			response.BadRequest(c, "内容安全审核未通过，请更换角色图片/文案/音色后重试")
		case strings.Contains(msg, "\"code\":50514") || strings.Contains(msg, "Pre Audio Risk Not Pass"):
			response.BadRequest(c, "音频内容安全审核未通过，请更换文案、音色或上传其他音频后重试")
		case strings.Contains(msg, "\"code\":50430") || strings.Contains(msg, "API Concurrent Limit"):
			// Volcengine concurrency limit.
			response.Error(c, 429, "TOO_MANY_REQUESTS", "请求过于频繁，请稍后重试")
		case strings.Contains(msg, "504 Gateway Time-out") || strings.Contains(msg, "volcengine http status 504"):
			response.Error(c, 504, "DIGITAL_HUMAN_UPSTREAM_TIMEOUT", "数字人服务响应超时，系统已自动走兼容链路；请稍后重试")
		default:
			response.InternalError(c, msg)
		}
		return
	}

	response.Success(c, result)
}

func normalizeDigitalHumanImage(fileName string, imageBytes []byte, contentType string) ([]byte, string, string, error) {
	decoded, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, "", "", err
	}

	bounds := decoded.Bounds()
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(canvas, bounds, decoded, bounds.Min, draw.Over)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, canvas, &jpeg.Options{Quality: 88}); err != nil {
		return nil, "", "", err
	}

	baseName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	if baseName == "" {
		baseName = "digital-human"
	}

	if len(buf.Bytes()) >= len(imageBytes) && strings.EqualFold(contentType, "image/jpeg") {
		return imageBytes, fileName, contentType, nil
	}

	return buf.Bytes(), baseName + ".jpg", "image/jpeg", nil
}
