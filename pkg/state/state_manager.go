package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/corynth/corynth/pkg/types"
	"github.com/corynth/corynth/pkg/workflow"
)

// StateBackend defines the interface for state storage backends
type StateBackend interface {
	SaveState(state *workflow.ExecutionState) error
	LoadState(stateID string) (*workflow.ExecutionState, error)
	ListStates() ([]workflow.ExecutionState, error)
	FindStatesByWorkflow(workflowName string) ([]workflow.ExecutionState, error)
	GetLatestState(workflowName string) (*workflow.ExecutionState, error)
	SaveWorkflowOutput(workflowName string, outputs map[string]interface{}) error
	LoadWorkflowOutput(workflowName string) (*types.WorkflowOutput, error)
	CleanupOldStates(maxAge time.Duration) error
}

// NewStateBackend creates a new state backend based on configuration
func NewStateBackend(backend string, backendConfig map[string]string) (StateBackend, error) {
	switch backend {
	case "local", "":
		// Default to local backend
		stateDir := backendConfig["path"]
		if stateDir == "" {
			stateDir = ".corynth/state"
		}
		return NewStateManager(stateDir), nil
	
	case "s3":
		return NewS3StateManager(backendConfig)
	
	default:
		return nil, fmt.Errorf("unsupported state backend: %s", backend)
	}
}

// StateManager manages workflow state persistence and retrieval
type StateManager struct {
	stateDir string
}

// NewStateManager creates a new state manager
func NewStateManager(stateDir string) *StateManager {
	return &StateManager{
		stateDir: stateDir,
	}
}

// SaveState saves workflow execution state to persistent storage
func (sm *StateManager) SaveState(state *workflow.ExecutionState) error {
	if err := sm.ensureStateDir(); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	stateFile := filepath.Join(sm.stateDir, fmt.Sprintf("%s.json", state.ID))
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState loads workflow execution state from persistent storage
func (sm *StateManager) LoadState(stateID string) (*workflow.ExecutionState, error) {
	stateFile := filepath.Join(sm.stateDir, fmt.Sprintf("%s.json", stateID))
	
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("state not found: %s", stateID)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state workflow.ExecutionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// ListStates returns a list of all available workflow states
func (sm *StateManager) ListStates() ([]workflow.ExecutionState, error) {
	if err := sm.ensureStateDir(); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(sm.stateDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob state files: %w", err)
	}

	var states []workflow.ExecutionState
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}

		var state workflow.ExecutionState
		if err := json.Unmarshal(data, &state); err != nil {
			continue // Skip files that can't be parsed
		}

		states = append(states, state)
	}

	return states, nil
}

// FindStatesByWorkflow finds all execution states for a specific workflow
func (sm *StateManager) FindStatesByWorkflow(workflowName string) ([]workflow.ExecutionState, error) {
	states, err := sm.ListStates()
	if err != nil {
		return nil, err
	}

	var matchingStates []workflow.ExecutionState
	for _, state := range states {
		if state.WorkflowName == workflowName {
			matchingStates = append(matchingStates, state)
		}
	}

	return matchingStates, nil
}

// GetLatestState returns the most recent execution state for a workflow
func (sm *StateManager) GetLatestState(workflowName string) (*workflow.ExecutionState, error) {
	states, err := sm.FindStatesByWorkflow(workflowName)
	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("no states found for workflow: %s", workflowName)
	}

	// Find the most recent state
	var latest *workflow.ExecutionState
	for i := range states {
		if latest == nil || states[i].StartTime.After(latest.StartTime) {
			latest = &states[i]
		}
	}

	return latest, nil
}

// SaveWorkflowOutput saves workflow outputs for use by other workflows
func (sm *StateManager) SaveWorkflowOutput(workflowName string, outputs map[string]interface{}) error {
	if err := sm.ensureStateDir(); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	outputFile := filepath.Join(sm.stateDir, fmt.Sprintf("outputs_%s.json", workflowName))
	
	outputData := types.WorkflowOutput{
		WorkflowName: workflowName,
		Outputs:      outputs,
		Timestamp:    time.Now(),
	}

	data, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal outputs: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write outputs file: %w", err)
	}

	return nil
}

// LoadWorkflowOutput loads outputs from a previous workflow execution
func (sm *StateManager) LoadWorkflowOutput(workflowName string) (*types.WorkflowOutput, error) {
	outputFile := filepath.Join(sm.stateDir, fmt.Sprintf("outputs_%s.json", workflowName))
	
	data, err := os.ReadFile(outputFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no outputs found for workflow: %s", workflowName)
		}
		return nil, fmt.Errorf("failed to read outputs file: %w", err)
	}

	var output types.WorkflowOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outputs: %w", err)
	}

	return &output, nil
}

// CleanupOldStates removes state files older than the specified duration
func (sm *StateManager) CleanupOldStates(maxAge time.Duration) error {
	if err := sm.ensureStateDir(); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(sm.stateDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob state files: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				// Log but don't fail the cleanup
				fmt.Printf("Warning: failed to remove old state file %s: %v\n", file, err)
			}
		}
	}

	return nil
}

// ensureStateDir creates the state directory if it doesn't exist
func (sm *StateManager) ensureStateDir() error {
	if err := os.MkdirAll(sm.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", sm.stateDir, err)
	}
	return nil
}

