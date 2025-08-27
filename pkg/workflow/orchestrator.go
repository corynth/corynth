package workflow

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/corynth/corynth/pkg/types"
)

// WorkflowEngine defines the interface for workflow execution
type WorkflowEngine interface {
	LoadWorkflowFile(path string) (*Workflow, error)
	Execute(ctx context.Context, workflowName string, variables map[string]interface{}, mode ExecutionMode) (*ExecutionState, error)
}

// WorkflowStateManager defines the interface for workflow state management
type WorkflowStateManager interface {
	SaveWorkflowOutput(workflowName string, outputs map[string]interface{}) error
	LoadWorkflowOutput(workflowName string) (*types.WorkflowOutput, error)
	FindStatesByWorkflow(workflowName string) ([]ExecutionState, error)
}

// Orchestrator manages workflow execution chains and dependencies
type Orchestrator struct {
	engine       WorkflowEngine
	stateManager WorkflowStateManager
}

// NewOrchestrator creates a new workflow orchestrator
func NewOrchestrator(engine WorkflowEngine, stateManager WorkflowStateManager) *Orchestrator {
	return &Orchestrator{
		engine:       engine,
		stateManager: stateManager,
	}
}

// ExecuteWorkflowChain executes a workflow and its dependencies/triggers
func (o *Orchestrator) ExecuteWorkflowChain(workflowFile string, variables map[string]interface{}) (*ExecutionState, error) {
	// Load and parse the workflow
	workflow, err := o.engine.LoadWorkflowFile(workflowFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow: %w", err)
	}

	// Process workflow dependencies first
	for _, dep := range workflow.DependsOnWorkflow {
		if err := o.executeDependency(&dep, variables); err != nil {
			return nil, fmt.Errorf("failed to execute dependency %s: %w", dep.WorkflowFile, err)
		}
		
		// Import variables from dependency if specified
		if dep.ImportAll || len(dep.ImportVars) > 0 {
			importedVars, err := o.importVariablesFromWorkflow(dep.WorkflowFile, dep.ImportVars, dep.ImportAll)
			if err != nil {
				return nil, fmt.Errorf("failed to import variables from %s: %w", dep.WorkflowFile, err)
			}
			
			// Merge imported variables (dependency variables take precedence)
			for k, v := range importedVars {
				variables[k] = v
			}
		}
	}

	// Execute the main workflow
	ctx := context.Background()
	state, err := o.engine.Execute(ctx, workflow.Name, variables, ModeApply)
	if err != nil {
		return nil, fmt.Errorf("failed to execute workflow: %w", err)
	}

	// Save workflow outputs for potential use by other workflows
	workflowOutputs := make(map[string]interface{})
	
	// First try to get outputs from execution state
	if len(state.Outputs) > 0 {
		for k, v := range state.Outputs {
			workflowOutputs[k] = v
		}
	}
	
	// If no execution outputs, check if workflow definition has outputs
	if len(workflowOutputs) == 0 && len(workflow.Outputs) > 0 {
		// Process workflow-level outputs defined in HCL
		// Since the engine doesn't support custom outputs yet, we'll use the static values
		for k, v := range workflow.Outputs {
			workflowOutputs[k] = v
		}
	}
	
	// If still no outputs, extract from successful step outputs  
	if len(workflowOutputs) == 0 {
		for _, step := range state.Steps {
			if step.Status == StatusSuccess && step.Outputs != nil {
				for k, v := range step.Outputs {
					// Only include non-standard outputs (not stdout, stderr, exit_code)
					if k != "stdout" && k != "stderr" && k != "exit_code" {
						workflowOutputs[k] = v
					}
				}
			}
		}
	}
	
	// Save outputs if we have any
	if len(workflowOutputs) > 0 {
		if err := o.stateManager.SaveWorkflowOutput(workflow.Name, workflowOutputs); err != nil {
			// Log but don't fail - this is for convenience, not critical
			fmt.Printf("Warning: failed to save workflow outputs: %v\n", err)
		}
	}

	// Process workflow triggers on success/failure
	if state.Status == StatusSuccess {
		for _, trigger := range workflow.TriggerWorkflows {
			if trigger.OnSuccess || (!trigger.OnSuccess && !trigger.OnFailure) {
				if err := o.executeTrigger(&trigger, state); err != nil {
					fmt.Printf("Warning: failed to execute success trigger %s: %v\n", trigger.WorkflowFile, err)
				}
			}
		}
	} else if state.Status == StatusFailure {
		for _, trigger := range workflow.TriggerWorkflows {
			if trigger.OnFailure {
				if err := o.executeTrigger(&trigger, state); err != nil {
					fmt.Printf("Warning: failed to execute failure trigger %s: %v\n", trigger.WorkflowFile, err)
				}
			}
		}
	}

	return state, nil
}

