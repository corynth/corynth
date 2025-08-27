package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	pluginv2 "github.com/corynth/corynth/pkg/plugin/v2"
	"github.com/hashicorp/terraform-exec/tfexec"
)

// TerraformPlugin implements Terraform operations via gRPC
type TerraformPlugin struct {
	*pluginv2.BasePlugin
	workingDir string
	tf         *tfexec.Terraform
}

// NewTerraformPlugin creates a new Terraform plugin
func NewTerraformPlugin() *TerraformPlugin {
	base := pluginv2.NewBuilder("terraform", "1.0.0").
		Description("Terraform infrastructure as code operations and management").
		Author("Corynth Team").
		Tags("terraform", "infrastructure", "iac", "provisioning").
		Action("init", "Initialize Terraform working directory").
		InputWithDefault("working_dir", "string", "Working directory path", ".").
		InputWithDefault("backend_config", "object", "Backend configuration", map[string]string{}).
		InputWithDefault("upgrade", "boolean", "Upgrade providers and modules", false).
		Output("working_dir", "string", "Initialized working directory").
		Output("status", "string", "Initialization status").
		Add().
		Action("plan", "Create Terraform execution plan").
		InputWithDefault("working_dir", "string", "Working directory path", ".").
		InputWithDefault("var_file", "string", "Variables file path", "").
		InputWithDefault("vars", "object", "Terraform variables", map[string]string{}).
		InputWithDefault("destroy", "boolean", "Plan destroy operation", false).
		InputWithDefault("out_file", "string", "Save plan to file", "").
		Output("plan_file", "string", "Plan file path").
		Output("changes", "number", "Number of planned changes").
		Output("status", "string", "Plan status").
		Add().
		Action("apply", "Apply Terraform configuration").
		InputWithDefault("working_dir", "string", "Working directory path", ".").
		InputWithDefault("plan_file", "string", "Plan file to apply", "").
		InputWithDefault("var_file", "string", "Variables file path", "").
		InputWithDefault("vars", "object", "Terraform variables", map[string]string{}).
		InputWithDefault("auto_approve", "boolean", "Auto-approve changes", false).
		Output("resources_added", "number", "Resources added").
		Output("resources_changed", "number", "Resources changed").
		Output("resources_destroyed", "number", "Resources destroyed").
		Output("status", "string", "Apply status").
		Add().
		Action("destroy", "Destroy Terraform-managed infrastructure").
		InputWithDefault("working_dir", "string", "Working directory path", ".").
		InputWithDefault("var_file", "string", "Variables file path", "").
		InputWithDefault("vars", "object", "Terraform variables", map[string]string{}).
		InputWithDefault("auto_approve", "boolean", "Auto-approve destruction", false).
		Output("resources_destroyed", "number", "Resources destroyed").
		Output("status", "string", "Destroy status").
		Add().
		Action("output", "Read Terraform output values").
		InputWithDefault("working_dir", "string", "Working directory path", ".").
		InputWithDefault("output_name", "string", "Specific output name", "").
		Output("outputs", "object", "Terraform outputs").
		Output("count", "number", "Number of outputs").
		Add().
		Action("validate", "Validate Terraform configuration").
		InputWithDefault("working_dir", "string", "Working directory path", ".").
		Output("valid", "boolean", "Configuration is valid").
		Output("error_count", "number", "Number of validation errors").
		Output("warning_count", "number", "Number of warnings").
		Output("status", "string", "Validation status").
		Add().
		Action("fmt", "Format Terraform configuration files").
		InputWithDefault("working_dir", "string", "Working directory path", ".").
		InputWithDefault("recursive", "boolean", "Format recursively", true).
		InputWithDefault("write", "boolean", "Write changes to files", true).
		Output("files_formatted", "array", "List of formatted files").
		Output("status", "string", "Format status").
		Add().
		Build()

	plugin := &TerraformPlugin{
		BasePlugin: base,
	}

	return plugin
}

