package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// NewPluginDevCommand creates the plugin development command group
func NewPluginDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin-dev",
		Short: "Plugin development tools",
		Long: `Plugin development tools for creating, testing, and scaffolding Corynth plugins.

Available commands:
  init     - Generate a new plugin from template
  test     - Test a plugin locally
  scaffold - Create plugin scaffolding`,
	}

	cmd.AddCommand(NewPluginInitCommand())
	cmd.AddCommand(NewPluginTestCommand())

	return cmd
}

// NewPluginInitCommand creates the plugin init command
func NewPluginInitCommand() *cobra.Command {
	var (
		pluginType   string
		outputDir    string
		authorName   string
		authorEmail  string
		license      string
		interactive  bool
	)

	cmd := &cobra.Command{
		Use:   "init [plugin-name]",
		Short: "Generate a new plugin from template",
		Long: `Generate a new plugin with complete scaffolding including:
  - Plugin interface implementation
  - Go module setup
  - Documentation templates
  - Sample workflows
  - Test framework
  - CI/CD configuration

Available plugin types:
  http      - HTTP client plugin for API calls
  database  - Database operations plugin
  command   - Command execution plugin
  file      - File system operations plugin
  api       - Generic API integration plugin
  custom    - Basic custom plugin template`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]
			return runPluginInit(pluginName, pluginType, outputDir, authorName, authorEmail, license, interactive)
		},
	}

	cmd.Flags().StringVarP(&pluginType, "type", "t", "custom", "Plugin type (http, database, command, file, api, custom)")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for the plugin")
	cmd.Flags().StringVar(&authorName, "author", "", "Author name")
	cmd.Flags().StringVar(&authorEmail, "email", "", "Author email")
	cmd.Flags().StringVar(&license, "license", "Apache-2.0", "License type")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode with prompts")

	return cmd
}


// NewPluginTestCommand creates the plugin test command
func NewPluginTestCommand() *cobra.Command {
	var (
		workflowFile string
		verbose      bool
	)

	cmd := &cobra.Command{
		Use:   "test [plugin-path]",
		Short: "Test a plugin locally",
		Long:  "Build and test a plugin with sample workflows",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginPath := args[0]
			return runPluginTest(pluginPath, workflowFile, verbose)
		},
	}

	cmd.Flags().StringVarP(&workflowFile, "workflow", "w", "", "Test with specific workflow file")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

// Plugin scaffolding data structures
type PluginTemplate struct {
	Name            string
	Type            string
	Description     string
	Author          string
	Email           string
	License         string
	Year            int
	ModulePath      string
	PackageName     string
	ClassName       string
	Actions         []ActionTemplate
	Dependencies    []string
	HasDatabase     bool
	HasHTTP         bool
	HasFileSystem   bool
	HasExternal     bool
}

type ActionTemplate struct {
	Name        string
	Description string
	Method      string
	Params      []ParamTemplate
	Returns     []ReturnTemplate
}

type ParamTemplate struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     string
}

type ReturnTemplate struct {
	Name        string
	Type        string
	Description string
}

