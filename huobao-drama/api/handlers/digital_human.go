package handlers

import (
	"bytes"
	"encoding/base64"
	"io"
	"strings"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type DigitalHumanHandler struct {
	service       *services.DigitalHumanService
	uploadService *services.UploadService
	log           *logger.Logger
}

func NewDigitalHumanHandler(cfg *config.Config, log *logger.Logger) (*DigitalHumanHandler, error) {
	uploadService, err := services.NewUploadService(cfg, log)
	if err != nil {
		return nil, err
	}

	return &DigitalHumanHandler{
		service:       services.NewDigitalHumanService(cfg, log),
		uploadService: uploadService,
		log:           log,
	}, nil
}

func (h *DigitalHumanHandler) Generate(c *gin.Context) {
	imageFile, imageHeader, err := c.Request.FormFile("image")
	if err != nil {
		response.BadRequest(c, "请上传图片")
		return
	}
	defer imageFile.Close()

	// `audio` is optional: the frontend can provide `audio_url` when only a voice is selected.
	audioFile, audioHeader, audioErr := c.Request.FormFile("audio")
	if audioErr == nil {
		defer audioFile.Close()
	}

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
		if !(strings.HasPrefix(audioContentType, "audio/") || audioContentType == "video/mp4") {
			response.BadRequest(c, "只支持常见音频格式 (mp3, wav, m4a, ogg, aac)")
			return
		}
		if audioHeader.Size > 20*1024*1024 {
			response.BadRequest(c, "音频大小不能超过20MB")
			return
		}
	} else {
		if audioURL == "" {
			response.BadRequest(c, "请上传音频或选择音色")
			return
		}
		if !strings.HasPrefix(audioURL, "http://") && !strings.HasPrefix(audioURL, "https://") {
			response.BadRequest(c, "audio_url 必须是 http/https URL")
			return
		}
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
	}

	speechText := strings.TrimSpace(c.PostForm("speech_text"))
	motionText := strings.TrimSpace(c.PostForm("motion_text"))

	result, err := h.service.Generate(c.Request.Context(), &services.DigitalHumanRequest{
		ImageURL:    uploadedImageURL,
		ImageBase64: imageBase64,
		AudioURL:   audioURL,
		SpeechText: speechText,
		MotionText: motionText,
	})
	if err != nil {
		h.log.Errorw("Failed to generate digital human video", "error", err)
		msg := err.Error()
		switch {
		case strings.Contains(msg, "\"code\":50218") || strings.Contains(msg, "Resource exists risk"):
			// Volcengine content risk control.
			response.BadRequest(c, "内容安全审核未通过，请更换角色图片/文案/音色后重试")
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
