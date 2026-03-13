#!/bin/bash
# Quick Regression Test Suite for RC Validation

echo "Libcode RC Regression Test Suite"
echo ""

# Run Go tests
echo "Running Go tests..."
go test ./internal/... -short || exit 1
go test ./tests/integration/... || exit 1
go test ./tests/e2e/... -run TestE2E_ConfigCompatibility || exit 1

echo ""
echo "✓ All regression tests passed!"
echo "RC is ready for release."
