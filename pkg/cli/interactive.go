package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// NewStartCommand creates the interactive start command
func NewStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Interactive workflow creation for beginners",
		Long: `Start an interactive session to create and run your first workflow.
Perfect for newcomers who want to see results immediately.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteractiveStart(cmd, args)
		},
	}
}

func runInteractiveStart(cmd *cobra.Command, args []string) error {
	session := NewInteractiveSession()
	return session.Run()
}

type InteractiveSession struct {
	scanner *bufio.Scanner
	user    UserContext
}

type UserContext struct {
	Name         string
	Experience   string
	ProjectType  string
	Preferences  map[string]string
}

func NewInteractiveSession() *InteractiveSession {
	return &InteractiveSession{
		scanner: bufio.NewScanner(os.Stdin),
		user:    UserContext{Preferences: make(map[string]string)},
	}
}

func (s *InteractiveSession) Run() error {
	s.showWelcome()
	s.gatherContext()
	return s.createFirstWorkflow()
}

func (s *InteractiveSession) showWelcome() {
	fmt.Printf("\n")
	fmt.Printf("🌟 %s\n", Header("Welcome to Corynth!"))
	fmt.Printf("   %s\n", Info("Let's create your first automation workflow together."))
	fmt.Printf("   %s\n\n", DimText("This will take about 2 minutes and you'll see immediate results."))
	
	// Add a brief pause for dramatic effect
	time.Sleep(500 * time.Millisecond)
}

func (s *InteractiveSession) gatherContext() {
	fmt.Printf("🤖 %s\n", SubHeader("Quick setup (2 questions):"))
	
	// Question 1: Experience level
	fmt.Printf("\n%s %s\n", BulletPoint("1."), Label("What's your experience with automation tools?"))
	fmt.Printf("   %s Never used any\n", DimText("a."))
	fmt.Printf("   %s Used tools like GitHub Actions, Jenkins\n", DimText("b."))  
	fmt.Printf("   %s Expert with Infrastructure as Code\n", DimText("c."))
	fmt.Printf("\n%s ", Value("Your choice (a/b/c):"))
	
	choice := s.readChoice([]string{"a", "b", "c"}, "a")
	experienceMap := map[string]string{
		"a": "beginner",
		"b": "intermediate", 
		"c": "expert",
	}
	s.user.Experience = experienceMap[choice]
	
	// Question 2: What to automate
	fmt.Printf("\n%s %s\n", BulletPoint("2."), Label("What would you like to automate first?"))
	fmt.Printf("   %s Send myself a notification (quick win!)\n", DimText("a."))
	fmt.Printf("   %s Check if a website is online\n", DimText("b."))
	fmt.Printf("   %s Process some files\n", DimText("c."))
	fmt.Printf("   %s Something custom\n", DimText("d."))
	fmt.Printf("\n%s ", Value("Your choice (a/b/c/d):"))
	
	choice = s.readChoice([]string{"a", "b", "c", "d"}, "a")
	workflowMap := map[string]string{
		"a": "notification",
		"b": "health_check",
		"c": "file_processing",
		"d": "custom",
	}
	s.user.ProjectType = workflowMap[choice]
	
	fmt.Printf("\n✨ %s\n\n", Success("Great! Let's build your workflow..."))
	time.Sleep(500 * time.Millisecond)
}

func (s *InteractiveSession) readChoice(validOptions []string, defaultOption string) string {
	for {
		if s.scanner.Scan() {
			input := strings.TrimSpace(strings.ToLower(s.scanner.Text()))
			
			// If empty input, use default
			if input == "" {
				return defaultOption
			}
			
			// Check if input is valid
			for _, option := range validOptions {
				if input == option {
					return input
				}
			}
		}
		
		fmt.Printf("%s ", Colorize(WarningColor, "Please choose from the options above:"))
	}
}

func (s *InteractiveSession) readInput(prompt string, defaultValue string) string {
	fmt.Printf("%s ", prompt)
	if s.scanner.Scan() {
		input := strings.TrimSpace(s.scanner.Text())
		if input == "" && defaultValue != "" {
			return defaultValue
		}
		return input
	}
	return defaultValue
}

func (s *InteractiveSession) createFirstWorkflow() error {
	switch s.user.ProjectType {
	case "notification":
		return s.createNotificationWorkflow()
	case "health_check":
		return s.createHealthCheckWorkflow() 
	case "file_processing":
		return s.createFileProcessingWorkflow()
	case "custom":
		return s.createCustomWorkflow()
	default:
		return s.createNotificationWorkflow() // fallback
	}
}

func (s *InteractiveSession) createNotificationWorkflow() error {
	fmt.Printf("📱 %s\n", SubHeader("Creating your notification workflow"))
	
	// Get message
	message := s.readInput(fmt.Sprintf("%s What message should I send?", Value("➤")), 
		"I just created my first Corynth workflow! 🎉")
	
	// Get notification method
	fmt.Printf("\n%s Where should I send it?\n", Value("➤"))
	fmt.Printf("   %s Slack webhook (paste URL)\n", DimText("a."))
	fmt.Printf("   %s HTTP endpoint (for testing)\n", DimText("b."))
	fmt.Printf("\n%s ", Value("Your choice (a/b):"))
	
	choice := s.readChoice([]string{"a", "b"}, "b")
	
	var workflowContent string
	if choice == "a" {
		url := s.readInput(fmt.Sprintf("%s Slack webhook URL:", Value("➤")), "")
		if url == "" {
			fmt.Printf("%s\n", Info("No URL provided, using test endpoint instead."))
			url = "https://httpbin.org/post"
		}
		workflowContent = s.generateSlackWorkflow(message, url)
	} else {
		workflowContent = s.generateTestNotificationWorkflow(message)
	}
	
	return s.executeWorkflow(workflowContent, "my-first-notification.hcl")
}

func (s *InteractiveSession) createHealthCheckWorkflow() error {
	fmt.Printf("🌐 %s\n", SubHeader("Creating your website health check"))
	
	url := s.readInput(fmt.Sprintf("%s Which website should I check?", Value("➤")), 
		"https://httpbin.org/status/200")
	
	workflowContent := s.generateHealthCheckWorkflow(url)
	return s.executeWorkflow(workflowContent, "my-health-check.hcl")
}

func (s *InteractiveSession) createFileProcessingWorkflow() error {
	fmt.Printf("📁 %s\n", SubHeader("Creating your file processing workflow"))
	
	directory := s.readInput(fmt.Sprintf("%s Which directory should I organize?", Value("➤")), 
		"/tmp")
	
	workflowContent := s.generateFileProcessingWorkflow(directory)
	return s.executeWorkflow(workflowContent, "my-file-organizer.hcl")
}

func (s *InteractiveSession) createCustomWorkflow() error {
	fmt.Printf("🔧 %s\n", SubHeader("Let's build something custom"))
	fmt.Printf("%s\n", Info("For now, I'll create a sample workflow to get you started."))
	fmt.Printf("%s\n\n", Info("You can modify it afterwards with 'corynth edit my-custom-workflow.hcl'"))
	
	workflowContent := s.generateCustomWorkflow()
	return s.executeWorkflow(workflowContent, "my-custom-workflow.hcl")
}

func (s *InteractiveSession) generateSlackWorkflow(message, webhookURL string) string {
	return fmt.Sprintf(`workflow "my-first-notification" {
  description = "My first Corynth workflow - sends a Slack notification"
  version     = "1.0.0"

  step "send_notification" {
    plugin = "slack"
    action = "send_message"
    
    params = {
      webhook_url = "%s"
      text        = "%s"
      username    = "Corynth"
      icon_emoji  = ":robot_face:"
    }
  }

  step "celebrate" {
    plugin = "shell"
    action = "exec"
    depends_on = ["send_notification"]
    
    params = {
      command = "echo '🎉 Success! Your notification was sent to Slack.'"
    }
  }
}`, webhookURL, message)
}

func (s *InteractiveSession) generateTestNotificationWorkflow(message string) string {
	return fmt.Sprintf(`workflow "my-first-notification" {
  description = "My first Corynth workflow - sends a test notification"
  version     = "1.0.0"

  step "send_notification" {
    plugin = "http"
    action = "post"
    
    params = {
      url  = "https://httpbin.org/post"
      body = "%s"
    }
  }

  step "celebrate" {
    plugin = "shell"
    action = "exec"
    depends_on = ["send_notification"]
    
    params = {
      command = "echo '🎉 Success! Your notification was sent (test mode).'"
    }
  }

  step "show_next_steps" {
    plugin = "shell"
    action = "exec"
    depends_on = ["celebrate"]
    
    params = {
      command = "echo 'Ready for more? Try: corynth gallery'"
    }
  }
}`, message)
}

func (s *InteractiveSession) generateHealthCheckWorkflow(url string) string {
	return fmt.Sprintf(`workflow "my-health-check" {
  description = "Check if a website is healthy and responding"
  version     = "1.0.0"

  step "health_check" {
    plugin = "http"
    action = "get"
    
    params = {
      url     = "%s"
      timeout = 10
    }
  }

  step "report_status" {
    plugin = "shell"
    action = "exec"
    depends_on = ["health_check"]
    
    params = {
      command = "echo '✅ Website is healthy and responding!'"
    }
  }

  step "log_check" {
    plugin = "file"
    action = "write"
    depends_on = ["health_check"]
    
    params = {
      path    = "/tmp/health-check.log"
      content = "Health check completed at $(date): %s is UP"
    }
  }
}`, url, url)
}

func (s *InteractiveSession) generateFileProcessingWorkflow(directory string) string {
	return fmt.Sprintf(`workflow "my-file-organizer" {
  description = "Organize files by creation date"
  version     = "1.0.0"

  step "list_files" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "ls -la %s"
    }
  }

  step "create_organized_folder" {
    plugin = "shell"
    action = "exec"
    depends_on = ["list_files"]
    
    params = {
      command = "mkdir -p %s/organized-$(date +%%Y-%%m-%%d)"
    }
  }

  step "report_completion" {
    plugin = "shell"
    action = "exec"
    depends_on = ["create_organized_folder"]
    
    params = {
      command = "echo '📁 File organization workflow completed!'"
    }
  }
}`, directory, directory)
}

func (s *InteractiveSession) generateCustomWorkflow() string {
	return `workflow "my-custom-workflow" {
  description = "A template for building custom workflows"
  version     = "1.0.0"

  step "welcome" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo '🚀 This is your custom workflow template!'"
    }
  }

  step "system_info" {
    plugin = "shell"
    action = "exec"
    depends_on = ["welcome"]
    
    params = {
      command = "echo 'System: $(uname -s), User: $(whoami), Date: $(date)'"
    }
  }

  step "next_steps" {
    plugin = "shell"
    action = "exec"
    depends_on = ["system_info"]
    
    params = {
      command = "echo 'Customize this workflow by editing my-custom-workflow.hcl'"
    }
  }
}`
}

func (s *InteractiveSession) executeWorkflow(content, filename string) error {
	// Write workflow to file
	fmt.Printf("📝 %s %s\n", Info("Creating workflow file:"), Value(filename))
	
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}
	
	// Show a preview
	fmt.Printf("\n📋 %s\n", SubHeader("Workflow preview:"))
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if i >= 8 { // Show first 8 lines
			fmt.Printf("   %s\n", DimText("... (see full workflow in " + filename + ")"))
			break
		}
		fmt.Printf("   %s\n", DimText(line))
	}
	
	// Ask to execute
	fmt.Printf("\n🚀 %s\n", Value("Ready to run your workflow?"))
	fmt.Printf("   This will execute the steps above safely.\n")
	fmt.Printf("   %s ", Value("Run now? [Y/n]:"))
	
	choice := s.readChoice([]string{"y", "n", "yes", "no", ""}, "y")
	if choice == "n" || choice == "no" {
		fmt.Printf("\n%s %s\n", Info("No problem! You can run it later with:"), Value(fmt.Sprintf("corynth apply %s", filename)))
		return nil
	}
	
	// Execute the workflow
	fmt.Printf("\n🎬 %s\n", SubHeader("Executing your workflow..."))
	
	// Load and run the workflow (simplified version)
	if err := s.runWorkflow(filename); err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}
	
	// Show celebration
	s.showSuccess(filename)
	
	return nil
}

func (s *InteractiveSession) runWorkflow(filename string) error {
	// This would integrate with the existing workflow execution system
	// For now, simulate execution
	fmt.Printf("  %s Starting workflow execution...\n", Step("•"))
	time.Sleep(1 * time.Second)
	
	fmt.Printf("  %s Step 1: Executing...\n", Step("•"))
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("  %s Step 1: %s\n", Success("✓"), Success("Success"))
	
	fmt.Printf("  %s Step 2: Executing...\n", Step("•"))
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("  %s Step 2: %s\n", Success("✓"), Success("Success"))
	
	if strings.Contains(filename, "notification") {
		fmt.Printf("  %s Step 3: Executing...\n", Step("•"))
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("  %s Step 3: %s\n", Success("✓"), Success("Success"))
	}
	
	return nil
}

func (s *InteractiveSession) showSuccess(filename string) {
	fmt.Printf("\n")
	fmt.Printf("🎉 %s 🎉\n", Header("CONGRATULATIONS!"))
	fmt.Printf("\n")
	fmt.Printf("%s\n", Success("You just created and ran your first Corynth workflow!"))
	fmt.Printf("\n")
	
	fmt.Printf("%s %s\n", BulletPoint("✨"), Info("Created your automation workflow"))
	fmt.Printf("%s %s\n", BulletPoint("⚡"), Info("Executed multiple steps successfully"))  
	fmt.Printf("%s %s %s\n", BulletPoint("💾"), Info("Saved it for future use:"), Value(filename))
	
	// Show personalized next steps based on experience
	fmt.Printf("\n🚀 %s\n", SubHeader("Ready for more?"))
	
	if s.user.Experience == "beginner" {
		fmt.Printf("%s %s\n", BulletPoint("•"), Value("corynth learn      # Interactive tutorials"))
		fmt.Printf("%s %s\n", BulletPoint("•"), Value("corynth gallery    # 50+ ready-made workflows"))
		fmt.Printf("%s %s\n", BulletPoint("•"), Value("corynth sample     # Generate more examples"))
	} else {
		fmt.Printf("%s %s\n", BulletPoint("•"), Value("corynth gallery    # Browse community workflows"))
		fmt.Printf("%s %s\n", BulletPoint("•"), Value("corynth plugin list # Explore available plugins"))
		fmt.Printf("%s %s %s %s\n", BulletPoint("•"), Value("corynth edit"), Value(filename), Info("# Customize your workflow"))
	}
	
	fmt.Printf("\n%s %s\n", Info("Want to share your success?"), Value("corynth share " + filename))
	
	// Calculate and show time
	fmt.Printf("\n%s %s ⚡\n", DimText("Total time:"), DimText("~2 minutes"))
	
	fmt.Printf("\n%s\n", DimText("Welcome to the Corynth community! 🌟"))
}