package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mcp-ecosystem/mcp-gateway/cmd/apiserver/internal/database"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	configPath = flag.String("conf", "", "path to configuration file or directory")
	db         database.Database
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Should set stricter checks in production
	},
}

type WebSocketMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
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

	// Load configuration
	cfg, err := config.LoadConfig("configs/apiserver.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database based on configuration
	switch cfg.Database.Type {
	case "postgres":
		db = database.NewPostgresDB(&cfg.Database)
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

	r := gin.Default()

	// Configure routes
	r.POST("/api/mcp-servers", handleMCPServerCreate)
	r.PUT("/api/mcp-servers/:name", handleMCPServerUpdate)
	r.GET("/ws/chat", handleWebSocket)
	r.GET("/api/mcp-servers", handleGetMCPServers)
	r.DELETE("/api/mcp-servers/:name", handleMCPServerDelete)
	r.POST("/api/mcp-servers/sync", handleMCPServerSync)

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
	// Get the server name from path parameter instead of query parameter
	name := c.Param("name")
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
	// Get sessionId from query parameter
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	// Check if session exists, if not create it
	exists, err := db.SessionExists(c.Request.Context(), sessionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check session"})
		return
	}
	if !exists {
		// Create new session with the provided sessionId
		if err := db.CreateSession(c.Request.Context(), sessionId); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
			return
		}
	}

	// Log connection attempt
	log.Printf("[WS] New connection attempt - SessionID: %s, RemoteAddr: %s", sessionId, c.Request.RemoteAddr)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] Failed to upgrade connection - SessionID: %s, Error: %v", sessionId, err)
		return
	}
	defer conn.Close()

	// Load existing messages
	messages, err := db.GetMessages(c.Request.Context(), sessionId)
	if err != nil {
		log.Printf("[WS] Failed to load messages - SessionID: %s, Error: %v", sessionId, err)
	}

	// Send existing messages to client
	for _, msg := range messages {
		response := WebSocketMessage{
			Type:      "message",
			Content:   msg.Content,
			Sender:    msg.Sender,
			Timestamp: msg.Timestamp.UnixMilli(),
		}
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("[WS] Error sending existing message - SessionID: %s, Error: %v", sessionId, err)
			return
		}
	}

	// Log successful connection
	log.Printf("[WS] Connection established - SessionID: %s, RemoteAddr: %s", sessionId, c.Request.RemoteAddr)

	for {
		var message WebSocketMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Printf("[WS] Error reading message - SessionID: %s, Error: %v", sessionId, err)
			break
		}

		// Save all incoming messages to database
		dbMessage := &database.Message{
			ID:        message.Type + "-" + time.Now().Format(time.RFC3339Nano),
			SessionID: sessionId,
			Content:   message.Content,
			Sender:    message.Sender,
			Timestamp: time.Now(),
		}
		if err := db.SaveMessage(c.Request.Context(), dbMessage); err != nil {
			log.Printf("[WS] Failed to save message - SessionID: %s, Error: %v", sessionId, err)
		}

		// Log received message
		log.Printf("[WS] Message received - SessionID: %s, Type: %s, Content: %s",
			sessionId, message.Type, message.Content)

		// Process message based on type
		switch message.Type {
		case "message":
			// Echo the message back for now
			response := WebSocketMessage{
				Type:      "message",
				Content:   "Echo: " + message.Content,
				Sender:    "bot",
				Timestamp: time.Now().UnixMilli(),
			}
			if err := conn.WriteJSON(response); err != nil {
				log.Printf("[WS] Error writing message - SessionID: %s, Error: %v", sessionId, err)
				break
			}

			// Save bot response to database
			dbMessage := &database.Message{
				ID:        "bot-" + time.Now().Format(time.RFC3339Nano),
				SessionID: sessionId,
				Content:   response.Content,
				Sender:    response.Sender,
				Timestamp: time.Now(),
			}
			if err := db.SaveMessage(c.Request.Context(), dbMessage); err != nil {
				log.Printf("[WS] Failed to save bot message - SessionID: %s, Error: %v", sessionId, err)
			}

			// Log sent message
			log.Printf("[WS] Message sent - SessionID: %s, Type: %s, Content: %s",
				sessionId, response.Type, response.Content)
		case "system":
			// Handle system messages if needed
			log.Printf("Received system message: %s", message.Content)
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

func handleMCPServerDelete(c *gin.Context) {
	// Get the server name from path parameter
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name parameter is required"})
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

	// Delete the file
	if err := os.Remove(configFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete MCP server configuration: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func handleMCPServerSync(c *gin.Context) {
	// Get the config directory
	configDir := getConfigPath()

	// Read all YAML files in the config directory
	files, err := os.ReadDir(configDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read config directory: " + err.Error(),
		})
		return
	}

	// Validate all YAML files
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(configDir, file.Name()))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to read config file: " + err.Error(),
			})
			return
		}

		// Validate the YAML content
		var cfg config.Config
		if err := yaml.Unmarshal(content, &cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid YAML content in " + file.Name() + ": " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(files),
	})
}
