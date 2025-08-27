package pluginv2

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
	"github.com/corynth/corynth/pkg/plugin"
	pb "github.com/corynth/corynth/pkg/plugin/proto"
)

// GRPCClientPlugin wraps a gRPC connection to a remote plugin
type GRPCClientPlugin struct {
	name     string
	conn     *grpc.ClientConn
	client   pb.PluginServiceClient
	cmd      *exec.Cmd
	metadata plugin.Metadata
	actions  []plugin.Action
}

// NewGRPCClientPlugin creates a new gRPC client plugin
func NewGRPCClientPlugin(name, executablePath string) (*GRPCClientPlugin, error) {
	// Start the plugin process with "serve" argument
	cmd := exec.Command(executablePath, "serve")
	
	// Get stdout to read handshake
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	
	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start plugin process: %w", err)
	}
	
	// Read handshake from stdout (Terraform pattern: "1|1|tcp|127.0.0.1:port|grpc")
	handshakeBuf := make([]byte, 256)
	n, err := stdout.Read(handshakeBuf)
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to read handshake: %w", err)
	}
	
	handshake := strings.TrimSpace(string(handshakeBuf[:n]))
	parts := strings.Split(handshake, "|")
	if len(parts) < 5 {
		cmd.Process.Kill()
		return nil, fmt.Errorf("invalid handshake format: %s", handshake)
	}
	
	// Extract connection info
	address := parts[3] // 127.0.0.1:port
	protocol := parts[4] // grpc
	
	if protocol != "grpc" {
		cmd.Process.Kill()
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
	
	// Connect to plugin via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to connect to plugin: %w", err)
	}
	
	client := pb.NewPluginServiceClient(conn)
	
	// Test connection with health check
	healthResp, err := client.Health(ctx, &pb.HealthRequest{})
	if err != nil || !healthResp.Healthy {
		conn.Close()
		cmd.Process.Kill()
		return nil, fmt.Errorf("plugin health check failed: %w", err)
	}
	
	grpcPlugin := &GRPCClientPlugin{
		name:   name,
		conn:   conn,
		client: client,
		cmd:    cmd,
	}
	
	// Load metadata and actions
	if err := grpcPlugin.loadMetadata(); err != nil {
		grpcPlugin.Close()
		return nil, fmt.Errorf("failed to load plugin metadata: %w", err)
	}
	
	if err := grpcPlugin.loadActions(); err != nil {
		grpcPlugin.Close()
		return nil, fmt.Errorf("failed to load plugin actions: %w", err)
	}
	
	return grpcPlugin, nil
}

// loadMetadata fetches and caches plugin metadata
func (p *GRPCClientPlugin) loadMetadata() error {
	resp, err := p.client.GetMetadata(context.Background(), &pb.MetadataRequest{})
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}
	
	p.metadata = plugin.Metadata{
		Name:        resp.Name,
		Version:     resp.Version,
		Description: resp.Description,
		Author:      resp.Author,
		Tags:        resp.Tags,
	}
	
	return nil
}

// loadActions fetches and caches plugin actions
func (p *GRPCClientPlugin) loadActions() error {
	resp, err := p.client.GetActions(context.Background(), &pb.ActionsRequest{})
	if err != nil {
		return fmt.Errorf("failed to get actions: %w", err)
	}
	
	var actions []plugin.Action
	for _, pbAction := range resp.Actions {
		action := plugin.Action{
			Name:        pbAction.Name,
			Description: pbAction.Description,
			Inputs:      make(map[string]plugin.InputSpec),
			Outputs:     make(map[string]plugin.OutputSpec),
		}
		
		// Convert input specs
		for name, spec := range pbAction.Inputs {
			action.Inputs[name] = plugin.InputSpec{
				Type:        spec.Type,
				Description: spec.Description,
				Required:    spec.Required,
				Default:     convertFromProtoValue(spec.DefaultValue),
			}
		}
		
		// Convert output specs
		for name, spec := range pbAction.Outputs {
			action.Outputs[name] = plugin.OutputSpec{
				Type:        spec.Type,
				Description: spec.Description,
			}
		}
		
		actions = append(actions, action)
	}
	
	p.actions = actions
	return nil
}

// Metadata returns cached plugin metadata
func (p *GRPCClientPlugin) Metadata() plugin.Metadata {
	return p.metadata
}

