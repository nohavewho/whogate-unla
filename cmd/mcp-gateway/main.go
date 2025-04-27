package main

import (
	"context"
	"fmt"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/storage/helper"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/storage"

	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/core"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configPath   string
	pidFile      string
	reloadChan   chan struct{}
	serverLock   sync.RWMutex
	httpServer   *http.Server
	reloadServer *http.Server
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of mcp-gateway",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcp-gateway version %s\n", version.Get())
	},
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the configuration of a running mcp-gateway instance",
	Run: func(cmd *cobra.Command, args []string) {
		pidBytes, err := os.ReadFile(pidFile)
		if err != nil {
			fmt.Printf("Failed to read PID file: %v\n", err)
			os.Exit(1)
		}

		pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if err != nil {
			fmt.Printf("Invalid PID in file: %v\n", err)
			os.Exit(1)
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("Failed to find process: %v\n", err)
			os.Exit(1)
		}

		if err := process.Signal(syscall.SIGHUP); err != nil {
			fmt.Printf("Failed to send reload signal: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Reload signal sent successfully")
	},
}

var rootCmd = &cobra.Command{
	Use:   "mcp-gateway",
	Short: "MCP Gateway service",
	Long:  `MCP Gateway is a service that provides API gateway functionality for MCP ecosystem`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "conf", "", "path to configuration file or directory")
	rootCmd.PersistentFlags().StringVar(&pidFile, "pid", "", "path to PID file")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(reloadCmd)
}

func getCfgPath() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	configsPath := filepath.Join(currentDir, "configs", "mcp-gateway.yaml")
	if _, err := os.Stat(configsPath); err == nil {
		return configsPath
	}

	return "/etc/mcp-gateway/mcp-gateway.yaml"
}

func getPIDFile() string {
	if pidFile != "" {
		return pidFile
	}

	cfgPath := getCfgPath()
	cfg, err := config.LoadConfig[config.MCPGatewayConfig](cfgPath)
	if err != nil {
		return "/var/run/mcp-gateway.pid"
	}

	return cfg.PID
}

func writePIDFile() error {
	dir := filepath.Dir(pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

func removePIDFile() error {
	return os.Remove(pidFile)
}

func handleReload(ctx context.Context, logger *zap.Logger, store storage.Store, srv *core.Server, cfg *config.MCPGatewayConfig) {
	logger.Info("Reloading MCP configuration")

	mcpConfigs, err := store.List(ctx)
	if err != nil {
		logger.Fatal("Failed to load MCP configurations",
			zap.Error(err))
	}
	newMCPCfg, err := helper.MergeConfigs(mcpConfigs)
	if err != nil {
		logger.Fatal("failed to merge MCP configurations",
			zap.Error(err))
	}

	serverLock.Lock()
	defer serverLock.Unlock()

	newRouter := gin.New()

	if err := srv.RegisterRoutes(newRouter, newMCPCfg); err != nil {
		logger.Error("failed to register new routes",
			zap.Error(err))
		return
	}

	if err := srv.UpdateConfig(newMCPCfg); err != nil {
		logger.Error("failed to update server configuration",
			zap.Error(err))
		return
	}

	newServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: newRouter,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown old server",
			zap.Error(err))
		return
	}

	httpServer = newServer
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start new server",
				zap.Error(err))
		}
	}()
}

func run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	pidFile = getPIDFile()

	if err := writePIDFile(); err != nil {
		logger.Fatal("Failed to write PID file",
			zap.String("path", pidFile),
			zap.Error(err))
	}
	defer removePIDFile()

	logger.Info("Starting mcp-gateway", zap.String("version", version.Get()))

	cfgPath := getCfgPath()
	logger.Info("Using configuration paths",
		zap.String("service_config", cfgPath))

	cfg, err := config.LoadConfig[config.MCPGatewayConfig](cfgPath)
	if err != nil {
		logger.Fatal("Failed to load service configuration",
			zap.String("path", cfgPath),
			zap.Error(err))
	}

	// Initialize storage and load initial configuration
	store, err := storage.NewStore(ctx, logger, &cfg.Storage)
	if err != nil {
		logger.Fatal("failed to initialize storage",
			zap.String("type", cfg.Storage.Type),
			zap.Error(err))
	}

	mcpConfigs, err := store.List(ctx)
	if err != nil {
		logger.Fatal("Failed to load MCP configurations",
			zap.Error(err))
	}
	mcpCfg, err := helper.MergeConfigs(mcpConfigs)
	if err != nil {
		logger.Fatal("failed to merge MCP configurations",
			zap.Error(err))
	}

	srv := core.NewServer(logger)
	router := gin.Default()

	if err := srv.RegisterRoutes(router, mcpCfg); err != nil {
		logger.Fatal("failed to register routes",
			zap.Error(err))
	}

	httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	notifier := store.GetNotifier(ctx)
	updateCh, err := notifier.Watch(ctx)
	if err != nil {
		logger.Fatal("failed to start watching for updates",
			zap.Error(err))
	}

	// Keep the original HTTP reload endpoint
	reloadChan = make(chan struct{})
	reloadRouter := gin.Default()
	reloadRouter.POST("/_reload", func(c *gin.Context) {
		reloadChan <- struct{}{}
		c.JSON(http.StatusOK, gin.H{"status": "reload triggered"})
	})

	reloadServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.InnerPort),
		Handler: reloadRouter,
	}

	go func() {
		logger.Info("Starting main server", zap.Int("port", cfg.Port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start main server",
				zap.Error(err))
		}
	}()

	go func() {
		logger.Info("Starting inner server", zap.Int("port", cfg.InnerPort))
		if err := reloadServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start inner server",
				zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-quit:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := httpServer.Shutdown(ctx); err != nil {
				cancel()
				logger.Error("failed to shutdown main server",
					zap.Error(err))
			}
			if err := reloadServer.Shutdown(ctx); err != nil {
				cancel()
				logger.Error("failed to shutdown reload server",
					zap.Error(err))
			}
			cancel()
			return
		case <-updateCh:
			logger.Info("Received reload signal")
			handleReload(ctx, logger, store, srv, cfg)
		case <-reloadChan:
			handleReload(ctx, logger, store, srv, cfg)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
