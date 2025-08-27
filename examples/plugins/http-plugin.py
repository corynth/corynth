#!/usr/bin/env python3
"""
Corynth HTTP RPC Plugin - Python Implementation
Handles HTTP requests via subprocess RPC interface
"""
import json
import sys
import requests
from typing import Dict, Any

class HttpPlugin:
    def __init__(self):
        self.metadata = {
            "name": "http",
            "version": "1.0.0", 
            "description": "HTTP client for REST API calls and web requests",
            "author": "Corynth Team",
            "tags": ["http", "web", "api", "rest"]
        }
        
    def get_metadata(self) -> Dict[str, Any]:
        return self.metadata
    
    def get_actions(self) -> Dict[str, Any]:
        return {
            "get": {
                "description": "Perform HTTP GET request",
                "inputs": {
                    "url": {"type": "string", "required": True, "description": "URL to request"},
                    "headers": {"type": "object", "required": False, "description": "Request headers"},
                    "timeout": {"type": "number", "required": False, "default": 30, "description": "Request timeout in seconds"}
                },
                "outputs": {
                    "status_code": {"type": "number", "description": "HTTP status code"},
                    "headers": {"type": "object", "description": "Response headers"},
                    "body": {"type": "string", "description": "Response body"}
                }
            },
            "post": {
                "description": "Perform HTTP POST request", 
                "inputs": {
                    "url": {"type": "string", "required": True, "description": "URL to request"},
                    "body": {"type": "string", "required": False, "description": "Request body"},
                    "headers": {"type": "object", "required": False, "description": "Request headers"},
                    "timeout": {"type": "number", "required": False, "default": 30, "description": "Request timeout in seconds"}
                },
                "outputs": {
                    "status_code": {"type": "number", "description": "HTTP status code"},
                    "headers": {"type": "object", "description": "Response headers"},
                    "body": {"type": "string", "description": "Response body"}
                }
            }
        }
    
    def execute(self, action: str, params: Dict[str, Any]) -> Dict[str, Any]:
        try:
            if action == "get":
                return self._handle_get(params)
            elif action == "post":
                return self._handle_post(params)
            else:
                raise ValueError(f"Unknown action: {action}")
        except Exception as e:
            return {"error": str(e)}
    
    def _handle_get(self, params: Dict[str, Any]) -> Dict[str, Any]:
        url = params.get("url")
        headers = params.get("headers", {})
        timeout = params.get("timeout", 30)
        
        if not url:
            raise ValueError("url parameter is required")
            
        response = requests.get(url, headers=headers, timeout=timeout)
        
        return {
            "status_code": response.status_code,
            "headers": dict(response.headers),
            "body": response.text
        }
    
    def _handle_post(self, params: Dict[str, Any]) -> Dict[str, Any]:
        url = params.get("url")
        body = params.get("body", "")
        headers = params.get("headers", {})
        timeout = params.get("timeout", 30)
        
        if not url:
            raise ValueError("url parameter is required")
            
        response = requests.post(url, data=body, headers=headers, timeout=timeout)
        
        return {
            "status_code": response.status_code,
            "headers": dict(response.headers),
            "body": response.text
        }

def main():
    if len(sys.argv) < 2:
        print(json.dumps({"error": "action required as first argument"}))
        sys.exit(1)
        
    action = sys.argv[1]
    
    # Read parameters from stdin
    try:
        params = json.loads(sys.stdin.read()) if sys.stdin.read() else {}
        sys.stdin.seek(0)  # Reset for re-reading
        params_data = sys.stdin.read().strip()
        if params_data:
            params = json.loads(params_data)
        else:
            params = {}
    except json.JSONDecodeError:
        params = {}
    
    plugin = HttpPlugin()
    
    if action == "metadata":
        result = plugin.get_metadata()
    elif action == "actions":
        result = plugin.get_actions()
    else:
        result = plugin.execute(action, params)
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()