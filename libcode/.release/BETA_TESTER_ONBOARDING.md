# Beta Tester Onboarding Guide

**Welcome to Libcode Beta Testing!**

Thank you for participating in the beta testing of Libcode v1.0.0.

## Installation

### Quick Install

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/gemone/libcode/releases/download/v1.0.0-beta.1/libcode-darwin-arm64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

#### Linux
```bash
curl -L https://github.com/gemone/libcode/releases/download/v1.0.0-beta.1/libcode-linux-amd64.tar.gz | tar xz
sudo mv libcode /usr/local/bin/
```

## First Time Setup

### 1. Configure API Key

```bash
libcode auth login openai
# Enter your API key when prompted
```

### 2. Verify Installation

```bash
libcode version
libcode auth status
```

### 3. Test Basic Commands

```bash
libcode chat
> Hello, can you help me?
```

## Testing Checklist

- [ ] Installation successful
- [ ] API key configured
- [ ] Interactive chat works
- [ ] Tool execution works
- [ ] Session management works
- [ ] Migration from JS works (if applicable)

## Providing Feedback

### Feedback Channels

- **GitHub Issues:** https://github.com/gemone/libcode/issues
- **Discord:** https://discord.gg/libcode-beta
- **Feedback Form:** https://forms.gle/libcode-beta-feedback

### What to Report

- Bugs and crashes
- Performance issues
- Feature requests
- Documentation improvements

## Getting Help

- [User Guide](https://github.com/gemone/libcode/blob/main/docs/user/README.md)
- [Troubleshooting](https://github.com/gemone/libcode/blob/main/docs/troubleshooting/README.md)
- Discord community for real-time help

## Beta Timeline

- **Week 1:** Beta release and initial feedback
- **Week 2:** Bug fixes and refinements  
- **Week 3:** Release candidates
- **Week 4:** Official launch

Thank you for testing Libcode!
