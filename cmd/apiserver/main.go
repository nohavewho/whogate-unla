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

// initLogger initializes the application logger
func initLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	return logger
}

// initConfig loads and returns the application configuration
func initConfig(logger *zap.Logger) *config.APIServerConfig {
	cfg, cfgPath, err := config.LoadConfig[config.APIServerConfig](configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration",
			zap.String("path", cfgPath), zap.Error(err))
	}
	logger.Info("Loaded configuration", zap.String("path", cfgPath))
	return cfg
}

// initNotifier initializes the notifier service
func initNotifier(ctx context.Context, logger *zap.Logger, cfg *config.NotifierConfig) notifier.Notifier {
	if notifier.Type(cfg.Type) == notifier.TypeComposite {
		log.Fatal("Composite notifier is not supported in apiserver")
	}
	ntf, err := notifier.NewNotifier(ctx, logger, cfg)
	if err != nil {
		logger.Fatal("Failed to initialize notifier", zap.Error(err))
	}
	return ntf
}

// initOpenAI initializes the OpenAI client
func initOpenAI(cfg *config.OpenAIConfig) *openai.Client {
	return openai.NewClient(cfg)
}

// initDatabase initializes the database connection
func initDatabase(logger *zap.Logger, cfg *config.DatabaseConfig) database.Database {
	logger.Info("Initializing database", zap.String("type", cfg.Type))
	db, err := database.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	logger.Info("Database initialized", zap.String("type", cfg.Type))
	return db
}

// initStore initializes the storage service
func initStore(logger *zap.Logger, cfg *config.StorageConfig) storage.Store {
	store, err := storage.NewStore(logger, cfg)
	if err != nil {
		logger.Fatal("Failed to initialize store", zap.Error(err))
	}
	return store
}

// initRouter initializes the HTTP router and handlers
func initRouter(db database.Database, store storage.Store, ntf notifier.Notifier, openaiClient *openai.Client) *gin.Engine {
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

	return r
}

// startServer starts the HTTP server
func startServer(logger *zap.Logger, router *gin.Engine) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5234"
	}

	logger.Info("Server starting", zap.String("port", port))
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	logger := initLogger()
	defer logger.Sync()

	// Load configuration
	cfg := initConfig(logger)

	// Initialize services
	ntf := initNotifier(ctx, logger, &cfg.Notifier)
	openaiClient := initOpenAI(&cfg.OpenAI)
	db := initDatabase(logger, &cfg.Database)
	defer db.Close()
	store := initStore(logger, &cfg.Storage)

	logger.Info("Starting apiserver", zap.String("version", version.Get()))

	// Initialize router and start server
	router := initRouter(db, store, ntf, openaiClient)
	startServer(logger, router)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
