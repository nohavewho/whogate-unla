package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mcp-ecosystem/mcp-gateway/internal/apiserver/database"
	"github.com/mcp-ecosystem/mcp-gateway/internal/apiserver/handler"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/cnst"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/storage"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/storage/notifier"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/openai"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configPath string
	db         database.Database

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
	rootCmd.PersistentFlags().StringVar(&configPath, "conf", cnst.ApiServerYaml, "path to configuration file, like /etc/mcp-gateway/apiserver.yaml")
	rootCmd.AddCommand(versionCmd)
}

func run() {
	ctx := context.Background()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, cfgPath, err := config.LoadConfig[config.APIServerConfig](configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration",
			zap.String("path", cfgPath), zap.Error(err))
	}
	logger.Info("Loaded configuration", zap.String("path", cfgPath))

	// Initialize notifier
	ntf, err := notifier.NewNotifier(ctx, logger, &cfg.Notifier)
	if err != nil {
		logger.Fatal("Failed to initialize notifier", zap.Error(err))
	}

	// Initialize OpenAI client
	openaiClient := openai.NewClient(&cfg.OpenAI)

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

	// Initialize store using factory
	store, err := storage.NewStore(logger, &cfg.Storage)
	if err != nil {
		logger.Fatal("Failed to initialize store", zap.Error(err))
	}

	logger.Info("Starting apiserver", zap.String("version", version.Get()))

	r := gin.Default()

	chatHandler := handler.NewChat(db)
	mcpHandler := handler.NewMCP(db, store, ntf)
	wsHandler := handler.NewWebSocket(db, openaiClient)

	// Configure routes
	r.POST("/api/mcp-servers", mcpHandler.HandleMCPServerCreate)
	r.PUT("/api/mcp-servers/:name", mcpHandler.HandleMCPServerUpdate)
	r.GET("/api/mcp-servers", mcpHandler.HandleListMCPServers)
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
