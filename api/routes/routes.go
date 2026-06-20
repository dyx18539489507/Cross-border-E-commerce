/**
 * 模块说明：后端 HTTP 路由注册。
 * 业务场景：数字丝路前端通过 /api/v1 访问 Agent、商品合规、数字人和分发等能力。
 * 核心职责：组装服务与 handler，注册数字丝路相关接口，同时保留旧业务接口的兼容路由。
 */
package routes

import (
	handlers2 "github.com/drama-generator/backend/api/handlers"
	middlewares2 "github.com/drama-generator/backend/api/middlewares"
	services2 "github.com/drama-generator/backend/application/services"
	storage2 "github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, log *logger.Logger, localStorage interface{}) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middlewares2.LoggerMiddleware(log))
	r.Use(middlewares2.CORSMiddleware(cfg.Server.CORSOrigins))
	r.Use(middlewares2.DeviceIdentityMiddleware())

	// 静态文件服务（用户上传的文件）
	r.Static("/static", cfg.Storage.LocalPath)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"app":     cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	aiService := services2.NewAIService(db, log)
	localStoragePtr := localStorage.(*storage2.LocalStorage)
	transferService := services2.NewResourceTransferService(db, log)
	notificationService := services2.NewNotificationService(db, log)
	workbenchService := services2.NewWorkbenchService(db, log)
	analyticsService := services2.NewAnalyticsService(db, log)
	dramaHandler := handlers2.NewDramaHandler(db, cfg, log, nil, notificationService)
	aiConfigHandler := handlers2.NewAIConfigHandler(db, cfg, log)
	aiAssistHandler := handlers2.NewAIAssistHandler(db, cfg, log)
	scriptGenHandler := handlers2.NewScriptGenerationHandler(db, cfg, log)
	imageGenService := services2.NewImageGenerationService(db, cfg, transferService, localStoragePtr, log)
	imageGenHandler := handlers2.NewImageGenerationHandler(db, cfg, log, transferService, localStoragePtr)
	videoGenHandler := handlers2.NewVideoGenerationHandler(db, transferService, localStoragePtr, aiService, log)
	videoMergeHandler := handlers2.NewVideoMergeHandler(db, nil, cfg.Storage.LocalPath, cfg.Storage.BaseURL, log)
	assetHandler := handlers2.NewAssetHandler(db, cfg, log)
	characterLibraryService := services2.NewCharacterLibraryService(db, log)
	characterLibraryHandler := handlers2.NewCharacterLibraryHandler(db, cfg, log, transferService, localStoragePtr)
	uploadHandler, err := handlers2.NewUploadHandler(cfg, log, characterLibraryService)
	if err != nil {
		log.Fatalw("Failed to create upload handler", "error", err)
	}
	digitalHumanHandler, err := handlers2.NewDigitalHumanHandler(db, cfg, log)
	if err != nil {
		log.Fatalw("Failed to create digital human handler", "error", err)
	}
	voiceLibraryHandler, err := handlers2.NewVoiceLibraryHandler(cfg, db, log)
	if err != nil {
		log.Fatalw("Failed to create voice library handler", "error", err)
	}
	storyboardHandler := handlers2.NewStoryboardHandler(db, cfg, log)
	sceneHandler := handlers2.NewSceneHandler(db, log, imageGenService)
	taskHandler := handlers2.NewTaskHandler(db, log)
	framePromptService := services2.NewFramePromptService(db, cfg, log)
	framePromptHandler := handlers2.NewFramePromptHandler(framePromptService, log)
	audioExtractionHandler := handlers2.NewAudioExtractionHandler(log, cfg.Storage.LocalPath)
	settingsHandler := handlers2.NewSettingsHandler(cfg, log)
	musicHandler := handlers2.NewMusicHandler(log)
	mediaHandler := handlers2.NewMediaHandler(log)
	sfxHandler := handlers2.NewSFXHandler(cfg, log)
	socialBindingHandler := handlers2.NewSocialBindingHandler(db, log)
	distributionService := services2.NewDistributionService(db, cfg, log)
	distributionHandler := handlers2.NewDistributionHandler(distributionService, log)
	silkroadAgentHistoryService := services2.NewSilkroadAgentHistoryService(db, log)
	silkroadAgentProjectService := services2.NewSilkroadAgentProjectService(db, log)
	// 丝路 Agent 独立于旧短剧生成服务，专门处理商品出海分析、追问和结果通知。
	silkroadAgentHandler := handlers2.NewSilkroadAgentHandler(cfg, log, notificationService, silkroadAgentHistoryService, silkroadAgentProjectService)
	notificationHandler := handlers2.NewNotificationHandler(notificationService, log)
	workbenchHandler := handlers2.NewWorkbenchHandler(workbenchService, log)
	analyticsHandler := handlers2.NewAnalyticsHandler(analyticsService, log)

	api := r.Group("/api/v1")
	{
		api.Use(middlewares2.RateLimitMiddleware())

		dramas := api.Group("/dramas")
		{
			dramas.GET("", dramaHandler.ListDramas)
			// 商品录入完成后先走合规预检，返回 compliance_token，再由创建接口校验同一份商品内容。
			dramas.POST("/compliance-check", dramaHandler.CheckCompliance)
			dramas.POST("", dramaHandler.CreateDrama)
			dramas.GET("/stats", dramaHandler.GetDramaStats)
			dramas.GET("/:id/characters", dramaHandler.GetCharacters)
			dramas.PUT("/:id/characters", dramaHandler.SaveCharacters)
			dramas.PUT("/:id/outline", dramaHandler.SaveOutline)
			dramas.PUT("/:id/episodes", dramaHandler.SaveEpisodes)
			dramas.PUT("/:id/progress", dramaHandler.SaveProgress)
			dramas.GET("/:id", dramaHandler.GetDrama)
			dramas.PUT("/:id", dramaHandler.UpdateDrama)
			dramas.DELETE("/:id", dramaHandler.DeleteDrama)
		}

		projects := api.Group("/projects")
		{
			projects.GET("", dramaHandler.ListDramas)
			projects.POST("/compliance-check", dramaHandler.CheckCompliance)
			projects.POST("", dramaHandler.CreateDrama)
			projects.GET("/stats", dramaHandler.GetDramaStats)
			projects.GET("/:id/script", dramaHandler.GetProjectScript)
			projects.PUT("/:id/script", dramaHandler.SaveProjectScript)
			projects.GET("/:id/timeline", dramaHandler.GetProjectTimeline)
			projects.PUT("/:id/timeline", dramaHandler.SaveProjectTimeline)
			projects.POST("/:id/timeline/export", dramaHandler.GetProjectTimeline)
			projects.GET("/:id/assets", dramaHandler.GetProjectAssets)
			projects.GET("/:id/tasks", dramaHandler.GetProjectTasks)
			projects.GET("/:id/characters", dramaHandler.GetCharacters)
			projects.PUT("/:id/characters", dramaHandler.SaveCharacters)
			projects.PUT("/:id/outline", dramaHandler.SaveOutline)
			projects.PUT("/:id/episodes", dramaHandler.SaveEpisodes)
			projects.PUT("/:id/progress", dramaHandler.SaveProgress)
			projects.GET("/:id", dramaHandler.GetDrama)
			projects.PUT("/:id", dramaHandler.UpdateDrama)
			projects.DELETE("/:id", dramaHandler.DeleteDrama)
		}

		aiConfigs := api.Group("/ai-configs")
		{
			aiConfigs.GET("", aiConfigHandler.ListConfigs)
			aiConfigs.POST("", aiConfigHandler.CreateConfig)
			aiConfigs.POST("/test", aiConfigHandler.TestConnection)
			aiConfigs.GET("/:id", aiConfigHandler.GetConfig)
			aiConfigs.PUT("/:id", aiConfigHandler.UpdateConfig)
			aiConfigs.DELETE("/:id", aiConfigHandler.DeleteConfig)
		}

		generation := api.Group("/generation")
		{
			generation.POST("/characters", scriptGenHandler.GenerateCharacters)
			generation.POST("/assist-script", aiAssistHandler.GenerateEpisodeScript)
		}

		agent := api.Group("/agent")
		{
			// /extract 用于轻量识别，/analyze 用于过渡页流式展示，/generate 生成最终结果页 JSON。
			agent.POST("/extract", silkroadAgentHandler.Extract)
			agent.POST("/generate", silkroadAgentHandler.Generate)
			agent.POST("/workflow", silkroadAgentHandler.GenerateWorkflow)
			agent.POST("/generate-workflow", silkroadAgentHandler.GenerateWorkflow)
			agent.POST("/analyze", silkroadAgentHandler.Analyze)
			agent.POST("/follow-up", silkroadAgentHandler.FollowUp)
			agent.GET("/history", silkroadAgentHandler.ListHistory)
			agent.GET("/history/:id", silkroadAgentHandler.GetHistory)
			agent.POST("/create-project", silkroadAgentHandler.CreateProject)
			agent.POST("/:id/create-project", silkroadAgentHandler.CreateProjectFromHistory)
		}

		workbench := api.Group("/workbench")
		{
			workbench.GET("/summary", workbenchHandler.Summary)
		}

		analytics := api.Group("/analytics")
		{
			analytics.GET("/summary", analyticsHandler.Summary)
		}

		notifications := api.Group("/notifications")
		{
			notifications.GET("", notificationHandler.List)
			notifications.POST("", notificationHandler.Create)
			notifications.GET("/unread-count", notificationHandler.UnreadCount)
			notifications.GET("/stream", notificationHandler.Stream)
			notifications.PATCH("/read-all", notificationHandler.MarkAllRead)
			notifications.PATCH("/:id/read", notificationHandler.MarkRead)
			notifications.DELETE("/:id", notificationHandler.Dismiss)
		}

		// 角色库路由
		characterLibrary := api.Group("/character-library")
		{
			characterLibrary.GET("", characterLibraryHandler.ListLibraryItems)
			characterLibrary.POST("", characterLibraryHandler.CreateLibraryItem)
			characterLibrary.GET("/:id", characterLibraryHandler.GetLibraryItem)
			characterLibrary.DELETE("/:id", characterLibraryHandler.DeleteLibraryItem)
		}

		// 角色图片相关路由
		characters := api.Group("/characters")
		{
			characters.PUT("/:id", characterLibraryHandler.UpdateCharacter)
			characters.DELETE("/:id", characterLibraryHandler.DeleteCharacter)
			characters.POST("/batch-generate-images", characterLibraryHandler.BatchGenerateCharacterImages)
			characters.POST("/:id/generate-image", characterLibraryHandler.GenerateCharacterImage)
			characters.POST("/:id/upload-image", uploadHandler.UploadCharacterImage)
			characters.PUT("/:id/image", characterLibraryHandler.UploadCharacterImage)
			characters.PUT("/:id/image-from-library", characterLibraryHandler.ApplyLibraryItemToCharacter)
			characters.POST("/:id/add-to-library", characterLibraryHandler.AddCharacterToLibrary)
		}

		// 文件上传路由
		upload := api.Group("/upload")
		{
			upload.POST("/image", uploadHandler.UploadImage)
			upload.POST("/file", uploadHandler.UploadFile)
		}

		// 数字人生成
		digitalHumans := api.Group("/digital-humans")
		{
			// 数字丝路内容创作阶段使用该接口把角色图和口播音频生成数字人视频素材。
			digitalHumans.GET("", digitalHumanHandler.ListTasks)
			digitalHumans.POST("", digitalHumanHandler.Generate)
			digitalHumans.GET("/history", digitalHumanHandler.ListTasks)
			digitalHumans.GET("/:id", digitalHumanHandler.GetTask)
			digitalHumans.GET("/:id/status", digitalHumanHandler.GetTaskStatus)
			digitalHumans.GET("/:id/result", digitalHumanHandler.GetTaskResult)
			digitalHumans.DELETE("/:id", digitalHumanHandler.DeleteTask)
		}

		// 音色库
		voiceLibrary := api.Group("/voice-library")
		{
			voiceLibrary.GET("", voiceLibraryHandler.List)
			voiceLibrary.POST("/custom", voiceLibraryHandler.CreateCustom)
			voiceLibrary.GET("/custom/:id/status", voiceLibraryHandler.GetCustomStatus)
		}

		// 分镜头路由
		episodes := api.Group("/episodes")
		{
			// 分镜头
			episodes.POST("/:episode_id/storyboards", storyboardHandler.GenerateStoryboard)
			episodes.GET("/:episode_id/storyboards", sceneHandler.GetStoryboardsForEpisode)
			episodes.POST("/:episode_id/finalize", dramaHandler.FinalizeEpisode)
			episodes.GET("/:episode_id/download", dramaHandler.DownloadEpisodeVideo)
		}

		// 任务路由
		tasks := api.Group("/tasks")
		{
			tasks.GET("/:task_id", taskHandler.GetTaskStatus)
			tasks.GET("", taskHandler.GetResourceTasks)
		}

		// 场景路由
		scenes := api.Group("/scenes")
		{
			scenes.PUT("/:scene_id", sceneHandler.UpdateScene)
			scenes.PUT("/:scene_id/prompt", sceneHandler.UpdateScenePrompt)
			scenes.DELETE("/:scene_id", sceneHandler.DeleteScene)
			scenes.POST("/generate-image", sceneHandler.GenerateSceneImage)
		}

		images := api.Group("/images")
		{
			images.GET("", imageGenHandler.ListImageGenerations)
			images.POST("", imageGenHandler.GenerateImage)
			images.POST("/manual", imageGenHandler.CreateImageRecord)
			images.GET("/:id", imageGenHandler.GetImageGeneration)
			images.DELETE("/:id", imageGenHandler.DeleteImageGeneration)
			images.POST("/scene/:scene_id", imageGenHandler.GenerateImagesForScene)
			images.GET("/episode/:episode_id/backgrounds", imageGenHandler.GetBackgroundsForEpisode)
			images.POST("/episode/:episode_id/backgrounds/extract", imageGenHandler.ExtractBackgroundsForEpisode)
			images.POST("/episode/:episode_id/batch", imageGenHandler.BatchGenerateForEpisode)
		}

		videos := api.Group("/videos")
		{
			videos.GET("", videoGenHandler.ListVideoGenerations)
			videos.POST("", videoGenHandler.GenerateVideo)
			videos.GET("/:id", videoGenHandler.GetVideoGeneration)
			videos.DELETE("/:id", videoGenHandler.DeleteVideoGeneration)
			videos.POST("/image/:image_gen_id", videoGenHandler.GenerateVideoFromImage)
			videos.POST("/episode/:episode_id/batch", videoGenHandler.BatchGenerateForEpisode)
		}

		videoMerges := api.Group("/video-merges")
		{
			videoMerges.GET("", videoMergeHandler.ListMerges)
			videoMerges.POST("", videoMergeHandler.MergeVideos)
			videoMerges.GET("/:merge_id", videoMergeHandler.GetMerge)
			videoMerges.GET("/:merge_id/distributions", videoMergeHandler.ListDistributions)
			videoMerges.POST("/:merge_id/distribute", videoMergeHandler.DistributeVideo)
			videoMerges.DELETE("/:merge_id", videoMergeHandler.DeleteMerge)
		}

		distributions := api.Group("/distributions")
		{
			// 分发接口按“账号/目标配置”和“发布任务”分层，前端可以先校验账号状态再提交任务。
			distributions.GET("/targets", distributionHandler.ListTargets)
			distributions.PUT("/targets/:id/default", distributionHandler.SetDefaultTarget)
			distributions.PUT("/targets/reddit/default", distributionHandler.SaveRedditDefaultTarget)
			distributions.POST("/targets/discord", distributionHandler.UpsertDiscordTarget)
			distributions.DELETE("/targets/:id", distributionHandler.DeleteTarget)

			distributions.POST("/upload-post/profile/ensure", distributionHandler.EnsureUploadPostProfile)
			distributions.POST("/upload-post/connect-link", distributionHandler.GenerateUploadPostConnectLink)
			distributions.POST("/upload-post/sync", distributionHandler.SyncUploadPostProfile)
			distributions.GET("/pinterest/boards", distributionHandler.ListPinterestBoards)

			distributions.GET("", distributionHandler.ListDistributionJobs)
			distributions.POST("", distributionHandler.CreateDistribution)
			distributions.GET("/:id", distributionHandler.GetDistributionJob)
			distributions.POST("/:id/retry", distributionHandler.RetryDistribution)
		}

		socialBindings := api.Group("/social-bindings")
		{
			socialBindings.GET("", socialBindingHandler.ListBindings)
			socialBindings.PUT("/:platform", socialBindingHandler.UpsertBinding)
		}

		assets := api.Group("/assets")
		{
			assets.GET("", assetHandler.ListAssets)
			assets.POST("", assetHandler.CreateAsset)
			assets.GET("/:id", assetHandler.GetAsset)
			assets.PUT("/:id", assetHandler.UpdateAsset)
			assets.DELETE("/:id", assetHandler.DeleteAsset)
			assets.POST("/import/image/:image_gen_id", assetHandler.ImportFromImageGen)
			assets.POST("/import/video/:video_gen_id", assetHandler.ImportFromVideoGen)
		}

		storyboards := api.Group("/storyboards")
		{
			storyboards.PUT("/:id", storyboardHandler.UpdateStoryboard)
			storyboards.DELETE("/:id", storyboardHandler.DeleteStoryboard)
			storyboards.POST("/:id/frame-prompt", framePromptHandler.GenerateFramePrompt)
			storyboards.GET("/:id/frame-prompts", handlers2.GetStoryboardFramePrompts(db, log))
		}

		audio := api.Group("/audio")
		{
			audio.POST("/extract", audioExtractionHandler.ExtractAudio)
			audio.POST("/extract/batch", audioExtractionHandler.BatchExtractAudio)
		}

		music := api.Group("/music")
		{
			music.GET("/netease/search", musicHandler.SearchNetease)
			music.GET("/netease/song-url", musicHandler.GetNeteaseSongURL)
			music.GET("/netease/stream", musicHandler.StreamNeteaseSong)
			music.GET("/search", musicHandler.SearchAll)
			music.GET("/stream", musicHandler.StreamMusic)
		}

		media := api.Group("/media")
		{
			media.GET("/proxy", mediaHandler.Proxy)
		}

		sfx := api.Group("/sfx")
		{
			sfx.GET("", sfxHandler.ListSFX)
			sfx.POST("/generate", sfxHandler.GenerateSFX)
		}

		settings := api.Group("/settings")
		{
			settings.GET("/language", settingsHandler.GetLanguage)
			settings.PUT("/language", settingsHandler.UpdateLanguage)
		}
	}

	// 前端静态文件服务（放在API路由之后，避免冲突）
	// 服务前端构建产物
	r.Static("/assets", "./web/dist/assets")
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	// NoRoute处理：对于所有未匹配的路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是API路径，返回404
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}

		// SPA fallback - 返回index.html
		c.File("./web/dist/index.html")
	})

	return r
}
