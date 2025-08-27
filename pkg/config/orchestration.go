package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a custom duration type that can parse from YAML strings
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler for Duration
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration format: %s", s)
	}
	
	*d = Duration(parsed)
	return nil
}

// MarshalYAML implements yaml.Marshaler for Duration
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// String returns the string representation
func (d Duration) String() string {
	return time.Duration(d).String()
}

// OrchestrationConfig defines configuration for workflow orchestration
type OrchestrationConfig struct {
	// Enable orchestration features
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// State management configuration
	State OrchestrationStateConfig `yaml:"state" json:"state"`
	
	// Execution limits and timeouts
	Execution ExecutionConfig `yaml:"execution" json:"execution"`
	
	// Retry and error handling
	Retry RetryConfig `yaml:"retry" json:"retry"`
	
	// Logging and observability
	Logging LoggingConfig `yaml:"logging" json:"logging"`
}

// OrchestrationStateConfig defines state management settings for orchestration
// This extends the base StateConfig with orchestration-specific options
type OrchestrationStateConfig struct {
	// Backend type: "local", "redis", "s3", "postgresql"
	Backend string `yaml:"backend" json:"backend"`
	
	// Directory for local state storage
	Directory string `yaml:"directory" json:"directory"`
	
	// Connection string for remote backends
	ConnectionString string `yaml:"connection_string" json:"connection_string"`
	
	// Backend-specific configuration
	BackendConfig map[string]string `yaml:"backend_config" json:"backend_config"`
	
	// Maximum number of execution states to keep
	MaxHistory int `yaml:"max_history" json:"max_history"`
	
	// How often to clean up old states
	CleanupInterval Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	
	// Enable state encryption
	Encryption bool `yaml:"encryption" json:"encryption"`
	
	// Encryption key (or path to key file)
	EncryptionKey string `yaml:"encryption_key" json:"encryption_key"`
}

// ExecutionConfig defines workflow execution limits
type ExecutionConfig struct {
	// Maximum number of concurrent workflow executions
	MaxConcurrent int `yaml:"max_concurrent" json:"max_concurrent"`
	
	// Maximum number of concurrent dependencies per workflow
	MaxConcurrentDependencies int `yaml:"max_concurrent_dependencies" json:"max_concurrent_dependencies"`
	
	// Maximum number of concurrent triggers per workflow
	MaxConcurrentTriggers int `yaml:"max_concurrent_triggers" json:"max_concurrent_triggers"`
	
	// Default timeout for workflow execution
	DefaultTimeout Duration `yaml:"default_timeout" json:"default_timeout"`
	
	// Default timeout for dependency execution
	DependencyTimeout Duration `yaml:"dependency_timeout" json:"dependency_timeout"`
	
	// Default timeout for trigger execution
	TriggerTimeout Duration `yaml:"trigger_timeout" json:"trigger_timeout"`
}

// RetryConfig defines default retry behavior
type RetryConfig struct {
	// Default maximum retry attempts
	DefaultMaxAttempts int `yaml:"default_max_attempts" json:"default_max_attempts"`
	
	// Default retry delay
	DefaultDelay Duration `yaml:"default_delay" json:"default_delay"`
	
	// Default backoff strategy: "linear", "exponential", "fixed"
	DefaultBackoff string `yaml:"default_backoff" json:"default_backoff"`
	
	// Maximum backoff delay
	MaxBackoffDelay Duration `yaml:"max_backoff_delay" json:"max_backoff_delay"`
}

// LoggingConfig defines logging behavior
type LoggingConfig struct {
	// Log level: "debug", "info", "warn", "error"
	Level string `yaml:"level" json:"level"`
	
	// Log format: "text", "json"
	Format string `yaml:"format" json:"format"`
	
	// Enable structured logging for orchestration events
	StructuredLogging bool `yaml:"structured_logging" json:"structured_logging"`
	
	// Enable metrics collection
	EnableMetrics bool `yaml:"enable_metrics" json:"enable_metrics"`
	
	// Metrics endpoint (for Prometheus)
	MetricsEndpoint string `yaml:"metrics_endpoint" json:"metrics_endpoint"`
}

// DefaultOrchestrationConfig returns the default configuration
func DefaultOrchestrationConfig() *OrchestrationConfig {
	homeDir, _ := os.UserHomeDir()
	defaultStateDir := filepath.Join(homeDir, ".corynth", "state")
	
	return &OrchestrationConfig{
		Enabled: true,
		
		State: OrchestrationStateConfig{
			Backend:         "local",
			Directory:       defaultStateDir,
			MaxHistory:      100,
			CleanupInterval: Duration(24 * time.Hour),
			Encryption:      false,
		},
		
		Execution: ExecutionConfig{
			MaxConcurrent:             5,
			MaxConcurrentDependencies: 3,
			MaxConcurrentTriggers:     2,
			DefaultTimeout:            Duration(30 * time.Minute),
			DependencyTimeout:         Duration(10 * time.Minute),
			TriggerTimeout:            Duration(5 * time.Minute),
		},
		
		Retry: RetryConfig{
			DefaultMaxAttempts: 1,
			DefaultDelay:       Duration(5 * time.Second),
			DefaultBackoff:     "linear",
			MaxBackoffDelay:    Duration(5 * time.Minute),
		},
		
		Logging: LoggingConfig{
			Level:             "info",
			Format:            "text",
			StructuredLogging: false,
			EnableMetrics:     false,
			MetricsEndpoint:   ":9090",
		},
	}
}

