package handlers

import (
	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	services2 "github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"mime"
	"path/filepath"
	"strings"
)

type UploadHandler struct {
	uploadService           *services2.UploadService
	characterLibraryService *services2.CharacterLibraryService
	log                     *logger.Logger
}

func NewUploadHandler(cfg *config.Config, log *logger.Logger, characterLibraryService *services2.CharacterLibraryService) (*UploadHandler, error) {
	uploadService, err := services2.NewUploadService(cfg, log)
	if err != nil {
		return nil, err
	}

	return &UploadHandler{
		uploadService:           uploadService,
		characterLibraryService: characterLibraryService,
		log:                     log,
	}, nil
}

// UploadImage 上传图片
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	// 检查文件类型
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 验证是图片类型
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !allowedTypes[contentType] {
		response.BadRequest(c, "只支持图片格式 (jpg, png, gif, webp)")
		return
	}

	// 检查文件大小 (10MB)
	if header.Size > 10*1024*1024 {
		response.BadRequest(c, "文件大小不能超过10MB")
		return
	}

	// 上传到MinIO
	fileURL, err := h.uploadService.UploadCharacterImage(file, header.Filename, contentType)
	if err != nil {
		h.log.Errorw("Failed to upload image", "error", err)
		response.InternalError(c, "上传失败")
		return
	}

	response.Success(c, gin.H{
		"url":      fileURL,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

// UploadFile 上传数字丝路商品素材
func (h *UploadHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	contentType := resolveUploadContentType(header.Header.Get("Content-Type"), header.Filename)
	if !isAllowedMaterialContentType(contentType) {
		response.BadRequest(c, "暂不支持该文件类型")
		return
	}

	if header.Size > 100*1024*1024 {
		response.BadRequest(c, "文件大小不能超过100MB")
		return
	}

	category := sanitizeUploadCategory(c.PostForm("category"))
	if category == "" {
		category = sanitizeUploadCategory(c.Query("category"))
	}
	if category == "" {
		category = "materials"
	}

	fileURL, err := h.uploadService.UploadFile(file, header.Filename, contentType, category)
	if err != nil {
		h.log.Errorw("Failed to upload material", "error", err, "category", category)
		response.InternalError(c, "上传失败")
		return
	}

	response.Success(c, gin.H{
		"url":          fileURL,
		"filename":     header.Filename,
		"size":         header.Size,
		"content_type": contentType,
		"category":     category,
		"asset_type":   inferMaterialAssetType(contentType),
	})
}

// UploadCharacterImage 上传角色图片（带角色ID）
func (h *UploadHandler) UploadCharacterImage(c *gin.Context) {
	deviceID := middlewares2.GetDeviceID(c)
	characterID := c.Param("id")

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	// 检查文件类型
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 验证是图片类型
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !allowedTypes[contentType] {
		response.BadRequest(c, "只支持图片格式 (jpg, png, gif, webp)")
		return
	}

	// 检查文件大小 (10MB)
	if header.Size > 10*1024*1024 {
		response.BadRequest(c, "文件大小不能超过10MB")
		return
	}

	// 上传到MinIO
	fileURL, err := h.uploadService.UploadCharacterImage(file, header.Filename, contentType)
	if err != nil {
		h.log.Errorw("Failed to upload character image", "error", err)
		response.InternalError(c, "上传失败")
		return
	}

	// 更新角色的image_url字段到数据库
	err = h.characterLibraryService.UploadCharacterImage(characterID, fileURL, deviceID)
	if err != nil {
		h.log.Errorw("Failed to update character image_url", "error", err, "character_id", characterID)
		response.InternalError(c, "更新角色图片失败")
		return
	}

	h.log.Infow("Character image uploaded and saved", "character_id", characterID, "url", fileURL)

	response.Success(c, gin.H{
		"url":      fileURL,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

func resolveUploadContentType(contentType string, filename string) string {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if contentType != "" && contentType != "application/octet-stream" {
		return contentType
	}

	if inferred := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); inferred != "" {
		return strings.Split(inferred, ";")[0]
	}
	return "application/octet-stream"
}

func sanitizeUploadCategory(category string) string {
	category = strings.TrimSpace(strings.ToLower(category))
	if category == "" {
		return ""
	}

	var builder strings.Builder
	for _, r := range category {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
		}
	}
	return strings.Trim(builder.String(), "-_")
}

func isAllowedMaterialContentType(contentType string) bool {
	if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/") {
		return true
	}

	switch contentType {
	case "application/pdf",
		"text/plain",
		"text/csv",
		"application/json",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/zip":
		return true
	default:
		return false
	}
}

func inferMaterialAssetType(contentType string) string {
	switch {
	case strings.HasPrefix(contentType, "video/"):
		return "video"
	case strings.HasPrefix(contentType, "audio/"):
		return "audio"
	default:
		return "image"
	}
}
