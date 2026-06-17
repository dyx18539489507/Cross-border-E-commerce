package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DramaService struct {
	db                 *gorm.DB
	log                *logger.Logger
	complianceService  *ComplianceService
	complianceCache    map[string]cachedComplianceResult
	complianceCacheMu  sync.RWMutex
	complianceTokens   map[string]issuedComplianceToken
	complianceTokensMu sync.RWMutex
}

var (
	ErrTargetCountryRequired     = errors.New("target_country is required")
	ErrComplianceRiskForbidden   = errors.New("compliance risk level red, creation forbidden")
	ErrCompliancePrecheckInvalid = errors.New("compliance precheck token invalid or expired")
)

const complianceCacheTTL = 10 * time.Minute

type cachedComplianceResult struct {
	result    *ComplianceResult
	expiresAt time.Time
}

type issuedComplianceToken struct {
	cacheKey  string
	deviceID  string
	result    *ComplianceResult
	expiresAt time.Time
}

func normalizeStaticURL(raw *string) {
	NormalizeImageURLPtr(raw)
}

func normalizeDramaImageURLs(drama *models.Drama) {
	for i := range drama.Characters {
		normalizeStaticURL(drama.Characters[i].ImageURL)
	}
	for i := range drama.Scenes {
		normalizeStaticURL(drama.Scenes[i].ImageURL)
	}
	for i := range drama.Episodes {
		for j := range drama.Episodes[i].Characters {
			normalizeStaticURL(drama.Episodes[i].Characters[j].ImageURL)
		}
		for j := range drama.Episodes[i].Scenes {
			normalizeStaticURL(drama.Episodes[i].Scenes[j].ImageURL)
		}
		for j := range drama.Episodes[i].Storyboards {
			normalizeStaticURL(drama.Episodes[i].Storyboards[j].ComposedImage)
			if drama.Episodes[i].Storyboards[j].Background != nil {
				normalizeStaticURL(drama.Episodes[i].Storyboards[j].Background.ImageURL)
			}
		}
	}
}

type imageGenerationStatusSnapshot struct {
	Status   string
	ErrorMsg *string
}

type imageGenerationStatusRow struct {
	TargetID uint
	Status   models.ImageGenerationStatus
	ErrorMsg *string
}

func appendUniqueIDs(dst []uint, seen map[uint]struct{}, ids ...uint) []uint {
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		dst = append(dst, id)
	}
	return dst
}

func (s *DramaService) loadLatestImageStatuses(column string, ids []uint) (map[uint]imageGenerationStatusSnapshot, error) {
	statuses := make(map[uint]imageGenerationStatusSnapshot, len(ids))
	if len(ids) == 0 {
		return statuses, nil
	}

	var rows []imageGenerationStatusRow
	activeStatuses := []string{string(models.ImageStatusPending), string(models.ImageStatusProcessing)}
	activeSelect := fmt.Sprintf("%s AS target_id, status, error_msg", column)

	if err := s.db.Model(&models.ImageGeneration{}).
		Select(activeSelect).
		Where(fmt.Sprintf("%s IN ?", column), ids).
		Where("status IN ?", activeStatuses).
		Order(fmt.Sprintf("%s ASC, created_at DESC", column)).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		if _, exists := statuses[row.TargetID]; exists {
			continue
		}
		statuses[row.TargetID] = imageGenerationStatusSnapshot{
			Status:   string(row.Status),
			ErrorMsg: row.ErrorMsg,
		}
	}

	rows = nil
	failedSelect := fmt.Sprintf("%s AS target_id, status, error_msg", column)
	if err := s.db.Model(&models.ImageGeneration{}).
		Select(failedSelect).
		Where(fmt.Sprintf("%s IN ?", column), ids).
		Where("status = ?", models.ImageStatusFailed).
		Order(fmt.Sprintf("%s ASC, created_at DESC", column)).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		if _, exists := statuses[row.TargetID]; exists {
			continue
		}
		statuses[row.TargetID] = imageGenerationStatusSnapshot{
			Status:   string(row.Status),
			ErrorMsg: row.ErrorMsg,
		}
	}

	return statuses, nil
}

