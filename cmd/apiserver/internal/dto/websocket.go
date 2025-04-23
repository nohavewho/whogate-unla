package dto

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	ID        string `json:"id"`
	Tools     []Tool `json:"tools,omitempty"`
}

// WebSocketResponse represents a response sent over WebSocket
type WebSocketResponse struct {
	Type      string             `json:"type"`
	Content   string             `json:"content"`
	Sender    string             `json:"sender"`
	Timestamp int64              `json:"timestamp"`
	ID        string             `json:"id"`
	ToolCalls []ToolCallResponse `json:"tool_calls,omitempty"`
}

// ToolParameters represents the parameters of a tool
type ToolParameters struct {
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

// ToolFunction represents the function details of a tool call
type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool represents a tool that can be called by the LLM
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

// ToolCall represents a tool call from the LLM
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolCallResponse represents the response for a tool call
type ToolCallResponse struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ResponseType represents the type of WebSocket response
const (
	ResponseTypeMessage  = "message"
	ResponseTypeStream   = "stream"
	ResponseTypeToolCall = "tool_call"
)
