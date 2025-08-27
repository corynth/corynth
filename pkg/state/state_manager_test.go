package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/corynth/corynth/pkg/workflow"
)

func TestStateManagerSaveAndLoadState(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "corynth-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStateManager(tempDir)

	// Create a test execution state
	testState := &workflow.ExecutionState{
		ID:           "test-execution-123",
		WorkflowName: "test-workflow",
		Status:       workflow.StatusSuccess,
		StartTime:    time.Now(),
		Variables: map[string]interface{}{
			"test_var": "test_value",
		},
		Outputs: map[string]interface{}{
			"result": "success",
		},
		ExecutionMode: workflow.ModeApply,
	}

	// Test saving state
	err = sm.SaveState(testState)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Test loading state
	loadedState, err := sm.LoadState("test-execution-123")
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify loaded state matches original
	if loadedState.ID != testState.ID {
		t.Errorf("Expected ID %s, got %s", testState.ID, loadedState.ID)
	}

	if loadedState.WorkflowName != testState.WorkflowName {
		t.Errorf("Expected WorkflowName %s, got %s", testState.WorkflowName, loadedState.WorkflowName)
	}

	if loadedState.Status != testState.Status {
		t.Errorf("Expected Status %s, got %s", testState.Status, loadedState.Status)
	}

	if loadedState.Variables["test_var"] != "test_value" {
		t.Errorf("Expected test_var to be 'test_value', got %v", loadedState.Variables["test_var"])
	}

	if loadedState.Outputs["result"] != "success" {
		t.Errorf("Expected result to be 'success', got %v", loadedState.Outputs["result"])
	}
}

func TestStateManagerListStates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "corynth-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStateManager(tempDir)

	// Create multiple test states
	states := []*workflow.ExecutionState{
		{
			ID:           "exec-1",
			WorkflowName: "workflow-a",
			Status:       workflow.StatusSuccess,
			StartTime:    time.Now(),
		},
		{
			ID:           "exec-2",
			WorkflowName: "workflow-b",
			Status:       workflow.StatusFailure,
			StartTime:    time.Now(),
		},
		{
			ID:           "exec-3",
			WorkflowName: "workflow-a",
			Status:       workflow.StatusSuccess,
			StartTime:    time.Now(),
		},
	}

	// Save all states
	for _, state := range states {
		err := sm.SaveState(state)
		if err != nil {
			t.Fatalf("Failed to save state %s: %v", state.ID, err)
		}
	}

	// Test listing all states
	allStates, err := sm.ListStates()
	if err != nil {
		t.Fatalf("Failed to list states: %v", err)
	}

	if len(allStates) != 3 {
		t.Errorf("Expected 3 states, got %d", len(allStates))
	}

	// Test finding states by workflow
	workflowAStates, err := sm.FindStatesByWorkflow("workflow-a")
	if err != nil {
		t.Fatalf("Failed to find states for workflow-a: %v", err)
	}

	if len(workflowAStates) != 2 {
		t.Errorf("Expected 2 states for workflow-a, got %d", len(workflowAStates))
	}

	workflowBStates, err := sm.FindStatesByWorkflow("workflow-b")
	if err != nil {
		t.Fatalf("Failed to find states for workflow-b: %v", err)
	}

	if len(workflowBStates) != 1 {
		t.Errorf("Expected 1 state for workflow-b, got %d", len(workflowBStates))
	}
}

func TestStateManagerWorkflowOutputs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "corynth-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStateManager(tempDir)

	// Test saving workflow outputs
	testOutputs := map[string]interface{}{
		"result":      "processing_complete",
		"record_count": 1500,
		"output_file":  "/tmp/results.json",
	}

	err = sm.SaveWorkflowOutput("data-processor", testOutputs)
	if err != nil {
		t.Fatalf("Failed to save workflow outputs: %v", err)
	}

	// Test loading workflow outputs
	loadedOutput, err := sm.LoadWorkflowOutput("data-processor")
	if err != nil {
		t.Fatalf("Failed to load workflow outputs: %v", err)
	}

	// Verify loaded outputs match original
	if loadedOutput.WorkflowName != "data-processor" {
		t.Errorf("Expected WorkflowName 'data-processor', got '%s'", loadedOutput.WorkflowName)
	}

	if loadedOutput.Outputs["result"] != "processing_complete" {
		t.Errorf("Expected result 'processing_complete', got %v", loadedOutput.Outputs["result"])
	}

	// JSON unmarshaling converts numbers to float64
	if loadedOutput.Outputs["record_count"] != float64(1500) {
		t.Errorf("Expected record_count 1500, got %v", loadedOutput.Outputs["record_count"])
	}

	if loadedOutput.Outputs["output_file"] != "/tmp/results.json" {
		t.Errorf("Expected output_file '/tmp/results.json', got %v", loadedOutput.Outputs["output_file"])
	}

	// Verify timestamp is recent
	if time.Since(loadedOutput.Timestamp) > time.Minute {
		t.Errorf("Expected recent timestamp, got %v", loadedOutput.Timestamp)
	}
}

