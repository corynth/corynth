package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SlackPlugin implements Slack messaging operations
type SlackPlugin struct{}

// NewSlackPlugin creates a new Slack plugin
func NewSlackPlugin() *SlackPlugin {
	return &SlackPlugin{}
}

func (p *SlackPlugin) Metadata() Metadata {
	return Metadata{
		Name:        "slack",
		Version:     "1.0.0",
		Description: "Slack messaging and notification operations",
		Author:      "Corynth Team",
		Tags:        []string{"slack", "messaging", "notifications", "webhook"},
		License:     "Apache-2.0",
	}
}

func (p *SlackPlugin) Actions() []Action {
	return []Action{
		{
			Name:        "send_message",
			Description: "Send a message to a Slack channel",
			Inputs: map[string]InputSpec{
				"webhook_url": {
					Type:        "string",
					Description: "Slack webhook URL",
					Required:    true,
				},
				"text": {
					Type:        "string",
					Description: "Message text to send",
					Required:    true,
				},
				"channel": {
					Type:        "string",
					Description: "Channel name or ID (optional if webhook is channel-specific)",
					Required:    false,
				},
				"username": {
					Type:        "string",
					Description: "Bot username to display",
					Required:    false,
					Default:     "Corynth",
				},
				"icon_emoji": {
					Type:        "string",
					Description: "Emoji icon for the bot",
					Required:    false,
					Default:     ":robot_face:",
				},
			},
			Outputs: map[string]OutputSpec{
				"success": {
					Type:        "bool",
					Description: "Whether the message was sent successfully",
				},
				"response": {
					Type:        "string",
					Description: "Response from Slack API",
				},
			},
		},
		{
			Name:        "send_rich_message",
			Description: "Send a rich message with attachments to Slack",
			Inputs: map[string]InputSpec{
				"webhook_url": {
					Type:        "string",
					Description: "Slack webhook URL",
					Required:    true,
				},
				"text": {
					Type:        "string",
					Description: "Main message text",
					Required:    true,
				},
				"color": {
					Type:        "string",
					Description: "Color of the attachment sidebar (good, warning, danger, or hex)",
					Required:    false,
					Default:     "good",
				},
				"fields": {
					Type:        "map",
					Description: "Key-value pairs to display as fields",
					Required:    false,
				},
				"channel": {
					Type:        "string",
					Description: "Channel name or ID",
					Required:    false,
				},
				"username": {
					Type:        "string",
					Description: "Bot username to display",
					Required:    false,
					Default:     "Corynth",
				},
			},
			Outputs: map[string]OutputSpec{
				"success": {
					Type:        "bool",
					Description: "Whether the message was sent successfully",
				},
				"response": {
					Type:        "string",
					Description: "Response from Slack API",
				},
			},
		},
		{
			Name:        "workflow_notification",
			Description: "Send workflow status notification to Slack",
			Inputs: map[string]InputSpec{
				"webhook_url": {
					Type:        "string",
					Description: "Slack webhook URL",
					Required:    true,
				},
				"workflow_name": {
					Type:        "string",
					Description: "Name of the workflow",
					Required:    true,
				},
				"status": {
					Type:        "string",
					Description: "Workflow status (started, success, failed, warning)",
					Required:    true,
				},
				"duration": {
					Type:        "string",
					Description: "Workflow execution duration",
					Required:    false,
				},
				"details": {
					Type:        "string",
					Description: "Additional details or error message",
					Required:    false,
				},
				"channel": {
					Type:        "string",
					Description: "Channel name or ID",
					Required:    false,
				},
			},
			Outputs: map[string]OutputSpec{
				"success": {
					Type:        "bool",
					Description: "Whether the notification was sent successfully",
				},
			},
		},
	}
}

func (p *SlackPlugin) Validate(params map[string]interface{}) error {
	// Basic validation - webhook URL should be provided for all actions
	if webhookURL, exists := params["webhook_url"]; exists {
		if url, ok := webhookURL.(string); ok {
			if !strings.HasPrefix(url, "https://hooks.slack.com/services/") {
				return fmt.Errorf("invalid Slack webhook URL format")
			}
		}
	}
	return nil
}

