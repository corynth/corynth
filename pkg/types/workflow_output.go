package types

import "time"

// WorkflowOutput represents the outputs from a workflow execution
type WorkflowOutput struct {
	WorkflowName string                 `json:"workflow_name"`
	Outputs      map[string]interface{} `json:"outputs"`
	Timestamp    time.Time              `json:"timestamp"`
}