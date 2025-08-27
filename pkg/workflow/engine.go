package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/corynth/corynth/pkg/plugin"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// Engine executes workflows
type Engine struct {
	pluginManager *plugin.Manager
	workflows     map[string]*Workflow
	mu            sync.RWMutex
	notifier      Notifier
	stateStore    StateStore
	hclParser     *hclparse.Parser
}

// Notifier interface for sending notifications
type Notifier interface {
	NotifyStart(workflow string, execution *ExecutionState)
	NotifyComplete(workflow string, execution *ExecutionState)
	NotifyFailure(workflow string, execution *ExecutionState, err error)
}

// StateStore interface for persisting execution state
type StateStore interface {
	SaveExecution(execution *ExecutionState) error
	LoadExecution(id string) (*ExecutionState, error)
	ListExecutions() ([]*ExecutionState, error)
	DeleteExecution(id string) error
}

// NewEngine creates a new workflow engine
func NewEngine(pluginManager *plugin.Manager) *Engine {
	return &Engine{
		pluginManager: pluginManager,
		workflows:     make(map[string]*Workflow),
		hclParser:     hclparse.NewParser(),
	}
}

// SetNotifier sets the notifier for the engine
func (e *Engine) SetNotifier(notifier Notifier) {
	e.notifier = notifier
}

// SetStateStore sets the state store for the engine
func (e *Engine) SetStateStore(store StateStore) {
	e.stateStore = store
}

// Execute executes a workflow
func (e *Engine) Execute(ctx context.Context, workflowName string, variables map[string]interface{}, mode ExecutionMode) (*ExecutionState, error) {
	workflow, exists := e.workflows[workflowName]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowName)
	}

	// Merge default variable values with provided variables
	mergedVariables := e.mergeVariables(workflow, variables)
	
	execution := &ExecutionState{
		ID:            uuid.New().String(),
		WorkflowName:  workflow.Name,
		Status:        StatusRunning,
		StartTime:     time.Now(),
		Variables:     mergedVariables,
		Steps:         make([]StepState, 0),
		Outputs:       make(map[string]interface{}),
		ExecutionMode: mode,
	}

	if e.stateStore != nil {
		if err := e.stateStore.SaveExecution(execution); err != nil {
			return nil, fmt.Errorf("failed to save execution state: %w", err)
		}
	}

	if e.notifier != nil {
		e.notifier.NotifyStart(workflow.Name, execution)
	}

	execErr := e.executeSteps(ctx, workflow, execution)
	
	endTime := time.Now()
	execution.EndTime = &endTime

	if execErr != nil {
		execution.Status = StatusFailure
		execution.SetError(execErr)
		
		if e.notifier != nil {
			e.notifier.NotifyFailure(workflow.Name, execution, execErr)
		}
	} else {
		execution.Status = StatusSuccess
		
		if e.notifier != nil {
			e.notifier.NotifyComplete(workflow.Name, execution)
		}
	}

	if e.stateStore != nil {
		if saveErr := e.stateStore.SaveExecution(execution); saveErr != nil {
			fmt.Printf("Warning: failed to save final execution state: %v\n", saveErr)
		}
	}

	return execution, execErr
}

// LoadWorkflow loads a workflow from an HCL file
func (e *Engine) LoadWorkflow(filePath string) (*Workflow, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if workflow, exists := e.workflows[filePath]; exists {
		return workflow, nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	file, diags := e.hclParser.ParseHCL(content, filePath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %s", diags.Error())
	}

	var config WorkflowConfig
	diags = gohcl.DecodeBody(file.Body, nil, &config)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode workflow: %s", diags.Error())
	}

	if len(config.Workflows) == 0 {
		return nil, fmt.Errorf("no workflow blocks found in file")
	}

	// Use the first workflow block
	workflowHCL := config.Workflows[0]
	
	// Convert WorkflowHCL to Workflow
	workflow := Workflow{
		Name:        workflowHCL.Name,
		Description: workflowHCL.Description,
		Version:     workflowHCL.Version,
		Extends:     workflowHCL.Extends,
		Secrets:     workflowHCL.Secrets,
		Steps:       workflowHCL.Steps,
		Parallel:    workflowHCL.Parallel,
		Variables:   make(map[string]Variable),
		Locals:      make(map[string]string),
		Templates:   make(map[string]StepTemplate),
		Metadata:    workflowHCL.Metadata,
		Outputs:     workflowHCL.Outputs,
	}

	// Convert variable blocks to map and extract proper default values
	for _, v := range workflowHCL.Variables {
		// Create a copy of the variable with properly extracted default value
		variable := v
		if v.Default != nil {
			// Extract the actual default value from HCL expressions
			variable.Default = e.extractDefaultValue(v.Default)
		}
		workflow.Variables[v.Name] = variable
	}

	// Set default values
	if workflow.Name == "" {
		workflow.Name = strings.TrimSuffix(filepath.Base(filePath), ".hcl")
	}

	e.workflows[filePath] = &workflow
	// Also store by workflow name for easier lookup
	e.workflows[workflow.Name] = &workflow
	return &workflow, nil
}

