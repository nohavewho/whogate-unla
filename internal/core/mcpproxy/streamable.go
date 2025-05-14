package mcpproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/template"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"
)

// FetchStreamableToolList fetches the list of available tools from a Streamable HTTP backend
func FetchStreamableToolList(ctx context.Context, mcpProxyCfg config.MCPServerConfig) ([]mcp.ToolSchema, error) {
	// Create Streamable HTTP transport
	streamableTransport, err := transport.NewStreamableHTTP(mcpProxyCfg.URL)
	if err != nil {
		return []mcp.ToolSchema{}, fmt.Errorf("failed to create Streamable HTTP transport: %w", err)
	}

	// Start the transport
	if err := streamableTransport.Start(ctx); err != nil {
		return []mcp.ToolSchema{}, fmt.Errorf("failed to start Streamable HTTP transport: %w", err)
	}

	// Create client with the transport
	c := client.NewClient(streamableTransport)
	defer c.Close()

	// Initialize the client
	initRequest := mcpgo.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpgo.Implementation{
		Name:    "mcp-gateway",
		Version: version.Get(),
	}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		return []mcp.ToolSchema{}, fmt.Errorf("failed to initialize Streamable HTTP client: %w", err)
	}

	// List available tools
	toolsResult, err := c.ListTools(ctx, mcpgo.ListToolsRequest{})
	if err != nil {
		return []mcp.ToolSchema{}, fmt.Errorf("failed to list tools: %w", err)
	}

	// Convert from mcpgo.Tool to mcp.ToolSchema
	tools := make([]mcp.ToolSchema, len(toolsResult.Tools))
	for i, schema := range toolsResult.Tools {
		// Create local mcp package ToolInputSchema
		inputSchema := mcp.ToolInputSchema{
			Type:       "object",
			Properties: make(map[string]any),
		}

		// Convert mcpgo InputSchema to local mcp format
		rawSchema, err := json.Marshal(schema.InputSchema)
		if err == nil {
			// Parse schema properties
			var schemaMap map[string]interface{}
			if err := json.Unmarshal(rawSchema, &schemaMap); err == nil {
				if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
					inputSchema.Properties = properties
				}
				if typ, ok := schemaMap["type"].(string); ok {
					inputSchema.Type = typ
				}
				if required, ok := schemaMap["required"].([]interface{}); ok {
					reqStrings := make([]string, len(required))
					for j, r := range required {
						if rStr, ok := r.(string); ok {
							reqStrings[j] = rStr
						}
					}
					inputSchema.Required = reqStrings
				}
			}
		}

		tools[i] = mcp.ToolSchema{
			Name:        schema.Name,
			Description: schema.Description,
			InputSchema: inputSchema,
		}
	}

	return tools, nil
}

// InvokeStreamableTool handles tool invocation for Streamable HTTP MCP protocol
func InvokeStreamableTool(c *gin.Context, mcpProxyCfg config.MCPServerConfig, params mcp.CallToolParams) (*mcp.CallToolResult, error) {
	// Convert arguments to map[string]any
	var args map[string]any
	if err := json.Unmarshal(params.Arguments, &args); err != nil {
		return nil, fmt.Errorf("invalid tool arguments: %w", err)
	}

	// Prepare template context for environment variables
	tmplCtx, err := template.PrepareTemplateContext(args, c.Request, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare template context: %w", err)
	}

	// Process environment variables with templates
	renderedClientEnv := make(map[string]string)
	for k, v := range mcpProxyCfg.Env {
		rendered, err := template.RenderTemplate(v, tmplCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to render env template: %w", err)
		}
		renderedClientEnv[k] = rendered
	}

	// Create HTTP headers for the transport
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	sessionID := c.GetHeader(mcp.HeaderMcpSessionID)
	if sessionID != "" {
		headers[mcp.HeaderMcpSessionID] = sessionID
	}

	// Create Streamable HTTP transport with headers
	streamableTransport, err := transport.NewStreamableHTTP(
		mcpProxyCfg.URL,
		transport.WithHTTPHeaders(headers),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Streamable HTTP transport: %w", err)
	}

	// Start the transport
	if err := streamableTransport.Start(c.Request.Context()); err != nil {
		return nil, fmt.Errorf("failed to start Streamable HTTP transport: %w", err)
	}

	// Create client with the transport
	mcpClient := client.NewClient(streamableTransport)
	defer mcpClient.Close()

	// Initialize the client
	initRequest := mcpgo.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpgo.Implementation{
		Name:    "mcp-gateway",
		Version: version.Get(),
	}

	_, err = mcpClient.Initialize(c.Request.Context(), initRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Streamable HTTP client: %w", err)
	}

	// Prepare tool call request parameters
	toolCallRequestParams := make(map[string]interface{})
	if err := json.Unmarshal(params.Arguments, &toolCallRequestParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	// Call tool
	callRequest := mcpgo.CallToolRequest{}
	callRequest.Params.Name = params.Name
	callRequest.Params.Arguments = toolCallRequestParams

	mcpgoResult, err := mcpClient.CallTool(c.Request.Context(), callRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	// Convert mcp-go result to local mcp format
	return convertMCPGoResult(mcpgoResult), nil
}
