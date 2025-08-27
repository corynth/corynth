package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewGalleryCommand creates the workflow gallery command
func NewGalleryCommand() *cobra.Command {
	var (
		category string
		detailed bool
	)

	cmd := &cobra.Command{
		Use:   "gallery",
		Short: "Browse community workflows",
		Long: `Browse and discover workflows created by the Corynth community.
Find ready-made solutions for common automation tasks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGallery(cmd, args, category, detailed)
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category")
	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed information")

	return cmd
}

func runGallery(cmd *cobra.Command, args []string, category string, detailed bool) error {
	fmt.Printf("🎨 %s\n", Header("Corynth Workflow Gallery"))
	fmt.Printf("   %s\n\n", Info("Discover ready-made workflows from the community"))

	// For now, show curated examples - in production this would fetch from a registry
	workflows := getCuratedWorkflows()

	if category != "" {
		workflows = filterByCategory(workflows, category)
		fmt.Printf("📂 %s %s\n\n", SubHeader("Category:"), Value(category))
	}

	if len(workflows) == 0 {
		fmt.Printf("%s\n", Info("No workflows found for the specified criteria"))
		return nil
	}

	// Group workflows by category
	categories := groupWorkflowsByCategory(workflows)

	for cat, workflowList := range categories {
		fmt.Printf("📁 %s %s\n", SubHeader(cat), DimText(fmt.Sprintf("(%d workflows)", len(workflowList))))
		
		for _, wf := range workflowList {
			if detailed {
				showDetailedWorkflow(wf)
			} else {
				showWorkflowSummary(wf)
			}
		}
		fmt.Println()
	}

	// Show footer with instructions
	fmt.Printf("💡 %s\n", SubHeader("How to use:"))
	fmt.Printf("  %s corynth gallery --category deployment\n", BulletPoint("•"))
	fmt.Printf("  %s corynth gallery --detailed\n", BulletPoint("•"))
	fmt.Printf("  %s corynth start  # Interactive workflow builder\n", BulletPoint("•"))
	
	return nil
}

type GalleryWorkflow struct {
	Name        string
	Description string
	Category    string
	Author      string
	Stars       int
	Tags        []string
	Difficulty  string
	Time        string
	Example     string
}

func getCuratedWorkflows() []GalleryWorkflow {
	return []GalleryWorkflow{
		{
			Name:        "slack-notification",
			Description: "Send notifications to Slack channels",
			Category:    "Communication",
			Author:      "corynth-team",
			Stars:       124,
			Tags:        []string{"slack", "notifications", "beginner"},
			Difficulty:  "Beginner",
			Time:        "2 min",
			Example:     "Perfect for CI/CD notifications and alerts",
		},
		{
			Name:        "website-health-check",
			Description: "Monitor website uptime and response times",
			Category:    "Monitoring",
			Author:      "community",
			Stars:       89,
			Tags:        []string{"monitoring", "http", "health-check"},
			Difficulty:  "Beginner",
			Time:        "3 min",
			Example:     "Get alerts when your website goes down",
		},
		{
			Name:        "docker-build-deploy",
			Description: "Build Docker images and deploy to production",
			Category:    "Deployment",
			Author:      "devops-team",
			Stars:       156,
			Tags:        []string{"docker", "deployment", "ci-cd"},
			Difficulty:  "Intermediate",
			Time:        "10 min",
			Example:     "Complete Docker deployment pipeline",
		},
		{
			Name:        "aws-s3-backup",
			Description: "Automated backups to AWS S3 with encryption",
			Category:    "Backup",
			Author:      "community",
			Stars:       67,
			Tags:        []string{"aws", "s3", "backup", "encryption"},
			Difficulty:  "Intermediate",
			Time:        "8 min",
			Example:     "Secure automated backups to the cloud",
		},
		{
			Name:        "kubernetes-deployment",
			Description: "Deploy applications to Kubernetes clusters",
			Category:    "Deployment",
			Author:      "k8s-experts",
			Stars:       203,
			Tags:        []string{"kubernetes", "deployment", "helm"},
			Difficulty:  "Advanced",
			Time:        "15 min",
			Example:     "Production-ready Kubernetes deployments",
		},
		{
			Name:        "database-maintenance",
			Description: "Automated database backups and maintenance tasks",
			Category:    "Database",
			Author:      "dba-team",
			Stars:       91,
			Tags:        []string{"mysql", "postgres", "backup", "maintenance"},
			Difficulty:  "Intermediate",
			Time:        "12 min",
			Example:     "Keep your databases healthy and backed up",
		},
		{
			Name:        "log-analysis",
			Description: "Parse and analyze application logs for issues",
			Category:    "Monitoring",
			Author:      "community",
			Stars:       78,
			Tags:        []string{"logs", "analysis", "monitoring", "alerts"},
			Difficulty:  "Intermediate",
			Time:        "6 min",
			Example:     "Automated log analysis and alerting",
		},
		{
			Name:        "file-organizer",
			Description: "Organize files by date, type, or custom rules",
			Category:    "Automation",
			Author:      "community",
			Stars:       45,
			Tags:        []string{"files", "organization", "cleanup"},
			Difficulty:  "Beginner",
			Time:        "4 min",
			Example:     "Keep your downloads folder clean",
		},
		{
			Name:        "api-integration-test",
			Description: "Comprehensive API testing and validation",
			Category:    "Testing",
			Author:      "qa-team",
			Stars:       112,
			Tags:        []string{"api", "testing", "validation", "http"},
			Difficulty:  "Intermediate",
			Time:        "7 min",
			Example:     "Automated API testing pipeline",
		},
		{
			Name:        "security-scan",
			Description: "Security vulnerability scanning and reporting",
			Category:    "Security",
			Author:      "security-team",
			Stars:       134,
			Tags:        []string{"security", "scanning", "vulnerability", "reports"},
			Difficulty:  "Advanced",
			Time:        "20 min",
			Example:     "Comprehensive security auditing",
		},
	}
}

func filterByCategory(workflows []GalleryWorkflow, category string) []GalleryWorkflow {
	var filtered []GalleryWorkflow
	categoryLower := strings.ToLower(category)
	
	for _, wf := range workflows {
		if strings.ToLower(wf.Category) == categoryLower {
			filtered = append(filtered, wf)
		}
	}
	
	return filtered
}

func groupWorkflowsByCategory(workflows []GalleryWorkflow) map[string][]GalleryWorkflow {
	groups := make(map[string][]GalleryWorkflow)
	
	for _, wf := range workflows {
		groups[wf.Category] = append(groups[wf.Category], wf)
	}
	
	return groups
}

func showWorkflowSummary(wf GalleryWorkflow) {
	stars := ""
	if wf.Stars > 0 {
		stars = fmt.Sprintf(" ⭐ %d", wf.Stars)
	}
	
	difficulty := getDifficultyColor(wf.Difficulty)
	
	fmt.Printf("  %s %s %s%s\n", 
		BulletPoint("•"), 
		Label(wf.Name), 
		Colorize(difficulty, wf.Difficulty),
		stars)
	fmt.Printf("    %s\n", DimText(wf.Description))
	
	if len(wf.Tags) > 0 {
		fmt.Printf("    %s %s\n", DimText("Tags:"), DimText(strings.Join(wf.Tags, ", ")))
	}
}

func showDetailedWorkflow(wf GalleryWorkflow) {
	difficulty := getDifficultyColor(wf.Difficulty)
	
	fmt.Printf("  %s %s\n", SubHeader("📋"), Header(wf.Name))
	fmt.Printf("    %s %s\n", Label("Description:"), Value(wf.Description))
	fmt.Printf("    %s %s\n", Label("Author:"), Value(wf.Author))
	fmt.Printf("    %s %s\n", Label("Difficulty:"), Colorize(difficulty, wf.Difficulty))
	fmt.Printf("    %s %s\n", Label("Est. Time:"), Value(wf.Time))
	
	if wf.Stars > 0 {
		fmt.Printf("    %s %s\n", Label("Community:"), Colorize(SuccessColor, fmt.Sprintf("⭐ %d stars", wf.Stars)))
	}
	
	if len(wf.Tags) > 0 {
		fmt.Printf("    %s %s\n", Label("Tags:"), DimText(strings.Join(wf.Tags, ", ")))
	}
	
	if wf.Example != "" {
		fmt.Printf("    %s %s\n", Label("Use case:"), DimText(wf.Example))
	}
	
	fmt.Printf("    %s corynth start --template %s\n", DimText("Try it:"), DimText(wf.Name))
	fmt.Println()
}

func getDifficultyColor(difficulty string) string {
	switch strings.ToLower(difficulty) {
	case "beginner":
		return SuccessColor
	case "intermediate":
		return WarningColor
	case "advanced":
		return ErrorColor
	default:
		return ""
	}
}