package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/corynth/corynth/pkg/types"
)

// MockStateManager implements WorkflowStateManager for testing
type MockStateManager struct {
	outputs map[string]*types.WorkflowOutput
	states  map[string][]ExecutionState
}

func NewMockStateManager() *MockStateManager {
	return &MockStateManager{
		outputs: make(map[string]*types.WorkflowOutput),
		states:  make(map[string][]ExecutionState),
	}
}

func (m *MockStateManager) SaveWorkflowOutput(workflowName string, outputs map[string]interface{}) error {
	m.outputs[workflowName] = &types.WorkflowOutput{
		WorkflowName: workflowName,
		Outputs:      outputs,
		Timestamp:    time.Now(),
	}
	return nil
}

func (m *MockStateManager) LoadWorkflowOutput(workflowName string) (*types.WorkflowOutput, error) {
	if output, exists := m.outputs[workflowName]; exists {
		return output, nil
	}
	return nil, &WorkflowOutputNotFoundError{WorkflowName: workflowName}
}

func (m *MockStateManager) FindStatesByWorkflow(workflowName string) ([]ExecutionState, error) {
	if states, exists := m.states[workflowName]; exists {
		return states, nil
	}
	return []ExecutionState{}, nil
}

// WorkflowOutputNotFoundError represents an error when workflow output is not found
type WorkflowOutputNotFoundError struct {
	WorkflowName string
}

func (e *WorkflowOutputNotFoundError) Error() string {
	return "no outputs found for workflow: " + e.WorkflowName
}

// MockEngine implements a mock workflow engine for testing
type MockEngine struct {
	workflows map[string]*Workflow
	executions map[string]*ExecutionState
}

func NewMockEngine() *MockEngine {
	return &MockEngine{
		workflows:  make(map[string]*Workflow),
		executions: make(map[string]*ExecutionState),
	}
}

func (m *MockEngine) LoadWorkflowFile(path string) (*Workflow, error) {
	if workflow, exists := m.workflows[path]; exists {
		return workflow, nil
	}
	return nil, &WorkflowNotFoundError{Path: path}
}

func (m *MockEngine) Execute(ctx context.Context, workflowName string, variables map[string]interface{}, mode ExecutionMode) (*ExecutionState, error) {
	// Create outputs based on workflow name for testing
	outputs := map[string]interface{}{
		"result": "success from " + workflowName,
	}
	
	// For specific test workflows, use predefined outputs
	if workflowName == "source-workflow" {
		outputs = map[string]interface{}{
			"shared_value": "imported_data",
			"other_value":  "not_imported",
		}
	}
	
	// Add input variables if they exist
	if testVar, exists := variables["test_var"]; exists {
		outputs["input_var"] = testVar
	}
	
	// Create a successful execution state for testing
	state := &ExecutionState{
		ID:           "test-execution-" + workflowName,
		WorkflowName: workflowName,
		Status:       StatusSuccess,
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

func TestOrchestratorBasicExecution(t *testing.T) {
	mockEngine := NewMockEngine()
	mockStateManager := NewMockStateManager()
	
	// Create a simple workflow
	workflow := &Workflow{
		Name: "test-workflow",
		DependsOnWorkflow: []WorkflowDependency{},
		TriggerWorkflows:  []WorkflowTrigger{},
	}
	
	mockEngine.workflows["test-workflow.hcl"] = workflow
	orchestrator := NewOrchestrator(mockEngine, mockStateManager)
	
	variables := map[string]interface{}{
		"test_var": "test_value",
	}
	
	state, err := orchestrator.ExecuteWorkflowChain("test-workflow.hcl", variables)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if state.Status != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, state.Status)
	}
	
	if state.WorkflowName != "test-workflow" {
		t.Errorf("Expected workflow name 'test-workflow', got '%s'", state.WorkflowName)
	}
}

