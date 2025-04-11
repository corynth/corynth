package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// StateManagerImpl is responsible for managing state
type StateManagerImpl struct{}

// NewStateManager creates a new StateManagerImpl
func NewStateManager() *StateManagerImpl {
	return &StateManagerImpl{}
}

// LoadState loads the state from the specified directory
func (m *StateManagerImpl) LoadState(dir string) (*State, error) {
	statePath := filepath.Join(dir, ".corynth", "state.json")

	// Check if state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		// Return empty state
		return &State{
			Flows: make(map[string]*FlowState),
		}, nil
	}

	// Read state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("error reading state file: %w", err)
	}

	// Parse JSON
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("error parsing state file: %w", err)
	}

	return &state, nil
}

// SaveState saves the state to the specified directory
func (m *StateManagerImpl) SaveState(dir string, state *State) error {
	statePath := filepath.Join(dir, ".corynth", "state.json")

	// Create .corynth directory if it doesn't exist
	corynthDir := filepath.Join(dir, ".corynth")
	if err := os.MkdirAll(corynthDir, 0755); err != nil {
		return fmt.Errorf("error creating .corynth directory: %w", err)
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling state: %w", err)
	}

	// Write state file
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("error writing state file: %w", err)
	}

	return nil
}

// UpdateFlowState updates the state of a flow
func (m *StateManagerImpl) UpdateFlowState(state *State, flow *Flow, status string) {
	// Create flow state if it doesn't exist
	if state.Flows == nil {
		state.Flows = make(map[string]*FlowState)
	}

	flowState, ok := state.Flows[flow.Name]
	if !ok {
		flowState = &FlowState{
			Steps: make(map[string]*StepState),
		}
		state.Flows[flow.Name] = flowState
	}

	// Update flow state
	flowState.Status = status
	flowState.StartTime = time.Now()
	flowState.EndTime = time.Now()
	flowState.Duration = flowState.EndTime.Sub(flowState.StartTime)

	// Update last apply time
	state.LastApply = time.Now()
}

// UpdateStepState updates the state of a step
func (m *StateManagerImpl) UpdateStepState(state *State, flow *Flow, step *Step) {
	// Create flow state if it doesn't exist
	if state.Flows == nil {
		state.Flows = make(map[string]*FlowState)
	}

	flowState, ok := state.Flows[flow.Name]
	if !ok {
		flowState = &FlowState{
			Steps: make(map[string]*StepState),
		}
		state.Flows[flow.Name] = flowState
	}

	// Create step state if it doesn't exist
	if flowState.Steps == nil {
		flowState.Steps = make(map[string]*StepState)
	}

	// Update step state
	stepState := &StepState{
		Status:    step.Result.Status,
		StartTime: step.Result.StartTime,
		EndTime:   step.Result.EndTime,
		Duration:  step.Result.Duration,
		Output:    step.Result.Output,
		Error:     step.Result.Error,
	}
	flowState.Steps[step.Name] = stepState
}