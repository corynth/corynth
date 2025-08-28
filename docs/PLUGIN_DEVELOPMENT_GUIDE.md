# Plugin Development Guide

Comprehensive guide for developing, testing, and distributing Corynth plugins using the JSON protocol architecture.

## Overview

Corynth uses a JSON protocol plugin architecture where plugins are executable scripts that communicate via JSON over stdin/stdout. This approach provides maximum language flexibility, version independence, and process isolation for plugin development.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Plugin Architecture](#plugin-architecture)
3. [Development Environment](#development-environment)
4. [Plugin Interface](#plugin-interface)
5. [Implementation Patterns](#implementation-patterns)
6. [Testing Framework](#testing-framework)
7. [Documentation Standards](#documentation-standards)
8. [Distribution Process](#distribution-process)
9. [Best Practices](#best-practices)
10. [Examples](#examples)
11. [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites
- Any programming language (Python, Node.js, Go, Rust, etc.)
- Basic understanding of JSON data structures
- Command-line development experience

### Create Your First Plugin

1. **Choose your language and create plugin file**:
```bash
# Python example
touch my-plugin.py
chmod +x my-plugin.py
```

2. **Implement the basic structure**:
```python
#!/usr/bin/env python3
import json
import sys

class MyPlugin:
    def get_metadata(self):
        return {
            "name": "my-plugin",
            "version": "1.0.0", 
            "description": "Description of what this plugin does",
            "author": "Your Name",
            "tags": ["category", "functionality"]
        }
    
    def get_actions(self):
        return {
            "my-action": {
                "description": "Description of the action",
                "inputs": {
                    "param1": {"type": "string", "required": True}
                },
                "outputs": {
                    "result": {"type": "string"}
                }
            }
        }
    
    def execute(self, action, params):
        if action == "my-action":
            return {"result": f"Hello {params.get('param1', 'World')}"}
        raise ValueError(f"Unknown action: {action}")

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "action required"}))
        sys.exit(1)
    
    action = sys.argv[1]
    plugin = MyPlugin()
    
    if action == "metadata":
        result = plugin.get_metadata()
    elif action == "actions":
        result = plugin.get_actions()
    else:
        params_data = sys.stdin.read().strip()
        params = json.loads(params_data) if params_data else {}
        result = plugin.execute(action, params)
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

3. **Test your plugin**:
```bash
# Test metadata
./my-plugin.py metadata

# Test actions  
./my-plugin.py actions

# Test execution
echo '{"param1": "World"}' | ./my-plugin.py my-action
```

## Plugin Architecture

### Core Concepts

Corynth JSON protocol plugins are:
- **Language agnostic** - Implement in any programming language
- **Process isolated** - Each plugin runs in its own process
- **JSON communication** - Simple JSON over stdin/stdout
- **Command-line driven** - Actions specified as command-line arguments
- **Stateless** - Each invocation is independent

### Plugin Communication Protocol

The plugin communication follows this simple pattern:

```bash
# Get plugin metadata
./plugin.py metadata
# Output: {"name": "my-plugin", "version": "1.0.0", ...}

# Get available actions
./plugin.py actions  
# Output: {"action1": {...}, "action2": {...}}

# Execute action with parameters
echo '{"param1": "value"}' | ./plugin.py action1
# Output: {"result": "action result"}
```

### Plugin Lifecycle

1. **Discovery**: Plugin discovered in plugin directory
2. **Metadata**: Corynth calls `./plugin metadata` to get plugin info
3. **Action Discovery**: Corynth calls `./plugin actions` to get available actions
4. **Execution**: Corynth calls `./plugin <action>` with JSON params via stdin
5. **Cleanup**: Process terminates after execution

## Development Environment

### Directory Structure
```
my-plugin/
├── plugin.py                  # Main plugin executable
├── README.md                  # Plugin documentation
├── requirements.txt           # Dependencies (if Python)
├── package.json              # Dependencies (if Node.js)
├── samples/                  # Example workflows
│   ├── basic-usage.hcl
│   └── advanced-features.hcl
└── tests/                    # Test files
    └── test_plugin.py
```

### Installation Location
Place your plugin in the Corynth plugins directory:
```bash
# Copy to plugin directory
cp my-plugin.py ~/.corynth/plugins/my-plugin

# Or create symbolic link for development
ln -s /path/to/development/my-plugin.py ~/.corynth/plugins/my-plugin
```

## Plugin Interface

### Required Commands

Every plugin must support three commands:

#### 1. Metadata Command
```bash
./plugin metadata
```
Returns plugin information:
```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Brief description of plugin functionality",
  "author": "Your Name",
  "tags": ["category", "functionality", "integration"],
  "license": "Apache-2.0"
}
```

#### 2. Actions Command  
```bash
./plugin actions
```
Returns available actions:
```json
{
  "action1": {
    "description": "Description of what this action does",
    "inputs": {
      "required_param": {
        "type": "string",
        "required": true,
        "description": "A required string parameter"
      },
      "optional_param": {
        "type": "number", 
        "required": false,
        "default": 42,
        "description": "An optional parameter with default"
      }
    },
    "outputs": {
      "result": {
        "type": "string",
        "description": "The operation result"
      },
      "status": {
        "type": "number", 
        "description": "Status code"
      }
    }
  }
}
```

#### 3. Action Execution
```bash
echo '{"param1": "value"}' | ./plugin action1
```
Executes the specified action with JSON parameters from stdin:
```json
{
  "result": "action completed successfully",
  "status": 200
}
```

### Parameter Types

Supported parameter types:
- **string**: Text values
- **number**: Numeric values (integers and floats)
- **boolean**: True/false values
- **array**: Lists of values
- **object**: Nested JSON objects

## Implementation Patterns

### 1. Python Implementation Pattern

```python
#!/usr/bin/env python3
import json
import sys
import requests
from typing import Dict, Any

class HttpPlugin:
    def get_metadata(self):
        return {
            "name": "http-client",
            "version": "1.0.0",
            "description": "HTTP client for making web requests",
            "author": "Corynth Team",
            "tags": ["http", "web", "api"]
        }
    
    def get_actions(self):
        return {
            "get": {
                "description": "Make HTTP GET request",
                "inputs": {
                    "url": {"type": "string", "required": True},
                    "headers": {"type": "object", "required": False},
                    "timeout": {"type": "number", "required": False, "default": 30}
                },
                "outputs": {
                    "status_code": {"type": "number"},
                    "body": {"type": "string"},
                    "headers": {"type": "object"}
                }
            },
            "post": {
                "description": "Make HTTP POST request", 
                "inputs": {
                    "url": {"type": "string", "required": True},
                    "data": {"type": "object", "required": False},
                    "headers": {"type": "object", "required": False},
                    "timeout": {"type": "number", "required": False, "default": 30}
                },
                "outputs": {
                    "status_code": {"type": "number"},
                    "body": {"type": "string"}, 
                    "headers": {"type": "object"}
                }
            }
        }
    
    def execute(self, action: str, params: Dict[str, Any]) -> Dict[str, Any]:
        try:
            if action == "get":
                return self._make_request("GET", params)
            elif action == "post":
                return self._make_request("POST", params)
            else:
                raise ValueError(f"Unknown action: {action}")
        except Exception as e:
            return {"error": str(e)}
    
    def _make_request(self, method: str, params: Dict[str, Any]) -> Dict[str, Any]:
        url = params.get("url")
        headers = params.get("headers", {})
        timeout = params.get("timeout", 30)
        
        if method == "POST":
            data = params.get("data")
            response = requests.request(method, url, json=data, headers=headers, timeout=timeout)
        else:
            response = requests.request(method, url, headers=headers, timeout=timeout)
        
        return {
            "status_code": response.status_code,
            "body": response.text,
            "headers": dict(response.headers)
        }

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "action required"}))
        sys.exit(1)
    
    action = sys.argv[1]
    plugin = HttpPlugin()
    
    if action == "metadata":
        result = plugin.get_metadata()
    elif action == "actions":
        result = plugin.get_actions()
    else:
        params_data = sys.stdin.read().strip()
        params = json.loads(params_data) if params_data else {}
        result = plugin.execute(action, params)
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

### 2. Node.js Implementation Pattern

```javascript
#!/usr/bin/env node

const fs = require('fs');
const { execSync } = require('child_process');

class ShellPlugin {
    getMetadata() {
        return {
            name: "shell-executor",
            version: "1.0.0", 
            description: "Execute shell commands safely",
            author: "Corynth Team",
            tags: ["shell", "command", "system"]
        };
    }
    
    getActions() {
        return {
            exec: {
                description: "Execute shell command",
                inputs: {
                    command: { type: "string", required: true },
                    working_dir: { type: "string", required: false },
                    timeout: { type: "number", required: false, default: 300 }
                },
                outputs: {
                    output: { type: "string" },
                    exit_code: { type: "number" },
                    success: { type: "boolean" }
                }
            }
        };
    }
    
    execute(action, params) {
        try {
            if (action === "exec") {
                return this.executeCommand(params);
            } else {
                throw new Error(`Unknown action: ${action}`);
            }
        } catch (error) {
            return { error: error.message };
        }
    }
    
    executeCommand(params) {
        const command = params.command;
        const workingDir = params.working_dir || process.cwd();
        const timeout = (params.timeout || 300) * 1000; // Convert to milliseconds
        
        try {
            const output = execSync(command, {
                cwd: workingDir,
                timeout: timeout,
                encoding: 'utf8'
            });
            
            return {
                output: output.toString(),
                exit_code: 0,
                success: true
            };
        } catch (error) {
            return {
                output: error.stdout ? error.stdout.toString() : "",
                exit_code: error.status || 1,
                success: false
            };
        }
    }
}

function main() {
    const args = process.argv.slice(2);
    if (args.length < 1) {
        console.log(JSON.stringify({ error: "action required" }));
        process.exit(1);
    }
    
    const action = args[0];
    const plugin = new ShellPlugin();
    
    let result;
    if (action === "metadata") {
        result = plugin.getMetadata();
    } else if (action === "actions") {
        result = plugin.getActions();
    } else {
        // Read parameters from stdin
        const input = fs.readFileSync(0, 'utf8').trim();
        const params = input ? JSON.parse(input) : {};
        result = plugin.execute(action, params);
    }
    
    console.log(JSON.stringify(result));
}

if (require.main === module) {
    main();
}
```

### 3. Go Implementation Pattern

```go
#!/usr/bin/env go run

package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "strconv"
    "time"
)

type FilePlugin struct{}

type Metadata struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    Author      string   `json:"author"`
    Tags        []string `json:"tags"`
}

type ActionSpec struct {
    Description string                 `json:"description"`
    Inputs      map[string]interface{} `json:"inputs"`
    Outputs     map[string]interface{} `json:"outputs"`
}

func (p *FilePlugin) GetMetadata() Metadata {
    return Metadata{
        Name:        "file-operations",
        Version:     "1.0.0",
        Description: "File system operations",
        Author:      "Corynth Team", 
        Tags:        []string{"file", "filesystem", "io"},
    }
}

func (p *FilePlugin) GetActions() map[string]ActionSpec {
    return map[string]ActionSpec{
        "read": {
            Description: "Read file contents",
            Inputs: map[string]interface{}{
                "path": map[string]interface{}{
                    "type":     "string",
                    "required": true,
                },
            },
            Outputs: map[string]interface{}{
                "content": map[string]interface{}{"type": "string"},
                "size":    map[string]interface{}{"type": "number"},
            },
        },
        "write": {
            Description: "Write content to file",
            Inputs: map[string]interface{}{
                "path": map[string]interface{}{
                    "type":     "string", 
                    "required": true,
                },
                "content": map[string]interface{}{
                    "type":     "string",
                    "required": true,
                },
            },
            Outputs: map[string]interface{}{
                "success": map[string]interface{}{"type": "boolean"},
                "size":    map[string]interface{}{"type": "number"},
            },
        },
    }
}

func (p *FilePlugin) Execute(action string, params map[string]interface{}) (map[string]interface{}, error) {
    switch action {
    case "read":
        return p.readFile(params)
    case "write":
        return p.writeFile(params)
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}

func (p *FilePlugin) readFile(params map[string]interface{}) (map[string]interface{}, error) {
    path, ok := params["path"].(string)
    if !ok {
        return map[string]interface{}{"error": "path parameter required"}, nil
    }
    
    content, err := os.ReadFile(path)
    if err != nil {
        return map[string]interface{}{"error": err.Error()}, nil
    }
    
    return map[string]interface{}{
        "content": string(content),
        "size":    len(content),
    }, nil
}

func (p *FilePlugin) writeFile(params map[string]interface{}) (map[string]interface{}, error) {
    path, ok := params["path"].(string)
    if !ok {
        return map[string]interface{}{"error": "path parameter required"}, nil
    }
    
    content, ok := params["content"].(string)
    if !ok {
        return map[string]interface{}{"error": "content parameter required"}, nil
    }
    
    err := os.WriteFile(path, []byte(content), 0644)
    if err != nil {
        return map[string]interface{}{"error": err.Error()}, nil
    }
    
    return map[string]interface{}{
        "success": true,
        "size":    len(content),
    }, nil
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println(`{"error": "action required"}`)
        os.Exit(1)
    }
    
    action := os.Args[1]
    plugin := &FilePlugin{}
    
    var result interface{}
    var err error
    
    if action == "metadata" {
        result = plugin.GetMetadata()
    } else if action == "actions" {
        result = plugin.GetActions()
    } else {
        // Read parameters from stdin
        input, _ := io.ReadAll(os.Stdin)
        var params map[string]interface{}
        if len(input) > 0 {
            json.Unmarshal(input, &params)
        } else {
            params = make(map[string]interface{})
        }
        result, err = plugin.Execute(action, params)
    }
    
    if err != nil {
        result = map[string]interface{}{"error": err.Error()}
    }
    
    output, _ := json.Marshal(result)
    fmt.Println(string(output))
}
```

## Testing Framework

### Unit Testing

Create comprehensive tests for your plugin:

```python
#!/usr/bin/env python3
import json
import subprocess
import unittest
from io import StringIO

class TestMyPlugin(unittest.TestCase):
    def setUp(self):
        self.plugin_path = "./my-plugin.py"
    
    def run_plugin(self, action, params=None):
        """Helper to run plugin command"""
        cmd = [self.plugin_path, action]
        
        if params:
            input_data = json.dumps(params)
            process = subprocess.run(
                cmd, 
                input=input_data, 
                text=True, 
                capture_output=True
            )
        else:
            process = subprocess.run(cmd, capture_output=True, text=True)
        
        if process.returncode != 0:
            raise Exception(f"Plugin failed: {process.stderr}")
        
        return json.loads(process.stdout)
    
    def test_metadata(self):
        """Test plugin metadata"""
        result = self.run_plugin("metadata")
        
        self.assertIn("name", result)
        self.assertIn("version", result)
        self.assertIn("description", result)
        self.assertIn("author", result)
        self.assertIn("tags", result)
        
        self.assertIsInstance(result["tags"], list)
        self.assertGreater(len(result["tags"]), 0)
    
    def test_actions(self):
        """Test plugin actions"""
        result = self.run_plugin("actions")
        
        self.assertIsInstance(result, dict)
        self.assertGreater(len(result), 0)
        
        # Check action structure
        for action_name, action_spec in result.items():
            self.assertIn("description", action_spec)
            self.assertIn("inputs", action_spec)
            self.assertIn("outputs", action_spec)
    
    def test_valid_execution(self):
        """Test valid plugin execution"""
        params = {"param1": "test_value"}
        result = self.run_plugin("my-action", params)
        
        self.assertNotIn("error", result)
        self.assertIn("result", result)
    
    def test_invalid_action(self):
        """Test invalid action handling"""
        try:
            result = self.run_plugin("invalid-action", {})
            self.assertIn("error", result)
        except Exception:
            pass  # Expected to fail
    
    def test_missing_required_param(self):
        """Test missing required parameter"""
        result = self.run_plugin("my-action", {})
        # Should handle gracefully
        
if __name__ == "__main__":
    unittest.main()
```

### Integration Testing with Corynth

Create test workflows:

```hcl
# test-workflow.hcl
workflow "test-my-plugin" {
  description = "Test workflow for my plugin"
  version     = "1.0.0"

  step "test_basic" {
    plugin = "my-plugin" 
    action = "my-action"
    
    params = {
      param1 = "test_value"
    }
  }
  
  step "test_with_dependency" {
    plugin = "my-plugin"
    action = "my-action"
    
    depends_on = ["test_basic"]
    
    params = {
      param1 = "${test_basic.result}"
    }
  }
}
```

Test with Corynth:
```bash
# Validate workflow syntax
corynth validate test-workflow.hcl

# Test plugin execution
corynth apply test-workflow.hcl --auto-approve
```

## Documentation Standards

### README.md Template

```markdown
# My Plugin

Brief description of what the plugin does and its primary use cases.

## Installation  

1. Download the plugin:
```bash
curl -o ~/.corynth/plugins/my-plugin https://raw.githubusercontent.com/username/my-plugin/main/plugin.py
chmod +x ~/.corynth/plugins/my-plugin
```

2. Install dependencies (if needed):
```bash
# Python
pip install -r requirements.txt

# Node.js
npm install
```

## Actions

### my-action

Description of the action and what it accomplishes.

**Parameters:**
- `required_param` (string, required): Description of the parameter
- `optional_param` (number, optional, default: 42): Description with default
- `list_param` (array, optional): Description of list parameter

**Returns:**
- `result` (string): Description of the result
- `status` (number): Operation status code

**Example:**
```hcl
step "example" {
  plugin = "my-plugin"
  action = "my-action"
  
  params = {
    required_param = "example_value"
    optional_param = 100
    list_param = ["item1", "item2"]
  }
}
```

## Development

### Testing
```bash
# Run unit tests
python test_plugin.py

# Test plugin directly
./plugin.py metadata
./plugin.py actions
echo '{"param1": "test"}' | ./plugin.py my-action
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

Apache-2.0 License
```

## Best Practices

### Security
- **Validate all inputs** thoroughly before processing
- **Sanitize file paths** to prevent directory traversal attacks  
- **Use subprocess with shell=False** when executing commands
- **Handle credentials via environment variables**, not parameters
- **Limit resource usage** with timeouts and size constraints

### Performance  
- **Keep plugins lightweight** - minimize dependencies
- **Use efficient algorithms** for data processing
- **Implement timeouts** for external API calls
- **Stream large data** instead of loading into memory
- **Cache expensive operations** when appropriate

### Error Handling
- **Return structured errors** as JSON objects
- **Provide meaningful error messages** with context
- **Handle edge cases gracefully**
- **Validate parameters early** before expensive operations
- **Use try-catch blocks** around all external operations

### Code Quality
```python
# Good practices example
import json
import sys
import logging
from typing import Dict, Any, Optional

class MyPlugin:
    def __init__(self):
        # Initialize logging
        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(__name__)
    
    def validate_params(self, params: Dict[str, Any], required: list) -> Optional[str]:
        """Validate required parameters"""
        for param in required:
            if param not in params:
                return f"Missing required parameter: {param}"
        return None
    
    def execute(self, action: str, params: Dict[str, Any]) -> Dict[str, Any]:
        """Execute action with proper error handling"""
        try:
            self.logger.info(f"Executing action: {action}")
            
            if action == "my-action":
                # Validate parameters
                error = self.validate_params(params, ["required_param"])
                if error:
                    return {"error": error}
                
                # Execute action
                result = self._do_my_action(params)
                self.logger.info("Action completed successfully")
                return result
            
            return {"error": f"Unknown action: {action}"}
            
        except Exception as e:
            self.logger.error(f"Action failed: {e}")
            return {"error": str(e)}
```

## Troubleshooting

### Common Issues

#### Plugin Not Executable
```
bash: ./my-plugin.py: Permission denied
```
**Solution**: Make plugin executable: `chmod +x my-plugin.py`

#### JSON Parsing Errors
```
{"error": "Expecting value: line 1 column 1 (char 0)"}
```
**Solution**: Ensure valid JSON is passed via stdin. Check for empty input.

#### Missing Dependencies  
```
ModuleNotFoundError: No module named 'requests'
```
**Solution**: Install required dependencies or include requirements.txt

#### Invalid Shebang Line
```
/usr/bin/env: 'python3': No such file or directory
```
**Solution**: Verify Python3 is installed or adjust shebang line

### Debugging Tips

1. **Test plugin manually** before using with Corynth:
```bash
# Test each command
./plugin.py metadata
./plugin.py actions
echo '{"test": "value"}' | ./plugin.py my-action
```

2. **Add debug logging**:
```python
import logging
logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)
logger.debug(f"Received params: {params}")
```

3. **Validate JSON output**:
```bash
./plugin.py metadata | jq .
```

4. **Check for common issues**:
- Proper shebang line (`#!/usr/bin/env python3`)
- Executable permissions (`chmod +x`)
- Valid JSON output format
- All required parameters handled
- Error cases return JSON with "error" key

## Examples

### Working Calculator Plugin

The calculator plugin demonstrates a complete implementation:

```python
#!/usr/bin/env python3
"""
Corynth Calculator JSON Protocol Plugin
"""
import json
import sys
from typing import Dict, Any

class CalculatorPlugin:
    def __init__(self):
        self.metadata = {
            "name": "calculator",
            "version": "1.0.0",
            "description": "Mathematical calculations and unit conversions",
            "author": "Corynth Team",
            "tags": ["math", "calculation", "utility"]
        }
    
    def get_metadata(self) -> Dict[str, Any]:
        return self.metadata
    
    def get_actions(self) -> Dict[str, Any]:
        return {
            "calculate": {
                "description": "Perform mathematical calculations",
                "inputs": {
                    "expression": {
                        "type": "string", 
                        "required": True, 
                        "description": "Mathematical expression to evaluate"
                    },
                    "precision": {
                        "type": "number", 
                        "required": False, 
                        "default": 2, 
                        "description": "Decimal precision"
                    }
                },
                "outputs": {
                    "result": {"type": "number", "description": "Calculation result"},
                    "expression": {"type": "string", "description": "Original expression"}
                }
            }
        }
    
    def execute(self, action: str, params: Dict[str, Any]) -> Dict[str, Any]:
        try:
            if action == "calculate":
                return self._handle_calculate(params)
            else:
                raise ValueError(f"Unknown action: {action}")
        except Exception as e:
            return {"error": str(e)}
    
    def _handle_calculate(self, params: Dict[str, Any]) -> Dict[str, Any]:
        expression = params.get("expression")
        precision = params.get("precision", 2)
        
        if not expression:
            raise ValueError("expression parameter is required")
        
        # Simple safe evaluation - only allow basic math
        allowed_chars = set("0123456789+-*/.()\\n\\t ")
        if not all(c in allowed_chars for c in expression):
            raise ValueError("Expression contains invalid characters")
        
        # Evaluate safely
        try:
            result = eval(expression, {"__builtins__": {}}, {})
            if isinstance(result, (int, float)):
                result = round(float(result), precision)
            else:
                raise ValueError("Expression must result in a number")
        except Exception as e:
            raise ValueError(f"Invalid expression: {e}")
        
        return {
            "result": result,
            "expression": expression
        }

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
    
    plugin = CalculatorPlugin()
    
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

This plugin can be tested directly:
```bash
# Install and test
cp calculator.py ~/.corynth/plugins/calculator
chmod +x ~/.corynth/plugins/calculator

# Test commands
./calculator metadata
./calculator actions
echo '{"expression": "10 + 5 * 2", "precision": 2}' | ./calculator calculate
```

---

This guide provides everything needed to create robust, language-independent Corynth plugins using the JSON protocol architecture. The approach offers maximum flexibility, version independence, and process isolation while maintaining simplicity and reliability.