// LoadOrchestrationConfig loads configuration from a YAML file
func LoadOrchestrationConfig(configPath string) (*OrchestrationConfig, error) {
	// Start with defaults
	config := DefaultOrchestrationConfig()
	
	// If no config file specified, try standard locations
	if configPath == "" {
		configPath = findOrchestrationConfigFile()
	}
	
	// If config file doesn't exist, return defaults
	if configPath == "" {
		return config, nil
	}
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}
	
	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

// findOrchestrationConfigFile looks for orchestration configuration files in standard locations
func findOrchestrationConfigFile() string {
	homeDir, _ := os.UserHomeDir()
	
	candidates := []string{
		"corynth.yaml",
		"corynth.yml",
		".corynth.yaml",
		".corynth.yml",
		filepath.Join(homeDir, ".corynth", "config.yaml"),
		filepath.Join(homeDir, ".corynth", "config.yml"),
		"/etc/corynth/config.yaml",
	}
	
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	
	return ""
}

// Validate checks if the configuration is valid
func (c *OrchestrationConfig) Validate() error {
	// Validate state backend
	validBackends := []string{"local", "redis", "s3", "postgresql"}
	found := false
	for _, backend := range validBackends {
		if c.State.Backend == backend {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid state backend '%s', must be one of: %v", c.State.Backend, validBackends)
	}
	
	// Validate execution limits
	if c.Execution.MaxConcurrent <= 0 {
		return fmt.Errorf("max_concurrent must be positive, got %d", c.Execution.MaxConcurrent)
	}
	
	if c.Execution.MaxConcurrentDependencies <= 0 {
		return fmt.Errorf("max_concurrent_dependencies must be positive, got %d", c.Execution.MaxConcurrentDependencies)
	}
	
	if c.Execution.MaxConcurrentTriggers <= 0 {
		return fmt.Errorf("max_concurrent_triggers must be positive, got %d", c.Execution.MaxConcurrentTriggers)
	}
	
	// Validate timeouts
	if time.Duration(c.Execution.DefaultTimeout) <= 0 {
		return fmt.Errorf("default_timeout must be positive, got %v", c.Execution.DefaultTimeout)
	}
	
	// Validate retry configuration
	if c.Retry.DefaultMaxAttempts < 1 {
		return fmt.Errorf("default_max_attempts must be at least 1, got %d", c.Retry.DefaultMaxAttempts)
	}
	
	validBackoffs := []string{"linear", "exponential", "fixed"}
	found = false
	for _, backoff := range validBackoffs {
		if c.Retry.DefaultBackoff == backoff {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid default_backoff '%s', must be one of: %v", c.Retry.DefaultBackoff, validBackoffs)
	}
	
	// Validate logging configuration
	validLevels := []string{"debug", "info", "warn", "error"}
	found = false
	for _, level := range validLevels {
		if c.Logging.Level == level {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid log level '%s', must be one of: %v", c.Logging.Level, validLevels)
	}
	
	validFormats := []string{"text", "json"}
	found = false
	for _, format := range validFormats {
		if c.Logging.Format == format {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid log format '%s', must be one of: %v", c.Logging.Format, validFormats)
	}
	
	return nil
}

// Save saves the configuration to a YAML file
func (c *OrchestrationConfig) Save(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}
	
	// Write config file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}
	
	return nil
}

// GetStateDirectory returns the resolved state directory path
func (c *OrchestrationConfig) GetStateDirectory() string {
	if c.State.Directory == "" {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".corynth", "state")
	}
	
	// Expand ~ to home directory
	if c.State.Directory[:2] == "~/" {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, c.State.Directory[2:])
	}
	
	return c.State.Directory
}

// GetBackendConfig returns the backend configuration map for state management
func (c *OrchestrationConfig) GetBackendConfig() map[string]string {
	config := make(map[string]string)
	
	// Copy backend_config
	for k, v := range c.State.BackendConfig {
		config[k] = v
	}
	
	// Add standard fields
	if c.State.Directory != "" {
		config["path"] = c.GetStateDirectory()
	}
	
	if c.State.ConnectionString != "" {
		config["connection_string"] = c.State.ConnectionString
	}
	
	if c.State.Encryption {
		config["encryption"] = "true"
	} else {
		config["encryption"] = "false"
	}
	
	return config
}