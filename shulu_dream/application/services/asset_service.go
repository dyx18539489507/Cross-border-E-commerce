package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/external/ffmpeg"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

type AssetService struct {
	db     *gorm.DB
	log    *logger.Logger
	ffmpeg *ffmpeg.FFmpeg
}

func NewAssetService(db *gorm.DB, log *logger.Logger) *AssetService {
	return &AssetService{
		db:     db,
		log:    log,
		ffmpeg: ffmpeg.NewFFmpeg(log),
	}
}

func firstAssetDeviceID(deviceIDs []string) string {
	if len(deviceIDs) == 0 {
		return ""
	}
	return strings.TrimSpace(deviceIDs[0])
}

func (s *AssetService) applyAssetDeviceScope(query *gorm.DB, deviceID string) *gorm.DB {
	if deviceID == "" {
		return query
	}
	dramaSubQuery := s.db.Model(&models.Drama{}).Select("id").Where("device_id = ?", deviceID)
	return query.Where("drama_id IN (?)", dramaSubQuery)
}

type CreateAssetRequest struct {
	DramaID      *string          `json:"drama_id"`
	Name         string           `json:"name" binding:"required"`
	Description  *string          `json:"description"`
	Type         models.AssetType `json:"type" binding:"required"`
	Category     *string          `json:"category"`
	URL          string           `json:"url" binding:"required"`
	ThumbnailURL *string          `json:"thumbnail_url"`
	LocalPath    *string          `json:"local_path"`
	FileSize     *int64           `json:"file_size"`
	MimeType     *string          `json:"mime_type"`
	Width        *int             `json:"width"`
	Height       *int             `json:"height"`
	Duration     *int             `json:"duration"`
	Format       *string          `json:"format"`
	ImageGenID   *uint            `json:"image_gen_id"`
	VideoGenID   *uint            `json:"video_gen_id"`
	TagIDs       []uint           `json:"tag_ids"`
}

type UpdateAssetRequest struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	Category     *string `json:"category"`
	ThumbnailURL *string `json:"thumbnail_url"`
	TagIDs       []uint  `json:"tag_ids"`
	IsFavorite   *bool   `json:"is_favorite"`
}

type ListAssetsRequest struct {
	DramaID      *string           `json:"drama_id"`
	EpisodeID    *uint             `json:"episode_id"`
	StoryboardID *uint             `json:"storyboard_id"`
	Type         *models.AssetType `json:"type"`
	Category     string            `json:"category"`
	TagIDs       []uint            `json:"tag_ids"`
	IsFavorite   *bool             `json:"is_favorite"`
	Search       string            `json:"search"`
	Page         int               `json:"page"`
	PageSize     int               `json:"page_size"`
}

func (s *AssetService) CreateAsset(req *CreateAssetRequest, deviceIDs ...string) (*models.Asset, error) {
	deviceID := firstAssetDeviceID(deviceIDs)
	var dramaID *uint
	if req.DramaID != nil && *req.DramaID != "" {
		id, err := strconv.ParseUint(*req.DramaID, 10, 32)
		if err == nil {
			uid := uint(id)
			dramaID = &uid
		}
	}

	if dramaID != nil {
		dramaQuery := s.db.Where("id = ?", *dramaID)
		if deviceID != "" {
			dramaQuery = dramaQuery.Where("device_id = ?", deviceID)
		}
		var drama models.Drama
		if err := dramaQuery.First(&drama).Error; err != nil {
			return nil, fmt.Errorf("drama not found")
		}
	} else if deviceID != "" {
		return nil, fmt.Errorf("drama_id is required")
	}

	asset := &models.Asset{
		DramaID:      dramaID,
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		Category:     req.Category,
		URL:          req.URL,
		ThumbnailURL: req.ThumbnailURL,
		LocalPath:    req.LocalPath,
		FileSize:     req.FileSize,
		MimeType:     req.MimeType,
		Width:        req.Width,
		Height:       req.Height,
		Duration:     req.Duration,
		Format:       req.Format,
		ImageGenID:   req.ImageGenID,
		VideoGenID:   req.VideoGenID,
	}

	if err := s.db.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return asset, nil
}

func (s *AssetService) UpdateAsset(assetID uint, req *UpdateAssetRequest, deviceIDs ...string) (*models.Asset, error) {
	deviceID := firstAssetDeviceID(deviceIDs)
	var asset models.Asset
	query := s.db.Model(&models.Asset{}).Where("id = ?", assetID)
	query = s.applyAssetDeviceScope(query, deviceID)
	if err := query.First(&asset).Error; err != nil {
		return nil, fmt.Errorf("asset not found")
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.ThumbnailURL != nil {
		updates["thumbnail_url"] = *req.ThumbnailURL
	}
	if req.IsFavorite != nil {
		updates["is_favorite"] = *req.IsFavorite
	}

	if len(updates) > 0 {
		if err := s.db.Model(&asset).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update asset: %w", err)
		}
	}

	if err := s.db.First(&asset, assetID).Error; err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *AssetService) GetAsset(assetID uint, deviceIDs ...string) (*models.Asset, error) {
	deviceID := firstAssetDeviceID(deviceIDs)
	var asset models.Asset
	query := s.db.Model(&models.Asset{}).Where("id = ?", assetID)
	query = s.applyAssetDeviceScope(query, deviceID)
	if err := query.First(&asset).Error; err != nil {
		return nil, err
	}

	s.db.Model(&asset).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	return &asset, nil
}

