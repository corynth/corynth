package pluginv2

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/corynth/corynth/pkg/plugin"
)

// PluginSDK provides a framework for creating Corynth gRPC plugins
type PluginSDK struct {
	plugin plugin.Plugin
}

// NewSDK creates a new plugin SDK instance
func NewSDK(plugin plugin.Plugin) *PluginSDK {
	return &PluginSDK{
		plugin: plugin,
	}
}

// Serve starts the plugin server following Terraform's pattern
// This should be called from the plugin's main function
func (sdk *PluginSDK) Serve() {
	// Check if running in serve mode
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		log.Printf("Usage: %s serve", os.Args[0])
		log.Printf("This is a Corynth plugin. Use 'serve' to start the gRPC server.")
		os.Exit(1)
	}
	
	// Start gRPC server
	if err := ServePlugin(sdk.plugin); err != nil {
		log.Fatalf("Failed to serve plugin: %v", err)
	}
}

// BasePlugin provides a foundation for plugin implementations
type BasePlugin struct {
	name        string
	version     string
	description string
	author      string
	tags        []string
	actions     []plugin.Action
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name, version, description, author string, tags []string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
		author:      author,
		tags:        tags,
		actions:     []plugin.Action{},
	}
}

// Metadata returns plugin metadata
func (p *BasePlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        p.name,
		Version:     p.version,
		Description: p.description,
		Author:      p.author,
		Tags:        p.tags,
	}
}

// Actions returns available actions
func (p *BasePlugin) Actions() []plugin.Action {
	return p.actions
}

// AddAction adds an action to the plugin
func (p *BasePlugin) AddAction(action plugin.Action) {
	p.actions = append(p.actions, action)
}

// Validate provides default validation (can be overridden)
func (p *BasePlugin) Validate(params map[string]interface{}) error {
	// Default validation - check required parameters
	for _, action := range p.actions {
		for paramName, spec := range action.Inputs {
			if spec.Required {
				if _, exists := params[paramName]; !exists {
					return fmt.Errorf("required parameter '%s' is missing", paramName)
				}
			}
		}
	}
	return nil
}

// Execute must be implemented by concrete plugins
func (p *BasePlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Execute method must be implemented by concrete plugin")
}

// PluginBuilder helps construct plugins fluently
type PluginBuilder struct {
	plugin *BasePlugin
}

// NewBuilder creates a new plugin builder
func NewBuilder(name, version string) *PluginBuilder {
	return &PluginBuilder{
		plugin: NewBasePlugin(name, version, "", "", []string{}),
	}
}

// Description sets the plugin description
func (b *PluginBuilder) Description(desc string) *PluginBuilder {
	b.plugin.description = desc
	return b
}

// Author sets the plugin author
func (b *PluginBuilder) Author(author string) *PluginBuilder {
	b.plugin.author = author
	return b
}

// Tags sets the plugin tags
func (b *PluginBuilder) Tags(tags ...string) *PluginBuilder {
	b.plugin.tags = tags
	return b
}

// Action adds an action to the plugin
func (b *PluginBuilder) Action(name, description string) *ActionBuilder {
	return &ActionBuilder{
		pluginBuilder: b,
		action: plugin.Action{
			Name:        name,
			Description: description,
			Inputs:      make(map[string]plugin.InputSpec),
			Outputs:     make(map[string]plugin.OutputSpec),
		},
	}
}

// Build returns the constructed base plugin
func (b *PluginBuilder) Build() *BasePlugin {
	return b.plugin
}

// ActionBuilder helps construct actions fluently
type ActionBuilder struct {
	pluginBuilder *PluginBuilder
	action        plugin.Action
}

// Input adds an input parameter to the action
func (ab *ActionBuilder) Input(name, paramType, description string, required bool) *ActionBuilder {
	ab.action.Inputs[name] = plugin.InputSpec{
		Type:        paramType,
		Description: description,
		Required:    required,
	}
	return ab
}

// InputWithDefault adds an input parameter with a default value
func (ab *ActionBuilder) InputWithDefault(name, paramType, description string, defaultVal interface{}) *ActionBuilder {
	ab.action.Inputs[name] = plugin.InputSpec{
		Type:        paramType,
		Description: description,
		Required:    false,
		Default:     defaultVal,
	}
	return ab
}

// Output adds an output parameter to the action
func (ab *ActionBuilder) Output(name, paramType, description string) *ActionBuilder {
	ab.action.Outputs[name] = plugin.OutputSpec{
		Type:        paramType,
		Description: description,
	}
	return ab
}

// Add completes the action and adds it to the plugin
func (ab *ActionBuilder) Add() *PluginBuilder {
	ab.pluginBuilder.plugin.AddAction(ab.action)
	return ab.pluginBuilder
}