func runPluginInit(pluginName, pluginType, outputDir, authorName, authorEmail, license string, interactive bool) error {
	// Interactive mode for gathering information
	if interactive {
		fmt.Printf("🚀 Corynth Plugin Generator\n\n")
		
		if authorName == "" {
			fmt.Print("Author name: ")
			fmt.Scanln(&authorName)
		}
		
		if authorEmail == "" {
			fmt.Print("Author email: ")
			fmt.Scanln(&authorEmail)
		}
		
		if pluginType == "custom" {
			fmt.Print("Plugin type (http, database, command, file, api, custom): ")
			fmt.Scanln(&pluginType)
		}
		
		fmt.Print("Description: ")
		var description string
		fmt.Scanln(&description)
	}

	// Validate plugin name
	if !isValidPluginName(pluginName) {
		return fmt.Errorf("invalid plugin name: %s (use lowercase letters, numbers, and hyphens only)", pluginName)
	}

	// Create plugin directory
	pluginDir := filepath.Join(outputDir, pluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Generate plugin template data
	templateData := &PluginTemplate{
		Name:        pluginName,
		Type:        pluginType,
		Author:      authorName,
		Email:       authorEmail,
		License:     license,
		Year:        time.Now().Year(),
		ModulePath:  fmt.Sprintf("github.com/corynth/corynthplugins/%s", pluginName),
		PackageName: strings.ReplaceAll(pluginName, "-", ""),
		ClassName:   toCamelCase(pluginName) + "Plugin",
	}

	// Set type-specific properties
	switch pluginType {
	case "http":
		templateData.Description = fmt.Sprintf("HTTP client plugin for %s API integration", pluginName)
		templateData.HasHTTP = true
		templateData.Actions = httpActions()
		templateData.Dependencies = []string{"net/http", "io", "time", "encoding/json"}
	case "database":
		templateData.Description = fmt.Sprintf("Database operations plugin for %s", pluginName)
		templateData.HasDatabase = true
		templateData.Actions = databaseActions()
		templateData.Dependencies = []string{"database/sql", "context", "fmt"}
	case "command":
		templateData.Description = fmt.Sprintf("Command execution plugin for %s operations", pluginName)
		templateData.HasExternal = true
		templateData.Actions = commandActions()
		templateData.Dependencies = []string{"os/exec", "context", "strings"}
	case "file":
		templateData.Description = fmt.Sprintf("File system operations plugin for %s", pluginName)
		templateData.HasFileSystem = true
		templateData.Actions = fileActions()
		templateData.Dependencies = []string{"os", "io", "path/filepath"}
	case "api":
		templateData.Description = fmt.Sprintf("API integration plugin for %s", pluginName)
		templateData.HasHTTP = true
		templateData.Actions = apiActions()
		templateData.Dependencies = []string{"net/http", "encoding/json", "fmt", "time"}
	default:
		templateData.Description = fmt.Sprintf("Custom plugin for %s functionality", pluginName)
		templateData.Actions = customActions()
		templateData.Dependencies = []string{"context", "fmt"}
	}

	fmt.Printf("🔨 Generating %s plugin '%s'...\n", pluginType, pluginName)

	// Generate all plugin files
	if err := generatePluginFiles(pluginDir, templateData); err != nil {
		return fmt.Errorf("failed to generate plugin files: %w", err)
	}

	// Print success message with next steps
	printSuccessMessage(pluginName, pluginDir, pluginType)

	return nil
}

func generatePluginFiles(pluginDir string, data *PluginTemplate) error {
	// Generate plugin.go
	if err := generateFromTemplate(filepath.Join(pluginDir, "plugin.go"), pluginGoTemplate, data); err != nil {
		return err
	}

	// Generate go.mod
	if err := generateFromTemplate(filepath.Join(pluginDir, "go.mod"), goModTemplate, data); err != nil {
		return err
	}

	// Generate README.md
	if err := generateFromTemplate(filepath.Join(pluginDir, "README.md"), readmeTemplate, data); err != nil {
		return err
	}

	// Generate plugin_test.go
	if err := generateFromTemplate(filepath.Join(pluginDir, "plugin_test.go"), testTemplate, data); err != nil {
		return err
	}

	// Create samples directory and files
	samplesDir := filepath.Join(pluginDir, "samples")
	if err := os.MkdirAll(samplesDir, 0755); err != nil {
		return err
	}

	if err := generateFromTemplate(filepath.Join(samplesDir, "basic-usage.hcl"), basicWorkflowTemplate, data); err != nil {
		return err
	}

	if err := generateFromTemplate(filepath.Join(samplesDir, "advanced-features.hcl"), advancedWorkflowTemplate, data); err != nil {
		return err
	}

	// Generate .github/workflows/test.yml
	workflowDir := filepath.Join(pluginDir, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return err
	}

	if err := generateFromTemplate(filepath.Join(workflowDir, "test.yml"), githubWorkflowTemplate, data); err != nil {
		return err
	}

	// Generate Makefile
	if err := generateFromTemplate(filepath.Join(pluginDir, "Makefile"), makefileTemplate, data); err != nil {
		return err
	}

	return nil
}

func generateFromTemplate(filename, templateText string, data *PluginTemplate) error {
	tmpl, err := template.New(filepath.Base(filename)).Funcs(template.FuncMap{
		"ToUpper":     strings.ToUpper,
		"ToCamelCase": toCamelCase,
	}).Parse(templateText)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	// Use a buffer to process the template first
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Replace GitHub Actions placeholders
	content := buf.String()
	content = strings.ReplaceAll(content, "RUNNER_OS", "${{ runner.os }}")
	content = strings.ReplaceAll(content, "HASH_FILES", "${{ hashFiles('**/go.sum') }}")

	// Write the final content
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func isValidPluginName(name string) bool {
	if len(name) == 0 {
		return false
	}
	
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	
	return true
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "-")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result
}

func printSuccessMessage(pluginName, pluginDir, pluginType string) {
	fmt.Printf("\n✅ Successfully generated %s plugin '%s'\n\n", pluginType, pluginName)
	fmt.Printf("📁 Plugin created in: %s\n\n", pluginDir)
	fmt.Printf("🚀 Next steps:\n")
	fmt.Printf("   1. cd %s\n", pluginName)
	fmt.Printf("   2. go mod tidy\n")
	fmt.Printf("   3. make test\n")
	fmt.Printf("   4. corynth plugin test .\n")
	fmt.Printf("   5. Edit plugin.go to implement your logic\n")
	fmt.Printf("   6. Test with: corynth apply samples/basic-usage.hcl\n\n")
	fmt.Printf("📚 Documentation: docs/PLUGIN_DEVELOPMENT_GUIDE.md\n")
	fmt.Printf("💡 Examples: Check samples/ directory for workflow examples\n\n")
}


func runPluginTest(pluginPath, workflowFile string, verbose bool) error {
	fmt.Printf("Testing plugin at %s - to be implemented\n", pluginPath)
	return nil
}

// Action generators for different plugin types
func httpActions() []ActionTemplate {
	return []ActionTemplate{
		{
			Name:        "get",
			Description: "Make HTTP GET request",
			Method:      "executeGet",
			Params: []ParamTemplate{
				{Name: "url", Type: "string", Description: "Target URL", Required: true},
				{Name: "headers", Type: "object", Description: "HTTP headers", Required: false},
				{Name: "timeout", Type: "number", Description: "Request timeout in seconds", Required: false, Default: "30"},
			},
			Returns: []ReturnTemplate{
				{Name: "status_code", Type: "number", Description: "HTTP status code"},
				{Name: "body", Type: "string", Description: "Response body"},
				{Name: "headers", Type: "object", Description: "Response headers"},
			},
		},
		{
			Name:        "post",
			Description: "Make HTTP POST request",
			Method:      "executePost",
			Params: []ParamTemplate{
				{Name: "url", Type: "string", Description: "Target URL", Required: true},
				{Name: "body", Type: "string", Description: "Request body", Required: false},
				{Name: "headers", Type: "object", Description: "HTTP headers", Required: false},
				{Name: "timeout", Type: "number", Description: "Request timeout in seconds", Required: false, Default: "30"},
			},
			Returns: []ReturnTemplate{
				{Name: "status_code", Type: "number", Description: "HTTP status code"},
				{Name: "body", Type: "string", Description: "Response body"},
				{Name: "headers", Type: "object", Description: "Response headers"},
			},
		},
	}
}

func databaseActions() []ActionTemplate {
	return []ActionTemplate{
		{
			Name:        "query",
			Description: "Execute SQL query",
			Method:      "executeQuery",
			Params: []ParamTemplate{
				{Name: "connection_string", Type: "string", Description: "Database connection string", Required: true},
				{Name: "query", Type: "string", Description: "SQL query to execute", Required: true},
				{Name: "timeout", Type: "number", Description: "Query timeout in seconds", Required: false, Default: "30"},
			},
			Returns: []ReturnTemplate{
				{Name: "rows", Type: "array", Description: "Query result rows"},
				{Name: "count", Type: "number", Description: "Number of rows returned"},
				{Name: "columns", Type: "array", Description: "Column names"},
			},
		},
		{
			Name:        "execute",
			Description: "Execute SQL statement",
			Method:      "executeStatement", 
			Params: []ParamTemplate{
				{Name: "connection_string", Type: "string", Description: "Database connection string", Required: true},
				{Name: "statement", Type: "string", Description: "SQL statement to execute", Required: true},
				{Name: "timeout", Type: "number", Description: "Statement timeout in seconds", Required: false, Default: "30"},
			},
			Returns: []ReturnTemplate{
				{Name: "affected_rows", Type: "number", Description: "Number of affected rows"},
				{Name: "last_insert_id", Type: "number", Description: "Last inserted ID"},
			},
		},
	}
}

func commandActions() []ActionTemplate {
	return []ActionTemplate{
		{
			Name:        "exec",
			Description: "Execute command",
			Method:      "executeCommand",
			Params: []ParamTemplate{
				{Name: "command", Type: "string", Description: "Command to execute", Required: true},
				{Name: "args", Type: "array", Description: "Command arguments", Required: false},
				{Name: "working_dir", Type: "string", Description: "Working directory", Required: false},
				{Name: "env", Type: "object", Description: "Environment variables", Required: false},
				{Name: "timeout", Type: "number", Description: "Command timeout in seconds", Required: false, Default: "300"},
			},
			Returns: []ReturnTemplate{
				{Name: "output", Type: "string", Description: "Command output"},
				{Name: "exit_code", Type: "number", Description: "Exit code"},
				{Name: "success", Type: "boolean", Description: "Whether command succeeded"},
			},
		},
	}
}

func fileActions() []ActionTemplate {
	return []ActionTemplate{
		{
			Name:        "read",
			Description: "Read file contents",
			Method:      "executeRead",
			Params: []ParamTemplate{
				{Name: "path", Type: "string", Description: "File path to read", Required: true},
				{Name: "encoding", Type: "string", Description: "File encoding", Required: false, Default: "utf-8"},
			},
			Returns: []ReturnTemplate{
				{Name: "content", Type: "string", Description: "File contents"},
				{Name: "size", Type: "number", Description: "File size in bytes"},
				{Name: "modified", Type: "number", Description: "Last modified timestamp"},
			},
		},
		{
			Name:        "write",
			Description: "Write file contents",
			Method:      "executeWrite",
			Params: []ParamTemplate{
				{Name: "path", Type: "string", Description: "File path to write", Required: true},
				{Name: "content", Type: "string", Description: "Content to write", Required: true},
				{Name: "mode", Type: "string", Description: "File permissions", Required: false, Default: "0644"},
			},
			Returns: []ReturnTemplate{
				{Name: "bytes_written", Type: "number", Description: "Number of bytes written"},
				{Name: "success", Type: "boolean", Description: "Whether write succeeded"},
			},
		},
	}
}

func apiActions() []ActionTemplate {
	return []ActionTemplate{
		{
			Name:        "call",
			Description: "Make API call",
			Method:      "executeAPICall",
			Params: []ParamTemplate{
				{Name: "endpoint", Type: "string", Description: "API endpoint", Required: true},
				{Name: "method", Type: "string", Description: "HTTP method", Required: false, Default: "GET"},
				{Name: "data", Type: "object", Description: "Request data", Required: false},
				{Name: "auth", Type: "object", Description: "Authentication credentials", Required: false},
			},
			Returns: []ReturnTemplate{
				{Name: "data", Type: "object", Description: "Response data"},
				{Name: "status", Type: "number", Description: "HTTP status code"},
				{Name: "success", Type: "boolean", Description: "Whether call succeeded"},
			},
		},
	}
}

func customActions() []ActionTemplate {
	return []ActionTemplate{
		{
			Name:        "execute",
			Description: "Execute custom action",
			Method:      "executeCustom",
			Params: []ParamTemplate{
				{Name: "input", Type: "string", Description: "Input parameter", Required: true},
				{Name: "options", Type: "object", Description: "Optional parameters", Required: false},
			},
			Returns: []ReturnTemplate{
				{Name: "result", Type: "string", Description: "Operation result"},
				{Name: "success", Type: "boolean", Description: "Whether operation succeeded"},
			},
		},
	}
}