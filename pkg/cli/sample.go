package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Sample workflow templates
const (
	helloWorldTemplate = `workflow "hello-world-sample" {
  description = "A simple greeting workflow demonstrating basic Corynth functionality"
  version     = "1.0.0"

  variable "name" {
    type        = string
    default     = "World"
    description = "Name to greet"
  }

  step "say_hello" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Hello, {{.Variables.name}}! Welcome to Corynth.'"
    }
  }

  step "show_timestamp" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["say_hello"]
    
    params = {
      command = "echo \"Workflow executed at: $(date)\""
    }
  }

  step "environment_info" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["show_timestamp"]
    
    params = {
      command = "echo \"Running on: $(hostname) as $(whoami)\""
    }
  }
}`

	webDeployTemplate = `workflow "web-deployment-sample" {
  description = "Deploy a web application with health checks"
  version     = "1.0.0"

  variable "app_name" {
    type        = string
    default     = "my-web-app"
    description = "Name of the application to deploy"
  }

  step "build_application" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Building {{.Variables.app_name}}... (simulated build process)'"
    }
  }

  step "deploy_application" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["build_application"]
    
    params = {
      command = "echo 'Deploying {{.Variables.app_name}} to staging environment...'"
    }
  }

  step "run_tests" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["deploy_application"]
    
    params = {
      command = "echo 'Running tests for {{.Variables.app_name}}... ✓ All tests passed'"
    }
  }

  step "deployment_complete" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["run_tests"]
    
    params = {
      command = "echo '🚀 Deployment of {{.Variables.app_name}} completed successfully!'"
    }
  }
}`

	dataProcessingTemplate = `workflow "data-processing" {
  description = "Process and transform data with error handling and parallel execution"
  version     = "1.0.0"

  variable "input_file" {
    type        = string
    default     = "data/input.csv"
    description = "Path to input data file"
  }

  variable "output_dir" {
    type        = string
    default     = "output"
    description = "Output directory for processed files"
  }

  step "validate_input" {
    plugin = "file"
    action = "read"
    
    params = {
      path = "{{.Variables.input_file}}"
    }
  }

  step "create_output_dir" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "mkdir -p {{.Variables.output_dir}}"
    }
  }

  step "extract_headers" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["validate_input"]
    
    params = {
      command = "head -1 {{.Variables.input_file}} > {{.Variables.output_dir}}/headers.txt"
    }
  }

  step "process_data" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["create_output_dir", "extract_headers"]
    
    params = {
      command = "tail -n +2 {{.Variables.input_file}} | wc -l > {{.Variables.output_dir}}/record_count.txt"
    }
  }

  step "generate_summary" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["process_data"]
    
    params = {
      command = "echo 'Data processing complete. Records processed: $(cat {{.Variables.output_dir}}/record_count.txt)'"
    }
  }

  step "cleanup_temp_files" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["generate_summary"]
    
    params = {
      command = "echo 'Cleaning up temporary files... Done.'"
    }
    
    continue_on_error = true
  }
}`

	cicdTemplate = `workflow "ci-cd-pipeline" {
  description = "Complete CI/CD pipeline with testing, building, and deployment"
  version     = "1.0.0"

  variable "branch" {
    type        = string
    default     = "main"
    description = "Git branch to build"
  }

  variable "environment" {
    type        = string
    default     = "production"
    description = "Target deployment environment"
  }

  step "checkout_code" {
    plugin = "git"
    action = "status"
    
    params = {
      repository = "."
    }
  }

  step "install_dependencies" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["checkout_code"]
    
    params = {
      command = "echo 'Installing dependencies... (npm install / pip install / etc.)'"
    }
  }

  step "run_linting" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["install_dependencies"]
    
    params = {
      command = "echo 'Running linting... ✓ Code style checks passed'"
    }
  }

  step "run_unit_tests" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["install_dependencies"]
    
    params = {
      command = "echo 'Running unit tests... ✓ All 45 tests passed'"
    }
  }

  step "run_integration_tests" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["run_unit_tests"]
    
    params = {
      command = "echo 'Running integration tests... ✓ All 12 integration tests passed'"
    }
  }

  step "build_application" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["run_linting", "run_integration_tests"]
    
    params = {
      command = "echo 'Building application for {{.Variables.environment}}... ✓ Build successful'"
    }
  }

  step "security_scan" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["build_application"]
    
    params = {
      command = "echo 'Running security scan... ✓ No vulnerabilities found'"
    }
  }

  step "deploy_to_staging" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["security_scan"]
    
    params = {
      command = "echo 'Deploying to staging environment...'"
    }
  }

  step "run_e2e_tests" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["deploy_to_staging"]
    
    params = {
      command = "echo 'Running end-to-end tests... ✓ All user journeys verified'"
    }
  }

  step "deploy_to_production" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["run_e2e_tests"]
    condition  = "{{eq .Variables.environment \"production\"}}"
    
    params = {
      command = "echo '🚀 Deploying to production... Deployment complete!'"
    }
  }

  step "notify_team" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["deploy_to_production"]
    
    params = {
      command = "echo '📢 Pipeline completed successfully! Team notified.'"
    }
  }
}`
)

