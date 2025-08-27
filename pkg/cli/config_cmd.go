package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/corynth/corynth/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewOrchestrationConfigCommand creates the orchestration config management command group
func NewOrchestrationConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Corynth configuration",
		Long: `Config commands help manage Corynth configuration files and settings.
These commands provide visibility into current configuration and help
create and modify configuration files.`,
	}

	cmd.AddCommand(NewConfigShowCommand())
	cmd.AddCommand(NewConfigInitCommand())
	cmd.AddCommand(NewConfigValidateCommand())
	cmd.AddCommand(NewConfigSetCommand())

	return cmd
}

// NewConfigShowCommand shows current configuration
func NewConfigShowCommand() *cobra.Command {
	var configFile string
	var format string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long: `Show displays the current configuration, including default values
and any overrides from configuration files.`,
		Example: `  # Show current config
  corynth config show
  
  # Show specific config file  
  corynth config show --config ./corynth.yaml
  
  # JSON output
  corynth config show --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(configFile, format)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")
	cmd.Flags().StringVar(&format, "format", "yaml", "Output format (yaml, json)")

	return cmd
}

// NewConfigInitCommand creates a new configuration file
func NewConfigInitCommand() *cobra.Command {
	var configFile string
	var force bool

	cmd := &cobra.Command{
		Use:   "init [config-file]",
		Short: "Initialize a new configuration file",
		Long: `Init creates a new configuration file with default values.
This provides a starting point for customizing Corynth behavior.`,
		Example: `  # Create default config file
  corynth config init
  
  # Create config in specific location
  corynth config init ~/my-corynth.yaml
  
  # Overwrite existing file
  corynth config init --force`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetFile := configFile
			if len(args) > 0 {
				targetFile = args[0]
			}
			return runConfigInit(targetFile, force)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing file")

	return cmd
}

// NewConfigValidateCommand validates configuration
func NewConfigValidateCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "validate [config-file]",
		Short: "Validate configuration file",
		Long: `Validate checks a configuration file for syntax errors and
validates all settings against expected values and constraints.`,
		Example: `  # Validate default config
  corynth config validate
  
  # Validate specific file
  corynth config validate ./my-config.yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetFile := configFile
			if len(args) > 0 {
				targetFile = args[0]
			}
			return runConfigValidate(targetFile)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")

	return cmd
}

// NewConfigSetCommand sets configuration values
func NewConfigSetCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set modifies a configuration value in the configuration file.
This provides a convenient way to update settings without manual editing.`,
		Example: `  # Enable orchestration
  corynth config set orchestration.enabled true
  
  # Set state directory
  corynth config set orchestration.state.directory /opt/corynth/state
  
  # Set concurrency limit
  corynth config set orchestration.execution.max_concurrent 10`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigSet(configFile, args[0], args[1])
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")

	return cmd
}

// Implementation functions

func runConfigShow(configFile, format string) error {
	// Load configuration
	orchConfig, err := config.LoadOrchestrationConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Show configuration based on format
	switch format {
	case "yaml", "yml":
		data, err := yaml.Marshal(orchConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		fmt.Printf("# Corynth Orchestration Configuration\n\n")
		fmt.Print(string(data))

	case "json":
		// TODO: Implement JSON output
		fmt.Println("JSON output not yet implemented")

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Show config source info
	fmt.Printf("\n# Configuration source: ")
	if configFile != "" {
		fmt.Printf("%s\n", configFile)
	} else {
		fmt.Printf("defaults + discovered config files\n")
	}

	return nil
}

func runConfigInit(configFile string, force bool) error {
	// Determine target file
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(homeDir, ".corynth", "config.yaml")
	}

	// Check if file already exists
	if _, err := os.Stat(configFile); err == nil && !force {
		return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configFile)
	}

	// Create default configuration
	defaultConfig := config.DefaultOrchestrationConfig()

	// Add comments to the YAML
	configWithComments := fmt.Sprintf(`# Corynth Orchestration Configuration
# Generated by: corynth config init

# Enable or disable orchestration features
enabled: %t

# State management configuration
state:
  # Backend type: local, redis, s3, postgresql
  backend: "%s"
  
  # Directory for local state storage  
  directory: "%s"
  
  # Maximum number of execution states to keep
  max_history: %d
  
  # How often to clean up old states
  cleanup_interval: "%s"
  
  # Enable state encryption
  encryption: %t

