package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/corynth/corynth/pkg/config"
	"github.com/corynth/corynth/pkg/plugin"
	"github.com/corynth/corynth/pkg/state"
	"github.com/corynth/corynth/pkg/workflow"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

// loadConfig loads the configuration file
func loadConfig(configPath string) (*config.Config, error) {
	if configPath == "" {
		configPath = "corynth.hcl"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		// If config doesn't exist, return default config
		if os.IsNotExist(err) || strings.Contains(err.Error(), "Configuration file not found") {
			return &config.Config{
				Version: "1.0",
			}, nil
		}
		return nil, err
	}

	return cfg, nil
}

// initializePluginManager initializes the plugin manager (using gRPC v2 under the hood)
func initializePluginManager(cfg *config.Config) (*plugin.Manager, error) {
	// Default paths - check both bin/plugins (local dev) and .corynth/plugins (installed)
	pluginPaths := []string{"bin/plugins", ".corynth/plugins"}
	cachePath := ".corynth/cache"
	autoInstall := true

	// Override with config
	if cfg.Plugins != nil {
		autoInstall = cfg.Plugins.AutoInstall
		if cfg.Plugins.LocalPath != "" {
			pluginPaths = []string{cfg.Plugins.LocalPath}
		}
		if cfg.Plugins.Cache != nil && cfg.Plugins.Cache.Path != "" {
			cachePath = cfg.Plugins.Cache.Path
		}
	}

	// Use the first existing plugin path
	var pluginPath string
	for _, path := range pluginPaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			pluginPath = path
			break
		}
	}
	if pluginPath == "" {
		pluginPath = pluginPaths[0] // Default to first if none exist
	}

	// Create standard manager with gRPC support
	manager := plugin.NewManager(pluginPath, cachePath, autoInstall)

	// Add configured repositories
	hasRepositories := false
	if cfg.Plugins != nil && len(cfg.Plugins.Repositories) > 0 {
		hasRepositories = true
		for _, repo := range cfg.Plugins.Repositories {
			manager.AddRepository(plugin.Repository{
				Name:     repo.Name,
				URL:      repo.URL,
				Branch:   repo.Branch,
				Token:    repo.Token,
				Priority: repo.Priority,
			})
		}
	}
	
	// Add default official repository if no repositories configured
	if !hasRepositories {
		manager.AddRepository(plugin.Repository{
			Name:     "official",
			URL:      "https://github.com/corynth/plugins",
			Branch:   "main",
			Priority: 1,
		})
	}

	// Load local plugins
	if err := manager.LoadLocal(); err != nil {
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}

	return manager, nil
}

// loadVariables loads variables from command line and files
func loadVariables(cmd *cobra.Command) (map[string]interface{}, error) {
	variables := make(map[string]interface{})

	// Load from var-file if specified
	varFile, _ := cmd.Flags().GetString("var-file")
	if varFile != "" {
		fileVars, err := loadVariableFile(varFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load variable file: %w", err)
		}
		for k, v := range fileVars {
			variables[k] = v
		}
	}

	// Override with command line variables
	varFlags, _ := cmd.Flags().GetStringSlice("var")
	for _, varFlag := range varFlags {
		parts := strings.SplitN(varFlag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable format: %s (expected key=value)", varFlag)
		}
		variables[parts[0]] = parts[1]
	}

	return variables, nil
}

// loadVariableFile loads variables from a file
func loadVariableFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var variables map[string]interface{}
	
	// Determine format by extension
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &variables)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &variables)
	case ".covars", ".hcl":
		// Parse HCL variable file
		variables, err = parseHCLVariables(data)
	default:
		// Try JSON first, then YAML, then HCL
		if err = json.Unmarshal(data, &variables); err != nil {
			if err = yaml.Unmarshal(data, &variables); err != nil {
				variables, err = parseHCLVariables(data)
			}
		}
	}

	return variables, err
}