func (s *DramaService) applyEpisodeImageStatuses(drama *models.Drama) error {
	characterSeen := make(map[uint]struct{})
	sceneSeen := make(map[uint]struct{})
	characterIDs := make([]uint, 0)
	sceneIDs := make([]uint, 0)

	for i := range drama.Episodes {
		for _, character := range drama.Episodes[i].Characters {
			characterIDs = appendUniqueIDs(characterIDs, characterSeen, character.ID)
		}
		for _, scene := range drama.Episodes[i].Scenes {
			sceneIDs = appendUniqueIDs(sceneIDs, sceneSeen, scene.ID)
		}
	}

	characterStatuses, err := s.loadLatestImageStatuses("character_id", characterIDs)
	if err != nil {
		return err
	}
	sceneStatuses, err := s.loadLatestImageStatuses("scene_id", sceneIDs)
	if err != nil {
		return err
	}

	for i := range drama.Episodes {
		for j := range drama.Episodes[i].Characters {
			status, exists := characterStatuses[drama.Episodes[i].Characters[j].ID]
			if !exists {
				drama.Episodes[i].Characters[j].ImageGenerationStatus = nil
				drama.Episodes[i].Characters[j].ImageGenerationError = nil
				continue
			}
			statusValue := status.Status
			drama.Episodes[i].Characters[j].ImageGenerationStatus = &statusValue
			drama.Episodes[i].Characters[j].ImageGenerationError = status.ErrorMsg
		}

		for j := range drama.Episodes[i].Scenes {
			status, exists := sceneStatuses[drama.Episodes[i].Scenes[j].ID]
			if !exists {
				drama.Episodes[i].Scenes[j].ImageGenerationStatus = nil
				drama.Episodes[i].Scenes[j].ImageGenerationError = nil
				continue
			}
			statusValue := status.Status
			drama.Episodes[i].Scenes[j].ImageGenerationStatus = &statusValue
			drama.Episodes[i].Scenes[j].ImageGenerationError = status.ErrorMsg
		}
	}

	return nil
}

func NewDramaService(db *gorm.DB, log *logger.Logger, complianceService *ComplianceService) *DramaService {
	return &DramaService{
		db:                db,
		log:               log,
		complianceService: complianceService,
		complianceCache:   make(map[string]cachedComplianceResult),
		complianceTokens:  make(map[string]issuedComplianceToken),
	}
}

type CreateDramaRequest struct {
	Title                  string   `json:"title" binding:"required,min=1,max=50"`
	Description            string   `json:"description" binding:"required,min=1,max=500"`
	TargetCountry          []string `json:"target_country" binding:"required,min=1"`
	MaterialComposition    string   `json:"material_composition" binding:"omitempty,max=200"`
	MarketingSellingPoints string   `json:"marketing_selling_points" binding:"omitempty,max=200"`
	ComplianceToken        string   `json:"compliance_token" binding:"omitempty,max=128"`
	Genre                  string   `json:"genre"`
	Tags                   string   `json:"tags"`
}

type UpdateDramaRequest struct {
	Title                  string   `json:"title" binding:"omitempty,min=1,max=50"`
	Description            string   `json:"description" binding:"omitempty,max=500"`
	TargetCountry          []string `json:"target_country" binding:"omitempty,min=1"`
	MaterialComposition    string   `json:"material_composition" binding:"omitempty,max=200"`
	MarketingSellingPoints string   `json:"marketing_selling_points" binding:"omitempty,max=200"`
	Genre                  string   `json:"genre"`
	Tags                   string   `json:"tags"`
	Status                 string   `json:"status" binding:"omitempty,oneof=draft planning production completed archived"`
}

type DramaListQuery struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Status   string `form:"status"`
	Genre    string `form:"genre"`
	Keyword  string `form:"keyword"`
}

type preparedCreateDramaInput struct {
	title                  string
	description            string
	targetCountries        []string
	targetCountry          string
	materialComposition    string
	marketingSellingPoints string
}

