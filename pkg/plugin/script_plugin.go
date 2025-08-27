package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// ScriptPlugin wraps a script-based plugin
type ScriptPlugin struct {
	metadata   Metadata
	scriptPath string
	actions    []Action
}

// Metadata returns plugin metadata
func (p *ScriptPlugin) Metadata() Metadata {
	return p.metadata
}

// Actions returns available actions
func (p *ScriptPlugin) Actions() []Action {
	return p.actions
}

// Execute runs the plugin action
func (p *ScriptPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	// Convert params to JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	// Set timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	
	// Execute script with action and parameters
	cmd := exec.CommandContext(timeoutCtx, p.scriptPath, action)
	cmd.Stdin = bytes.NewReader(paramsJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("script execution failed: %w\nstderr: %s", err, stderr.String())
	}

	// Parse output as JSON
	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		// If not JSON, return as string output
		return map[string]interface{}{
			"output": stdout.String(),
		}, nil
	}

	return result, nil
}

// Validate checks if the plugin configuration is valid
func (p *ScriptPlugin) Validate(params map[string]interface{}) error {
	// Basic validation - script plugins handle their own validation
	return nil
}