// Plan generates an execution plan for a workflow
func (e *Engine) Plan(workflowName string, variables map[string]interface{}) (*Plan, error) {
	workflow, exists := e.workflows[workflowName]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowName)
	}

	plan := &Plan{
		ID:           uuid.New().String(),
		WorkflowName: workflow.Name,
		CreatedAt:    time.Now(),
		Variables:    variables,
		Steps:        make([]PlannedStep, 0, len(workflow.Steps)),
	}

	for _, step := range workflow.Steps {
		plannedStep := PlannedStep{
			Name:         step.Name,
			Plugin:       step.Plugin,
			Action:       step.Action,
			Dependencies: step.DependsOn,
			Estimated:    time.Second * 30, // Default estimation
		}
		plan.Steps = append(plan.Steps, plannedStep)
	}

	return plan, nil
}

// LoadWorkflowFile loads a workflow from a file path (alias for LoadWorkflow)
func (e *Engine) LoadWorkflowFile(filePath string) (*Workflow, error) {
	return e.LoadWorkflow(filePath)
}

func (e *Engine) executeSteps(ctx context.Context, workflow *Workflow, execution *ExecutionState) error {
	// Create execution context with variables
	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(convertMapToCtyValue(execution.Variables)),
		},
		Functions: map[string]function.Function{
			"timestamp": function.New(&function.Spec{
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					return cty.StringVal(time.Now().Format(time.RFC3339)), nil
				},
			}),
		},
	}

	// Add some common stdlib functions
	evalCtx.Functions["abs"] = stdlib.AbsoluteFunc
	evalCtx.Functions["ceil"] = stdlib.CeilFunc
	evalCtx.Functions["floor"] = stdlib.FloorFunc
	evalCtx.Functions["max"] = stdlib.MaxFunc
	evalCtx.Functions["min"] = stdlib.MinFunc
	evalCtx.Functions["length"] = stdlib.LengthFunc
	evalCtx.Functions["upper"] = stdlib.UpperFunc
	evalCtx.Functions["lower"] = stdlib.LowerFunc
	evalCtx.Functions["split"] = stdlib.SplitFunc
	evalCtx.Functions["join"] = stdlib.JoinFunc

	// Add boolean and comparison functions
	evalCtx.Functions["equal"] = stdlib.EqualFunc
	evalCtx.Functions["notequal"] = stdlib.NotEqualFunc
	evalCtx.Functions["lessthan"] = stdlib.LessThanFunc
	evalCtx.Functions["lessequal"] = stdlib.LessThanOrEqualToFunc
	evalCtx.Functions["greaterthan"] = stdlib.GreaterThanFunc
	evalCtx.Functions["greaterequal"] = stdlib.GreaterThanOrEqualToFunc
	evalCtx.Functions["and"] = stdlib.AndFunc
	evalCtx.Functions["or"] = stdlib.OrFunc
	evalCtx.Functions["not"] = stdlib.NotFunc
	
	// Add conditional function
	evalCtx.Functions["if"] = function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "condition",
				Type:             cty.Bool,
				AllowDynamicType: false,
			},
			{
				Name:             "true_value",
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
			},
			{
				Name:             "false_value", 
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
			},
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			if args[0].True() {
				return args[1], nil
			}
			return args[2], nil
		},
	})
	evalCtx.Functions["signum"] = stdlib.SignumFunc
	evalCtx.Functions["strlen"] = stdlib.StrlenFunc
	evalCtx.Functions["substr"] = stdlib.SubstrFunc
	evalCtx.Functions["upper"] = stdlib.UpperFunc
	evalCtx.Functions["lower"] = stdlib.LowerFunc

	// Execute steps in dependency order
	for _, step := range workflow.Steps {
		if err := e.executeStep(ctx, &step, execution, evalCtx); err != nil {
			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}
	}

	return nil
}

