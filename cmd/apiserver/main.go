package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mcp-ecosystem/mcp-gateway/cmd/apiserver/internal/database"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/openai"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"
	openaiGo "github.com/openai/openai-go"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	configPath   string
	db           database.Database
	openaiClient *openai.Client
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of apiserver",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("apiserver version %s\n", version.Get())
	},
}

var rootCmd = &cobra.Command{
	Use:   "apiserver",
	Short: "MCP API Server",
	Long:  `MCP API Server provides API endpoints for MCP ecosystem`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "conf", "", "path to configuration file or directory")
	rootCmd.AddCommand(versionCmd)
}

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

	// Configure routes
	r.POST("/api/mcp-servers", handleMCPServerCreate)
	r.PUT("/api/mcp-servers/:name", handleMCPServerUpdate)
	r.GET("/ws/chat", handleWebSocket)
	r.GET("/api/mcp-servers", handleGetMCPServers)
	r.DELETE("/api/mcp-servers/:name", handleMCPServerDelete)
	r.POST("/api/mcp-servers/sync", handleMCPServerSync)
	r.GET("/api/chat/sessions", handleGetChatSessions)
	r.GET("/api/chat/sessions/:sessionId/messages", handleGetChatMessages)

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
	var cfg config.MCPConfig
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
			ID:        uuid.New().String(),
			SessionID: sessionId,
			Content:   message.Content,
			Sender:    message.Sender,
			Timestamp: time.Now(),
		}
		if err := db.SaveMessage(c.Request.Context(), dbMessage); err != nil {
			log.Printf("[WS] Failed to save message - SessionID: %s, Error: %v", sessionId, err)
		}

		// If this is the first message, update the session title
		messages, err := db.GetMessages(c.Request.Context(), sessionId)
		if err != nil {
			log.Printf("[WS] Failed to get messages - SessionID: %s, Error: %v", sessionId, err)
		} else if len(messages) == 1 {
			// Extract title from the first message (first 20 UTF-8 characters)
			title := message.Content
			runes := []rune(title)
			if len(runes) > 20 {
				title = string(runes[:20])
			}
			if err := db.UpdateSessionTitle(c.Request.Context(), sessionId, title); err != nil {
				log.Printf("[WS] Failed to update session title - SessionID: %s, Error: %v", sessionId, err)
			}
		}

		// Log received message
		log.Printf("[WS] Message received - SessionID: %s, Type: %s, Content: %s",
			sessionId, message.Type, message.Content)

		// Process message based on type
		switch message.Type {
		case "message":
			// Get conversation history from database
			messages, err := db.GetMessages(c.Request.Context(), sessionId)
			if err != nil {
				log.Printf("[WS] Failed to get conversation history - SessionID: %s, Error: %v", sessionId, err)
				continue
			}

			// Convert messages to OpenAI format
			openaiMessages := make([]openaiGo.ChatCompletionMessageParamUnion, len(messages))
			for i, msg := range messages {
				openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
					OfUser: &openaiGo.ChatCompletionUserMessageParam{
						Content: openaiGo.ChatCompletionUserMessageParamContentUnion{
							OfString: openaiGo.String(msg.Content),
						},
					},
				}
			}

			// Get streaming response from OpenAI
			stream, err := openaiClient.ChatCompletionStream(c.Request.Context(), openaiMessages)
			if err != nil {
				log.Printf("[WS] Failed to get OpenAI response - SessionID: %s, Error: %v", sessionId, err)
				continue
			}

			// Initialize response content
			responseContent := ""

			// Process stream chunks
			for stream.Next() {
				chunk := stream.Current()
				if chunk.Choices[0].Delta.Content != "" {
					responseContent += chunk.Choices[0].Delta.Content

					// Send chunk to client
					response := WebSocketMessage{
						Type:      "stream",
						Content:   chunk.Choices[0].Delta.Content,
						Sender:    "bot",
						Timestamp: time.Now().UnixMilli(),
					}
					if err := conn.WriteJSON(response); err != nil {
						log.Printf("[WS] Error writing message chunk - SessionID: %s, Error: %v", sessionId, err)
						break
					}
				}
			}

			if err := stream.Err(); err != nil {
				log.Printf("[WS] Error in stream - SessionID: %s, Error: %v", sessionId, err)
				continue
			}

			// Save complete bot response to database
			dbMessage := &database.Message{
				ID:        uuid.New().String(),
				SessionID: sessionId,
				Content:   responseContent,
				Sender:    "bot",
				Timestamp: time.Now(),
			}
			if err := db.SaveMessage(c.Request.Context(), dbMessage); err != nil {
				log.Printf("[WS] Failed to save bot message - SessionID: %s, Error: %v", sessionId, err)
			}

			// Log sent message
			log.Printf("[WS] Message sent - SessionID: %s, Type: %s, Content: %s",
				sessionId, "message", responseContent)
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
	var cfg config.MCPConfig
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

func sendReloadSignal() error {
	// Load configuration
	cfg, err := config.LoadConfig[config.APIServerConfig]("configs/apiserver.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Read gateway PID file
	pidBytes, err := os.ReadFile(cfg.GatewayPID)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return fmt.Errorf("invalid PID in file: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to send reload signal: %w", err)
	}

	return nil
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
		var cfg config.MCPConfig
		if err := yaml.Unmarshal(content, &cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid YAML content in " + file.Name() + ": " + err.Error(),
			})
			return
		}
	}

	// Send reload signal to gateway
	if err := sendReloadSignal(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to reload gateway: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(files),
	})
}

// handleGetChatSessions handles the GET /api/chat/sessions endpoint
func handleGetChatSessions(c *gin.Context) {
	sessions, err := db.GetSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chat sessions"})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// handleGetChatMessages handles the GET /api/chat/messages/:sessionId endpoint
func handleGetChatMessages(c *gin.Context) {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	// Get pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Get messages with pagination
	messages, err := db.GetMessagesWithPagination(c.Request.Context(), sessionId, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
