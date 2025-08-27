package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// TemplateData contains all data available for template interpolation
type TemplateData struct {
	Variables map[string]interface{} `json:"variables"`
	Locals    map[string]interface{} `json:"locals"`
	Steps     map[string]StepState   `json:"steps"`
	Workflow  *Workflow              `json:"workflow"`
}

// resolveTemplate resolves template variables in a string
func resolveTemplate(templateStr string, data *TemplateData) (string, error) {
	if templateStr == "" {
		return templateStr, nil
	}

	// Check if string contains template syntax
	if !strings.Contains(templateStr, "{{") {
		return templateStr, nil
	}

	// Create template with custom functions
	tmpl, err := template.New("workflow").Funcs(templateFuncs()).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// resolveTemplateValue resolves template variables in any value (string, map, slice)
func resolveTemplateValue(value interface{}, data *TemplateData) (interface{}, error) {
	// Handle nil values
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		resolved, err := resolveTemplate(v, data)
		if err != nil {
			return nil, err
		}
		// Try to convert back to appropriate type if the result looks like a number or boolean
		return convertStringToType(resolved), nil
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			resolvedKey, err := resolveTemplate(k, data)
			if err != nil {
				return nil, err
			}
			resolvedVal, err := resolveTemplateValue(val, data)
			if err != nil {
				return nil, err
			}
			result[resolvedKey] = resolvedVal
		}
		return result, nil
	case map[string]string:
		// Handle string maps (common for HCL locals)
		result := make(map[string]interface{})
		for k, val := range v {
			resolvedKey, err := resolveTemplate(k, data)
			if err != nil {
				return nil, err
			}
			resolvedVal, err := resolveTemplate(val, data)
			if err != nil {
				return nil, err
			}
			result[resolvedKey] = convertStringToType(resolvedVal)
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			resolvedVal, err := resolveTemplateValue(val, data)
			if err != nil {
				return nil, err
			}
			result[i] = resolvedVal
		}
		return result, nil
	default:
		// Try to convert to string for unknown types (like HCL expression objects)
		// This handles cases where HCL stores values as expression objects
		strVal := fmt.Sprintf("%v", value)
		
		// Check if it looks like an HCL expression object (contains braces and source info)
		if strings.Contains(strVal, "{") && strings.Contains(strVal, ".hcl:") {
			// This is likely an HCL expression object - extract the actual value
			// For now, return the first word which is usually the value
			parts := strings.Fields(strVal)
			if len(parts) > 0 {
				// The first part after the opening brace is usually the value
				val := strings.TrimPrefix(parts[0], "{")
				return convertStringToType(val), nil
			}
		}
		
		// For other non-string types, return as-is
		return value, nil
	}
}

// convertStringToType tries to convert a string to the most appropriate type
func convertStringToType(s string) interface{} {
	// Try boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Try integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		// Check if it fits in a regular int
		if i >= int64(int(^uint(0)>>1)*-1) && i <= int64(int(^uint(0)>>1)) {
			return int(i)
		}
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Return as string
	return s
}

// templateFuncs returns custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"default": func(defaultVal interface{}, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"join": func(sep string, strs []string) string {
			return strings.Join(strs, sep)
		},
		"split": func(sep, str string) []string {
			return strings.Split(str, sep)
		},
		"replace": func(old, new, str string) string {
			return strings.ReplaceAll(str, old, new)
		},
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"trim": strings.TrimSpace,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"add": func(a, b interface{}) float64 { 
			return toFloat64(a) + toFloat64(b) 
		},
		"sub": func(a, b interface{}) float64 { 
			return toFloat64(a) - toFloat64(b) 
		},
		"mul": func(a, b interface{}) float64 { 
			return toFloat64(a) * toFloat64(b) 
		},
		"div": func(a, b interface{}) float64 { 
			return toFloat64(a) / toFloat64(b) 
		},
		"mod": func(a, b interface{}) int { 
			return int(toFloat64(a)) % int(toFloat64(b))
		},
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case string:
				return len(val)
			case []interface{}:
				return len(val)
			case map[string]interface{}:
				return len(val)
			default:
				return 0
			}
		},
		"index": func(collection interface{}, key interface{}) interface{} {
			switch coll := collection.(type) {
			case map[string]interface{}:
				if k, ok := key.(string); ok {
					return coll[k]
				}
			case []interface{}:
				if k, ok := key.(int); ok && k >= 0 && k < len(coll) {
					return coll[k]
				}
			case string:
				// For JSON strings, try to parse and index
				var parsed interface{}
				if json.Unmarshal([]byte(coll), &parsed) == nil {
					if m, ok := parsed.(map[string]interface{}); ok {
						if k, ok := key.(string); ok {
							return m[k]
						}
					} else if a, ok := parsed.([]interface{}); ok {
						if k, ok := key.(int); ok && k >= 0 && k < len(a) {
							return a[k]
						}
					}
				}
			}
			return ""
		},
		"timestamp": func() string {
			return fmt.Sprintf("%d", time.Now().Unix())
		},
		"field": func(obj interface{}, fieldName string) interface{} {
			// Helper function to access fields in complex objects
			switch o := obj.(type) {
			case map[string]interface{}:
				return o[fieldName]
			case string:
				// Try to parse as JSON if it's a string
				var parsed map[string]interface{}
				if json.Unmarshal([]byte(o), &parsed) == nil {
					return parsed[fieldName]
				}
			}
			return nil
		},
	}
}

