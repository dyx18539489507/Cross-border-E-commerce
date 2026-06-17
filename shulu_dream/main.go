package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/drama-generator/backend/api/routes"
	"github.com/drama-generator/backend/infrastructure/database"
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

	logr.Info("Starting Drama Generator API Server...")

	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logr.Fatalw("Failed to connect to database", "error", err)
	}
	logr.Info("Database connected successfully")

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(db); err != nil {
		logr.Fatalw("Failed to migrate database", "error", err)
	}
	logr.Info("Database tables migrated successfully")

	// 初始化本地存储
	var localStorage *storage.LocalStorage
	if cfg.Storage.Type == "local" {
		localStorage, err = storage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.BaseURL)
		if err != nil {
			logr.Fatalw("Failed to initialize local storage", "error", err)
		}
		logr.Infow("Local storage initialized successfully", "path", cfg.Storage.LocalPath)
	}

	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	listenAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	if cfg.Server.Host != "" && cfg.Server.Host != "0.0.0.0" {
		listenAddr = net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
	}
	listener, err := listenWithDevPortTakeover(cfg, logr, "tcp", listenAddr)
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			logr.Fatalw(
				"Server port is already in use",
				"port", cfg.Server.Port,
				"listen_addr", listenAddr,
				"hint", fmt.Sprintf("Run `lsof -nP -iTCP:%d -sTCP:LISTEN` to find the process, then stop it or set SERVER_PORT to another port", cfg.Server.Port),
			)
		}
		logr.Fatalw("Failed to bind server port", "error", err, "listen_addr", listenAddr)
	}

	router := routes.SetupRouter(cfg, db, logr, localStorage)

	srv := &http.Server{
		Addr:         listenAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	go func() {
		logr.Infow("🚀 Server starting...",
			"port", cfg.Server.Port,
			"listen_addr", listenAddr,
			"mode", gin.Mode())
		logr.Info("📍 Access URLs:")
		logr.Info(fmt.Sprintf("   Frontend:  http://localhost:%d", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   API:       http://localhost:%d/api/v1", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   Health:    http://localhost:%d/health", cfg.Server.Port))
		logr.Info("📁 Static files:")
		logr.Info(fmt.Sprintf("   Uploads:   http://localhost:%d/static", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   Assets:    http://localhost:%d/assets", cfg.Server.Port))
		logr.Info("✅ Server is ready!")

		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			logr.Fatalw("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logr.Info("Shutting down server...")

	// 清理资源

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logr.Fatalw("Server forced to shutdown", "error", err)
	}

	logr.Info("Server exited")
}

func listenWithDevPortTakeover(cfg *config.Config, logr *logger.Logger, network, listenAddr string) (net.Listener, error) {
	listener, err := net.Listen(network, listenAddr)
	if err == nil || !cfg.App.Debug || !errors.Is(err, syscall.EADDRINUSE) {
		return listener, err
	}

	pid, findErr := findListeningPID(cfg.Server.Port)
	if findErr != nil {
		logr.Warnw("Failed to inspect existing listener while retrying dev startup", "port", cfg.Server.Port, "error", findErr)
		return retryBusyListen(network, listenAddr, 2*time.Second)
	}

	sameWorkspace, matchErr := isSameWorkspaceProcess(pid)
	if matchErr != nil {
		logr.Warnw("Failed to inspect existing listener while retrying dev startup", "pid", pid, "error", matchErr)
		return retryBusyListen(network, listenAddr, 2*time.Second)
	}
	if !sameWorkspace {
		return retryBusyListen(network, listenAddr, 2*time.Second)
	}

	logr.Warnw("Stopping previous dev server instance that is using the port", "port", cfg.Server.Port, "pid", pid)
	if killErr := syscall.Kill(pid, syscall.SIGTERM); killErr != nil {
		logr.Warnw("Failed to stop previous dev server instance", "pid", pid, "error", killErr)
		return nil, err
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)

		listener, retryErr := net.Listen(network, listenAddr)
		if retryErr == nil {
			logr.Infow("Reclaimed busy dev server port from previous instance", "port", cfg.Server.Port, "previous_pid", pid)
			return listener, nil
		}
		if !errors.Is(retryErr, syscall.EADDRINUSE) {
			return nil, retryErr
		}
	}

	return nil, err
}

func retryBusyListen(network, listenAddr string, timeout time.Duration) (net.Listener, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)

		listener, err := net.Listen(network, listenAddr)
		if err == nil {
			return listener, nil
		}
		lastErr = err
		if !errors.Is(err, syscall.EADDRINUSE) {
			return nil, err
		}
	}

	if lastErr == nil {
		lastErr = syscall.EADDRINUSE
	}
	return nil, lastErr
}

func findListeningPID(port int) (int, error) {
	lsofPath, err := exec.LookPath("lsof")
	if err != nil {
		return 0, err
	}

	out, err := exec.Command(lsofPath, "-nP", fmt.Sprintf("-iTCP:%d", port), "-sTCP:LISTEN", "-t").Output()
	if err != nil {
		return 0, err
	}

	firstLine := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	if firstLine == "" {
		return 0, fmt.Errorf("no listening process found for port %d", port)
	}

	pid, err := strconv.Atoi(firstLine)
	if err != nil {
		return 0, fmt.Errorf("parse pid %q: %w", firstLine, err)
	}
	return pid, nil
}

func isSameWorkspaceProcess(pid int) (bool, error) {
	if pid <= 0 || pid == os.Getpid() {
		return false, nil
	}

	currentWD, err := os.Getwd()
	if err != nil {
		return false, err
	}
	currentWD, err = filepath.EvalSymlinks(currentWD)
	if err != nil {
		return false, err
	}

	processWD, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	if err != nil {
		return false, err
	}
	processWD, err = filepath.EvalSymlinks(processWD)
	if err != nil {
		return false, err
	}

	return currentWD == processWD, nil
}
