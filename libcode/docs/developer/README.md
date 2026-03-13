# Libcode Developer Documentation

## Architecture Overview

Libcode is built with Go 1.23+ using a modular architecture:

```
libcode/
├── cmd/                    # CLI entry points
│   ├── libcode/            # Main CLI application
│   └── libcode-desktop/    # Desktop application (Wails)
├── internal/               # Private packages
│   ├── ai/                 # AI abstraction layer
│   │   ├── provider/       # AI provider implementations
│   │   └── streaming/      # Streaming infrastructure
│   ├── tool/               # Tool system
│   │   ├── builtin/        # Built-in tools (24 tools)
│   │   └── framework/      # Tool framework
│   ├── mcp/                # MCP protocol client
│   ├── lsp/                # LSP client
│   ├── pty/                # PTY handling
│   ├── session/            # Session management
│   ├── config/             # Configuration management
│   ├── storage/            # Database layer (SQLite)
│   ├── event/              # Event bus
│   ├── logger/             # Structured logging
│   └── plugin/             # Plugin system (WASM)
├── pkg/                    # Public packages
│   ├── ui/                 # TUI components (Bubbletea)
│   └── api/                # Shared APIs
├── desktop/                # Desktop app (Wails)
│   └── frontend/           # React + TypeScript UI
└── tests/                  # Tests
    ├── unit/               # Unit tests
    ├── integration/        # Integration tests
    ├── e2e/                # End-to-end tests
    └── benchmark/          # Performance benchmarks
```

## Core Components

### 1. AI Abstraction Layer (`internal/ai/`)

Provides a unified API across all AI providers.

**Interface:**
```go
type Provider interface {
    // Complete sends a request and waits for the full response
    Complete(ctx context.Context, req *Request) (*Response, error)

    // Stream sends a request and streams the response
    Stream(ctx context.Context, req *Request) <-chan StreamingChunk

    // ToolCall executes a tool call (if supported)
    ToolCall(ctx context.Context, req *Request, tool Tool) (*ToolCallResult, error)
}
```

**Supported Providers:**
- OpenAI (GPT-4, GPT-3.5, o1)
- Anthropic (Claude 3.5 Sonnet, Haiku, Opus)
- Google (Gemini, Vertex AI)
- Azure OpenAI
- AWS Bedrock
- OpenRouter
- And 10+ more

### 2. Tool System (`internal/tool/`)

**Tool Framework** (`internal/tool/framework/`)
```go
type Tool interface {
    ID() string                                    // Unique identifier
    Description() string                           // What the tool does
    Parameters() *Schema                          // Input schema
    Execute(ctx context.Context, params map[string]any, execCtx *ExecutionContext) (*Result, error)
}
```

**Built-in Tools** (24 total):
- **File Operations**: read, write, edit, glob, ls
- **Search**: grep, codesearch, websearch, webfetch
- **Execution**: bash, task
- **Advanced**: apply_patch, multiedit, lsp, mcp
- **AI/Agent**: question, plan, skill, task
- **Utilities**: todo, external_directory, invalid, truncation, batch

**Adding a New Tool:**
```go
// 1. Create file in internal/tool/builtin/mytool.go
package builtin

type MyTool struct {
    workingDir string
}

func NewMyTool(workingDir string) *MyTool {
    return &MyTool{workingDir: workingDir}
}

func (t *MyTool) ID() string {
    return "mytool"
}

func (t *MyTool) Description() string {
    return "Description of what mytool does"
}

func (t *MyTool) Parameters() *framework.Schema {
    return framework.NewSchema().
        AddProperty("param1", framework.Property{
            Type:        "string",
            Description: "Parameter description",
            Required:    true,
        })
}

func (t *MyTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
    // Tool implementation
    return &framework.Result{
        Title:  "Result",
        Output: "Tool output",
    }, nil
}

// 2. Register in internal/tool/builtin/registry.go
func NewRegistry(workingDir string) *Registry {
    r := &Registry{workingDir: workingDir}
    r.Register(NewMyTool(workingDir))
    // ... other tools
    return r
}
```

### 3. MCP Client (`internal/mcp/`)

Implements the Model Context Protocol for tool extensibility.

**Features:**
- stdio transport (local servers)
- SSE transport (remote servers)
- Tool discovery and calling
- Resource access
- Prompt templates
- OAuth authentication

**Using MCP Servers:**
```yaml
mcp:
  servers:
    - name: "filesystem"
      command: "npx"
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/allowed/path"]
```

### 4. LSP Client (`internal/lsp/`)

Provides code intelligence features.

