package executor

import (
	"fmt"
	"time"
)

// Executor is responsible for executing steps
type Executor struct {
	pluginManager PluginManager
}

// PluginManager is responsible for managing plugins
type PluginManager interface {
	LoadPlugin(name string) (Plugin, error)
	ExecutePluginAction(plugin Plugin, action string, params map[string]interface{}) (Result, error)
}

// Plugin represents a Corynth plugin
type Plugin interface {
	Name() string
	Execute(action string, params map[string]interface{}) (Result, error)
}

// Flow represents a Corynth flow definition
type Flow struct {
	Name        string
	Description string
	Steps       []*Step
	Chain       []*FlowChain
}

// FlowChain represents a chain of flows
type FlowChain struct {
	Flow      string
	OnSuccess string
	OnFailure string
}

// Step represents a step in a flow
type Step struct {
	Name      string
	Plugin    string
	Action    string
	Params    map[string]interface{}
	DependsOn []*Dependency
	Result    *Result
}

// Dependency represents a step dependency
type Dependency struct {
	Step   string
	Status string
}

// Result represents the result of a step execution
type Result struct {
	Status    string
	Output    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// NewExecutor creates a new Executor
func NewExecutor(pluginManager PluginManager) *Executor {
	return &Executor{
		pluginManager: pluginManager,
	}
}

// ExecuteFlow executes a flow
func (e *Executor) ExecuteFlow(flow *Flow, stepMap map[string]*Step) error {
	fmt.Printf("🚀 Executing flow: %s\n", flow.Name)

	// Execute steps
	for _, step := range flow.Steps {
		if err := e.executeStep(step, stepMap); err != nil {
			return fmt.Errorf("error executing step %s: %w", step.Name, err)
		}
	}

	fmt.Println("✨ Flow completed successfully!")
	return nil
}

// executeStep executes a step
func (e *Executor) executeStep(step *Step, stepMap map[string]*Step) error {
	// Check if dependencies are satisfied
	if !e.areDependenciesSatisfied(step, stepMap) {
		fmt.Printf("  ⏭️ Skipping step: %s (dependencies not satisfied)\n", step.Name)
		return nil
	}

	// Load plugin
	plugin, err := e.pluginManager.LoadPlugin(step.Plugin)
	if err != nil {
		step.Result = &Result{
			Status:    "error",
			Error:     fmt.Sprintf("error loading plugin: %s", err),
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
		fmt.Printf("  ❌ Step: %s (error loading plugin: %s)\n", step.Name, err)
		return fmt.Errorf("error loading plugin %s: %w", step.Plugin, err)
	}

	// Execute plugin action
	startTime := time.Now()
	result, err := e.pluginManager.ExecutePluginAction(plugin, step.Action, step.Params)
	endTime := time.Now()

	if err != nil {
		step.Result = &Result{
			Status:    "error",
			Error:     fmt.Sprintf("error executing plugin action: %s", err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}
		fmt.Printf("  ❌ Step: %s (error executing plugin action: %s)\n", step.Name, err)
		return fmt.Errorf("error executing plugin action %s.%s: %w", step.Plugin, step.Action, err)
	}

	step.Result = &result
	fmt.Printf("  ✅ Step: %s (%.1fs)\n", step.Name, result.Duration.Seconds())
	return nil
}

// areDependenciesSatisfied checks if all dependencies of a step are satisfied
func (e *Executor) areDependenciesSatisfied(step *Step, stepMap map[string]*Step) bool {
	if len(step.DependsOn) == 0 {
		return true
	}

	for _, dep := range step.DependsOn {
		// Find the dependent step
		dependentStep, ok := stepMap[dep.Step]
		if !ok {
			return false
		}

		// Check if the dependent step has been executed
		if dependentStep.Result == nil {
			return false
		}

		// Check if the dependent step has the required status
		if dependentStep.Result.Status != dep.Status {
			return false
		}
	}

	return true
}

// ExecuteFlowChain executes a chain of flows
func (e *Executor) ExecuteFlowChain(flowChain *FlowChain, flows map[string]*Flow, stepMap map[string]*Step) error {
	// Get the flow to execute
	flow, ok := flows[flowChain.Flow]
	if !ok {
		return fmt.Errorf("flow %s not found", flowChain.Flow)
	}

	// Execute the flow
	err := e.ExecuteFlow(flow, stepMap)
	if err != nil {
		// If the flow failed and there's an on_failure flow, execute it
		if flowChain.OnFailure != "" {
			onFailureFlow, ok := flows[flowChain.OnFailure]
			if !ok {
				return fmt.Errorf("on_failure flow %s not found", flowChain.OnFailure)
			}

			fmt.Printf("🔄 Flow %s failed, executing on_failure flow %s\n", flow.Name, onFailureFlow.Name)
			return e.ExecuteFlow(onFailureFlow, stepMap)
		}

		return err
	}

	// If the flow succeeded and there's an on_success flow, execute it
	if flowChain.OnSuccess != "" {
		onSuccessFlow, ok := flows[flowChain.OnSuccess]
		if !ok {
			return fmt.Errorf("on_success flow %s not found", flowChain.OnSuccess)
		}

		fmt.Printf("🔄 Flow %s succeeded, executing on_success flow %s\n", flow.Name, onSuccessFlow.Name)
		return e.ExecuteFlow(onSuccessFlow, stepMap)
	}

	return nil
}