func (e *Engine) executeStep(ctx context.Context, step *Step, execution *ExecutionState, evalCtx *hcl.EvalContext) error {
	// Check condition first
	if step.Condition != "" {
		shouldExecute, err := e.evaluateCondition(step.Condition, execution, evalCtx)
		if err != nil {
			return fmt.Errorf("failed to evaluate condition for step %s: %w", step.Name, err)
		}
		if !shouldExecute {
			// Skip step execution
			stepState := StepState{
				Name:      step.Name,
				Status:    StatusSkipped,
				StartTime: time.Now(),
				EndTime:   &[]time.Time{time.Now()}[0],
				Outputs:   make(map[string]interface{}),
			}
			execution.Steps = append(execution.Steps, stepState)
			return nil
		}
	}

	// Handle loop execution
	if step.Loop != nil {
		return e.executeStepLoop(ctx, step, execution, evalCtx)
	}

	return e.executeSingleStep(ctx, step, execution, evalCtx)
}

func (e *Engine) executeSingleStep(ctx context.Context, step *Step, execution *ExecutionState, evalCtx *hcl.EvalContext) error {
	stepState := StepState{
		Name:      step.Name,
		Status:    StatusRunning,
		StartTime: time.Now(),
		Outputs:   make(map[string]interface{}),
	}
	execution.Steps = append(execution.Steps, stepState)

	// Update state store
	if e.stateStore != nil {
		if err := e.stateStore.SaveExecution(execution); err != nil {
			fmt.Printf("Warning: failed to save step start state: %v\n", err)
		}
	}

	// Execute the plugin
	pluginInstance, err := e.pluginManager.Get(step.Plugin)
	if err != nil {
		return fmt.Errorf("plugin %s not found: %w", step.Plugin, err)
	}
	
	// Convert string map to interface{} map for plugin execution and process templates
	params := make(map[string]interface{})
	// Support both 'with' and 'params' syntax - merge them
	for k, v := range step.With {
		processed, err := e.processParameter(v, execution, evalCtx)
		if err != nil {
			return fmt.Errorf("failed to process parameter %s: %w", k, err)
		}
		params[k] = processed
	}
	for k, v := range step.Params {
		processed, err := e.processParameter(v, execution, evalCtx)
		if err != nil {
			return fmt.Errorf("failed to process parameter %s: %w", k, err)
		}
		params[k] = processed
	}
	
	result, err := pluginInstance.Execute(ctx, step.Action, params)
	
	endTime := time.Now()
	stepState.EndTime = &endTime

	if err != nil {
		stepState.Status = StatusFailure
		stepState.SetError(err)
		// Update the step in the execution
		for i := range execution.Steps {
			if execution.Steps[i].Name == step.Name {
				execution.Steps[i] = stepState
				break
			}
		}
		return err
	}

	stepState.Status = StatusSuccess
	stepState.Outputs = result
	
	// Update the step in the execution
	for i := range execution.Steps {
		if execution.Steps[i].Name == step.Name {
			execution.Steps[i] = stepState
			break
		}
	}

	// Add step outputs to evaluation context for subsequent steps
	if evalCtx.Variables == nil {
		evalCtx.Variables = make(map[string]cty.Value)
	}
	
	stepOutputs := make(map[string]cty.Value)
	for k, v := range result {
		stepOutputs[k] = convertInterfaceToCtyValue(v)
	}
	
	var stepMap map[string]cty.Value
	if evalCtx.Variables["step"] == cty.NilVal {
		stepMap = make(map[string]cty.Value)
	} else {
		stepMap = evalCtx.Variables["step"].AsValueMap()
		// Create a new map to avoid modifying a copy
		newStepMap := make(map[string]cty.Value)
		for k, v := range stepMap {
			newStepMap[k] = v
		}
		stepMap = newStepMap
	}
	
	stepMap[step.Name] = cty.ObjectVal(map[string]cty.Value{
		"outputs": cty.ObjectVal(stepOutputs),
	})
	evalCtx.Variables["step"] = cty.ObjectVal(stepMap)

	return nil
}

