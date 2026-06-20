package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/database"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	demoDeviceID  = "demo-device-digital-silk-road"
	demoProject   = "[DEMO] 智能温显保温杯出海营销"
	demoRequestID = "demo-seed-smart-bottle-v1"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fatal("加载配置失败", err)
	}

	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		fatal("连接数据库失败", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		fatal("自动迁移失败", err)
	}

	deviceID := strings.TrimSpace(os.Getenv("DEMO_DEVICE_ID"))
	if deviceID == "" {
		deviceID = demoDeviceID
	}
	if len(deviceID) < 16 {
		fatal("DEMO_DEVICE_ID 至少需要 16 个字符", nil)
	}

	projectID, err := seed(db, cfg, deviceID)
	if err != nil {
		fatal("写入 Demo 数据失败", err)
	}

	fmt.Println("Demo seed completed.")
	fmt.Printf("DEMO_DEVICE_ID=%s\n", deviceID)
	fmt.Printf("PROJECT_ID=%d\n", projectID)
	fmt.Printf("PROJECT_PATH=/projects/%d\n", projectID)
	var session models.SilkroadAgentSession
	if err := db.Where("device_id = ? AND request_id = ?", deviceID, demoRequestID).First(&session).Error; err == nil {
		fmt.Printf("AGENT_HISTORY_ID=%d\n", session.ID)
		fmt.Printf("AGENT_HISTORY_PATH=/agent/result?historyId=%d\n", session.ID)
	}
	fmt.Println("注意：这些记录均标记为 DEMO，不代表真实广告、订单或第三方生成结果。")
}

