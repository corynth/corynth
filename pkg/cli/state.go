package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/corynth/corynth/pkg/config"
	"github.com/corynth/corynth/pkg/state"
	"github.com/corynth/corynth/pkg/workflow"
	"github.com/spf13/cobra"
)

// NewOrchestrationStateCommand creates the orchestration state management command group
func NewOrchestrationStateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage workflow state and execution history",
		Long: `State commands provide visibility into workflow execution history,
state persistence, and orchestration data.

These commands help with debugging, monitoring, and understanding
workflow execution patterns.`,
	}

	cmd.AddCommand(NewStateListCommand())
	cmd.AddCommand(NewStateShowCommand())
	cmd.AddCommand(NewStateCleanupCommand())
	cmd.AddCommand(NewStateHistoryCommand())
	cmd.AddCommand(NewStateOutputsCommand())

	return cmd
}

// NewStateListCommand lists all workflow executions
func NewStateListCommand() *cobra.Command {
	var configFile string
	var workflowFilter string
	var statusFilter string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflow execution states",
		Long: `List shows all workflow executions with their status, timing,
and basic information. This provides an overview of workflow activity.`,
		Example: `  # List all executions
  corynth state list
  
  # Filter by workflow name
  corynth state list --workflow data-processing
  
  # Filter by status  
  corynth state list --status success
  
  # Limit results
  corynth state list --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateList(configFile, workflowFilter, statusFilter, limit)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")
	cmd.Flags().StringVar(&workflowFilter, "workflow", "", "Filter by workflow name")
	cmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status (success, failure, running)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of results")

	return cmd
}

// NewStateShowCommand shows details for a specific execution
func NewStateShowCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "show <execution-id>",
		Short: "Show detailed information for a workflow execution",
		Long: `Show displays complete information about a specific workflow execution,
including variables, outputs, timing, and step-by-step execution details.`,
		Example: `  # Show execution details
  corynth state show execution-123
  
  # Use custom config
  corynth state show --config ./corynth.yaml execution-123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateShow(configFile, args[0])
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")

	return cmd
}

