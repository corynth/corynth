package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/corynth/corynth/pkg/state"
	"github.com/corynth/corynth/pkg/workflow"
)

func TestWorkflowOrchestrationIntegration(t *testing.T) {
	// Create temporary directory for state and workflows
	tempDir, err := os.MkdirTemp("", "corynth-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create state manager
	stateManager := state.NewStateManager(filepath.Join(tempDir, "state"))
	
	// Create mock engine that simulates workflow execution
	mockEngine := &IntegrationMockEngine{
		workflows:  make(map[string]*workflow.Workflow),
		executions: make(map[string]*workflow.ExecutionState),
	}
	
	// Create dependency workflow
	depWorkflow := &workflow.Workflow{
		Name: "dependency-workflow",
		DependsOnWorkflow: []workflow.WorkflowDependency{},
		TriggerWorkflows:  []workflow.WorkflowTrigger{},
	}
	
	// Create main workflow with dependency
	mainWorkflow := &workflow.Workflow{
		Name: "main-workflow",
		DependsOnWorkflow: []workflow.WorkflowDependency{
			{
				WorkflowFile: "dependency-workflow.hcl",
				Required:     true,
				ImportVars:   []string{"dataset_path", "record_count"},
				Variables: map[string]string{
					"source_type": "integration_test",
				},
			},
		},
		TriggerWorkflows: []workflow.WorkflowTrigger{
			{
				WorkflowFile: "cleanup-workflow.hcl",
				OnSuccess:    true,
				ExportVars:   []string{"temp_files"},
			},
		},
	}
	
	// Create cleanup workflow
	cleanupWorkflow := &workflow.Workflow{
		Name: "cleanup-workflow",
		DependsOnWorkflow: []workflow.WorkflowDependency{},
		TriggerWorkflows:  []workflow.WorkflowTrigger{},
	}
	
	// Register workflows with mock engine
	mockEngine.workflows["dependency-workflow.hcl"] = depWorkflow
	mockEngine.workflows["main-workflow.hcl"] = mainWorkflow
	mockEngine.workflows["cleanup-workflow.hcl"] = cleanupWorkflow
	
	// Pre-populate dependency outputs (simulate previous execution)
	err = stateManager.SaveWorkflowOutput("dependency-workflow", map[string]interface{}{
		"dataset_path": "/tmp/test-dataset.json",
		"record_count": "1000",
	})
	if err != nil {
		t.Fatalf("Failed to save dependency outputs: %v", err)
	}
	
	// Create orchestrator
	orchestrator := workflow.NewOrchestrator(mockEngine, stateManager)
	
	// Execute main workflow (should trigger dependency and cleanup)
	variables := map[string]interface{}{
		"test_var": "integration_test_value",
	}
	
	state, err := orchestrator.ExecuteWorkflowChain("main-workflow.hcl", variables)
	if err != nil {
		t.Fatalf("Failed to execute workflow chain: %v", err)
	}
	
	// Verify main workflow executed successfully
	if state.Status != workflow.StatusSuccess {
		t.Errorf("Expected main workflow to succeed, got status: %s", state.Status)
	}
	
	if state.WorkflowName != "main-workflow" {
		t.Errorf("Expected workflow name 'main-workflow', got '%s'", state.WorkflowName)
	}
	
	// Verify dependency was executed
	depExecution, exists := mockEngine.executions["dependency-workflow"]
	if !exists {
		t.Error("Expected dependency workflow to be executed")
	} else {
		if depExecution.Variables["source_type"] != "integration_test" {
			t.Errorf("Expected dependency to receive source_type='integration_test', got %v", 
				depExecution.Variables["source_type"])
		}
	}
	
	// Verify main workflow received imported variables
	mainExecution, exists := mockEngine.executions["main-workflow"]
	if !exists {
		t.Error("Expected main workflow to be executed")
	} else {
		if mainExecution.Variables["dataset_path"] == nil {
			t.Error("Expected main workflow to import dataset_path variable")
		}
		if mainExecution.Variables["record_count"] == nil {
			t.Error("Expected main workflow to import record_count variable")
		}
	}
	
	// Verify cleanup workflow was triggered
	cleanupExecution, exists := mockEngine.executions["cleanup-workflow"]
	if !exists {
		t.Error("Expected cleanup workflow to be triggered")
	} else {
		if cleanupExecution.Variables["temp_files"] == nil {
			t.Error("Expected cleanup workflow to receive temp_files variable")
		}
	}
	
	// Verify state persistence
	savedOutput, err := stateManager.LoadWorkflowOutput("main-workflow")
	if err != nil {
		t.Errorf("Failed to load saved workflow output: %v", err)
	} else {
		if savedOutput.WorkflowName != "main-workflow" {
			t.Errorf("Expected saved output for 'main-workflow', got '%s'", savedOutput.WorkflowName)
		}
	}
}

func TestWorkflowOrchestrationErrorHandling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "corynth-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	stateManager := state.NewStateManager(filepath.Join(tempDir, "state"))
	
	mockEngine := &IntegrationMockEngine{
		workflows:  make(map[string]*workflow.Workflow),
		executions: make(map[string]*workflow.ExecutionState),
		simulateFailure: map[string]bool{
			"failing-dependency": true, // This workflow will fail
		},
	}
	
	// Create workflows
	failingDep := &workflow.Workflow{
		Name: "failing-dependency",
		DependsOnWorkflow: []workflow.WorkflowDependency{},
		TriggerWorkflows:  []workflow.WorkflowTrigger{},
	}
	
	mainWorkflow := &workflow.Workflow{
		Name: "main-with-failing-dep",
		DependsOnWorkflow: []workflow.WorkflowDependency{
			{
				WorkflowFile: "failing-dependency.hcl",
				Required:     true,
				ImportVars:   []string{"result"},
			},
		},
		TriggerWorkflows: []workflow.WorkflowTrigger{},
	}
	
	mockEngine.workflows["failing-dependency.hcl"] = failingDep
	mockEngine.workflows["main-with-failing-dep.hcl"] = mainWorkflow
	
	orchestrator := workflow.NewOrchestrator(mockEngine, stateManager)
	
	// Execute workflow with failing dependency
	variables := map[string]interface{}{
		"test_var": "error_test",
	}
	
	_, err = orchestrator.ExecuteWorkflowChain("main-with-failing-dep.hcl", variables)
	if err == nil {
		t.Error("Expected error when required dependency fails, got nil")
	}
	
	// Verify the error message mentions the dependency failure
	if err != nil && err.Error() == "" {
		t.Error("Expected descriptive error message")
	}
}

