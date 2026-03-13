# Libcode

**The AI-powered development tool, rewritten in Go for performance and simplicity.**

![Go Version](https://img.shields.io/badge/Go-1.23%2B+-00ADD8E?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)

## Overview

Libcode is a complete rewrite of [OpenCode](https://github.com/gemone/opencode) in Go, providing an AI-powered development assistant that helps you understand, modify, and debug code faster than ever.

## ⚡ Performance

| Metric | JavaScript Version | Go Version | Improvement |
|--------|-------------------|------------|-------------|
| **Startup Time** | ~300ms | ~50ms | **6x faster** |
| **Memory Usage** | ~300MB | ~50MB | **6x less** |
| **Binary Size** | N/A (interpreted) | 8MB | Self-contained |

## ✨ Features

- **24 Built-in Tools** - File operations, search, bash, LSP, MCP, and more
- **15+ AI Providers** - OpenAI, Anthropic, Google, Azure, AWS, and more
- **Multiple Interfaces** - CLI, TUI, Desktop app
- **Extensible** - MCP servers, LSP support, WASM plugins
- **Compatible** - Migrates seamlessly from OpenCode JS

## 📦 Installation

### Quick Install

#### macOS / Linux
```bash
go install github.com/gemone/libcode/cmd/libcode@latest
```

#### From Binary
```bash
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

## 🚀 Quick Start

```bash
# Configure your API key
libcode auth login openai

# Start chatting
libcode chat
> Help me understand this codebase
```

## 📚 Documentation

- [User Guide](docs/user/README.md) - Installation, configuration, usage
- [Developer Guide](docs/developer/README.md) - Architecture, contributing, API
- [Migration Guide](docs/migration/MIGRATION_GUIDE.md) - From OpenCode JS
- [Troubleshooting](docs/troubleshooting/README.md) - Common issues and solutions

## 🔄 Migration from OpenCode JS

```bash
# Automatic migration
libcode migrate

# All sessions preserved
libcode session list
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run benchmarks
go test ./tests/benchmark/... -bench=.
```

## 🗺️ Roadmap

### v1.0.0 (Current) - Stable Release
- ✅ Complete JS to Go migration
- ✅ All 24 tools implemented
- ✅ Production-ready

### v1.1.0 (Planned)
- Enhanced WASM plugin system
- Web UI
- Collaborative features

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- TUI powered by [Bubbletea](https://github.com/charmbracelet/bubbletea)
- Desktop app built with [Wails](https://wails.io)
- Original project: [OpenCode](https://github.com/gemone/opencode)

## 📮 Support

- **GitHub:** https://github.com/gemone/libcode
- **Discord:** https://discord.gg/libcode
- **Issues:** https://github.com/gemone/libcode/issues

---

**Built with Go. Powered by AI. Ready for production.**