// NewStateCleanupCommand cleans up old workflow states
func NewStateCleanupCommand() *cobra.Command {
	var configFile string
	var maxAge string
	var dryRun bool
	var force bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up old workflow execution states",
		Long: `Cleanup removes old workflow execution states to free up storage space.
By default, only states older than the configured retention period are removed.`,
		Example: `  # Clean up states older than 7 days
  corynth state cleanup --max-age 7d
  
  # Dry run to see what would be cleaned
  corynth state cleanup --max-age 7d --dry-run
  
  # Force cleanup without confirmation
  corynth state cleanup --max-age 1d --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateCleanup(configFile, maxAge, dryRun, force)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")
	cmd.Flags().StringVar(&maxAge, "max-age", "30d", "Maximum age of states to keep")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be cleaned without removing")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

// NewStateHistoryCommand shows execution history for a workflow
func NewStateHistoryCommand() *cobra.Command {
	var configFile string
	var limit int
	var format string

	cmd := &cobra.Command{
		Use:   "history <workflow-name>",
		Short: "Show execution history for a specific workflow",
		Long: `History displays the execution history for a specific workflow,
showing trends, success rates, and execution patterns over time.`,
		Example: `  # Show history for a workflow
  corynth state history data-processing
  
  # Limit to recent executions
  corynth state history --limit 20 data-processing
  
  # JSON output for scripting
  corynth state history --format json data-processing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateHistory(configFile, args[0], limit, format)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of history entries")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// NewStateOutputsCommand shows outputs from a workflow execution
func NewStateOutputsCommand() *cobra.Command {
	var configFile string
	var format string

	cmd := &cobra.Command{
		Use:   "outputs <workflow-name>",
		Short: "Show outputs from the latest execution of a workflow",
		Long: `Outputs displays the outputs from the most recent successful execution
of a workflow. These outputs are available for import by other workflows.`,
		Example: `  # Show workflow outputs
  corynth state outputs data-processing
  
  # JSON format for scripting
  corynth state outputs --format json data-processing`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStateOutputs(configFile, args[0], format)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// Implementation functions

func runStateList(configFile, workflowFilter, statusFilter string, limit int) error {
	// Load configuration
	orchConfig, err := config.LoadOrchestrationConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize state backend
	stateBackend, err := state.NewStateBackend(orchConfig.State.Backend, orchConfig.GetBackendConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize state backend: %w", err)
	}

	// Get all states
	states, err := stateBackend.ListStates()
	if err != nil {
		return fmt.Errorf("failed to list states: %w", err)
	}

	// Apply filters
	filteredStates := filterStates(states, workflowFilter, statusFilter)

	// Sort by start time (newest first)
	sort.Slice(filteredStates, func(i, j int) bool {
		return filteredStates[i].StartTime.After(filteredStates[j].StartTime)
	})

	// Apply limit
	if limit > 0 && len(filteredStates) > limit {
		filteredStates = filteredStates[:limit]
	}

	// Display results
	if len(filteredStates) == 0 {
		fmt.Println("No workflow executions found")
		return nil
	}

	fmt.Printf("Found %d workflow execution(s)\n\n", len(filteredStates))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EXECUTION ID\tWORKFLOW\tSTATUS\tSTARTED\tDURATION")
	fmt.Fprintln(w, "------------\t--------\t------\t-------\t--------")

	for _, state := range filteredStates {
		duration := calculateDuration(state)
		startTime := state.StartTime.Format("2006-01-02 15:04:05")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			truncateString(state.ID, 20),
			truncateString(state.WorkflowName, 25),
			formatStatus(state.Status),
			startTime,
			duration)
	}

	w.Flush()
	return nil
}

func runStateShow(configFile, executionID string) error {
	// Load configuration
	orchConfig, err := config.LoadOrchestrationConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize state backend
	stateBackend, err := state.NewStateBackend(orchConfig.State.Backend, orchConfig.GetBackendConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize state backend: %w", err)
	}

	// Load specific state
	state, err := stateBackend.LoadState(executionID)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Display detailed information
	fmt.Printf("📋 Workflow Execution Details\n\n")
	fmt.Printf("Execution ID: %s\n", state.ID)
	fmt.Printf("Workflow:     %s\n", state.WorkflowName)
	fmt.Printf("Status:       %s\n", formatStatus(state.Status))
	fmt.Printf("Started:      %s\n", state.StartTime.Format(time.RFC3339))

	if state.EndTime != nil {
		fmt.Printf("Ended:        %s\n", state.EndTime.Format(time.RFC3339))
		fmt.Printf("Duration:     %s\n", state.EndTime.Sub(state.StartTime))
	} else {
		fmt.Printf("Duration:     %s (ongoing)\n", time.Since(state.StartTime))
	}

	fmt.Printf("Mode:         %s\n", state.ExecutionMode)

	if state.TriggeredBy != "" {
		fmt.Printf("Triggered by: %s\n", state.TriggeredBy)
	}

	// Show variables
	if len(state.Variables) > 0 {
		fmt.Printf("\n📊 Variables:\n")
		for k, v := range state.Variables {
			fmt.Printf("  %s = %v\n", k, v)
		}
	}

	// Show outputs
	if len(state.Outputs) > 0 {
		fmt.Printf("\n📈 Outputs:\n")
		for k, v := range state.Outputs {
			fmt.Printf("  %s = %v\n", k, v)
		}
	}

	// Show steps
	if len(state.Steps) > 0 {
		fmt.Printf("\n🔄 Steps:\n")
		for _, step := range state.Steps {
			status := formatStatus(step.Status)
			duration := ""
			if step.EndTime != nil {
				duration = fmt.Sprintf(" (%s)", step.EndTime.Sub(step.StartTime))
			}
			fmt.Printf("  %s: %s%s\n", step.Name, status, duration)
			
			if step.Error != nil {
				fmt.Printf("    Error: %v\n", step.Error)
			}
		}
	}

	// Show error if any
	if state.Error != nil {
		fmt.Printf("\n❌ Error: %v\n", state.Error)
	}

	return nil
}

func runStateCleanup(configFile, maxAgeStr string, dryRun, force bool) error {
	// Parse max age
	maxAge, err := time.ParseDuration(maxAgeStr)
	if err != nil {
		return fmt.Errorf("invalid max-age format: %w", err)
	}

	// Load configuration
	orchConfig, err := config.LoadOrchestrationConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize state backend
	stateBackend, err := state.NewStateBackend(orchConfig.State.Backend, orchConfig.GetBackendConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize state backend: %w", err)
	}

	if dryRun {
		fmt.Printf("🧪 Dry run: would clean up states older than %s\n", maxAge)
		// TODO: Implement dry run logic to show what would be cleaned
		fmt.Printf("💡 Use --force to actually perform cleanup\n")
		return nil
	}

	// Confirm unless forced
	if !force {
		fmt.Printf("⚠️  About to clean up states older than %s\n", maxAge)
		fmt.Printf("   This will permanently delete old execution history.\n")
		fmt.Print("\nProceed? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Cleanup cancelled")
			return nil
		}
	}

	// Perform cleanup
	fmt.Printf("🧹 Cleaning up states older than %s...\n", maxAge)
	err = stateBackend.CleanupOldStates(maxAge)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	fmt.Println("✅ Cleanup completed successfully")
	return nil
}

func runStateHistory(configFile, workflowName string, limit int, format string) error {
	// Load configuration
	orchConfig, err := config.LoadOrchestrationConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize state backend
	stateBackend, err := state.NewStateBackend(orchConfig.State.Backend, orchConfig.GetBackendConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize state backend: %w", err)
	}

	// Get history for workflow
	history, err := stateBackend.FindStatesByWorkflow(workflowName)
	if err != nil {
		return fmt.Errorf("failed to get workflow history: %w", err)
	}

	if len(history) == 0 {
		fmt.Printf("No execution history found for workflow: %s\n", workflowName)
		return nil
	}

	// Sort by start time (newest first)
	sort.Slice(history, func(i, j int) bool {
		return history[i].StartTime.After(history[j].StartTime)
	})

	// Apply limit
	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	// Display based on format
	if format == "json" {
		// TODO: Implement JSON output
		fmt.Printf("JSON output not yet implemented\n")
		return nil
	}

	// Table format
	fmt.Printf("📜 Execution History for %s (%d entries)\n\n", workflowName, len(history))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EXECUTION ID\tSTATUS\tSTARTED\tDURATION\tTRIGGERED BY")
	fmt.Fprintln(w, "------------\t------\t-------\t--------\t------------")

	for _, state := range history {
		duration := calculateDuration(state)
		startTime := state.StartTime.Format("2006-01-02 15:04")
		triggeredBy := state.TriggeredBy
		if triggeredBy == "" {
			triggeredBy = "manual"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			truncateString(state.ID, 20),
			formatStatus(state.Status),
			startTime,
			duration,
			truncateString(triggeredBy, 15))
	}

	w.Flush()
	return nil
}

func runStateOutputs(configFile, workflowName string, format string) error {
	// Load configuration
	orchConfig, err := config.LoadOrchestrationConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize state backend
	stateBackend, err := state.NewStateBackend(orchConfig.State.Backend, orchConfig.GetBackendConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize state backend: %w", err)
	}

	// Get workflow outputs
	outputs, err := stateBackend.LoadWorkflowOutput(workflowName)
	if err != nil {
		return fmt.Errorf("failed to load workflow outputs: %w", err)
	}

	if format == "json" {
		// TODO: Implement JSON output
		fmt.Printf("JSON output not yet implemented\n")
		return nil
	}

	// Table format
	fmt.Printf("📈 Latest Outputs for %s\n", workflowName)
	fmt.Printf("Generated: %s\n\n", outputs.Timestamp.Format(time.RFC3339))

	if len(outputs.Outputs) == 0 {
		fmt.Println("No outputs available")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVALUE")
	fmt.Fprintln(w, "----\t-----")

	for k, v := range outputs.Outputs {
		fmt.Fprintf(w, "%s\t%v\n", k, v)
	}

	w.Flush()
	return nil
}

// Helper functions

func filterStates(states []workflow.ExecutionState, workflowFilter, statusFilter string) []workflow.ExecutionState {
	var filtered []workflow.ExecutionState

	for _, state := range states {
		// Apply workflow filter
		if workflowFilter != "" && state.WorkflowName != workflowFilter {
			continue
		}

		// Apply status filter
		if statusFilter != "" && strings.ToLower(string(state.Status)) != strings.ToLower(statusFilter) {
			continue
		}

		filtered = append(filtered, state)
	}

	return filtered
}

func formatStatus(status workflow.Status) string {
	switch status {
	case workflow.StatusSuccess:
		return "✅ success"
	case workflow.StatusFailure:
		return "❌ failure"
	case workflow.StatusRunning:
		return "🔄 running"
	case workflow.StatusPending:
		return "⏳ pending"
	case workflow.StatusSkipped:
		return "⏭️  skipped"
	case workflow.StatusCancelled:
		return "🚫 cancelled"
	default:
		return string(status)
	}
}

func calculateDuration(state workflow.ExecutionState) string {
	if state.EndTime != nil {
		duration := state.EndTime.Sub(state.StartTime)
		return formatDuration(duration)
	} else if state.Status == workflow.StatusRunning {
		duration := time.Since(state.StartTime)
		return formatDuration(duration) + " (ongoing)"
	}
	return "unknown"
}

// formatDuration is available from common.go

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}