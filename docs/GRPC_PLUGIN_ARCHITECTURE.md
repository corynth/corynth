# Corynth gRPC Plugin Architecture

Corynth has adopted **Terraform's proven gRPC plugin architecture**, eliminating version compatibility issues while enabling cross-platform distribution and multi-language plugin development.

## Architecture Overview

### Core Components

**Corynth Core:**
- Statically compiled Go binary
- Acts as gRPC client
- Manages workflow orchestration
- Handles plugin lifecycle

**Plugins:**
- Independent executable processes
- gRPC servers following Terraform's protocol
- Language agnostic (Go, Python, Node.js, etc.)
- Cross-platform compatible

### Communication Protocol

```
Corynth Core ←→ gRPC over TCP ←→ Plugin Process
     Client                         Server
```

**Handshake Protocol (Terraform Compatible):**
```
Plugin stdout: "1|1|tcp|127.0.0.1:PORT|grpc"
Core connects to: 127.0.0.1:PORT
Protocol: gRPC over TCP
```

## Key Benefits

### ✅ **Eliminates Go Plugin Issues**
- **No Version Coupling**: Plugins don't need exact Go version match
- **Cross-Platform**: Same binary works on all platforms
- **Process Isolation**: Plugin crashes don't affect core
- **Language Freedom**: Write plugins in any language

### ✅ **Modern Distribution**
- **Registry Support**: Centralized plugin distribution
- **Semantic Versioning**: Proper dependency management  
- **Auto-Installation**: Automatic download and compilation
- **Platform Binaries**: Native executables for each OS

### ✅ **Developer Experience**
- **Simple SDK**: Fluent builder pattern for plugin creation
- **Type Safety**: Full gRPC type system
- **Hot Reload**: Plugin updates without core restart
- **Rich Ecosystem**: Support for any language/runtime

## Plugin Development

### 1. Using the SDK (Go)

```go
package main

import (
    "context"
    
    pluginv2 "github.com/corynth/corynth/pkg/plugin/v2"
)

type MyPlugin struct {
    *pluginv2.BasePlugin
}

func NewMyPlugin() *MyPlugin {
    base := pluginv2.NewBuilder("myplugin", "1.0.0").
        Description("My awesome plugin").
        Author("Your Name").
        Tags("category", "feature").
        Action("myaction", "Does something useful").
        Input("input_param", "string", "Input parameter", true).
        Output("result", "string", "Action result").
        Add().
        Build()

    return &MyPlugin{BasePlugin: base}
}

func (p *MyPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
    switch action {
    case "myaction":
        return map[string]interface{}{
            "result": fmt.Sprintf("Processed: %s", params["input_param"]),
        }, nil
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}

func main() {
    plugin := NewMyPlugin()
    sdk := pluginv2.NewSDK(plugin)
    sdk.Serve()
}
```

### 2. Building and Testing

```bash
# Build plugin
go build -o corynth-plugin-myplugin main.go

# Test plugin server
./corynth-plugin-myplugin serve &

# Should output handshake: "1|1|tcp|127.0.0.1:PORT|grpc"
```

### 3. Installation

```bash
# Copy to plugins directory
cp corynth-plugin-myplugin ~/.corynth/plugins/

# Plugin automatically available in workflows
```

## Protocol Definition

### gRPC Service Interface

```protobuf
service PluginService {
  rpc GetMetadata(MetadataRequest) returns (MetadataResponse);
  rpc GetActions(ActionsRequest) returns (ActionsResponse);
  rpc ValidateParams(ValidateRequest) returns (ValidateResponse);
  rpc Execute(ExecuteRequest) returns (ExecuteResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

### Value System

Universal value type supporting JSON-like data:
- `string_value` - Text data
- `number_value` - Numeric data (float64)
- `bool_value` - Boolean data
- `array_value` - Arrays/lists
- `object_value` - Maps/objects
- `null_value` - Null values

## Plugin Lifecycle

### 1. Discovery
```
Workflow requests plugin → Check local cache → Download if missing
```

### 2. Loading
```
Start plugin process → Read handshake → Connect via gRPC → Load metadata
```

### 3. Execution
```
Validate params → Execute action → Return results → Cache connection
```

### 4. Cleanup
```
Close gRPC connection → Terminate process → Clean resources
```

## Multi-Language Support

### Go Plugins
```go
// Use the official SDK
import pluginv2 "github.com/corynth/corynth/pkg/plugin/v2"
```

### Python Plugins (Future)
```python
# Using corynth-plugin-sdk-python
from corynth import PluginBuilder

builder = PluginBuilder("myplugin", "1.0.0")
# ... implement plugin
builder.serve()
```

### Node.js Plugins (Future)
```javascript
// Using corynth-plugin-sdk-node
const { PluginBuilder } = require('@corynth/plugin-sdk');