# Execution limits and timeouts
execution:
  # Maximum number of concurrent workflow executions
  max_concurrent: %d
  
  # Maximum concurrent dependencies per workflow
  max_concurrent_dependencies: %d
  
  # Maximum concurrent triggers per workflow
  max_concurrent_triggers: %d
  
  # Default timeout for workflow execution
  default_timeout: "%s"
  
  # Default timeout for dependency execution
  dependency_timeout: "%s"
  
  # Default timeout for trigger execution
  trigger_timeout: "%s"

# Retry and error handling defaults
retry:
  # Default maximum retry attempts
  default_max_attempts: %d
  
  # Default retry delay
  default_delay: "%s"
  
  # Default backoff strategy: linear, exponential, fixed
  default_backoff: "%s"
  
  # Maximum backoff delay
  max_backoff_delay: "%s"

# Logging and observability
logging:
  # Log level: debug, info, warn, error
  level: "%s"
  
  # Log format: text, json
  format: "%s"
  
  # Enable structured logging for orchestration events
  structured_logging: %t
  
  # Enable metrics collection
  enable_metrics: %t
  
  # Metrics endpoint (for Prometheus)
  metrics_endpoint: "%s"
`,
		defaultConfig.Enabled,
		defaultConfig.State.Backend,
		defaultConfig.State.Directory,
		defaultConfig.State.MaxHistory,
		defaultConfig.State.CleanupInterval,
		defaultConfig.State.Encryption,
		defaultConfig.Execution.MaxConcurrent,
		defaultConfig.Execution.MaxConcurrentDependencies,
		defaultConfig.Execution.MaxConcurrentTriggers,
		defaultConfig.Execution.DefaultTimeout,
		defaultConfig.Execution.DependencyTimeout,
		defaultConfig.Execution.TriggerTimeout,
		defaultConfig.Retry.DefaultMaxAttempts,
		defaultConfig.Retry.DefaultDelay,
		defaultConfig.Retry.DefaultBackoff,
		defaultConfig.Retry.MaxBackoffDelay,
		defaultConfig.Logging.Level,
		defaultConfig.Logging.Format,
		defaultConfig.Logging.StructuredLogging,
		defaultConfig.Logging.EnableMetrics,
		defaultConfig.Logging.MetricsEndpoint,
	)

	// Create directory if needed
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write configuration file
	if err := os.WriteFile(configFile, []byte(configWithComments), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Configuration file created: %s\n", configFile)
	fmt.Printf("💡 Edit this file to customize Corynth behavior\n")
	fmt.Printf("💡 Use 'corynth config validate' to check your changes\n")

	return nil
}

func runConfigValidate(configFile string) error {
	fmt.Printf("🔍 Validating configuration")
	if configFile != "" {
		fmt.Printf(": %s", configFile)
	}
	fmt.Println()

	// Load and validate configuration
	orchConfig, err := config.LoadOrchestrationConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	// The LoadOrchestrationConfig function already validates the config
	// If we get here, validation passed
	fmt.Printf("✅ Configuration is valid\n")

	// Show some basic info
	fmt.Printf("\n📋 Configuration Summary:\n")
	fmt.Printf("  Orchestration: %s\n", enabledString(orchConfig.Enabled))
	fmt.Printf("  State Backend: %s\n", orchConfig.State.Backend)
	fmt.Printf("  State Directory: %s\n", orchConfig.GetStateDirectory())
	fmt.Printf("  Max Concurrent: %d\n", orchConfig.Execution.MaxConcurrent)
	fmt.Printf("  Log Level: %s\n", orchConfig.Logging.Level)

	return nil
}

func runConfigSet(configFile, key, value string) error {
	// This is a simplified implementation
	// A full implementation would need to parse the key path and update the config
	
	fmt.Printf("🔧 Setting %s = %s\n", key, value)
	
	// For now, just show what would be done
	fmt.Printf("💡 Config modification not yet implemented\n")
	fmt.Printf("💡 Edit the configuration file manually:\n")
	
	if configFile != "" {
		fmt.Printf("   %s\n", configFile)
	} else {
		homeDir, _ := os.UserHomeDir()
		defaultConfig := filepath.Join(homeDir, ".corynth", "config.yaml")
		fmt.Printf("   %s\n", defaultConfig)
	}
	
	return nil
}

// Helper functions

func enabledString(enabled bool) string {
	if enabled {
		return "✅ enabled"
	}
	return "❌ disabled"
}