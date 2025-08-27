package workflow

import (
	"fmt"
	"time"
)

// WorkflowConfig represents the root HCL configuration with workflow blocks
type WorkflowConfig struct {
	Workflows []WorkflowHCL `hcl:"workflow,block"`
}

// WorkflowHCL represents a workflow block in HCL format
type WorkflowHCL struct {
	Name            string          `hcl:"name,label"`
	Description     string          `hcl:"description,optional"`
	Version         string          `hcl:"version,optional"`
	Extends         []string        `hcl:"extends,optional"`
	Secrets         []string        `hcl:"secrets,optional"`
	Variables       []Variable      `hcl:"variable,block"`
	Locals          []LocalsHCL     `hcl:"locals,block"`
	Steps           []Step          `hcl:"step,block"`
	Parallel        []ParallelGroup `hcl:"parallel,block"`
	OnSuccessBlocks []OnSuccessBlock `hcl:"on_success,block"`
	OnFailureBlocks []OnFailureBlock `hcl:"on_failure,block"`
	Templates       []StepTemplate  `hcl:"template,block"`
	Imports         []Import        `hcl:"import,block"`
	OutputBlocks    []OutputHCL     `hcl:"output,block"`
	Metadata        map[string]string `hcl:"metadata,optional"`
	Outputs         map[string]string `hcl:"outputs,optional"`
	// New fields for workflow dependencies and chaining
	DependsOnWorkflow []WorkflowDependency `hcl:"depends_on_workflow,block"`
	TriggerWorkflows  []WorkflowTrigger    `hcl:"trigger_workflow,block"`
}

// LocalsHCL represents a locals block in HCL format
type LocalsHCL struct {
	Values map[string]interface{} `hcl:",remain"`
}

// OutputHCL represents an output block in HCL format
type OutputHCL struct {
	Name        string `hcl:"name,label"`
	Value       string `hcl:"value"`
	Description string `hcl:"description,optional"`
	Sensitive   bool   `hcl:"sensitive,optional"`
}

// WorkflowDependency represents a dependency on another workflow
type WorkflowDependency struct {
	WorkflowFile string            `hcl:"workflow_file"`
	Variables    map[string]string `hcl:"variables,optional"`
	ImportVars   []string          `hcl:"import_vars,optional"`
	ImportAll    bool              `hcl:"import_all,optional"`
	Required     bool              `hcl:"required,optional"`
}

// WorkflowTrigger represents a workflow to trigger after completion
type WorkflowTrigger struct {
	WorkflowFile string            `hcl:"workflow_file"`
	Variables    map[string]string `hcl:"variables,optional"`
	ExportVars   []string          `hcl:"export_vars,optional"`
	ExportAll    bool              `hcl:"export_all,optional"`
	OnSuccess    bool              `hcl:"on_success,optional"`
	OnFailure    bool              `hcl:"on_failure,optional"`
}

// OnSuccessBlock represents an on_success block in HCL
type OnSuccessBlock struct {
	Steps []Step `hcl:"step,block"`
}

// OnFailureBlock represents an on_failure block in HCL
type OnFailureBlock struct {
	Steps []Step `hcl:"step,block"`
}

// Workflow represents a complete workflow definition
type Workflow struct {
	Name        string                 `hcl:"name,label" yaml:"name"`
	Description string                 `hcl:"description,optional" yaml:"description"`
	Version     string                 `hcl:"version,optional" yaml:"version"`
	Extends     []string              `hcl:"extends,optional" yaml:"extends"`       // Support extending other workflows
	Imports     []Import              `hcl:"import,block" yaml:"imports"`          // Import other workflow definitions
	Triggers    []Trigger             `hcl:"trigger,block" yaml:"triggers"`
	Variables   map[string]Variable   `yaml:"variables"`                           // Internal map representation
	VariableBlocks []Variable          `hcl:"variable,block"`                       // HCL block representation
	Locals      map[string]string     `yaml:"locals"`                              // Local computed values
	Secrets     []string              `hcl:"secrets,optional" yaml:"secrets"`
	Steps       []Step                `hcl:"step,block" yaml:"steps"`
	Parallel    []ParallelGroup       `hcl:"parallel,block" yaml:"parallel"`        // Support parallel execution groups
	OnSuccess   []Step                `hcl:"on_success,block" yaml:"on_success"`
	OnFailure   []Step                `hcl:"on_failure,block" yaml:"on_failure"`
	Templates   map[string]StepTemplate `yaml:"templates"`                         // Internal map representation
	TemplateBlocks []StepTemplate      `hcl:"template,block"`                       // HCL block representation
	Metadata    map[string]string     `hcl:"metadata,optional" yaml:"metadata"`
	Outputs     map[string]string     `hcl:"outputs,optional" yaml:"outputs"`      // Add outputs support
	// Workflow composition and chaining support
	DependsOnWorkflow []WorkflowDependency `hcl:"depends_on_workflow,block" yaml:"depends_on_workflow"`
	TriggerWorkflows  []WorkflowTrigger    `hcl:"trigger_workflow,block" yaml:"trigger_workflows"`
}

