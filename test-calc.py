#!/usr/bin/env python3
"""
Corynth Calculator RPC Plugin - Python Implementation
"""
import json
import sys
import math
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
                    "expression": {"type": "string", "required": True, "description": "Mathematical expression to evaluate"},
                    "precision": {"type": "number", "required": False, "default": 2, "description": "Decimal precision"}
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
        
        # Simple safe evaluation of basic math expressions
        # Only allow basic operations for security
        allowed_chars = set("0123456789+-*/.()\n\t ")
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
        print(json.dumps({"error": "action required as first argument"}))
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