/**
 * 模块说明：项目创建与合规预检服务。
 * 业务场景：数字丝路商品录入完成后，需要先完成目标市场合规校验，再允许生成工作区项目记录。
 * 核心职责：本文件仍包含旧短剧项目服务；本次注释仅覆盖商品创建、合规缓存与 compliance_token 校验链路。
 */
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

func firstDramaDeviceID(deviceIDs []string) string {
	if len(deviceIDs) == 0 {
		return ""
	}
	return strings.TrimSpace(deviceIDs[0])
}

func scopeDramaByDevice(query *gorm.DB, deviceID string) *gorm.DB {
	if deviceID == "" {
		return query
	}
	return query.Where("device_id = ?", deviceID)
}

func (s *DramaService) claimLegacyDramas(deviceID string) error {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return nil
	}

	var total int64
	if err := s.db.Model(&models.Drama{}).Where("device_id = ?", deviceID).Count(&total).Error; err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	return s.db.Model(&models.Drama{}).Where("device_id = ''").Update("device_id", deviceID).Error
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

/**
 * 功能：把前端商品录入请求归一化为合规与创建共用的业务输入。
 * 参数：req 为商品标题、描述、目标市场、材质和卖点等录入字段。
 * 返回：去空格、目标国家规范化后的输入；目标市场缺失时返回业务错误。
 */
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

/**
 * 功能：复制合规结果，避免缓存对象被调用方意外修改。
 * 参数：result 为模型或规则引擎产生的合规评分、风险点和建议。
 * 返回：字段值相同但切片独立的新结果。
 */
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

/**
 * 功能：为一次商品合规预检生成内容指纹。
 * 参数：input 为商品业务字段，deviceID 为前端设备标识。
 * 返回：用于缓存和 token 绑定的哈希 key。
 */
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

	// 将设备 ID 纳入指纹，避免不同用户在同一商品文本下复用彼此的预检 token。
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

/**
 * 功能：生成短期合规预检 token。
 * 参数：无。
 * 返回：随机 token；系统熵不可用时用时间戳哈希兜底，保证流程不中断。
 */
func generateComplianceToken() string {
	randomBytes := make([]byte, 18)
	if _, err := rand.Read(randomBytes); err != nil {
		fallback := sha256.Sum256([]byte(fmt.Sprintf("compliance-%d", time.Now().UnixNano())))
		return hex.EncodeToString(fallback[:16])
	}
	return hex.EncodeToString(randomBytes)
}

/**
 * 功能：读取同一商品内容的短期合规缓存。
 * 参数：cacheKey 为商品字段和设备 ID 生成的内容指纹。
 * 返回：未过期的合规结果副本；不存在或过期时返回 nil。
 */
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
		// 过期数据即时清理，避免用户修改商品后仍看到旧市场风险判断。
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

/**
 * 功能：写入短期合规缓存。
 * 参数：cacheKey 为商品内容指纹，result 为合规模型或规则兜底的评估结果。
 * 返回：无。
 */
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

/**
 * 功能：签发合规预检通过后的短期 token。
 * 参数：cacheKey 绑定商品内容，deviceID 绑定当前设备，result 用于后续创建时复用。
 * 返回：创建接口可携带的 compliance_token。
 */
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

