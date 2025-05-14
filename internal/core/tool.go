package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mcp-ecosystem/mcp-gateway/internal/mcp/session"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/utils"
	"github.com/mcp-ecosystem/mcp-gateway/pkg/version"

	"github.com/mcp-ecosystem/mcp-gateway/pkg/mcp"

	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/template"
)

// renderTemplate renders a template with the given context
func renderTemplate(tmpl string, ctx *template.Context) (string, error) {
	renderer := template.NewRenderer()
	return renderer.Render(tmpl, ctx)
}

// prepareTemplateContext prepares the template context with request and config data
func prepareTemplateContext(args map[string]any, request *http.Request, serverCfg map[string]string) (*template.Context, error) {
	tmplCtx := template.NewContext()
	tmplCtx.Args = preprocessArgs(args)

	// Process request headers
	for k, v := range request.Header {
		if len(v) > 0 {
			tmplCtx.Request.Headers[k] = v[0]
		}
	}

	// Process request querystring
	for k, v := range request.URL.Query() {
		if len(v) > 0 {
			tmplCtx.Request.Query[k] = v[0]
		}
	}

	// Process request cookies
	for _, cookie := range request.Cookies() {
		if cookie.Name != "" {
			tmplCtx.Request.Cookies[cookie.Name] = cookie.Value
		}
	}

	// Only process server config templates if serverCfg is provided
	if serverCfg != nil {
		// Process server config templates
		for k, v := range serverCfg {
			rendered, err := renderTemplate(v, tmplCtx)
			if err != nil {
				return nil, fmt.Errorf("failed to render config template: %w", err)
			}
			serverCfg[k] = rendered
		}
		tmplCtx.Config = serverCfg
	}

	return tmplCtx, nil
}

func preprocessArgs(args map[string]any) map[string]any {
	processed := make(map[string]any)

	for k, v := range args {
		switch val := v.(type) {
		case []any:
			ss, _ := json.Marshal(val)
			processed[k] = string(ss)
		case float64:
			// If the float64 equals its integer conversion, it's an integer
			if val == float64(int64(val)) {
				processed[k] = int64(val)
			} else {
				processed[k] = val
			}
		default:
			processed[k] = v
		}
	}
	return processed
}