// executeDependency executes a workflow dependency
func (o *Orchestrator) executeDependency(dep *WorkflowDependency, parentVariables map[string]interface{}) error {
	// Build variables for the dependency
	depVariables := make(map[string]interface{})
	
	// Start with parent variables as base (if not overridden)
	for k, v := range parentVariables {
		depVariables[k] = v
	}
	
	// Override with dependency-specific variables
	for k, v := range dep.Variables {
		depVariables[k] = v
	}

	// Execute the dependency workflow
	depWorkflow, err := o.engine.LoadWorkflowFile(dep.WorkflowFile)
	if err != nil {
		if dep.Required {
			return fmt.Errorf("required dependency not found: %w", err)
		}
		// If not required, skip silently
		return nil
	}

	ctx := context.Background()
	depState, err := o.engine.Execute(ctx, depWorkflow.Name, depVariables, ModeApply)
	if err != nil {
		if dep.Required {
			return fmt.Errorf("required dependency failed: %w", err)
		}
		// If not required, log but continue
		fmt.Printf("Warning: optional dependency %s failed: %v\n", dep.WorkflowFile, err)
		return nil
	}
	
	// Save dependency outputs for potential use by other workflows
	if depState.Status == StatusSuccess {
		workflowOutputs := make(map[string]interface{})
		
		// First try to get outputs from execution state
		if len(depState.Outputs) > 0 {
			for k, v := range depState.Outputs {
				workflowOutputs[k] = v
			}
		}
		
		// If no execution outputs, check if workflow definition has outputs
		if len(workflowOutputs) == 0 && len(depWorkflow.Outputs) > 0 {
			// Process workflow-level outputs defined in HCL
			for k, v := range depWorkflow.Outputs {
				workflowOutputs[k] = v
			}
		}
		
		// If still no outputs, extract from successful step outputs
		if len(workflowOutputs) == 0 {
			for _, step := range depState.Steps {
				if step.Status == StatusSuccess && step.Outputs != nil {
					for k, v := range step.Outputs {
						// Only include non-standard outputs
						if k != "stdout" && k != "stderr" && k != "exit_code" {
							workflowOutputs[k] = v
						}
					}
				}
			}
		}
		
		// Save outputs if we have any
		if len(workflowOutputs) > 0 {
			if err := o.stateManager.SaveWorkflowOutput(depWorkflow.Name, workflowOutputs); err != nil {
				fmt.Printf("Warning: failed to save dependency outputs: %v\n", err)
			}
		}
	}

	return nil
}

// executeTrigger executes a workflow trigger
func (o *Orchestrator) executeTrigger(trigger *WorkflowTrigger, parentState *ExecutionState) error {
	// Build variables for the trigger
	triggerVariables := make(map[string]interface{})
	
	// Export specified variables from parent
	if trigger.ExportAll {
		// Export all outputs and variables
		for k, v := range parentState.Outputs {
			triggerVariables[k] = v
		}
		for k, v := range parentState.Variables {
			triggerVariables[k] = v
		}
	} else if len(trigger.ExportVars) > 0 {
		// Export only specified variables
		for _, varName := range trigger.ExportVars {
			if val, exists := parentState.Outputs[varName]; exists {
				triggerVariables[varName] = val
			} else if val, exists := parentState.Variables[varName]; exists {
				triggerVariables[varName] = val
			}
		}
	}
	
	// Override with trigger-specific variables
	for k, v := range trigger.Variables {
		triggerVariables[k] = v
	}

	// Execute the trigger workflow
	triggerWorkflow, err := o.engine.LoadWorkflowFile(trigger.WorkflowFile)
	if err != nil {
		return fmt.Errorf("trigger workflow not found: %w", err)
	}

	// Execute in background or synchronously based on configuration
	ctx := context.Background()
	_, err = o.engine.Execute(ctx, triggerWorkflow.Name, triggerVariables, ModeApply)
	return err
}

// importVariablesFromWorkflow imports variables from a previously executed workflow
func (o *Orchestrator) importVariablesFromWorkflow(workflowFile string, importVars []string, importAll bool) (map[string]interface{}, error) {
	// Load the workflow to get the actual workflow name from HCL
	workflow, err := o.engine.LoadWorkflowFile(workflowFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow %s: %w", workflowFile, err)
	}
	
	// Load outputs from the workflow (dependency should have been executed already)
	output, err := o.stateManager.LoadWorkflowOutput(workflow.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to load outputs from workflow %s: %w", workflow.Name, err)
	}

	importedVars := make(map[string]interface{})
	
	if importAll {
		// Import all available outputs
		for k, v := range output.Outputs {
			importedVars[k] = v
		}
	} else {
		// Import only specified variables
		for _, varName := range importVars {
			if val, exists := output.Outputs[varName]; exists {
				importedVars[varName] = val
			}
		}
	}

	return importedVars, nil
}

// extractWorkflowName extracts the workflow name from a file path
func (o *Orchestrator) extractWorkflowName(workflowFile string) string {
	base := filepath.Base(workflowFile)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// GetWorkflowHistory returns execution history for a workflow
func (o *Orchestrator) GetWorkflowHistory(workflowName string) ([]ExecutionState, error) {
	return o.stateManager.FindStatesByWorkflow(workflowName)
}

// GetWorkflowOutputs returns the latest outputs from a workflow
func (o *Orchestrator) GetWorkflowOutputs(workflowName string) (map[string]interface{}, error) {
	output, err := o.stateManager.LoadWorkflowOutput(workflowName)
	if err != nil {
		return nil, err
	}
	
	return output.Outputs, nil
}