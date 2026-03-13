# Libcode v1.0.0 - Official Launch Announcement

**Release Date:** April 1, 2026
**Status:** Ready for Launch

---

## 🎉 Announcing Libcode v1.0.0: The AI-Powered Development Tool, Rewritten in Go

We're thrilled to announce the official release of **Libcode v1.0.0** — a complete rewrite of OpenCode in Go, delivering dramatic performance improvements and a streamlined distribution experience.

---

## What is Libcode?

Libcode is an AI-powered development assistant that helps you:
- **Understand codebases** quickly with intelligent context awareness
- **Write better code** with AI-assisted development
- **Debug efficiently** with powerful search and analysis tools
- **Automate tasks** with 24 built-in tools and extensibility

---

## 🚀 Why Go? The Performance Revolution

### 6x Faster Startup
- **Before (Node.js):** ~300ms cold start
- **After (Go):** ~50ms cold start
- **Result:** Near-instant command response

### 6x Less Memory
- **Before (Node.js):** ~300MB idle memory
- **After (Go):** ~50MB idle memory
- **Result:** Lighter footprint, better for resource-constrained environments

### Single Binary Distribution
- **Before:** `npm install` with hundreds of dependencies
- **After:** One 8MB binary, no dependencies required
- **Result:** Simpler installation, more reliable updates

---

## ✨ Complete Feature Parity

Everything you loved about OpenCode JS, now in Go:

**24 Built-in Tools:**
- File operations: read, write, edit, glob, ls
- Search: grep, codesearch, websearch, webfetch
- Advanced: apply_patch, multiedit, lsp, mcp
- AI/Agent: question, plan, skill, task
- And more...

**15+ AI Providers:**
- OpenAI (GPT-4, GPT-3.5, o1)
- Anthropic (Claude 3.5 Sonnet, Haiku, Opus)
- Google (Gemini, Vertex AI)
- Azure OpenAI, AWS Bedrock, OpenRouter
- And more...

**Multiple Interfaces:**
- CLI for terminal workflows
- TUI for interactive sessions
- Desktop app for GUI users (macOS, Windows, Linux)

---

## 📦 Getting Started

### Quick Install

#### macOS / Linux
```bash
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### Build from Source
```bash
go install github.com/gemone/libcode/cmd/libcode@latest
```

### First Run
```bash
# Set your API key
libcode auth login openai

# Start chatting
libcode chat
> Help me understand this codebase
```

---

## 🔄 Migration from OpenCode JS

Coming from OpenCode JS? Migration is seamless:

```bash
# Automatic migration
libcode migrate

# All your sessions and settings preserved
libcode session list
```

See [Migration Guide](https://github.com/gemone/libcode/blob/main/docs/migration/MIGRATION_GUIDE.md)

---

## 📊 Performance Benchmarks

| Metric | JS Version | Go Version | Improvement |
|--------|------------|------------|-------------|
| CLI Startup | 300ms | 50ms | **6x faster** |
| Memory Usage | 300MB | 50MB | **6x less** |
| Binary Size | N/A | 8MB | Self-contained |
| Session Create | 20ms | 10ms | **2x faster** |
| Concurrent Reads | 10ms | 4.8ms | **2x faster** |

---

## 🎯 Use Cases

### For Individual Developers
- Quick code understanding
- Bug investigation and debugging
- Refactoring assistance
- Documentation generation

### For Teams
- Onboarding new developers
- Code review assistance
- Consistent coding patterns
- Knowledge sharing

### For DevOps
- Infrastructure automation
- Script generation
- Configuration management
- Troubleshooting assistance

---

## 🛣️ Roadmap

### v1.0.0 (April 2026) - Current Release
- ✅ Complete JS to Go migration
- ✅ All 24 tools implemented
- ✅ 15+ AI providers
- ✅ CLI, TUI, Desktop interfaces

### v1.1.0 (Q2 2026) - Planned
- Enhanced plugin system (WASM)
- Web UI
- Collaborative features

### v1.2.0 (Q3 2026) - Planned
- Advanced code analysis
- Custom AI provider support
- Enterprise features

---

## 🙏 Acknowledgments

- **OpenCode Community** - For feedback and patience during the rewrite
- **Beta Testers** - For comprehensive testing and feedback
- **Go Team** - For an amazing language and toolchain
- **Contributors** - Everyone who helped port features from JS to Go

---

## 📚 Resources

- **GitHub:** https://github.com/gemone/libcode
- **Documentation:** https://github.com/gemone/libcode/tree/main/docs
- **Migration Guide:** https://github.com/gemone/libcode/blob/main/docs/migration/MIGRATION_GUIDE.md
- **Discord:** https://discord.gg/libcode
- **Twitter:** @libcode

---

## 🎉 Try It Today!

Download Libcode v1.0.0: https://github.com/gemone/libcode/releases/latest

**Built with Go. Powered by AI. Ready for production.**

---

*Libcode — The AI-powered development tool, rewritten for performance and simplicity.*
