package initialize

import (
	"fmt"
	"os"
	"path/filepath"
)

// Execute runs the initialize command
func Execute(dir string) {
	fmt.Printf("Initializing Corynth project in %s\n", dir)

	// Create directory structure
	createDirectoryStructure(dir)

	// Generate sample flow definition
	generateSampleFlow(dir)

	// Initialize plugin configuration
	initializePluginConfig(dir)

	// Create .corynth directory for state
	createCorynthDir(dir)

	fmt.Println("✅ Created directory structure")
	fmt.Println("✅ Generated sample flow definition")
	fmt.Println("✅ Plugin configuration initialized")
	fmt.Println("✨ Project initialized! Run 'corynth plan' to validate your flows")
}

func createDirectoryStructure(dir string) {
	// Create main directories
	dirs := []string{
		"flows",
		"plugins",
	}

	for _, d := range dirs {
		path := filepath.Join(dir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %s\n", path, err)
			os.Exit(1)
		}
	}
}

func generateSampleFlow(dir string) {
	sampleFlow := `flow:
  name: "sample_flow"
  description: "Sample deployment flow"
  steps:
    - name: "clone_repo"
      plugin: "git"
      action: "clone"
      params:
        repo: "https://github.com/user/repo.git"
        branch: "main"
    
    - name: "setup_env"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'Setting up environment'"
      depends_on:
        - step: "clone_repo"
          status: "success"
    
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'Deploying application'"
      depends_on:
        - step: "setup_env"
          status: "success"
`

	flowPath := filepath.Join(dir, "flows", "sample_flow.yaml")
	if err := os.WriteFile(flowPath, []byte(sampleFlow), 0644); err != nil {
		fmt.Printf("Error creating sample flow file: %s\n", err)
		os.Exit(1)
	}
}

func initializePluginConfig(dir string) {
	pluginConfig := `plugins:
  - name: "git"
    type: "core"
  - name: "ansible"
    type: "core"
  - name: "shell"
    type: "core"
`

	configPath := filepath.Join(dir, "plugins", "config.yaml")
	if err := os.WriteFile(configPath, []byte(pluginConfig), 0644); err != nil {
		fmt.Printf("Error creating plugin configuration file: %s\n", err)
		os.Exit(1)
	}
}

func createCorynthDir(dir string) {
	corynthDir := filepath.Join(dir, ".corynth")
	if err := os.MkdirAll(corynthDir, 0755); err != nil {
		fmt.Printf("Error creating .corynth directory: %s\n", err)
		os.Exit(1)
	}

	// Create empty state file
	statePath := filepath.Join(corynthDir, "state.json")
	emptyState := "{}"
	if err := os.WriteFile(statePath, []byte(emptyState), 0644); err != nil {
		fmt.Printf("Error creating state file: %s\n", err)
		os.Exit(1)
	}
}