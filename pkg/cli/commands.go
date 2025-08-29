package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/corynth/corynth/pkg/plugin"
	"github.com/corynth/corynth/pkg/workflow"
	"github.com/spf13/cobra"
)

// NewValidateCommand creates the validate command
func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [workflow-file]",
		Short: "Validate workflow syntax and configuration",
		Long: `Validate checks workflow files for syntax errors and validates:
  - YAML syntax
  - Workflow structure
  - Plugin availability
  - Variable requirements
  - Step dependencies`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFile := ""
			if len(args) > 0 {
				workflowFile = args[0]
			}
			return runValidate(cmd, workflowFile)
		},
	}
	return cmd
}

func runValidate(cmd *cobra.Command, workflowFile string) error {
	if workflowFile == "" {
		workflowFile = findDefaultWorkflowFile()
		if workflowFile == "" {
			return fmt.Errorf("no workflow file found")
		}
	}

	// Load configuration
	cfg, err := loadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize plugin manager
	manager, err := initializePluginManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin manager: %w", err)
	}

	// Create workflow engine
	engine := workflow.NewEngine(manager)

	// Load and validate workflow
	fmt.Printf("%s %s...\n", Step("Validating"), Value(workflowFile))
	wf, err := engine.LoadWorkflow(workflowFile)
	if err != nil {
		fmt.Printf("%s %v\n", Failed("Validation failed"), err)
		return err
	}

	fmt.Printf("%s\n", Success("Workflow is valid"))
	fmt.Printf("%s %s\n", Label("Name:"), Value(wf.Name))
	fmt.Printf("%s %s\n", Label("Description:"), Value(wf.Description))
	fmt.Printf("%s %s\n", Label("Steps:"), Value(fmt.Sprintf("%d", len(wf.Steps))))
	
	return nil
}

