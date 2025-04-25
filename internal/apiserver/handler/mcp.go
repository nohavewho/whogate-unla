package handler

import (
	"fmt"
	"github.com/mcp-ecosystem/mcp-gateway/internal/apiserver/database"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type MCP struct {
	db database.Database
}

func NewMCP(db database.Database) *MCP {
	return &MCP{db: db}
}

func (h *MCP) HandleMCPServerUpdate(c *gin.Context) {
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

func (h *MCP) HandleGetMCPServers(c *gin.Context) {
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

func (h *MCP) HandleMCPServerCreate(c *gin.Context) {
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

func (h *MCP) HandleMCPServerDelete(c *gin.Context) {
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

func (h *MCP) HandleMCPServerSync(c *gin.Context) {
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

// Helper functions
func getConfigPath() string {
	// 1. Check environment variable
	if envPath := os.Getenv("CONFIG_DIR"); envPath != "" {
		return envPath
	}

	// 2. Default to APPDATA/.mcp/gateway
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
