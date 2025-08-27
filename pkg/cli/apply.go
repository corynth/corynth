package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/corynth/corynth/pkg/workflow"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// NewApplyCommand creates the apply command
func NewApplyCommand() *cobra.Command {
	var workflowName string
	var autoApprove bool
	var parallel int
	var quiet bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "apply [workflow-file]",
		Short: "Execute workflow",
		Long: `Apply executes the workflow and applies all planned changes.

This command:
  - Loads and validates the workflow
  - Executes steps in the correct order
  - Handles dependencies and parallel execution
  - Provides real-time progress updates
  - Sends notifications on success/failure`,
		Example: `  # Apply default workflow
  corynth apply

  # Apply specific workflow file
  corynth apply workflow.hcl

  # Auto-approve without confirmation
  corynth apply --auto-approve

  # Show only command outputs
  corynth apply --quiet --auto-approve

  # Limit parallel execution
  corynth apply --parallel 2`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFile := ""
			if len(args) > 0 {
				workflowFile = args[0]
			}
			return runApply(cmd, workflowFile, workflowName, autoApprove, parallel, quiet, verbose)
		},
	}

	cmd.Flags().StringVarP(&workflowName, "name", "n", "", "Workflow name to execute")
	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "Skip interactive approval")
	cmd.Flags().IntVar(&parallel, "parallel", 0, "Maximum parallel executions (0 = unlimited)")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Show only command outputs (implies --auto-approve)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed execution information")

	return cmd
}

func runApply(cmd *cobra.Command, workflowFile, workflowName string, autoApprove bool, parallel int, quiet bool, verbose bool) error {
	// Load configuration
	configPath, _ := cmd.Flags().GetString("config")
	cfg, err := loadConfig(configPath)
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

	// Set up notifications if configured
	if cfg.Notifications != nil && cfg.Notifications.Enabled {
		notifier, err := createNotifier(cfg.Notifications)
		if err != nil {
			fmt.Printf("%s Failed to initialize notifications: %v\n", Warning(""), err)
		} else {
			engine.SetNotifier(notifier)
		}
	}

	// Set up state store
	stateStore, err := createStateStore(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize state store: %w", err)
	}
	engine.SetStateStore(stateStore)

	// TODO: Set up governance components
	// auditLogger, err := createAuditLogger(cfg)
	// if err != nil {
	// 	fmt.Printf("%s Failed to initialize audit logging: %v\n", Warning(""), err)
	// } else {
	// 	engine.SetAuditLogger(auditLogger)
	// }

	// accessController := createAccessController(auditLogger)
	// engine.SetAccessController(accessController)

	// Load variables
	variables, err := loadVariables(cmd)
	if err != nil {
		return fmt.Errorf("failed to load variables: %w", err)
	}

	// Find workflow file if not specified
	if workflowFile == "" {
		workflowFile = findDefaultWorkflowFile()
		if workflowFile == "" {
			return fmt.Errorf("no workflow file found. Create workflow.hcl or specify file path")
		}
	}

	// Load workflow
	if !quiet {
		fmt.Printf("%s %s...\n", Step("Loading workflow from"), Value(workflowFile))
	}
	wf, err := engine.LoadWorkflow(workflowFile)
	if err != nil {
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	// Use workflow name from file if not specified
	if workflowName == "" {
		workflowName = wf.Name
	}

	// Generate plan first
	plan, err := engine.Plan(workflowName, variables)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// Show plan summary (only if not quiet)
	if !quiet {
		fmt.Printf("\n%s %s\n", Label("Workflow:"), Header(wf.Name))
		fmt.Printf("%s %s\n", Label("Description:"), Value(wf.Description))
		fmt.Printf("%s %s\n", Label("Steps to execute:"), Value(fmt.Sprintf("%d", len(plan.Steps))))
		fmt.Printf("%s %s\n", Label("Estimated duration:"), Value(estimateTotalDuration(plan).String()))

		// Show step list
		fmt.Printf("\n%s\n", SubHeader("Steps:"))
		for i, step := range plan.Steps {
			fmt.Printf("%s %d. %s (%s.%s)\n", 
				BulletPoint(""), 
				i+1, 
				Value(step.Name), 
				DimText(step.Plugin), 
				DimText(step.Action))
		}
	}

	// Quiet mode implies auto-approve
	if quiet {
		autoApprove = true
	}
	
	// Ask for confirmation unless auto-approve
	if !autoApprove {
		if !quiet {
			fmt.Printf("\n%s ", Info("Do you want to execute this workflow? (yes/no):"))
		}
		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			if !quiet {
				fmt.Printf("%s\n", Info("Execution cancelled."))
			}
			return nil
		}
	}

	// Execute workflow
	if !quiet {
		fmt.Printf("\n%s\n", Header("Starting workflow execution..."))
	}
	
	startTime := time.Now()
	state, err := executeWithProgress(engine, workflowName, variables, len(plan.Steps))
	duration := time.Since(startTime)

	if err != nil {
		if !quiet {
			fmt.Printf("\n%s\n", Failed("Workflow execution failed!"))
			fmt.Printf("%s %v\n", Label("Error:"), err)
			fmt.Printf("%s %s\n", Label("Duration:"), Value(duration.Round(time.Second).String()))
			
			// Show failed step details
			if state != nil {
				for _, step := range state.Steps {
					if step.Status == workflow.StatusFailure {
						fmt.Printf("%s %s\n", Label("Failed step:"), Error(step.Name))
						if step.Error != nil {
							fmt.Printf("%s %v\n", Label("Error:"), step.Error)
						}
						break
					}
				}
			}
		}
		return err
	}

	// Show step console outputs first (both modes)
	for _, step := range state.Steps {
		if step.Outputs != nil {
			if stdout, ok := step.Outputs["stdout"]; ok && stdout != "" {
				fmt.Print(stdout)
				if !strings.HasSuffix(stdout.(string), "\n") {
					fmt.Print("\n")
				}
			}
		}
	}

	if quiet {
		// Quiet mode: No workflow details, just console output
		// (console output already shown above)
	} else if verbose {
		// Verbose mode: Full workflow details + execution summary
		fmt.Printf("\n%s\n", Success("Workflow executed successfully!"))
		fmt.Printf("%s %s\n", Label("Duration:"), Value(duration.Round(time.Second).String()))
		fmt.Printf("%s %s/%s\n", Label("Steps completed:"), Value(fmt.Sprintf("%d", countCompletedSteps(state))), Value(fmt.Sprintf("%d", len(plan.Steps))))

		// Show outputs
		if len(state.Outputs) > 0 {
			fmt.Printf("\n%s\n", SubHeader("Outputs:"))
			for key, value := range state.Outputs {
				fmt.Printf("%s %s = %v\n", BulletPoint(""), Label(key), Value(fmt.Sprintf("%v", value)))
			}
		}

		// Show execution summary
		fmt.Printf("\n%s\n", Header("Execution Summary"))
		fmt.Printf("%s\n", strings.Repeat("=", 17))
		fmt.Printf("%s %s\n", Label("Workflow ID:"), Value(state.ID))
		fmt.Printf("%s %s\n", Label("Started:"), Value(state.StartTime.Format(time.RFC3339)))
		fmt.Printf("%s %s\n", Label("Completed:"), Value(state.EndTime.Format(time.RFC3339)))
		fmt.Printf("%s %s\n", Label("Status:"), Success(string(state.Status)))
	} else {
		// Normal mode: Basic workflow summary
		fmt.Printf("\n%s %s\n", Success("✓ Workflow completed:"), Value(wf.Name))
		fmt.Printf("%s %s\n", Label("Duration:"), Value(duration.Round(time.Second).String()))
		fmt.Printf("%s %s/%s\n", Label("Steps:"), Value(fmt.Sprintf("%d", countCompletedSteps(state))), Value(fmt.Sprintf("%d", len(plan.Steps))))
	}

	return nil
}

