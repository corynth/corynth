package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	pluginv2 "github.com/corynth/corynth/pkg/plugin/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// DockerPlugin implements Docker operations via gRPC
type DockerPlugin struct {
	*pluginv2.BasePlugin
	client *client.Client
}

// NewDockerPlugin creates a new Docker plugin
func NewDockerPlugin() *DockerPlugin {
	base := pluginv2.NewBuilder("docker", "1.0.0").
		Description("Docker container and image management operations").
		Author("Corynth Team").
		Tags("docker", "containers", "images", "orchestration").
		Action("build", "Build Docker image from Dockerfile").
		Input("dockerfile", "string", "Dockerfile content or path", true).
		Input("tag", "string", "Image tag", true).
		InputWithDefault("context", "string", "Build context path", ".").
		InputWithDefault("build_args", "object", "Build arguments", map[string]string{}).
		Output("image_id", "string", "Built image ID").
		Output("tag", "string", "Image tag").
		Output("status", "string", "Build status").
		Add().
		Action("run", "Run Docker container").
		Input("image", "string", "Docker image", true).
		InputWithDefault("name", "string", "Container name", "").
		InputWithDefault("ports", "object", "Port mappings", map[string]string{}).
		InputWithDefault("env", "object", "Environment variables", map[string]string{}).
		InputWithDefault("volumes", "object", "Volume mounts", map[string]string{}).
		InputWithDefault("command", "string", "Command to run", "").
		InputWithDefault("detach", "boolean", "Run in background", true).
		Output("container_id", "string", "Container ID").
		Output("container_name", "string", "Container name").
		Output("status", "string", "Container status").
		Add().
		Action("stop", "Stop running container").
		Input("container", "string", "Container name or ID", true).
		InputWithDefault("timeout", "number", "Stop timeout in seconds", 30.0).
		Output("container_id", "string", "Stopped container ID").
		Output("status", "string", "Stop operation status").
		Add().
		Action("remove", "Remove container or image").
		Input("target", "string", "Container/image name or ID", true).
		Input("type", "string", "Type: container or image", true).
		InputWithDefault("force", "boolean", "Force removal", false).
		Output("target", "string", "Removed target").
		Output("type", "string", "Removed type").
		Output("status", "string", "Removal status").
		Add().
		Action("ps", "List containers").
		InputWithDefault("all", "boolean", "Show all containers", false).
		Output("containers", "array", "List of containers").
		Output("count", "number", "Number of containers").
		Add().
		Action("images", "List images").
		InputWithDefault("all", "boolean", "Show all images", false).
		Output("images", "array", "List of images").
		Output("count", "number", "Number of images").
		Add().
		Action("logs", "Get container logs").
		Input("container", "string", "Container name or ID", true).
		InputWithDefault("tail", "number", "Number of lines to show", 100.0).
		InputWithDefault("follow", "boolean", "Follow log output", false).
		Output("logs", "string", "Container logs").
		Output("container_id", "string", "Container ID").
		Add().
		Build()

	plugin := &DockerPlugin{
		BasePlugin: base,
	}

	return plugin
}

