package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/server"
	"github.com/mcp-ecosystem/mcp-gateway/internal/server/storage"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("conf", "configs", "path to configuration file or directory")
	dataDir    = flag.String("data-dir", "data", "path to data directory")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Load configuration
	cfgLoader := config.NewLoader(logger)

	// Check if config path is a directory
	info, err := os.Stat(*configPath)
	if err != nil {
		logger.Fatal("failed to stat config path",
			zap.String("path", *configPath),
			zap.Error(err))
	}

	var cfg *config.Config
	if info.IsDir() {
		cfg, err = cfgLoader.LoadFromDir(*configPath)
	} else {
		cfg, err = cfgLoader.LoadFromFile(*configPath)
	}

	if err != nil {
		logger.Fatal("failed to load configuration",
			zap.String("path", *configPath),
			zap.Error(err))
	}

	// Initialize storage
	store, err := storage.NewDiskStorage(logger, *dataDir)
	if err != nil {
		logger.Fatal("failed to initialize storage",
			zap.String("path", *dataDir),
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
		if err := router.Run(":8080"); err != nil {
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
