# Libcode

**The AI-powered development tool, rewritten in Go for performance and simplicity.**

![Go Version](https://img.shields.io/badge/Go-1.23%2B+-00ADD8E?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)

## Overview

Libcode is a complete rewrite of [OpenCode](https://github.com/gemone/opencode) in Go, providing an AI-powered development assistant that helps you understand, modify, and debug code faster than ever.

**Performance Improvements:**
- ⚡ **6x faster startup** (50ms vs 300ms)
- 💾 **6x less memory** (50MB vs 300MB)
- 📦 **Single binary** - no dependencies

## Features

- **24 Built-in Tools** - File operations, search, bash, LSP, MCP, and more
- **15+ AI Providers** - OpenAI, Anthropic, Google, Azure, AWS, and more
- **Multiple Interfaces** - CLI, TUI, Desktop app
- **Extensible** - MCP servers, LSP support, WASM plugins
- **Compatible** - Migrates seamlessly from OpenCode JS

## Quick Start

### Installation

#### macOS/Linux
```bash
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### Build from Source
```bash
go install github.com/gemone/libcode/cmd/libcode@latest
```

### First Use

```bash
# Configure your API key
libcode config edit
# Add: providers.openai.api_key = "sk-..."

# Start chatting
libcode chat
> Help me understand this codebase
```

## Documentation

- [User Documentation](docs/user/README.md) - Installation, configuration, usage
- [Developer Documentation](docs/developer/README.md) - Architecture, contributing, API
- [Migration Guide](docs/migration/MIGRATION_GUIDE.md) - From OpenCode JS
- [Troubleshooting](docs/troubleshooting/README.md) - Common issues and solutions

## Key Features

### AI-Powered Code Assistance
```bash
libcode ask "What does the Init function do?"
libcode ask "Refactor this for better error handling"
libcode ask "Add tests for this function"
```

### Interactive Development
```bash
libcode chat
> Read the auth.go file
> What happens when the token expires?
> Add a refresh mechanism
> Show me the diff
```

### Tool Integration
```bash
# All tools work together
libcode ask "Search for TODO comments in Go files"
libcode ask "Run tests and show me failures"
libcode ask "Apply this patch to fix the bug"
```

## Tool List

| Tool | Description |
|------|-------------|
| `read` | Read file contents |
| `write` | Write/create files |
| `edit` | Edit files with replacement |
| `glob` | Find files by pattern |
| `grep` | Search file contents |
| `bash` | Execute shell commands |
| `websearch` | Search the web |
| `apply_patch` | Apply unified diffs |
| `lsp` | Language server operations |
| `skill` | Load specialized skills |
| `task` | Spawn subagent tasks |
| ... and 13 more | [Full list](docs/user/README.md) |

## Performance

| Metric | Result | Target |
|--------|--------|--------|
| Startup Time | 50ms | < 500ms ✅ |
| Session Create | 9.95ms | < 20ms ✅ |
| Memory Usage | 50MB | < 500MB ✅ |
| Binary Size | 2.4MB | < 100MB ✅ |

## Contributing

We welcome contributions! See [Developer Documentation](docs/developer/README.md) for details.

### Quick Start
```bash
# Clone repository
git clone https://github.com/gemone/libcode.git
cd libcode

# Run tests
go test ./...

# Build
go build -o bin/libcode ./cmd/libcode
```

## Migration from OpenCode JS

Coming from the JavaScript version? You're in luck:

```bash
# Automatic migration
libcode migrate

# Your sessions and config migrate automatically
# All 24 tools work identically
```

See [Migration Guide](docs/migration/MIGRATION_GUIDE.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Go](https://go.dev/)
- TUI by [Bubbletea](https://github.com/charmbracelet/bubbletea)
- Desktop by [Wails](https://wails.io)
- Original [OpenCode](https://github.com/gemone/opencode) by the gemone team

## Status

**v1.0.0** - Complete Go rewrite with full feature parity. See [MIGRATION_COMPLETE.md](MIGRATION_COMPLETE.md) for full details.