func (p *SlackPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "send_message":
		return p.executeSendMessage(ctx, params)
	case "send_rich_message":
		return p.executeSendRichMessage(ctx, params)
	case "workflow_notification":
		return p.executeWorkflowNotification(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (p *SlackPlugin) executeSendMessage(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	webhookURL, _ := params["webhook_url"].(string)
	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url parameter is required")
	}

	text, _ := params["text"].(string)
	if text == "" {
		return nil, fmt.Errorf("text parameter is required")
	}

	// Build Slack message payload
	payload := map[string]interface{}{
		"text": text,
	}

	if channel, ok := params["channel"].(string); ok && channel != "" {
		payload["channel"] = channel
	}

	if username, ok := params["username"].(string); ok && username != "" {
		payload["username"] = username
	} else {
		payload["username"] = "Corynth"
	}

	if iconEmoji, ok := params["icon_emoji"].(string); ok && iconEmoji != "" {
		payload["icon_emoji"] = iconEmoji
	} else {
		payload["icon_emoji"] = ":robot_face:"
	}

	return p.sendToSlack(ctx, webhookURL, payload)
}

func (p *SlackPlugin) executeSendRichMessage(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	webhookURL, _ := params["webhook_url"].(string)
	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url parameter is required")
	}

	text, _ := params["text"].(string)
	if text == "" {
		return nil, fmt.Errorf("text parameter is required")
	}

	// Build attachment
	attachment := map[string]interface{}{
		"text": text,
	}

	if color, ok := params["color"].(string); ok && color != "" {
		attachment["color"] = color
	} else {
		attachment["color"] = "good"
	}

	// Add fields if provided
	if fieldsParam, ok := params["fields"]; ok {
		if fieldsMap, ok := fieldsParam.(map[string]interface{}); ok {
			var fields []map[string]interface{}
			for key, value := range fieldsMap {
				fields = append(fields, map[string]interface{}{
					"title": key,
					"value": fmt.Sprintf("%v", value),
					"short": true,
				})
			}
			attachment["fields"] = fields
		}
	}

	// Build main payload
	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{attachment},
	}

	if channel, ok := params["channel"].(string); ok && channel != "" {
		payload["channel"] = channel
	}

	if username, ok := params["username"].(string); ok && username != "" {
		payload["username"] = username
	} else {
		payload["username"] = "Corynth"
	}

	return p.sendToSlack(ctx, webhookURL, payload)
}

func (p *SlackPlugin) executeWorkflowNotification(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	webhookURL, _ := params["webhook_url"].(string)
	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url parameter is required")
	}

	workflowName, _ := params["workflow_name"].(string)
	if workflowName == "" {
		return nil, fmt.Errorf("workflow_name parameter is required")
	}

	status, _ := params["status"].(string)
	if status == "" {
		return nil, fmt.Errorf("status parameter is required")
	}

	// Determine message details based on status
	var color, emoji, statusText string
	switch strings.ToLower(status) {
	case "started":
		color = "#2196F3"
		emoji = ":arrow_forward:"
		statusText = "Started"
	case "success":
		color = "good"
		emoji = ":white_check_mark:"
		statusText = "Completed Successfully"
	case "failed":
		color = "danger"
		emoji = ":x:"
		statusText = "Failed"
	case "warning":
		color = "warning"
		emoji = ":warning:"
		statusText = "Completed with Warnings"
	default:
		color = "#9E9E9E"
		emoji = ":information_source:"
		statusText = status
	}

	// Build fields
	fields := []map[string]interface{}{
		{
			"title": "Status",
			"value": statusText,
			"short": true,
		},
	}

	if duration, ok := params["duration"].(string); ok && duration != "" {
		fields = append(fields, map[string]interface{}{
			"title": "Duration",
			"value": duration,
			"short": true,
		})
	}

	if details, ok := params["details"].(string); ok && details != "" {
		fields = append(fields, map[string]interface{}{
			"title": "Details",
			"value": details,
			"short": false,
		})
	}

	// Build attachment
	attachment := map[string]interface{}{
		"color":  color,
		"title":  fmt.Sprintf("%s Workflow: %s", emoji, workflowName),
		"fields": fields,
		"ts":     time.Now().Unix(),
	}

	// Build payload
	payload := map[string]interface{}{
		"username":    "Corynth Workflow",
		"icon_emoji":  ":gear:",
		"attachments": []map[string]interface{}{attachment},
	}

	if channel, ok := params["channel"].(string); ok && channel != "" {
		payload["channel"] = channel
	}

	return p.sendToSlack(ctx, webhookURL, payload)
}

func (p *SlackPlugin) sendToSlack(ctx context.Context, webhookURL string, payload map[string]interface{}) (map[string]interface{}, error) {
	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)
	responseText := responseBody.String()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	if !success {
		return map[string]interface{}{
			"success":  false,
			"response": responseText,
		}, fmt.Errorf("Slack API returned error: %s (status: %d)", responseText, resp.StatusCode)
	}

	return map[string]interface{}{
		"success":  true,
		"response": responseText,
	}, nil
}