func prepareCreateDramaInput(req *CreateDramaRequest) (*preparedCreateDramaInput, error) {
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	targetCountries := normalizeCountryCodes(req.TargetCountry)
	if len(targetCountries) == 0 {
		return nil, ErrTargetCountryRequired
	}

	return &preparedCreateDramaInput{
		title:                  title,
		description:            description,
		targetCountries:        targetCountries,
		targetCountry:          strings.Join(targetCountries, ","),
		materialComposition:    strings.TrimSpace(req.MaterialComposition),
		marketingSellingPoints: strings.TrimSpace(req.MarketingSellingPoints),
	}, nil
}

func cloneComplianceResult(result *ComplianceResult) *ComplianceResult {
	if result == nil {
		return nil
	}

	return &ComplianceResult{
		Score:                    result.Score,
		Level:                    result.Level,
		LevelLabel:               result.LevelLabel,
		Summary:                  result.Summary,
		NonCompliancePoints:      append([]string{}, result.NonCompliancePoints...),
		RectificationSuggestions: append([]string{}, result.RectificationSuggestions...),
		SuggestedCategories:      append([]string{}, result.SuggestedCategories...),
	}
}

func buildComplianceCacheKey(input *preparedCreateDramaInput, deviceID string) string {
	payload := struct {
		DeviceID               string   `json:"device_id"`
		Title                  string   `json:"title"`
		Description            string   `json:"description"`
		TargetCountry          []string `json:"target_country"`
		MaterialComposition    string   `json:"material_composition"`
		MarketingSellingPoints string   `json:"marketing_selling_points"`
	}{
		DeviceID:               strings.TrimSpace(deviceID),
		Title:                  input.title,
		Description:            input.description,
		TargetCountry:          append([]string{}, input.targetCountries...),
		MaterialComposition:    input.materialComposition,
		MarketingSellingPoints: input.marketingSellingPoints,
	}

	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func generateComplianceToken() string {
	randomBytes := make([]byte, 18)
	if _, err := rand.Read(randomBytes); err != nil {
		fallback := sha256.Sum256([]byte(fmt.Sprintf("compliance-%d", time.Now().UnixNano())))
		return hex.EncodeToString(fallback[:16])
	}
	return hex.EncodeToString(randomBytes)
}

func (s *DramaService) getCachedComplianceResult(cacheKey string) *ComplianceResult {
	if cacheKey == "" {
		return nil
	}

	now := time.Now()

	s.complianceCacheMu.RLock()
	entry, ok := s.complianceCache[cacheKey]
	s.complianceCacheMu.RUnlock()
	if !ok {
		return nil
	}

	if now.After(entry.expiresAt) {
		s.complianceCacheMu.Lock()
		current, exists := s.complianceCache[cacheKey]
		if exists && now.After(current.expiresAt) {
			delete(s.complianceCache, cacheKey)
		}
		s.complianceCacheMu.Unlock()
		return nil
	}

	return cloneComplianceResult(entry.result)
}

func (s *DramaService) setCachedComplianceResult(cacheKey string, result *ComplianceResult) {
	if cacheKey == "" || result == nil {
		return
	}

	s.complianceCacheMu.Lock()
	s.complianceCache[cacheKey] = cachedComplianceResult{
		result:    cloneComplianceResult(result),
		expiresAt: time.Now().Add(complianceCacheTTL),
	}
	s.complianceCacheMu.Unlock()
}

func (s *DramaService) issueComplianceToken(cacheKey, deviceID string, result *ComplianceResult) string {
	if cacheKey == "" || result == nil {
		return ""
	}

	token := generateComplianceToken()
	s.complianceTokensMu.Lock()
	s.complianceTokens[token] = issuedComplianceToken{
		cacheKey:  cacheKey,
		deviceID:  strings.TrimSpace(deviceID),
		result:    cloneComplianceResult(result),
		expiresAt: time.Now().Add(complianceCacheTTL),
	}
	s.complianceTokensMu.Unlock()
	return token
}

func (s *DramaService) getComplianceResultByToken(token string, input *preparedCreateDramaInput, deviceID string) (*ComplianceResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrCompliancePrecheckInvalid
	}

	now := time.Now()
	s.complianceTokensMu.RLock()
	entry, ok := s.complianceTokens[token]
	s.complianceTokensMu.RUnlock()
	if !ok {
		return nil, ErrCompliancePrecheckInvalid
	}

	if now.After(entry.expiresAt) {
		s.complianceTokensMu.Lock()
		current, exists := s.complianceTokens[token]
		if exists && now.After(current.expiresAt) {
			delete(s.complianceTokens, token)
		}
		s.complianceTokensMu.Unlock()
		return nil, ErrCompliancePrecheckInvalid
	}

	normalizedDeviceID := strings.TrimSpace(deviceID)
	if entry.deviceID != normalizedDeviceID {
		return nil, ErrCompliancePrecheckInvalid
	}

	expectedCacheKey := buildComplianceCacheKey(input, normalizedDeviceID)
	if entry.cacheKey != expectedCacheKey {
		return nil, ErrCompliancePrecheckInvalid
	}

	if cached := s.getCachedComplianceResult(entry.cacheKey); cached != nil {
		return cached, nil
	}

	if entry.result == nil {
		return nil, ErrCompliancePrecheckInvalid
	}

	s.setCachedComplianceResult(entry.cacheKey, entry.result)
	return cloneComplianceResult(entry.result), nil
}

