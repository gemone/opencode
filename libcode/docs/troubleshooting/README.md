# Libcode Troubleshooting Guide

## Installation Issues

### "command not found: libcode"

**Problem:** libcode is not in your PATH.

**Solution:**
```bash
# Check if binary exists
which libcode

# If not found, install:
# macOS/Linux
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/

# Verify installation
libcode --version
```

### "Permission denied when installing"

**Problem:** Insufficient permissions to install to /usr/local/bin.

**Solution:**
```bash
# Option 1: Use sudo
sudo mv libcode /usr/local/bin/

# Option 2: Install to user directory
mkdir -p ~/.local/bin
mv libcode ~/.local/bin/
# Add to PATH in ~/.zshrc or ~/.bashrc:
# export PATH="$HOME/.local/bin:$PATH"
```

### Build fails with "go: module not found"

**Problem:** Go dependencies not downloaded.

**Solution:**
```bash
cd libcode
go mod download
go mod tidy
go build -o bin/libcode ./cmd/libcode
```

## Configuration Issues

### "No API key configured"

**Problem:** libcode cannot find your API keys.

**Solution:**
```bash
# Option 1: Edit config
libcode config edit
# Add your API key:
# providers:
#   openai:
#     api_key: "sk-..."

# Option 2: Set environment variable
export LIBCODE_OPENAI_API_KEY="sk-..."
export LIBCODE_ANTHROPIC_API_KEY="sk-ant-..."

# Option 3: Pass via flag
libcode chat --openai-api-key "sk-..."
```

### "Config file not found"

**Problem:** libcode cannot find configuration file.

**Solution:**
```bash
# Create default config
mkdir -p ~/.libcode
libcode config edit

# Or specify custom location
libcode --config /path/to/config.yaml chat
```

### "Invalid config format"

**Problem:** YAML syntax error in config file.

**Solution:**
```bash
# Validate config
libcode config validate

# Common YAML errors:
# - Use spaces, not tabs for indentation
# - Ensure proper nesting
# - Quote strings with special characters

# Example of correct config:
providers:
  openai:
    api_key: "sk-..."  # Use quotes
    enabled: true
    model: "gpt-4"     # Use spaces for indentation
```

## Session Issues

### "Session not found"

**Problem:** Session ID doesn't exist or database is corrupted.

**Solution:**
```bash
# List available sessions
libcode session list

# Check session database
libcode doctor --sessions

# Repair database
libcode migrate --repair
```

### "Cannot save session"

**Problem:** Permissions issue or disk full.

**Solution:**
```bash
# Check session directory permissions
ls -la ~/.libcode/sessions/

# Fix permissions
chmod 755 ~/.libcode/sessions/

# Check disk space
df -h
```

### "Session too large"

**Problem:** Session exceeds memory limits (default: 100 messages).

**Solution:**
```bash
# Option 1: Increase limit
libcode config set max_history 200

# Option 2: Start new session
libcode chat

# Option 3: Export and split session
libcode session export <id> > session.json
```

## AI Provider Issues

### "OpenAI API error"

**Problem:** API key invalid or quota exceeded.

**Solution:**
```bash
# Verify API key
libcode config show | grep api_key

# Test API key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Check quota and billing
# Visit: https://platform.openai.com/usage
```

### "Anthropic API error"

**Problem:** Anthropic API authentication failure.

**Solution:**
```bash
# Verify API key format
# Anthropic keys start with: sk-ant-

libcode config set providers.anthropic.api_key "sk-ant-..."

# Test connection
libcode ask "test" --model anthropic/claude-3-5-sonnet-20241022
```

### "Rate limit exceeded"

**Problem:** Too many requests to AI provider.

**Solution:**
```bash
# Option 1: Wait and retry
# Rate limits typically reset after 1 minute

# Option 2: Switch providers
libcode chat --model anthropic/claude-3-5-sonnet-20241022

# Option 3: Reduce concurrent requests
libcode config set providers.openai.max_concurrent 1
```

