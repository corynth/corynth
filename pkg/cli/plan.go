package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/corynth/corynth/pkg/plugin"
	"github.com/corynth/corynth/pkg/workflow"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// NewPlanCommand creates the plan command
func NewPlanCommand() *cobra.Command {
	var workflowName string
	var out string
	var detailed bool

	cmd := &cobra.Command{
		Use:   "plan [workflow-file]",
		Short: "Show execution plan for workflow",
		Long: `Plan analyzes workflow files and shows what actions will be performed.

This command:
  - Parses workflow files and validates syntax
  - Resolves plugin dependencies
  - Shows step execution order and dependencies
  - Estimates execution time
  - Validates required variables and secrets`,
		Example: `  # Plan execution for default workflow
  corynth plan

  # Plan execution for specific workflow file
  corynth plan workflow.hcl

  # Plan with detailed output
  corynth plan --detailed

  # Save plan to file
  corynth plan --out plan.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowFile := ""
			if len(args) > 0 {
				workflowFile = args[0]
			}
			return runPlan(cmd, workflowFile, workflowName, out, detailed)
		},
	}

	cmd.Flags().StringVarP(&workflowName, "name", "n", "", "Workflow name to plan")
	cmd.Flags().StringVarP(&out, "out", "o", "", "Save plan to file")
	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed plan information")

	return cmd
}

func runPlan(cmd *cobra.Command, workflowFile, workflowName, out string, detailed bool) error {
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

	// Set up governance components for access control checking
	// TODO: Add audit logging and access control in future version

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
	fmt.Printf("%s %s...\n", Step("Loading workflow from"), Value(workflowFile))
	wf, err := engine.LoadWorkflow(workflowFile)
	if err != nil {
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	// Use workflow name from file if not specified
	if workflowName == "" {
		workflowName = wf.Name
	}

	// Execute in plan mode
	fmt.Printf("\n%s\n", Header("Planning workflow execution..."))
	_, err = engine.Execute(cmd.Context(), workflowName, variables, workflow.ModePlan)
	if err != nil {
		return fmt.Errorf("planning failed: %w", err)
	}

	// Generate plan
	plan, err := engine.Plan(workflowName, variables)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// Display plan
	displayPlan(plan, wf, detailed)

	// Save plan if requested
	if out != "" {
		if err := savePlan(plan, out); err != nil {
			return fmt.Errorf("failed to save plan: %w", err)
		}
		fmt.Printf("\n%s %s\n", Success("Plan saved to"), Value(out))
	}

	// Display summary
	fmt.Printf("\n%s\n", Header("Plan Summary"))
	fmt.Printf("%s\n", strings.Repeat("=", 12))
	fmt.Printf("%s %s\n", Label("Workflow:"), Value(wf.Name))
	fmt.Printf("%s %s\n", Label("Description:"), Value(wf.Description))
	fmt.Printf("%s %s\n", Label("Steps:"), Value(fmt.Sprintf("%d", len(plan.Steps))))
	if len(wf.Parallel) > 0 {
		fmt.Printf("%s %s\n", Label("Parallel Groups:"), Value(fmt.Sprintf("%d", len(wf.Parallel))))
	}
	fmt.Printf("%s %s\n", Label("Estimated Duration:"), Value(estimateTotalDuration(plan).String()))

	// Show warnings
	warnings := validatePlan(plan, wf, manager)
	if len(warnings) > 0 {
		fmt.Printf("\n%s\n", SubHeader("Warnings:"))
		for _, warning := range warnings {
			fmt.Printf("%s %s\n", BulletPoint(""), Warning(warning))
		}
	}

	fmt.Printf("\n%s %s\n", Info("To execute this plan, run:"), Command("corynth apply"))

	return nil
}

// displayPlan displays the execution plan
func displayPlan(plan *workflow.Plan, wf *workflow.Workflow, detailed bool) {
	fmt.Printf("\n%s '%s':\n", Header("Execution Plan for"), Value(plan.WorkflowName))
	fmt.Printf("%s\n", strings.Repeat("=", 24))

	if detailed {
		displayDetailedPlan(plan, wf)
	} else {
		displaySimplePlan(plan)
	}
}

// displaySimplePlan shows a simple plan view
func displaySimplePlan(plan *workflow.Plan) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Step", "Plugin", "Action", "Est. Time"})
	table.SetBorder(false)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
	)

	for i, step := range plan.Steps {
		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			step.Name,
			step.Plugin,
			step.Action,
			step.Estimated.String(),
		})
	}

	table.Render()
}

// displayDetailedPlan shows a detailed plan view
func displayDetailedPlan(plan *workflow.Plan, wf *workflow.Workflow) {
	for i, step := range plan.Steps {
		fmt.Printf("\n%d. %s\n", i+1, Header(step.Name))
		fmt.Printf("   %s %s\n", Label("Plugin:"), Value(step.Plugin))
		fmt.Printf("   %s %s\n", Label("Action:"), Value(step.Action))
		fmt.Printf("   %s %s\n", Label("Estimated Time:"), Value(step.Estimated.String()))
		
		if len(step.Dependencies) > 0 {
			fmt.Printf("   %s %s\n", Label("Dependencies:"), Value(strings.Join(step.Dependencies, ", ")))
		}

		// Find original step for more details
		for _, originalStep := range wf.Steps {
			if originalStep.Name == step.Name {
				if originalStep.Condition != "" {
					fmt.Printf("   %s %s\n", Label("Condition:"), Value(originalStep.Condition))
				}
				if originalStep.RetryPolicy != nil {
					fmt.Printf("   %s %d attempts, %s delay\n", 
						Label("Retry Policy:"), 
						originalStep.RetryPolicy.MaxAttempts, 
						originalStep.RetryPolicy.Delay)
				}
				if originalStep.Timeout != "" {
					fmt.Printf("   %s %s\n", Label("Timeout:"), Value(originalStep.Timeout))
				}
				break
			}
		}
	}
}

// validatePlan validates the plan and returns warnings
func validatePlan(plan *workflow.Plan, wf *workflow.Workflow, manager *plugin.Manager) []string {
	var warnings []string

	// Check plugin availability
	pluginsSeen := make(map[string]bool)
	for _, step := range plan.Steps {
		if step.Plugin == "" {
			continue
		}
		
		if !pluginsSeen[step.Plugin] {
			pluginsSeen[step.Plugin] = true
			if _, err := manager.Get(step.Plugin); err != nil {
				warnings = append(warnings, fmt.Sprintf("Plugin '%s' not found", step.Plugin))
			}
		}
	}

	// Check for long-running workflows
	totalDuration := estimateTotalDuration(plan)
	if totalDuration.Minutes() > 30 {
		warnings = append(warnings, "Workflow estimated to run for more than 30 minutes")
	}

	// Check for missing variables
	for name, variable := range wf.Variables {
		if variable.Required && plan.Variables[name] == nil {
			warnings = append(warnings, fmt.Sprintf("Required variable '%s' not provided", name))
		}
	}

	// Check for circular dependencies
	if hasCircularDependencies(plan) {
		warnings = append(warnings, "Circular dependencies detected in workflow steps")
	}

	return warnings
}

// hasCircularDependencies checks for circular dependencies
func hasCircularDependencies(plan *workflow.Plan) bool {
	// Simple cycle detection using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	
	// Build dependency graph
	deps := make(map[string][]string)
	for _, step := range plan.Steps {
		deps[step.Name] = step.Dependencies
	}

	var hasCycle func(string) bool
	hasCycle = func(node string) bool {
		if recursionStack[node] {
			return true
		}
		if visited[node] {
			return false
		}

		visited[node] = true
		recursionStack[node] = true

		for _, dep := range deps[node] {
			if hasCycle(dep) {
				return true
			}
		}

		recursionStack[node] = false
		return false
	}

	for step := range deps {
		if hasCycle(step) {
			return true
		}
	}

	return false
}

// estimateTotalDuration estimates total workflow duration
func estimateTotalDuration(plan *workflow.Plan) time.Duration {
	// Simple estimation: sum of all step durations
	// In reality, parallel steps would run concurrently
	var total time.Duration
	for _, step := range plan.Steps {
		total += step.Estimated
	}
	return total
}

// savePlan saves the plan to a file
func savePlan(plan *workflow.Plan, filename string) error {
	// Create plan output directory if needed
	dir := filepath.Dir(filename)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// For now, save as JSON
	// TODO: Support other formats based on file extension
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