// evaluateCondition evaluates a condition expression and returns whether it's true
func (e *Engine) evaluateCondition(condition string, execution *ExecutionState, evalCtx *hcl.EvalContext) (bool, error) {
	// Parse condition as HCL expression
	expr, diags := hclsyntax.ParseExpression([]byte(condition), "condition", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return false, fmt.Errorf("failed to parse condition: %v", diags)
	}

	// Evaluate expression
	result, diags := expr.Value(evalCtx)
	if diags.HasErrors() {
		return false, fmt.Errorf("failed to evaluate condition: %v", diags)
	}

	// Convert result to boolean
	if result.Type() == cty.Bool {
		return result.True(), nil
	}

	// Handle string comparisons and other types
	if result.Type() == cty.String {
		str := result.AsString()
		return str != "" && str != "false" && str != "0", nil
	}

	if result.Type() == cty.Number {
		val, _ := result.AsBigFloat().Float64()
		return val != 0, nil
	}

	// For other types, consider non-null as true
	return !result.IsNull(), nil
}

// executeStepLoop handles loop execution for a step
func (e *Engine) executeStepLoop(ctx context.Context, step *Step, execution *ExecutionState, evalCtx *hcl.EvalContext) error {
	loop := step.Loop

	// Parse the loop items
	items, err := e.parseLoopItems(loop, execution, evalCtx)
	if err != nil {
		return fmt.Errorf("failed to parse loop items for step %s: %w", step.Name, err)
	}

	if len(items) == 0 {
		// No items to iterate over, skip the step
		stepState := StepState{
			Name:      step.Name,
			Status:    StatusSkipped,
			StartTime: time.Now(),
			EndTime:   &[]time.Time{time.Now()}[0],
			Outputs:   make(map[string]interface{}),
		}
		execution.Steps = append(execution.Steps, stepState)
		return nil
	}

	// Create a copy of the step for iteration
	loopResults := make([]map[string]interface{}, 0, len(items))

	for i, item := range items {
		// Create a new evaluation context for this iteration
		iterCtx := &hcl.EvalContext{
			Variables: make(map[string]cty.Value),
			Functions: evalCtx.Functions,
		}

		// Copy existing variables
		for k, v := range evalCtx.Variables {
			iterCtx.Variables[k] = v
		}

		// Set loop variables
		iterCtx.Variables[loop.Variable] = convertInterfaceToCtyValue(item)
		iterCtx.Variables["loop"] = cty.ObjectVal(map[string]cty.Value{
			"index": cty.NumberIntVal(int64(i)),
			"count": cty.NumberIntVal(int64(len(items))),
			"first": cty.BoolVal(i == 0),
			"last":  cty.BoolVal(i == len(items)-1),
		})

		// Create a modified step name for this iteration
		iterStep := *step
		iterStep.Name = fmt.Sprintf("%s[%d]", step.Name, i)
		iterStep.Loop = nil // Remove loop config for single execution

		// Execute the step for this iteration
		if err := e.executeSingleStep(ctx, &iterStep, execution, iterCtx); err != nil {
			return fmt.Errorf("loop iteration %d failed for step %s: %w", i, step.Name, err)
		}

		// Collect the results
		for j := len(execution.Steps) - 1; j >= 0; j-- {
			if execution.Steps[j].Name == iterStep.Name {
				loopResults = append(loopResults, execution.Steps[j].Outputs)
				break
			}
		}
	}

	// Create a summary step for the loop
	stepState := StepState{
		Name:      step.Name,
		Status:    StatusSuccess,
		StartTime: time.Now(),
		EndTime:   &[]time.Time{time.Now()}[0],
		Outputs: map[string]interface{}{
			"results": loopResults,
			"count":   len(items),
		},
	}
	execution.Steps = append(execution.Steps, stepState)

	return nil
}

