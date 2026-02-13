package handlers

import (
	"bytes"
	"io"
	"mime"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VoiceLibraryHandler struct {
	service       *services.VoiceLibraryService
	uploadService *services.UploadService
	log           *logger.Logger
}

func NewVoiceLibraryHandler(cfg *config.Config, db *gorm.DB, log *logger.Logger) (*VoiceLibraryHandler, error) {
	service, err := services.NewVoiceLibraryService(cfg, db, log)
	if err != nil {
		return nil, err
	}

	uploadService, err := services.NewUploadService(cfg, log)
	if err != nil {
		return nil, err
	}

	return &VoiceLibraryHandler{
		service:       service,
		uploadService: uploadService,
		log:           log,
	}, nil
}

func (h *VoiceLibraryHandler) List(c *gin.Context) {
	voices, err := h.service.ListSpeakers(c.Request.Context())
	if err != nil {
		h.log.Errorw("Failed to list voice library", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, voices)
}

func (h *VoiceLibraryHandler) CreateCustom(c *gin.Context) {
	audioFile, audioHeader, err := c.Request.FormFile("audio")
	if err != nil {
		response.BadRequest(c, "请上传音色音频")
		return
	}
	defer audioFile.Close()

	name := strings.TrimSpace(c.PostForm("name"))
	audioContentType := strings.TrimSpace(audioHeader.Header.Get("Content-Type"))
	if audioContentType == "" {
		audioContentType = "application/octet-stream"
	}

	if !isAllowedVoiceAudioUpload(audioContentType, audioHeader.Filename) {
		response.BadRequest(c, "仅支持常见音频格式 (mp3, wav, m4a, ogg, aac, flac)")
		return
	}
	if audioHeader.Size > 20*1024*1024 {
		response.BadRequest(c, "音频大小不能超过20MB")
		return
	}

	audioBytes, err := io.ReadAll(io.LimitReader(audioFile, 20*1024*1024+1))
	if err != nil {
		h.log.Errorw("Failed to read custom voice audio", "error", err)
		response.InternalError(c, "读取音频失败")
		return
	}
	if len(audioBytes) > 20*1024*1024 {
		response.BadRequest(c, "音频大小不能超过20MB")
		return
	}

	uploadedAudioURL, err := h.uploadService.UploadFile(bytes.NewReader(audioBytes), audioHeader.Filename, audioContentType, "voice-library/custom")
	if err != nil {
		h.log.Errorw("Failed to upload custom voice audio", "error", err)
		response.InternalError(c, "上传音频失败")
		return
	}

	voice, err := h.service.CreateCustomVoice(c.Request.Context(), &services.CreateCustomVoiceRequest{
		Name:           name,
		SourceAudioURL: uploadedAudioURL,
		AudioBytes:     audioBytes,
		AudioFormat:    detectVoiceAudioFormat(audioHeader.Filename, audioContentType),
	})
	if err != nil {
		h.log.Errorw("Failed to create custom voice", "error", err)
		msg := err.Error()
		switch {
		case strings.Contains(msg, "license not found") || strings.Contains(msg, "requested resource not granted"):
			response.Error(c, 400, "VOICE_CLONE_NOT_ENABLED", "当前账号未开通音色复刻训练能力，请先开通后重试")
		case strings.Contains(msg, "no speaker slot") || strings.Contains(msg, "available training"):
			response.BadRequest(c, "当前没有可用音色槽位，请先在控制台购买或检查音色额度")
		default:
			response.InternalError(c, msg)
		}
		return
	}

	response.Success(c, voice)
}

func (h *VoiceLibraryHandler) GetCustomStatus(c *gin.Context) {
	idParam := strings.TrimSpace(c.Param("id"))
	id, ok := parseCustomVoiceIDParam(idParam)
	if !ok {
		response.BadRequest(c, "无效的音色ID")
		return
	}

	voice, err := h.service.RefreshCustomVoiceStatus(c.Request.Context(), id)
	if err != nil {
		h.log.Errorw("Failed to refresh custom voice status", "id", idParam, "error", err)
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "音色不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, voice)
}

func parseCustomVoiceIDParam(raw string) (uint, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, false
	}
	if strings.HasPrefix(trimmed, "custom-") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "custom-"))
	}
	parsed, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(parsed), true
}

func isAllowedVoiceAudioUpload(contentType, filename string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if strings.HasPrefix(ct, "audio/") {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}

	byExt := strings.ToLower(mime.TypeByExtension(ext))
	if strings.HasPrefix(byExt, "audio/") {
		return true
	}

	switch ext {
	case ".mp3", ".wav", ".m4a", ".ogg", ".aac", ".flac":
		return true
	default:
		return false
	}
}

func detectVoiceAudioFormat(filename, contentType string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch ext {
	case "mp3", "wav", "m4a", "ogg", "aac", "flac":
		return ext
	}

	ct := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.Contains(ct, "mpeg"):
		return "mp3"
	case strings.Contains(ct, "wav"):
		return "wav"
	case strings.Contains(ct, "x-m4a") || strings.Contains(ct, "mp4"):
		return "m4a"
	case strings.Contains(ct, "ogg"):
		return "ogg"
	case strings.Contains(ct, "aac"):
		return "aac"
	case strings.Contains(ct, "flac"):
		return "flac"
	default:
		return "wav"
	}
}
