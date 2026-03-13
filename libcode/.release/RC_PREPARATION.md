# Release Candidate Preparation Guide

**Week 3:** March 15-21, 2026
**Status:** In Progress
**Goal:** Prepare and validate Release Candidate builds

---

## Overview

Week 3 focuses on creating Release Candidate (RC) builds that incorporate beta feedback and are ready for production deployment.

---

## RC Build Process

### Version Strategy

**Version Numbering:**
- Alpha: `0.1.0-alpha`
- Beta: `1.0.0-beta.1`
- RC1: `1.0.0-rc.1`
- RC2 (if needed): `1.0.0-rc.2`
- Stable: `1.0.0`

### Tagging Strategy

```bash
# RC1 Tag
git tag -a v1.0.0-rc.1 -m "Release Candidate 1"

# RC2 Tag (if needed)
git tag -a v1.0.0-rc.2 -m "Release Candidate 2"

# Stable Tag
git tag -a v1.0.0 -m "Stable Release 1.0.0"
```

### Build Commands

```bash
# Set version
export VERSION=1.0.0-rc.1

# Build all platforms
make build-all

# Or with explicit version
make build-all VERSION=1.0.0-rc.1
```

---

## RC Checklist

### Pre-RC Checks

- [ ] All beta critical bugs fixed
- [ ] All beta high-priority issues addressed
- [ ] Performance benchmarks verified
- [ ] All tests passing (100%)
- [ ] Documentation updated
- [ ] Migration guide tested
- [ ] Platform compatibility verified

### RC1 Build

- [ ] Build for all platforms
- [ ] Verify binary signatures
- [ ] Test installation on macOS
- [ ] Test installation on Linux
- [ ] Test installation on Windows
- [ ] Verify desktop app

### RC1 Testing

- [ ] Run full test suite
- [ ] Run regression tests
- [ ] Run stress tests
- [ ] Verify migration from JS
- [ ] Test all CLI commands
- [ ] Test TUI interface
- [ ] Test desktop app

### RC Sign-Off Criteria

RC is ready for release when:
- ✅ Zero critical bugs
- ✅ < 5 high-priority bugs
- ✅ All tests passing
- ✅ Performance targets met
- ✅ No regressions from beta

---

## RC Build Script

Create `scripts/build-rc.sh`:

```bash
#!/bin/bash
set -e

VERSION=${1:-"1.0.0-rc.1"}
BUILD_DIR=".release/rc"
ARTIFACTS_DIR="$BUILD_DIR/artifacts"

echo "Building Libcode $VERSION"

# Clean and create directories
rm -rf "$BUILD_DIR"
mkdir -p "$ARTIFACTS_DIR"

# Build for all platforms
echo "Building for all platforms..."
mkdir -p bin

GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$VERSION" -o bin/libcode-darwin-arm64 ./cmd/libcode
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o bin/libcode-darwin-amd64 ./cmd/libcode
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o bin/libcode-linux-amd64 ./cmd/libcode
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o bin/libcode-windows-amd64.exe ./cmd/libcode

# Copy artifacts
cp bin/libcode-* "$ARTIFACTS_DIR/"

# Create checksums
cd "$ARTIFACTS_DIR"
shasum * > SHA256SUMS
cd -

# Build desktop app (if Wails available)
if command -v wails &> /dev/null; then
    echo "Building desktop app..."
    cd desktop
    wails build -clean -u "$VERSION"
    cp build/bin/libcode-desktop "$ARTIFACTS_DIR/"
    cd ..
fi

# Run tests
echo "Running tests..."
go test ./... -v -race

# Show results
echo ""
echo "Build complete!"
echo "Version: $VERSION"
echo "Artifacts: $ARTIFACTS_DIR"
ls -lh "$ARTIFACTS_DIR"
```

---

## RC Release Notes Template

### RC1 Release Notes

```markdown
# Libcode v1.0.0-rc.1 Release Candidate

**Release Date:** March 15, 2026
**Status:** Release Candidate 1

## What's New Since Beta

### Bug Fixes
- Fixed issue with session export on Windows
- Improved error messages for API failures
- Fixed memory leak in long-running sessions

### Improvements
- 20% faster startup on Linux
- Better migration error handling
- Enhanced TUI keyboard shortcuts

### Known Issues
- Desktop app may have rendering issues on some Linux distros
- TUI accessibility features still in progress

## Testing Status

- All tests: PASS ✅
- Regression tests: PASS ✅
- Stress tests: PASS ✅
- Migration tests: PASS ✅

## Installation

Same as beta. See [Installation Guide](https://github.com/gemone/libcode/blob/main/docs/user/README.md#installation).

## Feedback

Please report any issues at:
https://github.com/gemone/libcode/issues
```

---

## RC Validation Process

### 1. Automated Testing

```bash
# Run all tests
go test ./... -v -race -cover

# Run benchmarks
go test ./tests/benchmark/... -bench=. -benchmem

# Run integration tests
go test ./tests/integration/... -v

# Run E2E tests
go test ./tests/e2e/... -v
```

### 2. Manual Testing

Use the RC testing checklist:

```bash
# Test 1: Fresh install
rm -rf ~/.libcode
./libcode version
./libcode auth login openai
./libcode chat

# Test 2: Migration
./libcode migrate
./libcode session list

# Test 3: All commands
./libcode run "test message"
./libcode session export <id>
./libcode mcp list
./libcode models list

# Test 4: Performance
time ./libcode version
# Check memory usage
```

### 3. Platform Testing

Test on actual hardware:
- [ ] macOS 14 (arm64)
- [ ] macOS 14 (amd64)
- [ ] Ubuntu 22.04
- [ ] Windows 11

---

## RC Decision Process

### Go/No-Go Criteria

**GO for release if:**
- All critical bugs fixed
- No regressions from beta
- All tests passing
- Performance targets met
- At least 48 hours of beta testing completed

**NO-GO if:**
- Critical bugs found
- Regressions detected
- Performance degradation
- Negative user feedback

### RC2 Scenarios

Create RC2 if:
- Critical bug found in RC1
- Performance issue discovered
- Breaking change needed

---

## Timeline

**Day 1-2:** RC1 Build and Testing
- Build RC1
- Run full test suite
- Manual validation

**Day 3-4:** RC1 Release
- Publish RC1 to GitHub
- Announce to beta testers
- Monitor feedback

**Day 5-7:** Evaluation
- Collect feedback
- Fix issues if needed
- Decide on RC2

**End of Week:** Either launch or RC2

---

## Success Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Bugs | < 5 total | GitHub Issues |
| Performance | Same or better | Benchmarks |
| Test Pass Rate | 100% | Test suite |
| User Satisfaction | 95%+ | Feedback form |

---

## Next Steps

After RC validation:

1. **If RC passes:** Proceed to Week 4 (Official Launch)
2. **If RC fails:** Create RC2 and repeat validation
3. **If major issues:** Return to bug fixing phase

---

**RC Build Status:** ⏳ In Progress
**Target Date:** March 15, 2026
**Responsible:** Release Team