// Execute implements the plugin execution
func (p *TerraformPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	// Initialize Terraform
	if err := p.initTerraform(params); err != nil {
		return nil, fmt.Errorf("failed to initialize Terraform: %w", err)
	}

	switch action {
	case "init":
		return p.handleInit(ctx, params)
	case "plan":
		return p.handlePlan(ctx, params)
	case "apply":
		return p.handleApply(ctx, params)
	case "destroy":
		return p.handleDestroy(ctx, params)
	case "output":
		return p.handleOutput(ctx, params)
	case "validate":
		return p.handleValidate(ctx, params)
	case "fmt":
		return p.handleFormat(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Validate validates parameters
func (p *TerraformPlugin) Validate(params map[string]interface{}) error {
	if err := p.BasePlugin.Validate(params); err != nil {
		return err
	}

	// Validate working directory exists
	if workingDir, exists := params["working_dir"]; exists {
		if dir, ok := workingDir.(string); ok && dir != "" {
			if info, err := os.Stat(dir); err != nil || !info.IsDir() {
				return fmt.Errorf("working directory does not exist or is not a directory: %s", dir)
			}
		}
	}

	return nil
}

// initTerraform initializes the Terraform executor
func (p *TerraformPlugin) initTerraform(params map[string]interface{}) error {
	workingDir := getStringParam(params, "working_dir", ".")
	
	// Convert to absolute path
	absPath, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	p.workingDir = absPath

	// Find terraform executable
	terraformPath := "terraform"
	if customPath := os.Getenv("TERRAFORM_PATH"); customPath != "" {
		terraformPath = customPath
	}

	// Create terraform executor
	p.tf, err = tfexec.NewTerraform(p.workingDir, terraformPath)
	if err != nil {
		return fmt.Errorf("failed to create Terraform executor: %w", err)
	}

	return nil
}

// handleInit initializes Terraform working directory
func (p *TerraformPlugin) handleInit(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	upgrade := getBoolParam(params, "upgrade", false)
	backendConfig := getBackendConfig(params)

	initOptions := []tfexec.InitOption{}
	
	if upgrade {
		initOptions = append(initOptions, tfexec.Upgrade(true))
	}

	for key, value := range backendConfig {
		initOptions = append(initOptions, tfexec.BackendConfig(fmt.Sprintf("%s=%s", key, value)))
	}

	err := p.tf.Init(ctx, initOptions...)
	if err != nil {
		return nil, fmt.Errorf("terraform init failed: %w", err)
	}

	return map[string]interface{}{
		"working_dir": p.workingDir,
		"status":      "initialized",
	}, nil
}

// handlePlan creates execution plan
func (p *TerraformPlugin) handlePlan(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	varFile := getStringParam(params, "var_file", "")
	vars := getVars(params)
	destroy := getBoolParam(params, "destroy", false)
	outFile := getStringParam(params, "out_file", "")

	planOptions := []tfexec.PlanOption{}

	if varFile != "" {
		planOptions = append(planOptions, tfexec.VarFile(varFile))
	}

	for key, value := range vars {
		planOptions = append(planOptions, tfexec.Var(fmt.Sprintf("%s=%s", key, value)))
	}

	if destroy {
		planOptions = append(planOptions, tfexec.Destroy(true))
	}

	if outFile != "" {
		planOptions = append(planOptions, tfexec.Out(outFile))
	}

	hasChanges, err := p.tf.Plan(ctx, planOptions...)
	if err != nil {
		return nil, fmt.Errorf("terraform plan failed: %w", err)
	}

	changes := 0.0
	if hasChanges {
		changes = 1.0 // Simplified - would need to parse plan for exact count
	}

	result := map[string]interface{}{
		"changes": changes,
		"status":  "planned",
	}

	if outFile != "" {
		result["plan_file"] = outFile
	}

	return result, nil
}

// handleApply applies the configuration
func (p *TerraformPlugin) handleApply(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	planFile := getStringParam(params, "plan_file", "")
	varFile := getStringParam(params, "var_file", "")
	vars := getVars(params)
	autoApprove := getBoolParam(params, "auto_approve", false)

	applyOptions := []tfexec.ApplyOption{}

	if planFile != "" {
		applyOptions = append(applyOptions, tfexec.DirOrPlan(planFile))
	} else {
		if varFile != "" {
			applyOptions = append(applyOptions, tfexec.VarFile(varFile))
		}

		for key, value := range vars {
			applyOptions = append(applyOptions, tfexec.Var(fmt.Sprintf("%s=%s", key, value)))
		}
	}

	if autoApprove {
		applyOptions = append(applyOptions, tfexec.Lock(false))
	}

	err := p.tf.Apply(ctx, applyOptions...)
	if err != nil {
		return nil, fmt.Errorf("terraform apply failed: %w", err)
	}

	// For simplicity, return generic success metrics
	// In a real implementation, you'd parse the apply output
	return map[string]interface{}{
		"resources_added":     1.0, // Simplified
		"resources_changed":   0.0,
		"resources_destroyed": 0.0,
		"status":              "applied",
	}, nil
}

// handleDestroy destroys the infrastructure
func (p *TerraformPlugin) handleDestroy(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	varFile := getStringParam(params, "var_file", "")
	vars := getVars(params)
	autoApprove := getBoolParam(params, "auto_approve", false)

	destroyOptions := []tfexec.DestroyOption{}

	if varFile != "" {
		destroyOptions = append(destroyOptions, tfexec.VarFile(varFile))
	}

	for key, value := range vars {
		destroyOptions = append(destroyOptions, tfexec.Var(fmt.Sprintf("%s=%s", key, value)))
	}

	if autoApprove {
		destroyOptions = append(destroyOptions, tfexec.Lock(false))
	}

	err := p.tf.Destroy(ctx, destroyOptions...)
	if err != nil {
		return nil, fmt.Errorf("terraform destroy failed: %w", err)
	}

	return map[string]interface{}{
		"resources_destroyed": 1.0, // Simplified
		"status":              "destroyed",
	}, nil
}

// handleOutput reads output values
func (p *TerraformPlugin) handleOutput(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	outputName := getStringParam(params, "output_name", "")

	// Get all outputs (simplified)
	outputs, err := p.tf.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get outputs: %w", err)
	}

	outputMap := make(map[string]interface{})
	for name, output := range outputs {
		outputMap[name] = fmt.Sprintf("%v", output)
	}

	if outputName != "" {
		// Return specific output if requested
		if value, exists := outputMap[outputName]; exists {
			return map[string]interface{}{
				"outputs": map[string]interface{}{
					outputName: value,
				},
				"count": 1.0,
			}, nil
		}
		return nil, fmt.Errorf("output %s not found", outputName)
	}

	return map[string]interface{}{
		"outputs": outputMap,
		"count":   float64(len(outputs)),
	}, nil
}

// handleValidate validates the configuration
func (p *TerraformPlugin) handleValidate(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	validateOutput, err := p.tf.Validate(ctx)
	if err != nil {
		return nil, fmt.Errorf("terraform validate failed: %w", err)
	}

	valid := validateOutput.Valid
	errorCount := float64(len(validateOutput.Diagnostics))
	warningCount := 0.0 // Simplified for now

	status := "invalid"
	if valid {
		status = "valid"
	}

	return map[string]interface{}{
		"valid":         valid,
		"error_count":   errorCount,
		"warning_count": warningCount,
		"status":        status,
	}, nil
}

// handleFormat formats Terraform files
func (p *TerraformPlugin) handleFormat(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	recursive := getBoolParam(params, "recursive", true)

	var err error
	if recursive {
		err = p.tf.FormatWrite(ctx, tfexec.Recursive(true))
	} else {
		err = p.tf.FormatWrite(ctx)
	}
	
	if err != nil {
		return nil, fmt.Errorf("terraform fmt failed: %w", err)
	}

	// For simplicity, return generic success
	// In a real implementation, you'd capture the list of formatted files
	return map[string]interface{}{
		"files_formatted": []string{"*.tf"}, // Simplified
		"status":          "formatted",
	}, nil
}

// Helper functions
func getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, exists := params[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getNumberParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if val, exists := params[key]; exists {
		if num, ok := val.(float64); ok {
			return num
		}
		if str, ok := val.(string); ok {
			if num, err := strconv.ParseFloat(str, 64); err == nil {
				return num
			}
		}
	}
	return defaultValue
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, exists := params[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getBackendConfig(params map[string]interface{}) map[string]string {
	config := make(map[string]string)
	if backendConfig, exists := params["backend_config"]; exists {
		if configMap, ok := backendConfig.(map[string]interface{}); ok {
			for k, v := range configMap {
				config[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	return config
}

func getVars(params map[string]interface{}) map[string]string {
	vars := make(map[string]string)
	if varsParam, exists := params["vars"]; exists {
		if varsMap, ok := varsParam.(map[string]interface{}); ok {
			for k, v := range varsMap {
				vars[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	return vars
}

func main() {
	terraformPlugin := NewTerraformPlugin()
	sdk := pluginv2.NewSDK(terraformPlugin)
	sdk.Serve()
}