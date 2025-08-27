package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/corynth/corynth/pkg/config"
	"github.com/corynth/corynth/pkg/plugin"
	"github.com/corynth/corynth/pkg/workflow"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	var upgrade bool
	var backend string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Corynth configuration and plugins",
		Long: `Initialize prepares your working directory for Corynth workflows.

This command performs the following actions:
  - Creates .corynth directory for local state
  - Downloads and installs required plugins
  - Validates configuration files
  - Sets up state backend`,
		Example: `  # Initialize with default settings
  corynth init

  # Initialize with backend configuration
  corynth init --backend s3

  # Upgrade plugins during init
  corynth init --upgrade`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, upgrade, backend)
		},
	}

	cmd.Flags().BoolVar(&upgrade, "upgrade", false, "Upgrade plugins to latest version")
	cmd.Flags().StringVar(&backend, "backend", "local", "State backend type (local, s3, consul)")

	return cmd
}

func runInit(cmd *cobra.Command, upgrade bool, backend string) error {
	// Print initialization header
	fmt.Printf("%s\n", Header("Initializing Corynth..."))

	// Get config file path
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = "corynth.hcl"
	}

	// Load configuration
	fmt.Printf("%s %s...\n", Step("Loading configuration from"), Value(configPath))
	cfg, err := loadConfig(configPath)
	if err != nil {
		// Check if config doesn't exist, create a default one
		if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
			fmt.Printf("%s\n", Info("No configuration found, creating default..."))
			if err := createDefaultConfig(configPath); err != nil {
				return fmt.Errorf("failed to create default config: %w", err)
			}
			cfg, err = loadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load default config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	fmt.Printf("%s\n", Success("Configuration validated"))

	// Create .corynth directory
	corynthDir := ".corynth"
	if err := os.MkdirAll(corynthDir, 0755); err != nil {
		return fmt.Errorf("failed to create .corynth directory: %w", err)
	}
	
	// Create subdirectories
	dirs := []string{
		filepath.Join(corynthDir, "state"),
		filepath.Join(corynthDir, "plugins"),
		filepath.Join(corynthDir, "cache"),
		filepath.Join(corynthDir, "logs"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	fmt.Printf("%s\n", Success("Created .corynth directory structure"))

	// Initialize plugin manager
	fmt.Printf("%s\n", Step("Initializing plugin system..."))
	pluginPath := filepath.Join(corynthDir, "plugins")
	cachePath := filepath.Join(corynthDir, "cache")
	
	autoInstall := true
	if cfg.Plugins != nil {
		autoInstall = cfg.Plugins.AutoInstall
		if cfg.Plugins.LocalPath != "" {
			pluginPath = cfg.Plugins.LocalPath
		}
		if cfg.Plugins.Cache != nil && cfg.Plugins.Cache.Path != "" {
			cachePath = cfg.Plugins.Cache.Path
		}
	}

	manager := plugin.NewManager(pluginPath, cachePath, autoInstall)

	// Add configured repositories
	if cfg.Plugins != nil {
		for _, repo := range cfg.Plugins.Repositories {
			manager.AddRepository(plugin.Repository{
				Name:     repo.Name,
				URL:      repo.URL,
				Branch:   repo.Branch,
				Token:    repo.Token,
				Priority: repo.Priority,
			})
		}
	}

	// Load local plugins
	if err := manager.LoadLocal(); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	// List loaded plugins
	plugins := manager.List()
	fmt.Printf("%s %s\n", Header("Loaded plugins"), DimText(fmt.Sprintf("(%d)", len(plugins))))
	for _, p := range plugins {
		fmt.Printf("%s %s %s - %s\n", 
			BulletPoint(""), 
			Label(p.Name), 
			DimText("v"+p.Version), 
			Value(p.Description))
	}
	fmt.Printf("%s\n", Success("Plugin system initialized"))

	// Load workflow files
	fmt.Printf("%s\n", Step("Scanning for workflow files..."))
	workflowFiles := findWorkflowFiles(".")
	if len(workflowFiles) > 0 {
		fmt.Printf("%s %s\n", Header("Found workflow files"), DimText(fmt.Sprintf("(%d)", len(workflowFiles))))
		engine := workflow.NewEngine(manager)
		for _, file := range workflowFiles {
			wf, err := engine.LoadWorkflow(file)
			if err != nil {
				fmt.Printf("%s %s: %v\n", BulletPoint(""), Warning(file), err)
			} else {
				fmt.Printf("%s %s: %s\n", BulletPoint(""), Success(file), Value(wf.Name))
			}
		}
	} else {
		fmt.Printf("%s\n", Info("No workflow files found"))
	}

	// Initialize state backend
	fmt.Printf("%s %s state backend...\n", Step("Initializing"), Value(backend))
	if err := initializeStateBackend(backend, cfg); err != nil {
		return fmt.Errorf("failed to initialize state backend: %w", err)
	}
	fmt.Printf("%s\n", Success("State backend initialized"))

	// Upgrade plugins if requested
	if upgrade {
		fmt.Printf("%s\n", Step("Checking for plugin updates..."))
		for _, p := range plugins {
			fmt.Printf("%s %s...\n", Step("Checking"), Value(p.Name))
			if err := manager.Update(p.Name); err != nil {
				fmt.Printf("    %s Failed to update: %v\n", Warning(""), err)
			} else {
				fmt.Printf("    %s\n", Success("Updated to latest version"))
			}
		}
	}

	// Print summary
	fmt.Println()
	fmt.Printf("%s\n", Success("Corynth has been successfully initialized!"))
	fmt.Println()
	fmt.Printf("%s\n", SubHeader("You can now run workflows with:"))
	fmt.Printf("%s %s - %s\n", BulletPoint(""), Command("corynth sample"), Value("Generate sample workflows"))
	fmt.Printf("%s %s - %s\n", BulletPoint(""), Command("corynth plan"), Value("Preview workflow execution"))
	fmt.Printf("%s %s - %s\n", BulletPoint(""), Command("corynth apply"), Value("Execute workflow"))
	fmt.Println()
	fmt.Printf("%s\n", Info("To get started, generate a sample workflow file."))

	return nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(path string) error {
	defaultConfig := `# Corynth Configuration File
# This file configures Corynth plugins and state management

plugins {
  local_path = ".corynth/plugins"
  auto_install = true
  
  cache {
    path = ".corynth/cache"
  }
}

state {
  backend = "local"
  backend_config = {
    path = ".corynth/state"
  }
}
`

	return os.WriteFile(path, []byte(defaultConfig), 0644)
}

// findWorkflowFiles finds all workflow YAML files
func findWorkflowFiles(root string) []string {
	var files []string
	
	patterns := []string{
		"workflow.hcl",
		"workflow.tf",
		"*.workflow.hcl",
		"*.workflow.tf",
		"workflows/*.hcl",
		"workflows/*.tf",
		// Legacy YAML support
		"workflow.yaml",
		"workflow.yml",
		"*.workflow.yaml",
		"*.workflow.yml",
		"workflows/*.yaml",
		"workflows/*.yml",
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		files = append(files, matches...)
	}

	return files
}

// initializeStateBackend initializes the state backend
func initializeStateBackend(backend string, cfg *config.Config) error {
	stateDir := ".corynth/state"
	
	switch backend {
	case "local":
		// Create local state file
		stateFile := filepath.Join(stateDir, "corynth.tfstate")
		if _, err := os.Stat(stateFile); os.IsNotExist(err) {
			// Create empty state file
			initialState := `{
  "version": 1,
  "executions": []
}`
			if err := os.WriteFile(stateFile, []byte(initialState), 0644); err != nil {
				return fmt.Errorf("failed to create state file: %w", err)
			}
		}
	case "s3":
		// TODO: Initialize S3 backend
		fmt.Printf("%s\n", Info("S3 backend configuration required in corynth.hcl"))
	case "consul":
		// TODO: Initialize Consul backend
		fmt.Printf("%s\n", Info("Consul backend configuration required in corynth.hcl"))
	default:
		return fmt.Errorf("unsupported backend: %s", backend)
	}

	return nil
}