func (s *DramaService) evaluateCompliance(input *preparedCreateDramaInput, deviceID string) *ComplianceResult {
	cacheKey := buildComplianceCacheKey(input, deviceID)
	if cached := s.getCachedComplianceResult(cacheKey); cached != nil {
		return cached
	}

	complianceResult := &ComplianceResult{
		Score:                    0,
		Level:                    ComplianceRiskGreen,
		LevelLabel:               "低",
		Summary:                  "未进行合规校验",
		NonCompliancePoints:      []string{},
		RectificationSuggestions: []string{},
		SuggestedCategories:      []string{},
	}
	if s.complianceService != nil {
		if evaluated, err := s.complianceService.Evaluate(ComplianceRequest{
			Title:                  input.title,
			Description:            input.description,
			TargetCountry:          input.targetCountries,
			MaterialComposition:    input.materialComposition,
			MarketingSellingPoints: input.marketingSellingPoints,
		}); err == nil && evaluated != nil {
			complianceResult = evaluated
		} else if err != nil {
			s.log.Warnw("Compliance evaluation error, continue with default result", "error", err)
		}
	}

	s.setCachedComplianceResult(cacheKey, complianceResult)
	return complianceResult
}

func (s *DramaService) EvaluateCompliance(req *CreateDramaRequest, deviceIDs ...string) (*ComplianceResult, string, error) {
	input, err := prepareCreateDramaInput(req)
	if err != nil {
		return nil, "", err
	}
	deviceID := ""
	if len(deviceIDs) > 0 {
		deviceID = strings.TrimSpace(deviceIDs[0])
	}
	complianceResult := s.evaluateCompliance(input, deviceID)
	cacheKey := buildComplianceCacheKey(input, deviceID)
	complianceToken := s.issueComplianceToken(cacheKey, deviceID, complianceResult)
	return complianceResult, complianceToken, nil
}

