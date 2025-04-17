package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Should set stricter checks in production
	},
}

type Config struct {
	// Configuration structure
	// Add more fields as needed
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create config loader
	loader := config.NewLoader(logger)

	r := gin.Default()

	// Configure routes
	r.POST("/api/configs", handleConfigUpdate)
	r.GET("/ws/chat", handleWebSocket)
	r.GET("/api/configs", func(c *gin.Context) {
		handleGetConfigs(c, loader)
	})

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

func handleConfigUpdate(c *gin.Context) {
	var cfg Config
	if err := c.ShouldBindYAML(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: Save configuration
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Process message
		response := "Received: " + string(message)
		if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}
}

// handleGetConfigs handles the GET /api/configs endpoint
func handleGetConfigs(c *gin.Context, loader *config.Loader) {
	// Get the config directory from environment variable or use default
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "./configs"
	}

	// Get all yaml files in the directory
	files, err := os.ReadDir(configDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read config directory: " + err.Error(),
		})
		return
	}

	// Load configurations from each yaml file
	configs := make([]*config.Config, 0)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".yaml") {
			continue
		}

		cfg, err := loader.LoadFromFile(filepath.Join(configDir, file.Name()))
		if err != nil {
			// Log error but continue with other files
			log.Printf("Failed to load config file %s: %v", file.Name(), err)
			continue
		}

		configs = append(configs, cfg)
	}

	// Return the list of configurations
	c.JSON(http.StatusOK, configs)
}
