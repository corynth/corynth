package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/corynth/corynth/internal/engine"
)

// KubernetesPlugin is a plugin for Kubernetes operations
type KubernetesPlugin struct{
	kubeconfig string
}

// Result represents the result of a step execution
type Result struct {
	Status    string
	Output    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// New creates a new KubernetesPlugin - this is the required export for Go plugins
func New() engine.Plugin {
	return NewKubernetesPlugin()
}

// NewKubernetesPlugin creates a new KubernetesPlugin
func NewKubernetesPlugin() *KubernetesPlugin {
	// Default to using the KUBECONFIG environment variable
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// Fall back to default location
		homeDir, err := os.UserHomeDir()
		if err == nil {
			kubeconfig = fmt.Sprintf("%s/.kube/config", homeDir)
		}
	}
	
	return &KubernetesPlugin{
		kubeconfig: kubeconfig,
	}
}

// Name returns the name of the plugin
func (p *KubernetesPlugin) Name() string {
	return "kubernetes"
}

// Execute executes a Kubernetes action
func (p *KubernetesPlugin) Execute(action string, params map[string]interface{}) (engine.Result, error) {
	startTime := time.Now()

	// Override kubeconfig if provided in params
	if kubeconfigParam, ok := params["kubeconfig"].(string); ok && kubeconfigParam != "" {
		p.kubeconfig = kubeconfigParam
	}

	switch action {
	case "apply":
		return p.apply(params)
	case "delete":
		return p.delete(params)
	case "get":
		return p.get(params)
	case "describe":
		return p.describe(params)
	case "exec":
		return p.exec(params)
	case "logs":
		return p.logs(params)
	case "port-forward":
		return p.portForward(params)
	default:
		endTime := time.Now()
		return engine.Result{
			Status:    "error",
			Error:     fmt.Sprintf("unknown action: %s", action),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("unknown action: %s", action)
	}
}

// apply applies Kubernetes resources
func (p *KubernetesPlugin) apply(params map[string]interface{}) (engine.Result, error) {
	startTime := time.Now()

	// Get parameters
	var filename string
	var manifest string
	var ok bool

	// Either filename or manifest must be provided
	if filename, ok = params["filename"].(string); !ok {
		if manifest, ok = params["manifest"].(string); !ok {
			endTime := time.Now()
			return engine.Result{
				Status:    "error",
				Error:     "either filename or manifest parameter is required",
				StartTime: startTime,
				EndTime:   endTime,
				Duration:  endTime.Sub(startTime),
			}, fmt.Errorf("either filename or manifest parameter is required")
		}
	}

	namespace, _ := params["namespace"].(string)
	wait, _ := params["wait"].(bool)

	// Build command
	args := []string{"apply"}

	if p.kubeconfig != "" {
		args = append(args, "--kubeconfig", p.kubeconfig)
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if filename != "" {
		args = append(args, "-f", filename)
	}

	if wait {
		args = append(args, "--wait")
	}

	var cmd *exec.Cmd
	var err error
	var output []byte

	if filename != "" {
		// Execute command with file
		cmd = exec.Command("kubectl", args...)
		output, err = cmd.CombinedOutput()
	} else {
		// Execute command with manifest from stdin
		cmd = exec.Command("kubectl", args...)
		cmd.Stdin = strings.NewReader(manifest)
		output, err = cmd.CombinedOutput()
	}

	endTime := time.Now()

	if err != nil {
		return engine.Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing kubectl apply: %w", err)
	}

	return engine.Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// delete deletes Kubernetes resources
func (p *KubernetesPlugin) delete(params map[string]interface{}) (engine.Result, error) {
	startTime := time.Now()

	// Get parameters
	var resource string
	var filename string
	var ok bool

	// Either resource or filename must be provided
	if resource, ok = params["resource"].(string); !ok {
		if filename, ok = params["filename"].(string); !ok {
			endTime := time.Now()
			return engine.Result{
				Status:    "error",
				Error:     "either resource or filename parameter is required",
				StartTime: startTime,
				EndTime:   endTime,
				Duration:  endTime.Sub(startTime),
			}, fmt.Errorf("either resource or filename parameter is required")
		}
	}

	namespace, _ := params["namespace"].(string)
	name, _ := params["name"].(string)
	wait, _ := params["wait"].(bool)

	// Build command
	args := []string{"delete"}

	if p.kubeconfig != "" {
		args = append(args, "--kubeconfig", p.kubeconfig)
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if resource != "" {
		args = append(args, resource)
		if name != "" {
			args = append(args, name)
		}
	}

	if filename != "" {
		args = append(args, "-f", filename)
	}

	if wait {
		args = append(args, "--wait")
	}

	// Execute command
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return engine.Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing kubectl delete: %w", err)
	}

	return engine.Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// get gets Kubernetes resources
func (p *KubernetesPlugin) get(params map[string]interface{}) (engine.Result, error) {
	startTime := time.Now()

	// Get parameters
	resource, ok := params["resource"].(string)
	if !ok {
		endTime := time.Now()
		return engine.Result{
			Status:    "error",
			Error:     "resource parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("resource parameter is required")
	}

	namespace, _ := params["namespace"].(string)
	name, _ := params["name"].(string)
	outputFormat, _ := params["output"].(string)
	selector, _ := params["selector"].(string)

	// Build command
	args := []string{"get", resource}

	if p.kubeconfig != "" {
		args = append(args, "--kubeconfig", p.kubeconfig)
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if name != "" {
		args = append(args, name)
	}

	if outputFormat != "" {
		args = append(args, "-o", outputFormat)
	}

	if selector != "" {
		args = append(args, "-l", selector)
	}

	// Execute command
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return engine.Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing kubectl get: %w", err)
	}

	return engine.Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// describe describes Kubernetes resources
func (p *KubernetesPlugin) describe(params map[string]interface{}) (engine.Result, error) {
	startTime := time.Now()

	// Get parameters
	resource, ok := params["resource"].(string)
	if !ok {
		endTime := time.Now()
		return engine.Result{
			Status:    "error",
			Error:     "resource parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("resource parameter is required")
	}

	namespace, _ := params["namespace"].(string)
	name, _ := params["name"].(string)
	selector, _ := params["selector"].(string)

	// Build command
	args := []string{"describe", resource}

	if p.kubeconfig != "" {
		args = append(args, "--kubeconfig", p.kubeconfig)
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if name != "" {
		args = append(args, name)
	}

	if selector != "" {
		args = append(args, "-l", selector)
	}

	// Execute command
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return engine.Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing kubectl describe: %w", err)
	}

	return engine.Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// exec executes a command in a container
func (p *KubernetesPlugin) exec(params map[string]interface{}) (engine.Result, error) {
	startTime := time.Now()

	// Get parameters
	pod, ok := params["pod"].(string)
	if !ok {
		endTime := time.Now()
		return engine.Result{
			Status:    "error",
			Error:     "pod parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("pod parameter is required")
	}

	command, ok := params["command"].(string)
	if !ok {
		endTime := time.Now()
		return engine.Result{
			Status:    "error",
			Error:     "command parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("command parameter is required")
	}

	namespace, _ := params["namespace"].(string)
	container, _ := params["container"].(string)

	// Build command
	args := []string{"exec"}

	if p.kubeconfig != "" {
		args = append(args, "--kubeconfig", p.kubeconfig)
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	args = append(args, pod)

	if container != "" {
		args = append(args, "-c", container)
	}

	args = append(args, "--")
	args = append(args, strings.Split(command, " ")...)

	// Execute command
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return engine.Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing kubectl exec: %w", err)
	}

	return engine.Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// logs gets logs from a container
func (p *KubernetesPlugin) logs(params map[string]interface{}) (engine.Result, error) {
	startTime := time.Now()

	// Get parameters
	pod, ok := params["pod"].(string)
	if !ok {
		endTime := time.Now()
		return engine.Result{
			Status:    "error",
			Error:     "pod parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("pod parameter is required")
	}

	namespace, _ := params["namespace"].(string)
	container, _ := params["container"].(string)
	tail, _ := params["tail"].(int)
	follow, _ := params["follow"].(bool)

	// Build command
	args := []string{"logs"}

	if p.kubeconfig != "" {
		args = append(args, "--kubeconfig", p.kubeconfig)
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if container != "" {
		args = append(args, "-c", container)
	}

	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}

	if follow {
		args = append(args, "-f")
	}

	args = append(args, pod)

	// Execute command
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return engine.Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing kubectl logs: %w", err)
	}

	return engine.Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// portForward forwards a port to a pod
func (p *KubernetesPlugin) portForward(params map[string]interface{}) (engine.Result, error) {
	startTime := time.Now()

	// Get parameters
	pod, ok := params["pod"].(string)
	if !ok {
		endTime := time.Now()
		return engine.Result{
			Status:    "error",
			Error:     "pod parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("pod parameter is required")
	}

	ports, ok := params["ports"].(string)
	if !ok {
		endTime := time.Now()
		return engine.Result{
			Status:    "error",
			Error:     "ports parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("ports parameter is required")
	}

	namespace, _ := params["namespace"].(string)
	address, _ := params["address"].(string)

	// Build command
	args := []string{"port-forward"}

	if p.kubeconfig != "" {
		args = append(args, "--kubeconfig", p.kubeconfig)
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if address != "" {
		args = append(args, "--address", address)
	}

	args = append(args, pod)
	args = append(args, strings.Split(ports, ",")...)

	// Execute command
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return engine.Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing kubectl port-forward: %w", err)
	}

	return engine.Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}