func (s *DramaService) CreateDrama(req *CreateDramaRequest, deviceID string) (*models.Drama, *ComplianceResult, error) {
	input, err := prepareCreateDramaInput(req)
	if err != nil {
		return nil, nil, err
	}

	var complianceResult *ComplianceResult
	if strings.TrimSpace(req.ComplianceToken) != "" {
		complianceResult, err = s.getComplianceResultByToken(req.ComplianceToken, input, deviceID)
		if err != nil {
			return nil, nil, err
		}
	} else {
		complianceResult = s.evaluateCompliance(input, deviceID)
	}

	if complianceResult.Level == ComplianceRiskRed {
		s.log.Warnw(
			"Drama creation blocked by compliance red risk",
			"title", input.title,
			"score", complianceResult.Score,
			"level", complianceResult.Level,
			"device_id", deviceID,
		)
		return nil, complianceResult, ErrComplianceRiskForbidden
	}

	complianceReportJSON, _ := json.Marshal(complianceResult)

	drama := &models.Drama{
		DeviceID:         deviceID,
		Title:            input.title,
		Status:           "draft",
		TargetCountry:    input.targetCountry,
		ComplianceScore:  complianceResult.Score,
		ComplianceLevel:  string(complianceResult.Level),
		ComplianceReport: datatypes.JSON(complianceReportJSON),
	}

	if input.description != "" {
		drama.Description = &input.description
	}
	if req.Genre != "" {
		drama.Genre = &req.Genre
	}
	if input.materialComposition != "" {
		drama.MaterialComposition = &input.materialComposition
	}
	if input.marketingSellingPoints != "" {
		drama.MarketingSellingPoints = &input.marketingSellingPoints
	}

	if err := s.db.Create(drama).Error; err != nil {
		s.log.Errorw("Failed to create drama", "error", err)
		return nil, nil, err
	}

	s.log.Infow("Drama created", "drama_id", drama.ID, "compliance_score", complianceResult.Score, "risk_level", complianceResult.Level)
	return drama, complianceResult, nil
}

func (s *DramaService) GetDrama(dramaID string, deviceID string) (*models.Drama, error) {
	var drama models.Drama
	err := s.db.Where("id = ? AND device_id = ?", dramaID, deviceID).
		Preload("Characters").          // 加载Drama级别的角色
		Preload("Episodes.Characters"). // 加载每个章节关联的角色
		Preload("Episodes.Scenes").     // 加载每个章节关联的场景
		Preload("Episodes.Storyboards", func(db *gorm.DB) *gorm.DB {
			return db.Order("storyboards.storyboard_number ASC")
		}).
		First(&drama).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drama not found")
		}
		s.log.Errorw("Failed to get drama", "error", err)
		return nil, err
	}

	normalizeDramaImageURLs(&drama)
	if err := s.applyEpisodeImageStatuses(&drama); err != nil {
		s.log.Errorw("Failed to load episode image statuses", "error", err, "drama_id", dramaID)
		return nil, err
	}

	// 统计每个剧集的时长（基于场景时长之和）
	for i := range drama.Episodes {
		totalDuration := 0
		for _, scene := range drama.Episodes[i].Storyboards {
			totalDuration += scene.Duration
		}
		// 更新剧集时长（秒转分钟，向上取整）
		durationMinutes := (totalDuration + 59) / 60
		originalDuration := drama.Episodes[i].Duration
		drama.Episodes[i].Duration = durationMinutes

		// 如果数据库中的时长与计算的不一致，更新数据库
		if originalDuration != durationMinutes {
			s.db.Model(&models.Episode{}).Where("id = ?", drama.Episodes[i].ID).Update("duration", durationMinutes)
		}
	}

	// 整合所有剧集的场景到Drama级别的Scenes字段
	sceneMap := make(map[uint]*models.Scene) // 用于去重
	for i := range drama.Episodes {
		for j := range drama.Episodes[i].Scenes {
			scene := &drama.Episodes[i].Scenes[j]
			sceneMap[scene.ID] = scene
		}
	}

	// 将整合的场景添加到drama.Scenes
	drama.Scenes = make([]models.Scene, 0, len(sceneMap))
	for _, scene := range sceneMap {
		drama.Scenes = append(drama.Scenes, *scene)
	}

	return &drama, nil
}

