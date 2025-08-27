package workflow

import (
)

// WorkflowFile represents the root HCL file structure
type WorkflowFile struct {
	Workflows []WorkflowBlock `hcl:"workflow,block"`
	Variables []VariableBlock `hcl:"variable,block"`
	Locals    []LocalsBlock   `hcl:"locals,block"`
	Steps     []StepBlock     `hcl:"step,block"`
	Outputs   []OutputBlock   `hcl:"output,block"`
	Templates []TemplateBlock `hcl:"template,block"`
	Imports   []ImportBlock   `hcl:"import,block"`
}

// WorkflowBlock represents a workflow definition block
type WorkflowBlock struct {
	Name        string   `hcl:"name,label"`
	Description string   `hcl:"description,optional"`
	Version     string   `hcl:"version,optional"`
	Extends     []string `hcl:"extends,optional"`
	Secrets     []string `hcl:"secrets,optional"`
	Metadata    map[string]string `hcl:"metadata,optional"`
}

// VariableBlock represents a variable definition
type VariableBlock struct {
	Name        string           `hcl:"name,label"`
	Type        string           `hcl:"type,optional"`
	Default     string           `hcl:"default,optional"`
	Description string           `hcl:"description,optional"`
	Required    bool             `hcl:"required,optional"`
	Sensitive   bool             `hcl:"sensitive,optional"`
	Validation  *ValidationBlock `hcl:"validation,block"`
}

// ValidationBlock represents variable validation rules
type ValidationBlock struct {
	Condition    string `hcl:"condition"`
	ErrorMessage string `hcl:"error_message"`
}

// LocalsBlock represents local value definitions
type LocalsBlock struct {
	Values map[string]string `hcl:",remain"`
}

// StepBlock represents a step definition
type StepBlock struct {
	Name        string            `hcl:"name,label"`
	Plugin      string            `hcl:"plugin,optional"`
	Action      string            `hcl:"action,optional"`
	Template    string            `hcl:"template,optional"`
	Params      map[string]string `hcl:"params,optional"`
	With        map[string]string `hcl:"with,optional"`
	Condition   string            `hcl:"condition,optional"`
	Timeout     string            `hcl:"timeout,optional"`
	DependsOn   []string          `hcl:"depends_on,optional"`
	SubWorkflow string            `hcl:"subworkflow,optional"`
	Loop        *LoopBlock        `hcl:"loop,block"`
	Retry       *RetryBlock       `hcl:"retry,block"`
}

// LoopBlock defines loop configuration
type LoopBlock struct {
	Over       string `hcl:"over"`
	Variable   string      `hcl:"variable,optional"`
	Parallel   bool        `hcl:"parallel,optional"`
	MaxWorkers int         `hcl:"max_workers,optional"`
}

// RetryBlock defines retry configuration
type RetryBlock struct {
	MaxAttempts int    `hcl:"max_attempts"`
	Delay       string `hcl:"delay,optional"`
	Backoff     string `hcl:"backoff,optional"`
}

// OutputBlock represents an output definition
type OutputBlock struct {
	Name        string      `hcl:"name,label"`
	Value       string `hcl:"value"`
	Description string      `hcl:"description,optional"`
	Sensitive   bool        `hcl:"sensitive,optional"`
}

// TemplateBlock represents a step template
type TemplateBlock struct {
	Name     string            `hcl:"name,label"`
	Plugin   string            `hcl:"plugin"`
	Action   string            `hcl:"action"`
	Defaults map[string]string `hcl:"defaults,optional"`
}

// ImportBlock represents an import statement
type ImportBlock struct {
	Path      string                 `hcl:"path"`
	As        string                 `hcl:"as,optional"`
	Variables map[string]string `hcl:"variables,optional"`
}

// Helper function to convert HCL blocks to internal workflow structure
func (wf *WorkflowFile) ToWorkflow() *Workflow {
	if len(wf.Workflows) == 0 {
		// Create a default workflow if none specified
		workflow := &Workflow{
			Name: "default",
		}
		return wf.populateWorkflow(workflow)
	}

	// Use the first workflow block as the main workflow
	workflowBlock := wf.Workflows[0]
	workflow := &Workflow{
		Name:        workflowBlock.Name,
		Description: workflowBlock.Description,
		Version:     workflowBlock.Version,
		Extends:     workflowBlock.Extends,
		Secrets:     workflowBlock.Secrets,
		Metadata:    workflowBlock.Metadata,
	}

	return wf.populateWorkflow(workflow)
}

func (wf *WorkflowFile) populateWorkflow(workflow *Workflow) *Workflow {
	// Convert variables
	workflow.Variables = make(map[string]Variable)
	for _, vb := range wf.Variables {
		variable := Variable{
			Type:        vb.Type,
			Default:     vb.Default, // Now string type for HCL compatibility
			Description: vb.Description,
			Required:    vb.Required,
			Sensitive:   vb.Sensitive,
		}
		
		// Convert validation if present
		if vb.Validation != nil {
			variable.Validation = &Validation{
				Condition:    vb.Validation.Condition,
				ErrorMessage: vb.Validation.ErrorMessage,
			}
		}
		
		workflow.Variables[vb.Name] = variable
	}

	// Convert locals
	workflow.Locals = make(map[string]string)
	for _, lb := range wf.Locals {
		for key, value := range lb.Values {
			workflow.Locals[key] = value
		}
	}

	// Convert steps
	workflow.Steps = make([]Step, len(wf.Steps))
	for i, sb := range wf.Steps {
		step := Step{
			Name:        sb.Name,
			Plugin:      sb.Plugin,
			Action:      sb.Action,
			Template:    sb.Template,
			Params:      sb.Params, // Now string map for HCL compatibility
			With:        sb.With,   // Now string map for HCL compatibility
			Condition:   sb.Condition,
			Timeout:     sb.Timeout,
			DependsOn:   sb.DependsOn,
			SubWorkflow: sb.SubWorkflow,
		}

		if sb.Loop != nil {
			step.Loop = &LoopConfig{
				Over:       sb.Loop.Over, // Now string type for HCL compatibility
				Variable:   sb.Loop.Variable,
				Parallel:   sb.Loop.Parallel,
				MaxWorkers: sb.Loop.MaxWorkers,
			}
		}

		if sb.Retry != nil {
			step.RetryPolicy = &RetryPolicy{
				MaxAttempts: sb.Retry.MaxAttempts,
				Delay:       sb.Retry.Delay, // Keep as string for runtime parsing
				Backoff:     sb.Retry.Backoff,
			}
		}

		workflow.Steps[i] = step
	}

	// Convert templates
	workflow.Templates = make(map[string]StepTemplate)
	for _, tb := range wf.Templates {
		workflow.Templates[tb.Name] = StepTemplate{
			Plugin:   tb.Plugin,
			Action:   tb.Action,
			Defaults: tb.Defaults, // Now string map for HCL compatibility
		}
	}

	// Convert imports
	workflow.Imports = make([]Import, len(wf.Imports))
	for i, ib := range wf.Imports {
		workflow.Imports[i] = Import{
			Path:      ib.Path,
			As:        ib.As,
			Variables: ib.Variables, // Now string map for HCL compatibility
		}
	}

	return workflow
}