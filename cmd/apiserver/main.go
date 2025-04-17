package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	configPath = flag.String("conf", "", "path to configuration file or directory")
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

func getConfigPath() string {
	// 1. Check command line flag
	if *configPath != "" {
		return *configPath
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

func main() {
	flag.Parse()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Get configuration path
	configDir := getConfigPath()
	logger.Info("Using configuration directory", zap.String("path", configDir))

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Fatal("Failed to create config directory",
			zap.String("path", configDir),
			zap.Error(err))
	}

	r := gin.Default()

	// Configure routes
	r.POST("/api/mcp-servers", handleMCPServerCreate)
	r.PUT("/api/mcp-servers", handleMCPServerUpdate)
	r.GET("/ws/chat", handleWebSocket)
	r.GET("/api/mcp-servers", handleGetMCPServers)

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

func handleMCPServerUpdate(c *gin.Context) {
	// Get the server name from query parameter
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name parameter is required"})
		return
	}

	// Read the raw YAML content from request body
	content, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body: " + err.Error()})
		return
	}

	// Validate the YAML content
	var cfg config.Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid YAML content: " + err.Error()})
		return
	}

	// Check if the server name in config matches the name parameter
	if len(cfg.Servers) == 0 || cfg.Servers[0].Name != name {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server name in configuration must match name parameter"})
		return
	}

	// Get the config directory
	configDir := getConfigPath()

	// Create the config file path
	configFile := filepath.Join(configDir, name+".yaml")

	// Check if the file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "MCP server not found"})
		return
	}

	// Write the content to file
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save MCP server configuration: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"path":   configFile,
	})
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

// handleGetMCPServers handles the GET /api/mcp-servers endpoint
func handleGetMCPServers(c *gin.Context) {
	// Get the config directory
	configDir := getConfigPath()

	// Get all yaml files in the directory
	files, err := os.ReadDir(configDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read MCP servers directory: " + err.Error(),
		})
		return
	}

	// Load configurations from each yaml file
	servers := make([]map[string]string, 0)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".yaml") {
			continue
		}

		// Read the raw YAML content
		content, err := os.ReadFile(filepath.Join(configDir, file.Name()))
		if err != nil {
			log.Printf("Failed to read MCP server file %s: %v", file.Name(), err)
			continue
		}

		// Add the YAML content to the response
		servers = append(servers, map[string]string{
			"name":   strings.TrimSuffix(file.Name(), ".yaml"),
			"config": string(content),
		})
	}

	// Return the list of MCP servers
	c.JSON(http.StatusOK, servers)
}

func handleMCPServerCreate(c *gin.Context) {
	// Read the raw YAML content from request body
	content, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body: " + err.Error()})
		return
	}

	// Validate the YAML content and get the server name
	var cfg config.Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid YAML content: " + err.Error()})
		return
	}

	// Check if there is at least one server in the config
	if len(cfg.Servers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no server configuration found in YAML"})
		return
	}

	// Use the first server's name
	name := cfg.Servers[0].Name
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server name is required in configuration"})
		return
	}

	// Get the config directory
	configDir := getConfigPath()

	// Create the config file path
	configFile := filepath.Join(configDir, name+".yaml")

	// Check if the file already exists
	if _, err := os.Stat(configFile); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "MCP server already exists"})
		return
	}

	// Write the content to file
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save MCP server configuration: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"path":   configFile,
	})
}
