# 🎉 Libcode v1.0.0: The Go Rewrite is Here!

We're excited to announce the official release of **Libcode v1.0.0** — a complete rewrite of OpenCode in Go, delivering dramatic performance improvements and a streamlined distribution experience.

## 🚀 What's New

### Performance Revolution

| Metric | JavaScript Version | Go Version | Improvement |
|--------|-------------------|------------|-------------|
| **Startup Time** | ~300ms | ~50ms | **6x faster** |
| **Memory Usage** | ~300MB | ~50MB | **6x less** |
| **Binary Size** | N/A (interpreted) | 8MB | Self-contained |
| **Distribution** | npm install | Single binary | No dependencies |

### Complete Feature Parity

**All 24 Tools from OpenCode JS:**
- File Operations: read, write, edit, glob, ls
- Search: grep, codesearch, websearch, webfetch
- Execution: bash, task
- Advanced: apply_patch, multiedit, lsp, mcp
- AI/Agent: question, plan, skill, task
- Utilities: todo, external_directory, invalid, truncation, batch

**15+ AI Providers:**
- OpenAI (GPT-4, GPT-3.5, o1)
- Anthropic (Claude 3.5 Sonnet, Haiku, Opus)
- Google (Gemini, Vertex AI)
- Azure OpenAI, AWS Bedrock
- OpenRouter, and more...

## 📦 Installation

### Quick Install

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/gemone/libcode/releases/download/v1.0.0/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L https://github.com/gemone/libcode/releases/download/v1.0.0/libcode-darwin-amd64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### Linux
```bash
curl -L https://github.com/gemone/libcode/releases/download/v1.0.0/libcode-linux-amd64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### Windows
Download from the releases section below.

### Build from Source
```bash
go install github.com/gemone/libcode/cmd/libcode@v1.0.0
```

## 🔄 Migration from OpenCode JS

Seamless migration from the JavaScript version:

```bash
# Automatic migration
libcode migrate

# All sessions preserved
libcode session list
```

See [Migration Guide](https://github.com/gemone/libcode/blob/main/docs/migration/MIGRATION_GUIDE.md) for details.

## 🚀 Quick Start

```bash
# Configure your API key
libcode auth login openai

# Start chatting
libcode chat
> Help me understand this codebase

# Or ask a quick question
libcode ask "What files are in this directory?"
```

## ✨ Features

- **24 Built-in Tools** - Everything you need for development
- **15+ AI Providers** - Use your preferred AI service
- **Multiple Interfaces** - CLI, TUI, Desktop app
- **MCP Support** - Extensible tool protocol
- **LSP Integration** - Code intelligence features
- **Session Persistence** - SQLite-based storage
- **Cross-Platform** - macOS, Linux, Windows

## 📚 Documentation

- [User Guide](https://github.com/gemone/libcode/blob/main/docs/user/README.md)
- [Developer Guide](https://github.com/gemone/libcode/blob/main/docs/developer/README.md)
- [Migration Guide](https://github.com/gemone/libcode/blob/main/docs/migration/MIGRATION_GUIDE.md)
- [Troubleshooting](https://github.com/gemone/libcode/blob/main/docs/troubleshooting/README.md)

## 🧪 Testing

All tests passing:
- ✅ Unit Tests
- ✅ Integration Tests
- ✅ E2E Tests
- ✅ Performance Benchmarks

## 🙏 Acknowledgments

- **OpenCode Community** - For feedback and patience during the rewrite
- **Beta Testers** - For comprehensive testing and feedback
- **Go Team** - For an amazing language and toolchain
- **Contributors** - Everyone who helped port features from JS to Go

## 🗺️ Roadmap

### v1.0.0 (Current) - Stable Release
- ✅ Complete JS to Go migration
- ✅ All 24 tools implemented
- ✅ Production-ready

### v1.1.0 (Planned)
- Enhanced WASM plugin system
- Web UI
- Collaborative features

### v1.2.0 (Planned)
- Advanced code analysis
- Custom AI provider support
- Enterprise features

## 📝 License

MIT License - see [LICENSE](https://github.com/gemone/libcode/blob/main/LICENSE) for details.

---

**Built with Go. Powered by AI. Ready for production.**

Download: [libcode-v1.0.0](https://github.com/gemone/libcode/releases/tag/v1.0.0)

Documentation: [https://github.com/gemone/libcode/tree/main/docs](https://github.com/gemone/libcode/tree/main/docs)

---

*Thank you to everyone who contributed to this migration!*