func (s *DramaService) ListDramas(query *DramaListQuery, deviceID string) ([]models.Drama, int64, error) {
	var dramas []models.Drama
	var total int64

	db := s.db.Model(&models.Drama{}).Where("device_id = ?", deviceID)

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.Genre != "" {
		db = db.Where("genre = ?", query.Genre)
	}

	if query.Keyword != "" {
		likeKeyword := "%" + query.Keyword + "%"
		db = db.Where(
			"title LIKE ? OR description LIKE ? OR target_country LIKE ? OR material_composition LIKE ? OR marketing_selling_points LIKE ?",
			likeKeyword, likeKeyword, likeKeyword, likeKeyword, likeKeyword,
		)
	}

	if err := db.Count(&total).Error; err != nil {
		s.log.Errorw("Failed to count dramas", "error", err)
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("updated_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Preload("Episodes.Storyboards", func(db *gorm.DB) *gorm.DB {
			return db.Order("storyboards.storyboard_number ASC")
		}).
		Find(&dramas).Error

	if err != nil {
		s.log.Errorw("Failed to list dramas", "error", err)
		return nil, 0, err
	}

	// 统计每个剧本的每个剧集的时长（基于场景时长之和）
	for i := range dramas {
		normalizeDramaImageURLs(&dramas[i])
		for j := range dramas[i].Episodes {
			totalDuration := 0
			for _, scene := range dramas[i].Episodes[j].Storyboards {
				totalDuration += scene.Duration
			}
			// 更新剧集时长（秒转分钟，向上取整）
			durationMinutes := (totalDuration + 59) / 60
			dramas[i].Episodes[j].Duration = durationMinutes
		}
	}

	return dramas, total, nil
}

func (s *DramaService) UpdateDrama(dramaID string, req *UpdateDramaRequest, deviceID string) (*models.Drama, error) {
	var drama models.Drama
	if err := s.db.Where("id = ? AND device_id = ?", dramaID, deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drama not found")
		}
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if len(req.TargetCountry) > 0 {
		normalizedCountries := normalizeCountryCodes(req.TargetCountry)
		if len(normalizedCountries) > 0 {
			updates["target_country"] = strings.Join(normalizedCountries, ",")
		}
	}
	if req.MaterialComposition != "" {
		updates["material_composition"] = strings.TrimSpace(req.MaterialComposition)
	}
	if req.MarketingSellingPoints != "" {
		updates["marketing_selling_points"] = strings.TrimSpace(req.MarketingSellingPoints)
	}
	if req.Genre != "" {
		updates["genre"] = req.Genre
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	updates["updated_at"] = time.Now()

	if err := s.db.Model(&drama).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to update drama", "error", err)
		return nil, err
	}

	s.log.Infow("Drama updated", "drama_id", dramaID)
	return &drama, nil
}

func (s *DramaService) DeleteDrama(dramaID string, deviceID string) error {
	result := s.db.Where("id = ? AND device_id = ?", dramaID, deviceID).Delete(&models.Drama{})

	if result.Error != nil {
		s.log.Errorw("Failed to delete drama", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("drama not found")
	}

	s.log.Infow("Drama deleted", "drama_id", dramaID)
	return nil
}

func (s *DramaService) GetDramaStats(deviceID string) (map[string]interface{}, error) {
	var total int64
	var byStatus []struct {
		Status string
		Count  int64
	}

	if err := s.db.Model(&models.Drama{}).Where("device_id = ?", deviceID).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.Drama{}).
		Where("device_id = ?", deviceID).
		Select("status, count(*) as count").
		Group("status").
		Scan(&byStatus).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":     total,
		"by_status": byStatus,
	}

	return stats, nil
}

type SaveOutlineRequest struct {
	Title   string   `json:"title" binding:"required"`
	Summary string   `json:"summary" binding:"required"`
	Genre   string   `json:"genre"`
	Tags    []string `json:"tags"`
}

type SaveCharactersRequest struct {
	Characters []models.Character `json:"characters" binding:"required"`
	EpisodeID  *uint              `json:"episode_id"` // 可选：如果提供则关联到指定章节
}

type SaveProgressRequest struct {
	CurrentStep string                 `json:"current_step" binding:"required"`
	StepData    map[string]interface{} `json:"step_data"`
}

type SaveEpisodesRequest struct {
	Episodes []models.Episode `json:"episodes" binding:"required"`
}

func (s *DramaService) SaveOutline(dramaID string, req *SaveOutlineRequest, deviceID string) error {
	var drama models.Drama
	if err := s.db.Where("id = ? AND device_id = ?", dramaID, deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("drama not found")
		}
		return err
	}

	updates := map[string]interface{}{
		"title":       req.Title,
		"description": req.Summary,
		"updated_at":  time.Now(),
	}

	if req.Genre != "" {
		updates["genre"] = req.Genre
	}

	if len(req.Tags) > 0 {
		tagsJSON, err := json.Marshal(req.Tags)
		if err != nil {
			s.log.Errorw("Failed to marshal tags", "error", err)
			return err
		}
		updates["tags"] = tagsJSON
	}

	if err := s.db.Model(&drama).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to save outline", "error", err)
		return err
	}

	s.log.Infow("Outline saved", "drama_id", dramaID)
	return nil
}

