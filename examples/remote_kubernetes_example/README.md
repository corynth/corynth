# Remote Kubernetes Plugin Example

This example demonstrates how to use the Kubernetes plugin as a remote plugin that gets downloaded automatically when needed.

## Prerequisites

- Corynth installed
- Kubernetes cluster configured and accessible via kubectl
- Internet connection to download the plugin

## Setup

1. Create a new Corynth project:

```bash
corynth init remote-k8s-example
cd remote-k8s-example
```

2. Copy the plugins.yaml file to your project:

```bash
cp plugins.yaml .
```

3. Copy the kubernetes_flow.yaml file to your project's flows directory:

```bash
cp kubernetes_flow.yaml flows/
```

## Running the Example

1. Plan the deployment:

```bash
corynth plan
```

During the planning phase, Corynth will detect that the Kubernetes plugin is needed and will download it automatically.

2. Apply the deployment:

```bash
corynth apply
```

This will execute the flow, which will create a namespace, deploy an application, create a service, and then clean up the resources.

## How It Works

The `plugins.yaml` file specifies the Kubernetes plugin as a remote plugin:

```yaml
plugins:
  - name: "kubernetes"
    repository: "https://github.com/corynth/plugins"
    version: "v1.2.0"
    path: "kubernetes"
```

When Corynth encounters a step that uses the "kubernetes" plugin, it will:

1. Check if the plugin is already loaded
2. If not, check if it's already downloaded
3. If not, download it from the specified repository
4. Load the plugin and execute the requested action

This allows you to use plugins without having to install them manually, making it easier to share and distribute workflows.