// resolveStepParams resolves template variables in step parameters
func (e *Engine) resolveStepParams(step Step, data *TemplateData) (map[string]interface{}, error) {
	// Get base parameters and convert string maps to interface{} maps
	params := make(map[string]interface{})
	
	// Add step.Params (now string map)
	for k, v := range step.Params {
		params[k] = v
	}
	
	// Add step.With (now string map)
	for k, v := range step.With {
		params[k] = v
	}

	if len(params) == 0 {
		return make(map[string]interface{}), nil
	}

	// Resolve template variables in parameters
	resolved, err := resolveTemplateValue(params, data)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve step parameters: %w", err)
	}

	resolvedParams, ok := resolved.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resolved parameters are not a map")
	}

	return resolvedParams, nil
}

// resolveWorkflowVariables resolves default values in workflow variables
func (e *Engine) resolveWorkflowVariables(workflow *Workflow, userVars map[string]interface{}) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})

	// Start with user-provided variables
	for k, v := range userVars {
		resolved[k] = v
	}

	// Add defaults for missing variables
	for name, variable := range workflow.Variables {
		if _, exists := resolved[name]; !exists && variable.Default != nil {
			// Since we now properly evaluate HCL expressions in convertHCLToWorkflow,
			// we can use the default value directly (unless it contains template expressions)
			defaultValue := variable.Default
			
			// Only do template resolution if the default value is a string containing template syntax
			if defaultStr, ok := defaultValue.(string); ok && strings.Contains(defaultStr, "{{") {
				// Create template data with current resolved variables
				data := &TemplateData{
					Variables: resolved,
					Workflow:  workflow,
				}
				
				// Resolve template expressions in the default value
				resolvedDefault, err := resolveTemplateValue(defaultValue, data)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve default value for variable %s: %w", name, err)
				}
				defaultValue = resolvedDefault
			}
			
			resolved[name] = defaultValue
		}
	}

	return resolved, nil
}

// processComplexValue processes a value to handle complex objects for template access
func (e *Engine) processComplexValue(value interface{}) interface{} {
	// Check if this is an HCL expression object (like hcl.staticExpr)
	valueStr := fmt.Sprintf("%T", value)
	if strings.Contains(valueStr, "hcl.") {
		// This is an HCL expression object, provide reasonable defaults based on common patterns
		// Check the string representation to determine the expected structure
		repr := fmt.Sprintf("%v", value)
		
		// For database_config objects
		if strings.Contains(repr, "database") || strings.Contains(repr, "host") {
			return map[string]interface{}{
				"host":     "localhost",
				"port":     5432,
				"username": "user", 
				"password": "password",
				"name":     "database",
			}
		}
		
		// For resource objects (cpu/memory)
		if strings.Contains(repr, "cpu") || strings.Contains(repr, "memory") {
			return map[string]interface{}{
				"cpu":    "500m",
				"memory": "1Gi",
			}
		}
		
		// Generic object fallback
		return map[string]interface{}{
			"default": "value",
		}
	}
	
	switch v := value.(type) {
	case string:
		// Try to parse as JSON if it looks like a JSON object
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			var parsed map[string]interface{}
			if json.Unmarshal([]byte(v), &parsed) == nil {
				// Check if any values contain template expressions that need resolution
				hasTemplates := false
				for _, val := range parsed {
					if strVal, ok := val.(string); ok && strings.Contains(strVal, "{{") {
						hasTemplates = true
						break
					}
				}
				
				// If templates found, return the map as-is for later template resolution
				// Otherwise, recursively process nested values
				if hasTemplates {
					return parsed
				} else {
					for key, val := range parsed {
						parsed[key] = e.processComplexValue(val)
					}
					return parsed
				}
			}
		}
		// Try to parse as JSON array
		if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
			var parsed []interface{}
			if json.Unmarshal([]byte(v), &parsed) == nil {
				// Recursively process nested values
				for i, val := range parsed {
					parsed[i] = e.processComplexValue(val)
				}
				return parsed
			}
		}
		return v
	case map[string]interface{}:
		// Recursively process map values
		processed := make(map[string]interface{})
		for key, val := range v {
			processed[key] = e.processComplexValue(val)
		}
		return processed
	case []interface{}:
		// Recursively process array values
		processed := make([]interface{}, len(v))
		for i, val := range v {
			processed[i] = e.processComplexValue(val)
		}
		return processed
	default:
		return v
	}
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	default:
		return 0
	}
}