func (s *DramaService) GetCharacters(dramaID string, episodeID *string, deviceID string) ([]models.Character, error) {
	var drama models.Drama
	if err := s.db.Where("id = ? AND device_id = ?", dramaID, deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drama not found")
		}
		return nil, err
	}

	var characters []models.Character

	// 如果指定了episodeID，只获取该章节关联的角色
	if episodeID != nil {
		var episode models.Episode
		if err := s.db.Preload("Characters").Where("id = ? AND drama_id = ?", *episodeID, dramaID).First(&episode).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("episode not found")
			}
			return nil, err
		}
		characters = episode.Characters
	} else {
		// 如果没有指定episodeID，获取项目的所有角色
		if err := s.db.Where("drama_id = ?", dramaID).Find(&characters).Error; err != nil {
			s.log.Errorw("Failed to get characters", "error", err)
			return nil, err
		}
	}

	// 查询每个角色的图片生成任务状态
	for i := range characters {
		NormalizeImageURLPtr(characters[i].ImageURL)
		// 查询该角色最新的图片生成任务
		var imageGen models.ImageGeneration
		err := s.db.Where("character_id = ?", characters[i].ID).
			Order("created_at DESC").
			First(&imageGen).Error

		if err == nil {
			// 如果有进行中的任务，填充状态信息
			if imageGen.Status == models.ImageStatusPending || imageGen.Status == models.ImageStatusProcessing {
				statusStr := string(imageGen.Status)
				characters[i].ImageGenerationStatus = &statusStr
			} else if imageGen.Status == models.ImageStatusFailed {
				statusStr := "failed"
				characters[i].ImageGenerationStatus = &statusStr
				if imageGen.ErrorMsg != nil {
					characters[i].ImageGenerationError = imageGen.ErrorMsg
				}
			}
		}
	}

	return characters, nil
}

func (s *DramaService) SaveCharacters(dramaID string, req *SaveCharactersRequest, deviceID string) error {
	// 转换dramaID
	id, err := strconv.ParseUint(dramaID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid drama ID")
	}
	dramaIDUint := uint(id)

	var drama models.Drama
	if err := s.db.Where("id = ? AND device_id = ?", dramaIDUint, deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("drama not found")
		}
		return err
	}

	// 如果指定了EpisodeID，验证章节存在性
	if req.EpisodeID != nil {
		var episode models.Episode
		if err := s.db.Where("id = ? AND drama_id = ?", *req.EpisodeID, dramaIDUint).First(&episode).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("episode not found")
			}
			return err
		}
	}

	// 获取该项目已存在的所有角色
	var existingCharacters []models.Character
	if err := s.db.Where("drama_id = ?", dramaIDUint).Find(&existingCharacters).Error; err != nil {
		s.log.Errorw("Failed to get existing characters", "error", err)
		return err
	}

	// 创建角色名称到角色的映射
	existingCharMap := make(map[string]*models.Character)
	for i := range existingCharacters {
		existingCharMap[existingCharacters[i].Name] = &existingCharacters[i]
	}

	// 收集需要关联到章节的角色ID
	var characterIDs []uint

	// 创建新角色或复用已有角色
	for _, char := range req.Characters {
		if existingChar, exists := existingCharMap[char.Name]; exists {
			// 角色已存在，直接复用
			s.log.Infow("Character already exists, reusing", "name", char.Name, "character_id", existingChar.ID)
			characterIDs = append(characterIDs, existingChar.ID)
			continue
		}

		// 角色不存在，创建新角色
		character := models.Character{
			DramaID:     dramaIDUint,
			Name:        char.Name,
			Role:        char.Role,
			Description: char.Description,
			Personality: char.Personality,
			Appearance:  char.Appearance,
		}

		if err := s.db.Create(&character).Error; err != nil {
			s.log.Errorw("Failed to create character", "error", err, "name", char.Name)
			continue
		}

		s.log.Infow("New character created", "character_id", character.ID, "name", char.Name)
		characterIDs = append(characterIDs, character.ID)
	}

	// 如果指定了EpisodeID，建立角色与章节的关联
	if req.EpisodeID != nil && len(characterIDs) > 0 {
		var episode models.Episode
		if err := s.db.First(&episode, *req.EpisodeID).Error; err != nil {
			return err
		}

		// 获取角色对象
		var characters []models.Character
		if err := s.db.Where("id IN ?", characterIDs).Find(&characters).Error; err != nil {
			s.log.Errorw("Failed to get characters", "error", err)
			return err
		}

		// 使用GORM的Association API建立多对多关系（会自动去重）
		if err := s.db.Model(&episode).Association("Characters").Append(&characters); err != nil {
			s.log.Errorw("Failed to associate characters with episode", "error", err)
			return err
		}

		s.log.Infow("Characters associated with episode", "episode_id", *req.EpisodeID, "character_count", len(characterIDs))
	}

	if err := s.db.Model(&drama).Update("updated_at", time.Now()).Error; err != nil {
		s.log.Errorw("Failed to update drama timestamp", "error", err)
	}

	s.log.Infow("Characters saved", "drama_id", dramaID, "count", len(req.Characters))
	return nil
}

