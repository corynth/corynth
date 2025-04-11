package apply

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Execute runs the apply command
func Execute(dir string, flowName string) {
	fmt.Printf("Applying Corynth flows in %s\n", dir)

	// Check if the directory is a Corynth project
	if !isCorynthProject(dir) {
		fmt.Println("❌ Not a Corynth project. Run 'corynth init' first.")
		os.Exit(1)
	}

	// Check if a plan exists
	if !planExists(dir) {
		fmt.Println("❌ No execution plan found. Run 'corynth plan' first.")
		os.Exit(1)
	}

	// Load the execution plan
	plan, err := loadPlan(dir)
	if err != nil {
		fmt.Printf("❌ Error loading execution plan: %s\n", err)
		os.Exit(1)
	}

	// Execute the flows
	if flowName == "" {
		// Execute all flows in the plan
		for _, flow := range plan.Flows {
			executeFlow(dir, flow)
		}
	} else {
		// Execute only the specified flow
		found := false
		for _, flow := range plan.Flows {
			if flow == flowName {
				executeFlow(dir, flow)
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("❌ Flow '%s' not found in the execution plan.\n", flowName)
			os.Exit(1)
		}
	}

	// Update state
	if err := updateState(dir, plan); err != nil {
		fmt.Printf("❌ Error updating state: %s\n", err)
		os.Exit(1)
	}
}

func isCorynthProject(dir string) bool {
	// Check for .corynth directory
	corynthDir := filepath.Join(dir, ".corynth")
	if _, err := os.Stat(corynthDir); os.IsNotExist(err) {
		return false
	}

	// Check for flows directory
	flowsDir := filepath.Join(dir, "flows")
	if _, err := os.Stat(flowsDir); os.IsNotExist(err) {
		return false
	}

	return true
}

func planExists(dir string) bool {
	planPath := filepath.Join(dir, ".corynth", "plan.json")
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Plan represents an execution plan
type Plan struct {
	Flows  []string `json:"flows"`
	Status string   `json:"status"`
}

func loadPlan(dir string) (*Plan, error) {
	planPath := filepath.Join(dir, ".corynth", "plan.json")
	planData, err := os.ReadFile(planPath)
	if err != nil {
		return nil, err
	}

	var plan Plan
	if err := json.Unmarshal(planData, &plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

// Flow represents a Corynth flow definition
type Flow struct {
	Name        string
	Description string
	Steps       []Step
}

// Step represents a step in a flow
type Step struct {
	Name      string
	Plugin    string
	Action    string
	Params    map[string]interface{}
	DependsOn []Dependency
}

// Dependency represents a step dependency
type Dependency struct {
	Step   string
	Status string
}

func executeFlow(dir string, flowName string) {
	// This is a placeholder for actual flow execution logic
	// In a real implementation, this would load the flow definition,
	// resolve dependencies, and execute each step in the correct order

	fmt.Printf("🚀 Executing flow: %s\n", flowName)

	// Mock execution of steps
	mockSteps := []struct {
		Name     string
		Duration float64
	}{
		{"clone_repo", 2.3},
		{"setup_env", 5.7},
		{"deploy", 10.2},
	}

	for _, step := range mockSteps {
		// Simulate step execution
		time.Sleep(time.Duration(step.Duration * float64(time.Second) / 10)) // Reduced for demo
		fmt.Printf("  ✅ Step: %s (%.1fs)\n", step.Name, step.Duration)
	}

	fmt.Println("✨ Flow completed successfully!")
}

func updateState(dir string, plan *Plan) error {
	// This is a placeholder for actual state update logic
	// In a real implementation, this would update the state with the results of the execution

	statePath := filepath.Join(dir, ".corynth", "state.json")
	stateContent := map[string]interface{}{
		"last_apply": time.Now().Format(time.RFC3339),
		"flows":      plan.Flows,
		"status":     "applied",
	}

	stateData, err := json.MarshalIndent(stateContent, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, stateData, 0644)
}