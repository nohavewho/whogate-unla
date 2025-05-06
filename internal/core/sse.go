package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/session"

	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleSSE handles SSE connections
func (s *Server) handleSSE(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Get the prefix from the request path
	prefix := strings.TrimSuffix(c.Request.URL.Path, "/sse")
	if prefix == "" {
		prefix = "/"
	}

	sessionID := uuid.New().String()
	meta := &session.Meta{
		ID:        sessionID,
		CreatedAt: time.Now(),
		Prefix:    prefix,
		Type:      "sse",
		Extra:     nil,
	}
	conn, err := s.sessions.Register(c.Request.Context(), meta)
	if err != nil {
		s.sendProtocolError(c, sessionID, "Failed to create SSE connection", http.StatusInternalServerError, mcp.ErrorCodeInternalError)
		return
	}

	// Send the initial endpoint event
	_, err = fmt.Fprintf(c.Writer, "event: endpoint\ndata: %s\n\n",
		fmt.Sprintf("%s/message?sessionId=%s", strings.TrimSuffix(c.Request.URL.Path, "/sse"), meta.ID))
	if err != nil {
		s.sendProtocolError(c, sessionID, "Failed to initialize SSE connection", http.StatusInternalServerError, mcp.ErrorCodeInternalError)
		return
	}
	c.Writer.Flush()

	// Main event loop
	for {
		select {
		case event := <-conn.EventQueue():
			switch event.Event {
			case "message":
				_, err = fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", event.Data)
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

// sendErrorResponse sends an error response through SSE channel and returns Accepted status
func (s *Server) sendErrorResponse(c *gin.Context, conn session.Connection, req mcp.JSONRPCRequest, errorMsg string) {
	response := mcp.JSONRPCErrorSchema{
		JSONRPCBaseResult: mcp.JSONRPCBaseResult{
			JSONRPC: mcp.JSPNRPCVersion,
			ID:      req.Id,
		},
		Error: mcp.JSONRPCError{
			Code:    mcp.ErrorCodeInternalError,
			Message: errorMsg,
		},
	}
	eventData, err := json.Marshal(response)
	if err != nil {
		c.String(http.StatusAccepted, mcp.Accepted)
		return
	}
	err = conn.Send(c.Request.Context(), &session.Message{
		Event: "message",
		Data:  eventData,
	})
	if err != nil {
		c.String(http.StatusAccepted, mcp.Accepted)
		return
	}
	c.String(http.StatusAccepted, mcp.Accepted)
}

// handleMessage processes incoming JSON-RPC messages
func (s *Server) handleMessage(c *gin.Context) {
	s.logger.Debug("Received message", zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path))

	// Get the session ID from the query parameter
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		c.String(http.StatusNotFound, "Missing sessionId parameter")
		s.sendProtocolError(c, nil, "Missing sessionId parameter", http.StatusBadRequest, mcp.ErrorCodeInvalidRequest)
		return
	}

	conn, err := s.sessions.Get(c.Request.Context(), sessionId)
	if err != nil {
		c.String(http.StatusNotFound, "Session not found")
		return
	}
	s.handlePostMessage(c, conn)
}

func (s *Server) handlePostMessage(c *gin.Context, conn session.Connection) {
	if conn == nil {
		c.String(http.StatusInternalServerError, "SSE connection not established")
		return
	}

	// Validate Content-Type header
	contentType := c.GetHeader("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		c.String(http.StatusNotAcceptable, "Unsupported Media Type: Content-Type must be application/json")
		return
	}

	// TODO: support auth

	// Parse the JSON-RPC message
	var req mcp.JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid message")
		return
	}

	switch req.Method {
	case mcp.NotificationInitialized:
		s.sendAcceptedResponse(c)
	case mcp.Initialize:
		var params mcp.InitializeRequestParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendProtocolError(c, req.Id, "Invalid initialize parameters", http.StatusBadRequest, mcp.ErrorCodeInvalidParams)
			return
		}

		result := mcp.InitializedResult{
			ProtocolVersion: mcp.LatestProtocolVersion,
			ServerInfo: mcp.ImplementationSchema{
				Name:    "mcp-gateway",
				Version: "0.1.0",
			},
		}
		s.sendSuccessResponse(c, conn, req, result, true)
	case mcp.ToolsList:
		// Get tools for this prefix
		tools, ok := s.state.prefixToTools[conn.Meta().Prefix]
		if !ok {
			tools = []mcp.ToolSchema{} // Return empty list if prefix not found
		}

		result := mcp.ListToolsResult{
			Tools: tools,
		}
		s.sendSuccessResponse(c, conn, req, result, true)
	case mcp.ToolsCall:
		// Execute the tool and return the result
		var params mcp.CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendProtocolError(c, req.Id, "Invalid tool call parameters", http.StatusBadRequest, mcp.ErrorCodeInvalidParams)
			return
		}

		// Find the tool in the precomputed map
		tool, exists := s.state.toolMap[params.Name]
		if !exists {
			s.sendProtocolError(c, req.Id, "Tool not found", http.StatusNotFound, mcp.ErrorCodeMethodNotFound)
			return
		}

		// Convert arguments to map[string]any
		var args map[string]any
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			s.sendProtocolError(c, req.Id, "Invalid tool arguments", http.StatusBadRequest, mcp.ErrorCodeInvalidParams)
			return
		}

		// Get server configuration
		serverCfg, ok := s.state.prefixToServerConfig[conn.Meta().Prefix]
		if !ok {
			s.sendProtocolError(c, req.Id, "Server configuration not found", http.StatusInternalServerError, mcp.ErrorCodeInternalError)
			return
		}

		// Execute the tool
		result, err := s.executeTool(tool, args, c.Request, serverCfg.Config)
		if err != nil {
			s.sendToolExecutionError(c, conn, req, err, true)
			return
		}

		// Send the result
		toolResult := mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: result,
				},
			},
			IsError: false,
		}
		s.sendSuccessResponse(c, conn, req, toolResult, true)
	default:
		s.sendProtocolError(c, req.Id, "Unknown method", http.StatusNotFound, mcp.ErrorCodeMethodNotFound)
	}
}