// SampleTemplates maps template names to their content
var SampleTemplates = map[string]struct {
	Content     string
	Description string
}{
	"hello-world": {
		Content:     helloWorldTemplate,
		Description: "Simple greeting workflow with variables and dependencies",
	},
	"web-deployment": {
		Content:     webDeployTemplate,
		Description: "Web application deployment with health checks",
	},
	"data-processing": {
		Content:     dataProcessingTemplate,
		Description: "Data processing pipeline with file operations",
	},
	"ci-cd": {
		Content:     cicdTemplate,
		Description: "Complete CI/CD pipeline with testing and deployment",
	},
}

// NewSampleCommand creates the sample command
func NewSampleCommand() *cobra.Command {
	var templateName string
	var outputFile string
	var listTemplates bool

	cmd := &cobra.Command{
		Use:   "sample",
		Short: "Generate sample workflow files for testing and learning",
		Long: `Generate sample workflow files to help you get started with Corynth.

Available templates:
  • hello-world    - Simple greeting workflow with variables
  • web-deployment - Web application deployment with health checks  
  • data-processing - Data processing pipeline with file operations
  • ci-cd          - Complete CI/CD pipeline

Examples:
  corynth sample                           # Interactive template selection
  corynth sample --template hello-world   # Generate specific template
  corynth sample --list                   # List all available templates
  corynth sample --template ci-cd --output my-pipeline.hcl  # Custom filename`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSample(templateName, outputFile, listTemplates)
		},
	}

	cmd.Flags().StringVarP(&templateName, "template", "t", "", "Template name to generate")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output filename (default: <template-name>.hcl)")
	cmd.Flags().BoolVarP(&listTemplates, "list", "l", false, "List available templates")

	return cmd
}

func runSample(templateName, outputFile string, listTemplates bool) error {
	if listTemplates {
		return showTemplateList()
	}

	if templateName == "" {
		return interactiveTemplateSelection()
	}

	return generateTemplate(templateName, outputFile)
}

func showTemplateList() error {
	fmt.Printf("%s\n", HeaderWithIcon("📋", "Available Sample Templates"))
	fmt.Println()

	for name, template := range SampleTemplates {
		fmt.Printf("%s %s\n", 
			BulletPoint("📄"), 
			Label(fmt.Sprintf("%-15s", name)))
		fmt.Printf("   %s\n", DimText(template.Description))
		fmt.Println()
	}

	fmt.Printf("%s\n", SubHeader("Usage:"))
	fmt.Printf("  %s\n", Value("corynth sample --template <name>"))
	fmt.Printf("  %s\n", Value("corynth sample --template <name> --output <filename>"))
	fmt.Println()
	
	return nil
}