// executeWithProgress executes the workflow with a progress bar
func executeWithProgress(engine *workflow.Engine, name string, variables map[string]interface{}, totalSteps int) (*workflow.ExecutionState, error) {
	// Create progress bar
	bar := progressbar.NewOptions(totalSteps,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetDescription("Executing workflow..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)

	// Channel to track progress
	progressCh := make(chan string, 10)
	doneCh := make(chan struct{})

	// Start progress updater
	go func() {
		stepCount := 0
		for {
			select {
			case stepName := <-progressCh:
				stepCount++
				bar.Describe(fmt.Sprintf("Executing: %s", stepName))
				bar.Set(stepCount)
			case <-doneCh:
				bar.Finish()
				return
			}
		}
	}()

	// Create a notifier that updates progress
	_ = &ProgressNotifier{
		progressCh: progressCh,
	}

	// Note: In a real implementation, we'd need to modify the engine
	// to support multiple notifiers or chaining

	// Execute workflow
	state, err := engine.Execute(context.Background(), name, variables, workflow.ModeApply)

	// Signal completion
	close(doneCh)

	return state, err
}

// ProgressNotifier updates progress during execution
type ProgressNotifier struct {
	progressCh chan<- string
}

func (p *ProgressNotifier) NotifyStart(workflow string, execution *workflow.ExecutionState) {
	// Implementation would send to progress channel
}

func (p *ProgressNotifier) NotifyStepStart(workflow string, step string, execution *workflow.ExecutionState) {
	if p.progressCh != nil {
		select {
		case p.progressCh <- step:
		default:
		}
	}
}

func (p *ProgressNotifier) NotifyStepComplete(workflow string, step string, execution *workflow.ExecutionState) {
	// Could update with completion status
}

func (p *ProgressNotifier) NotifyComplete(workflow string, execution *workflow.ExecutionState) {
	// Final notification
}

func (p *ProgressNotifier) NotifyFailure(workflow string, execution *workflow.ExecutionState, err error) {
	// Failure notification
}

// countCompletedSteps counts successfully completed steps
func countCompletedSteps(state *workflow.ExecutionState) int {
	count := 0
	for _, step := range state.Steps {
		if step.Status == workflow.StatusSuccess {
			count++
		}
	}
	return count
}

// TODO: Re-implement governance functions when governance package is available
// createAuditLogger creates an audit logger based on configuration
// func createAuditLogger(cfg *config.Config) (governance.AuditLogger, error) {
// 	// For now, create a file-based audit logger
// 	// In production, this would be configurable
// 	logDir := ".corynth/logs"
// 	if cfg.State != nil && cfg.State.BackendConfig != nil {
// 		if path, ok := cfg.State.BackendConfig["path"]; ok {
// 			logDir = filepath.Join(filepath.Dir(path), "logs")
// 		}
// 	}
// 	
// 	logPath := filepath.Join(logDir, "audit.log")
// 	return governance.NewFileAuditLogger(logPath)
// }

// createAccessController creates an access controller with audit logging
// func createAccessController(auditLogger governance.AuditLogger) governance.AccessController {
// 	// For now, use the AllowAll controller as requested
// 	// In production, this would be configurable to use RBAC
// 	return governance.NewAllowAllAccessController(auditLogger)
// }