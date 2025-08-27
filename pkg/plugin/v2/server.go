package pluginv2

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	
	"github.com/corynth/corynth/pkg/plugin"
	pb "github.com/corynth/corynth/pkg/plugin/proto"
)

// GRPCPluginServer implements the gRPC server for plugins
type GRPCPluginServer struct {
	pb.UnimplementedPluginServiceServer
	plugin plugin.Plugin
}

// NewGRPCPluginServer creates a new gRPC server for a plugin
func NewGRPCPluginServer(plugin plugin.Plugin) *GRPCPluginServer {
	return &GRPCPluginServer{
		plugin: plugin,
	}
}

// GetMetadata implements the gRPC metadata method
func (s *GRPCPluginServer) GetMetadata(ctx context.Context, req *pb.MetadataRequest) (*pb.MetadataResponse, error) {
	meta := s.plugin.Metadata()
	
	return &pb.MetadataResponse{
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Author:      meta.Author,
		Tags:        meta.Tags,
	}, nil
}

// GetActions implements the gRPC actions method
func (s *GRPCPluginServer) GetActions(ctx context.Context, req *pb.ActionsRequest) (*pb.ActionsResponse, error) {
	actions := s.plugin.Actions()
	
	var pbActions []*pb.Action
	for _, action := range actions {
		pbAction := &pb.Action{
			Name:        action.Name,
			Description: action.Description,
			Inputs:      make(map[string]*pb.InputSpec),
			Outputs:     make(map[string]*pb.OutputSpec),
		}
		
		// Convert input specs
		for name, spec := range action.Inputs {
			pbAction.Inputs[name] = &pb.InputSpec{
				Type:        spec.Type,
				Description: spec.Description,
				Required:    spec.Required,
				DefaultValue: convertToProtoValue(spec.Default),
			}
		}
		
		// Convert output specs
		for name, spec := range action.Outputs {
			pbAction.Outputs[name] = &pb.OutputSpec{
				Type:        spec.Type,
				Description: spec.Description,
			}
		}
		
		pbActions = append(pbActions, pbAction)
	}
	
	return &pb.ActionsResponse{
		Actions: pbActions,
	}, nil
}

// ValidateParams implements the gRPC validation method
func (s *GRPCPluginServer) ValidateParams(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	params := convertFromProtoParams(req.Params)
	
	err := s.plugin.Validate(params)
	if err != nil {
		return &pb.ValidateResponse{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}
	
	return &pb.ValidateResponse{
		Valid:  true,
		Errors: []string{},
	}, nil
}

// Execute implements the gRPC execution method
func (s *GRPCPluginServer) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	params := convertFromProtoParams(req.Params)
	
	// Execute the plugin
	outputs, err := s.plugin.Execute(ctx, req.Action, params)
	if err != nil {
		return &pb.ExecuteResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	// Convert outputs to proto values
	pbOutputs := make(map[string]*pb.Value)
	for k, v := range outputs {
		pbOutputs[k] = convertToProtoValue(v)
	}
	
	return &pb.ExecuteResponse{
		Success: true,
		Outputs: pbOutputs,
	}, nil
}

// Health implements the health check method
func (s *GRPCPluginServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Healthy: true,
		Status:  "OK",
	}, nil
}

// ServePlugin starts a gRPC server for the plugin following Terraform's pattern
func ServePlugin(plugin plugin.Plugin) error {
	// Create listener on available port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	
	// Create gRPC server
	server := grpc.NewServer()
	
	// Register plugin service
	pluginServer := NewGRPCPluginServer(plugin)
	pb.RegisterPluginServiceServer(server, pluginServer)
	
	// Print handshake info to stdout (Terraform pattern)
	addr := lis.Addr().(*net.TCPAddr)
	fmt.Printf("1|1|tcp|127.0.0.1:%d|grpc\n", addr.Port)
	
	// Serve requests
	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	
	return nil
}

func convertFromProtoParams(params map[string]*pb.Value) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range params {
		result[k] = convertFromProtoValue(v)
	}
	return result
}