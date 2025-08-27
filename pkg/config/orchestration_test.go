package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultOrchestrationConfig(t *testing.T) {
	config := DefaultOrchestrationConfig()
	
	// Check basic defaults
	if !config.Enabled {
		t.Error("Expected orchestration to be enabled by default")
	}
	
	if config.State.Backend != "local" {
		t.Errorf("Expected default backend to be 'local', got %s", config.State.Backend)
	}
	
	if config.Execution.MaxConcurrent <= 0 {
		t.Error("Expected positive max concurrent workflows")
	}
	
	if config.Retry.DefaultMaxAttempts < 1 {
		t.Error("Expected at least 1 retry attempt by default")
	}
	
	// Validate the default config
	if err := config.Validate(); err != nil {
		t.Errorf("Default config should be valid, got error: %v", err)
	}
}

func TestLoadOrchestrationConfigFromFile(t *testing.T) {
	// Create temporary config file
	tempDir, err := os.MkdirTemp("", "corynth-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
enabled: true
state:
  backend: "redis"
  connection_string: "redis://localhost:6379"
  max_history: 200
  cleanup_interval: "48h"
execution:
  max_concurrent: 10
  default_timeout: "1h"
retry:
  default_max_attempts: 3
  default_delay: "10s"
  default_backoff: "exponential"
logging:
  level: "debug"
  format: "json"
  structured_logging: true
`
	
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Load config
	config, err := LoadOrchestrationConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify loaded values
	if config.State.Backend != "redis" {
		t.Errorf("Expected backend 'redis', got %s", config.State.Backend)
	}
	
	if config.State.ConnectionString != "redis://localhost:6379" {
		t.Errorf("Expected connection string 'redis://localhost:6379', got %s", config.State.ConnectionString)
	}
	
	if config.State.MaxHistory != 200 {
		t.Errorf("Expected max history 200, got %d", config.State.MaxHistory)
	}
	
	if config.Execution.MaxConcurrent != 10 {
		t.Errorf("Expected max concurrent 10, got %d", config.Execution.MaxConcurrent)
	}
	
	if config.Retry.DefaultMaxAttempts != 3 {
		t.Errorf("Expected default max attempts 3, got %d", config.Retry.DefaultMaxAttempts)
	}
	
	if config.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", config.Logging.Level)
	}
	
	if !config.Logging.StructuredLogging {
		t.Error("Expected structured logging to be enabled")
	}
}

func TestLoadOrchestrationConfigNonExistent(t *testing.T) {
	// Loading non-existent config should return defaults
	config, err := LoadOrchestrationConfig("/non/existent/path")
	if err != nil {
		t.Fatalf("Loading non-existent config should return defaults, got error: %v", err)
	}
	
	// Should be the same as default config
	defaultConfig := DefaultOrchestrationConfig()
	if config.State.Backend != defaultConfig.State.Backend {
		t.Errorf("Expected default backend, got %s", config.State.Backend)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		modifyConfig func(*OrchestrationConfig)
		expectError string
	}{
		{
			name: "invalid state backend",
			modifyConfig: func(c *OrchestrationConfig) {
				c.State.Backend = "invalid"
			},
			expectError: "invalid state backend",
		},
		{
			name: "zero max concurrent",
			modifyConfig: func(c *OrchestrationConfig) {
				c.Execution.MaxConcurrent = 0
			},
			expectError: "max_concurrent must be positive",
		},
		{
			name: "zero max concurrent dependencies",
			modifyConfig: func(c *OrchestrationConfig) {
				c.Execution.MaxConcurrentDependencies = 0
			},
			expectError: "max_concurrent_dependencies must be positive",
		},
		{
			name: "zero timeout",
			modifyConfig: func(c *OrchestrationConfig) {
				c.Execution.DefaultTimeout = 0
			},
			expectError: "default_timeout must be positive",
		},
		{
			name: "zero retry attempts",
			modifyConfig: func(c *OrchestrationConfig) {
				c.Retry.DefaultMaxAttempts = 0
			},
			expectError: "default_max_attempts must be at least 1",
		},
		{
			name: "invalid backoff strategy",
			modifyConfig: func(c *OrchestrationConfig) {
				c.Retry.DefaultBackoff = "invalid"
			},
			expectError: "invalid default_backoff",
		},
		{
			name: "invalid log level",
			modifyConfig: func(c *OrchestrationConfig) {
				c.Logging.Level = "invalid"
			},
			expectError: "invalid log level",
		},
		{
			name: "invalid log format",
			modifyConfig: func(c *OrchestrationConfig) {
				c.Logging.Format = "invalid"
			},
			expectError: "invalid log format",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultOrchestrationConfig()
			tt.modifyConfig(config)
			
			err := config.Validate()
			if err == nil {
				t.Errorf("Expected validation error containing '%s', got nil", tt.expectError)
			} else if !contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectError, err.Error())
			}
		})
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "corynth-config-save-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a custom config
	originalConfig := DefaultOrchestrationConfig()
	originalConfig.State.Backend = "postgresql"
	originalConfig.State.MaxHistory = 500
	originalConfig.Execution.MaxConcurrent = 20
	originalConfig.Retry.DefaultMaxAttempts = 5
	originalConfig.Logging.Level = "debug"
	
	// Save config
	configFile := filepath.Join(tempDir, "test-config.yaml")
	err = originalConfig.Save(configFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Load config back
	loadedConfig, err := LoadOrchestrationConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	
	// Verify values match
	if loadedConfig.State.Backend != originalConfig.State.Backend {
		t.Errorf("Backend mismatch: expected %s, got %s", 
			originalConfig.State.Backend, loadedConfig.State.Backend)
	}
	
	if loadedConfig.State.MaxHistory != originalConfig.State.MaxHistory {
		t.Errorf("MaxHistory mismatch: expected %d, got %d", 
			originalConfig.State.MaxHistory, loadedConfig.State.MaxHistory)
	}
	
	if loadedConfig.Execution.MaxConcurrent != originalConfig.Execution.MaxConcurrent {
		t.Errorf("MaxConcurrent mismatch: expected %d, got %d", 
			originalConfig.Execution.MaxConcurrent, loadedConfig.Execution.MaxConcurrent)
	}
}

func TestGetStateDirectory(t *testing.T) {
	config := DefaultOrchestrationConfig()
	
	// Test default directory
	stateDir := config.GetStateDirectory()
	if stateDir == "" {
		t.Error("State directory should not be empty")
	}
	
	// Test custom directory
	config.State.Directory = "/custom/state/path"
	stateDir = config.GetStateDirectory()
	if stateDir != "/custom/state/path" {
		t.Errorf("Expected custom path, got %s", stateDir)
	}
	
	// Test tilde expansion
	config.State.Directory = "~/custom/state"
	stateDir = config.GetStateDirectory()
	if stateDir[:2] == "~/" {
		t.Error("Tilde should have been expanded")
	}
}

func TestParseTimeoutsFromString(t *testing.T) {
	// Test that YAML parsing handles time.Duration correctly
	tempDir, err := os.MkdirTemp("", "corynth-timeout-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
execution:
  default_timeout: "2h30m"
  dependency_timeout: "45m"
state:
  cleanup_interval: "168h"  # 7 days
retry:
  default_delay: "30s"
  max_backoff_delay: "10m"
`
	
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	config, err := LoadOrchestrationConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify durations are parsed correctly
	if time.Duration(config.Execution.DefaultTimeout) != 2*time.Hour+30*time.Minute {
		t.Errorf("Expected 2h30m, got %v", config.Execution.DefaultTimeout)
	}
	
	if time.Duration(config.Execution.DependencyTimeout) != 45*time.Minute {
		t.Errorf("Expected 45m, got %v", config.Execution.DependencyTimeout)
	}
	
	if time.Duration(config.State.CleanupInterval) != 7*24*time.Hour {
		t.Errorf("Expected 7 days, got %v", config.State.CleanupInterval)
	}
	
	if time.Duration(config.Retry.DefaultDelay) != 30*time.Second {
		t.Errorf("Expected 30s, got %v", config.Retry.DefaultDelay)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}