func seed(db *gorm.DB, cfg *config.Config, deviceID string) (uint, error) {
	now := time.Now()
	description := "304 不锈钢智能温显保温杯，适合通勤、办公室和礼赠场景。杯盖可显示水温，避免直接饮用过热饮品。"
	material := "304 不锈钢内胆、食品接触级密封圈、LED 温度显示杯盖"
	sellingPoints := "触碰显示温度；便携防漏；通勤与礼赠双场景；简洁中性外观"
	genre := "跨境商品短视频"
	tags := mustJSON([]string{"DEMO", "智能保温杯", "Amazon", "TikTok Shop"})
	compliance := mustJSON(map[string]interface{}{
		"score":                     88,
		"level":                     "green",
		"level_label":               "低风险",
		"summary":                   "Demo 合规结果：避免将温度显示描述为医疗或安全保证，保留材质与检测报告复核提示。",
		"non_compliance_points":     []string{},
		"rectification_suggestions": []string{"使用 helps monitor beverage temperature，避免 absolute safety guarantee。", "上架前复核食品接触材料检测报告和平台类目要求。"},
		"suggested_categories":      []string{"Home & Kitchen", "Insulated Beverage Containers"},
		"disclaimer":                "本结果仅用于跨境电商营销合规辅助，不构成法律意见；实际上架与投放前建议结合目标国家法规、平台政策和专业合规意见进行复核。",
	})
	metadata := mustJSON(map[string]interface{}{
		"demo_seed":       true,
		"demo_case":       "smart-temperature-bottle",
		"target_platform": "Amazon / TikTok Shop",
		"data_boundary":   "平台内演示数据，不含真实广告、订单和销售额",
	})

	var project models.Drama
	err := db.Where("device_id = ? AND title = ?", deviceID, demoProject).First(&project).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	project.DeviceID = deviceID
	project.Title = demoProject
	project.Description = &description
	project.TargetCountry = "美国,英国"
	project.MaterialComposition = &material
	project.MarketingSellingPoints = &sellingPoints
	project.ComplianceScore = 88
	project.ComplianceLevel = "green"
	project.ComplianceReport = compliance
	project.Genre = &genre
	project.Style = "clean_product_demo"
	project.TotalEpisodes = 1
	project.TotalDuration = 30
	project.Status = "completed"
	project.Tags = tags
	project.Metadata = metadata
	project.CreatedAt = now
	project.UpdatedAt = now
	if project.ID == 0 {
		if err := db.Create(&project).Error; err != nil {
			return 0, err
		}
	} else if err := db.Save(&project).Error; err != nil {
		return 0, err
	}

	script := "0-3 秒：手指轻触杯盖，温度数字亮起。\n3-18 秒：展示通勤包侧袋、防漏杯盖和办公室饮用场景。\n18-27 秒：强调 304 不锈钢内胆与简洁礼赠包装。\n27-30 秒：字幕引导查看商品详情，避免绝对安全承诺。"
	episodeDescription := "面向美国通勤人群的 30 秒智能温显保温杯营销内容。"
	episode := models.Episode{}
	if err := db.Where("drama_id = ? AND episode_number = ?", project.ID, 1).First(&episode).Error; err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	episode.DramaID = project.ID
	episode.EpisodeNum = 1
	episode.Title = "温度一触即见：通勤杯 30 秒讲解"
	episode.ScriptContent = &script
	episode.Description = &episodeDescription
	episode.Duration = 30
	episode.Status = "completed"
	episode.CreatedAt = now
	episode.UpdatedAt = now
	if episode.ID == 0 {
		if err := db.Create(&episode).Error; err != nil {
			return 0, err
		}
	} else if err := db.Save(&episode).Error; err != nil {
		return 0, err
	}

	scenePrompt := "明亮现代办公室桌面，智能温显保温杯居中，手指触碰杯盖显示温度，产品结构清晰。"
	scene := models.Scene{}
	if err := db.Where("drama_id = ? AND episode_id = ? AND location = ?", project.ID, episode.ID, "现代办公室桌面").First(&scene).Error; err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	scene.DramaID = project.ID
	scene.EpisodeID = &episode.ID
	scene.Location = "现代办公室桌面"
	scene.Time = "日间"
	scene.Prompt = scenePrompt
	scene.StoryboardCount = 3
	scene.Status = "generated"
	scene.CreatedAt = now
	scene.UpdatedAt = now
	if scene.ID == 0 {
		if err := db.Create(&scene).Error; err != nil {
			return 0, err
		}
	} else if err := db.Save(&scene).Error; err != nil {
		return 0, err
	}

	assetURL, localPath, err := writeDemoAsset(cfg)
	if err != nil {
		return 0, err
	}
	project.Thumbnail = &assetURL
	if err := db.Model(&project).Updates(map[string]interface{}{"thumbnail": assetURL, "updated_at": now}).Error; err != nil {
		return 0, err
	}
	if err := db.Model(&scene).Updates(map[string]interface{}{"image_url": assetURL, "status": "generated", "updated_at": now}).Error; err != nil {
		return 0, err
	}

	role := "main"
	characterDescription := "亲和、可信的英文商品讲解数字人形象建议；Demo seed 不调用数字人供应商。"
	appearance := "简洁商务休闲风，适合美区通勤与家居用品讲解"
	personality := "清晰、克制、可信"
	character := models.Character{}
	if err := db.Where("drama_id = ? AND name = ?", project.ID, "Mia · 商品讲解员").First(&character).Error; err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	character.DramaID = project.ID
	character.Name = "Mia · 商品讲解员"
	character.Role = &role
	character.Description = &characterDescription
	character.Appearance = &appearance
	character.Personality = &personality
	character.SortOrder = 1
	character.CreatedAt = now
	character.UpdatedAt = now
	if character.ID == 0 {
		if err := db.Create(&character).Error; err != nil {
			return 0, err
		}
	} else if err := db.Save(&character).Error; err != nil {
		return 0, err
	}

	completedAt := now
	image := models.ImageGeneration{}
	if err := db.Where("drama_id = ? AND provider = ? AND image_type = ?", project.ID, "demo-seed", string(models.ImageTypeScene)).First(&image).Error; err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	image.DramaID = project.ID
	image.SceneID = &scene.ID
	image.ImageType = string(models.ImageTypeScene)
	image.Provider = "demo-seed"
	image.Prompt = scenePrompt
	image.Model = "local-svg-demo-asset"
	image.Size = "1200x675"
	image.Quality = "demo"
	image.ImageURL = &assetURL
	image.LocalPath = &localPath
	image.Status = models.ImageStatusCompleted
	image.CompletedAt = &completedAt
	image.CreatedAt = now
	image.UpdatedAt = now
	if image.ID == 0 {
		if err := db.Create(&image).Error; err != nil {
			return 0, err
		}
	} else if err := db.Save(&image).Error; err != nil {
		return 0, err
	}

	assetDescription := "Demo seed 生成的本地 SVG 商品占位素材，不代表第三方 AI 生成结果。"
	category := "demo-product"
	mimeType := "image/svg+xml"
	format := "svg"
	width, height := 1200, 675
	asset := models.Asset{}
	if err := db.Where("drama_id = ? AND name = ?", project.ID, "[DEMO] 智能温显保温杯主视觉").First(&asset).Error; err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	asset.DramaID = &project.ID
	asset.Name = "[DEMO] 智能温显保温杯主视觉"
	asset.Description = &assetDescription
	asset.Type = models.AssetTypeImage
	asset.Category = &category
	asset.URL = assetURL
	asset.ThumbnailURL = &assetURL
	asset.LocalPath = &localPath
	asset.MimeType = &mimeType
	asset.Format = &format
	asset.Width = &width
	asset.Height = &height
	asset.ImageGenID = &image.ID
	asset.CreatedAt = now
	asset.UpdatedAt = now
	if asset.ID == 0 {
		if err := db.Create(&asset).Error; err != nil {
			return 0, err
		}
	} else if err := db.Save(&asset).Error; err != nil {
		return 0, err
	}

	taskID := "demo-digital-human-task-0001"
	task := models.AsyncTask{}
	if err := db.Where("id = ?", taskID).First(&task).Error; err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	task.ID = taskID
	task.DeviceID = deviceID
	task.Type = "digital_human_generation"
	task.Status = "failed"
	task.Progress = 0
	task.Message = "Demo seed 未调用第三方数字人服务"
	task.Error = "未配置或未调用数字人供应商；配置火山引擎 Key 后请从项目页创建真实任务。"
	task.ResourceID = strconv.FormatUint(uint64(project.ID), 10)
	task.CreatedAt = now
	task.UpdatedAt = now
	task.CompletedAt = &completedAt
	if task.ID == "" {
		task.ID = uuid.NewString()
	}
	if err := db.Save(&task).Error; err != nil {
		return 0, err
	}

	if err := seedAgentHistory(db, deviceID, now); err != nil {
		return 0, err
	}

	return project.ID, nil
}

