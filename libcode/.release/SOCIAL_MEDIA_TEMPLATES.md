# Social Media Announcement Templates

**For:** Libcode v1.0.0 Launch
**Date:** April 1, 2026

---

## Twitter/X (Thread)

**Tweet 1:**
🚀 Announcing Libcode v1.0.0 — The AI-powered dev tool, rewritten in Go!

⚡ 6x faster startup (50ms vs 300ms)
💾 6x less memory (50MB vs 300MB)
📦 Single binary — no dependencies

Complete feature parity with OpenCode JS + massive performance gains.

Download now: https://github.com/gemone/libcode/releases

#Go #OpenSource #DevTools

**Tweet 2:**
Migrating from OpenCode JS?

🔄 Your sessions and settings migrate automatically
🛠️ All 24 tools you know and love
📖 Migration guide: https://github.com/gemone/libcode/blob/main/docs/migration/MIGRATION_GUIDE.md

The switch to Go is painless. Try it today!

#Libcode #DeveloperTools

**Tweet 3:**
What's new in Libcode v1.0.0?

✅ 24 built-in tools (read, write, edit, grep, bash, lsp, mcp...)
✅ 15+ AI providers (OpenAI, Anthropic, Google, Azure...)
✅ CLI, TUI, and Desktop interfaces
✅ 10x better performance

Built with Go. Ready for production.

#GoLang #AI #Coding

**Tweet 4:**
Benchmark results: Libcode vs OpenCode JS

Startup: 50ms vs 300ms (6x faster)
Memory: 50MB vs 300MB (6x less)
Binary: 8MB single file vs npm install

See full benchmarks: https://github.com/gemone/libcode/blob/main/docs/user/README.md#performance

---

## Reddit (r/golang, r/devtools, r/programming)

**Title:** [Release] Libcode v1.0.0 - AI-powered dev tool rewritten in Go with 6x performance improvement

**Body:**

Hi r/golang!

I'm excited to announce the v1.0.0 release of Libcode — a complete rewrite of [OpenCode](https://github.com/gemone/opencode) from JavaScript/TypeScript to Go.

**What is Libcode?**

Libcode is an AI-powered development assistant that helps you understand codebases, write better code, debug issues, and automate tasks. It provides 24 built-in tools and integrates with 15+ AI providers.

**Why Rewrite in Go?**

The rewrite was motivated by performance and distribution challenges:

| Metric | JS Version | Go Version | Improvement |
|--------|------------|------------|-------------|
| Startup Time | ~300ms | ~50ms | **6x faster** |
| Memory Usage | ~300MB | ~50MB | **6x less** |
| Distribution | npm + dependencies | Single 8MB binary | Self-contained |

**Features**

- **24 Built-in Tools:** File operations, search, bash, LSP, MCP, web search, and more
- **15+ AI Providers:** OpenAI, Anthropic, Google, Azure, AWS, OpenRouter, etc.
- **Multiple Interfaces:** CLI, TUI (Bubbletea), Desktop (Wails)
- **100% Feature Parity:** All OpenCode JS features migrated
- **Compatible:** Seamless migration from JS version

**Installation**

```bash
# macOS/Linux
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/

# Build from source
go install github.com/gemone/libcode/cmd/libcode@latest
```

**Example Usage**

```bash
# Configure
libcode auth login openai

# Chat with AI assistant
libcode chat
> Help me understand this codebase
> Read the main.go file and explain what Init does

# Quick questions
libcode ask "Search for TODO comments"
libcode ask "List all Go files in tests/"
```

**Links**

- GitHub: https://github.com/gemone/libcode
- Documentation: https://github.com/gemone/libcode/tree/main/docs
- Migration Guide: https://github.com/gemone/libcode/blob/main/docs/migration/MIGRATION_GUIDE.md

Feedback welcome! This is our 1.0 release and we're actively looking for user feedback to guide v1.1 development.

---

## Hacker News

**Title:** Show Libcode: AI-powered dev tool rewritten in Go (6x faster, 6x less memory)

**Body:**

Libcode v1.0.0 is a complete rewrite of OpenCode (an AI-powered development assistant) from Node.js to Go. The rewrite delivers:

- **6x faster startup** (50ms vs 300ms)
- **6x less memory** (50MB vs 300MB)
- **Single binary distribution** (8MB, no dependencies)
- **100% feature parity** with 24 tools and 15+ AI providers

The project uses:
- Go 1.23+ for the CLI
- Bubbletea for the TUI
- Wails for the desktop app
- Modernc.org/sqlite for embedded database
- Zap for structured logging

**Why Go?**

The Node.js version was slow to start and memory-hungry. The Go rewrite started as an experiment to see if we could dramatically improve performance while maintaining all features. The results exceeded expectations.

**Installation:**

```bash
go install github.com/gemone/libcode/cmd/libcode@latest
libcode auth login openai
libcode chat
```

GitHub: https://github.com/gemone/libcode

---

## LinkedIn

**Post:**

Excited to announce the v1.0.0 release of Libcode! 🚀

Libcode is an AI-powered development assistant that helps developers understand code, write better code, and debug faster. This is a complete rewrite from Node.js to Go, delivering:

⚡ 6x faster performance
💾 6x smaller memory footprint
📦 Single binary distribution

Key features:
- 24 built-in tools for code operations
- 15+ AI provider integrations
- CLI, terminal UI, and desktop interfaces
- Seamless migration from OpenCode JS

Great for individual developers and teams looking to accelerate their development workflow.

Download: https://github.com/gemone/libcode/releases/latest

#Go #DeveloperTools #OpenSource #AI #SoftwareDevelopment

---

## Discord/Community Announcement

**Subject:** 🎉 Libcode v1.0.0 is Live!

**Message:**

Hey everyone!

Big news today — Libcode v1.0.0 is officially released! 🎉

**What's Libcode?**
It's an AI-powered dev assistant that helps you:
- Understand codebases quickly
- Write better code with AI assistance
- Debug faster with smart search tools
- Automate repetitive tasks

**Why v1.0.0 is Special:**
We rewrote the entire thing from JavaScript to Go, and the results are insane:
- 6x faster startup
- 6x less memory
- Single binary (no npm install headaches)

**Try It Out:**
```bash
go install github.com/gemone/libcode/cmd/libcode@latest
libcode chat
```

**Migration from OpenCode JS:**
If you're using OpenCode JS, migration is seamless:
```bash
libcode migrate
```
All your sessions and settings just work!

**Feedback Welcome:**
This is 1.0.0 — we're already planning v1.1 features (WASM plugins, Web UI, etc.). Let us know what you'd like to see!

Docs: https://github.com/gemone/libcode/tree/main/docs
Issues: https://github.com/gemone/libcode/issues

Let's build something awesome together! 🚀

---

## Email Newsletter (Subscriber Announcement)

**Subject:** 🚀 Libcode v1.0.0: The Go Rewrite is Here!

**Body:**

Hi [Name],

The moment we've been working toward for months is here: **Libcode v1.0.0** is live!

If you've been following along, you know we completely rewrote Libcode from JavaScript to Go. The results are better than we ever imagined:

**Performance:**
- ⚡ Startup: 50ms (was 300ms) — **6x faster**
- 💾 Memory: 50MB (was 300MB) — **6x less**
- 📦 Size: 8MB single binary — no dependencies

**Features:**
- ✅ All 24 tools from the JS version
- ✅ 15+ AI providers (OpenAI, Anthropic, Google, etc.)
- ✅ CLI, TUI, and Desktop interfaces
- ✅ Seamless migration from OpenCode JS

**What's New in v1.0.0:**

This is our first stable Go release and includes:
- Complete feature parity with OpenCode JS
- Dramatically improved performance
- Simplified installation (no more npm!)
- Comprehensive documentation (2,100+ lines)
- Production-ready reliability

**Get Started:**

```bash
go install github.com/gemone/libcode/cmd/libcode@latest
libcode auth login openai
libcode chat
```

**What's Next:**

We're already planning v1.1.0 with:
- Enhanced WASM plugin system
- Web UI
- Collaborative features
- More AI providers

**Your Feedback:**

As an early user, your feedback shapes the future of Libcode. Reply to this email or join our Discord to share your thoughts!

Download: https://github.com/gemone/libcode/releases/latest

Happy coding!

The Libcode Team

---
Unsubscribe: [Link]