func TestOrchestratorWorkflowDependencies(t *testing.T) {
	mockEngine := NewMockEngine()
	mockStateManager := NewMockStateManager()
	
	// Create dependency workflow
	depWorkflow := &Workflow{
		Name: "dependency-workflow",
		DependsOnWorkflow: []WorkflowDependency{},
		TriggerWorkflows:  []WorkflowTrigger{},
	}
	
	// Create main workflow with dependency
	mainWorkflow := &Workflow{
		Name: "main-workflow",
		DependsOnWorkflow: []WorkflowDependency{
			{
				WorkflowFile: "dependency-workflow.hcl",
				Required:     true,
				ImportVars:   []string{"result"},
				Variables: map[string]string{
					"dep_var": "dependency_value",
				},
			},
		},
		TriggerWorkflows: []WorkflowTrigger{},
	}
	
	mockEngine.workflows["dependency-workflow.hcl"] = depWorkflow
	mockEngine.workflows["main-workflow.hcl"] = mainWorkflow
	
	// Pre-populate dependency output
	mockStateManager.SaveWorkflowOutput("dependency-workflow", map[string]interface{}{
		"result": "dependency_result",
	})
	
	orchestrator := NewOrchestrator(mockEngine, mockStateManager)
	
	variables := map[string]interface{}{
		"main_var": "main_value",
	}
	
	state, err := orchestrator.ExecuteWorkflowChain("main-workflow.hcl", variables)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if state.Status != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, state.Status)
	}
	
	// Check that dependency was executed
	if _, exists := mockEngine.executions["dependency-workflow"]; !exists {
		t.Error("Expected dependency workflow to be executed")
	}
	
	// Check that main workflow was executed
	if _, exists := mockEngine.executions["main-workflow"]; !exists {
		t.Error("Expected main workflow to be executed")
	}
}

func TestOrchestratorWorkflowTriggers(t *testing.T) {
	mockEngine := NewMockEngine()
	mockStateManager := NewMockStateManager()
	
	// Create trigger workflow
	triggerWorkflow := &Workflow{
		Name: "trigger-workflow",
		DependsOnWorkflow: []WorkflowDependency{},
		TriggerWorkflows:  []WorkflowTrigger{},
	}
	
	// Create main workflow with trigger
	mainWorkflow := &Workflow{
		Name: "main-workflow",
		DependsOnWorkflow: []WorkflowDependency{},
		TriggerWorkflows: []WorkflowTrigger{
			{
				WorkflowFile: "trigger-workflow.hcl",
				OnSuccess:    true,
				ExportVars:   []string{"result"},
				Variables: map[string]string{
					"trigger_var": "trigger_value",
				},
			},
		},
	}
	
	mockEngine.workflows["trigger-workflow.hcl"] = triggerWorkflow
	mockEngine.workflows["main-workflow.hcl"] = mainWorkflow
	
	orchestrator := NewOrchestrator(mockEngine, mockStateManager)
	
	variables := map[string]interface{}{
		"main_var": "main_value",
	}
	
	state, err := orchestrator.ExecuteWorkflowChain("main-workflow.hcl", variables)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if state.Status != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, state.Status)
	}
	
	// Check that main workflow was executed
	if _, exists := mockEngine.executions["main-workflow"]; !exists {
		t.Error("Expected main workflow to be executed")
	}
	
	// Check that trigger workflow was executed
	if _, exists := mockEngine.executions["trigger-workflow"]; !exists {
		t.Error("Expected trigger workflow to be executed")
	}
}

