package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	pluginv2 "github.com/corynth/corynth/pkg/plugin/v2"
)

// HTTPPlugin implements HTTP operations via gRPC
type HTTPPlugin struct {
	*pluginv2.BasePlugin
}

// NewHTTPPlugin creates a new HTTP plugin
func NewHTTPPlugin() *HTTPPlugin {
	base := pluginv2.NewBuilder("http", "2.0.0").
		Description("HTTP client for REST API calls and web requests").
		Author("Corynth Team").
		Tags("http", "web", "api", "rest").
		Action("get", "Perform HTTP GET request").
		Input("url", "string", "URL to request", true).
		InputWithDefault("timeout", "number", "Request timeout in seconds", 30.0).
		InputWithDefault("headers", "object", "HTTP headers", map[string]string{}).
		Output("status_code", "number", "HTTP status code").
		Output("body", "string", "Response body").
		Output("headers", "object", "Response headers").
		Add().
		Action("post", "Perform HTTP POST request").
		Input("url", "string", "URL to request", true).
		Input("body", "string", "Request body", true).
		InputWithDefault("timeout", "number", "Request timeout in seconds", 30.0).
		InputWithDefault("headers", "object", "HTTP headers", map[string]string{"Content-Type": "application/json"}).
		Output("status_code", "number", "HTTP status code").
		Output("body", "string", "Response body").
		Output("headers", "object", "Response headers").
		Add().
		Build()

	return &HTTPPlugin{
		BasePlugin: base,
	}
}

// Execute implements the plugin execution
func (p *HTTPPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "get":
		return p.handleGet(ctx, params)
	case "post":
		return p.handlePost(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Validate validates parameters
func (p *HTTPPlugin) Validate(params map[string]interface{}) error {
	// Use base validation first
	if err := p.BasePlugin.Validate(params); err != nil {
		return err
	}
	
	// URL validation
	if url, exists := params["url"]; exists {
		if urlStr, ok := url.(string); ok {
			if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
				return fmt.Errorf("URL must start with http:// or https://")
			}
		}
	}
	
	return nil
}

// handleGet performs HTTP GET request
func (p *HTTPPlugin) handleGet(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	url := params["url"].(string)
	timeout := getTimeout(params)
	headers := getHeaders(params)
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	
	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Convert response headers
	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		responseHeaders[k] = strings.Join(v, ", ")
	}
	
	return map[string]interface{}{
		"status_code": float64(resp.StatusCode),
		"body":        string(body),
		"headers":     responseHeaders,
	}, nil
}

// handlePost performs HTTP POST request
func (p *HTTPPlugin) handlePost(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	url := params["url"].(string)
	body := params["body"].(string)
	timeout := getTimeout(params)
	headers := getHeaders(params)
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	
	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Convert response headers
	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		responseHeaders[k] = strings.Join(v, ", ")
	}
	
	return map[string]interface{}{
		"status_code": float64(resp.StatusCode),
		"body":        string(responseBody),
		"headers":     responseHeaders,
	}, nil
}

// Helper functions
func getTimeout(params map[string]interface{}) int {
	if timeout, exists := params["timeout"]; exists {
		if timeoutFloat, ok := timeout.(float64); ok {
			return int(timeoutFloat)
		}
		if timeoutStr, ok := timeout.(string); ok {
			if val, err := strconv.Atoi(timeoutStr); err == nil {
				return val
			}
		}
	}
	return 30 // Default timeout
}

func getHeaders(params map[string]interface{}) map[string]string {
	headers := make(map[string]string)
	
	if headersParam, exists := params["headers"]; exists {
		if headersMap, ok := headersParam.(map[string]interface{}); ok {
			for k, v := range headersMap {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	return headers
}

func main() {
	httpPlugin := NewHTTPPlugin()
	sdk := pluginv2.NewSDK(httpPlugin)
	sdk.Serve()
}