// NewDestroyCommand creates the destroy command
func NewDestroyCommand() *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy resources created by workflow",
		Long: `Destroy removes resources that were created by previous workflow executions.
This command reverses the actions performed by the workflow.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDestroy(cmd, autoApprove)
		},
	}

	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")
	return cmd
}

func runDestroy(cmd *cobra.Command, autoApprove bool) error {
	fmt.Printf("%s\n", Info("Destroy functionality not yet implemented"))
	return nil
}

// NewPluginCommand creates the plugin command
func NewPluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
		Long:  "Manage Corynth plugins: list, install, update, and remove plugins.",
	}

	cmd.AddCommand(
		newPluginListCommand(),
		newPluginInstallCommand(),
		newPluginUpdateCommand(),
		newPluginRemoveCommand(),
		newPluginSearchCommand(),
		newPluginDiscoverCommand(),
		newPluginInfoCommand(),
		newPluginCategoriesCommand(),
		NewPluginInitCommand(),
		newPluginDoctorCommand(),
		newPluginCleanCommand(),
		newPluginSecurityCommand(),
		newPluginStatsCommand(),
	)

	return cmd
}

func newPluginListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			plugins := manager.List()
			if len(plugins) == 0 {
				fmt.Printf("%s\n", Info("No plugins installed"))
				return nil
			}

			fmt.Printf("%s %s\n", Header("Installed plugins"), DimText(fmt.Sprintf("(%d)", len(plugins))))
			for _, p := range plugins {
				fmt.Printf("%s %s %s - %s\n", 
					BulletPoint(""), 
					Label(p.Name), 
					DimText("v"+p.Version), 
					Value(p.Description))
			}
			return nil
		},
	}
}

func newPluginInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install <plugin-name>",
		Short: "Install a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			pluginName := args[0]
			fmt.Printf("%s %s...\n", Step("Installing plugin"), Value(pluginName))
			
			if err := manager.InstallFromRepository(pluginName); err != nil {
				return fmt.Errorf("failed to install plugin: %w", err)
			}

			fmt.Printf("%s\n", Success(fmt.Sprintf("Plugin %s installed successfully", pluginName)))
			return nil
		},
	}
}

func newPluginUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update <plugin-name>",
		Short: "Update a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			pluginName := args[0]
			fmt.Printf("%s %s...\n", Step("Updating plugin"), Value(pluginName))
			
			if err := manager.Update(pluginName); err != nil {
				return fmt.Errorf("failed to update plugin: %w", err)
			}

			fmt.Printf("%s\n", Success(fmt.Sprintf("Plugin %s updated successfully", pluginName)))
			return nil
		},
	}
}

func newPluginRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <plugin-name>",
		Short: "Remove a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			pluginName := args[0]
			fmt.Printf("%s %s...\n", Step("Removing plugin"), Value(pluginName))
			
			if err := manager.Remove(pluginName); err != nil {
				return fmt.Errorf("failed to remove plugin: %w", err)
			}

			fmt.Printf("%s\n", Success(fmt.Sprintf("Plugin %s removed successfully", pluginName)))
			return nil
		},
	}
}

func newPluginSearchCommand() *cobra.Command {
	var tags []string
	
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for plugins locally and in registries",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			// Search locally first
			fmt.Printf("%s\n", SubHeader("Installed plugins:"))
			results := manager.Search(query, tags)
			if len(results) == 0 {
				fmt.Printf("%s\n", DimText("  No local plugins found"))
			} else {
				for _, p := range results {
					fmt.Printf("%s %s %s - %s\n", 
						BulletPoint(""), 
						Label(p.Name), 
						DimText("v"+p.Version), 
						Value(p.Description))
				}
			}

			// Search in registry
			fmt.Printf("\n%s\n", SubHeader("Available in registry:"))
			registryResults, err := manager.SearchRegistry(query, tags)
			if err != nil {
				fmt.Printf("%s\n", DimText("  Registry not available"))
			} else if len(registryResults) == 0 {
				fmt.Printf("%s\n", DimText("  No plugins found in registry"))
			} else {
				for _, p := range registryResults {
					installed := ""
					if _, err := manager.Get(p.Name); err == nil {
						installed = Colorize(SuccessColor, " ✓ installed")
					}
					fmt.Printf("%s %s %s - %s%s\n", 
						BulletPoint(""), 
						Label(p.Name), 
						DimText("v"+p.Version), 
						Value(p.Description),
						installed)
				}
			}
			
			return nil
		},
	}
	
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Filter by tags")
	return cmd
}

func newPluginDiscoverCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "discover",
		Short: "Discover available plugins from registries",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", Header("Discovering plugins from registries..."))
			
			plugins, err := manager.ListAvailable()
			if err != nil {
				return fmt.Errorf("failed to fetch plugin registry: %w", err)
			}

			if len(plugins) == 0 {
				fmt.Printf("%s\n", Info("No plugins available"))
				return nil
			}

			// Get featured plugins
			featured, _ := manager.GetFeaturedPlugins()
			featuredMap := make(map[string]bool)
			for _, f := range featured {
				featuredMap[f] = true
			}

			fmt.Printf("\n%s %s\n", Header("Available plugins"), DimText(fmt.Sprintf("(%d total)", len(plugins))))
			
			// Group by installation status
			var installed, notInstalled []plugin.RegistryPlugin
			for _, p := range plugins {
				if _, err := manager.Get(p.Name); err == nil {
					installed = append(installed, p)
				} else {
					notInstalled = append(notInstalled, p)
				}
			}

			if len(notInstalled) > 0 {
				fmt.Printf("\n%s\n", SubHeader("Available to install:"))
				for _, p := range notInstalled {
					star := ""
					if featuredMap[p.Name] {
						star = Colorize(StepColor, " ★")
					}
					fmt.Printf("%s %s %s - %s%s\n", 
						BulletPoint(""), 
						Label(p.Name), 
						DimText(fmt.Sprintf("v%s, %s", p.Version, p.Size)), 
						Value(p.Description),
						star)
					if len(p.Tags) > 0 {
						fmt.Printf("    %s %s\n", DimText("Tags:"), DimText(strings.Join(p.Tags, ", ")))
					}
				}
			}

			if len(installed) > 0 {
				fmt.Printf("\n%s\n", SubHeader("Already installed:"))
				for _, p := range installed {
					fmt.Printf("%s %s %s %s\n", 
						BulletPoint(""), 
						Label(p.Name), 
						DimText("v"+p.Version),
						Colorize(SuccessColor, "✓"))
				}
			}

			fmt.Printf("\n%s\n", Info("Use 'corynth plugin install <name>' to install a plugin"))
			
			return nil
		},
	}
}

func newPluginInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info <plugin-name>",
		Short: "Show detailed plugin information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			pluginName := args[0]
			
			// Try to get from registry
			info, err := manager.GetPluginInfo(pluginName)
			if err != nil {
				// Try local plugin
				if p, err := manager.Get(pluginName); err == nil {
					meta := p.Metadata()
					fmt.Printf("%s %s\n", Label("Name:"), Value(meta.Name))
					fmt.Printf("%s %s\n", Label("Version:"), Value(meta.Version))
					fmt.Printf("%s %s\n", Label("Description:"), Value(meta.Description))
					fmt.Printf("%s %s\n", Label("Author:"), Value(meta.Author))
					if len(meta.Tags) > 0 {
						fmt.Printf("%s %s\n", Label("Tags:"), Value(strings.Join(meta.Tags, ", ")))
					}
					fmt.Printf("%s %s\n", Label("Status:"), Colorize(SuccessColor, "Installed locally"))
					
					// Show actions
					actions := p.Actions()
					if len(actions) > 0 {
						fmt.Printf("\n%s\n", SubHeader("Available actions:"))
						for _, action := range actions {
							fmt.Printf("%s %s - %s\n", 
								BulletPoint(""), 
								Label(action.Name), 
								Value(action.Description))
						}
					}
				} else {
					return fmt.Errorf("plugin '%s' not found", pluginName)
				}
			} else {
				// Show registry info
				fmt.Printf("%s %s\n", Label("Name:"), Value(info.Name))
				fmt.Printf("%s %s\n", Label("Version:"), Value(info.Version))
				fmt.Printf("%s %s\n", Label("Description:"), Value(info.Description))
				fmt.Printf("%s %s\n", Label("Author:"), Value(info.Author))
				fmt.Printf("%s %s\n", Label("Size:"), Value(info.Size))
				fmt.Printf("%s %s\n", Label("Format:"), Value(info.Format))
				
				if len(info.Tags) > 0 {
					fmt.Printf("%s %s\n", Label("Tags:"), Value(strings.Join(info.Tags, ", ")))
				}
				
				// Check if installed
				if _, err := manager.Get(pluginName); err == nil {
					fmt.Printf("%s %s\n", Label("Status:"), Colorize(SuccessColor, "✓ Installed"))
				} else {
					fmt.Printf("%s %s\n", Label("Status:"), DimText("Not installed"))
				}
				
				// Show actions
				if len(info.Actions) > 0 {
					fmt.Printf("\n%s\n", SubHeader("Available actions:"))
					for _, action := range info.Actions {
						fmt.Printf("%s %s - %s\n", 
							BulletPoint(""), 
							Label(action.Name), 
							Value(action.Description))
						if action.Example != "" {
							fmt.Printf("    %s %s\n", DimText("Example:"), DimText(action.Example))
						}
					}
				}
				
				// Show requirements
				if info.Requirements.Corynth != "" {
					fmt.Printf("\n%s\n", SubHeader("Requirements:"))
					fmt.Printf("%s Corynth %s\n", BulletPoint(""), info.Requirements.Corynth)
					if len(info.Requirements.OS) > 0 {
						fmt.Printf("%s OS: %s\n", BulletPoint(""), strings.Join(info.Requirements.OS, ", "))
					}
					if len(info.Requirements.Arch) > 0 {
						fmt.Printf("%s Architecture: %s\n", BulletPoint(""), strings.Join(info.Requirements.Arch, ", "))
					}
				}
			}
			
			return nil
		},
	}
}

func newPluginCategoriesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "categories",
		Short: "List plugin categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			categories, err := manager.GetCategories()
			if err != nil {
				return fmt.Errorf("failed to fetch categories: %w", err)
			}

			if len(categories) == 0 {
				fmt.Printf("%s\n", Info("No categories available"))
				return nil
			}

			fmt.Printf("%s\n", Header("Plugin Categories"))
			
			for category, plugins := range categories {
				fmt.Printf("\n%s %s\n", SubHeader(category), DimText(fmt.Sprintf("(%d plugins)", len(plugins))))
				for _, plugin := range plugins {
					// Check if installed
					installed := ""
					if _, err := manager.Get(plugin); err == nil {
						installed = Colorize(SuccessColor, " ✓")
					}
					fmt.Printf("%s %s%s\n", BulletPoint(""), Value(plugin), installed)
				}
			}
			
			return nil
		},
	}
}

// NewWorkflowCommand creates the workflow command
func NewWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage workflows",
		Long:  "Manage workflows: list, show, and validate workflow definitions.",
	}

	cmd.AddCommand(
		newWorkflowListCommand(),
		newWorkflowShowCommand(),
	)

	return cmd
}

func newWorkflowListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available workflows",
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFiles := findWorkflowFiles(".")
			if len(workflowFiles) == 0 {
				fmt.Printf("%s\n", Info("No workflow files found"))
				return nil
			}

			fmt.Printf("%s %s\n", Header("Available workflows"), DimText(fmt.Sprintf("(%d)", len(workflowFiles))))
			for _, file := range workflowFiles {
				fmt.Printf("%s %s\n", BulletPoint(""), Value(file))
			}
			return nil
		},
	}
}

func newWorkflowShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <workflow-file>",
		Short: "Show workflow details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFile := args[0]
			
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			engine := workflow.NewEngine(manager)
			wf, err := engine.LoadWorkflow(workflowFile)
			if err != nil {
				return fmt.Errorf("failed to load workflow: %w", err)
			}

			fmt.Printf("%s %s\n", Label("Name:"), Value(wf.Name))
			fmt.Printf("%s %s\n", Label("Description:"), Value(wf.Description))
			fmt.Printf("%s %s\n", Label("Version:"), Value(wf.Version))
			fmt.Printf("%s %s\n", Label("Steps:"), Value(fmt.Sprintf("%d", len(wf.Steps))))
			
			if len(wf.Variables) > 0 {
				fmt.Printf("\n%s\n", SubHeader("Variables:"))
				for name, variable := range wf.Variables {
					required := ""
					if variable.Required {
						required = Colorize(WarningColor, " (required)")
					}
					typeStr := fmt.Sprintf("%v", variable.Type)
					fmt.Printf("%s %s: %s%s\n", 
						BulletPoint(""), 
						Label(name), 
						Value(typeStr), 
						required)
				}
			}

			return nil
		},
	}
}

// NewStateCommand creates the state command
func NewStateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage workflow execution state",
		Long:  "Manage workflow execution state: list executions, show details, and cleanup.",
	}

	cmd.AddCommand(
		newStateListCommand(),
		newStateShowCommand(),
	)

	return cmd
}

func newStateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List workflow executions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			stateStore, err := createStateStore(cfg)
			if err != nil {
				return err
			}

			executions, err := stateStore.ListExecutions()
			if err != nil {
				return fmt.Errorf("failed to list executions: %w", err)
			}

			if len(executions) == 0 {
				fmt.Printf("%s\n", Info("No executions found"))
				return nil
			}

			fmt.Printf("%s %s\n", Header("Recent executions"), DimText(fmt.Sprintf("(%d)", len(executions))))
			for _, exec := range executions {
				duration := ""
				if exec.EndTime != nil {
					duration = exec.EndTime.Sub(exec.StartTime).Round(time.Second).String()
				} else {
					duration = Colorize(StepColor, "running")
				}
				
				statusColor := SuccessColor
				if exec.Status == "failed" {
					statusColor = ErrorColor
				} else if exec.Status == "running" {
					statusColor = StepColor
				}
				
				fmt.Printf("%s %s: %s %s - %s\n", 
					BulletPoint(""), 
					DimText(exec.ID[:8]), 
					Value(exec.WorkflowName), 
					Colorize(statusColor, fmt.Sprintf("(%s)", exec.Status)), 
					DimText(duration))
			}
			return nil
		},
	}
}

func newStateShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <execution-id>",
		Short: "Show execution details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			stateStore, err := createStateStore(cfg)
			if err != nil {
				return err
			}

			executionID := args[0]
			state, err := stateStore.LoadExecution(executionID)
			if err != nil {
				return fmt.Errorf("failed to load execution: %w", err)
			}

			fmt.Printf("%s %s\n", Label("Execution ID:"), Value(state.ID))
			fmt.Printf("%s %s\n", Label("Workflow:"), Value(state.WorkflowName))
			
			statusColor := SuccessColor
			if state.Status == "failed" {
				statusColor = ErrorColor
			} else if state.Status == "running" {
				statusColor = StepColor
			}
			fmt.Printf("%s %s\n", Label("Status:"), Colorize(statusColor, string(state.Status)))
			
			fmt.Printf("%s %s\n", Label("Started:"), Value(state.StartTime.Format(time.RFC3339)))
			if state.EndTime != nil {
				fmt.Printf("%s %s\n", Label("Completed:"), Value(state.EndTime.Format(time.RFC3339)))
				fmt.Printf("%s %s\n", Label("Duration:"), Value(state.EndTime.Sub(state.StartTime).Round(time.Second).String()))
			}

			if len(state.Steps) > 0 {
				fmt.Printf("\n%s\n", SubHeader("Steps:"))
				for _, step := range state.Steps {
					stepStatus := string(step.Status)
					stepStatusColor := SuccessColor
					if step.Error != nil {
						stepStatus = fmt.Sprintf("%s (%v)", stepStatus, step.Error)
						stepStatusColor = ErrorColor
					} else if step.Status == "running" {
						stepStatusColor = StepColor
					}
					fmt.Printf("%s %s: %s\n", 
						BulletPoint(""), 
						Label(step.Name), 
						Colorize(stepStatusColor, stepStatus))
				}
			}

			return nil
		},
	}
}

// NewConfigCommand creates the config command
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Manage Corynth configuration: validate and show current settings.",
	}

	cmd.AddCommand(
		newConfigValidateCommand(),
		newConfigShowCommand(),
	)

	return cmd
}

func newConfigValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			fmt.Printf("%s %s...\n", Step("Validating config"), Value(configPath))
			
			cfg, err := loadConfig(configPath)
			if err != nil {
				fmt.Printf("%s %v\n", Failed("Config load failed"), err)
				return fmt.Errorf("failed to load config: %w", err)
			}

			if err := cfg.Validate(); err != nil {
				fmt.Printf("%s %v\n", Failed("Configuration is invalid"), err)
				return err
			}

			fmt.Printf("%s\n", Success("Configuration is valid"))
			return nil
		},
	}
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			cfg, err := loadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Printf("%s %s\n", Label("Configuration file:"), Value(configPath))
			fmt.Printf("%s %s\n", Label("Version:"), Value(cfg.Version))
			
			if cfg.Notifications != nil {
				status := map[bool]string{true: "enabled", false: "disabled"}[cfg.Notifications.Enabled]
				fmt.Printf("%s %s\n", Label("Notifications:"), Value(status))
			}
			
			if cfg.Plugins != nil {
				autoInstall := map[bool]string{true: "enabled", false: "disabled"}[cfg.Plugins.AutoInstall]
				fmt.Printf("%s %s\n", Label("Plugin auto-install:"), Value(autoInstall))
				fmt.Printf("%s %s\n", Label("Plugin repositories:"), Value(fmt.Sprintf("%d", len(cfg.Plugins.Repositories))))
			}

			return nil
		},
	}
}

// newPluginSecurityCommand creates the plugin security command
func newPluginSecurityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security [plugin-name]",
		Short: "Show plugin security information",
		Long:  "Display detailed security information for a plugin including trust level, checksums, and scan results.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			pluginName := args[0]
			securityInfo, err := manager.GetPluginSecurityInfo(pluginName)
			if err != nil {
				return fmt.Errorf("failed to get security info for plugin '%s': %w", pluginName, err)
			}

			fmt.Printf("%s %s\n", Header("Security Information for"), Success(pluginName))
			fmt.Printf("%s %s\n", Label("Trust Level:"), getTrustLevelDisplay(securityInfo.TrustLevel))
			fmt.Printf("%s %s\n", Label("Publisher:"), Value(securityInfo.Publisher))
			fmt.Printf("%s %s\n", Label("Verified:"), getBoolDisplay(securityInfo.Verified))
			
			if securityInfo.SHA256 != "" {
				fmt.Printf("%s %s\n", Label("SHA256:"), Value(securityInfo.SHA256))
			}
			
			if securityInfo.ScannedAt != "" {
				fmt.Printf("%s %s\n", Label("Last Scan:"), Value(securityInfo.ScannedAt))
			}
			
			if len(securityInfo.ScanResults) > 0 {
				fmt.Printf("%s\n", Label("Scan Results:"))
				for _, result := range securityInfo.ScanResults {
					fmt.Printf("  %s %s\n", BulletPoint(""), Value(result))
				}
			}
			
			if len(securityInfo.AuditTrail) > 0 {
				fmt.Printf("%s\n", Label("Audit Trail:"))
				for _, entry := range securityInfo.AuditTrail {
					fmt.Printf("  %s %s by %s at %s\n", 
						BulletPoint(""), 
						Success(entry.Action),
						Value(entry.Actor),
						Value(entry.Timestamp))
					if entry.Details != "" {
						fmt.Printf("    %s\n", DimText(entry.Details))
					}
				}
			}

			return nil
		},
	}
	return cmd
}

// newPluginStatsCommand creates the plugin statistics command
func newPluginStatsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show plugin registry security statistics",
		Long:  "Display security statistics for all plugins in the registry including trust levels and verification status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig("")
			if err != nil {
				return err
			}

			manager, err := initializePluginManager(cfg)
			if err != nil {
				return err
			}

			stats, err := manager.GetSecurityStats()
			if err != nil {
				return fmt.Errorf("failed to get security statistics: %w", err)
			}

			fmt.Printf("%s\n", Header("Plugin Registry Security Statistics"))
			fmt.Printf("%s %s\n", Label("Total Plugins:"), Value(fmt.Sprintf("%d", stats["total"])))
			fmt.Printf("%s %s\n", Label("Official:"), Colorize(SuccessColor, fmt.Sprintf("%d", stats["official"])))
			fmt.Printf("%s %s\n", Label("Verified:"), Colorize(InfoColor, fmt.Sprintf("%d", stats["verified"])))
			fmt.Printf("%s %s\n", Label("Community:"), Colorize(WarningColor, fmt.Sprintf("%d", stats["community"])))
			fmt.Printf("%s %s\n", Label("Unscanned:"), Colorize(ErrorColor, fmt.Sprintf("%d", stats["unscanned"])))
			
			// Calculate percentages
			total := stats["total"]
			if total > 0 {
				fmt.Printf("\n%s\n", Header("Trust Level Distribution"))
				fmt.Printf("%s %s%%\n", Label("Official:"), Value(fmt.Sprintf("%.1f", float64(stats["official"])*100/float64(total))))
				fmt.Printf("%s %s%%\n", Label("Verified:"), Value(fmt.Sprintf("%.1f", float64(stats["verified"])*100/float64(total))))
				fmt.Printf("%s %s%%\n", Label("Community:"), Value(fmt.Sprintf("%.1f", float64(stats["community"])*100/float64(total))))
				fmt.Printf("%s %s%%\n", Label("Unscanned:"), Value(fmt.Sprintf("%.1f", float64(stats["unscanned"])*100/float64(total))))
			}

			return nil
		},
	}
	return cmd
}

// Helper functions for security display

func getTrustLevelDisplay(trustLevel string) string {
	switch trustLevel {
	case "official":
		return Colorize(SuccessColor, "Official")
	case "verified":
		return Colorize(InfoColor, "Verified")
	case "community":
		return Colorize(WarningColor, "Community")
	default:
		return Colorize(ErrorColor, "Unknown")
	}
}

func getBoolDisplay(value bool) string {
	if value {
		return Colorize(SuccessColor, "Yes")
	}
	return Colorize(ErrorColor, "No")
}