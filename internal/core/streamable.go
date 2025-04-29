package core

import (
	"encoding/json"
	"fmt"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/session"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleMCP handles MCP connections
func (s *Server) handleMCP(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodOptions:
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Status(http.StatusOK)
		return

	case http.MethodGet:
		prefix := strings.TrimSuffix(c.Request.URL.Path, "/mcp")
		if prefix == "" {
			prefix = "/"
		}

		sessionID := c.GetHeader("Mcp-Session-Id")
		if sessionID == "" {
			sessionID = uuid.New().String()
			c.Header("Mcp-Session-Id", sessionID)
		}

		// TODO: complete
	case http.MethodPost:
		var req mcp.JSONRPCRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		sessionID := c.GetHeader("Mcp-Session-Id")
		if sessionID == "" {
			sessionID = uuid.New().String()
			c.Header("Mcp-Session-Id", sessionID)
		}

		var (
			conn session.Connection
			err  error
		)
		if req.Method == mcp.Initialize {
			prefix := strings.TrimSuffix(c.Request.URL.Path, "/mcp")
			if prefix == "" {
				prefix = "/"
			}

			meta := &session.Meta{
				ID:        sessionID,
				CreatedAt: time.Now(),
				Prefix:    prefix,
				Type:      "streamable",
			}
			conn, err = s.sessions.Register(c.Request.Context(), meta)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to create session")
				return
			}
		} else {
			conn, err = s.sessions.Get(c.Request.Context(), sessionID)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to get session")
				return
			}
		}

		if err := s.handleMCPRequest(c, req, conn); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		return
	case http.MethodDelete:
		sessionID := c.GetHeader("Mcp-Session-Id")
		if sessionID == "" {
			c.String(http.StatusBadRequest, "Missing session ID")
			return
		}

		err := s.sessions.Unregister(c.Request.Context(), sessionID)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid or expired session")
			return
		}
		c.String(http.StatusOK, "Session terminated")

	default:
		c.String(http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleMCPRequest(c *gin.Context, req mcp.JSONRPCRequest, conn session.Connection) error {
	// Process the request based on its method
	switch req.Method {
	case mcp.Initialize:
		// Handle initialization request
		var params mcp.InitializeRequestParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendErrorResponse(c, conn, req, fmt.Sprintf("invalid initialize parameters: %v", err))
			return nil
		}

		response := mcp.InitializeResult{
			JSONRPCBaseResult: mcp.JSONRPCBaseResult{
				JSONRPC: mcp.JSPNRPCVersion,
				ID:      req.Id,
			},
			Result: mcp.InitializedResult{
				ProtocolVersion: mcp.LatestProtocolVersion,
				Capabilities: mcp.ServerCapabilitiesSchema{
					Logging: mcp.LoggingCapabilitySchema{},
					Tools: mcp.ToolsCapabilitySchema{
						ListChanged: true,
					},
				},
				ServerInfo: mcp.ImplementationSchema{
					Name:    "mcp-gateway",
					Version: "1.0.0",
				},
			},
		}

		eventData, err := json.Marshal(response)
		if err != nil {
			s.sendErrorResponse(c, conn, req, fmt.Sprintf("failed to marshal response: %v", err))
			return nil
		}

		// Send response directly
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		// TODO: maybe we need to send this to the session store too.
		c.String(http.StatusOK, fmt.Sprintf("event: message\ndata: %s\n\n", eventData))
		return nil

	case mcp.NotificationInitialized:
		// Set response headers
		c.Header("X-Powered-By", "Express")
		c.Header("Connection", "keep-alive")
		c.Header("Keep-Alive", "timeout=5")
		c.Header("Transfer-Encoding", "chunked")

		// Return accepted response
		c.String(http.StatusAccepted, mcp.Accepted)
		return nil

	case mcp.ToolsList:
		// Get tools for this prefix
		tools, ok := s.prefixToTools[conn.Meta().Prefix]
		if !ok {
			tools = []mcp.ToolSchema{}
		}

		response := mcp.JSONRPCResponse{
			JSONRPCBaseResult: mcp.JSONRPCBaseResult{
				JSONRPC: mcp.JSPNRPCVersion,
				ID:      req.Id,
			},
			Result: mcp.ListToolsResult{
				Tools: tools,
			},
		}

		eventData, err := json.Marshal(response)
		if err != nil {
			s.sendErrorResponse(c, conn, req, fmt.Sprintf("failed to marshal response: %v", err))
			return nil
		}

		// Send response directly
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.String(http.StatusOK, fmt.Sprintf("event: message\ndata: %s\n\n", eventData))
		return nil

	case mcp.ToolsCall:
		// Handle tool call request
		var params mcp.CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendErrorResponse(c, conn, req, fmt.Sprintf("invalid tool call parameters: %v", err))
			return nil
		}

		// Find the tool in the precomputed map
		tool, exists := s.toolMap[params.Name]
		if !exists {
			s.sendErrorResponse(c, conn, req, "Tool not found")
			return nil
		}

		// Convert arguments to map[string]any
		var args map[string]any
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			s.sendErrorResponse(c, conn, req, "Invalid tool arguments")
			return nil
		}

		serverCfg, ok := s.prefixToServerConfig[conn.Meta().Prefix]
		if !ok {
			s.sendErrorResponse(c, conn, req, "Server config not found")
			return nil
		}

		// Execute the tool
		result, err := s.executeTool(tool, args, c.Request, serverCfg.Config)
		if err != nil {
			s.logger.Error("failed to execute tool", zap.Error(err))
			s.sendErrorResponse(c, conn, req, fmt.Sprintf("Error: %s", err.Error()))
			return nil
		}

		response := mcp.JSONRPCResponse{
			JSONRPCBaseResult: mcp.JSONRPCBaseResult{
				JSONRPC: mcp.JSPNRPCVersion,
				ID:      req.Id,
			},
			Result: mcp.CallToolResult{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: result,
					},
				},
			},
		}

		eventData, err := json.Marshal(response)
		if err != nil {
			s.sendErrorResponse(c, conn, req, fmt.Sprintf("failed to marshal response: %v", err))
			return nil
		}

		// Send response directly
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.String(http.StatusOK, fmt.Sprintf("event: message\ndata: %s\n\n", eventData))
		return nil

	default:
		s.sendErrorResponse(c, conn, req, fmt.Sprintf("unknown method: %s", req.Method))
		return nil
	}
}
