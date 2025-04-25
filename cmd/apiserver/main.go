package main

import (
	"context"
	"encoding/json"
	"fmt"
	database2 "github.com/mcp-ecosystem/mcp-gateway/internal/apiserver/database"
	config2 "github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/dto"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/openai"
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
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"
	openaiGo "github.com/openai/openai-go"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	configPath   string
	db           database2.Database
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Should set stricter checks in production
	},
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
	cfg, err := config2.LoadConfig[config2.APIServerConfig]("configs/apiserver.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize OpenAI client
	openaiClient = openai.NewClient(&cfg.OpenAI)

	// Initialize database based on configuration
	switch cfg.Database.Type {
	case "postgres":
		db = database2.NewPostgresDB(&cfg.Database)
	case "sqlite":
		db = database2.NewSQLiteDB(&cfg.Database)
	case "mysql":
		db = database2.NewMySQLDB(&cfg.Database)
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
	var cfg config2.MCPConfig
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

	// Send reload signal to gateway
	if err := sendReloadSignal(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to reload gateway: " + err.Error(),
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
		var message dto.WebSocketMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Printf("[WS] Error reading message - SessionID: %s, Error: %v", sessionId, err)
			break
		}

		// Log received message
		log.Printf("[WS] Message received - SessionID: %s, Type: %s, Content: %s",
			sessionId, message.Type, message.Content)

		// Process message based on type
		switch message.Type {
		case dto.MsgTypeMessage:

			// Save all incoming messages to database
			msg := &database2.Message{
				ID:        uuid.New().String(),
				SessionID: sessionId,
				Content:   message.Content,
				Sender:    message.Sender,
				Timestamp: time.Now(),
			}
			if err := db.SaveMessage(c.Request.Context(), msg); err != nil {
				log.Printf("[WS] Failed to save message - SessionID: %s, Error: %v", sessionId, err)
			}

			// Get conversation history from database
			messages, err := db.GetMessages(c.Request.Context(), sessionId)
			if err != nil {
				log.Printf("[WS] Failed to get conversation history - SessionID: %s, Error: %v", sessionId, err)
				continue
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

			// Convert messages to OpenAI format
			openaiMessages := make([]openaiGo.ChatCompletionMessageParamUnion, len(messages))
			for i, msg := range messages {
				if msg.Sender == "bot" {
					if msg.ToolCalls != "" {
						// For bot messages with tool calls, use OfAssistant type with ToolCalls field
						var toolCalls []dto.ToolCall
						if err := json.Unmarshal([]byte(msg.ToolCalls), &toolCalls); err != nil {
							log.Printf("[WS] Failed to unmarshal tool calls - SessionID: %s, Error: %v", sessionId, err)
							continue
						}
						openaiToolCalls := make([]openaiGo.ChatCompletionMessageToolCallParam, len(toolCalls))
						for j, tc := range toolCalls {
							openaiToolCalls[j] = openaiGo.ChatCompletionMessageToolCallParam{
								ID:   tc.ID,
								Type: "function",
								Function: openaiGo.ChatCompletionMessageToolCallFunctionParam{
									Name:      tc.Function.Name,
									Arguments: tc.Function.Arguments,
								},
							}
						}
						openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
							OfAssistant: &openaiGo.ChatCompletionAssistantMessageParam{
								ToolCalls: openaiToolCalls,
							},
						}
					} else {
						// For regular bot messages, use OfAssistant type
						openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
							OfAssistant: &openaiGo.ChatCompletionAssistantMessageParam{
								Content: openaiGo.ChatCompletionAssistantMessageParamContentUnion{
									OfString: openaiGo.String(msg.Content),
								},
							},
						}
					}
				} else {
					if msg.ToolResult != "" {
						// For user messages with tool results, use OfTool type
						var toolResult dto.ToolResult
						if err := json.Unmarshal([]byte(msg.ToolResult), &toolResult); err != nil {
							log.Printf("[WS] Failed to unmarshal tool result - SessionID: %s, Error: %v", sessionId, err)
							continue
						}
						openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
							OfTool: &openaiGo.ChatCompletionToolMessageParam{
								ToolCallID: toolResult.ToolCallID,
								Content: openaiGo.ChatCompletionToolMessageParamContentUnion{
									OfString: openaiGo.String(toolResult.Result),
								},
							},
						}
					} else {
						// For regular user messages, use OfUser type
						openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
							OfUser: &openaiGo.ChatCompletionUserMessageParam{
								Content: openaiGo.ChatCompletionUserMessageParamContentUnion{
									OfString: openaiGo.String(msg.Content),
								},
							},
						}
					}
				}
			}

			// Convert tools to OpenAI format if provided
			var openaiTools []openaiGo.ChatCompletionToolParam
			if len(message.Tools) > 0 {
				openaiTools = make([]openaiGo.ChatCompletionToolParam, len(message.Tools))
				for i, tool := range message.Tools {
					openaiTools[i] = openaiGo.ChatCompletionToolParam{
						Function: openaiGo.FunctionDefinitionParam{
							Name:        tool.Name,
							Description: openaiGo.String(tool.Description),
							Parameters: openaiGo.FunctionParameters{
								"type":       "object",
								"properties": tool.Parameters.Properties,
								"required":   tool.Parameters.Required,
							},
							Strict: openaiGo.Bool(true),
						},
					}
				}
			}

			// Get streaming response from OpenAI
			stream, err := openaiClient.ChatCompletionStream(c.Request.Context(), openaiMessages, openaiTools)
			if err != nil {
				log.Printf("[WS] Failed to get OpenAI response - SessionID: %s, Error: %v", sessionId, err)
				continue
			}

			// Initialize response content
			responseContent := ""
			var toolCall *dto.ToolCall
			var toolCallArguments string

			// Process stream chunks
			for stream.Next() {
				chunk := stream.Current()
				// Check if this is a tool call
				if chunk.Choices[0].Delta.ToolCalls != nil {
					// Initialize tool call if not exists
					if toolCall == nil {
						toolCall = &dto.ToolCall{
							ID:   chunk.Choices[0].Delta.ToolCalls[0].ID,
							Type: chunk.Choices[0].Delta.ToolCalls[0].Type,
							Function: dto.ToolFunction{
								Name:      chunk.Choices[0].Delta.ToolCalls[0].Function.Name,
								Arguments: "",
							},
						}
					}

					// Accumulate arguments
					if chunk.Choices[0].Delta.ToolCalls[0].Function.Arguments != "" {
						toolCallArguments += chunk.Choices[0].Delta.ToolCalls[0].Function.Arguments
					}
					continue
				}

				// If this is the last chunk of tool call, send the complete tool call to client
				if chunk.Choices[0].FinishReason == "tool_calls" {
					var toolCalls []dto.ToolCall
					if toolCall != nil {
						toolCall.Function.Arguments = toolCallArguments
						toolCalls = append(toolCalls, *toolCall)
					}
					response := dto.WebSocketResponse{
						Type:      dto.MsgTypeToolCall,
						Content:   "",
						Sender:    "bot",
						Timestamp: time.Now().UnixMilli(),
						ID:        uuid.New().String(),
						ToolCalls: toolCalls,
					}
					if err := conn.WriteJSON(response); err != nil {
						log.Printf("[WS] Error writing tool call message - SessionID: %s, Error: %v", sessionId, err)
						break
					}

					s, err := json.Marshal(toolCalls)
					if err != nil {
						log.Printf("[WS] Failed to marshal tool calls - SessionID: %s, Error: %v", sessionId, err)
						continue
					}
					msg := &database2.Message{
						ID:        uuid.New().String(),
						SessionID: sessionId,
						Content:   "",
						Sender:    "bot",
						Timestamp: time.Now(),
						ToolCalls: string(s),
					}
					if err := db.SaveMessage(c.Request.Context(), msg); err != nil {
						log.Printf("[WS] Failed to save tool call message - SessionID: %s, Error: %v", sessionId, err)
					}
					continue
				}

				// Handle regular content
				if chunk.Choices[0].Delta.Content != "" {
					responseContent += chunk.Choices[0].Delta.Content
					response := dto.WebSocketResponse{
						Type:      dto.MsgTypeStream,
						Content:   chunk.Choices[0].Delta.Content,
						Sender:    "bot",
						Timestamp: time.Now().UnixMilli(),
						ID:        uuid.New().String(),
					}
					if err := conn.WriteJSON(response); err != nil {
						log.Printf("[WS] Error writing stream message - SessionID: %s, Error: %v", sessionId, err)
						break
					}
				}

				// If this is the last chunk, save the complete message
				if chunk.Choices[0].FinishReason == "stop" {
					// Save the complete message to database
					dbMessage := &database2.Message{
						ID:        uuid.New().String(),
						SessionID: sessionId,
						Content:   responseContent,
						Sender:    "bot",
						Timestamp: time.Now(),
					}
					if err := db.SaveMessage(c.Request.Context(), dbMessage); err != nil {
						log.Printf("[WS] Failed to save bot message - SessionID: %s, Error: %v", sessionId, err)
					}
				}
			}

		case dto.MsgTypeToolResult:
			s, err := json.Marshal(message.ToolResult)
			if err != nil {
				log.Printf("[WS] Failed to marshal tool result - SessionID: %s, Error: %v", sessionId, err)
				continue
			}
			msg := &database2.Message{
				ID:         uuid.New().String(),
				SessionID:  sessionId,
				Content:    "",
				Sender:     "user",
				Timestamp:  time.Now(),
				ToolResult: string(s),
			}
			if err := db.SaveMessage(c.Request.Context(), msg); err != nil {
				log.Printf("[WS] Failed to save tool result message - SessionID: %s, Error: %v", sessionId, err)
			}

			// Get conversation history from database
			messages, err := db.GetMessages(c.Request.Context(), sessionId)
			if err != nil {
				log.Printf("[WS] Failed to get conversation history - SessionID: %s, Error: %v", sessionId, err)
				continue
			}

			// Convert messages to OpenAI format
			openaiMessages := make([]openaiGo.ChatCompletionMessageParamUnion, len(messages))
			for i, msg := range messages {
				if msg.Sender == "bot" {
					if msg.ToolCalls != "" {
						// For bot messages with tool calls, use OfAssistant type with ToolCalls field
						var toolCalls []dto.ToolCall
						if err := json.Unmarshal([]byte(msg.ToolCalls), &toolCalls); err != nil {
							log.Printf("[WS] Failed to unmarshal tool calls - SessionID: %s, Error: %v", sessionId, err)
							continue
						}
						openaiToolCalls := make([]openaiGo.ChatCompletionMessageToolCallParam, len(toolCalls))
						for j, tc := range toolCalls {
							openaiToolCalls[j] = openaiGo.ChatCompletionMessageToolCallParam{
								ID:   tc.ID,
								Type: "function",
								Function: openaiGo.ChatCompletionMessageToolCallFunctionParam{
									Name:      tc.Function.Name,
									Arguments: tc.Function.Arguments,
								},
							}
						}
						openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
							OfAssistant: &openaiGo.ChatCompletionAssistantMessageParam{
								ToolCalls: openaiToolCalls,
							},
						}
					} else {
						// For regular bot messages, use OfAssistant type
						openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
							OfAssistant: &openaiGo.ChatCompletionAssistantMessageParam{
								Content: openaiGo.ChatCompletionAssistantMessageParamContentUnion{
									OfString: openaiGo.String(msg.Content),
								},
							},
						}
					}
				} else {
					if msg.ToolResult != "" {
						// For user messages with tool results, use OfTool type
						var toolResult dto.ToolResult
						if err := json.Unmarshal([]byte(msg.ToolResult), &toolResult); err != nil {
							log.Printf("[WS] Failed to unmarshal tool result - SessionID: %s, Error: %v", sessionId, err)
							continue
						}
						openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
							OfTool: &openaiGo.ChatCompletionToolMessageParam{
								ToolCallID: toolResult.ToolCallID,
								Content: openaiGo.ChatCompletionToolMessageParamContentUnion{
									OfString: openaiGo.String(toolResult.Result),
								},
							},
						}
					} else {
						// For regular user messages, use OfUser type
						openaiMessages[i] = openaiGo.ChatCompletionMessageParamUnion{
							OfUser: &openaiGo.ChatCompletionUserMessageParam{
								Content: openaiGo.ChatCompletionUserMessageParamContentUnion{
									OfString: openaiGo.String(msg.Content),
								},
							},
						}
					}
				}
			}

			// Get streaming response from OpenAI
			stream, err := openaiClient.ChatCompletionStream(c.Request.Context(), openaiMessages, nil)
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
					response := dto.WebSocketResponse{
						Type:      dto.MsgTypeStream,
						Content:   chunk.Choices[0].Delta.Content,
						Sender:    "bot",
						Timestamp: time.Now().UnixMilli(),
						ID:        uuid.New().String(),
					}
					if err := conn.WriteJSON(response); err != nil {
						log.Printf("[WS] Error writing stream message - SessionID: %s, Error: %v", sessionId, err)
						break
					}
				}

				// If this is the last chunk, save the complete message
				if chunk.Choices[0].FinishReason == "stop" {
					// Save the complete message to database
					dbMessage := &database2.Message{
						ID:        uuid.New().String(),
						SessionID: sessionId,
						Content:   responseContent,
						Sender:    "bot",
						Timestamp: time.Now(),
					}
					if err := db.SaveMessage(c.Request.Context(), dbMessage); err != nil {
						log.Printf("[WS] Failed to save bot message - SessionID: %s, Error: %v", sessionId, err)
					}
				}
			}
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
	var cfg config2.MCPConfig
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

	// Send reload signal to gateway
	if err := sendReloadSignal(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to reload gateway: " + err.Error(),
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
	cfg, err := config2.LoadConfig[config2.APIServerConfig]("configs/apiserver.yaml")
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
		var cfg config2.MCPConfig
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
