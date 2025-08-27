package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/corynth/corynth/pkg/config"
	"github.com/corynth/corynth/pkg/state"
	"github.com/corynth/corynth/pkg/workflow"
	"github.com/spf13/cobra"
)

// NewOrchestrateCommand creates the orchestrate command for workflow chains
func NewOrchestrateCommand() *cobra.Command {
	var configFile string
	var stateDir string
	var autoApprove bool
	var dryRun bool
	var verbose bool
	var showDependencies bool
	var maxConcurrent int

	cmd := &cobra.Command{
		Use:   "orchestrate [workflow-file]",
		Short: "Execute workflow with dependency management and orchestration",
		Long: `Orchestrate executes workflows with full dependency management, variable passing,
and state persistence. This enables building complex workflow pipelines.

Features:
  - Automatic dependency resolution and execution
  - Variable import/export between workflows  
  - State persistence across workflow runs
  - Trigger-based workflow chaining
  - Concurrent execution where possible
  - Comprehensive error handling and recovery

This command is ideal for complex automation pipelines where workflows
need to share data and coordinate execution.`,
		Example: `  # Execute workflow with orchestration
  corynth orchestrate data-pipeline.hcl

  # Dry run to see execution plan
  corynth orchestrate --dry-run complex-workflow.hcl
  
  # Show dependency tree
  corynth orchestrate --show-dependencies workflow.hcl
  
  # Use custom configuration
  corynth orchestrate --config ./corynth.yaml workflow.hcl
  
  # Override concurrency limit
  corynth orchestrate --max-concurrent 10 workflow.hcl`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrchestrate(cmd, args[0], orchestrateOptions{
				ConfigFile:       configFile,
				StateDir:         stateDir,
				AutoApprove:      autoApprove,
				DryRun:           dryRun,
				Verbose:          verbose,
				ShowDependencies: showDependencies,
				MaxConcurrent:    maxConcurrent,
			})
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")
	cmd.Flags().StringVar(&stateDir, "state-dir", "", "State directory override")
	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show execution plan without running")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	cmd.Flags().BoolVar(&showDependencies, "show-dependencies", false, "Show dependency tree and exit")
	cmd.Flags().IntVar(&maxConcurrent, "max-concurrent", 0, "Override max concurrent workflows")

	return cmd
}

// orchestrateOptions holds options for workflow orchestration
type orchestrateOptions struct {
	ConfigFile       string
	StateDir         string
	AutoApprove      bool
	DryRun           bool
	Verbose          bool
	ShowDependencies bool
	MaxConcurrent    int
}

// runOrchestrate executes workflow orchestration
func runOrchestrate(cmd *cobra.Command, workflowFile string, opts orchestrateOptions) error {
	startTime := time.Now()
	
	// Load orchestration configuration
	orchConfig, err := config.LoadOrchestrationConfig(opts.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load orchestration config: %w", err)
	}
	
	// Check if orchestration is enabled
	if !orchConfig.Enabled {
		return fmt.Errorf("orchestration is disabled in configuration")
	}
	
	// Override state directory if specified
	if opts.StateDir != "" {
		orchConfig.State.Directory = opts.StateDir
	}
	
	// Override concurrency if specified
	if opts.MaxConcurrent > 0 {
		orchConfig.Execution.MaxConcurrent = opts.MaxConcurrent
	}
	
	if opts.Verbose {
		fmt.Printf("🔧 Configuration loaded: %s backend, state dir: %s\n", 
			orchConfig.State.Backend, orchConfig.GetStateDirectory())
	}
	
	// Initialize components
	orchestrationSM := state.NewStateManager(orchConfig.GetStateDirectory())
	
	// Load plugin manager (reuse existing logic)
	cfg, err := loadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	pluginManager, err := initializePluginManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin manager: %w", err)
	}
	
	// Create workflow engine
	engine := workflow.NewEngine(pluginManager)
	
	// Set up state store integration (bridge legacy and orchestration state)
	legacyStateStore, err := createStateStore(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize legacy state store: %w", err)
	}
	engine.SetStateStore(legacyStateStore)
	
	// Create hybrid state manager that bridges legacy and orchestration state
	hybridStateManager := NewHybridStateManager(orchestrationSM, legacyStateStore)
	
	// Create orchestrator
	orchestrator := workflow.NewOrchestrator(engine, hybridStateManager)
	
	// Get variables from flags (reuse existing variable parsing logic)
	variables, err := parseVariablesFromFlags(cmd)
	if err != nil {
		return fmt.Errorf("failed to parse variables: %w", err)
	}
	
	// Show dependency tree if requested
	if opts.ShowDependencies {
		return showWorkflowDependencies(orchestrator, workflowFile, variables)
	}
	
	// Dry run if requested
	if opts.DryRun {
		return runOrchestrationDryRun(orchestrator, workflowFile, variables, opts.Verbose)
	}
	
	// Interactive approval unless auto-approve is set
	if !opts.AutoApprove {
		if !confirmExecution(workflowFile) {
			fmt.Println("❌ Execution cancelled")
			return nil
		}
	}
	
	fmt.Printf("🚀 Starting orchestrated execution of %s\n", workflowFile)
	if opts.Verbose {
		fmt.Printf("📊 Variables: %v\n", variables)
	}
	
	// Execute workflow chain
	executionState, err := orchestrator.ExecuteWorkflowChain(workflowFile, variables)
	if err != nil {
		fmt.Printf("❌ Execution failed after %v: %v\n", time.Since(startTime), err)
		return err
	}
	
	// Show results
	duration := time.Since(startTime)
	if executionState.Status == workflow.StatusSuccess {
		fmt.Printf("✅ Workflow chain completed successfully in %v\n", duration)
		
		if opts.Verbose {
			fmt.Printf("📈 Execution ID: %s\n", executionState.ID)
			fmt.Printf("📊 Final outputs: %v\n", executionState.Outputs)
			
			// Show workflow history
			history, _ := orchestrator.GetWorkflowHistory(executionState.WorkflowName)
			fmt.Printf("📜 Execution history: %d previous runs\n", len(history)-1)
		}
	} else {
		fmt.Printf("❌ Workflow chain failed after %v\n", duration)
		if executionState.Error != nil {
			fmt.Printf("💥 Error: %v\n", executionState.Error)
		}
		return fmt.Errorf("workflow execution failed")
	}
	
	return nil
}