const plugin = new PluginBuilder('myplugin', '1.0.0');
// ... implement plugin
plugin.serve();
```

## Registry & Distribution

### Repository Structure
```
corynth/plugins/
├── official/          # Core Corynth team plugins
│   ├── http/
│   ├── file/
│   └── calculator/
├── community/         # Community plugins
└── registry.json      # Plugin metadata
```

### Registry Format
```json
{
  "name": "myplugin",
  "version": "1.0.0", 
  "description": "Plugin description",
  "author": "Author Name",
  "platforms": ["linux", "darwin", "windows"],
  "architectures": ["amd64", "arm64"],
  "language": "go",
  "executable": "corynth-plugin-myplugin",
  "actions": [
    {
      "name": "action",
      "description": "What it does",
      "inputs": {"param": "string"},
      "outputs": {"result": "string"}
    }
  ]
}
```

### Installation Process
1. **Discovery**: Search registry for plugin
2. **Download**: Clone repository or download binary
3. **Build**: Compile source for target platform
4. **Install**: Copy executable to plugins directory
5. **Cache**: Remember for future use

## Performance Characteristics

### Benchmarks
- **Plugin startup**: ~100ms (including gRPC handshake)
- **Action execution**: ~10ms overhead (vs native function call)
- **Memory usage**: ~5MB per plugin process
- **Connection reuse**: Persistent connections for multiple actions

### Scalability
- **Concurrent plugins**: No hard limits (OS process limits apply)
- **Plugin isolation**: Full process isolation prevents interference
- **Resource management**: Individual plugin resource limits
- **Network efficiency**: Binary protobuf serialization

## Security Features

### Process Isolation
- Each plugin runs in separate process
- OS-level security boundaries
- No shared memory or file handles
- Plugin crashes don't affect core

### Input Validation
- Type-safe parameter validation
- Required parameter enforcement
- Custom validation logic support
- Sanitized error messages

### Resource Limits
- Process-level timeouts
- Memory usage monitoring
- CPU usage limits (future)
- Network access control (future)

## Troubleshooting

### Common Issues

#### Plugin Not Starting
**Symptoms**: "failed to read handshake" errors

**Solutions**:
1. Check plugin executable permissions: `chmod +x plugin`
2. Test plugin manually: `./plugin serve`
3. Check for missing dependencies: `ldd plugin`

#### gRPC Connection Failures
**Symptoms**: "failed to connect to plugin" errors

**Solutions**:
1. Verify plugin prints handshake correctly
2. Check port availability: `netstat -an | grep PORT`
3. Test with minimal plugin implementation

#### Invalid Handshake Format
**Symptoms**: "invalid handshake format" errors

**Solutions**:
1. Ensure plugin outputs: `1|1|tcp|127.0.0.1:PORT|grpc`
2. Check for extra output before handshake
3. Verify plugin starts gRPC server correctly

### Debugging Commands

```bash
# Test plugin directly
./corynth-plugin-name serve

# Check process status  
ps aux | grep corynth-plugin

# Monitor gRPC traffic (with grpcurl)
grpcurl -plaintext 127.0.0.1:PORT list

# Test plugin health
grpcurl -plaintext 127.0.0.1:PORT corynth.plugin.v1.PluginService/Health
```

## Migration Guide

### From Compiled Plugins

**Before (Go Plugin):**
```go
var ExportedPlugin plugin.Plugin = &MyPlugin{}
```

**After (gRPC Plugin):**
```go
func main() {
    plugin := NewMyPlugin()
    sdk := pluginv2.NewSDK(plugin)
    sdk.Serve()
}
```

### Benefits of Migration
1. **No Version Issues**: Plugins work across Corynth versions
2. **Cross Platform**: Single source, multiple binaries
3. **Better Testing**: Direct plugin testing with `serve` mode
4. **Multi-Language**: Expand beyond Go
5. **Process Safety**: Plugin crashes don't kill workflows

## Future Enhancements

### Planned Features
- **Plugin Registry API**: REST API for plugin discovery
- **Binary Distribution**: Pre-compiled plugins for faster installation
- **Plugin Signing**: Code signing for security verification
- **Resource Limits**: CPU/memory constraints per plugin
- **Plugin Monitoring**: Health checks and performance metrics

### SDK Expansion
- **Python SDK**: `corynth-plugin-sdk-python`
- **Node.js SDK**: `@corynth/plugin-sdk`
- **Rust SDK**: `corynth-plugin-sdk-rs`
- **Go Templates**: Plugin scaffolding generators

## Conclusion

The gRPC plugin architecture provides a **robust, scalable, and developer-friendly** foundation for Corynth's plugin ecosystem. By following Terraform's proven patterns, we've eliminated the fundamental issues of Go's plugin system while enabling cross-platform distribution and multi-language development.

This architecture positions Corynth for **long-term growth** with a plugin ecosystem that can scale to hundreds of plugins across multiple languages and platforms.