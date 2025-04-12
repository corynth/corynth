# Kubernetes Plugin for Corynth

This plugin provides Kubernetes integration for Corynth, allowing you to manage Kubernetes resources as part of your Corynth flows.

## Features

- Apply Kubernetes manifests from files or inline YAML
- Delete Kubernetes resources
- Get information about Kubernetes resources
- Describe Kubernetes resources in detail
- Execute commands in containers
- Retrieve logs from containers
- Set up port forwarding to pods

## Prerequisites

- `kubectl` command-line tool installed and available in the PATH
- Kubernetes configuration file (kubeconfig) properly set up

## Installation

The Kubernetes plugin is a remote plugin that will be automatically downloaded by Corynth when needed. To use it, add the following to your `plugins.yaml` file:

```yaml
plugins:
  - name: "kubernetes"
    repository: "https://github.com/corynth/plugins"
    version: "v1.2.0"
    path: "kubernetes"
```

## Building from Source

To build the plugin from source:

```bash
cd plugins/remote/kubernetes
make linux  # For Linux (optimized for Mac M4)
```

This will create a `build/kubernetes.so` file that can be used as a Corynth plugin.

## Usage

### Apply a Kubernetes Manifest

```yaml
flow:
  name: "deploy_application"
  description: "Deploy an application to Kubernetes"
  steps:
    - name: "deploy_app"
      plugin: "kubernetes"
      action: "apply"
      params:
        filename: "./manifests/deployment.yaml"
        namespace: "default"
        wait: true
```

Or with an inline manifest:

```yaml
flow:
  name: "deploy_application"
  description: "Deploy an application to Kubernetes"
  steps:
    - name: "deploy_app"
      plugin: "kubernetes"
      action: "apply"
      params:
        manifest: |
          apiVersion: apps/v1
          kind: Deployment
          metadata:
            name: my-app
            namespace: default
          spec:
            replicas: 3
            selector:
              matchLabels:
                app: my-app
            template:
              metadata:
                labels:
                  app: my-app
              spec:
                containers:
                - name: my-app
                  image: my-app:latest
                  ports:
                  - containerPort: 8080
        namespace: "default"
        wait: true
```

### Delete Kubernetes Resources

```yaml
flow:
  name: "cleanup_resources"
  description: "Clean up Kubernetes resources"
  steps:
    - name: "delete_deployment"
      plugin: "kubernetes"
      action: "delete"
      params:
        resource: "deployment"
        name: "my-app"
        namespace: "default"
        wait: true
```

### Get Kubernetes Resources

```yaml
flow:
  name: "check_resources"
  description: "Check Kubernetes resources"
  steps:
    - name: "get_pods"
      plugin: "kubernetes"
      action: "get"
      params:
        resource: "pods"
        namespace: "default"
        selector: "app=my-app"
        output: "wide"
```

### Describe Kubernetes Resources

```yaml
flow:
  name: "diagnose_resources"
  description: "Diagnose Kubernetes resources"
  steps:
    - name: "describe_pod"
      plugin: "kubernetes"
      action: "describe"
      params:
        resource: "pod"
        name: "my-app-pod"
        namespace: "default"
```

### Execute Commands in Containers

```yaml
flow:
  name: "run_commands"
  description: "Run commands in containers"
  steps:
    - name: "check_version"
      plugin: "kubernetes"
      action: "exec"
      params:
        pod: "my-app-pod"
        namespace: "default"
        container: "my-app"
        command: "cat /etc/os-release"
```

### Get Container Logs

```yaml
flow:
  name: "check_logs"
  description: "Check container logs"
  steps:
    - name: "get_logs"
      plugin: "kubernetes"
      action: "logs"
      params:
        pod: "my-app-pod"
        namespace: "default"
        container: "my-app"
        tail: 100
```

### Port Forward to a Pod

```yaml
flow:
  name: "access_app"
  description: "Access application via port forwarding"
  steps:
    - name: "port_forward"
      plugin: "kubernetes"
      action: "port-forward"
      params:
        pod: "my-app-pod"
        namespace: "default"
        ports: "8080:8080"
```

## Parameters

### Common Parameters

- `kubeconfig`: Path to the kubeconfig file (optional, defaults to environment variable or ~/.kube/config)
- `namespace`: Kubernetes namespace (optional, defaults to current context)

### Action-Specific Parameters

#### apply

- `filename`: Path to the Kubernetes manifest file (required if manifest not provided)
- `manifest`: Inline Kubernetes manifest (required if filename not provided)
- `wait`: Whether to wait for resources to be ready (optional, defaults to false)

#### delete

- `resource`: Resource type to delete (required if filename not provided)
- `name`: Resource name to delete (optional, required if resource provided)
- `filename`: Path to the Kubernetes manifest file (required if resource not provided)
- `wait`: Whether to wait for deletion to complete (optional, defaults to false)

#### get

- `resource`: Resource type to get (required)
- `name`: Resource name to get (optional)
- `output`: Output format (optional, e.g., "json", "yaml", "wide")
- `selector`: Label selector (optional)

#### describe

- `resource`: Resource type to describe (required)
- `name`: Resource name to describe (optional)
- `selector`: Label selector (optional)

#### exec

- `pod`: Pod name (required)
- `command`: Command to execute (required)
- `container`: Container name (optional, defaults to first container)

#### logs

- `pod`: Pod name (required)
- `container`: Container name (optional, defaults to first container)
- `tail`: Number of lines to show (optional)
- `follow`: Whether to follow the logs (optional, defaults to false)

#### port-forward

- `pod`: Pod name (required)
- `ports`: Port mapping (required, format: "local:remote" or "local")
- `address`: Address to bind to (optional, defaults to localhost)

## License

MIT