func seedAgentHistory(db *gorm.DB, deviceID string, now time.Time) error {
	input := services.SilkroadAgentInput{
		RequestID:         demoRequestID,
		ProductName:       "智能温显保温杯",
		Category:          "家居用品",
		TargetMarket:      "美国",
		TargetPlatform:    "Amazon / TikTok Shop",
		TargetAudience:    "22-40 岁通勤人群与礼赠买家",
		CoreSellingPoints: []string{"触碰显示温度", "便携防漏", "304 不锈钢内胆"},
		MaterialSpec:      "304 不锈钢内胆、食品接触级密封圈",
		UsageScenario:     "通勤、办公室、节日礼赠",
		RawPrompt:         "Demo seed：为智能温显保温杯生成美国市场营销方案。",
	}
	result := services.SilkroadAgentResult{
		RecognizedInfo: services.RecognizedInfo{
			ProductName:        input.ProductName,
			Category:           input.Category,
			TargetMarket:       input.TargetMarket,
			TargetPlatform:     input.TargetPlatform,
			TargetAudience:     input.TargetAudience,
			CoreSellingPoints:  input.CoreSellingPoints,
			ImageUnderstanding: "Demo seed 未调用视觉模型；商品信息来自固定演示案例。",
		},
		Overview: services.AgentOverview{
			ComplianceRiskLevel:     "低风险",
			MarketStrategy:          "突出通勤场景、温度可视化和礼赠属性，避免绝对安全承诺。",
			RecommendedVideoStyle:   "明亮产品演示 + 生活方式切片",
			RecommendedDigitalHuman: "亲和型英文商品讲解员",
		},
		Compliance: services.CompliancePlan{
			Title:       "美国市场营销合规辅助",
			Summary:     "避免医疗、安全保证和未经验证的保温时长承诺。",
			Suggestions: []string{"保留材质检测报告复核提示", "使用 helps monitor temperature 等克制表达"},
			Level:       "green",
			Score:       88,
			Disclaimer:  "本结果仅用于跨境电商营销合规辅助，不构成法律意见。",
		},
		Localization: services.Localization{
			Direction:        "美国通勤与礼赠场景",
			Reason:           "温度可视化适合短视频前三秒展示，杯型适合办公室和礼赠场景。",
			Keywords:         []string{"commute", "temperature display", "gift-ready"},
			Tone:             "clear, practical, trustworthy",
			SceneSuggestions: []string{"办公室桌面", "通勤包侧袋", "礼赠开箱"},
		},
		Script: services.VideoScript{
			Title:    "Tap. Check. Sip with confidence.",
			Duration: "30s",
			Opening:  services.ScriptSegment{Time: "0-3s", Content: "Tap the lid to reveal the temperature."},
			Middle:   services.ScriptSegment{Time: "3-24s", Content: "Show leak-resistant commuting and office use."},
			Ending:   services.ScriptSegment{Time: "24-30s", Content: "Explore details before purchase."},
		},
		DigitalHuman: services.DigitalHuman{
			Persona:        "亲和型英文商品讲解员",
			Tone:           "清晰、克制、可信",
			VideoRatio:     "9:16",
			SubtitleAdvice: "英文主字幕，关键词高亮",
			VisualStyle:    "明亮办公室产品演示",
			ShootingStyle:  "中景口播与商品特写交替",
		},
		Promotion: services.PromotionPlan{
			Platforms:          []string{"Amazon", "TikTok Shop"},
			ContentTags:        []string{"#CommuteEssentials", "#SmartBottle"},
			FocusMetrics:       []string{"平台内素材完成率", "内容发布状态"},
			OptimizationAdvice: "真实广告花费、订单和转化率需接入平台 API 后统计。",
		},
		AgentMessage: services.AgentMessage{
			Summary:      "这是明确标识的 Demo seed 方案，用于稳定演示页面与业务数据流。",
			QuickActions: []string{"查看营销项目", "进入合规分析", "检查数字人任务"},
		},
		IsMock: true,
		Model:  "demo-seed-no-external-ai",
	}

	inputJSON, _ := json.Marshal(input)
	resultJSON, _ := json.Marshal(result)
	session := models.SilkroadAgentSession{}
	if err := db.Where("device_id = ? AND request_id = ?", deviceID, demoRequestID).First(&session).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	session.DeviceID = deviceID
	session.RequestID = demoRequestID
	session.ProductName = input.ProductName
	session.Category = input.Category
	session.TargetMarket = input.TargetMarket
	session.TargetPlatform = input.TargetPlatform
	session.TargetAudience = input.TargetAudience
	session.RawPrompt = input.RawPrompt
	session.InputSnapshot = datatypes.JSON(inputJSON)
	session.ResultSnapshot = datatypes.JSON(resultJSON)
	session.Status = "completed"
	session.Model = result.Model
	session.CreatedAt = now
	session.UpdatedAt = now
	if session.ID == 0 {
		return db.Create(&session).Error
	}
	return db.Save(&session).Error
}