// findDefaultWorkflowFile finds the default workflow file
func findDefaultWorkflowFile() string {
	candidates := []string{
		"workflow.hcl",
		"workflow.tf",
		"corynth.hcl",
		"corynth.tf",
		// Legacy YAML support
		"workflow.yaml",
		"workflow.yml",
		"corynth.yaml",
		"corynth.yml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// createNotifier creates a notification handler
func createNotifier(cfg *config.NotificationConfig) (workflow.Notifier, error) {
	// TODO: Implement notification system
	return nil, nil
}

// createStateStore creates a state store
func createStateStore(cfg *config.Config) (workflow.StateStore, error) {
	if cfg.State == nil {
		return state.NewLocalStateStore(".corynth/state"), nil
	}

	switch cfg.State.Backend {
	case "local", "":
		path := ".corynth/state"
		if cfg.State.BackendConfig != nil {
			if p, ok := cfg.State.BackendConfig["path"]; ok {
				path = p
			}
		}
		return state.NewLocalStateStore(path), nil
	case "s3":
		// S3 state store not implemented yet, fallback to local
		return state.NewLocalStateStore(".corynth/state"), nil
	case "consul":
		// Consul state store not implemented yet, fallback to local
		return state.NewLocalStateStore(".corynth/state"), nil
	default:
		return nil, fmt.Errorf("unsupported state backend: %s", cfg.State.Backend)
	}
}

// formatDuration formats a duration nicely
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// printJSON prints an object as formatted JSON
func printJSON(obj interface{}) error {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// printTable prints data in a table format
func printTable(headers []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
	)

	for _, row := range rows {
		table.Append(row)
	}

	table.Render()
}

// validateWorkflowFile validates a workflow file exists and is readable
func validateWorkflowFile(path string) error {
	if path == "" {
		return fmt.Errorf("workflow file path is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("workflow file not found: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("workflow file path is a directory, not a file")
	}

	// Try to read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read workflow file: %w", err)
	}

	// Basic YAML validation
	var temp interface{}
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid YAML in workflow file: %w", err)
	}

	return nil
}

// ensureCorynthDir ensures the .corynth directory exists
func ensureCorynthDir() error {
	dirs := []string{
		".corynth",
		".corynth/state",
		".corynth/plugins",
		".corynth/cache",
		".corynth/logs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// parseHCLVariables parses HCL variable files (.covars, .hcl)
func parseHCLVariables(data []byte) (map[string]interface{}, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, "variables.hcl")
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %s", diags.Error())
	}

	// Define a simple structure to capture variable assignments
	var variableFile struct {
		Variables map[string]*hcl.Attribute `hcl:",remain"`
	}

	// Create evaluation context for HCL parsing
	evalCtx := &hcl.EvalContext{}

	// Use gohcl to decode the entire file content as variables
	diags = gohcl.DecodeBody(file.Body, evalCtx, &variableFile)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode HCL variables: %s", diags.Error())
	}

	// Evaluate each variable to get actual values instead of HCL attributes
	variables := make(map[string]interface{})
	for name, attr := range variableFile.Variables {
		val, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate variable %s: %s", name, diags.Error())
		}
		
		// Convert cty.Value to Go value
		if val.IsKnown() && !val.IsNull() {
			switch {
			case val.Type().Equals(cty.String):
				variables[name] = val.AsString()
			case val.Type().Equals(cty.Number):
				if floatVal, accuracy := val.AsBigFloat().Float64(); accuracy == 0 {
					if floatVal == float64(int64(floatVal)) {
						variables[name] = int64(floatVal)
					} else {
						variables[name] = floatVal
					}
				}
			case val.Type().Equals(cty.Bool):
				variables[name] = val.True()
			case val.Type().IsObjectType() || val.Type().IsMapType():
				// Convert object/map to Go map
				if !val.IsNull() {
					result := make(map[string]interface{})
					for it := val.ElementIterator(); it.Next(); {
						keyVal, elemVal := it.Element()
						key := keyVal.AsString()
						
						// Recursively convert nested values
						if elemVal.Type().Equals(cty.String) {
							result[key] = elemVal.AsString()
						} else if elemVal.Type().Equals(cty.Number) {
							if floatVal, accuracy := elemVal.AsBigFloat().Float64(); accuracy == 0 {
								if floatVal == float64(int64(floatVal)) {
									result[key] = int64(floatVal)
								} else {
									result[key] = floatVal
								}
							}
						} else if elemVal.Type().Equals(cty.Bool) {
							result[key] = elemVal.True()
						} else {
							result[key] = fmt.Sprintf("%v", elemVal)
						}
					}
					variables[name] = result
				}
			case val.Type().IsListType() || val.Type().IsSetType():
				// Convert list/set to Go slice
				if !val.IsNull() {
					var items []interface{}
					for it := val.ElementIterator(); it.Next(); {
						_, elemVal := it.Element()
						if elemVal.Type().Equals(cty.String) {
							items = append(items, elemVal.AsString())
						} else if elemVal.Type().Equals(cty.Number) {
							if floatVal, accuracy := elemVal.AsBigFloat().Float64(); accuracy == 0 {
								if floatVal == float64(int64(floatVal)) {
									items = append(items, int64(floatVal))
								} else {
									items = append(items, floatVal)
								}
							}
						} else if elemVal.Type().Equals(cty.Bool) {
							items = append(items, elemVal.True())
						} else {
							items = append(items, fmt.Sprintf("%v", elemVal))
						}
					}
					variables[name] = items
				}
			default:
				// For other complex types, convert to string representation
				variables[name] = fmt.Sprintf("%v", val)
			}
		}
	}

	return variables, nil
}