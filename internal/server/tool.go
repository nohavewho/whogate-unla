package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	texttemplate "text/template"

	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"github.com/mcp-ecosystem/mcp-gateway/internal/template"
)

// executeTool executes a tool with the given arguments
func (s *Server) executeTool(tool *config.ToolConfig, args map[string]any, request *http.Request) (string, error) {
	// Create HTTP client
	client := &http.Client{}

	// Create template context
	tmplCtx := template.NewContext()
	tmplCtx.Args = args

	// Set request headers in template context
	for k, v := range request.Header {
		if len(v) > 0 {
			tmplCtx.Request.Headers[k] = v[0]
		}
	}

	// Parse endpoint template
	endpointTmpl, err := texttemplate.New("endpoint").Parse(tool.Endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse endpoint template: %w", err)
	}

	// Execute endpoint template
	var endpointBuf bytes.Buffer
	if err := endpointTmpl.Execute(&endpointBuf, tmplCtx); err != nil {
		return "", fmt.Errorf("failed to execute endpoint template: %w", err)
	}
	endpoint := endpointBuf.String()

	// Parse request body template if provided
	var reqBody io.Reader
	if tool.RequestBody != "" {
		bodyTmpl, err := texttemplate.New("body").Parse(tool.RequestBody)
		if err != nil {
			return "", fmt.Errorf("failed to parse request body template: %w", err)
		}

		var bodyBuf bytes.Buffer
		if err := bodyTmpl.Execute(&bodyBuf, tmplCtx); err != nil {
			return "", fmt.Errorf("failed to execute request body template: %w", err)
		}
		reqBody = &bodyBuf
	}

	// Create request
	req, err := http.NewRequest(tool.Method, endpoint, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for k, v := range tool.Headers {
		headerTmpl, err := texttemplate.New("header").Parse(v)
		if err != nil {
			return "", fmt.Errorf("failed to parse header template: %w", err)
		}

		var headerBuf bytes.Buffer
		if err := headerTmpl.Execute(&headerBuf, tmplCtx); err != nil {
			return "", fmt.Errorf("failed to execute header template: %w", err)
		}
		req.Header.Set(k, headerBuf.String())
	}

	// Add arguments to headers, query params, or path params
	for _, arg := range tool.Args {
		value := fmt.Sprint(args[arg.Name])
		switch strings.ToLower(arg.Position) {
		case "header":
			req.Header.Set(arg.Name, value)
		case "query":
			q := req.URL.Query()
			q.Add(arg.Name, value)
			req.URL.RawQuery = q.Encode()
		}
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response body template if provided
	if tool.ResponseBody != "" {
		respTmpl, err := texttemplate.New("response").Parse(tool.ResponseBody)
		if err != nil {
			return "", fmt.Errorf("failed to parse response body template: %w", err)
		}

		// Create data map for response template
		var respData map[string]any
		if err := json.Unmarshal(respBody, &respData); err != nil {
			return "", fmt.Errorf("failed to parse response JSON: %w", err)
		}

		// Wrap response data in response.data
		tmplCtx.Response.Data = respData
		var respBuf bytes.Buffer
		if err := respTmpl.Execute(&respBuf, tmplCtx); err != nil {
			return "", fmt.Errorf("failed to execute response body template: %w", err)
		}
		return respBuf.String(), nil
	}

	return string(respBody), nil
}