// Import allows importing definitions from other files
type Import struct {
	Path      string            `hcl:"path,label" yaml:"path"`
	As        string            `hcl:"as,optional" yaml:"as"`
	Variables map[string]string `hcl:"variables,optional" yaml:"variables"` // Use string for HCL compatibility
}

// ParallelGroup defines steps that can run in parallel
type ParallelGroup struct {
	Name       string `hcl:"name,label" yaml:"name"`
	Steps      []Step `hcl:"step,block" yaml:"steps"`
	MaxWorkers int    `hcl:"max_workers,optional" yaml:"max_workers"`
}

// StepTemplate defines a reusable step template
type StepTemplate struct {
	Name        string            `hcl:"name,label" yaml:"name"`
	Plugin      string            `hcl:"plugin" yaml:"plugin"`
	Action      string            `hcl:"action" yaml:"action"`
	With        map[string]string `hcl:"with,optional" yaml:"with"`     // Use string for HCL compatibility
	Defaults    map[string]string `hcl:"defaults,optional" yaml:"defaults"` // Use string for HCL compatibility
}

// Trigger defines when a workflow should run
type Trigger struct {
	Name     string            `hcl:"name,label" yaml:"name"`
	Type     string            `hcl:"type" yaml:"type"` // manual, schedule, webhook, event
	Schedule string            `hcl:"schedule,optional" yaml:"schedule"`
	Event    string            `hcl:"event,optional" yaml:"event"`
	Config   map[string]string `hcl:"config,optional" yaml:"config"` // Use string for HCL compatibility
}

// Variable represents a workflow variable
type Variable struct {
	Name        string      `hcl:"name,label" yaml:"name"`
	Type        interface{} `hcl:"type" yaml:"type"` // Use interface{} for complex type expressions
	Default     interface{} `hcl:"default,optional" yaml:"default"` // Use interface{} for complex defaults
	Description string      `hcl:"description,optional" yaml:"description"`
	Required    bool        `hcl:"required,optional" yaml:"required"`
	Sensitive   bool        `hcl:"sensitive,optional" yaml:"sensitive"`
	Validation  *Validation `hcl:"validation,block" yaml:"validation,omitempty"`
}

// Validation represents variable validation rules
type Validation struct {
	Condition    string `hcl:"condition" yaml:"condition"`
	ErrorMessage string `hcl:"error_message" yaml:"error_message"`
}

// Step represents a single step in the workflow
type Step struct {
	Name        string            `hcl:"name,label" yaml:"name"`
	Plugin      string            `hcl:"plugin,optional" yaml:"plugin"`
	Action      string            `hcl:"action,optional" yaml:"action"`
	Template    string            `hcl:"template,optional" yaml:"template"`     // Reference to a template
	Loop        *LoopConfig       `hcl:"loop,block" yaml:"loop"`            // Support looping
	With        map[string]string `hcl:"with,optional" yaml:"with"`        // Use string for HCL compatibility
	Params      map[string]string `hcl:"params,optional" yaml:"params"`    // Use string for HCL compatibility
	Condition   string            `hcl:"condition,optional" yaml:"condition"`
	RetryPolicy *RetryPolicy      `hcl:"retry,block" yaml:"retry"`
	Timeout     string            `hcl:"timeout,optional" yaml:"timeout"`
	ContinueOn  *ContinueOn       `hcl:"continue_on,block" yaml:"continue_on"`
	Outputs     map[string]string `hcl:"outputs,optional" yaml:"outputs"`
	DependsOn   []string          `hcl:"depends_on,optional" yaml:"depends_on"`
	SubWorkflow string            `hcl:"subworkflow,optional" yaml:"subworkflow"`  // Execute another workflow as a step
}

