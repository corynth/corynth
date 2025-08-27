package plugin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPPlugin implements HTTP client functionality
type HTTPPlugin struct {
	metadata Metadata
}

// NewHTTPPlugin creates a new HTTP plugin instance
func NewHTTPPlugin() *HTTPPlugin {
	return &HTTPPlugin{
		metadata: Metadata{
			Name:        "http",
			Version:     "1.0.0",
			Description: "HTTP client for REST API calls and web requests",
			Author:      "Corynth Team",
			Tags:        []string{"network", "http", "api"},
		},
	}
}

// Metadata returns plugin metadata
func (p *HTTPPlugin) Metadata() Metadata {
	return p.metadata
}

// Actions returns available actions
func (p *HTTPPlugin) Actions() []Action {
	return []Action{
		{
			Name:        "get",
			Description: "Perform HTTP GET request",
			Inputs: map[string]InputSpec{
				"url": {
					Type:        "string",
					Description: "URL to request",
					Required:    true,
				},
				"timeout": {
					Type:        "number",
					Description: "Request timeout in seconds (default: 30)",
					Required:    false,
					Default:     30,
				},
				"headers": {
					Type:        "object",
					Description: "HTTP headers to send",
					Required:    false,
				},
			},
			Outputs: map[string]OutputSpec{
				"body": {
					Type:        "string",
					Description: "Response body",
				},
				"status_code": {
					Type:        "number",
					Description: "HTTP status code",
				},
				"headers": {
					Type:        "object",
					Description: "Response headers",
				},
				"success": {
					Type:        "boolean",
					Description: "Whether request was successful (2xx status)",
				},
			},
		},
		{
			Name:        "post",
			Description: "Perform HTTP POST request",
			Inputs: map[string]InputSpec{
				"url": {
					Type:        "string",
					Description: "URL to request",
					Required:    true,
				},
				"body": {
					Type:        "string",
					Description: "Request body",
					Required:    false,
				},
				"timeout": {
					Type:        "number",
					Description: "Request timeout in seconds (default: 30)",
					Required:    false,
					Default:     30,
				},
				"headers": {
					Type:        "object",
					Description: "HTTP headers to send",
					Required:    false,
				},
			},
			Outputs: map[string]OutputSpec{
				"body": {
					Type:        "string",
					Description: "Response body",
				},
				"status_code": {
					Type:        "number",
					Description: "HTTP status code",
				},
				"headers": {
					Type:        "object",
					Description: "Response headers",
				},
				"success": {
					Type:        "boolean",
					Description: "Whether request was successful (2xx status)",
				},
			},
		},
	}
}

// Execute runs the plugin action
func (p *HTTPPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "get":
		return p.performRequest(ctx, "GET", params)
	case "post":
		return p.performRequest(ctx, "POST", params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Validate checks if the plugin configuration is valid
func (p *HTTPPlugin) Validate(params map[string]interface{}) error {
	url, ok := params["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("url parameter is required and must be a non-empty string")
	}

	if timeout, exists := params["timeout"]; exists {
		if _, ok := timeout.(float64); !ok {
			if _, ok := timeout.(int); !ok {
				return fmt.Errorf("timeout parameter must be a number")
			}
		}
	}

	if headers, exists := params["headers"]; exists {
		if _, ok := headers.(map[string]interface{}); !ok {
			return fmt.Errorf("headers parameter must be an object")
		}
	}

	return nil
}

// performRequest performs the HTTP request
func (p *HTTPPlugin) performRequest(ctx context.Context, method string, params map[string]interface{}) (map[string]interface{}, error) {
	url := params["url"].(string)
	
	// Get timeout (default to 30 seconds)
	timeout := time.Duration(30) * time.Second
	if t, exists := params["timeout"]; exists {
		if tf, ok := t.(float64); ok {
			timeout = time.Duration(tf) * time.Second
		} else if ti, ok := t.(int); ok {
			timeout = time.Duration(ti) * time.Second
		}
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request body for POST
	var body io.Reader
	if method == "POST" {
		if bodyStr, exists := params["body"].(string); exists {
			body = strings.NewReader(bodyStr)
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	if headers, exists := params["headers"].(map[string]interface{}); exists {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	// Set default Content-Type for POST if not specified
	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"body":        "",
			"status_code": 0,
			"headers":     make(map[string]interface{}),
			"success":     false,
		}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert response headers to map
	respHeaders := make(map[string]interface{})
	for key, values := range resp.Header {
		if len(values) == 1 {
			respHeaders[key] = values[0]
		} else {
			respHeaders[key] = values
		}
	}

	// Determine success based on status code
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return map[string]interface{}{
		"body":        string(respBody),
		"status_code": resp.StatusCode,
		"headers":     respHeaders,
		"success":     success,
	}, nil
}