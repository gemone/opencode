// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"sync"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	maxBatchTools = 25
)

// Disallowed tools in batch mode
var disallowedInBatch = map[string]bool{
	"batch": true,
}

// BatchTool executes multiple tool calls in parallel
type BatchTool struct {
	registry *framework.Registry
}

// NewBatchTool creates a new batch tool
func NewBatchTool(registry *framework.Registry) *BatchTool {
	return &BatchTool{registry: registry}
}

// ID returns the tool identifier
func (t *BatchTool) ID() string {
	return "batch"
}

// Description returns the tool description
func (t *BatchTool) Description() string {
	return "Execute multiple tool calls in parallel for optimal performance. Up to 25 tools can be batched together."
}

// Parameters returns the parameter schema
func (t *BatchTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("tool_calls", framework.Property{
			Type:        "array",
			Description: "Array of tool calls to execute in parallel (max 25)",
			Required:    true,
		})
}

// Execute executes the batch tool calls
func (t *BatchTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	rawToolCalls, ok := params["tool_calls"].([]any)
	if !ok {
		return nil, fmt.Errorf("tool_calls must be an array")
	}

	// Limit to max batch size
	limit := len(rawToolCalls)
	if limit > maxBatchTools {
		limit = maxBatchTools
	}

	toolCalls := rawToolCalls[:limit]
	var results []*batchResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Execute tool calls in parallel
	for i, rawCall := range toolCalls {
		call, ok := rawCall.(map[string]any)
		if !ok {
			results = append(results, &batchResult{
				index:  i,
				success: false,
				error:  fmt.Errorf("tool call at index %d is not an object", i),
			})
			continue
		}

		toolName, _ := call["tool"].(string)
		callParams, _ := call["parameters"].(map[string]any)

		wg.Add(1)
		go func(idx int, name string, params map[string]any) {
			defer wg.Done()

			result := &batchResult{index: idx, tool: name}

			// Check if disallowed
			if disallowedInBatch[name] {
				result.success = false
				result.error = fmt.Errorf("tool '%s' is not allowed in batch", name)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				return
			}

			// Find tool
			tool, err := t.registry.Get(name)
			if err != nil {
				result.success = false
				result.error = fmt.Errorf("tool '%s' not found in registry: %w", name, err)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				return
			}

			// Execute tool
			output, err := tool.Execute(ctx, params, execCtx)
			if err != nil {
				result.success = false
				result.error = err
			} else {
				result.success = true
				result.output = output
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(i, toolName, callParams)
	}

	wg.Wait()

	// Count successes and failures
	successful := 0
	failed := 0
	for _, r := range results {
		if r.success {
			successful++
		} else {
			failed++
		}
	}

	var message string
	if failed > 0 {
		message = fmt.Sprintf("Executed %d/%d tools successfully. %d failed.", successful, len(results), failed)
	} else {
		message = fmt.Sprintf("All %d tools executed successfully.", successful)
	}

	return &framework.Result{
		Title:  fmt.Sprintf("Batch execution (%d/%d successful)", successful, len(results)),
		Output: message,
		Metadata: map[string]any{
			"totalCalls": len(results),
			"successful": successful,
			"failed":     failed,
			"details":    results,
		},
	}, nil
}

type batchResult struct {
	index   int
	tool    string
	success bool
	output  *framework.Result
	error   error
}