// prepareRequest prepares the HTTP request with templates and arguments
func prepareRequest(tool *config.ToolConfig, tmplCtx *template.Context) (*http.Request, error) {
	// Process endpoint template
	endpoint, err := renderTemplate(tool.Endpoint, tmplCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to render endpoint template: %w", err)
	}

	// Process request body template
	var reqBody io.Reader
	if tool.RequestBody != "" {
		rendered, err := renderTemplate(tool.RequestBody, tmplCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to render request body template: %w", err)
		}
		reqBody = strings.NewReader(rendered)
	}

	req, err := http.NewRequest(tool.Method, endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Process header templates
	for k, v := range tool.Headers {
		rendered, err := renderTemplate(v, tmplCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to render header template: %w", err)
		}
		req.Header.Set(k, rendered)
	}

	return req, nil
}

// processArguments processes tool arguments and adds them to the request
func processArguments(req *http.Request, tool *config.ToolConfig, args map[string]any) {
	for _, arg := range tool.Args {
		value := fmt.Sprint(args[arg.Name])
		switch strings.ToLower(arg.Position) {
		case "header":
			req.Header.Set(arg.Name, value)
		case "query":
			q := req.URL.Query()
			q.Add(arg.Name, value)
			req.URL.RawQuery = q.Encode()
		case "form-data":
			var b bytes.Buffer
			writer := multipart.NewWriter(&b)

			if err := writer.WriteField(arg.Name, value); err != nil {
				continue
			}

			if err := writer.Close(); err != nil {
				continue
			}

			req.Body = io.NopCloser(&b)
			req.Header.Set("Content-Type", writer.FormDataContentType())
		}
	}
}

// preprocessResponseData processes response data to handle []any type
func preprocessResponseData(data map[string]any) map[string]any {
	processed := make(map[string]any)

	for k, v := range data {
		switch val := v.(type) {
		case []any:
			ss, _ := json.Marshal(val)
			processed[k] = string(ss)
		case map[string]any:
			processed[k] = preprocessResponseData(val)
		default:
			processed[k] = v
		}
	}
	return processed
}

// executeHTTPTool executes a tool with the given arguments
func (s *Server) executeHTTPTool(tool *config.ToolConfig, args map[string]any, request *http.Request, serverCfg map[string]string) (*mcp.CallToolResult, error) {
	// Prepare template context
	tmplCtx, err := prepareTemplateContext(args, request, serverCfg)
	if err != nil {
		return nil, err
	}

	// Prepare HTTP request
	req, err := prepareRequest(tool, tmplCtx)
	if err != nil {
		return nil, err
	}

	// Process arguments
	processArguments(req, tool, args)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	// Process response
	callToolResult, err := s.toolRespHandler.Handle(resp, tool, tmplCtx)
	if err != nil {
		return nil, err
	}
	return callToolResult, nil
}

func (s *Server) executeStdioTool(
	c *gin.Context,
	tool *config.MCPServerConfig,
	args map[string]any,
	request *http.Request,
	params mcp.CallToolParams,
) (*mcp.CallToolResult, error) {
	// Prepare template context
	tmplCtx, err := prepareTemplateContext(args, request, nil)
	if err != nil {
		return nil, err
	}

	renderedClientEnv := make(map[string]string)
	for k, v := range tool.Env {
		rendered, err := renderTemplate(v, tmplCtx)
		if err != nil {
			return nil, err
		}
		renderedClientEnv[k] = rendered
	}

	toolCallRequestParams := make(map[string]interface{})
	if err := json.Unmarshal(params.Arguments, &toolCallRequestParams); err != nil {
		return nil, err
	}

	// Create stdio transport
	stdioTransport := transport.NewStdio(
		tool.Command,
		utils.MapToEnvList(renderedClientEnv),
		tool.Args...,
	)

	// Start the transport
	if err := stdioTransport.Start(c.Request.Context()); err != nil {
		return nil, fmt.Errorf("failed to start stdio transport: %w", err)
	}

	// Create client
	mcpClient := client.NewClient(stdioTransport)
	defer mcpClient.Close()

	// Initialize client
	initRequest := mcpgo.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpgo.Implementation{
		Name:    "mcp-gateway",
		Version: version.Get(),
	}

	_, err = mcpClient.Initialize(c.Request.Context(), initRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stdio client: %w", err)
	}

	// Call tool
	callRequest := mcpgo.CallToolRequest{}
	callRequest.Params.Name = params.Name

	// Convert parameters to mcp-go format
	callRequest.Params.Arguments = toolCallRequestParams

	mcpgoResult, err := mcpClient.CallTool(c.Request.Context(), callRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	// Convert mcp-go result to local mcp format
	result := &mcp.CallToolResult{
		IsError: mcpgoResult.IsError,
	}

	// Process content items
	if len(mcpgoResult.Content) > 0 {
		var validContents []mcp.Content

		for _, content := range mcpgoResult.Content {
			// Skip null content
			if content == nil {
				continue
			}

			// Try to get content type
			contentType := ""
			switch c := content.(type) {
			case *mcpgo.TextContent:
				contentType = "text"
				validContents = append(validContents, &mcp.TextContent{
					Type: "text",
					Text: c.Text,
				})
			case *mcpgo.ImageContent:
				contentType = "image"
				validContents = append(validContents, &mcp.ImageContent{
					Type:     "image",
					Data:     c.Data,
					MimeType: c.MIMEType,
				})
			case *mcpgo.AudioContent:
				contentType = "audio"
				validContents = append(validContents, &mcp.AudioContent{
					Type:     "audio",
					Data:     c.Data,
					MimeType: c.MIMEType,
				})
			default:
				// Try to parse from raw content
				rawContent, err := json.Marshal(content)
				if err == nil {
					var contentMap map[string]interface{}
					if json.Unmarshal(rawContent, &contentMap) == nil {
						if typ, ok := contentMap["type"].(string); ok {
							contentType = typ

							switch contentType {
							case "text":
								if text, ok := contentMap["text"].(string); ok {
									validContents = append(validContents, &mcp.TextContent{
										Type: "text",
										Text: text,
									})
								}
							case "image":
								data, _ := contentMap["data"].(string)
								mimeType, _ := contentMap["mimeType"].(string)
								validContents = append(validContents, &mcp.ImageContent{
									Type:     "image",
									Data:     data,
									MimeType: mimeType,
								})
							case "audio":
								data, _ := contentMap["data"].(string)
								mimeType, _ := contentMap["mimeType"].(string)
								validContents = append(validContents, &mcp.AudioContent{
									Type:     "audio",
									Data:     data,
									MimeType: mimeType,
								})
							}
						}
					}
				}
			}
		}

		if len(validContents) > 0 {
			result.Content = validContents
		}
	}

	return result, nil
}

func (s *Server) fetchHTTPToolList(conn session.Connection) ([]mcp.ToolSchema, error) {
	// Get http tools for this prefix
	tools, ok := s.state.prefixToTools[conn.Meta().Prefix]
	if !ok {
		tools = []mcp.ToolSchema{} // Return empty list if prefix not found
	}

	return tools, nil
}

func (s *Server) fetchStdioToolList(ctx context.Context, conn session.Connection) ([]mcp.ToolSchema, error) {
	// Get stdio tools for this prefix
	stdioCfg, ok := s.state.prefixToMCPServerConfig[conn.Meta().Prefix]
	if !ok {
		return []mcp.ToolSchema{}, nil
	}

	// Create stdio transport with the command and arguments
	stdioTransport := transport.NewStdio(
		stdioCfg.Command,
		utils.MapToEnvList(stdioCfg.Env),
		stdioCfg.Args...,
	)

	// Start the transport
	if err := stdioTransport.Start(ctx); err != nil {
		return []mcp.ToolSchema{}, fmt.Errorf("failed to start stdio transport: %w", err)
	}

	// Create client with the transport
	c := client.NewClient(stdioTransport)
	defer c.Close()

	// Initialize the client
	initRequest := mcpgo.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpgo.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpgo.Implementation{
		Name:    "mcp-gateway",
		Version: version.Get(),
	}

	_, err := c.Initialize(ctx, initRequest)
	if err != nil {
		return []mcp.ToolSchema{}, fmt.Errorf("failed to initialize: %w", err)
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

func (s *Server) invokeHTTPTool(c *gin.Context, req mcp.JSONRPCRequest, conn session.Connection, params mcp.CallToolParams) *mcp.CallToolResult {
	// Find the tool in the precomputed map
	tool, exists := s.state.toolMap[params.Name]
	if !exists {
		errMsg := "Tool not found"
		s.sendProtocolError(c, req.Id, errMsg, http.StatusNotFound, mcp.ErrorCodeMethodNotFound)
		return nil
	}

	// Convert arguments to map[string]any
	var args map[string]any
	if err := json.Unmarshal(params.Arguments, &args); err != nil {
		errMsg := "Invalid tool arguments"
		s.sendProtocolError(c, req.Id, errMsg, http.StatusBadRequest, mcp.ErrorCodeInvalidParams)
		return nil
	}

	// Get server configuration
	serverCfg, ok := s.state.prefixToServerConfig[conn.Meta().Prefix]
	if !ok {
		errMsg := "Server configuration not found"
		s.sendProtocolError(c, req.Id, errMsg, http.StatusInternalServerError, mcp.ErrorCodeInternalError)
		return nil
	}

	// Execute the tool
	result, err := s.executeHTTPTool(tool, args, c.Request, serverCfg.Config)
	if err != nil {
		s.sendToolExecutionError(c, conn, req, err, true)
		return nil
	}

	return result
}

func (s *Server) invokeStdioTool(c *gin.Context, req mcp.JSONRPCRequest, conn session.Connection, params mcp.CallToolParams) *mcp.CallToolResult {
	// Get stdio tools for this prefix
	stdioCfg, ok := s.state.prefixToMCPServerConfig[conn.Meta().Prefix]
	if !ok {
		errMsg := "Server configuration not found"
		s.sendProtocolError(c, req.Id, errMsg, http.StatusNotFound, mcp.ErrorCodeMethodNotFound)
		return nil
	}

	// Convert arguments to map[string]any
	var args map[string]any
	if err := json.Unmarshal(params.Arguments, &args); err != nil {
		errMsg := "Invalid tool arguments"
		s.sendProtocolError(c, req.Id, errMsg, http.StatusBadRequest, mcp.ErrorCodeInvalidParams)
		return nil
	}

	result, err := s.executeStdioTool(c, &stdioCfg, args, c.Request, params)
	if err != nil {
		s.sendToolExecutionError(c, conn, req, err, true)
		return nil
	}

	return result
}