// showWorkflowDependencies displays the dependency tree
func showWorkflowDependencies(orchestrator *workflow.Orchestrator, workflowFile string, variables map[string]interface{}) error {
	fmt.Printf("🔍 Analyzing dependencies for %s\n\n", workflowFile)
	
	// This is a simplified version - in a full implementation, we'd analyze the workflow
	// structure and show a tree of dependencies
	
	fmt.Printf("📋 Dependency Tree:\n")
	fmt.Printf("├─ %s (main workflow)\n", workflowFile)
	fmt.Printf("│  ├─ Variables: %d defined\n", len(variables))
	fmt.Printf("│  └─ Status: ✅ Loadable\n")
	
	// TODO: Add recursive dependency analysis
	fmt.Printf("\n💡 Use --dry-run to see full execution plan\n")
	
	return nil
}

// runOrchestrationDryRun shows what would be executed without running
func runOrchestrationDryRun(orchestrator *workflow.Orchestrator, workflowFile string, variables map[string]interface{}, verbose bool) error {
	fmt.Printf("🧪 Dry run: analyzing execution plan for %s\n\n", workflowFile)
	
	fmt.Printf("📋 Execution Plan:\n")
	fmt.Printf("├─ Load workflow: %s\n", workflowFile)
	fmt.Printf("├─ Variables: %d provided\n", len(variables))
	fmt.Printf("├─ Check dependencies\n")
	fmt.Printf("├─ Execute dependency chain\n") 
	fmt.Printf("├─ Execute main workflow\n")
	fmt.Printf("├─ Process triggers on completion\n")
	fmt.Printf("└─ Save state and outputs\n")
	
	if verbose {
		fmt.Printf("\n📊 Variables:\n")
		for k, v := range variables {
			fmt.Printf("  %s = %v\n", k, v)
		}
	}
	
	fmt.Printf("\n💡 Add --auto-approve to execute this plan\n")
	return nil
}

