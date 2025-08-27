package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/corynth/corynth/pkg/workflow"
)

// LocalStateStore implements local file-based state storage
type LocalStateStore struct {
	basePath string
	mu       sync.RWMutex
}

// StateFile represents the structure of the state file
type StateFile struct {
	Version    int                          `json:"version"`
	Executions []workflow.ExecutionState    `json:"executions"`
}

// NewLocalStateStore creates a new local state store
func NewLocalStateStore(basePath string) *LocalStateStore {
	return &LocalStateStore{
		basePath: basePath,
	}
}

// SaveExecution saves execution state to local storage (implements StateStore interface)
func (s *LocalStateStore) SaveExecution(state *workflow.ExecutionState) error {
	return s.Save(state)
}

// LoadExecution loads execution state by ID (implements StateStore interface)
func (s *LocalStateStore) LoadExecution(id string) (*workflow.ExecutionState, error) {
	return s.Load(id)
}

// ListExecutions returns all execution states (implements StateStore interface)
func (s *LocalStateStore) ListExecutions() ([]*workflow.ExecutionState, error) {
	executions, err := s.List()
	if err != nil {
		return nil, err
	}
	
	// Convert to slice of pointers
	result := make([]*workflow.ExecutionState, len(executions))
	for i := range executions {
		result[i] = &executions[i]
	}
	return result, nil
}

// DeleteExecution removes an execution from state (implements StateStore interface)
func (s *LocalStateStore) DeleteExecution(id string) error {
	return s.Delete(id)
}

// Save saves execution state to local storage
func (s *LocalStateStore) Save(state *workflow.ExecutionState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	stateFile := filepath.Join(s.basePath, "corynth.tfstate")
	
	// Load existing state
	var stateData StateFile
	if data, err := os.ReadFile(stateFile); err == nil {
		if err := json.Unmarshal(data, &stateData); err != nil {
			return fmt.Errorf("failed to parse existing state: %w", err)
		}
	} else {
		// Initialize new state file
		stateData = StateFile{
			Version:    1,
			Executions: []workflow.ExecutionState{},
		}
	}

	// Update or add execution
	found := false
	for i, exec := range stateData.Executions {
		if exec.ID == state.ID {
			stateData.Executions[i] = *state
			found = true
			break
		}
	}

	if !found {
		stateData.Executions = append(stateData.Executions, *state)
	}

	// Sort by start time (newest first)
	sort.Slice(stateData.Executions, func(i, j int) bool {
		return stateData.Executions[i].StartTime.After(stateData.Executions[j].StartTime)
	})

	// Keep only recent executions (limit to 100)
	if len(stateData.Executions) > 100 {
		stateData.Executions = stateData.Executions[:100]
	}

	// Save to file
	data, err := json.MarshalIndent(stateData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// Load loads execution state by ID
func (s *LocalStateStore) Load(id string) (*workflow.ExecutionState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stateFile := filepath.Join(s.basePath, "corynth.tfstate")
	
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("execution not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var stateData StateFile
	if err := json.Unmarshal(data, &stateData); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Find execution by ID (support partial ID matching)
	for _, exec := range stateData.Executions {
		if exec.ID == id || (len(id) >= 8 && exec.ID[:len(id)] == id) {
			// Restore error fields from error messages
			exec.RestoreErrorFromMessage()
			for i := range exec.Steps {
				exec.Steps[i].RestoreErrorFromMessage()
			}
			return &exec, nil
		}
	}

	return nil, fmt.Errorf("execution not found: %s", id)
}

// List returns all execution states
func (s *LocalStateStore) List() ([]workflow.ExecutionState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stateFile := filepath.Join(s.basePath, "corynth.tfstate")
	
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []workflow.ExecutionState{}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var stateData StateFile
	if err := json.Unmarshal(data, &stateData); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Restore error fields from error messages for all executions
	for i := range stateData.Executions {
		stateData.Executions[i].RestoreErrorFromMessage()
		for j := range stateData.Executions[i].Steps {
			stateData.Executions[i].Steps[j].RestoreErrorFromMessage()
		}
	}

	return stateData.Executions, nil
}

// Delete removes an execution from state
func (s *LocalStateStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stateFile := filepath.Join(s.basePath, "corynth.tfstate")
	
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var stateData StateFile
	if err := json.Unmarshal(data, &stateData); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	// Find and remove execution
	for i, exec := range stateData.Executions {
		if exec.ID == id {
			stateData.Executions = append(stateData.Executions[:i], stateData.Executions[i+1:]...)
			
			// Save updated state
			newData, err := json.MarshalIndent(stateData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal state: %w", err)
			}

			if err := os.WriteFile(stateFile, newData, 0644); err != nil {
				return fmt.Errorf("failed to write state file: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("execution not found: %s", id)
}

// Cleanup removes old executions
func (s *LocalStateStore) Cleanup(maxAge int, maxCount int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stateFile := filepath.Join(s.basePath, "corynth.tfstate")
	
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var stateData StateFile
	if err := json.Unmarshal(data, &stateData); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	// Sort by start time (newest first)
	sort.Slice(stateData.Executions, func(i, j int) bool {
		return stateData.Executions[i].StartTime.After(stateData.Executions[j].StartTime)
	})

	// Apply count limit
	if maxCount > 0 && len(stateData.Executions) > maxCount {
		stateData.Executions = stateData.Executions[:maxCount]
	}

	// Save cleaned state
	newData, err := json.MarshalIndent(stateData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(stateFile, newData, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}