func (s *DramaService) SaveEpisodes(dramaID string, req *SaveEpisodesRequest, deviceID string) error {
	// 转换dramaID
	id, err := strconv.ParseUint(dramaID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid drama ID")
	}
	dramaIDUint := uint(id)

	var drama models.Drama
	if err := s.db.Where("id = ? AND device_id = ?", dramaIDUint, deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("drama not found")
		}
		return err
	}

	// 删除旧剧集
	if err := s.db.Where("drama_id = ?", dramaIDUint).Delete(&models.Episode{}).Error; err != nil {
		s.log.Errorw("Failed to delete old episodes", "error", err)
		return err
	}

	// 创建新剧集（不包含场景，场景由后续步骤生成）
	for _, ep := range req.Episodes {
		episode := models.Episode{
			DramaID:       dramaIDUint,
			EpisodeNum:    ep.EpisodeNum,
			Title:         ep.Title,
			Description:   ep.Description,
			ScriptContent: ep.ScriptContent,
			Duration:      ep.Duration,
			Status:        "draft",
		}

		if err := s.db.Create(&episode).Error; err != nil {
			s.log.Errorw("Failed to create episode", "error", err, "episode", ep.EpisodeNum)
			continue
		}
	}

	if err := s.db.Model(&drama).Update("updated_at", time.Now()).Error; err != nil {
		s.log.Errorw("Failed to update drama timestamp", "error", err)
	}

	s.log.Infow("Episodes saved", "drama_id", dramaID, "count", len(req.Episodes))
	return nil
}

func (s *DramaService) SaveProgress(dramaID string, req *SaveProgressRequest, deviceID string) error {
	var drama models.Drama
	if err := s.db.Where("id = ? AND device_id = ?", dramaID, deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("drama not found")
		}
		return err
	}

	// 构建metadata对象
	metadata := make(map[string]interface{})

	// 保留现有metadata
	if drama.Metadata != nil {
		if err := json.Unmarshal(drama.Metadata, &metadata); err != nil {
			s.log.Warnw("Failed to unmarshal existing metadata", "error", err)
		}
	}

	// 更新progress信息
	metadata["current_step"] = req.CurrentStep
	if req.StepData != nil {
		metadata["step_data"] = req.StepData
	}

	// 序列化metadata
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		s.log.Errorw("Failed to marshal metadata", "error", err)
		return err
	}

	updates := map[string]interface{}{
		"metadata":   metadataJSON,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&drama).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to save progress", "error", err)
		return err
	}

	s.log.Infow("Progress saved", "drama_id", dramaID, "step", req.CurrentStep)
	return nil
}