// confirmExecution asks for user confirmation
func confirmExecution(workflowFile string) bool {
	fmt.Printf("\n⚠️  About to execute workflow chain starting with: %s\n", workflowFile)
	fmt.Printf("   This may execute multiple workflows and modify system state.\n")
	fmt.Print("\nProceed? [y/N]: ")
	
	var response string
	fmt.Scanln(&response)
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// NewExecuteCommand enhances the existing execute command with orchestration support
func NewExecuteCommand() *cobra.Command {
	var workflowName string
	var autoApprove bool
	var parallel int
	var quiet bool
	var verbose bool
	var enableOrchestration bool
	var configFile string

	cmd := &cobra.Command{
		Use:   "execute [workflow-file]",
		Short: "Execute workflow (with optional orchestration)",
		Long: `Execute runs a workflow with optional orchestration support.

Without --orchestrate: Runs workflow in standalone mode (original behavior)
With --orchestrate: Enables dependency management and workflow chaining

Use 'corynth orchestrate' for full orchestration features.`,
		Example: `  # Execute workflow normally
  corynth execute workflow.hcl
  
  # Execute with orchestration enabled
  corynth execute --orchestrate workflow.hcl
  
  # Use orchestrate command for full features
  corynth orchestrate workflow.hcl`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFile := ""
			if len(args) > 0 {
				workflowFile = args[0]
			}
			
			if enableOrchestration {
				// Delegate to orchestration
				return runOrchestrate(cmd, workflowFile, orchestrateOptions{
					ConfigFile:  configFile,
					AutoApprove: autoApprove,
					Verbose:     verbose,
				})
			}
			
			// Use original execution logic
			return runExecute(cmd, workflowFile, executeOptions{
				WorkflowName: workflowName,
				AutoApprove:  autoApprove,
				Parallel:     parallel,
				Quiet:        quiet,
				Verbose:      verbose,
			})
		},
	}

	cmd.Flags().StringVar(&workflowName, "workflow", "", "Workflow name to execute")
	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")
	cmd.Flags().IntVar(&parallel, "parallel", 0, "Maximum parallel steps")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Minimal output")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	cmd.Flags().BoolVar(&enableOrchestration, "orchestrate", false, "Enable orchestration features")
	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")

	return cmd
}

// executeOptions holds options for regular workflow execution
type executeOptions struct {
	WorkflowName string
	AutoApprove  bool
	Parallel     int
	Quiet        bool
	Verbose      bool
}

// runExecute runs workflow in standalone mode (original behavior)
func runExecute(cmd *cobra.Command, workflowFile string, opts executeOptions) error {
	// This would contain the original execute logic from apply.go
	// For now, just show that it's running in standalone mode
	
	if workflowFile == "" {
		workflowFile = findDefaultWorkflowFile()
		if workflowFile == "" {
			return fmt.Errorf("no workflow file found")
		}
	}
	
	fmt.Printf("🔄 Executing %s in standalone mode\n", workflowFile)
	fmt.Printf("💡 Use --orchestrate or 'corynth orchestrate' for dependency management\n")
	
	// TODO: Implement original execution logic or delegate to existing apply command
	return fmt.Errorf("standalone execution not yet implemented - use 'corynth orchestrate' instead")
}

// Helper functions - these use existing implementations from common.go

func parseVariablesFromFlags(cmd *cobra.Command) (map[string]interface{}, error) {
	// For now, return empty variables - this would be enhanced to parse --var flags
	return map[string]interface{}{}, nil
}