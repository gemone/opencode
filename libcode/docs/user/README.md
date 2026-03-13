# Libcode User Documentation

## Installation

### Prerequisites
- Go 1.23+ for building from source
- For desktop app: macOS 12+, Windows 10+, or Linux (Ubuntu 20.04+)

### Install from Binary

#### macOS (arm64)
```bash
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-darwin-amd64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### Linux
```bash
curl -L https://github.com/gemone/libcode/releases/latest/download/libcode-linux-amd64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### Windows
```powershell
# Download from https://github.com/gemone/libcode/releases/latest
# Extract and add to PATH
```

### Build from Source

```bash
git clone https://github.com/gemone/libcode.git
cd libcode
go build -o bin/libcode ./cmd/libcode
sudo mv bin/libcode /usr/local/bin/
```

### Desktop Application

#### macOS
```bash
cd libcode/desktop
wails build
# Open libcode-desktop.app
```

#### Linux
```bash
cd libcode/desktop
wails build
# Install and run from applications menu
```

#### Windows
```bash
cd libcode\desktop
wails build
# Run libcode-desktop.exe
```

## Quick Start

### Initial Setup

1. **Configure AI Provider**
```bash
libcode config edit
```

Add your API key (choose one):
```yaml
providers:
  openai:
    api_key: "sk-..."  # Add your OpenAI API key
    enabled: true
    model: "gpt-4"

  anthropic:
    api_key: "sk-ant-..."  # Add your Anthropic API key
    enabled: true
    model: "claude-3-5-sonnet-20241022"
```

2. **Start a Chat Session**
```bash
libcode chat
```

3. **Ask a Question**
```
> Help me understand this codebase
> What files handle authentication?
> How do I add a new tool?
```

### First Session Example

```bash
$ libcode chat
> Analyze the main.go file
[Libcode reads main.go and provides analysis]

> Refactor it to use error wrapping
[Libcode suggests improvements with code examples]

> Apply those changes
[Libcode edits the file]

> Show me the diff
[Libcode displays the changes]
```

## Configuration

### Config File Location

Config is loaded from (in order of priority):
1. `./libcode.yaml` (current directory)
2. `~/.config/libcode/config.yaml`
3. `~/.libcode/config.yaml`

### Config Options

```yaml
# AI Provider Settings
default_model: "openai/gpt-4"  # Default model to use
model_variant: ""               # Model variant (if applicable)

# Session Settings
session_dir: "~/.libcode/sessions"  # Where to store sessions
max_history: 100                   # Maximum messages per session

# Logging
log_level: "INFO"              # DEBUG, INFO, WARN, ERROR
log_file: ""                   # Empty = stdout only

# LSP Integration
lsp:
  enabled: true
  server_timeout: 30           # LSP server timeout in seconds
  languages:
    - typescript
    - go
    - python
    - rust

# MCP Integration
mcp:
  enabled: true
  server_timeout: 30           # MCP server timeout in seconds
  servers:
    - name: "filesystem"
      command: "npx"
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed"]

# Features
features:
  tui: true                    # Enable terminal UI
  desktop_mode: false          # Enable desktop mode
  stream_response: true        # Stream AI responses

# Providers Configuration
providers:
  openai:
    api_key: "sk-..."
    enabled: true
    model: "gpt-4"
    base_url: ""               # Optional: custom base URL

  anthropic:
    api_key: "sk-ant-..."
    enabled: true
    model: "claude-3-5-sonnet-20241022"
    base_url: ""

  google:
    api_key: "..."
    enabled: false
    model: "gemini-2.0-flash-exp"
```

### Environment Variables

You can also configure via environment variables:

```bash
export LIBCODE_OPENAI_API_KEY="sk-..."
export LIBCODE_ANTHROPIC_API_KEY="sk-ant-..."
export LIBCODE_DEFAULT_MODEL="openai/gpt-4"
export LIBCODE_LOG_LEVEL="DEBUG"
export LIBCODE_SESSION_DIR="~/.libcode/sessions"
```

## Command Reference

### Main Commands

#### `libcode chat`
Start an interactive chat session.

```bash
libcode chat                    # Start new session
libcode chat --session <id>     # Resume existing session
libcode chat --model <model>    # Use specific model
```

#### `libcode ask`
Ask a single question without starting a session.

```bash
libcode ask "What does this function do?"
libcode ask "Explain the auth flow" --model claude-3-5-sonnet-20241022
```

#### `libcode config`
Manage configuration.

```bash
libcode config edit             # Open config in editor
libcode config validate         # Validate config file
libcode config show             # Display current config
libcode config set <key> <val>  # Set a config value
```

#### `libcode session`
Manage sessions.

```bash
libcode session list            # List all sessions
libcode session show <id>       # Show session details
libcode session delete <id>     # Delete a session
libcode session export <id>     # Export session to JSON
```

#### `libcode mcp`
Manage MCP servers.