### "Streaming not working"

**Problem:** Streaming responses disabled or not supported.

**Solution:**
```bash
# Enable streaming
libcode config set features.stream_response true

# Check provider supports streaming
libcode doctor --providers
```

## Tool Issues

### "Tool not found"

**Problem:** Tool name misspelled or not available.

**Solution:**
```bash
# List available tools
libcode tool list

# Tool names are snake_case:
# Correct: web_search, apply_patch, multi_edit
# Incorrect: webSearch, ApplyPatch, multi-edit
```

### "Permission denied for tool"

**Problem:** Tool requires permission but was denied.

**Solution:**
```bash
# Grant permission when prompted
# Or pre-approve in config:

libcode config edit
# Add:
# permissions:
#   allow_bash: true
#   allow_external_directory: true
```

### "Tool execution timeout"

**Problem:** Tool took too long to execute.

**Solution:**
```bash
# Increase timeout
libcode config set tool_timeout 120  # seconds

# Or run in background for long operations
libcode ask "Run this in background: long-command"
```

## LSP Issues

### "LSP server failed to start"

**Problem:** Language server not installed or wrong path.

**Solution:**
```bash
# Check LSP status
libcode lsp status

# Install missing LSP
# TypeScript: npm install -g typescript-language-server
# Go: go install golang.org/x/tools/gopls@latest
# Python: pip install python-lsp-server
# Rust: rustup component add rust-analyzer
```

### "No diagnostics shown"

**Problem:** LSP not reporting errors.

**Solution:**
```bash
# Restart LSP
libcode lsp restart typescript

# Check LSP is enabled
libcode config show | grep lsp

# Force diagnostics refresh
libcode lsp diagnostics
```

### "LSP slows down editor"

**Problem:** LSP using too many resources.

**Solution:**
```bash
# Reduce LSP timeout
libcode config set lsp.server_timeout 15

# Disable unused languages
libcode config edit
# Keep only languages you use:
# lsp:
#   languages:
#     - typescript
#     - go
```

## MCP Issues

### "MCP server not responding"

**Problem:** MCP server failed to start or crashed.

**Solution:**
```bash
# Test MCP server
libcode mcp test <server-name>

# Check server logs
libcode mcp logs <server-name>

# Restart server
libcode mcp restart <server-name>
```

### "MCP tools not available"

**Problem:** MCP server not exposing tools.

**Solution:**
```bash
# List MCP tools
libcode mcp list-tools <server-name>

# Check server initialization
libcode mcp initialize <server-name>

# Verify server config
libcode config show | grep -A 10 mcp
```

## Performance Issues

### "Slow startup"

**Problem:** libcode takes >1 second to start.

**Solution:**
```bash
# Check what's slowing startup
libcode --debug version

# Common causes:
# 1. Slow DNS resolution - fix /etc/hosts
# 2. Large session history - archive old sessions
# 3. Many LSP servers - disable unused ones

# Archive old sessions
libcode session archive --before 2024-01-01
```

### "High memory usage"

**Problem:** libcode using >500MB RAM.

**Solution:**
```bash
# Check session sizes
libcode session list --sizes

# Reduce history limit
libcode config set max_history 50

# Clear cache
libcode cache clear
```

### "Slow streaming response"

**Problem:** Streaming chunks arriving slowly.

**Solution:**
```bash
# Measure TTFB (time to first byte)
libcode ask "test" --debug

# Common causes:
# 1. Network latency - switch to closer region
# 2. Provider throttling - check quota
# 3. Large context - reduce session history

# Reduce context window
libcode chat --context-window 4000
```

## Desktop App Issues

### "Desktop app won't open"

**Problem:** Application fails to start.

**Solution:**
```bash
# macOS: Check app is not quarantined
xattr -d /Applications/Libcode.app

# Linux: Check dependencies
ldd bin/libcode-desktop

# Windows: Run as administrator
# Right-click → Run as administrator
```

