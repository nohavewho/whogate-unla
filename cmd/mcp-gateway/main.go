package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/server"
	"github.com/mcp-ecosystem/mcp-gateway/internal/server/storage"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configPath string
	dataDir    string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of mcp-gateway",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcp-gateway version %s\n", version.Get())
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
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "data", "path to data directory")
	rootCmd.AddCommand(versionCmd)
}

func getConfigPath() string {
	// 1. Check command line flag
	if configPath != "" {
		return configPath
	}

	// 2. Check environment variable
	if envPath := os.Getenv("CONFIG_DIR"); envPath != "" {
		return envPath
	}

	// 3. Default to APPDATA/.mcp/gateway
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// For non-Windows systems, use HOME
		appData = os.Getenv("HOME")
		if appData == "" {
			panic("Neither APPDATA nor HOME environment variable is set")
		}
	}
	return filepath.Join(appData, ".mcp", "gateway")
}

func run() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("Starting mcp-gateway", zap.String("version", version.Get()))

	// Get configuration path
	configDir := getConfigPath()
	logger.Info("Using configuration directory", zap.String("path", configDir))

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Fatal("Failed to create config directory",
			zap.String("path", configDir),
			zap.Error(err))
	}

	// Load configuration
	cfgLoader := config.NewLoader(logger)

	// Check if config path is a directory
	info, err := os.Stat(configDir)
	if err != nil {
		logger.Fatal("failed to stat config path",
			zap.String("path", configDir),
			zap.Error(err))
	}

	var cfg *config.Config
	if info.IsDir() {
		cfg, err = cfgLoader.LoadFromDir(configDir)
	} else {
		cfg, err = cfgLoader.LoadFromFile(configDir)
	}

	if err != nil {
		logger.Fatal("failed to load configuration",
			zap.String("path", configDir),
			zap.Error(err))
	}

	// Initialize storage
	store, err := storage.NewDiskStorage(logger, dataDir)
	if err != nil {
		logger.Fatal("failed to initialize storage",
			zap.String("path", dataDir),
			zap.Error(err))
	}

	// Initialize server
	srv := server.NewServer(logger, store)

	// Initialize router
	router := gin.Default()

	// Register routes
	if err := srv.RegisterRoutes(router, cfg); err != nil {
		logger.Fatal("failed to register routes",
			zap.Error(err))
	}

	// Start server
	go func() {
		if err := router.Run(":5235"); err != nil {
			logger.Fatal("failed to start server",
				zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown server",
			zap.Error(err))
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
