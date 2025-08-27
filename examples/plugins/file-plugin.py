#!/usr/bin/env python3
"""
Corynth File RPC Plugin - Python Implementation
"""
import json
import sys
import os
from typing import Dict, Any

class FilePlugin:
    def __init__(self):
        self.metadata = {
            "name": "file",
            "version": "1.0.0",
            "description": "File system operations (read, write, copy, move)",
            "author": "Corynth Team", 
            "tags": ["file", "filesystem", "io"]
        }
    
    def get_metadata(self) -> Dict[str, Any]:
        return self.metadata
    
    def get_actions(self) -> Dict[str, Any]:
        return {
            "read": {
                "description": "Read file contents",
                "inputs": {
                    "path": {"type": "string", "required": True, "description": "File path to read"}
                },
                "outputs": {
                    "content": {"type": "string", "description": "File contents"}
                }
            },
            "write": {
                "description": "Write content to file",
                "inputs": {
                    "path": {"type": "string", "required": True, "description": "File path to write"},
                    "content": {"type": "string", "required": True, "description": "Content to write"},
                    "create_dirs": {"type": "boolean", "required": False, "default": False, "description": "Create parent directories if needed"}
                },
                "outputs": {
                    "success": {"type": "boolean", "description": "Write operation success"}
                }
            }
        }
    
    def execute(self, action: str, params: Dict[str, Any]) -> Dict[str, Any]:
        try:
            if action == "read":
                return self._handle_read(params)
            elif action == "write":
                return self._handle_write(params)
            else:
                raise ValueError(f"Unknown action: {action}")
        except Exception as e:
            return {"error": str(e)}
    
    def _handle_read(self, params: Dict[str, Any]) -> Dict[str, Any]:
        path = params.get("path")
        if not path:
            raise ValueError("path parameter is required")
            
        with open(path, 'r') as f:
            content = f.read()
        
        return {"content": content}
    
    def _handle_write(self, params: Dict[str, Any]) -> Dict[str, Any]:
        path = params.get("path")
        content = params.get("content", "")
        create_dirs = params.get("create_dirs", False)
        
        if not path:
            raise ValueError("path parameter is required")
        
        if create_dirs:
            os.makedirs(os.path.dirname(path), exist_ok=True)
            
        with open(path, 'w') as f:
            f.write(content)
        
        return {"success": True}

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "action required as first argument"}))
        sys.exit(1)
        
    action = sys.argv[1]
    
    # Read parameters from stdin
    try:
        params_data = sys.stdin.read().strip()
        params = json.loads(params_data) if params_data else {}
    except json.JSONDecodeError:
        params = {}
    
    plugin = FilePlugin()
    
    if action == "metadata":
        result = plugin.get_metadata()
    elif action == "actions":
        result = plugin.get_actions()
    else:
        result = plugin.execute(action, params)
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()