func TestStateManagerLoadNonexistentState(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "corynth-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStateManager(tempDir)

	// Try to load a state that doesn't exist
	_, err = sm.LoadState("nonexistent-state")
	if err == nil {
		t.Error("Expected error when loading nonexistent state, got nil")
	}

	// Try to load outputs for a workflow that doesn't exist
	_, err = sm.LoadWorkflowOutput("nonexistent-workflow")
	if err == nil {
		t.Error("Expected error when loading outputs for nonexistent workflow, got nil")
	}
}

func TestStateManagerGetLatestState(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "corynth-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStateManager(tempDir)

	// Create states with different timestamps
	now := time.Now()
	states := []*workflow.ExecutionState{
		{
			ID:           "exec-1",
			WorkflowName: "test-workflow",
			Status:       workflow.StatusSuccess,
			StartTime:    now.Add(-2 * time.Hour),
		},
		{
			ID:           "exec-2",
			WorkflowName: "test-workflow",
			Status:       workflow.StatusFailure,
			StartTime:    now.Add(-1 * time.Hour),
		},
		{
			ID:           "exec-3",
			WorkflowName: "test-workflow",
			Status:       workflow.StatusSuccess,
			StartTime:    now, // Most recent
		},
	}

	// Save all states
	for _, state := range states {
		err := sm.SaveState(state)
		if err != nil {
			t.Fatalf("Failed to save state %s: %v", state.ID, err)
		}
	}

	// Get latest state
	latestState, err := sm.GetLatestState("test-workflow")
	if err != nil {
		t.Fatalf("Failed to get latest state: %v", err)
	}

	// Should be exec-3 (most recent)
	if latestState.ID != "exec-3" {
		t.Errorf("Expected latest state ID to be 'exec-3', got '%s'", latestState.ID)
	}

	if latestState.Status != workflow.StatusSuccess {
		t.Errorf("Expected latest state status to be %s, got %s", workflow.StatusSuccess, latestState.Status)
	}
}

func TestStateManagerCleanupOldStates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "corynth-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewStateManager(tempDir)

	// Create a test state
	testState := &workflow.ExecutionState{
		ID:           "old-state",
		WorkflowName: "test-workflow",
		Status:       workflow.StatusSuccess,
		StartTime:    time.Now(),
	}

	err = sm.SaveState(testState)
	if err != nil {
		t.Fatalf("Failed to save test state: %v", err)
	}

	// Verify the file exists
	stateFile := filepath.Join(tempDir, "old-state.json")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Fatalf("State file should exist but doesn't: %s", stateFile)
	}

	// Wait a moment, then change the file's modification time to simulate an old file
	time.Sleep(10 * time.Millisecond)
	oldTime := time.Now().Add(-25 * time.Hour)
	err = os.Chtimes(stateFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("Failed to change file time: %v", err)
	}

	// Clean up states older than 24 hours
	err = sm.CleanupOldStates(24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to cleanup old states: %v", err)
	}

	// Verify the old file was removed
	if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
		t.Error("Old state file should have been removed but still exists")
	}
}

func TestStateManagerEnsureStateDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "corynth-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a subdirectory that doesn't exist yet
	stateDir := filepath.Join(tempDir, "nested", "state", "directory")
	sm := NewStateManager(stateDir)

	// Create a test state (this should trigger directory creation)
	testState := &workflow.ExecutionState{
		ID:           "test-state",
		WorkflowName: "test-workflow",
		Status:       workflow.StatusSuccess,
		StartTime:    time.Now(),
	}

	err = sm.SaveState(testState)
	if err != nil {
		t.Fatalf("Failed to save state (should create directories): %v", err)
	}

	// Verify the nested directory was created
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		t.Errorf("State directory should have been created but doesn't exist: %s", stateDir)
	}

	// Verify the state file was created
	stateFile := filepath.Join(stateDir, "test-state.json")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Errorf("State file should have been created but doesn't exist: %s", stateFile)
	}
}