**Supported Features:**
- Go to definition
- Find references
- Hover information
- Document symbols
- Workspace symbols
- Diagnostics
- Code actions
- Rename
- Completion
- Signature help

**Supported Languages:**
- TypeScript/JavaScript
- Go
- Python
- Rust
- And more via configuration

### 5. Session Management (`internal/session/`)

Manages conversation state and history.

**Session Structure:**
```go
type Session struct {
    ID        string
    Title     string
    Model     string
    Messages  []Message
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Storage:**
- SQLite database (`~/.libcode/sessions/sessions.db`)
- Data compatible with JavaScript version
- Migration support from JS schema

### 6. Configuration (`internal/config/`)

Uses Viper for flexible configuration management.

**Config Loading Order:**
1. Flags (command line)
2. Environment variables
3. Config file (`libcode.yaml`)
4. Defaults

**Configuration Schema:**
```go
type Config struct {
    DefaultModel string                 `json:"default_model"`
    ModelVariant string                 `json:"model_variant"`
    SessionDir   string                 `json:"session_dir"`
    MaxHistory   int                    `json:"max_history"`
    LogLevel     string                 `json:"log_level"`
    LogFile      string                 `json:"log_file"`
    LSP          LSPConfig              `json:"lsp"`
    MCP          MCPConfig              `json:"mcp"`
    Features     FeatureConfig          `json:"features"`
    Providers    map[string]ProviderCfg `json:"providers"`
}
```

## Design Decisions

### Why Go?

**Performance:**
- Compiled binary, fast startup (~50ms cold)
- Low memory footprint (~50MB idle)
- Efficient concurrency (goroutines)

**Distribution:**
- Single binary, no dependencies
- Cross-platform builds
- Small binary size (~2-5MB CLI, ~14MB desktop)

**Type Safety:**
- Strong typing catches errors at compile time
- Excellent tooling (gopls, go vet)
- Easy refactoring

### Why SQLite for Sessions?

- Zero configuration
- Embedded (no separate server)
- ACID compliant
- Cross-platform
- Data compatibility with JS version

### Why Bubbletea for TUI?

- Pure Go, no CGo
- Elegant Elm architecture
- Excellent accessibility story
- Easy testing

### Why Wails for Desktop?

- Native performance (no Electron overhead)
- Go backend + Web frontend
- Smaller bundle size than Electron (~14MB vs ~80MB)
- Cross-platform (macOS, Windows, Linux)

## Development Workflow

### Prerequisites

```bash
go version  # Go 1.23+
```

### Setup

```bash
# Clone repository
git clone https://github.com/gemone/libcode.git
cd libcode

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o bin/libcode ./cmd/libcode
```

### Running Tests

```bash
# All tests
go test ./...

# Unit tests only
go test ./internal/...

# Integration tests
go test ./tests/integration/...

# E2E tests
go test ./tests/e2e/...

# Benchmarks
go test ./tests/benchmark/... -bench=. -benchmem

# With coverage
go test -cover ./...
```

### Building

```bash
# CLI binary
go build -o bin/libcode ./cmd/libcode

# Desktop app
cd desktop && wails build

# Multi-platform builds
make build-all
```

### Code Style

We use standard Go formatting:

```bash
gofmt -w .
go vet ./...
golangci-lint run
```

## Contributing

### Making Changes

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```
3. **Make your changes**
4. **Write tests**
   ```bash
   # Add tests in <package>_test.go
   go test ./...
   ```
5. **Ensure all tests pass**
   ```bash
   go test ./...
   go vet ./...
   ```
6. **Commit with conventional commits**
   ```bash
   git commit -m "feat: add new tool for X"
   ```
7. **Push and create PR**
   ```bash
   git push origin feature/my-feature
   ```

### Pull Request Guidelines

- Describe what your PR does
- Link related issues
- Include tests for new features
- Update documentation if needed
- Ensure CI passes

### Code Review Process

1. Automated checks (CI)
2. Peer review (at least one approval)
3. Merge by maintainers

## API Reference

### Tool API

**Creating Tools:**

See "Adding a New Tool" above in Core Components.

**Tool Execution Context:**

```go
type ExecutionContext struct {
    SessionID    string                 // Current session ID
    MessageID    string                 // Current message ID
    CallID       string                 // Tool call ID
    Agent        string                 // Agent name
    Abort        context.CancelFunc     // Cancel function
    Messages     []session.Message      // Message history
    Extra        map[string]any         // Extra context
    WorkingDir   string                 // Working directory
    PermissionAsker PermissionAsker      // Permission handler
    QuestionAsker QuestionAsker        // Question handler
}
```

### AI Provider API

**Implementing a New Provider:**