// IntegrationMockEngine implements workflow.WorkflowEngine for integration testing
type IntegrationMockEngine struct {
	workflows       map[string]*workflow.Workflow
	executions      map[string]*workflow.ExecutionState
	simulateFailure map[string]bool
}

func (m *IntegrationMockEngine) LoadWorkflowFile(path string) (*workflow.Workflow, error) {
	if wf, exists := m.workflows[path]; exists {
		return wf, nil
	}
	return nil, &WorkflowNotFoundError{Path: path}
}

func (m *IntegrationMockEngine) Execute(ctx context.Context, workflowName string, variables map[string]interface{}, mode workflow.ExecutionMode) (*workflow.ExecutionState, error) {
	// Check if we should simulate failure for this workflow
	if m.simulateFailure[workflowName] {
		state := &workflow.ExecutionState{
			ID:           "failed-execution-" + workflowName,
			WorkflowName: workflowName,
			Status:       workflow.StatusFailure,
			StartTime:    time.Now(),
			Variables:    variables,
			Outputs:      map[string]interface{}{},
			ExecutionMode: mode,
		}
		m.executions[workflowName] = state
		return state, &WorkflowExecutionError{WorkflowName: workflowName, Message: "Simulated failure"}
	}
	
	// Create successful execution state
	outputs := map[string]interface{}{
		"result":     "success from " + workflowName,
		"temp_files": "/tmp/integration-test/" + workflowName,
	}
	
	// Add realistic outputs based on workflow name
	switch workflowName {
	case "dependency-workflow":
		outputs["dataset_path"] = "/tmp/test-dataset.json"
		outputs["record_count"] = "1000"
	case "main-workflow":
		outputs["processing_result"] = "processed " + workflowName
	case "cleanup-workflow":
		outputs["cleanup_status"] = "completed"
	}
	
	state := &workflow.ExecutionState{
		ID:           "integration-execution-" + workflowName,
		WorkflowName: workflowName,
		Status:       workflow.StatusSuccess,
		StartTime:    time.Now(),
		Variables:    variables,
		Outputs:      outputs,
		ExecutionMode: mode,
	}
	
	m.executions[workflowName] = state
	return state, nil
}

// WorkflowNotFoundError represents an error when workflow is not found
type WorkflowNotFoundError struct {
	Path string
}

func (e *WorkflowNotFoundError) Error() string {
	return "workflow not found: " + e.Path
}

// WorkflowExecutionError represents a workflow execution error
type WorkflowExecutionError struct {
	WorkflowName string
	Message      string
}

func (e *WorkflowExecutionError) Error() string {
	return "Workflow '" + e.WorkflowName + "' failed: " + e.Message
}