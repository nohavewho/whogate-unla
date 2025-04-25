package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// sseSession represents an active SSE connection
type sseSession struct {
	writer              http.ResponseWriter
	flusher             http.Flusher
	done                chan struct{}
	eventQueue          chan string
	sessionID           string
	notificationChannel chan mcp.JSONRPCNotification
	initialized         bool
}

func (s *sseSession) SessionID() string {
	return s.sessionID
}

func (s *sseSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return s.notificationChannel
}

func (s *sseSession) Initialize() {
	s.initialized = true
}

func (s *sseSession) Initialized() bool {
	return s.initialized
}

// Close closes the SSE session
func (s *sseSession) Close() {
	close(s.done)
}

// handleSSE handles SSE connections
func (s *Server) handleSSE(c *gin.Context) {
	w := c.Writer
	r := c.Request

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	// Get the prefix from the request path
	prefix := strings.TrimSuffix(c.Request.URL.Path, "/sse")
	if prefix == "" {
		prefix = "/"
	}

	sessionID := uuid.New().String()
	session := &sseSession{
		writer:              w,
		flusher:             flusher,
		done:                make(chan struct{}),
		eventQueue:          make(chan string, 100),
		sessionID:           sessionID,
		notificationChannel: make(chan mcp.JSONRPCNotification, 100),
	}

	// Store session to prefix mapping
	s.sessionToPrefix.Store(sessionID, prefix)

	s.sessions.Store(sessionID, session)
	defer s.sessions.Delete(sessionID)
	defer s.sessionToPrefix.Delete(sessionID)

	// Start notification handler for this session
	go func() {
		for {
			select {
			case notification := <-session.notificationChannel:
				eventData, err := json.Marshal(notification)
				if err == nil {
					select {
					case session.eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", eventData):
						// Event queued successfully
					case <-session.done:
						return
					}
				}
			case <-session.done:
				return
			case <-r.Context().Done():
				return
			}
		}
	}()

	// Send the initial endpoint event
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: %s\r\n\r\n",
		fmt.Sprintf("%s/message?sessionId=%s", c.Request.URL.Path[:len(c.Request.URL.Path)-4], sessionID))
	flusher.Flush()

	// Main event loop
	for {
		select {
		case event := <-session.eventQueue:
			_, _ = fmt.Fprint(w, event)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// sendErrorResponse sends an error response through SSE channel and returns Accepted status
func (s *Server) sendErrorResponse(c *gin.Context, session *sseSession, req mcp.JSONRPCRequest, errorMsg string) {
	response := mcp.JSONRPCResponse{
		JSONRPCBaseResult: mcp.JSONRPCBaseResult{
			JSONRPC: mcp.JSPNRPCVersion,
			ID:      req.Id,
		},
		Result: mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				{
					Type: "text",
					Text: errorMsg,
				},
			},
		},
	}
	eventData, err := json.Marshal(response)
	if err != nil {
		c.String(http.StatusAccepted, mcp.Accepted)
		return
	}
	select {
	case session.eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", eventData):
		// Event queued successfully
	case <-session.done:
		// Session is closed
	}
	c.String(http.StatusAccepted, mcp.Accepted)
}

// handleMessage processes incoming JSON-RPC messages
func (s *Server) handleMessage(c *gin.Context) {
	// Parse the JSON-RPC message
	var req mcp.JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusAccepted, mcp.Accepted)
		return
	}

	// Get the session ID from the query parameter
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		c.String(http.StatusAccepted, mcp.Accepted)
		return
	}

	// Get the session from the map
	sessionI, ok := s.sessions.Load(sessionId)
	if !ok {
		c.String(http.StatusAccepted, mcp.Accepted)
		return
	}
	session := sessionI.(*sseSession)

	switch req.Method {
	case mcp.NotificationInitialized:
		// Do nothing, just acknowledge
		c.String(http.StatusAccepted, mcp.Accepted)
	case mcp.Initialize:
		var params mcp.InitializeRequestParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendErrorResponse(c, session, req, "Invalid initialize parameters")
			return
		}

		response := mcp.InitializeResult{
			JSONRPCBaseResult: mcp.JSONRPCBaseResult{
				JSONRPC: mcp.JSPNRPCVersion,
				ID:      req.Id,
			},
			Result: mcp.InitializedResult{
				ProtocolVersion: mcp.LatestProtocolVersion,
				ServerInfo: mcp.ImplementationSchema{
					Name:    "mcp-gateway",
					Version: "0.1.0",
				},
			},
		}

		// Send response via SSE
		eventData, err := json.Marshal(response)
		if err != nil {
			s.sendErrorResponse(c, session, req, "Failed to marshal response")
			return
		}
		select {
		case session.eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", eventData):
			// Event queued successfully
		case <-session.done:
			// Session is closed
		}
		// Also send HTTP response
		c.String(http.StatusAccepted, mcp.Accepted)
		// Initialize the session
		session.Initialize()
	case mcp.ToolsList:
		// Get the prefix for this session
		prefixI, ok := s.sessionToPrefix.Load(session.sessionID)
		if !ok {
			s.sendErrorResponse(c, session, req, "Session not found")
			return
		}
		prefix := prefixI.(string)

		// Get tools for this prefix
		tools, ok := s.prefixToTools[prefix]
		if !ok {
			tools = []mcp.ToolSchema{} // Return empty list if prefix not found
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

		// Send response via SSE
		eventData, err := json.Marshal(response)
		if err != nil {
			s.sendErrorResponse(c, session, req, "Failed to marshal response")
			return
		}
		select {
		case session.eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", eventData):
			// Event queued successfully
		case <-session.done:
			// Session is closed
		}
		// Also send HTTP response
		c.String(http.StatusAccepted, mcp.Accepted)
	case mcp.ToolsCall:
		// Execute the tool and return the result
		var params mcp.CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendErrorResponse(c, session, req, "Invalid tool call parameters")
			return
		}

		// Find the tool in the precomputed map
		tool, exists := s.toolMap[params.Name]
		if !exists {
			s.sendErrorResponse(c, session, req, "Tool not found")
			return
		}

		// Convert arguments to map[string]any
		var args map[string]any
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			s.sendErrorResponse(c, session, req, "Invalid tool arguments")
			return
		}

		prefixI, ok := s.sessionToPrefix.Load(session.sessionID)
		if !ok {
			s.sendErrorResponse(c, session, req, "Session not found")
			return
		}
		prefix := prefixI.(string)
		serverCfg, ok := s.prefixToServerConfig[prefix]

		// Execute the tool
		result, err := s.executeTool(tool, args, c.Request, serverCfg.Config)
		if err != nil {
			s.sendErrorResponse(c, session, req, fmt.Sprintf("Error: %s", err.Error()))
			return
		}

		// Send the result
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
		// Send response via SSE
		eventData, err := json.Marshal(response)
		if err != nil {
			s.sendErrorResponse(c, session, req, "Failed to marshal response")
			return
		}
		select {
		case session.eventQueue <- fmt.Sprintf("event: message\ndata: %s\n\n", eventData):
			// Event queued successfully
		case <-session.done:
			// Session is closed
		}
		// Also send HTTP response
		c.String(http.StatusAccepted, mcp.Accepted)
	default:
		s.sendErrorResponse(c, session, req, "Unknown method")
	}
}
