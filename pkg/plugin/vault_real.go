package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VaultRealPlugin implements actual HashiCorp Vault integration
type VaultRealPlugin struct {
	client *http.Client
}

// NewVaultRealPlugin creates a new real Vault plugin with HTTP client
func NewVaultRealPlugin() *VaultRealPlugin {
	return &VaultRealPlugin{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *VaultRealPlugin) Metadata() Metadata {
	return Metadata{
		Name:        "vault",
		Version:     "2.0.0",
		Description: "HashiCorp Vault secrets management with real API integration",
		Author:      "Corynth Team",
		Tags:        []string{"secrets", "vault", "security", "encryption", "hashicorp"},
		License:     "Apache-2.0",
	}
}

func (p *VaultRealPlugin) Actions() []Action {
	return []Action{
		{
			Name:        "read",
			Description: "Read a secret from Vault KV store",
			Inputs: map[string]InputSpec{
				"address": {
					Type:        "string",
					Description: "Vault server address (e.g., https://vault.company.com)",
					Required:    false,
					Default:     "http://localhost:8200",
				},
				"token": {
					Type:        "string",
					Description: "Vault authentication token",
					Required:    true,
				},
				"path": {
					Type:        "string",
					Description: "Secret path (e.g., secret/data/myapp/config)",
					Required:    true,
				},
				"mount": {
					Type:        "string",
					Description: "KV mount path",
					Required:    false,
					Default:     "secret",
				},
			},
			Outputs: map[string]OutputSpec{
				"data": {
					Type:        "object",
					Description: "Secret data",
				},
				"version": {
					Type:        "number",
					Description: "Secret version",
				},
				"created_time": {
					Type:        "string",
					Description: "When the secret was created",
				},
			},
		},
		{
			Name:        "write",
			Description: "Write a secret to Vault KV store",
			Inputs: map[string]InputSpec{
				"address": {
					Type:        "string",
					Description: "Vault server address",
					Required:    false,
					Default:     "http://localhost:8200",
				},
				"token": {
					Type:        "string",
					Description: "Vault authentication token",
					Required:    true,
				},
				"path": {
					Type:        "string",
					Description: "Secret path (e.g., secret/data/myapp/config)",
					Required:    true,
				},
				"data": {
					Type:        "object",
					Description: "Secret data to write",
					Required:    true,
				},
				"mount": {
					Type:        "string",
					Description: "KV mount path",
					Required:    false,
					Default:     "secret",
				},
			},
			Outputs: map[string]OutputSpec{
				"success": {
					Type:        "boolean",
					Description: "Whether the write succeeded",
				},
				"version": {
					Type:        "number",
					Description: "New secret version",
				},
			},
		},
		{
			Name:        "delete",
			Description: "Delete a secret from Vault KV store",
			Inputs: map[string]InputSpec{
				"address": {
					Type:        "string",
					Description: "Vault server address",
					Required:    false,
					Default:     "http://localhost:8200",
				},
				"token": {
					Type:        "string",
					Description: "Vault authentication token",
					Required:    true,
				},
				"path": {
					Type:        "string",
					Description: "Secret path to delete",
					Required:    true,
				},
				"mount": {
					Type:        "string",
					Description: "KV mount path",
					Required:    false,
					Default:     "secret",
				},
			},
			Outputs: map[string]OutputSpec{
				"deleted": {
					Type:        "boolean",
					Description: "Whether the secret was deleted",
				},
			},
		},
		{
			Name:        "list",
			Description: "List secrets at a path in Vault KV store",
			Inputs: map[string]InputSpec{
				"address": {
					Type:        "string",
					Description: "Vault server address",
					Required:    false,
					Default:     "http://localhost:8200",
				},
				"token": {
					Type:        "string",
					Description: "Vault authentication token",
					Required:    true,
				},
				"path": {
					Type:        "string",
					Description: "Path to list (e.g., secret/metadata/myapp/)",
					Required:    true,
				},
				"mount": {
					Type:        "string",
					Description: "KV mount path",
					Required:    false,
					Default:     "secret",
				},
			},
			Outputs: map[string]OutputSpec{
				"keys": {
					Type:        "array",
					Description: "List of secret keys",
				},
				"count": {
					Type:        "number",
					Description: "Number of secrets found",
				},
			},
		},
		{
			Name:        "status",
			Description: "Check Vault server status and health",
			Inputs: map[string]InputSpec{
				"address": {
					Type:        "string",
					Description: "Vault server address",
					Required:    false,
					Default:     "http://localhost:8200",
				},
			},
			Outputs: map[string]OutputSpec{
				"sealed": {
					Type:        "boolean",
					Description: "Whether Vault is sealed",
				},
				"initialized": {
					Type:        "boolean",
					Description: "Whether Vault is initialized",
				},
				"version": {
					Type:        "string",
					Description: "Vault server version",
				},
			},
		},
	}
}

func (p *VaultRealPlugin) Validate(params map[string]interface{}) error {
	// Validate required fields based on action
	if token, exists := params["token"]; exists {
		if tokenStr, ok := token.(string); !ok || tokenStr == "" {
			return fmt.Errorf("token must be a non-empty string")
		}
	}
	
	if path, exists := params["path"]; exists {
		if pathStr, ok := path.(string); !ok || pathStr == "" {
			return fmt.Errorf("path must be a non-empty string")
		}
	}
	
	if address, exists := params["address"]; exists {
		if addrStr, ok := address.(string); !ok || addrStr == "" {
			return fmt.Errorf("address must be a non-empty string")
		}
		// Basic URL validation
		if addrStr, ok := address.(string); ok {
			if !strings.HasPrefix(addrStr, "http://") && !strings.HasPrefix(addrStr, "https://") {
				return fmt.Errorf("address must start with http:// or https://")
			}
		}
	}
	
	return nil
}

func (p *VaultRealPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "read":
		return p.executeRead(ctx, params)
	case "write":
		return p.executeWrite(ctx, params)
	case "delete":
		return p.executeDelete(ctx, params)
	case "list":
		return p.executeList(ctx, params)
	case "status":
		return p.executeStatus(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (p *VaultRealPlugin) executeRead(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	token, _ := params["token"].(string)
	path, _ := params["path"].(string)
	address := p.getAddress(params)
	mount := p.getMount(params)
	
	// Construct KV v2 API path
	apiPath := fmt.Sprintf("/v1/%s/data/%s", mount, strings.TrimPrefix(path, "/"))
	url := strings.TrimSuffix(address, "/") + apiPath
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Vault: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Vault returned error %d: %s", resp.StatusCode, string(body))
	}
	
	var vaultResp struct {
		Data struct {
			Data     map[string]interface{} `json:"data"`
			Metadata struct {
				Version     int    `json:"version"`
				CreatedTime string `json:"created_time"`
			} `json:"metadata"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("failed to decode Vault response: %w", err)
	}
	
	return map[string]interface{}{
		"data":         vaultResp.Data.Data,
		"version":      vaultResp.Data.Metadata.Version,
		"created_time": vaultResp.Data.Metadata.CreatedTime,
		"path":         path,
	}, nil
}

func (p *VaultRealPlugin) executeWrite(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	token, _ := params["token"].(string)
	path, _ := params["path"].(string)
	data, _ := params["data"].(map[string]interface{})
	address := p.getAddress(params)
	mount := p.getMount(params)
	
	// Construct KV v2 API path
	apiPath := fmt.Sprintf("/v1/%s/data/%s", mount, strings.TrimPrefix(path, "/"))
	url := strings.TrimSuffix(address, "/") + apiPath
	
	// Wrap data for KV v2
	payload := map[string]interface{}{
		"data": data,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Vault: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Vault returned error %d: %s", resp.StatusCode, string(body))
	}
	
	var vaultResp struct {
		Data struct {
			Version int `json:"version"`
		} `json:"data"`
	}
	
	// Try to decode response, but don't fail if empty
	json.NewDecoder(resp.Body).Decode(&vaultResp)
	
	return map[string]interface{}{
		"success": true,
		"version": vaultResp.Data.Version,
		"path":    path,
	}, nil
}

func (p *VaultRealPlugin) executeDelete(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	token, _ := params["token"].(string)
	path, _ := params["path"].(string)
	address := p.getAddress(params)
	mount := p.getMount(params)
	
	// Construct KV v2 API path for deletion
	apiPath := fmt.Sprintf("/v1/%s/data/%s", mount, strings.TrimPrefix(path, "/"))
	url := strings.TrimSuffix(address, "/") + apiPath
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", token)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Vault: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Vault returned error %d: %s", resp.StatusCode, string(body))
	}
	
	return map[string]interface{}{
		"deleted": true,
		"path":    path,
	}, nil
}

func (p *VaultRealPlugin) executeList(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	token, _ := params["token"].(string)
	path, _ := params["path"].(string)
	address := p.getAddress(params)
	mount := p.getMount(params)
	
	// Construct KV v2 API path for listing
	apiPath := fmt.Sprintf("/v1/%s/metadata/%s", mount, strings.TrimPrefix(path, "/"))
	url := strings.TrimSuffix(address, "/") + apiPath + "?list=true"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", token)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Vault: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return map[string]interface{}{
			"keys":  []string{},
			"count": 0,
			"path":  path,
		}, nil
	}
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Vault returned error %d: %s", resp.StatusCode, string(body))
	}
	
	var vaultResp struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("failed to decode Vault response: %w", err)
	}
	
	return map[string]interface{}{
		"keys":  vaultResp.Data.Keys,
		"count": len(vaultResp.Data.Keys),
		"path":  path,
	}, nil
}

func (p *VaultRealPlugin) executeStatus(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	address := p.getAddress(params)
	
	// Check health endpoint
	url := strings.TrimSuffix(address, "/") + "/v1/sys/health"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Vault: %w", err)
	}
	defer resp.Body.Close()
	
	var healthResp struct {
		Initialized bool   `json:"initialized"`
		Sealed      bool   `json:"sealed"`
		Version     string `json:"version"`
	}
	
	// Health endpoint returns different status codes based on state
	// 200 = unsealed and initialized
	// 429 = unsealed but uninitialized  
	// 472 = disaster recovery mode
	// 473 = performance standby
	// 501 = not initialized
	// 503 = sealed
	
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}
	
	return map[string]interface{}{
		"sealed":      healthResp.Sealed,
		"initialized": healthResp.Initialized,
		"version":     healthResp.Version,
		"status_code": resp.StatusCode,
	}, nil
}

// Helper functions
func (p *VaultRealPlugin) getAddress(params map[string]interface{}) string {
	if addr, ok := params["address"].(string); ok && addr != "" {
		return addr
	}
	return "http://localhost:8200"
}

func (p *VaultRealPlugin) getMount(params map[string]interface{}) string {
	if mount, ok := params["mount"].(string); ok && mount != "" {
		return mount
	}
	return "secret"
}