// Execute implements the plugin execution
func (p *DockerPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	// Initialize Docker client
	if err := p.initClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize Docker client: %w", err)
	}

	switch action {
	case "build":
		return p.handleBuild(ctx, params)
	case "run":
		return p.handleRun(ctx, params)
	case "stop":
		return p.handleStop(ctx, params)
	case "remove":
		return p.handleRemove(ctx, params)
	case "ps":
		return p.handlePS(ctx, params)
	case "images":
		return p.handleImages(ctx, params)
	case "logs":
		return p.handleLogs(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Validate validates parameters
func (p *DockerPlugin) Validate(params map[string]interface{}) error {
	if err := p.BasePlugin.Validate(params); err != nil {
		return err
	}

	// Validate type parameter for remove action
	if actionType, exists := params["type"]; exists {
		if t, ok := actionType.(string); ok {
			if t != "container" && t != "image" {
				return fmt.Errorf("type must be 'container' or 'image', got: %s", t)
			}
		}
	}

	return nil
}

// initClient initializes the Docker client
func (p *DockerPlugin) initClient() error {
	if p.client != nil {
		return nil
	}

	var err error
	p.client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	return nil
}

// handleBuild builds a Docker image
func (p *DockerPlugin) handleBuild(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	dockerfile := params["dockerfile"].(string)
	tag := params["tag"].(string)
	buildArgs := getBuildArgs(params)

	// Create tar archive for build context
	tarBuffer := new(bytes.Buffer)
	tarWriter := tar.NewWriter(tarBuffer)

	// Add Dockerfile to tar
	header := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfile)),
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile header: %w", err)
	}
	if _, err := tarWriter.Write([]byte(dockerfile)); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile: %w", err)
	}
	
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Build options
	buildOptions := types.ImageBuildOptions{
		Tags:      []string{tag},
		Remove:    true,
		BuildArgs: buildArgs,
	}

	// Build image
	response, err := p.client.ImageBuild(ctx, bytes.NewReader(tarBuffer.Bytes()), buildOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}
	defer response.Body.Close()

	// Read build output
	scanner := bufio.NewScanner(response.Body)
	var imageID string
	for scanner.Scan() {
		var buildResult map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &buildResult); err == nil {
			if stream, ok := buildResult["stream"].(string); ok {
				if strings.Contains(stream, "Successfully built") {
					parts := strings.Fields(stream)
					if len(parts) >= 3 {
						imageID = parts[2]
					}
				}
			}
		}
	}

	if imageID == "" {
		imageID = tag // fallback to tag if ID not found
	}

	return map[string]interface{}{
		"image_id": imageID,
		"tag":      tag,
		"status":   "built",
	}, nil
}

// handleRun runs a Docker container
func (p *DockerPlugin) handleRun(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	imageName := params["image"].(string)
	containerName := getStringParam(params, "name", "")
	ports := getPortMappings(params)
	env := getEnvironmentVars(params)
	volumes := getVolumeMounts(params)
	command := getStringParam(params, "command", "")
	detach := getBoolParam(params, "detach", true)

	// Container configuration
	config := &container.Config{
		Image: imageName,
		Env:   env,
	}

	if command != "" {
		config.Cmd = strings.Fields(command)
	}

	// Host configuration
	hostConfig := &container.HostConfig{
		PortBindings: ports,
		Mounts:       volumes,
	}

	// Create container
	resp, err := p.client.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := p.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	status := "running"
	if !detach {
		// Wait for container to finish
		statusCh, errCh := p.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				status = "error"
			}
		case <-statusCh:
			status = "completed"
		}
	}

	return map[string]interface{}{
		"container_id":   resp.ID,
		"container_name": containerName,
		"status":         status,
	}, nil
}

// handleStop stops a container
func (p *DockerPlugin) handleStop(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	containerID := params["container"].(string)
	timeout := int(getNumberParam(params, "timeout", 30))

	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}

	if err := p.client.ContainerStop(ctx, containerID, stopOptions); err != nil {
		return nil, fmt.Errorf("failed to stop container: %w", err)
	}

	return map[string]interface{}{
		"container_id": containerID,
		"status":       "stopped",
	}, nil
}