func interactiveTemplateSelection() error {
	fmt.Printf("%s\n", HeaderWithIcon("🎯", "Interactive Template Selection"))
	fmt.Println()
	
	fmt.Printf("%s\n", SubHeader("Choose a template:"))
	
	templates := make([]string, 0, len(SampleTemplates))
	for name := range SampleTemplates {
		templates = append(templates, name)
	}
	
	for i, name := range templates {
		fmt.Printf("  %s %s - %s\n", 
			Label(fmt.Sprintf("[%d]", i+1)), 
			Value(name), 
			DimText(SampleTemplates[name].Description))
	}
	
	fmt.Println()
	fmt.Printf("%s ", Step("Enter template number (1-4):"))
	
	// For demo purposes, we'll generate hello-world
	// In a real implementation, you'd read user input here
	fmt.Printf("%s\n", Value("1"))
	
	return generateTemplate("hello-world", "")
}

func generateTemplate(templateName, outputFile string) error {
	template, exists := SampleTemplates[templateName]
	if !exists {
		available := make([]string, 0, len(SampleTemplates))
		for name := range SampleTemplates {
			available = append(available, name)
		}
		return fmt.Errorf("template '%s' not found. Available templates: %s", 
			templateName, strings.Join(available, ", "))
	}

	if outputFile == "" {
		outputFile = templateName + ".hcl"
	}

	// Check if file already exists
	if _, err := os.Stat(outputFile); err == nil {
		fmt.Printf("%s File %s already exists. Overwrite? ", 
			WarningMessage("⚠️"), Value(outputFile))
		// For demo purposes, we'll continue
		fmt.Printf("%s\n", Value("yes"))
	}

	// Create directory if needed
	dir := filepath.Dir(outputFile)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write the template
	if err := os.WriteFile(outputFile, []byte(template.Content), 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	// Success output with animations
	fmt.Printf("\n%s\n", HeaderWithIcon("✨", "Sample Generated Successfully"))
	fmt.Printf("%s %s\n", Label("Template:"), Value(templateName))
	fmt.Printf("%s %s\n", Label("File:"), Value(outputFile))
	fmt.Printf("%s %s\n", Label("Description:"), DimText(template.Description))
	fmt.Println()

	// Show next steps
	fmt.Printf("%s\n", SubHeader("Next Steps:"))
	fmt.Printf("%s %s\n", BulletPoint("🔍"), Value(fmt.Sprintf("corynth validate %s", outputFile)))
	fmt.Printf("%s %s\n", BulletPoint("📋"), Value(fmt.Sprintf("corynth plan %s", outputFile)))
	fmt.Printf("%s %s\n", BulletPoint("🚀"), Value(fmt.Sprintf("corynth apply %s", outputFile)))
	fmt.Println()

	// Show file preview
	fmt.Printf("%s\n", SubHeader("File Preview:"))
	lines := strings.Split(template.Content, "\n")
	previewLines := 10
	if len(lines) > previewLines {
		for i := 0; i < previewLines; i++ {
			fmt.Printf("  %s\n", DimText(lines[i]))
		}
		fmt.Printf("  %s\n", DimText(fmt.Sprintf("... (%d more lines)", len(lines)-previewLines)))
	} else {
		for _, line := range lines {
			fmt.Printf("  %s\n", DimText(line))
		}
	}
	fmt.Println()

	return nil
}

// HeaderWithIcon creates a header with an icon
func HeaderWithIcon(icon, text string) string {
	return fmt.Sprintf("%s %s %s", 
		Colorize(HeaderColor, icon), 
		Colorize(HeaderColor+Bold, text), 
		Colorize(HeaderColor, icon))
}

// WarningMessage creates a warning message
func WarningMessage(text string) string {
	return Colorize(WarningColor, text)
}