```bash
libcode mcp list                # List configured servers
libcode mcp add <name>          # Add a new server
libcode mcp remove <name>       # Remove a server
libcode mcp test <name>         # Test server connection
```

#### `libcode lsp`
Manage LSP servers.

```bash
libcode lsp list                # List available LSP servers
libcode lsp start <lang>        # Start LSP for language
libcode lsp stop <lang>         # Stop LSP
libcode lsp status              # Show LSP status
```

#### `libcode skill`
Manage skills.

```bash
libcode skill list              # List available skills
libcode skill show <name>       # Show skill details
libcode skill install <url>     # Install skill from URL
```

### Advanced Commands

#### `libcode update`
Update to the latest version.

```bash
libcode update                  # Check and install updates
libcode update --check-only     # Only check for updates
```

#### `libcode doctor`
Run diagnostics.

```bash
libcode doctor                  # Run all diagnostic checks
libcode doctor --config         # Check configuration only
libcode doctor --providers      # Check AI providers only
```

#### `libcode migrate`
Migrate from JavaScript version.

```bash
libcode migrate                 # Migrate sessions and config
libcode migrate --dry-run       # Preview migration without changes
```

## Features

### Chat Modes

#### Interactive Chat
```bash
libcode chat
```
Features:
- Streaming responses
- Tool execution (read files, run commands, search code)
- Session history
- Multi-file context

#### Quick Ask
```bash
libcode ask "your question"
```
Best for:
- Single questions
- Scripts and automation
- CI/CD integration

#### Desktop App
```bash
libcode-desktop
```
Features:
- Native UI
- Keyboard shortcuts
- Session sidebar
- Settings panel

### Tools

Libcode includes 24 built-in tools:

**File Operations**
- `read` - Read file contents
- `write` - Write/create files
- `edit` - Edit files with string replacement
- `glob` - Find files by pattern
- `ls` - List directory contents

**Search**
- `grep` - Search file contents
- `codesearch` - Advanced code search
- `websearch` - Search the web

**Execution**
- `bash` - Execute shell commands
- `task` - Spawn subagent tasks

**Advanced**
- `apply_patch` - Apply unified diff patches
- `multiedit` - Batch file editing
- `lsp` - Language server operations
- `skill` - Load specialized skills
- `question` - Ask user questions
- `plan` - Planning mode management

### Skills

Skills provide specialized instructions for specific tasks:

**Available Skills**
- `tdd` - Test-driven development workflow
- `security-review` - Security code review
- `code-review` - Comprehensive code review
- `refactor` - Code refactoring patterns
- `debug` - Debugging assistance
- `architecture` - Software architecture review

**Using Skills**

In chat:
```
> Use the tdd skill to help me add tests for auth.go
```

Or via CLI:
```bash
libcode skill show tdd
```

## Troubleshooting

### Common Issues

#### "No API key configured"
```bash
# Add your API key to config
libcode config set providers.openai.api_key "sk-..."

# Or set environment variable
export LIBCODE_OPENAI_API_KEY="sk-..."
```

#### "LSP server failed to start"
```bash
# Check if LSP is installed
libcode lsp status

# Install missing LSP
libcode lsp install typescript
```

#### "Session not found"
```bash
# List available sessions
libcode session list

# Use correct session ID
libcode chat --session <correct-id>
```

#### "Permission denied"
```bash
# Libcode may need permission to access files
# Grant permission when prompted, or:
libcode config set permissions.allow_external true
```

### Debug Mode

Enable debug logging:

```bash
libcode config set log_level DEBUG
libcode chat  # Run with debug output
```

### Get Help

```bash
libcode --help           # General help
libcode <command> --help # Command-specific help
libcode doctor           # Run diagnostics
```

## FAQ

**Q: How do I switch between AI providers?**
```bash
libcode config edit
# Change default_model to your preferred provider
# Example: default_model: "anthropic/claude-3-5-sonnet-20241022"
```

**Q: Can I use libcode offline?**
Partially. Libcode requires internet for AI providers, but file operations, LSP, and local tools work offline.

**Q: How much memory does libcode use?**
Idle: ~50MB RAM
Active session: ~100-200MB RAM depending on context size

**Q: Is my code sent to AI providers?**
Only when you explicitly use AI features. File operations, LSP, and local tools don't send code externally.

**Q: Can I use libcode in CI/CD?**
Yes! Use `libcode ask` for one-off questions or automate with scripts.

**Q: How do I migrate from the JavaScript version?**
```bash
libcode migrate --dry-run  # Preview migration
libcode migrate             # Migrate sessions and config
```

See [Migration Guide](../migration/MIGRATION_GUIDE.md) for details.

## Next Steps

- Read [Developer Documentation](../developer/README.md) to contribute
- Check [Migration Guide](../migration/MIGRATION_GUIDE.md) if coming from JS version
- See [Troubleshooting](../troubleshooting/README.md) for more help