// handleRemove removes a container or image
func (p *DockerPlugin) handleRemove(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	target := params["target"].(string)
	resourceType := params["type"].(string)
	force := getBoolParam(params, "force", false)

	var err error
	switch resourceType {
	case "container":
		removeOptions := types.ContainerRemoveOptions{
			Force: force,
		}
		err = p.client.ContainerRemove(ctx, target, removeOptions)
	case "image":
		removeOptions := types.ImageRemoveOptions{
			Force: force,
		}
		_, err = p.client.ImageRemove(ctx, target, removeOptions)
	default:
		return nil, fmt.Errorf("unsupported type: %s", resourceType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to remove %s: %w", resourceType, err)
	}

	return map[string]interface{}{
		"target": target,
		"type":   resourceType,
		"status": "removed",
	}, nil
}

// handlePS lists containers
func (p *DockerPlugin) handlePS(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	showAll := getBoolParam(params, "all", false)

	containers, err := p.client.ContainerList(ctx, types.ContainerListOptions{All: showAll})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var containerList []map[string]interface{}
	for _, c := range containers {
		containerList = append(containerList, map[string]interface{}{
			"id":      c.ID[:12],
			"image":   c.Image,
			"command": c.Command,
			"created": time.Unix(c.Created, 0).Format(time.RFC3339),
			"status":  c.Status,
			"ports":   formatPorts(c.Ports),
			"names":   c.Names,
		})
	}

	return map[string]interface{}{
		"containers": containerList,
		"count":      float64(len(containers)),
	}, nil
}

// handleImages lists images
func (p *DockerPlugin) handleImages(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	showAll := getBoolParam(params, "all", false)

	images, err := p.client.ImageList(ctx, types.ImageListOptions{All: showAll})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	var imageList []map[string]interface{}
	for _, img := range images {
		imageList = append(imageList, map[string]interface{}{
			"id":         img.ID[:12],
			"repository": strings.Join(img.RepoTags, ", "),
			"tag":        "latest", // simplified
			"created":    time.Unix(img.Created, 0).Format(time.RFC3339),
			"size":       float64(img.Size),
		})
	}

	return map[string]interface{}{
		"images": imageList,
		"count":  float64(len(images)),
	}, nil
}

// handleLogs gets container logs
func (p *DockerPlugin) handleLogs(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	containerID := params["container"].(string)
	tail := fmt.Sprintf("%.0f", getNumberParam(params, "tail", 100))

	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	}

	reader, err := p.client.ContainerLogs(ctx, containerID, logOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	return map[string]interface{}{
		"logs":         string(logs),
		"container_id": containerID,
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

func getBuildArgs(params map[string]interface{}) map[string]*string {
	buildArgs := make(map[string]*string)
	if args, exists := params["build_args"]; exists {
		if argsMap, ok := args.(map[string]interface{}); ok {
			for k, v := range argsMap {
				strVal := fmt.Sprintf("%v", v)
				buildArgs[k] = &strVal
			}
		}
	}
	return buildArgs
}

func getPortMappings(params map[string]interface{}) nat.PortMap {
	portMap := make(nat.PortMap)
	if ports, exists := params["ports"]; exists {
		if portsMap, ok := ports.(map[string]interface{}); ok {
			for containerPort, hostPort := range portsMap {
				port, err := nat.NewPort("tcp", containerPort)
				if err == nil {
					portMap[port] = []nat.PortBinding{
						{
							HostPort: fmt.Sprintf("%v", hostPort),
						},
					}
				}
			}
		}
	}
	return portMap
}

func getEnvironmentVars(params map[string]interface{}) []string {
	var env []string
	if envVars, exists := params["env"]; exists {
		if envMap, ok := envVars.(map[string]interface{}); ok {
			for k, v := range envMap {
				env = append(env, fmt.Sprintf("%s=%v", k, v))
			}
		}
	}
	return env
}

func getVolumeMounts(params map[string]interface{}) []mount.Mount {
	var mounts []mount.Mount
	if volumes, exists := params["volumes"]; exists {
		if volumesMap, ok := volumes.(map[string]interface{}); ok {
			for source, target := range volumesMap {
				mounts = append(mounts, mount.Mount{
					Type:   mount.TypeBind,
					Source: source,
					Target: fmt.Sprintf("%v", target),
				})
			}
		}
	}
	return mounts
}

func formatPorts(ports []types.Port) []string {
	var portStrings []string
	for _, port := range ports {
		if port.PublicPort > 0 {
			portStrings = append(portStrings, fmt.Sprintf("%d->%d/%s", port.PublicPort, port.PrivatePort, port.Type))
		}
	}
	return portStrings
}

func main() {
	dockerPlugin := NewDockerPlugin()
	sdk := pluginv2.NewSDK(dockerPlugin)
	sdk.Serve()
}