### "Desktop app not responding"

**Problem:** UI freezes or crashes.

**Solution:**
```bash
# Check logs
# macOS: ~/Library/Logs/Libcode/
# Linux: ~/.config/Libcode/logs/
# Windows: %APPDATA%/Libcode/logs/

# Restart with clean state
libcode-desktop --reset-state

# Reinstall desktop app
cd desktop && wails build
```

### "Cannot connect to backend"

**Problem:** Frontend cannot reach Go backend.

**Solution:**
```bash
# Check if backend is running
ps aux | grep libcode

# Restart desktop app
# macOS: Force Quit and reopen
# Linux: killall libcode-desktop && libcode-desktop &
# Windows: Task Manager → End Task → Restart
```

## Debugging

### Enable Debug Logging

```bash
# Set log level to DEBUG
libcode config set log_level DEBUG

# Run with debug output
libcode --debug chat

# Check log file
tail -f ~/.libcode/libcode.log
```

### Run Diagnostics

```bash
# Full diagnostic check
libcode doctor

# Specific checks
libcode doctor --config      # Check configuration
libcode doctor --providers   # Check AI providers
libcode doctor --sessions    # Check sessions
libcode doctor --lsp         # Check LSP servers
libcode doctor --mcp         # Check MCP servers
```

### Export Diagnostic Info

```bash
# Export full diagnostic report
libcode doctor --export diagnostic-report.json

# Share with support team
# (Sanitize sensitive data first!)
```

## Common Error Messages

### "context deadline exceeded"

**Meaning:** Operation took too long and timed out.

**Solutions:**
- Increase timeout in config
- Check network connectivity
- Reduce context window
- Try non-streaming mode

### "unexpected EOF"

**Meaning:** Connection was closed unexpectedly.

**Solutions:**
- Check network stability
- Verify API endpoint is correct
- Try different provider
- Check proxy settings

### "constraint failed"

**Meaning:** Database constraint violation.

**Solutions:**
- Session ID already exists
- Database corruption
- Run: `libcode migrate --repair`

### "unauthorized"

**Meaning:** Invalid API key or credentials.

**Solutions:**
- Verify API key in config
- Check key hasn't been revoked
- Ensure key has required permissions

## Getting Additional Help

### Check Documentation

- [User Guide](../user/README.md)
- [Developer Guide](../developer/README.md)
- [Migration Guide](../migration/MIGRATION_GUIDE.md)

### Search Issues

- [GitHub Issues](https://github.com/gemone/libcode/issues)
- Check existing issues before creating new ones

### Join Community

- [Discussions](https://github.com/gemone/libcode/discussions)
- [Discord](https://discord.gg/libcode)

### Report Bugs

When reporting bugs, include:

1. **libcode version**: `libcode --version`
2. **Go version**: `go version`
3. **OS**: `uname -a` (macOS/Linux) or Windows version
4. **Error message**: Full error output
5. **Steps to reproduce**: Minimal reproduction
6. **Diagnostic info**: `libcode doctor --export diagnostic.json`

### Support Checklist

Before asking for help:

- [ ] Read this troubleshooting guide
- [ ] Checked documentation
- [ ] Searched existing issues
- [ ] Run `libcode doctor`
- [ ] Enabled debug logging
- [ ] Tried latest version

## Emergency Recovery

### Factory Reset

If all else fails, reset to default state:

```bash
# Backup current state
cp -r ~/.libcode ~/.libcode.backup

# Factory reset
rm -rf ~/.libcode
libcode config edit  # Creates fresh config

# Restore sessions from backup
cp -r ~/.libcode.backup/sessions ~/.libcode/
```

### Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/libcode

# Remove data (optional)
rm -rf ~/.libcode
rm -rf ~/.config/libcode

# Note: This removes all sessions and configuration
```
