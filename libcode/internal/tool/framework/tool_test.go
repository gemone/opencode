package framework

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseTool(t *testing.T) {
	tool := NewBaseTool(
		"test-tool",
		"A test tool",
		NewSchema().AddProperty("input", StringProperty("Test input").WithRequired()),
		func(ctx context.Context, params map[string]any, execCtx *ExecutionContext) (*Result, error) {
			return &Result{
				Title:  "Test Result",
				Output: params["input"].(string),
				Metadata: map[string]any{
					"test": true,
				},
			}, nil
		},
	)

	assert.Equal(t, "test-tool", tool.ID())
	assert.Equal(t, "A test tool", tool.Description())
	assert.Len(t, tool.Parameters().Properties, 1)

	ctx := context.Background()
	params := map[string]any{"input": "hello"}
	execCtx := &ExecutionContext{}

	result, err := tool.Execute(ctx, params, execCtx)
	require.NoError(t, err)
	assert.Equal(t, "Test Result", result.Title)
	assert.Equal(t, "hello", result.Output)
	assert.True(t, result.Metadata["test"].(bool))
}

func TestValidateParams(t *testing.T) {
	schema := NewSchema().
		AddProperty("required_string", StringProperty("A required string").WithRequired()).
		AddProperty("optional_number", IntegerProperty("An optional number")).
		AddProperty("enum_field", StringProperty("An enum field").WithEnum("a", "b", "c"))

	tests := []struct {
		name    string
		params  map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: map[string]any{
				"required_string": "test",
				"optional_number": 42,
			},
			wantErr: false,
		},
		{
			name: "missing required",
			params: map[string]any{
				"optional_number": 42,
			},
			wantErr: true,
			errMsg:  "required_string",
		},
		{
			name: "wrong type",
			params: map[string]any{
				"required_string": 123, // should be string
			},
			wantErr: true,
			errMsg:  "expected string",
		},
		{
			name: "unknown field",
			params: map[string]any{
				"required_string": "test",
				"unknown_field":   "value",
			},
			wantErr: true,
			errMsg:  "unknown parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParams(tt.params, schema)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	tool1 := NewBaseTool("tool1", "Tool 1", NewSchema(), nil)
	tool2 := NewBaseTool("tool2", "Tool 2", NewSchema(), nil)

	// Test registration
	err := registry.Register(tool1)
	require.NoError(t, err)

	err = registry.Register(tool2)
	require.NoError(t, err)

	// Test duplicate registration
	err = registry.Register(tool1)
	assert.Error(t, err)

	// Test Get
	retrieved, err := registry.Get("tool1")
	require.NoError(t, err)
	assert.Equal(t, tool1.ID(), retrieved.ID())

	// Test Get non-existent
	_, err = registry.Get("nonexistent")
	assert.Error(t, err)

	// Test List
	ids := registry.List()
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, "tool1")
	assert.Contains(t, ids, "tool2")

	// Test Count
	assert.Equal(t, 2, registry.Count())

	// Test Unregister
	err = registry.Unregister("tool1")
	require.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	// Test Clear
	registry.Clear()
	assert.Equal(t, 0, registry.Count())
}

func TestRegistryAlias(t *testing.T) {
	registry := NewRegistry()

	tool := NewBaseTool("my-tool", "My Tool", NewSchema(), nil)
	err := registry.Register(tool)
	require.NoError(t, err)

	err = registry.RegisterAlias("alias1", "my-tool")
	require.NoError(t, err)

	// Get by alias
	retrieved, err := registry.Get("alias1")
	require.NoError(t, err)
	assert.Equal(t, "my-tool", retrieved.ID())

	// Alias to non-existent tool
	err = registry.RegisterAlias("bad-alias", "nonexistent")
	assert.Error(t, err)
}

func TestRegistryGroups(t *testing.T) {
	registry := NewRegistry()

	tool1 := NewBaseTool("tool1", "Tool 1", NewSchema(), nil)
	tool2 := NewBaseTool("tool2", "Tool 2", NewSchema(), nil)
	tool3 := NewBaseTool("tool3", "Tool 3", NewSchema(), nil)

	registry.Register(tool1)
	registry.Register(tool2)
	registry.Register(tool3)

	err := registry.RegisterGroup("group1", "tool1", "tool2")
	require.NoError(t, err)

	tools := registry.ListByGroup("group1")
	assert.Len(t, tools, 2)
	assert.Contains(t, tools, "tool1")
	assert.Contains(t, tools, "tool2")

	// Group with non-existent tool
	err = registry.RegisterGroup("group2", "tool1", "nonexistent")
	assert.Error(t, err)
}

