package plugin

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FilePlugin implements file system operations
type FilePlugin struct {
	metadata Metadata
}

// NewFilePlugin creates a new file plugin instance
func NewFilePlugin() *FilePlugin {
	return &FilePlugin{
		metadata: Metadata{
			Name:        "file",
			Version:     "1.0.0",
			Description: "File system operations (read, write, copy, move)",
			Author:      "Corynth Team",
			Tags:        []string{"filesystem", "io"},
		},
	}
}

// Metadata returns plugin metadata
func (p *FilePlugin) Metadata() Metadata {
	return p.metadata
}

// Actions returns available actions
func (p *FilePlugin) Actions() []Action {
	return []Action{
		{
			Name:        "read",
			Description: "Read file contents",
			Inputs: map[string]InputSpec{
				"path": {
					Type:        "string",
					Description: "Path to file to read",
					Required:    true,
				},
			},
			Outputs: map[string]OutputSpec{
				"content": {
					Type:        "string",
					Description: "File contents",
				},
				"size": {
					Type:        "number",
					Description: "File size in bytes",
				},
				"exists": {
					Type:        "boolean",
					Description: "Whether file exists",
				},
			},
		},
		{
			Name:        "write",
			Description: "Write content to file",
			Inputs: map[string]InputSpec{
				"path": {
					Type:        "string",
					Description: "Path to file to write",
					Required:    true,
				},
				"content": {
					Type:        "string",
					Description: "Content to write",
					Required:    true,
				},
				"create_dirs": {
					Type:        "boolean",
					Description: "Create parent directories if they don't exist (default: false)",
					Required:    false,
					Default:     false,
				},
			},
			Outputs: map[string]OutputSpec{
				"bytes_written": {
					Type:        "number",
					Description: "Number of bytes written",
				},
				"success": {
					Type:        "boolean",
					Description: "Whether write was successful",
				},
			},
		},
		{
			Name:        "copy",
			Description: "Copy file from source to destination",
			Inputs: map[string]InputSpec{
				"source": {
					Type:        "string",
					Description: "Source file path",
					Required:    true,
				},
				"destination": {
					Type:        "string",
					Description: "Destination file path",
					Required:    true,
				},
				"create_dirs": {
					Type:        "boolean",
					Description: "Create parent directories if they don't exist (default: false)",
					Required:    false,
					Default:     false,
				},
			},
			Outputs: map[string]OutputSpec{
				"bytes_copied": {
					Type:        "number",
					Description: "Number of bytes copied",
				},
				"success": {
					Type:        "boolean",
					Description: "Whether copy was successful",
				},
			},
		},
		{
			Name:        "delete",
			Description: "Delete file or directory",
			Inputs: map[string]InputSpec{
				"path": {
					Type:        "string",
					Description: "Path to file or directory to delete",
					Required:    true,
				},
			},
			Outputs: map[string]OutputSpec{
				"success": {
					Type:        "boolean",
					Description: "Whether deletion was successful",
				},
			},
		},
	}
}

// Execute runs the plugin action
func (p *FilePlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "read":
		return p.readFile(params)
	case "write":
		return p.writeFile(params)
	case "copy":
		return p.copyFile(params)
	case "delete":
		return p.deleteFile(params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Validate checks if the plugin configuration is valid
func (p *FilePlugin) Validate(params map[string]interface{}) error {
	// All actions require at least a path
	path, ok := params["path"].(string)
	if !ok || path == "" {
		// Check if it's a copy action with source/destination
		if source, hasSource := params["source"].(string); hasSource {
			if dest, hasDest := params["destination"].(string); hasDest {
				if source == "" || dest == "" {
					return fmt.Errorf("source and destination parameters must be non-empty strings")
				}
				return nil // Valid copy operation
			}
		}
		return fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	return nil
}

// readFile reads a file and returns its contents
func (p *FilePlugin) readFile(params map[string]interface{}) (map[string]interface{}, error) {
	path := params["path"].(string)
	
	// Check if file exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return map[string]interface{}{
			"content": "",
			"size":    0,
			"exists":  false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read file contents
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]interface{}{
		"content": string(content),
		"size":    info.Size(),
		"exists":  true,
	}, nil
}

// writeFile writes content to a file
func (p *FilePlugin) writeFile(params map[string]interface{}) (map[string]interface{}, error) {
	path := params["path"].(string)
	content := params["content"].(string)
	
	// Check if we should create parent directories
	if createDirs, exists := params["create_dirs"].(bool); exists && createDirs {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return map[string]interface{}{
				"bytes_written": 0,
				"success":       false,
			}, fmt.Errorf("failed to create parent directories: %w", err)
		}
	}

	// Write file
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return map[string]interface{}{
			"bytes_written": 0,
			"success":       false,
		}, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"bytes_written": len(content),
		"success":       true,
	}, nil
}

// copyFile copies a file from source to destination
func (p *FilePlugin) copyFile(params map[string]interface{}) (map[string]interface{}, error) {
	source := params["source"].(string)
	destination := params["destination"].(string)
	
	// Check if we should create parent directories
	if createDirs, exists := params["create_dirs"].(bool); exists && createDirs {
		dir := filepath.Dir(destination)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return map[string]interface{}{
				"bytes_copied": 0,
				"success":      false,
			}, fmt.Errorf("failed to create parent directories: %w", err)
		}
	}

	// Open source file
	srcFile, err := os.Open(source)
	if err != nil {
		return map[string]interface{}{
			"bytes_copied": 0,
			"success":      false,
		}, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destination)
	if err != nil {
		return map[string]interface{}{
			"bytes_copied": 0,
			"success":      false,
		}, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy contents
	bytesCopied, err := io.Copy(destFile, srcFile)
	if err != nil {
		return map[string]interface{}{
			"bytes_copied": bytesCopied,
			"success":      false,
		}, fmt.Errorf("failed to copy file: %w", err)
	}

	return map[string]interface{}{
		"bytes_copied": bytesCopied,
		"success":      true,
	}, nil
}

// deleteFile deletes a file or directory
func (p *FilePlugin) deleteFile(params map[string]interface{}) (map[string]interface{}, error) {
	path := params["path"].(string)
	
	// Delete file or directory
	err := os.RemoveAll(path)
	if err != nil {
		return map[string]interface{}{
			"success": false,
		}, fmt.Errorf("failed to delete file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}