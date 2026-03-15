#!/usr/bin/env bash
set -euo pipefail

# Generates a .golangci.yml config and runs initial lint.
# Usage: bash scripts/setup-lint.sh [local-prefix]
# Example: bash scripts/setup-lint.sh github.com/myorg/myrepo

LOCAL_PREFIX="${1:-}"

if [ -f .golangci.yml ]; then
    echo "Error: .golangci.yml already exists. Remove it first to regenerate."
    exit 1
fi

cat > .golangci.yml << 'YAML'
linters:
  enable:
    - errcheck
    - goimports
    - revive
    - govet
    - staticcheck

linters-settings:
  revive:
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: error-return
      - name: error-strings
      - name: exported

run:
  timeout: 5m
YAML

if [ -n "$LOCAL_PREFIX" ]; then
    sed -i.bak "s|run:|linters-settings:\n  goimports:\n    local-prefixes: ${LOCAL_PREFIX}\n\nrun:|" .golangci.yml
    rm -f .golangci.yml.bak
fi

echo "Created .golangci.yml"
echo "Running golangci-lint..."
golangci-lint run ./... || echo "Lint issues found — fix them category by category (formatting first, then vet, then style)."