func TestRegistryFind(t *testing.T) {
	registry := NewRegistry()

	tool1 := NewBaseTool("file-read", "Read File", NewSchema(), nil)
	tool2 := NewBaseTool("file-write", "Write File", NewSchema(), nil)
	tool3 := NewBaseTool("bash-exec", "Execute Bash", NewSchema(), nil)

	registry.Register(tool1)
	registry.Register(tool2)
	registry.Register(tool3)

	// Find by pattern
	results := registry.Find("file")
	assert.Len(t, results, 2)
	assert.Contains(t, results, "file-read")
	assert.Contains(t, results, "file-write")

	results = registry.Find("bash")
	assert.Len(t, results, 1)
	assert.Contains(t, results, "bash-exec")
}

func TestSchemaBuilder(t *testing.T) {
	tagProp := StringProperty("Tag")
	schema := NewSchema().
		AddProperty("name", StringProperty("The name").WithRequired()).
		AddProperty("age", IntegerProperty("The age").WithMin(0).WithMax(150)).
		AddProperty("active", BooleanProperty("Is active").WithDefault(false)).
		AddProperty("tags", ArrayProperty("Tags", &tagProp))

	assert.Len(t, schema.Properties, 4)
	assert.Len(t, schema.Required, 1) // only "name" is required

	// Check string property
	nameProp := schema.Properties["name"]
	assert.Equal(t, "string", nameProp.Type)
	assert.True(t, nameProp.Required)

	// Check number property with range
	ageProp := schema.Properties["age"]
	assert.Equal(t, "integer", ageProp.Type)
	assert.NotNil(t, ageProp.Min)
	assert.Equal(t, 0.0, *ageProp.Min)
	assert.NotNil(t, ageProp.Max)
	assert.Equal(t, 150.0, *ageProp.Max)

	// Check boolean with default
	activeProp := schema.Properties["active"]
	assert.Equal(t, "boolean", activeProp.Type)
	assert.False(t, activeProp.Required)
	assert.Equal(t, false, activeProp.Default)

	// Check array property
	tagsProp := schema.Properties["tags"]
	assert.Equal(t, "array", tagsProp.Type)
	assert.NotNil(t, tagsProp.Items)
}

func TestPermissionManager(t *testing.T) {
	asker := &mockAsker{granted: make(map[string]bool)}
	pm := NewPermissionManager(asker)

	ctx := context.Background()
	sessionID := "test-session"

	// First request should ask
	req := &PermissionRequest{
		Permission: "read",
		Patterns:   []string{"/tmp/file.txt"},
		Always:     []string{"*"},
	}

	err := pm.Ask(ctx, sessionID, req)
	require.NoError(t, err)
	assert.True(t, asker.granted["/tmp/file.txt"])

	// Second request should be cached
	asker.granted = make(map[string]bool) // reset
	err = pm.Ask(ctx, sessionID, req)
	require.NoError(t, err)
	assert.False(t, asker.granted["/tmp/file.txt"]) // didn't ask

	// Different session should ask again
	otherSession := "other-session"
	err = pm.Ask(ctx, otherSession, req)
	require.NoError(t, err)
	assert.True(t, asker.granted["/tmp/file.txt"])

	// Test dangerous paths
	pm.Clear(sessionID) // Clear previous grants
	pm.AddDangerousPath("/etc/passwd")
	dangerousReq := &PermissionRequest{
		Permission: "read",
		Patterns:   []string{"/etc/passwd"},
		Always:     []string{"*"},
	}

	err = pm.Ask(ctx, sessionID, dangerousReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dangerous")
}

func TestExecuteWithTimeout(t *testing.T) {
	tool := NewBaseTool(
		"fast-tool",
		"Fast tool",
		NewSchema(),
		func(ctx context.Context, params map[string]any, execCtx *ExecutionContext) (*Result, error) {
			return &Result{Title: "Done", Output: "fast"}, nil
		},
	)

	ctx := context.Background()
	result, err := ExecuteWithTimeout(ctx, tool, nil, &ExecutionContext{}, 1000)
	require.NoError(t, err)
	assert.Equal(t, "Done", result.Title)
}

func TestCommonSchemas(t *testing.T) {
	// Test file path schema
	fp := FilePathSchema()
	assert.Equal(t, "string", fp.Type)
	assert.Equal(t, "uri", fp.Format)

	// Test timeout schema
	ts := TimeoutSchema()
	assert.Equal(t, "integer", ts.Type)
	assert.Equal(t, 120000, ts.Default)
	assert.Equal(t, 0.0, *ts.Min)

	// Test limit schema
	ls := LimitSchema()
	assert.Equal(t, "integer", ls.Type)
	assert.Equal(t, 100, ls.Default)
	assert.Equal(t, 1.0, *ls.Min)
}

// mockAsker is a test implementation of PermissionAsker
type mockAsker struct {
	granted map[string]bool
}

func (m *mockAsker) Ask(req *PermissionRequest) error {
	for _, pattern := range req.Patterns {
		m.granted[pattern] = true
	}
	return nil
}
