package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/apiserver/database"
	"github.com/mcp-ecosystem/mcp-gateway/internal/apiserver/handler"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/openai"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"os"
	"path/filepath"
)

var (
	configPath   string
	db           database.Database
	openaiClient *openai.Client

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of apiserver",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("apiserver version %s\n", version.Get())
		},
	}

	rootCmd = &cobra.Command{
		Use:   "apiserver",
		Short: "MCP API Server",
		Long:  `MCP API Server provides API endpoints for MCP ecosystem`,
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "conf", "", "path to configuration file or directory")
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
			log.Fatal("Neither APPDATA nor HOME environment variable is set")
		}
	}
	return filepath.Join(appData, ".mcp", "gateway")
}

func run() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig[config.APIServerConfig]("configs/apiserver.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize OpenAI client
	openaiClient = openai.NewClient(&cfg.OpenAI)

	// Initialize database based on configuration
	switch cfg.Database.Type {
	case "postgres":
		db = database.NewPostgresDB(&cfg.Database)
	case "sqlite":
		db = database.NewSQLiteDB(&cfg.Database)
	case "mysql":
		db = database.NewMySQLDB(&cfg.Database)
	default:
		logger.Fatal("Unsupported database type", zap.String("type", cfg.Database.Type))
	}

	if err := db.Init(context.Background()); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Get configuration path
	configDir := getConfigPath()
	logger.Info("Using configuration directory", zap.String("path", configDir))

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Fatal("Failed to create config directory",
			zap.String("path", configDir),
			zap.Error(err))
	}

	logger.Info("Starting apiserver", zap.String("version", version.Get()))

	r := gin.Default()

	chatHandler := handler.NewChat(db)
	mcpHandler := handler.NewMCP(db)
	wsHandler := handler.NewWebSocket(db, openaiClient)

	// Configure routes
	r.POST("/api/mcp-servers", mcpHandler.HandleMCPServerCreate)
	r.PUT("/api/mcp-servers/:name", mcpHandler.HandleMCPServerUpdate)
	r.GET("/api/mcp-servers", mcpHandler.HandleGetMCPServers)
	r.DELETE("/api/mcp-servers/:name", mcpHandler.HandleMCPServerDelete)
	r.POST("/api/mcp-servers/sync", mcpHandler.HandleMCPServerSync)

	r.GET("/ws/chat", wsHandler.HandleWebSocket)

	r.GET("/api/chat/sessions", chatHandler.HandleGetChatSessions)
	r.GET("/api/chat/sessions/:sessionId/messages", chatHandler.HandleGetChatMessages)

	// Static file server
	r.Static("/static", "./static")

	port := os.Getenv("PORT")
	if port == "" {
		port = "5234"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
