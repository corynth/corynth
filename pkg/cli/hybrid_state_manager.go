package cli

import (
	"fmt"

	"github.com/corynth/corynth/pkg/state"
	"github.com/corynth/corynth/pkg/types"
	"github.com/corynth/corynth/pkg/workflow"
)

// HybridStateManager bridges the legacy state store and orchestration state manager
type HybridStateManager struct {
	orchestrationSM *state.StateManager
	legacyStateStore workflow.StateStore
}

// NewHybridStateManager creates a new hybrid state manager
func NewHybridStateManager(orchestrationSM *state.StateManager, legacyStateStore workflow.StateStore) *HybridStateManager {
	return &HybridStateManager{
		orchestrationSM:  orchestrationSM,
		legacyStateStore: legacyStateStore,
	}
}

// SaveWorkflowOutput saves workflow outputs for use by other workflows
func (h *HybridStateManager) SaveWorkflowOutput(workflowName string, outputs map[string]interface{}) error {
	// Save using orchestration state manager
	return h.orchestrationSM.SaveWorkflowOutput(workflowName, outputs)
}

// LoadWorkflowOutput loads outputs from a previous workflow execution
// Only uses orchestration state for now to avoid legacy state parsing issues
func (h *HybridStateManager) LoadWorkflowOutput(workflowName string) (*types.WorkflowOutput, error) {
	// Only use orchestration state for now
	return h.orchestrationSM.LoadWorkflowOutput(workflowName)
}

// FindStatesByWorkflow finds all execution states for a specific workflow
func (h *HybridStateManager) FindStatesByWorkflow(workflowName string) ([]workflow.ExecutionState, error) {
	return h.orchestrationSM.FindStatesByWorkflow(workflowName)
}

// extractFromLegacyState extracts workflow outputs from legacy state store
func (h *HybridStateManager) extractFromLegacyState(workflowName string) (*types.WorkflowOutput, error) {
	// Get states from legacy store
	allStates, err := h.legacyStateStore.ListExecutions()
	if err != nil {
		return nil, fmt.Errorf("failed to list states from legacy store: %w", err)
	}
	
	// Filter states for the specific workflow
	var states []workflow.ExecutionState
	for _, state := range allStates {
		if state.WorkflowName == workflowName {
			states = append(states, *state)
		}
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("no outputs found for workflow: %s", workflowName)
	}

	// Get the latest successful state
	var latestState *workflow.ExecutionState
	for i := range states {
		state := &states[i]
		if state.Status == workflow.StatusSuccess {
			if latestState == nil || state.StartTime.After(latestState.StartTime) {
				latestState = state
			}
		}
	}

	if latestState == nil {
		return nil, fmt.Errorf("no successful execution found for workflow: %s", workflowName)
	}

	// Extract outputs from the state
	workflowOutputs := make(map[string]interface{})
	
	// If workflow has explicit outputs, use those
	if len(latestState.Outputs) > 0 {
		for k, v := range latestState.Outputs {
			workflowOutputs[k] = v
		}
	} else {
		// Extract from step outputs if no workflow outputs
		for _, step := range latestState.Steps {
			if step.Status == workflow.StatusSuccess && step.Outputs != nil {
				for k, v := range step.Outputs {
					// Only include non-standard outputs
					if k != "stdout" && k != "stderr" && k != "exit_code" {
						workflowOutputs[k] = v
					}
				}
			}
		}
	}

	// Save to orchestration state for future use
	if len(workflowOutputs) > 0 {
		if err := h.SaveWorkflowOutput(workflowName, workflowOutputs); err != nil {
			fmt.Printf("Warning: failed to cache workflow outputs: %v\n", err)
		}
	}

	return &types.WorkflowOutput{
		WorkflowName: workflowName,
		Outputs:      workflowOutputs,
		Timestamp:    latestState.StartTime,
	}, nil
}