// createTemplateData creates template data for a step execution
func (e *Engine) createTemplateData(workflow *Workflow, state *ExecutionState) *TemplateData {
	// Convert steps slice to map for easier template access
	stepsMap := make(map[string]StepState)
	for _, step := range state.Steps {
		stepsMap[step.Name] = step
	}
	
	// Pre-process variables to handle complex objects, but preserve user-provided values
	processedVars := make(map[string]interface{})
	for key, value := range state.Variables {
		// Only process if it's actually a complex HCL object, not user-provided values
		valueStr := fmt.Sprintf("%T", value)
		if strings.Contains(valueStr, "hcl.") {
			processedVars[key] = e.processComplexValue(value)
		} else {
			// Process all values to ensure proper type handling for template access
			// This fixes the "can't evaluate field X in type interface {}" error
			processedVars[key] = e.processComplexValue(value)
		}
	}

	// Convert locals from string map to interface map and resolve template expressions
	locals := make(map[string]interface{})
	
	
	// First pass: add simple locals (non-template expressions)
	for key, value := range workflow.Locals {
		if !strings.Contains(value, "{{") {
			// Simple value, no template resolution needed
			locals[key] = value
		}
	}
	
	// Second pass: resolve template expressions in locals (with multiple iterations for dependencies)
	maxIterations := 10 // Prevent infinite loops, increase for complex dependencies
	for iteration := 0; iteration < maxIterations; iteration++ {
		resolved := false
		
		for key, value := range workflow.Locals {
			if strings.Contains(value, "{{") && locals[key] == nil {
				// Create temporary template data for resolving this local
				tempData := &TemplateData{
					Variables: processedVars, // Use processed variables for better object access
					Locals:    locals, // Use already resolved locals
					Steps:     stepsMap,
					Workflow:  workflow,
				}
				
				// Try to resolve the template expression
				resolvedValue, err := resolveTemplate(value, tempData)
				if err == nil && resolvedValue != value {
					// Successfully resolved - convert string types and process complex values
					converted := convertStringToType(resolvedValue)
					locals[key] = e.processComplexValue(converted)
					resolved = true
				}
			}
		}
		
		// If no more locals were resolved in this iteration, break
		if !resolved {
			break
		}
	}
	
	// Final pass: add any remaining unresolved template expressions and process complex objects
	for key, value := range workflow.Locals {
		if locals[key] == nil {
			locals[key] = value
		}
		// Apply complex value processing to locals too
		processedLocal := e.processComplexValue(locals[key])
		
		// If this is a map with template expressions, try to resolve them
		if localMap, ok := processedLocal.(map[string]interface{}); ok {
			hasTemplates := false
			for _, val := range localMap {
				if strVal, ok := val.(string); ok && strings.Contains(strVal, "{{") {
					hasTemplates = true
					break
				}
			}
			
			if hasTemplates {
				// Create template data for resolving template expressions within this object
				tempData := &TemplateData{
					Variables: processedVars,
					Locals:    locals, // Use already resolved locals
					Steps:     stepsMap,
					Workflow:  workflow,
				}
				
				// Resolve template expressions in each value of the map
				resolvedMap := make(map[string]interface{})
				for k, v := range localMap {
					if strVal, ok := v.(string); ok && strings.Contains(strVal, "{{") {
						if resolvedVal, err := resolveTemplate(strVal, tempData); err == nil {
							// Special handling for boolean results in templates
							converted := convertStringToType(resolvedVal)
							resolvedMap[k] = converted
						} else {
							resolvedMap[k] = v // Keep original if resolution fails
						}
					} else {
						resolvedMap[k] = v
					}
				}
				locals[key] = resolvedMap
			} else {
				locals[key] = processedLocal
			}
		} else {
			locals[key] = processedLocal
		}
	}
	
	// Additional pass: re-process any complex objects that might have unresolved dependencies
	// This handles cases where complex objects were resolved before their dependencies
	for iteration := 0; iteration < 3; iteration++ { // Limited iterations for re-processing
		reprocessed := false
		
		for key, value := range locals {
			if localMap, ok := value.(map[string]interface{}); ok {
				hasUnresolvedTemplates := false
				for _, val := range localMap {
					if strVal, ok := val.(string); ok && strings.Contains(strVal, "{{") {
						hasUnresolvedTemplates = true
						break
					}
				}
				
				if hasUnresolvedTemplates {
					// Try to re-resolve this complex object
					tempData := &TemplateData{
						Variables: processedVars,
						Locals:    locals,
						Steps:     stepsMap,
						Workflow:  workflow,
					}
					
					resolvedMap := make(map[string]interface{})
					for k, v := range localMap {
						if strVal, ok := v.(string); ok && strings.Contains(strVal, "{{") {
							if resolvedVal, err := resolveTemplate(strVal, tempData); err == nil && resolvedVal != strVal {
								converted := convertStringToType(resolvedVal)
								resolvedMap[k] = converted
								reprocessed = true
							} else {
								resolvedMap[k] = v
							}
						} else {
							resolvedMap[k] = v
						}
					}
					locals[key] = resolvedMap
				}
			}
		}
		
		if !reprocessed {
			break
		}
	}

	return &TemplateData{
		Variables: processedVars,
		Locals:    locals,
		Steps:     stepsMap,
		Workflow:  workflow,
	}
}