package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FlowParser is responsible for parsing flow definitions
type FlowParser struct{}

// Flow represents a Corynth flow definition
type Flow struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Steps       []*Step      `yaml:"steps"`
	Chain       []*FlowChain `yaml:"chain"`
}

// FlowChain represents a chain of flows
type FlowChain struct {
	Flow      string `yaml:"flow"`
	OnSuccess string `yaml:"on_success"`
	OnFailure string `yaml:"on_failure"`
}

// Step represents a step in a flow
type Step struct {
	Name      string                 `yaml:"name"`
	Plugin    string                 `yaml:"plugin"`
	Action    string                 `yaml:"action"`
	Params    map[string]interface{} `yaml:"params"`
	DependsOn []*Dependency          `yaml:"depends_on"`
	Result    *Result                `yaml:"-"`
}

// Dependency represents a step dependency
type Dependency struct {
	Step   string `yaml:"step"`
	Status string `yaml:"status"`
}

// Result represents the result of a step execution
type Result struct {
	Status   string
	Output   string
	Error    string
	Duration float64
}

// FlowDefinition is the top-level YAML structure
type FlowDefinition struct {
	Flow Flow `yaml:"flow"`
}

// NewFlowParser creates a new FlowParser
func NewFlowParser() *FlowParser {
	return &FlowParser{}
}

// ParseFlow parses a flow definition from a file
func (p *FlowParser) ParseFlow(path string) (*Flow, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading flow file: %w", err)
	}

	// Parse YAML
	var flowDef FlowDefinition
	if err := yaml.Unmarshal(data, &flowDef); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	return &flowDef.Flow, nil
}

// ParseFlows parses all flow definitions in a directory
func (p *FlowParser) ParseFlows(dir string) ([]*Flow, error) {
	// Get all YAML files in the flows directory
	flowsDir := filepath.Join(dir, "flows")
	files, err := os.ReadDir(flowsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading flows directory: %w", err)
	}

	var flows []*Flow
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if file is a YAML file
		ext := filepath.Ext(file.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		// Parse flow
		flowPath := filepath.Join(flowsDir, file.Name())
		flow, err := p.ParseFlow(flowPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing flow %s: %w", file.Name(), err)
		}

		flows = append(flows, flow)
	}

	return flows, nil
}

// ValidateFlow validates a flow definition
func (p *FlowParser) ValidateFlow(flow *Flow) error {
	// Check if flow has a name
	if flow.Name == "" {
		return fmt.Errorf("flow must have a name")
	}

	// Check if flow has steps
	if len(flow.Steps) == 0 && len(flow.Chain) == 0 {
		return fmt.Errorf("flow must have steps or chain")
	}

	// Validate steps
	stepNames := make(map[string]bool)
	for _, step := range flow.Steps {
		// Check if step has a name
		if step.Name == "" {
			return fmt.Errorf("step must have a name")
		}

		// Check if step name is unique
		if stepNames[step.Name] {
			return fmt.Errorf("duplicate step name: %s", step.Name)
		}
		stepNames[step.Name] = true

		// Check if step has a plugin
		if step.Plugin == "" {
			return fmt.Errorf("step %s must have a plugin", step.Name)
		}

		// Check if step has an action
		if step.Action == "" {
			return fmt.Errorf("step %s must have an action", step.Name)
		}

		// Validate dependencies
		for _, dep := range step.DependsOn {
			// Check if dependency has a step
			if dep.Step == "" {
				return fmt.Errorf("dependency in step %s must have a step", step.Name)
			}

			// Check if dependency has a status
			if dep.Status == "" {
				return fmt.Errorf("dependency in step %s must have a status", step.Name)
			}

			// Check if dependency step exists
			if !stepNames[dep.Step] {
				return fmt.Errorf("dependency step %s in step %s does not exist", dep.Step, step.Name)
			}
		}
	}

	return nil
}