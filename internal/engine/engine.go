package engine

import (
	"fmt"
	"time"
)

// Engine represents the Corynth execution engine
type Engine struct {
	flowParser      FlowParser
	pluginManager   PluginManager
	dependencyGraph DependencyGraph
	stateManager    StateManager
}

// FlowParser is responsible for parsing flow definitions
type FlowParser interface {
	ParseFlow(path string) (*Flow, error)
	ParseFlows(dir string) ([]*Flow, error)
}

// PluginManager is responsible for managing plugins
type PluginManager interface {
	LoadPlugin(name string) (Plugin, error)
	ExecutePluginAction(plugin Plugin, action string, params map[string]interface{}) (Result, error)
}

// DependencyGraph is responsible for resolving dependencies between steps
type DependencyGraph interface {
	BuildGraph(flows []*Flow) error
	GetExecutionOrder() []*Step
}

// StateManager is responsible for managing state
type StateManager interface {
	LoadState(dir string) (*State, error)
	SaveState(dir string, state *State) error
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

// Plugin represents a Corynth plugin
type Plugin interface {
	Name() string
	Execute(action string, params map[string]interface{}) (Result, error)
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

// State represents the Corynth state
type State struct {
	LastApply time.Time
	Flows     map[string]*FlowState
}

// FlowState represents the state of a flow
type FlowState struct {
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Steps     map[string]*StepState
}

// StepState represents the state of a step
type StepState struct {
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Output    string
	Error     string
}

// NewEngine creates a new Corynth engine
func NewEngine(
	flowParser FlowParser,
	pluginManager PluginManager,
	dependencyGraph DependencyGraph,
	stateManager StateManager,
) *Engine {
	return &Engine{
		flowParser:      flowParser,
		pluginManager:   pluginManager,
		dependencyGraph: dependencyGraph,
		stateManager:    stateManager,
	}
}

// Plan creates an execution plan for the flows in the specified directory
func (e *Engine) Plan(dir string) ([]*Flow, error) {
	// Parse all flows in the directory
	flows, err := e.flowParser.ParseFlows(dir)
	if err != nil {
		return nil, fmt.Errorf("error parsing flows: %w", err)
	}

	// Build dependency graph
	if err := e.dependencyGraph.BuildGraph(flows); err != nil {
		return nil, fmt.Errorf("error building dependency graph: %w", err)
	}

	return flows, nil
}

// Apply executes the flows in the specified directory
func (e *Engine) Apply(dir string, flowName string) error {
	// Load state
	state, err := e.stateManager.LoadState(dir)
	if err != nil {
		return fmt.Errorf("error loading state: %w", err)
	}

	// Parse all flows in the directory
	flows, err := e.flowParser.ParseFlows(dir)
	if err != nil {
		return fmt.Errorf("error parsing flows: %w", err)
	}

	// Build dependency graph
	if err := e.dependencyGraph.BuildGraph(flows); err != nil {
		return fmt.Errorf("error building dependency graph: %w", err)
	}

	// Execute flows
	if flowName == "" {
		// Execute all flows
		for _, flow := range flows {
			if err := e.executeFlow(flow); err != nil {
				return fmt.Errorf("error executing flow %s: %w", flow.Name, err)
			}
		}
	} else {
		// Execute only the specified flow
		var flowToExecute *Flow
		for _, flow := range flows {
			if flow.Name == flowName {
				flowToExecute = flow
				break
			}
		}

		if flowToExecute == nil {
			return fmt.Errorf("flow %s not found", flowName)
		}

		if err := e.executeFlow(flowToExecute); err != nil {
			return fmt.Errorf("error executing flow %s: %w", flowName, err)
		}
	}

	// Save state
	if err := e.stateManager.SaveState(dir, state); err != nil {
		return fmt.Errorf("error saving state: %w", err)
	}

	return nil
}

// executeFlow executes a single flow
func (e *Engine) executeFlow(flow *Flow) error {
	fmt.Printf("🚀 Executing flow: %s\n", flow.Name)

	// Get execution order from dependency graph
	steps := e.dependencyGraph.GetExecutionOrder()

	// Execute steps
	for _, step := range steps {
		// Check if step belongs to this flow
		if !containsStep(flow.Steps, step) {
			continue
		}

		// Check if dependencies are satisfied
		if !e.areDependenciesSatisfied(step) {
			fmt.Printf("  ⏭️ Skipping step: %s (dependencies not satisfied)\n", step.Name)
			continue
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
	}

	fmt.Println("✨ Flow completed successfully!")
	return nil
}

// areDependenciesSatisfied checks if all dependencies of a step are satisfied
func (e *Engine) areDependenciesSatisfied(step *Step) bool {
	if len(step.DependsOn) == 0 {
		return true
	}

	for _, dep := range step.DependsOn {
		// Find the dependent step
		var dependentStep *Step
		for _, s := range e.dependencyGraph.GetExecutionOrder() {
			if s.Name == dep.Step {
				dependentStep = s
				break
			}
		}

		if dependentStep == nil {
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

// containsStep checks if a step is in a list of steps
func containsStep(steps []*Step, step *Step) bool {
	for _, s := range steps {
		if s.Name == step.Name {
			return true
		}
	}
	return false
}