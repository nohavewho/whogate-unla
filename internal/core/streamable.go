package core

import (
	"encoding/json"
	"fmt"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/session"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// handleMCP handles MCP connections
func (s *Server) handleMCP(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodOptions:
		c.Status(http.StatusOK)
		return

	case http.MethodGet:
		s.handleGet(c)
	case http.MethodPost:
		s.handlePost(c)
		return
	case http.MethodDelete:
		s.handleDelete(c)
		return

	default:
		c.Header("Allow", "GET, POST, DELETE")
		s.sendProtocolError(c, nil, mcp.ErrorCodeConnectionClosed, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// handleGet handles GET requests for SSE stream
func (s *Server) handleGet(c *gin.Context) {
	// Check Accept header for text/event-stream
	acceptHeader := c.GetHeader("Accept")
	if !strings.Contains(acceptHeader, "text/event-stream") {
		s.sendProtocolError(c, nil, mcp.ErrorCodeInvalidRequest, "Not Acceptable: Client must accept text/event-stream", http.StatusNotAcceptable)
		return
	}

	conn := s.getSession(c)
	if conn == nil {
		return
	}

	// TODO: replay events according to the last-event-id in request headers.

	// TODO: only support one sse stream per session, we can detect it when sub the redis topic

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Mcp-Session-Id", conn.Meta().ID)
	c.Writer.Flush()

	for {
		select {
		case event := <-conn.EventQueue():
			switch event.Event {
			case "message":
				_, err := fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", event.Data)
				if err != nil {
					s.logger.Error("failed to send SSE message", zap.Error(err))
				}
			}
			_, _ = fmt.Fprint(c.Writer, event)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		case <-s.shutdownCh:
			return
		}
	}
}

// handlePost handles POST requests containing JSON-RPC messages
func (s *Server) handlePost(c *gin.Context) {
	// Validate Accept header
	acceptHeader := c.GetHeader("Accept")
	if !strings.Contains(acceptHeader, "application/json") || !strings.Contains(acceptHeader, "text/event-stream") {
		s.sendProtocolError(c, nil, mcp.ErrorCodeConnectionClosed, "Not Acceptable: Client must accept both application/json and text/event-stream", http.StatusNotAcceptable)
		return
	}

	// Validate Content-Type header
	contentType := c.GetHeader("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		s.sendProtocolError(c, nil, mcp.ErrorCodeConnectionClosed, "Unsupported Media Type: Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// TODO: support batch messages
	var req mcp.JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendProtocolError(c, nil, mcp.ErrorCodeParseError, "Invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	sessionID := c.GetHeader("Mcp-Session-Id")

	var (
		conn session.Connection
		err  error
	)
	if req.Method == mcp.Initialize {
		if sessionID != "" {
			// confirm if it's registered
			conn, err = s.sessions.Get(c.Request.Context(), sessionID)
			if err != nil {
				s.sendProtocolError(c, req.Id, mcp.ErrorCodeInternalError, "Failed to get session", http.StatusInternalServerError)
				return
			}
			if conn != nil {
				s.sendProtocolError(c, req.Id, mcp.ErrorCodeInvalidRequest, "Invalid Request: Server already initialized", http.StatusBadRequest)
				return
			}
		} else {
			sessionID = uuid.New().String()
			c.Header("Mcp-Session-Id", sessionID)

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
				s.sendProtocolError(c, req.Id, mcp.ErrorCodeInternalError, "Failed to create session", http.StatusInternalServerError)
				return
			}
		}
	} else {
		conn, err = s.sessions.Get(c.Request.Context(), sessionID)
		if err != nil {
			s.sendProtocolError(c, req.Id, mcp.ErrorCodeInvalidRequest, "Invalid Request: Session not found", http.StatusBadRequest)
			return
		}
	}

	s.handleMCPRequest(c, req, conn)
}

// handleDelete handles DELETE requests to terminate sessions
func (s *Server) handleDelete(c *gin.Context) {
	conn := s.getSession(c)
	if conn == nil {
		return
	}

	err := s.sessions.Unregister(c.Request.Context(), conn.Meta().ID)
	if err != nil {
		s.sendProtocolError(c, conn.Meta().ID, mcp.ErrorCodeInternalError, "Failed to terminate session", http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (s *Server) handleMCPRequest(c *gin.Context, req mcp.JSONRPCRequest, conn session.Connection) {
	// Process the request based on its method
	switch req.Method {
	case mcp.Initialize:
		// Handle initialization request
		var params mcp.InitializeRequestParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendProtocolError(c, req.Id, mcp.ErrorCodeInvalidParams, fmt.Sprintf("invalid initialize parameters: %v", err), http.StatusBadRequest)
			return
		}

		s.sendSuccessResponse(c, conn, req, mcp.InitializedResult{
			ProtocolVersion: mcp.LatestProtocolVersion,
			Capabilities: mcp.ServerCapabilitiesSchema{
				Logging: mcp.LoggingCapabilitySchema{},
				Tools: mcp.ToolsCapabilitySchema{
					ListChanged: true,
				},
			},
			ServerInfo: mcp.ImplementationSchema{
				Name:    "mcp-gateway",
				Version: version.Get(),
			},
		}, false)
		return

	case mcp.NotificationInitialized:
		c.Status(http.StatusAccepted)
		return

	case mcp.ToolsList:
		// Get tools for this prefix
		tools, ok := s.prefixToTools[conn.Meta().Prefix]
		if !ok {
			tools = []mcp.ToolSchema{}
		}

		s.sendSuccessResponse(c, conn, req, mcp.ListToolsResult{
			Tools: tools,
		}, false)
		return

	case mcp.ToolsCall:
		// Handle tool call request
		var params mcp.CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendProtocolError(c, req.Id, mcp.ErrorCodeInvalidParams, fmt.Sprintf("invalid tool call parameters: %v", err), http.StatusBadRequest)
			return
		}

		// Find the tool in the precomputed map
		tool, exists := s.toolMap[params.Name]
		if !exists {
			s.sendProtocolError(c, req.Id, mcp.ErrorCodeMethodNotFound, "Tool not found", http.StatusNotFound)
			return
		}

		// Convert arguments to map[string]any
		var args map[string]any
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			s.sendProtocolError(c, req.Id, mcp.ErrorCodeInvalidParams, "Invalid tool arguments", http.StatusBadRequest)
			return
		}

		serverCfg, ok := s.prefixToServerConfig[conn.Meta().Prefix]
		if !ok {
			s.sendProtocolError(c, req.Id, mcp.ErrorCodeInternalError, "Server config not found", http.StatusInternalServerError)
			return
		}

		// Execute the tool
		result, err := s.executeTool(tool, args, c.Request, serverCfg.Config)
		if err != nil {
			s.logger.Error("failed to execute tool", zap.Error(err))
			// For tool execution errors, return them in the result with isError=true
			s.sendToolExecutionError(c, conn, req, err, false)
			return
		}

		s.sendSuccessResponse(c, conn, req, mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: result,
				},
			},
			IsError: false,
		}, false)
		return

	default:
		s.sendProtocolError(c, req.Id, mcp.ErrorCodeMethodNotFound, "Method not found", http.StatusNotFound)
		return
	}
}

func (s *Server) getSession(c *gin.Context) session.Connection {
	sessionID := c.GetHeader("Mcp-Session-Id")
	if sessionID == "" {
		s.sendProtocolError(c, nil, mcp.ErrorCodeConnectionClosed, "Bad Request: Mcp-Session-Id header is required", http.StatusBadRequest)
		return nil
	}
	conn, err := s.sessions.Get(c.Request.Context(), sessionID)
	if err != nil {
		s.sendProtocolError(c, nil, mcp.ErrorCodeRequestTimeout, "Session not found", http.StatusNotFound)
		return nil
	}
	return conn
}
