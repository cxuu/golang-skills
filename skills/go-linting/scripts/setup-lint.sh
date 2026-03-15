#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Generate .golangci.yml and run initial lint

USAGE
    bash $SCRIPT_NAME [options] [local-prefix]

DESCRIPTION
    Creates a .golangci.yml with a curated set of linters (errcheck,
    goimports, revive, govet, staticcheck) and runs golangci-lint.
    If local-prefix is provided, configures goimports to group local
    imports separately.

    Exits 0 if lint passes, 1 if lint issues found, 2 on error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --force          Overwrite existing .golangci.yml
    --dry-run        Print generated config to stdout without writing

ARGUMENTS
    local-prefix     Module path prefix for goimports grouping
                     (e.g., github.com/myorg/myrepo)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME github.com/myorg/myrepo
    bash $SCRIPT_NAME --force github.com/myorg/myrepo
    bash $SCRIPT_NAME --dry-run github.com/myorg/myrepo
    bash $SCRIPT_NAME --json
EOF
}

JSON_OUTPUT=false
FORCE=false
DRY_RUN=false
LOCAL_PREFIX=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -v|--version) echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --json)       JSON_OUTPUT=true; shift ;;
        --force)      FORCE=true; shift ;;
        --dry-run)    DRY_RUN=true; shift ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)            LOCAL_PREFIX="$1"; shift ;;
    esac
done

generate_config() {
    cat <<'YAML'
linters:
  enable:
    - errcheck
    - goimports
    - revive
    - govet
    - staticcheck

linters-settings:
YAML

    if [[ -n "$LOCAL_PREFIX" ]]; then
        cat <<YAML
  goimports:
    local-prefixes: ${LOCAL_PREFIX}
YAML
    fi

    cat <<'YAML'
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
}

if $DRY_RUN; then
    generate_config
    exit 0
fi

CONFIG_PATH=".golangci.yml"

if [[ -f "$CONFIG_PATH" ]] && ! $FORCE; then
    echo "error: $CONFIG_PATH already exists (use --force to overwrite)" >&2
    exit 2
fi

generate_config > "$CONFIG_PATH"

LINT_OUTPUT=""
LINT_EXIT=0
if ! command -v golangci-lint &>/dev/null; then
    echo "error: golangci-lint is not installed" >&2
    exit 2
fi

LINT_OUTPUT=$(golangci-lint run ./... 2>&1) || LINT_EXIT=$?

if $JSON_OUTPUT; then
    LINT_ESC="${LINT_OUTPUT//\"/\\\"}"
    LINT_ESC="${LINT_ESC//$'\n'/\\n}"
    CREATED=true
    HAS_ISSUES=$( [[ $LINT_EXIT -ne 0 ]] && echo true || echo false )
    cat <<EOF
{"config_path":"$CONFIG_PATH","local_prefix":"$LOCAL_PREFIX","created":$CREATED,"lint_issues":$HAS_ISSUES,"lint_output":"$LINT_ESC"}
EOF
else
    echo "Created $CONFIG_PATH"
    if [[ $LINT_EXIT -ne 0 ]]; then
        echo ""
        echo "$LINT_OUTPUT"
        echo ""
        echo "Lint issues found — fix them category by category (formatting first, then vet, then style)."
    else
        echo "golangci-lint: all clean."
    fi
fi

if [[ $LINT_EXIT -ne 0 ]]; then
    exit 1
fi
exit 0