// Actions returns cached plugin actions
func (p *GRPCClientPlugin) Actions() []plugin.Action {
	return p.actions
}

// Validate validates parameters via gRPC
func (p *GRPCClientPlugin) Validate(params map[string]interface{}) error {
	pbParams := make(map[string]*pb.Value)
	for k, v := range params {
		pbParams[k] = convertToProtoValue(v)
	}
	
	resp, err := p.client.ValidateParams(context.Background(), &pb.ValidateRequest{
		Params: pbParams,
	})
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	if !resp.Valid {
		return fmt.Errorf("validation errors: %s", strings.Join(resp.Errors, ", "))
	}
	
	return nil
}

// Execute executes an action via gRPC
func (p *GRPCClientPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	pbParams := make(map[string]*pb.Value)
	for k, v := range params {
		pbParams[k] = convertToProtoValue(v)
	}
	
	resp, err := p.client.Execute(ctx, &pb.ExecuteRequest{
		Action: action,
		Params: pbParams,
	})
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	
	if !resp.Success {
		return nil, fmt.Errorf("plugin execution failed: %s", resp.Error)
	}
	
	// Convert outputs back
	outputs := make(map[string]interface{})
	for k, v := range resp.Outputs {
		outputs[k] = convertFromProtoValue(v)
	}
	
	return outputs, nil
}

// Close closes the gRPC connection and terminates the plugin process
func (p *GRPCClientPlugin) Close() error {
	var errs []error
	
	// Close gRPC connection
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close gRPC connection: %w", err))
		}
	}
	
	// Terminate plugin process
	if p.cmd != nil && p.cmd.Process != nil {
		if err := p.cmd.Process.Kill(); err != nil {
			errs = append(errs, fmt.Errorf("failed to kill plugin process: %w", err))
		}
		p.cmd.Wait() // Wait for process cleanup
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("errors closing plugin: %v", errs)
	}
	
	return nil
}

// Helper functions for value conversion (same as in grpc_server.go)
func convertToProtoValue(v interface{}) *pb.Value {
	if v == nil {
		return &pb.Value{
			Kind: &pb.Value_NullValue{},
		}
	}
	
	switch val := v.(type) {
	case string:
		return &pb.Value{
			Kind: &pb.Value_StringValue{StringValue: val},
		}
	case float64:
		return &pb.Value{
			Kind: &pb.Value_NumberValue{NumberValue: val},
		}
	case int:
		return &pb.Value{
			Kind: &pb.Value_NumberValue{NumberValue: float64(val)},
		}
	case bool:
		return &pb.Value{
			Kind: &pb.Value_BoolValue{BoolValue: val},
		}
	case []interface{}:
		values := make([]*pb.Value, len(val))
		for i, item := range val {
			values[i] = convertToProtoValue(item)
		}
		return &pb.Value{
			Kind: &pb.Value_ArrayValue{
				ArrayValue: &pb.ValueArray{Values: values},
			},
		}
	case map[string]interface{}:
		fields := make(map[string]*pb.Value)
		for k, item := range val {
			fields[k] = convertToProtoValue(item)
		}
		return &pb.Value{
			Kind: &pb.Value_ObjectValue{
				ObjectValue: &pb.ValueObject{Fields: fields},
			},
		}
	default:
		// Fallback to string representation
		return &pb.Value{
			Kind: &pb.Value_StringValue{StringValue: fmt.Sprintf("%v", val)},
		}
	}
}

func convertFromProtoValue(v *pb.Value) interface{} {
	if v == nil {
		return nil
	}
	
	switch kind := v.Kind.(type) {
	case *pb.Value_StringValue:
		return kind.StringValue
	case *pb.Value_NumberValue:
		return kind.NumberValue
	case *pb.Value_BoolValue:
		return kind.BoolValue
	case *pb.Value_ArrayValue:
		values := make([]interface{}, len(kind.ArrayValue.Values))
		for i, val := range kind.ArrayValue.Values {
			values[i] = convertFromProtoValue(val)
		}
		return values
	case *pb.Value_ObjectValue:
		fields := make(map[string]interface{})
		for k, val := range kind.ObjectValue.Fields {
			fields[k] = convertFromProtoValue(val)
		}
		return fields
	case *pb.Value_NullValue:
		return nil
	default:
		return nil
	}
}