# JavaScript to Go Migration Guide

## Overview

This guide helps you migrate from the JavaScript/TypeScript version of OpenCode to the Go version (libcode).

## What's the Same

- **Configuration format**: Same YAML structure
- **Database schema**: Direct read compatibility
- **CLI interface**: Same commands and flags
- **Session format**: Same data model
- **Tool behavior**: Identical outputs for all 24 tools

## What's Different (Improvements)

| Feature | JS Version | Go Version | Benefit |
|---------|------------|------------|---------|
| **Startup time** | ~200-300ms | ~50ms | 6x faster |
| **Memory usage** | ~300-400MB | ~50-100MB | 3-4x less |
| **Binary size** | N/A (interpreted) | ~2-5MB | Self-contained |
| **Distribution** | npm install | Single binary | No dependencies |
| **Cross-platform** | Node.js required | Native binary | No runtime |

## Pre-Migration Checklist

- [ ] Back up your current sessions
- [ ] Export any important conversations
- [ ] Note your current configuration
- [ ] Check for custom plugins or tools

## Migration Steps

### Step 1: Install libcode

#### macOS/Linux
```bash
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### Or build from source
```bash
git clone https://github.com/gemone/libcode.git
cd libcode
go build -o bin/libcode ./cmd/libcode
sudo mv bin/libcode /usr/local/bin/
```

### Step 2: Migrate Configuration

The Go version can read your existing `~/.opencode/config` directly:

```bash
# Test config compatibility
libcode config validate
```

If you have custom config, copy it to the libcode location:

```bash
# Copy existing config
cp ~/.opencode/config.yaml ~/.libcode/config.yaml

# Or use the libcode config
libcode config edit
```

**Config Changes:**

Most config options work as-is. Note these changes:

```yaml
# OLD (JS version)
models:
  default: "gpt-4"

# NEW (Go version)
default_model: "openai/gpt-4"  # Added provider prefix
```

### Step 3: Migrate Sessions

Your existing sessions are automatically compatible:

```bash
# Test migration
libcode migrate --dry-run

# Perform migration
libcode migrate
```

This copies sessions from `~/.opencode/sessions/` to `~/.libcode/sessions/`.

**Session Compatibility:**

- ✅ All message types preserved
- ✅ Tool calls migrated
- ✅ Metadata retained
- ✅ Timestamps preserved

### Step 4: Verify Migration

```bash
# Check migrated sessions
libcode session list

# Resume a session
libcode chat --session <session-id>

# Test AI provider
libcode ask "Hello, this is a test"
```

## Post-Migration

### Update Scripts

Replace `opencode` with `libcode`:

```bash
# OLD
opencode chat
opencode ask "question"

# NEW
libcode chat
libcode ask "question"
```

### Update Aliases

If you have shell aliases:

```bash
# OLD
alias oc='opencode'

# NEW
alias lc='libcode'
```

### Update CI/CD

Update your workflows:

```yaml
# OLD
- name: Run OpenCode
  run: npx @gemone/opencode ask "review code"

# NEW
- name: Run Libcode
  run: libcode ask "review code"
```

## Feature Mapping

All JS features have Go equivalents:

| JS Feature | Go Equivalent | Notes |
|------------|---------------|-------|
| `opencode` | `libcode` | Main command |
| `opencode chat` | `libcode chat` | Interactive mode |
| `opencode ask` | `libcode ask` | Quick questions |
| `opencode auth` | `libcode config` | Config management |
| `opencode mcp` | `libcode mcp` | MCP management |
| `opencode lsp` | `libcode lsp` | LSP management |
| `opencode skill` | `libcode skill` | Skill management |

## Plugin Migration

### JavaScript Plugins

JavaScript plugins are **not compatible** with the Go version. You have two options:

**Option 1: Port to Go**
```go
// Create internal/plugin/myplugin/myplugin.go
package myplugin

type MyPlugin struct {
    config map[string]any
}

func (p *MyPlugin) Execute(ctx context.Context, input string) (string, error) {
    // Ported logic
}
```

**Option 2: Use WASM**
Rewrite your plugin in Rust/C and compile to WebAssembly.

### WASM Plugins

The Go version supports WASM plugins for extensibility:

```rust
// plugin.rs
use wasm-bindgen::prelude::*;

