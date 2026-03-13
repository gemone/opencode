// Package example provides an example plugin for libcode
package example

import (
	"context"
	"fmt"
	"time"

	"github.com/gemone/libcode/internal/plugin"
	"github.com/gemone/libcode/internal/tool/framework"
)

// ExamplePlugin demonstrates a simple plugin
type ExamplePlugin struct {
	*plugin.BasePlugin
}

// NewExamplePlugin creates a new example plugin
func NewExamplePlugin() *ExamplePlugin {
	base := plugin.NewBasePlugin(
		"example",
		"1.0.0",
		"An example plugin demonstrating the plugin system",
	)

	p := &ExamplePlugin{BasePlugin: base}

	// Add a simple tool
	p.AddTool(framework.NewBaseTool(
		"example.hello",
		"Says hello to the world",
		framework.NewSchema().AddProperty("name", framework.StringProperty("Your name")),
		func(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
			name, _ := params["name"].(string)
			if name == "" {
				name = "World"
			}
			return &framework.Result{
				Output: fmt.Sprintf("Hello, %s! from example plugin", name),
				Metadata: map[string]any{
					"plugin":  "example",
					"version": "1.0.0",
				},
			}, nil
		},
	))

	// Set up event hook
	p.SetEventHook(func(ctx context.Context, event *plugin.Event) error {
		fmt.Printf("[Example Plugin] Received event: %s\n", event.Type)
		return nil
	})

	// Set up shell environment hook
	p.SetShellEnvHook(func(ctx context.Context, cwd string, env map[string]string) (map[string]string, error) {
		// Add a custom environment variable
		result := make(map[string]string)
		for k, v := range env {
			result[k] = v
		}
		result["EXAMPLE_PLUGIN"] = "1"
		return result, nil
	})

	// Set up tool before hook
	p.SetToolBeforeHook(func(ctx context.Context, tool string, callID string, args map[string]any) (map[string]any, error) {
		if tool == "bash" {
			fmt.Printf("[Example Plugin] Before bash tool execution: %s\n", callID)
			// Could modify args here
		}
		return nil, nil
	})

	// Set up tool after hook
	p.SetToolAfterHook(func(ctx context.Context, tool string, callID string, args map[string]any, result *plugin.ToolResult) error {
		if tool == "bash" {
			fmt.Printf("[Example Plugin] After bash tool execution: %s\n", callID)
			result.Metadata = map[string]any{
				"processed_by": "example_plugin",
				"timestamp":    time.Now().Unix(),
			}
		}
		return nil
	})

	return p
}

// Init initializes the plugin
func (p *ExamplePlugin) Init(ctx context.Context, manager *plugin.Manager) error {
	fmt.Printf("[Example Plugin] Initializing plugin %s v%s\n", p.Name(), p.Version())
	return nil
}

// Cleanup cleans up plugin resources
func (p *ExamplePlugin) Cleanup() error {
	fmt.Printf("[Example Plugin] Cleaning up plugin %s\n", p.Name())
	return nil
}