// parseLoopItems parses the loop configuration and returns the items to iterate over
func (e *Engine) parseLoopItems(loop *LoopConfig, execution *ExecutionState, evalCtx *hcl.EvalContext) ([]interface{}, error) {
	var items []interface{}

	// Handle 'over' syntax
	if loop.Over != "" {
		expr, diags := hclsyntax.ParseExpression([]byte(loop.Over), "loop_over", hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to parse loop 'over' expression: %v", diags)
		}

		result, diags := expr.Value(evalCtx)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate loop 'over' expression: %v", diags)
		}

		// Handle different types of results
		if result.Type().IsListType() {
			for _, item := range result.AsValueSlice() {
				items = append(items, convertCtyValueToInterface(item))
			}
		} else if result.Type().IsSetType() {
			for _, item := range result.AsValueSet().Values() {
				items = append(items, convertCtyValueToInterface(item))
			}
		} else if result.Type().IsTupleType() {
			// Handle tuples (arrays in HCL)
			for _, item := range result.AsValueSlice() {
				items = append(items, convertCtyValueToInterface(item))
			}
		} else if result.Type().IsObjectType() {
			// Iterate over object keys
			for key := range result.AsValueMap() {
				items = append(items, key)
			}
		} else {
			return nil, fmt.Errorf("loop 'over' must evaluate to a list, set, tuple, or object, got %s", result.Type().FriendlyName())
		}
	}

	// Handle 'items' syntax (alternative)
	if loop.Items != "" {
		expr, diags := hclsyntax.ParseExpression([]byte(loop.Items), "loop_items", hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to parse loop 'items' expression: %v", diags)
		}

		result, diags := expr.Value(evalCtx)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate loop 'items' expression: %v", diags)
		}

		if result.Type().IsListType() {
			for _, item := range result.AsValueSlice() {
				items = append(items, convertCtyValueToInterface(item))
			}
		} else {
			return nil, fmt.Errorf("loop 'items' must evaluate to a list, got %s", result.Type().FriendlyName())
		}
	}

	return items, nil
}

func convertMapToCtyValue(m map[string]interface{}) map[string]cty.Value {
	result := make(map[string]cty.Value)
	for k, v := range m {
		result[k] = convertInterfaceToCtyValue(v)
	}
	return result
}

func convertInterfaceToCtyValue(v interface{}) cty.Value {
	switch val := v.(type) {
	case string:
		return cty.StringVal(val)
	case int:
		return cty.NumberIntVal(int64(val))
	case int64:
		return cty.NumberIntVal(val)
	case float64:
		return cty.NumberFloatVal(val)
	case bool:
		return cty.BoolVal(val)
	case map[string]interface{}:
		objVals := make(map[string]cty.Value)
		for k, v := range val {
			objVals[k] = convertInterfaceToCtyValue(v)
		}
		return cty.ObjectVal(objVals)
	case []interface{}:
		listVals := make([]cty.Value, len(val))
		for i, v := range val {
			listVals[i] = convertInterfaceToCtyValue(v)
		}
		return cty.ListVal(listVals)
	default:
		return cty.StringVal(fmt.Sprintf("%v", v))
	}
}

// processParameter processes a parameter value, handling template substitution
func (e *Engine) processParameter(value interface{}, execution *ExecutionState, evalCtx *hcl.EvalContext) (interface{}, error) {
	// Handle string parameters with template processing
	if strVal, ok := value.(string); ok {
		// Check if string contains template expressions
		if strings.Contains(strVal, "{{") && strings.Contains(strVal, "}}") {
			// Process simple template substitution for variables
			result := strVal
			
			// Process {{.Variables.name}} patterns
			if strings.Contains(result, "{{.Variables.") {
				result = e.processVariableTemplates(result, execution.Variables)
			}
			
			// Process ${var.name} patterns (HCL style)
			if strings.Contains(result, "${var.") {
				result = e.processHCLVariableTemplates(result, execution.Variables)
			}
			
			// Process step output references like ${step_name.output}
			if strings.Contains(result, "${") && execution.Steps != nil {
				result = e.processStepOutputTemplates(result, execution.Steps)
			}
			
			return result, nil
		}
	}
	
	// For non-string parameters, return as-is
	return value, nil
}

// processVariableTemplates processes {{.Variables.name}} style templates
func (e *Engine) processVariableTemplates(template string, variables map[string]interface{}) string {
	result := template
	
	// Simple regex-like replacement for {{.Variables.name}} patterns
	for varName, varValue := range variables {
		placeholder := fmt.Sprintf("{{.Variables.%s}}", varName)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", varValue))
		}
	}
	
	return result
}