```go
package myprovider

import (
    "github.com/gemone/libcode/internal/ai"
)

type MyProvider struct {
    apiKey string
    baseURL string
    model   string
}

func NewMyProvider(apiKey, baseURL, model string) *MyProvider {
    return &MyProvider{
        apiKey: apiKey,
        baseURL: baseURL,
        model:   model,
    }
}

func (p *MyProvider) Complete(ctx context.Context, req *ai.Request) (*ai.Response, error) {
    // Implement completion
}

func (p *MyProvider) Stream(ctx context.Context, req *ai.Request) <-chan ai.StreamingChunk {
    // Implement streaming
}

// Register in provider registry
func init() {
    provider.Register("myprovider", func(cfg ProviderConfig) (ai.Provider, error) {
        return NewMyProvider(cfg.APIKey, cfg.BaseURL, cfg.Model), nil
    })
}
```

### Session API

**Creating Sessions:**

```go
import "github.com/gemone/libcode/internal/storage"

db, err := storage.Open("path/to/sessions.db")
if err != nil {
    return err
}
defer db.Close()

session := &storage.Session{
    ID:        "session-123",
    Title:     "My Session",
    Model:     "openai/gpt-4",
    Messages:  []storage.Message{...},
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

err = db.CreateSession(session)
```

**Retrieving Sessions:**

```go
session, err := db.GetSession("session-123")
if err != nil {
    return err
}

sessions, err := db.ListSessions(100)
if err != nil {
    return err
}
```

## Plugin Development

### WASM Plugins

Libcode supports WebAssembly plugins for extensibility.

**Plugin Structure:**

```rust
// plugin.wat (WebAssembly Text)
(module
  (export "execute" (func $execute))
  (export "metadata" (func $metadata))
)
```

**Or compile from Rust:**

```rust
// lib.rs
#[no_mangle]
pub extern "C" fn execute(input_ptr: *const u8, input_len: usize) -> usize {
    // Plugin logic
}
```

**Loading Plugins:**

```yaml
plugins:
  - name: "my-plugin"
    path: "/path/to/plugin.wasm"
    enabled: true
```

### Native Go Plugins

For trusted plugins, you can write native Go code:

```go
// internal/plugin/myplugin/myplugin.go
package myplugin

type MyPlugin struct {
    config map[string]any
}

func (p *MyPlugin) Execute(ctx context.Context, input string) (string, error) {
    // Plugin logic
    return "result", nil
}
```

## Testing

### Unit Tests

```go
func TestMyTool_Execute(t *testing.T) {
    tool := NewMyTool("/tmp")
    ctx := context.Background()
    params := map[string]any{"param1": "value"}

    result, err := tool.Execute(ctx, params, nil)

    assert.NoError(t, err)
    assert.Equal(t, "expected", result.Output)
}
```

### Integration Tests

```go
func TestDatabase_ConcurrentAccess(t *testing.T) {
    tempDir := t.TempDir()
    db, err := storage.Open(filepath.Join(tempDir, "test.db"))
    require.NoError(t, err)
    defer db.Close()

    // Test concurrent access
    // ...
}
```

### Benchmarks

```go
func BenchmarkMyTool(b *testing.B) {
    tool := NewMyTool("/tmp")
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        tool.Execute(ctx, map[string]any{"param": "value"}, nil)
    }
}
```

## Performance Considerations

### Streaming

Always use streaming for AI responses to reduce perceived latency:

```go
chunks := provider.Stream(ctx, req)
for chunk := range chunks {
    if chunk.Error != nil {
        return err
    }
    fmt.Print(chunk.Content)
}
```

### Database Connections

Use connection pooling for concurrent access:

```go
db, err := storage.Open("file:sessions.db?cache=shared&mode=rwc")
```

### Memory Management

- Limit session history (default: 100 messages)
- Use streaming for large responses
- Close unused sessions

### Caching

AI providers may cache responses. Configure cache duration:

```yaml
providers:
  openai:
    cache_ttl: 300  # Cache for 5 minutes
```

## Debugging

### Debug Logging

```go
import "github.com/gemone/libcode/internal/logger"

logger.Debug("Detailed debug info", zap.String("key", "value"))
logger.Info("Info message", zap.String("key", "value"))
logger.Warn("Warning message", zap.String("key", "value"))
logger.Error("Error message", zap.Error(err))
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### Tracing

Libcode uses OpenTelemetry for distributed tracing:

```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("libcode")
ctx, span := tracer.Start(ctx, "operation")
defer span.End()
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Wails Documentation](https://wails.io/docs)
- [SQLite in Go](https://gitlab.com/cznic/sqlite)
- [Project Board](https://github.com/gemone/libcode/projects)
