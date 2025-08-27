package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pluginv2 "github.com/corynth/corynth/pkg/plugin/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sPlugin implements Kubernetes operations via gRPC
type K8sPlugin struct {
	*pluginv2.BasePlugin
	client kubernetes.Interface
}

// NewK8sPlugin creates a new Kubernetes plugin
func NewK8sPlugin() *K8sPlugin {
	base := pluginv2.NewBuilder("k8s", "1.0.0").
		Description("Kubernetes cluster operations and resource management").
		Author("Corynth Team").
		Tags("k8s", "kubernetes", "containers", "orchestration").
		Action("deploy", "Deploy application to Kubernetes cluster").
		Input("name", "string", "Deployment name", true).
		Input("image", "string", "Container image", true).
		Input("namespace", "string", "Kubernetes namespace", true).
		InputWithDefault("replicas", "number", "Number of replicas", 1.0).
		InputWithDefault("port", "number", "Container port", 80.0).
		InputWithDefault("kubeconfig", "string", "Path to kubeconfig file", "").
		Output("deployment_name", "string", "Created deployment name").
		Output("namespace", "string", "Deployment namespace").
		Output("replicas", "number", "Number of replicas").
		Output("status", "string", "Deployment status").
		Add().
		Action("scale", "Scale deployment replicas").
		Input("name", "string", "Deployment name", true).
		Input("namespace", "string", "Kubernetes namespace", true).
		Input("replicas", "number", "Number of replicas", true).
		InputWithDefault("kubeconfig", "string", "Path to kubeconfig file", "").
		Output("deployment_name", "string", "Scaled deployment name").
		Output("replicas", "number", "New replica count").
		Output("status", "string", "Scale operation status").
		Add().
		Action("delete", "Delete Kubernetes resources").
		Input("name", "string", "Resource name", true).
		Input("namespace", "string", "Kubernetes namespace", true).
		Input("type", "string", "Resource type (deployment, service, pod)", true).
		InputWithDefault("kubeconfig", "string", "Path to kubeconfig file", "").
		Output("resource_name", "string", "Deleted resource name").
		Output("resource_type", "string", "Deleted resource type").
		Output("status", "string", "Deletion status").
		Add().
		Action("get", "Get Kubernetes resource information").
		Input("name", "string", "Resource name", true).
		Input("namespace", "string", "Kubernetes namespace", true).
		Input("type", "string", "Resource type (deployment, service, pod)", true).
		InputWithDefault("kubeconfig", "string", "Path to kubeconfig file", "").
		Output("resource_name", "string", "Resource name").
		Output("resource_type", "string", "Resource type").
		Output("status", "string", "Resource status").
		Output("details", "object", "Resource details").
		Add().
		Build()

	plugin := &K8sPlugin{
		BasePlugin: base,
	}

	return plugin
}

// Execute implements the plugin execution
func (p *K8sPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	// Initialize Kubernetes client
	if err := p.initClient(params); err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}

	switch action {
	case "deploy":
		return p.handleDeploy(ctx, params)
	case "scale":
		return p.handleScale(ctx, params)
	case "delete":
		return p.handleDelete(ctx, params)
	case "get":
		return p.handleGet(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Validate validates parameters
func (p *K8sPlugin) Validate(params map[string]interface{}) error {
	if err := p.BasePlugin.Validate(params); err != nil {
		return err
	}

	// Validate namespace
	if namespace, exists := params["namespace"]; exists {
		if ns, ok := namespace.(string); ok {
			if ns == "" {
				return fmt.Errorf("namespace cannot be empty")
			}
		}
	}

	// Validate resource type for get/delete actions
	if resourceType, exists := params["type"]; exists {
		if rt, ok := resourceType.(string); ok {
			validTypes := []string{"deployment", "service", "pod", "configmap", "secret"}
			valid := false
			for _, validType := range validTypes {
				if rt == validType {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid resource type: %s. Must be one of: %s", rt, strings.Join(validTypes, ", "))
			}
		}
	}

	return nil
}

// initClient initializes the Kubernetes client
func (p *K8sPlugin) initClient(params map[string]interface{}) error {
	if p.client != nil {
		return nil
	}

	var config *rest.Config
	var err error

	// Try to use provided kubeconfig
	if kubeconfigPath, exists := params["kubeconfig"]; exists {
		if path, ok := kubeconfigPath.(string); ok && path != "" {
			config, err = clientcmd.BuildConfigFromFlags("", path)
			if err != nil {
				return fmt.Errorf("failed to build config from kubeconfig: %w", err)
			}
		}
	}

	// If no kubeconfig provided, try default locations
	if config == nil {
		// Try in-cluster config first (for pods running in K8s)
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fall back to default kubeconfig location
			homeDir, _ := os.UserHomeDir()
			kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			if err != nil {
				return fmt.Errorf("failed to build Kubernetes config: %w", err)
			}
		}
	}

	// Create the client
	p.client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return nil
}

// handleDeploy creates a deployment and service
func (p *K8sPlugin) handleDeploy(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	name := params["name"].(string)
	image := params["image"].(string)
	namespace := params["namespace"].(string)
	replicas := int32(getNumberParam(params, "replicas", 1))
	port := int32(getNumberParam(params, "port", 80))

	// Create deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"managed-by": "corynth",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
								},
							},
						},
					},
				},
			},
		},
	}

	deploymentClient := p.client.AppsV1().Deployments(namespace)
	
	// Check if deployment already exists
	_, err := deploymentClient.Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		// Update existing deployment
		_, err = deploymentClient.Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to update deployment: %w", err)
		}
	} else if errors.IsNotFound(err) {
		// Create new deployment
		_, err = deploymentClient.Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create deployment: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to check deployment: %w", err)
	}

	// Create service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"managed-by": "corynth",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt(int(port)),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	serviceClient := p.client.CoreV1().Services(namespace)
	
	// Check if service already exists
	_, err = serviceClient.Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		// Update existing service
		_, err = serviceClient.Update(ctx, service, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to update service: %w", err)
		}
	} else if errors.IsNotFound(err) {
		// Create new service
		_, err = serviceClient.Create(ctx, service, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create service: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to check service: %w", err)
	}

	return map[string]interface{}{
		"deployment_name": name,
		"namespace":       namespace,
		"replicas":        float64(replicas),
		"status":          "deployed",
	}, nil
}