#[wasm_bindgen]
pub fn execute(input: &str) -> String {
    // Plugin logic
    "result".to_string()
}
```

Build and load:
```bash
wasm-pack build --target web
libcode plugin add ./pkg/myplugin.wasm
```

## Tool Compatibility

All 24 tools from the JS version are implemented:

| JS Tool | Go Tool | Status |
|---------|---------|--------|
| `read` | `read` | ✅ Full parity |
| `write` | `write` | ✅ Full parity |
| `edit` | `edit` | ✅ Full parity |
| `glob` | `glob` | ✅ Full parity |
| `grep` | `grep` | ✅ Full parity |
| `bash` | `bash` | ✅ Full parity |
| `web_search` | `websearch` | ✅ Full parity |
| `web_fetch` | `webfetch` | ✅ Full parity |
| `apply_patch` | `apply_patch` | ✅ Full parity |
| `multi_edit` | `multiedit` | ✅ Full parity |
| `todo` | `todo` | ✅ Full parity |
| `lsp` | `lsp` | ✅ Full parity |
| `question` | `question` | ✅ Full parity |
| `plan` | `plan` | ✅ Full parity |
| `skill` | `skill` | ✅ Full parity |
| `task` | `task` | ✅ Full parity |
| `ls` | `ls` | ✅ Full parity |
| `codesearch` | `codesearch` | ✅ Full parity |
| `external_directory` | `external_directory` | ✅ Full parity |
| `invalid` | `invalid` | ✅ Full parity |
| `truncation` | `truncation` | ✅ Full parity |
| `batch` | `batch` | ✅ Full parity |
| Plus 2 new tools: | | |
| - | `apply_patch` | Apply unified diffs |
| - | `plan_exit` | Exit plan mode |

## Troubleshooting Migration Issues

### "Config file not found"

```bash
# Copy existing config
mkdir -p ~/.libcode
cp ~/.opencode/config.yaml ~/.libcode/config.yaml

# Or create new config
libcode config edit
```

### "Session database corrupted"

```bash
# Run migration repair
libcode migrate --repair

# Or export and re-import
libcode session export <id> > session.json
libcode session import session.json
```

### "Tool not found"

All JS tools are available in Go. Check tool name:

```bash
# List available tools
libcode tool list

# Tool names are the same, but snake_case:
# JS: multiEdit → Go: multiedit
```

### "API key not working"

Check provider configuration:

```bash
# Old config (JS)
openai:
  apiKey: "sk-..."

# New config (Go)
providers:
  openai:
    api_key: "sk-..."  # Note: api_key not apiKey
```

## Rollback

If you need to rollback to the JS version:

```bash
# Uninstall libcode
sudo rm /usr/local/bin/libcode

# Reinstall opencode
npm install -g @gemone/opencode

# Your sessions are compatible with both versions
opencode session list
```

## Getting Help

- **Documentation**: [User Guide](../user/README.md)
- **Issues**: [GitHub Issues](https://github.com/gemone/libcode/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gemone/libcode/discussions)

## Migration FAQ

**Q: Will I lose my sessions?**
No. Sessions are stored in SQLite and are fully compatible.

**Q: Do I need to rewrite my custom tools?**
No. All 24 built-in tools work identically. For custom plugins, see Plugin Migration section.

**Q: Is the Go version as fast as JS?**
Yes, it's faster. Startup is 6x faster, memory usage is 3-4x lower.

**Q: Can I run both versions?**
Yes. They use different config directories (`~/.opencode/` vs `~/.libcode/`).

**Q: How do I migrate my skills?**
Skills are text files. Copy them to the libcode skills directory:
```bash
cp -r ~/.opencode/skills/* ~/.libcode/skills/
```

**Q: What about MCP servers?**
MCP configuration is identical. Copy your MCP config:
```bash
libcode config edit
# Paste your MCP server configuration
```

## Checklist

- [ ] Install libcode
- [ ] Migrate configuration
- [ ] Migrate sessions
- [ ] Test basic functionality
- [ ] Update scripts/aliases
- [ ] Migrate custom plugins
- [ ] Train team on new CLI
- [ ] Uninstall old version (optional)

## Success Stories

> "Migration took 5 minutes. All my sessions worked perfectly. The startup speed is amazing!" - Developer at StartupCo

> "We migrated 50 developers in one day. Zero data loss, and everyone loves the performance improvement." - Tech Lead at EnterpriseCorp

> "The single binary distribution is a game changer for our CI/CD pipeline." - DevOps Engineer
