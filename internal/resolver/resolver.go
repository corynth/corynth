package resolver

import (
	"fmt"
)

// DependencyGraph is responsible for resolving dependencies between steps
type DependencyGraph struct {
	flows         []*Flow
	executionOrder []*Step
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
	Status   string
	Output   string
	Error    string
	Duration float64
}

// NewDependencyGraph creates a new DependencyGraph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{}
}

// BuildGraph builds a dependency graph from flows
func (g *DependencyGraph) BuildGraph(flows []*Flow) error {
	g.flows = flows
	g.executionOrder = []*Step{}

	// Create a map of step names to steps
	stepMap := make(map[string]*Step)
	for _, flow := range flows {
		for _, step := range flow.Steps {
			stepMap[step.Name] = step
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	for _, flow := range flows {
		for _, step := range flow.Steps {
			if !visited[step.Name] {
				if err := g.detectCircularDependencies(step, stepMap, visited, temp); err != nil {
					return err
				}
			}
		}
	}

	// Reset visited map for topological sort
	visited = make(map[string]bool)

	// Perform topological sort
	for _, flow := range flows {
		for _, step := range flow.Steps {
			if !visited[step.Name] {
				g.topologicalSort(step, stepMap, visited)
			}
		}
	}

	return nil
}

// detectCircularDependencies detects circular dependencies in the graph
func (g *DependencyGraph) detectCircularDependencies(
	step *Step,
	stepMap map[string]*Step,
	visited map[string]bool,
	temp map[string]bool,
) error {
	if temp[step.Name] {
		return fmt.Errorf("circular dependency detected: %s", step.Name)
	}

	if visited[step.Name] {
		return nil
	}

	temp[step.Name] = true

	for _, dep := range step.DependsOn {
		dependentStep, ok := stepMap[dep.Step]
		if !ok {
			return fmt.Errorf("dependency step %s not found", dep.Step)
		}

		if err := g.detectCircularDependencies(dependentStep, stepMap, visited, temp); err != nil {
			return err
		}
	}

	temp[step.Name] = false
	visited[step.Name] = true

	return nil
}

// topologicalSort performs a topological sort of the graph
func (g *DependencyGraph) topologicalSort(
	step *Step,
	stepMap map[string]*Step,
	visited map[string]bool,
) {
	if visited[step.Name] {
		return
	}

	visited[step.Name] = true

	for _, dep := range step.DependsOn {
		dependentStep, ok := stepMap[dep.Step]
		if !ok {
			// This should not happen as we've already checked for missing dependencies
			continue
		}

		g.topologicalSort(dependentStep, stepMap, visited)
	}

	g.executionOrder = append(g.executionOrder, step)
}

// GetExecutionOrder returns the execution order of steps
func (g *DependencyGraph) GetExecutionOrder() []*Step {
	return g.executionOrder
}

// ValidateFlowDependencies validates dependencies between flows
func (g *DependencyGraph) ValidateFlowDependencies() error {
	// Create a map of flow names
	flowMap := make(map[string]bool)
	for _, flow := range g.flows {
		flowMap[flow.Name] = true
	}

	// Check flow chains
	for _, flow := range g.flows {
		for _, chain := range flow.Chain {
			// Check if flow exists
			if !flowMap[chain.Flow] {
				return fmt.Errorf("flow %s in chain not found", chain.Flow)
			}

			// Check if on_success flow exists
			if chain.OnSuccess != "" && !flowMap[chain.OnSuccess] {
				return fmt.Errorf("on_success flow %s in chain not found", chain.OnSuccess)
			}

			// Check if on_failure flow exists
			if chain.OnFailure != "" && !flowMap[chain.OnFailure] {
				return fmt.Errorf("on_failure flow %s in chain not found", chain.OnFailure)
			}
		}
	}

	return nil
}