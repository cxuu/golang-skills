#!/usr/bin/env bash
set -euo pipefail

# Runs automated pre-review checks before manual code review.
# Usage: bash scripts/pre-review.sh [path]
# Example: bash scripts/pre-review.sh ./pkg/...

TARGET="${1:-./...}"
FAILED=0

echo "=== gofmt ==="
UNFORMATTED=$(gofmt -l . 2>&1) || true
if [ -n "$UNFORMATTED" ]; then
    echo "Unformatted files:"
    echo "$UNFORMATTED"
    FAILED=1
else
    echo "OK"
fi

echo ""
echo "=== go vet ==="
if ! go vet "$TARGET" 2>&1; then
    FAILED=1
fi

echo ""
echo "=== golangci-lint ==="
if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run "$TARGET" 2>&1; then
        FAILED=1
    fi
else
    echo "Warning: golangci-lint not installed. Skipping."
fi

echo ""
if [ "$FAILED" -eq 1 ]; then
    echo "Pre-review checks FAILED — fix issues before manual review."
    exit 1
else
    echo "All pre-review checks passed."
fi
