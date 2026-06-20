package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drama-generator/backend/api/routes"
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/infrastructure/database"
	scheduler2 "github.com/drama-generator/backend/infrastructure/scheduler"
	"github.com/drama-generator/backend/infrastructure/storage"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logr := logger.NewLogger(cfg.App.Debug)
	defer logr.Sync()

	logr.Info("Starting Digital Silk Road API Server...")

	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logr.Fatal("Failed to connect to database", "error", err)
	}
	logr.Info("Database connected successfully")

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(db); err != nil {
		logr.Fatal("Failed to migrate database", "error", err)
	}
	logr.Info("Database tables migrated successfully")

	// 本地目录同时承担本地存储和远端对象存储的上传缓存，所有部署模式都必须初始化。
	localStorage, err := storage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.BaseURL)
	if err != nil {
		logr.Fatal("Failed to initialize local storage", "error", err)
	}
	logr.Info("Local storage initialized successfully", "path", cfg.Storage.LocalPath)

	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := routes.SetupRouter(cfg, db, logr, localStorage)

	distributionService := services.NewDistributionService(db, cfg, logr)
	distributionScheduler := scheduler2.NewDistributionScheduler(
		distributionService,
		time.Duration(cfg.Distribution.StatusPollIntervalSecond)*time.Second,
		logr,
	)
	distributionScheduler.Start()

	readTimeout := time.Duration(cfg.Server.ReadTimeout) * time.Second
	if readTimeout <= 0 {
		readTimeout = 10 * time.Minute
	}
	writeTimeout := time.Duration(cfg.Server.WriteTimeout) * time.Second
	if writeTimeout <= 0 {
		writeTimeout = 10 * time.Minute
	}
	serverHost := cfg.Server.Host
	if serverHost == "" {
		serverHost = "0.0.0.0"
	}
	srv := &http.Server{
		Addr:         net.JoinHostPort(serverHost, fmt.Sprintf("%d", cfg.Server.Port)),
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	go func() {
		logr.Infow("🚀 Server starting...",
			"port", cfg.Server.Port,
			"mode", gin.Mode())
		logr.Info("📍 Access URLs:")
		logr.Info(fmt.Sprintf("   Frontend:  http://localhost:%d", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   API:       http://localhost:%d/api/v1", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   Health:    http://localhost:%d/health", cfg.Server.Port))
		logr.Info("📁 Static files:")
		logr.Info(fmt.Sprintf("   Uploads:   http://localhost:%d/static", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   Assets:    http://localhost:%d/assets", cfg.Server.Port))
		logr.Info("✅ Server is ready!")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logr.Fatal("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logr.Info("Shutting down server...")

	// 清理资源
	distributionScheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logr.Fatal("Server forced to shutdown", "error", err)
	}

	logr.Info("Server exited")
}
