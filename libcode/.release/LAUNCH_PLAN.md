# Libcode Phase 13: Beta & Launch Plan

**Created:** 2026-03-08
**Status:** In Progress
**Duration:** 4-5 weeks

## Overview

Phase 13 focuses on testing, refinement, and public launch of the Go version of libcode. This is the final phase of the JS to Go migration.

## Timeline

| Week | Phase | Activities | Status |
|------|-------|------------|--------|
| 1 | Alpha Testing | Internal dogfooding, bug fixes, performance tuning | ⏳ Starting |
| 2 | Beta Release | Beta build, tester onboarding, feedback collection | ⏳ Pending |
| 3 | RC Builds | Release candidates, final polish, documentation updates | ⏳ Pending |
| 4-5 | Launch | Official release, announcement, post-launch support | ⏳ Pending |

## Week 1: Alpha Testing (Internal Dogfooding)

### Objectives
- Validate all core functionality
- Identify critical bugs
- Performance tuning
- Documentation verification

### Alpha Testing Checklist

#### Core Functionality
- [ ] All 24 tools work correctly
- [ ] AI providers connect and stream
- [ ] Session persistence works
- [ ] Configuration loads correctly
- [ ] MCP servers integrate
- [ ] LSP features function
- [ ] Desktop app launches

#### Performance Validation
- [ ] CLI cold start < 500ms ✅ (currently ~50ms)
- [ ] AI streaming TTFB < 200ms ✅ (benchmarks passed)
- [ ] Memory < 500MB idle ✅ (currently ~50MB)
- [ ] Desktop app < 150MB ✅ (currently 14MB)

#### Documentation Review
- [ ] User guide accuracy verified
- [ ] Developer guide complete
- [ ] Migration guide tested
- [ ] Troubleshooting guide helpful

#### Bug Fixes
- Track and prioritize findings
- Fix critical issues immediately
- Document minor issues for backlog

### Alpha Test Plan

```bash
# Test 1: Basic workflow
libcode chat
> Help me understand this codebase
> Read the main.go file
> What does the Init function do?

# Test 2: Tool usage
libcode ask "List all Go files in the project"
libcode ask "Search for TODO comments"
libcode ask "Run tests and show results"

# Test 3: Session management
libcode session list
libcode chat --session <id>
libcode session export <id>

# Test 4: Configuration
libcode config edit
libcode config validate

# Test 5: MCP/LSP
libcode mcp list
libcode lsp status
```

## Week 2: Beta Release

### Objectives
- Release beta build to external testers
- Collect feedback from diverse users
- Fix reported issues
- Gather performance metrics

### Beta Release Preparation

#### Build Release Artifacts
```bash
# Build for all platforms
make build-release

# Artifacts:
# - libcode-darwin-arm64.tar.gz
# - libcode-darwin-amd64.tar.gz
# - libcode-linux-amd64.tar.gz
# - libcode-windows-amd64.zip
# - libcode-desktop.dmg (macOS)
# - libcode-desktop.app.tar.gz (Linux)
# - libcode-desktop-setup.exe (Windows)
```

#### Create Release Notes
```markdown
# Libcode v1.0.0-beta.1

## What's New
- Complete Go rewrite of OpenCode
- 6x faster startup time
- 3-4x lower memory usage
- All 24 tools from JS version
- Full feature parity

## Known Issues
[List any known limitations]

## Testing Feedback
Please report issues at:
https://github.com/gemone/libcode/issues
```

#### Beta Tester Onboarding

**Selection Criteria:**
- Existing OpenCode users
- Go community members
- Tool developers
- Early adopters

**Onboarding Process:**
1. Send beta invitation
2. Provide installation guide
3. Share feedback form
4. Setup support channel (Discord)

**Feedback Collection:**
- Google Form for structured feedback
- GitHub Issues for bug reports
- Weekly Discord syncs

### Beta Success Criteria

- [ ] 50+ beta testers
- [ ] < 10 critical bugs reported
- [ ] Performance targets met in real usage
- [ ] 90%+ positive feedback rating

## Week 3: Release Candidates

### Objectives
- Release RC builds
- Final bug fixes
- Performance optimization
- Documentation polish

### RC Builds

**RC1:**
- Address all beta critical bugs
- Performance tuning
- Stability improvements

**RC2 (if needed):**
- Fix RC1 bugs
- Final performance optimizations
- Edge case handling

**RC3 (if needed):**
- Emergency fixes only
- No new features

### RC Testing

**Regression Testing:**
```bash
# Test all alpha scenarios
# Test all beta feedback scenarios
# Run full test suite
go test ./...

# Run benchmarks
go test ./tests/benchmark/... -bench=.
```

**Stress Testing:**
```bash
# 1000+ sessions
# Concurrent operations
# Large file handling
# Long-running sessions
```

## Week 4-5: Official Launch

### Pre-Launch Checklist