// handleScale scales a deployment
func (p *K8sPlugin) handleScale(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	name := params["name"].(string)
	namespace := params["namespace"].(string)
	replicas := int32(getNumberParam(params, "replicas", 1))

	deploymentClient := p.client.AppsV1().Deployments(namespace)
	
	// Get current deployment
	deployment, err := deploymentClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Update replicas
	deployment.Spec.Replicas = &replicas
	
	_, err = deploymentClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to scale deployment: %w", err)
	}

	return map[string]interface{}{
		"deployment_name": name,
		"replicas":        float64(replicas),
		"status":          "scaled",
	}, nil
}

// handleDelete deletes a resource
func (p *K8sPlugin) handleDelete(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	name := params["name"].(string)
	namespace := params["namespace"].(string)
	resourceType := params["type"].(string)

	var err error
	switch resourceType {
	case "deployment":
		err = p.client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "service":
		err = p.client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "pod":
		err = p.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "configmap":
		err = p.client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "secret":
		err = p.client.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to delete %s: %w", resourceType, err)
	}

	return map[string]interface{}{
		"resource_name": name,
		"resource_type": resourceType,
		"status":        "deleted",
	}, nil
}

// handleGet gets resource information
func (p *K8sPlugin) handleGet(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	name := params["name"].(string)
	namespace := params["namespace"].(string)
	resourceType := params["type"].(string)

	var details interface{}
	var status string
	var err error

	switch resourceType {
	case "deployment":
		deployment, err := p.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment: %w", err)
		}
		status = fmt.Sprintf("%d/%d ready", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
		details = map[string]interface{}{
			"replicas":       float64(*deployment.Spec.Replicas),
			"ready_replicas": float64(deployment.Status.ReadyReplicas),
			"image":          deployment.Spec.Template.Spec.Containers[0].Image,
			"creation_time":  deployment.CreationTimestamp.Format(time.RFC3339),
		}
	case "service":
		service, err := p.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get service: %w", err)
		}
		status = string(service.Spec.Type)
		ports := make([]map[string]interface{}, len(service.Spec.Ports))
		for i, port := range service.Spec.Ports {
			ports[i] = map[string]interface{}{
				"port":        float64(port.Port),
				"target_port": port.TargetPort.String(),
				"protocol":    string(port.Protocol),
			}
		}
		details = map[string]interface{}{
			"type":          string(service.Spec.Type),
			"cluster_ip":    service.Spec.ClusterIP,
			"ports":         ports,
			"creation_time": service.CreationTimestamp.Format(time.RFC3339),
		}
	case "pod":
		pod, err := p.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod: %w", err)
		}
		status = string(pod.Status.Phase)
		details = map[string]interface{}{
			"phase":         string(pod.Status.Phase),
			"node":          pod.Spec.NodeName,
			"pod_ip":        pod.Status.PodIP,
			"creation_time": pod.CreationTimestamp.Format(time.RFC3339),
		}
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"resource_name": name,
		"resource_type": resourceType,
		"status":        status,
		"details":       details,
	}, nil
}

// Helper functions
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

func main() {
	k8sPlugin := NewK8sPlugin()
	sdk := pluginv2.NewSDK(k8sPlugin)
	sdk.Serve()
}