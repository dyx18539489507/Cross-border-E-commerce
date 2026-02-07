package handlers

import (
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

	audioFile, audioHeader, err := c.Request.FormFile("audio")
	if err != nil {
		response.BadRequest(c, "请上传音频")
		return
	}
	defer audioFile.Close()

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

	audioContentType := audioHeader.Header.Get("Content-Type")
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

	imageURL, err := h.uploadService.UploadFile(imageFile, imageHeader.Filename, imageContentType, "digital-human/images")
	if err != nil {
		h.log.Errorw("Failed to upload image", "error", err)
		response.InternalError(c, "图片上传失败")
		return
	}

	audioURL, err := h.uploadService.UploadFile(audioFile, audioHeader.Filename, audioContentType, "digital-human/audios")
	if err != nil {
		h.log.Errorw("Failed to upload audio", "error", err)
		response.InternalError(c, "音频上传失败")
		return
	}

	speechText := strings.TrimSpace(c.PostForm("speech_text"))
	motionText := strings.TrimSpace(c.PostForm("motion_text"))

	result, err := h.service.Generate(c.Request.Context(), &services.DigitalHumanRequest{
		ImageURL:   imageURL,
		AudioURL:   audioURL,
		SpeechText: speechText,
		MotionText: motionText,
	})
	if err != nil {
		h.log.Errorw("Failed to generate digital human video", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}