// convertCtyValueToInterface converts a cty.Value to interface{}
func convertCtyValueToInterface(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}

	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Number:
		if val.Type().Equals(cty.Number) {
			bf := val.AsBigFloat()
			if bf.IsInt() {
				i64, _ := bf.Int64()
				return i64
			}
			f64, _ := bf.Float64()
			return f64
		}
	case cty.Bool:
		return val.True()
	}

	if val.Type().IsListType() {
		var result []interface{}
		for _, item := range val.AsValueSlice() {
			result = append(result, convertCtyValueToInterface(item))
		}
		return result
	}

	if val.Type().IsSetType() {
		var result []interface{}
		for _, item := range val.AsValueSet().Values() {
			result = append(result, convertCtyValueToInterface(item))
		}
		return result
	}

	if val.Type().IsObjectType() {
		result := make(map[string]interface{})
		for key, item := range val.AsValueMap() {
			result[key] = convertCtyValueToInterface(item)
		}
		return result
	}

	// Fallback: return string representation
	return val.AsString()
}

// mergeVariables merges workflow default variables with provided variables
func (e *Engine) mergeVariables(workflow *Workflow, providedVars map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// First, set default values from workflow variables
	for name, variable := range workflow.Variables {
		if variable.Default != nil {
			result[name] = variable.Default
		}
	}
	
	// Then, override with provided variables
	for name, value := range providedVars {
		result[name] = value
	}
	
	return result
}

// processHCLVariableTemplates processes ${var.name} style templates
func (e *Engine) processHCLVariableTemplates(template string, variables map[string]interface{}) string {
	result := template
	
	// Process ${var.name} patterns
	for varName, varValue := range variables {
		placeholder := fmt.Sprintf("${var.%s}", varName)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", varValue))
		}
	}
	
	return result
}


// processStepOutputTemplates processes ${step_name.output} style templates
func (e *Engine) processStepOutputTemplates(template string, steps []StepState) string {
	result := template
	
	// Process ${step_name.output_key} patterns
	for _, step := range steps {
		for outputKey, outputValue := range step.Outputs {
			placeholder := fmt.Sprintf("${%s.%s}", step.Name, outputKey)
			if strings.Contains(result, placeholder) {
				result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", outputValue))
			}
		}
	}
	
	return result
}

// extractDefaultValue extracts the actual value from an HCL default expression
func (e *Engine) extractDefaultValue(hclValue interface{}) interface{} {
	// For simple cases where HCL already parsed the value correctly
	switch v := hclValue.(type) {
	case string:
		return v
	case int, int64, float64, bool:
		return v
	case cty.Value:
		// Handle HCL cty.Value types
		if v.IsNull() {
			return nil
		}
		switch v.Type() {
		case cty.String:
			return v.AsString()
		case cty.Number:
			f, _ := v.AsBigFloat().Float64()
			return f
		case cty.Bool:
			return v.True()
		default:
			return v.GoString()
		}
	case *hcl.Attribute:
		// Handle HCL attributes - evaluate the expression
		evalCtx := &hcl.EvalContext{
			Variables: make(map[string]cty.Value),
			Functions: make(map[string]function.Function),
		}
		
		val, diags := v.Expr.Value(evalCtx)
		if diags.HasErrors() {
			// If evaluation fails, return nil
			return nil
		}
		
		// Convert cty.Value to Go value
		return e.ctyValueToGo(val)
	default:
		// For complex HCL expressions, try to extract string representation
		// Check if it's an HCL literal expression (common pattern)
		if str := fmt.Sprintf("%v", v); str != "" {
			// If it looks like a simple quoted string, extract it
			if len(str) > 2 && str[0] == '"' && str[len(str)-1] == '"' {
				return str[1 : len(str)-1] // Remove quotes
			}
			return str
		}
		return v
	}
}

// ctyValueToGo converts a cty.Value to a Go interface{} value
func (e *Engine) ctyValueToGo(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}
	
	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Number:
		f, _ := val.AsBigFloat().Float64()
		return f
	case cty.Bool:
		return val.True()
	default:
		// For complex types, return string representation
		return val.GoString()
	}
}