// LoopConfig defines loop configuration for a step
type LoopConfig struct {
	Over       string `hcl:"over" yaml:"over"`                // Can be a list or range expression (as string for HCL compatibility)
	Items      string `hcl:"items,optional" yaml:"items"`     // Alternative syntax for iteration (as string for HCL compatibility)
	Variable   string `hcl:"variable" yaml:"variable"`        // Loop variable name
	Parallel   bool   `hcl:"parallel,optional" yaml:"parallel"`
	MaxWorkers int    `hcl:"max_workers,optional" yaml:"max_workers"`
}

// RetryPolicy defines retry behavior for a step
type RetryPolicy struct {
	MaxAttempts int    `hcl:"max_attempts" yaml:"max_attempts"`
	Delay       string `hcl:"delay" yaml:"delay"` // Duration string like "5s", "1m"
	Backoff     string `hcl:"backoff,optional" yaml:"backoff"` // linear, exponential
}

// ContinueOn defines when to continue execution despite failures
type ContinueOn struct {
	Error   bool `hcl:"error,optional" yaml:"error"`
	Failure bool `hcl:"failure,optional" yaml:"failure"`
}

// ExecutionState represents the state of a workflow execution
type ExecutionState struct {
	ID                string                 `json:"id"`
	WorkflowName      string                 `json:"workflow_name"`
	Status            Status                 `json:"status"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           *time.Time             `json:"end_time,omitempty"`
	Steps             []StepState            `json:"steps"`
	Variables         map[string]interface{} `json:"variables"`
	Outputs           map[string]interface{} `json:"outputs"`
	Error             error                  `json:"-"` // Exclude from JSON
	ErrorMessage      string                 `json:"error_message,omitempty"` // JSON-serializable error
	TriggeredBy       string                 `json:"triggered_by"`
	ExecutionMode     ExecutionMode          `json:"execution_mode"`
	// New fields for workflow chaining
	ParentWorkflowID  string                 `json:"parent_workflow_id"` // ID of parent workflow that triggered this one
	ChildWorkflowIDs  []string               `json:"child_workflow_ids"` // IDs of child workflows triggered by this one
	ImportedVariables map[string]interface{} `json:"imported_variables"` // Variables imported from other workflows
	PersistentState   map[string]interface{} `json:"persistent_state"` // State that persists across workflow runs
}

// StepState represents the state of a step execution
type StepState struct {
	Name         string                 `json:"name"`
	Status       Status                 `json:"status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Outputs      map[string]interface{} `json:"outputs"`
	Error        error                  `json:"-"` // Exclude from JSON
	ErrorMessage string                 `json:"error_message,omitempty"` // JSON-serializable error
	Attempts     int                    `json:"attempts"`
}

// SetError sets the error and error message for ExecutionState
func (es *ExecutionState) SetError(err error) {
	es.Error = err
	if err != nil {
		es.ErrorMessage = err.Error()
	} else {
		es.ErrorMessage = ""
	}
}

// SetError sets the error and error message for StepState
func (ss *StepState) SetError(err error) {
	ss.Error = err
	if err != nil {
		ss.ErrorMessage = err.Error()
	} else {
		ss.ErrorMessage = ""
	}
}

// RestoreErrorFromMessage restores the error field from the error message after JSON unmarshaling
func (es *ExecutionState) RestoreErrorFromMessage() {
	if es.ErrorMessage != "" {
		es.Error = fmt.Errorf("%s", es.ErrorMessage)
	}
}

// RestoreErrorFromMessage restores the error field from the error message after JSON unmarshaling
func (ss *StepState) RestoreErrorFromMessage() {
	if ss.ErrorMessage != "" {
		ss.Error = fmt.Errorf("%s", ss.ErrorMessage)
	}
}

// Status represents the execution status
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSuccess   Status = "success"
	StatusFailure   Status = "failure"
	StatusSkipped   Status = "skipped"
	StatusCancelled Status = "cancelled"
)

// ExecutionMode represents how the workflow is being executed
type ExecutionMode string

const (
	ModeInit  ExecutionMode = "init"
	ModePlan  ExecutionMode = "plan"
	ModeApply ExecutionMode = "apply"
)

// Plan represents the execution plan for a workflow
type Plan struct {
	ID           string
	WorkflowName string
	Steps        []PlannedStep
	CreatedAt    time.Time
	Variables    map[string]interface{}
}

// PlannedStep represents a step in the execution plan
type PlannedStep struct {
	Name         string
	Plugin       string
	Action       string
	Dependencies []string
	Estimated    time.Duration
}