func TestOrchestratorVariableImport(t *testing.T) {
	mockEngine := NewMockEngine()
	mockStateManager := NewMockStateManager()
	
	// Pre-populate outputs from a dependency
	mockStateManager.SaveWorkflowOutput("source-workflow", map[string]interface{}{
		"shared_value": "imported_data",
		"other_value":  "not_imported",
	})
	
	// Create workflow with dependency that imports specific variables
	workflow := &Workflow{
		Name: "importing-workflow",
		DependsOnWorkflow: []WorkflowDependency{
			{
				WorkflowFile: "source-workflow.hcl",
				Required:     false,
				ImportVars:   []string{"shared_value"},
			},
		},
		TriggerWorkflows: []WorkflowTrigger{},
	}
	
	sourceWorkflow := &Workflow{
		Name: "source-workflow",
		DependsOnWorkflow: []WorkflowDependency{},
		TriggerWorkflows:  []WorkflowTrigger{},
	}
	
	mockEngine.workflows["importing-workflow.hcl"] = workflow
	mockEngine.workflows["source-workflow.hcl"] = sourceWorkflow
	
	orchestrator := NewOrchestrator(mockEngine, mockStateManager)
	
	variables := map[string]interface{}{
		"local_var": "local_value",
	}
	
	state, err := orchestrator.ExecuteWorkflowChain("importing-workflow.hcl", variables)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if state.Status != StatusSuccess {
		t.Errorf("Expected status %s, got %s", StatusSuccess, state.Status)
	}
	
	// Verify that the importing workflow received the imported variable
	importingExecution := mockEngine.executions["importing-workflow"]
	if importingExecution == nil {
		t.Fatal("Expected importing workflow to be executed")
	}
	
	if importingExecution.Variables["shared_value"] != "imported_data" {
		t.Errorf("Expected imported variable 'shared_value' to be 'imported_data', got %v", 
			importingExecution.Variables["shared_value"])
	}
	
	// Verify that the non-imported variable is not present
	if _, exists := importingExecution.Variables["other_value"]; exists {
		t.Error("Expected 'other_value' to not be imported")
	}
	
	// Verify that local variables are preserved
	if importingExecution.Variables["local_var"] != "local_value" {
		t.Errorf("Expected local variable 'local_var' to be preserved as 'local_value', got %v", 
			importingExecution.Variables["local_var"])
	}
}

func TestOrchestratorGetWorkflowHistory(t *testing.T) {
	mockEngine := NewMockEngine()
	mockStateManager := NewMockStateManager()
	
	// Mock some historical states
	historicalStates := []ExecutionState{
		{
			ID:           "exec-1",
			WorkflowName: "test-workflow",
			Status:       StatusSuccess,
			StartTime:    time.Now().Add(-2 * time.Hour),
		},
		{
			ID:           "exec-2",
			WorkflowName: "test-workflow",
			Status:       StatusFailure,
			StartTime:    time.Now().Add(-1 * time.Hour),
		},
	}
	
	mockStateManager.states["test-workflow"] = historicalStates
	
	orchestrator := NewOrchestrator(mockEngine, mockStateManager)
	
	history, err := orchestrator.GetWorkflowHistory("test-workflow")
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(history) != 2 {
		t.Errorf("Expected 2 historical states, got %d", len(history))
	}
	
	if history[0].ID != "exec-1" {
		t.Errorf("Expected first execution ID to be 'exec-1', got '%s'", history[0].ID)
	}
	
	if history[1].Status != StatusFailure {
		t.Errorf("Expected second execution status to be %s, got %s", StatusFailure, history[1].Status)
	}
}

func TestOrchestratorGetWorkflowOutputs(t *testing.T) {
	mockEngine := NewMockEngine()
	mockStateManager := NewMockStateManager()
	
	// Mock some workflow outputs
	expectedOutputs := map[string]interface{}{
		"result":    "test_result",
		"timestamp": "2024-01-01T00:00:00Z",
	}
	
	mockStateManager.SaveWorkflowOutput("test-workflow", expectedOutputs)
	
	orchestrator := NewOrchestrator(mockEngine, mockStateManager)
	
	outputs, err := orchestrator.GetWorkflowOutputs("test-workflow")
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if outputs["result"] != "test_result" {
		t.Errorf("Expected output 'result' to be 'test_result', got %v", outputs["result"])
	}
	
	if outputs["timestamp"] != "2024-01-01T00:00:00Z" {
		t.Errorf("Expected output 'timestamp' to be '2024-01-01T00:00:00Z', got %v", outputs["timestamp"])
	}
}

func TestOrchestratorExtractWorkflowName(t *testing.T) {
	mockEngine := NewMockEngine()
	mockStateManager := NewMockStateManager()
	orchestrator := NewOrchestrator(mockEngine, mockStateManager)
	
	testCases := []struct {
		input    string
		expected string
	}{
		{"workflow.hcl", "workflow"},
		{"/path/to/workflow.hcl", "workflow"},
		{"workflow.yaml", "workflow"},
		{"complex-workflow-name.hcl", "complex-workflow-name"},
		{"workflow", "workflow"},
	}
	
	for _, tc := range testCases {
		result := orchestrator.extractWorkflowName(tc.input)
		if result != tc.expected {
			t.Errorf("extractWorkflowName(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}