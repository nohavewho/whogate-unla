package mcp

import "encoding/json"

type (
	JSONRPCBaseResult struct {
		JSONRPC string `json:"jsonrpc"`
		ID      any    `json:"id"`
	}

	// BaseRequestParams represents the base parameters for all requests
	BaseRequestParams struct {
		// Meta information for the request
		Meta RequestMeta `json:"_meta"`
	}

	// RequestMeta represents the meta information for a request
	RequestMeta struct {
		// Progress token for tracking request progress
		// Can be string or number
		ProgressToken any `json:"progressToken"`
	}

	// JSONRPCRequest represents a JSON-RPC request that expects a response
	JSONRPCRequest struct {
		// JSONRPC version, must be "2.0"
		JSONRPC string `json:"jsonrpc"`
		// A uniquely identifying ID for a request in JSON-RPC
		Id any `json:"id"`
		// The method to be invoked
		Method string `json:"method"`
		// The parameters to be passed to the method
		Params json.RawMessage `json:"params"`
	}

	// JSONRPCResponse represents a JSON-RPC response
	JSONRPCResponse struct {
		JSONRPCBaseResult
		Result any `json:"result"`
	}

	// JSONRPCNotification represents a JSON-RPC notification
	JSONRPCNotification struct {
		JSONRPCBaseResult
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}

	// ToolSchema represents a tool definition
	ToolSchema struct {
		// The name of the tool
		Name string `json:"name"`
		// A human-readable description of the tool
		Description string `json:"description"`
		// A JSON Schema object defining the expected parameters for the tool
		InputSchema json.RawMessage `json:"inputSchema"`
	}

	// ListToolsResult represents the result of a tools/list request
	ListToolsResult struct {
		Tools []ToolSchema `json:"tools"`
	}

	// CallToolParams represents parameters for a tools/call request
	CallToolParams struct {
		BaseRequestParams
		// The name of the tool to call
		Name string `json:"name"`
		// The arguments to pass to the tool
		Arguments json.RawMessage `json:"arguments"`
	}

	// Content represents a content item in a tool call result
	Content struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	// TextContent represents a text content item
	TextContent struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	// CallToolResult represents the result of a tools/call request
	CallToolResult struct {
		Content []Content `json:"content"`
		IsError bool      `json:"isError"`
	}

	// ImplementationSchema describes the name and version of an MCP implementation
	ImplementationSchema struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	// ClientCapabilitiesSchema represents capabilities a client may support
	ClientCapabilitiesSchema struct {
		Experimental map[string]any        `json:"experimental"`
		Sampling     map[string]any        `json:"sampling"`
		Roots        RootsCapabilitySchema `json:"roots"`
	}

	// RootsCapabilitySchema represents roots-related capabilities
	RootsCapabilitySchema struct {
		ListChanged bool `json:"listChanged"`
	}

	// InitializeRequestParams represents parameters for initialize request
	InitializeRequestParams struct {
		BaseRequestParams
		// The latest version of the Model Context Protocol that the client supports
		ProtocolVersion string `json:"protocolVersion"`
		// Client capabilities
		Capabilities ClientCapabilitiesSchema `json:"capabilities"`
		// Client implementation information
		ClientInfo ImplementationSchema `json:"clientInfo"`
	}

	// InitializeRequestSchema represents an initialize request
	InitializeRequestSchema struct {
		JSONRPCRequest
	}

	// ServerCapabilitiesSchema represents capabilities a server may support
	ServerCapabilitiesSchema struct {
		Experimental ExperimentalCapabilitySchema `json:"experimental"`
		Logging      LoggingCapabilitySchema      `json:"logging"`
		Prompts      PromptsCapabilitySchema      `json:"prompts"`
		Resources    ResourcesCapabilitySchema    `json:"resources"`
		Tools        ToolsCapabilitySchema        `json:"tools"`
	}

	ExperimentalCapabilitySchema struct {
	}

	LoggingCapabilitySchema struct {
	}

	// PromptsCapabilitySchema represents prompts-related capabilities
	PromptsCapabilitySchema struct {
		ListChanged bool `json:"listChanged"`
	}

	// ResourcesCapabilitySchema represents resources-related capabilities
	ResourcesCapabilitySchema struct {
		Subscribe   bool `json:"subscribe"`
		ListChanged bool `json:"listChanged"`
	}

	// ToolsCapabilitySchema represents tools-related capabilities
	ToolsCapabilitySchema struct {
		ListChanged bool `json:"listChanged"`
	}

	// InitializeResult represents the result of an initialize request
	InitializeResult struct {
		JSONRPCBaseResult
		Result InitializedResult `json:"result"`
	}
	InitializedResult struct {
		// The version of the Model Context Protocol that the server wants to use
		ProtocolVersion string `json:"protocolVersion"`
		// Server capabilities
		Capabilities ServerCapabilitiesSchema `json:"capabilities"`
		// Server implementation information
		ServerInfo ImplementationSchema `json:"serverInfo"`
		// Instructions describing how to use the server and its features
		Instructions string `json:"instructions"`
	}

	// InitializedNotification represents an initialized notification
	InitializedNotification struct {
		JSONRPCRequest
	}

	// PingRequest represents a ping request
	PingRequest struct {
		JSONRPCRequest
	}
)

// NewInitializeRequest creates a new initialize request
func NewInitializeRequest(id int64, params InitializeRequestParams) InitializeRequestSchema {
	paramsBytes, _ := json.Marshal(params)
	return InitializeRequestSchema{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPC: JSPNRPCVersion,
			Id:      id,
			Method:  Initialize,
			Params:  paramsBytes,
		},
	}
}

// NewPingRequest creates a new ping request
func NewPingRequest(id int64) PingRequest {
	return PingRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPC: JSPNRPCVersion,
			Id:      id,
			Method:  Ping,
		},
	}
}

func NewJSONRPCBaseResult() JSONRPCBaseResult {
	return JSONRPCBaseResult{
		JSONRPC: JSPNRPCVersion,
		ID:      0,
	}
}

func (j JSONRPCBaseResult) WithID(id int) JSONRPCBaseResult {
	j.ID = id
	return j
}