func writeDemoAsset(cfg *config.Config) (string, string, error) {
	localPath := filepath.Join(cfg.Storage.LocalPath, "demo", "smart-temperature-bottle.svg")
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return "", "", err
	}
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="675" viewBox="0 0 1200 675">
  <rect width="1200" height="675" fill="#f5f7fb"/>
  <rect x="80" y="72" width="1040" height="531" rx="18" fill="#ffffff" stroke="#d9e0ea"/>
  <rect x="506" y="170" width="188" height="350" rx="70" fill="#dbe4ea" stroke="#202a35" stroke-width="8"/>
  <rect x="526" y="145" width="148" height="88" rx="32" fill="#1e2936"/>
  <text x="600" y="202" text-anchor="middle" font-family="Arial" font-size="40" fill="#4ee5c2">56°C</text>
  <text x="138" y="160" font-family="Arial" font-size="28" font-weight="700" fill="#172033">DEMO ASSET</text>
  <text x="138" y="210" font-family="Arial" font-size="44" font-weight="700" fill="#172033">Smart Temperature Bottle</text>
  <text x="138" y="260" font-family="Arial" font-size="24" fill="#536173">Local seed visual · not an external AI generation result</text>
  <text x="138" y="520" font-family="Arial" font-size="22" fill="#536173">Digital Silk Road · Competition Demo</text>
</svg>`
	if err := os.WriteFile(localPath, []byte(svg), 0o644); err != nil {
		return "", "", err
	}
	url := strings.TrimSuffix(cfg.Storage.BaseURL, "/") + "/demo/smart-temperature-bottle.svg"
	return url, localPath, nil
}

func mustJSON(value interface{}) datatypes.JSON {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return datatypes.JSON(data)
}

func fatal(message string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
	} else {
		fmt.Fprintln(os.Stderr, message)
	}
	os.Exit(1)
}
