# Corynth RPC Plugin Examples

This directory contains example implementations of Corynth plugins using the RPC/subprocess architecture.

## RPC Plugin Architecture

Instead of Go plugins (which have compatibility issues), Corynth now uses subprocess-based RPC plugins that:

- ✅ **No version conflicts**: Each plugin runs independently
- ✅ **Language agnostic**: Python, Node.js, Rust, shell scripts, etc.
- ✅ **Better isolation**: Plugin crashes don't affect main engine
- ✅ **Easier development**: Standard debugging tools work
- ✅ **Simpler distribution**: No compilation required

## Plugin Interface

All RPC plugins must support these command-line interfaces:

### 1. Metadata Query
```bash
./plugin-name.py metadata
```
Returns JSON with plugin information:
```json
{
  "name": "plugin-name",
  "version": "1.0.0", 
  "description": "Plugin description",
  "author": "Author Name",
  "tags": ["tag1", "tag2"]
}
```

### 2. Actions Query  
```bash
./plugin-name.py actions
```
Returns JSON with available actions:
```json
{
  "action-name": {
    "description": "Action description",
    "inputs": {
      "param1": {"type": "string", "required": true, "description": "Parameter description"}
    },
    "outputs": {
      "result": {"type": "string", "description": "Result description"}
    }
  }
}
```

### 3. Action Execution
```bash
echo '{"param1": "value1"}' | ./plugin-name.py action-name
```
- Parameters passed via stdin as JSON
- Results returned via stdout as JSON
- Errors returned as `{"error": "error message"}`

## Example Plugins

### calculator-plugin.py
Mathematical calculations and basic operations.

Usage:
```bash
echo '{"expression": "15 + 25", "precision": 2}' | ./calculator-plugin.py calculate
```

### file-plugin.py  
File system operations (read, write).

Usage:
```bash
# Write file
echo '{"path": "/tmp/test.txt", "content": "hello", "create_dirs": true}' | ./file-plugin.py write

# Read file
echo '{"path": "/tmp/test.txt"}' | ./file-plugin.py read
```

### http-plugin.py
HTTP client for REST API calls.

Usage:
```bash
# GET request
echo '{"url": "https://api.github.com/users/octocat", "timeout": 10}' | ./http-plugin.py get

# POST request  
echo '{"url": "https://httpbin.org/post", "body": "test data"}' | ./http-plugin.py post
```

## Plugin Development Template

### Python Template
```python
#!/usr/bin/env python3
import json
import sys
from typing import Dict, Any

class MyPlugin:
    def __init__(self):
        self.metadata = {
            "name": "my-plugin",
            "version": "1.0.0",
            "description": "Plugin description",
            "author": "Your Name",
            "tags": ["tag1", "tag2"]
        }
    
    def get_metadata(self) -> Dict[str, Any]:
        return self.metadata
    
    def get_actions(self) -> Dict[str, Any]:
        return {
            "my-action": {
                "description": "Action description",
                "inputs": {
                    "param1": {"type": "string", "required": True, "description": "Parameter description"}
                },
                "outputs": {
                    "result": {"type": "string", "description": "Result description"}
                }
            }
        }
    
    def execute(self, action: str, params: Dict[str, Any]) -> Dict[str, Any]:
        if action == "my-action":
            # Implement your action logic here
            return {"result": "success"}
        else:
            return {"error": f"Unknown action: {action}"}

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "action required"}))
        sys.exit(1)
        
    action = sys.argv[1]
    
    # Read parameters from stdin
    try:
        params_data = sys.stdin.read().strip()
        params = json.loads(params_data) if params_data else {}
    except json.JSONDecodeError:
        params = {}
    
    plugin = MyPlugin()
    
    if action == "metadata":
        result = plugin.get_metadata()
    elif action == "actions":
        result = plugin.get_actions()
    else:
        result = plugin.execute(action, params)
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

## Installation & Usage

1. **Place plugin script** in `examples/plugins/` directory
2. **Make executable**: `chmod +x plugin-name.py`
3. **Use in workflows**: Reference by name (e.g., `plugin = "my-plugin"`)
4. **Auto-discovery**: Corynth automatically detects and loads plugins on first use

## Benefits Over Go Plugins

- **No compilation**: Scripts run directly
- **Cross-platform**: Works on any OS with the interpreter
- **Easy debugging**: Use standard language debugging tools
- **Rapid development**: Edit and test immediately
- **No dependency conflicts**: Each plugin has its own runtime environment