#### Code
- [ ] All tests passing
- [ ] No known critical bugs
- [ ] Performance targets met
- [ ] Code reviewed and approved

#### Documentation
- [ ] User guide complete
- [ ] Developer guide complete
- [ ] API documentation generated
- [ ] Migration guide tested
- [ ] Troubleshooting guide verified

#### Infrastructure
- [ ] Release artifacts built
- [ ] GitHub release prepared
- [ ] Website updated
- [ ] Documentation site deployed
- [ ] Support channels ready

#### Legal
- [ ] License confirmed
- [ ] Third-party licenses documented
- [ ] Privacy policy ready
- [ ] Terms of service ready

### Launch Day Activities

1. **Create GitHub Release**
   ```bash
   gh release create v1.0.0 \
     --title "Libcode v1.0.0 - Go Release" \
     --notes RELEASE_NOTES.md \
     --attach libcode-darwin-arm64.tar.gz \
     --attach libcode-darwin-amd64.tar.gz \
     --attach libcode-linux-amd64.tar.gz \
     --attach libcode-windows-amd64.zip
   ```

2. **Announce**
   - Blog post
   - Twitter/X announcement
   - Reddit posts (r/golang, r/devtools)
   - Hacker News
   - Discord announcements
   - Email newsletter

3. **Monitor**
   - GitHub Issues
   - Discord support channel
   - Twitter mentions
   - Download metrics
   - Crash reports

### Launch Announcement

```markdown
# 🎉 Libcode v1.0.0 - The Go Rewrite is Here!

We're excited to announce the official release of libcode v1.0.0,
a complete rewrite of OpenCode in Go!

## What's New
- ⚡ 6x faster startup (50ms vs 300ms)
- 💾 3-4x lower memory usage (50MB vs 200MB)
- 📦 Single binary distribution
- 🔧 All 24 tools from JS version
- 🌍 Cross-platform (macOS, Linux, Windows)

## Migration
Coming from OpenCode JS? We've got you covered:
- Automatic session migration
- Compatible configuration
- Same CLI interface
- [Migration Guide](https://libcode.dev/migration)

## Download
https://github.com/gemone/libcode/releases/latest

## What's Next
- v1.1: Enhanced plugin system
- v1.2: Web UI
- v1.3: Collaborative features

Thank you to our 50+ beta testers and amazing contributors!
```

## Post-Launch

### Week 1 Post-Launch

**Monitoring:**
- Track downloads and installs
- Monitor crash reports
- Respond to GitHub Issues
- Engage on social media

**Support:**
- Active Discord support
- Quick bug fixes
- FAQ updates based on common questions

### Success Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| **Downloads** | 1000+ in week 1 | GitHub release stats |
| **Issues** | < 5 critical bugs in week 1 | GitHub Issues priority |
| **Satisfaction** | > 4/5 star rating | User survey |
| **Adoption** | 50+ active users | Discord analytics |

### Maintenance Plan

**v1.0.1** (1 week post-launch)
- Critical bug fixes
- Security patches
- Documentation improvements

**v1.0.2** (2 weeks post-launch)
- Performance improvements
- Feature requests from feedback
- Stability enhancements

**v1.1.0** (1 month post-launch)
- New features
- Enhanced plugin system
- Performance optimizations

## Risk Management

### High-Risk Items

**1. Critical Bug in Production**
- **Mitigation**: Quick release process (v1.0.1)
- **Rollback**: JS version still available

**2. Migration Data Loss**
- **Mitigation**: Comprehensive migration testing
- **Backup**: Export before migration

**3. Performance Regression**
- **Mitigation**: Continuous benchmarking
- **Monitoring**: Real-user metrics

### Contingency Plans

**If > 20 critical bugs:**
- Delay launch by 1 week
- Focus on bug fixes only

**If migration issues > 10%:**
- Improve migration guide
- Create migration video tutorial
- Offer 1-on-1 migration help

**If performance targets not met:**
- Profile and optimize hot paths
- Reduce feature scope
- Document known limitations

## Team Responsibilities

| Role | Name | Responsibilities |
|------|------|------------------|
| **Release Manager** | TBD | Coordinate release, sign-offs |
| **Engineering Lead** | TBD | Bug triage, fixes |
| **Documentation** | TBD | Docs polish, release notes |
| **Community Manager** | TBD | Beta testers, Discord support |
| **DevOps** | TBD | Build infrastructure, deployments |

## Communication Plan

### Internal Updates
- Daily standups during alpha
- Weekly beta summaries
- Pre-launch briefing

### External Updates
- Beta tester weekly emails
- Public blog posts
- Social media updates
- Release day announcements

## Conclusion

Phase 13 marks the culmination of the JS to Go migration. With 12 phases complete and comprehensive testing, we're ready for a successful launch.

**Key Success Factors:**
- Thorough alpha testing
- Active beta community
- Rapid bug response
- Clear communication
- User-focused support

**Next Step:** Begin alpha testing immediately.