/**
 * 功能：校验创建请求携带的合规 token，并取回对应结果。
 * 参数：token 来自预检接口，input 是当前创建内容，deviceID 是前端设备标识。
 * 返回：与当前内容完全匹配的合规结果；token 过期、串设备或内容变化时返回错误。
 */
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
		// token 只作为“同内容已预检”的短期凭证，过期后必须重新评估目标市场风险。
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
		// 设备维度校验让前端无需登录也能防止不同浏览器互相复用预检结果。
		return nil, ErrCompliancePrecheckInvalid
	}

	expectedCacheKey := buildComplianceCacheKey(input, normalizedDeviceID)
	if entry.cacheKey != expectedCacheKey {
		// 商品标题、目标市场、材质或卖点变化后，原合规判断不能直接用于创建。
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

/**
 * 功能：执行或复用一次商品合规评估。
 * 参数：input 为归一化商品录入信息，deviceID 用于隔离缓存。
 * 返回：合规评分、风险等级、整改建议和推荐品类；模型不可用时返回可继续展示的默认结果。
 */
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
		// 合规服务根据商品信息和目标市场调用 DeepSeek 兼容模型；失败时不阻断预检页展示。
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

/**
 * 功能：供前端合规预检接口调用，返回结果和后续创建凭证。
 * 参数：req 为商品录入内容，deviceIDs 为请求头中的匿名设备标识。
 * 返回：合规结果、compliance_token 和可能的输入校验错误。
 */
func (s *DramaService) EvaluateCompliance(req *CreateDramaRequest, deviceIDs ...string) (*ComplianceResult, string, error) {
	input, err := prepareCreateDramaInput(req)
	if err != nil {
		return nil, "", err
	}

	deviceID := firstDramaDeviceID(deviceIDs)
	complianceResult := s.evaluateCompliance(input, deviceID)
	cacheKey := buildComplianceCacheKey(input, deviceID)
	complianceToken := s.issueComplianceToken(cacheKey, deviceID, complianceResult)
	return complianceResult, complianceToken, nil
}

/**
 * 功能：创建数字丝路商品项目，并把合规结论固化到项目记录。
 * 参数：req 为商品创建请求，可携带 compliance_token；deviceIDs 用于校验同设备预检。
 * 返回：创建后的项目、合规结果；红色风险或 token 失效时返回业务错误。
 */
func (s *DramaService) CreateDrama(req *CreateDramaRequest, deviceIDs ...string) (*models.Drama, *ComplianceResult, error) {
	deviceID := firstDramaDeviceID(deviceIDs)
	input, err := prepareCreateDramaInput(req)
	if err != nil {
		return nil, nil, err
	}

	var complianceResult *ComplianceResult
	if strings.TrimSpace(req.ComplianceToken) != "" {
		// 前端先预检再创建时，复用同内容合规结果，避免用户确认后再次等待模型调用。
		complianceResult, err = s.getComplianceResultByToken(req.ComplianceToken, input, deviceID)
		if err != nil {
			return nil, nil, err
		}
	} else {
		complianceResult = s.evaluateCompliance(input, deviceID)
	}

	if complianceResult.Level == ComplianceRiskRed {
		// 红色风险代表目标市场准入或广告表达风险过高，后端必须拦截创建而不能只依赖前端按钮状态。
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
	if tags := strings.TrimSpace(req.Tags); tags != "" && json.Valid([]byte(tags)) {
		// New project-only fields are persisted as JSON metadata while the legacy
		// table remains unchanged during the compatibility period.
		drama.Metadata = datatypes.JSON([]byte(tags))
	}

	if err := s.db.Create(drama).Error; err != nil {
		s.log.Errorw("Failed to create drama", "error", err)
		return nil, nil, err
	}

	s.log.Infow("Drama created", "drama_id", drama.ID, "compliance_score", complianceResult.Score, "risk_level", complianceResult.Level)
	return drama, complianceResult, nil
}

func (s *DramaService) GetDrama(dramaID string, deviceIDs ...string) (*models.Drama, error) {
	deviceID := firstDramaDeviceID(deviceIDs)
	var drama models.Drama
	err := scopeDramaByDevice(s.db.Where("id = ?", dramaID), deviceID).
		Preload("Characters").          // 加载Drama级别的角色
		Preload("Scenes").              // 加载Drama级别的场景
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

	// 统计每个剧集的时长（基于场景时长之和）
	for i := range drama.Episodes {
		totalDuration := 0
		for _, scene := range drama.Episodes[i].Storyboards {
			totalDuration += scene.Duration
		}
		// 更新剧集时长（秒转分钟，向上取整）
		durationMinutes := (totalDuration + 59) / 60
		drama.Episodes[i].Duration = durationMinutes

		// 如果数据库中的时长与计算的不一致，更新数据库
		if drama.Episodes[i].Duration != durationMinutes {
			s.db.Model(&models.Episode{}).Where("id = ?", drama.Episodes[i].ID).Update("duration", durationMinutes)
		}

		// 查询角色的图片生成状态
		for j := range drama.Episodes[i].Characters {
			var imageGen models.ImageGeneration
			err := s.db.Where("character_id = ? AND (status = ? OR status = ?)",
				drama.Episodes[i].Characters[j].ID, "pending", "processing").
				Order("created_at DESC").
				First(&imageGen).Error

			if err == nil {
				// 找到生成中的记录，设置状态
				statusStr := string(imageGen.Status)
				drama.Episodes[i].Characters[j].ImageGenerationStatus = &statusStr
				if imageGen.ErrorMsg != nil {
					drama.Episodes[i].Characters[j].ImageGenerationError = imageGen.ErrorMsg
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 检查是否有失败的记录
				err := s.db.Where("character_id = ? AND status = ?",
					drama.Episodes[i].Characters[j].ID, "failed").
					Order("created_at DESC").
					First(&imageGen).Error

				if err == nil {
					statusStr := string(imageGen.Status)
					drama.Episodes[i].Characters[j].ImageGenerationStatus = &statusStr
					if imageGen.ErrorMsg != nil {
						drama.Episodes[i].Characters[j].ImageGenerationError = imageGen.ErrorMsg
					}
				}
			}
		}

		// 查询场景的图片生成状态
		for j := range drama.Episodes[i].Scenes {
			var imageGen models.ImageGeneration
			err := s.db.Where("scene_id = ? AND (status = ? OR status = ?)",
				drama.Episodes[i].Scenes[j].ID, "pending", "processing").
				Order("created_at DESC").
				First(&imageGen).Error

			if err == nil {
				// 找到生成中的记录，设置状态
				statusStr := string(imageGen.Status)
				drama.Episodes[i].Scenes[j].ImageGenerationStatus = &statusStr
				if imageGen.ErrorMsg != nil {
					drama.Episodes[i].Scenes[j].ImageGenerationError = imageGen.ErrorMsg
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 检查是否有失败的记录
				err := s.db.Where("scene_id = ? AND status = ?",
					drama.Episodes[i].Scenes[j].ID, "failed").
					Order("created_at DESC").
					First(&imageGen).Error

				if err == nil {
					statusStr := string(imageGen.Status)
					drama.Episodes[i].Scenes[j].ImageGenerationStatus = &statusStr
					if imageGen.ErrorMsg != nil {
						drama.Episodes[i].Scenes[j].ImageGenerationError = imageGen.ErrorMsg
					}
				}
			}
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

func (s *DramaService) ListDramas(query *DramaListQuery, deviceIDs ...string) ([]models.Drama, int64, error) {
	deviceID := firstDramaDeviceID(deviceIDs)
	var dramas []models.Drama
	var total int64

	if err := s.claimLegacyDramas(deviceID); err != nil {
		s.log.Warnw("Failed to claim legacy dramas", "error", err, "device_id", deviceID)
	}

	db := scopeDramaByDevice(s.db.Model(&models.Drama{}), deviceID)

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

func (s *DramaService) UpdateDrama(dramaID string, req *UpdateDramaRequest, deviceIDs ...string) (*models.Drama, error) {
	deviceID := firstDramaDeviceID(deviceIDs)
	var drama models.Drama
	if err := scopeDramaByDevice(s.db.Where("id = ?", dramaID), deviceID).First(&drama).Error; err != nil {
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
		if json.Valid([]byte(req.Tags)) {
			updates["metadata"] = datatypes.JSON([]byte(req.Tags))
		}
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

func (s *DramaService) DeleteDrama(dramaID string, deviceIDs ...string) error {
	deviceID := firstDramaDeviceID(deviceIDs)
	result := scopeDramaByDevice(s.db.Where("id = ?", dramaID), deviceID).Delete(&models.Drama{})

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

func (s *DramaService) GetDramaStats(deviceIDs ...string) (map[string]interface{}, error) {
	deviceID := firstDramaDeviceID(deviceIDs)
	var total int64
	var byStatus []struct {
		Status string
		Count  int64
	}

	if err := s.claimLegacyDramas(deviceID); err != nil {
		s.log.Warnw("Failed to claim legacy dramas for stats", "error", err, "device_id", deviceID)
	}

	if err := scopeDramaByDevice(s.db.Model(&models.Drama{}), deviceID).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := scopeDramaByDevice(s.db.Model(&models.Drama{}), deviceID).
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

func (s *DramaService) SaveOutline(dramaID string, req *SaveOutlineRequest, deviceIDs ...string) error {
	deviceID := firstDramaDeviceID(deviceIDs)
	var drama models.Drama
	if err := scopeDramaByDevice(s.db.Where("id = ?", dramaID), deviceID).First(&drama).Error; err != nil {
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

func (s *DramaService) GetCharacters(dramaID string, episodeID *string, deviceIDs ...string) ([]models.Character, error) {
	deviceID := firstDramaDeviceID(deviceIDs)
	var drama models.Drama
	if err := scopeDramaByDevice(s.db.Where("id = ?", dramaID), deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("drama not found")
		}
		return nil, err
	}

	var characters []models.Character

	// 如果指定了episodeID，只获取该章节关联的角色
	if episodeID != nil {
		var episode models.Episode
		if err := s.db.Preload("Characters").Where("id = ? AND drama_id = ?", *episodeID, drama.ID).First(&episode).Error; err != nil {
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

func (s *DramaService) SaveCharacters(dramaID string, req *SaveCharactersRequest, deviceIDs ...string) error {
	deviceID := firstDramaDeviceID(deviceIDs)
	// 转换dramaID
	id, err := strconv.ParseUint(dramaID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid drama ID")
	}
	dramaIDUint := uint(id)

	var drama models.Drama
	if err := scopeDramaByDevice(s.db.Where("id = ?", dramaIDUint), deviceID).First(&drama).Error; err != nil {
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
			updates := map[string]interface{}{
				"role":        char.Role,
				"description": char.Description,
				"personality": char.Personality,
				"appearance":  char.Appearance,
				"updated_at":  time.Now(),
			}
			if char.VoiceStyle != nil {
				updates["voice_style"] = char.VoiceStyle
			}
			if char.ImageURL != nil {
				updates["image_url"] = char.ImageURL
			}
			if char.SortOrder > 0 {
				updates["sort_order"] = char.SortOrder
			}
			if err := s.db.Model(existingChar).Updates(updates).Error; err != nil {
				s.log.Errorw("Failed to update existing character", "error", err, "name", char.Name)
				return err
			}
			s.log.Infow("Character already exists, updated and reused", "name", char.Name, "character_id", existingChar.ID)
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

func (s *DramaService) SaveEpisodes(dramaID string, req *SaveEpisodesRequest, deviceIDs ...string) error {
	deviceID := firstDramaDeviceID(deviceIDs)
	// 转换dramaID
	id, err := strconv.ParseUint(dramaID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid drama ID")
	}
	dramaIDUint := uint(id)

	var drama models.Drama
	if err := scopeDramaByDevice(s.db.Where("id = ?", dramaIDUint), deviceID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("drama not found")
		}
		return err
	}

	now := time.Now()
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var existingEpisodes []models.Episode
		if err := tx.Where("drama_id = ?", dramaIDUint).Find(&existingEpisodes).Error; err != nil {
			s.log.Errorw("Failed to load existing episodes", "error", err)
			return err
		}

		existingByNumber := make(map[int]models.Episode, len(existingEpisodes))
		for _, episode := range existingEpisodes {
			existingByNumber[episode.EpisodeNum] = episode
		}

		seenEpisodeNumbers := make(map[int]struct{}, len(req.Episodes))
		for index, ep := range req.Episodes {
			episodeNumber := ep.EpisodeNum
			if episodeNumber <= 0 {
				episodeNumber = index + 1
			}
			if _, exists := seenEpisodeNumbers[episodeNumber]; exists {
				return fmt.Errorf("duplicate episode number: %d", episodeNumber)
			}
			seenEpisodeNumbers[episodeNumber] = struct{}{}

			title := strings.TrimSpace(ep.Title)
			if title == "" {
				title = fmt.Sprintf("第%d集", episodeNumber)
			}

			status := strings.TrimSpace(ep.Status)
			if status == "" {
				status = "draft"
			}

			if existing, exists := existingByNumber[episodeNumber]; exists {
				updates := map[string]interface{}{
					"title":          title,
					"description":    ep.Description,
					"script_content": ep.ScriptContent,
					"duration":       ep.Duration,
					"status":         status,
					"updated_at":     now,
				}
				if err := tx.Model(&models.Episode{}).
					Where("id = ? AND drama_id = ?", existing.ID, dramaIDUint).
					Updates(updates).Error; err != nil {
					s.log.Errorw("Failed to update episode", "error", err, "episode", episodeNumber)
					return err
				}
				continue
			}

			episode := models.Episode{
				DramaID:       dramaIDUint,
				EpisodeNum:    episodeNumber,
				Title:         title,
				Description:   ep.Description,
				ScriptContent: ep.ScriptContent,
				Duration:      ep.Duration,
				Status:        status,
			}

			if err := tx.Create(&episode).Error; err != nil {
				s.log.Errorw("Failed to create episode", "error", err, "episode", episodeNumber)
				return err
			}
		}

		for _, existing := range existingEpisodes {
			if _, keep := seenEpisodeNumbers[existing.EpisodeNum]; keep {
				continue
			}
			if err := tx.Delete(&existing).Error; err != nil {
				s.log.Errorw("Failed to delete removed episode", "error", err, "episode", existing.EpisodeNum)
				return err
			}
		}

		if err := tx.Model(&models.Drama{}).
			Where("id = ?", dramaIDUint).
			Updates(map[string]interface{}{
				"total_episodes": len(req.Episodes),
				"updated_at":     now,
			}).Error; err != nil {
			s.log.Errorw("Failed to update drama episode metadata", "error", err)
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	s.log.Infow("Episodes saved", "drama_id", dramaID, "count", len(req.Episodes))
	return nil
}

func (s *DramaService) SaveProgress(dramaID string, req *SaveProgressRequest, deviceIDs ...string) error {
	deviceID := firstDramaDeviceID(deviceIDs)
	var drama models.Drama
	if err := scopeDramaByDevice(s.db.Where("id = ?", dramaID), deviceID).First(&drama).Error; err != nil {
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