func (s *AssetService) ListAssets(req *ListAssetsRequest, deviceIDs ...string) ([]models.Asset, int64, error) {
	deviceID := firstAssetDeviceID(deviceIDs)
	query := s.db.Model(&models.Asset{})
	query = s.applyAssetDeviceScope(query, deviceID)

	if req.DramaID != nil {
		var dramaID uint64
		dramaID, _ = strconv.ParseUint(*req.DramaID, 10, 32)
		query = query.Where("drama_id = ?", uint(dramaID))
	}

	if req.EpisodeID != nil {
		query = query.Where("episode_id = ?", *req.EpisodeID)
	}

	if req.StoryboardID != nil {
		query = query.Where("storyboard_id = ?", *req.StoryboardID)
	}

	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	if req.IsFavorite != nil {
		query = query.Where("is_favorite = ?", *req.IsFavorite)
	}

	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var assets []models.Asset
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(req.PageSize).Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (s *AssetService) DeleteAsset(assetID uint, deviceIDs ...string) error {
	deviceID := firstAssetDeviceID(deviceIDs)
	query := s.db.Model(&models.Asset{}).Where("id = ?", assetID)
	query = s.applyAssetDeviceScope(query, deviceID)
	result := query.Delete(&models.Asset{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("asset not found")
	}
	return nil
}

func (s *AssetService) ImportFromImageGen(imageGenID uint, deviceIDs ...string) (*models.Asset, error) {
	deviceID := firstAssetDeviceID(deviceIDs)
	var imageGen models.ImageGeneration
	imageQuery := s.db.Model(&models.ImageGeneration{}).Where("id = ?", imageGenID)
	if deviceID != "" {
		dramaSubQuery := s.db.Model(&models.Drama{}).Select("id").Where("device_id = ?", deviceID)
		imageQuery = imageQuery.Where("drama_id IN (?)", dramaSubQuery)
	}
	if err := imageQuery.First(&imageGen).Error; err != nil {
		return nil, fmt.Errorf("image generation not found")
	}

	if imageGen.Status != models.ImageStatusCompleted || imageGen.ImageURL == nil {
		return nil, fmt.Errorf("image is not ready")
	}

	dramaID := imageGen.DramaID
	asset := &models.Asset{
		Name:       fmt.Sprintf("Image_%d", imageGen.ID),
		Type:       models.AssetTypeImage,
		URL:        *imageGen.ImageURL,
		DramaID:    &dramaID,
		ImageGenID: &imageGenID,
		Width:      imageGen.Width,
		Height:     imageGen.Height,
	}

	if err := s.db.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return asset, nil
}

func (s *AssetService) ImportFromVideoGen(videoGenID uint, deviceIDs ...string) (*models.Asset, error) {
	deviceID := firstAssetDeviceID(deviceIDs)
	var existing models.Asset
	existingQuery := s.db.Model(&models.Asset{}).Where("type = ? AND video_gen_id = ?", models.AssetTypeVideo, videoGenID)
	existingQuery = s.applyAssetDeviceScope(existingQuery, deviceID)
	if err := existingQuery.First(&existing).Error; err == nil {
		return nil, fmt.Errorf("该视频已在素材库中")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing video asset: %w", err)
	}

	var videoGen models.VideoGeneration
	videoQuery := s.db.Preload("Storyboard.Episode").Model(&models.VideoGeneration{}).Where("id = ?", videoGenID)
	if deviceID != "" {
		dramaSubQuery := s.db.Model(&models.Drama{}).Select("id").Where("device_id = ?", deviceID)
		videoQuery = videoQuery.Where("drama_id IN (?)", dramaSubQuery)
	}
	if err := videoQuery.First(&videoGen).Error; err != nil {
		return nil, fmt.Errorf("video generation not found")
	}

	if videoGen.Status != models.VideoStatusCompleted || videoGen.VideoURL == nil {
		return nil, fmt.Errorf("video is not ready")
	}

	dramaID := videoGen.DramaID

	var episodeID *uint
	var storyboardNum *int
	if videoGen.Storyboard != nil {
		episodeID = &videoGen.Storyboard.Episode.ID
		storyboardNum = &videoGen.Storyboard.StoryboardNumber
	}

	asset := &models.Asset{
		Name:          fmt.Sprintf("Video_%d", videoGen.ID),
		Type:          models.AssetTypeVideo,
		URL:           *videoGen.VideoURL,
		DramaID:       &dramaID,
		EpisodeID:     episodeID,
		StoryboardID:  videoGen.StoryboardID,
		StoryboardNum: storyboardNum,
		VideoGenID:    &videoGenID,
		Duration:      videoGen.Duration,
		Width:         videoGen.Width,
		Height:        videoGen.Height,
	}

	if videoGen.FirstFrameURL != nil {
		asset.ThumbnailURL = videoGen.FirstFrameURL
	}

	if err := s.db.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return asset, nil
}
