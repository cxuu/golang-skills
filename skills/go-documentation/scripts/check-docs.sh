#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="1.1.0"

for arg in "$@"; do
    case "$arg" in
        -h|--help)
            cat <<EOF
check-docs.sh v$VERSION - Check for missing doc comments on exported Go symbols

USAGE
    bash check-docs.sh [options] [path]

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --strict         Also check unexported types/functions
    --limit N        Show at most N results (default: all)
EOF
            exit 0
            ;;
        -v|--version)
            echo "check-docs.sh v$VERSION"
            exit 0
            ;;
    esac
done

if ! command -v go >/dev/null 2>&1; then
    echo "error: go is not installed or not in PATH" >&2
    exit 2
fi

CACHE_ROOT="${XDG_CACHE_HOME:-${HOME:-${TMPDIR:-/tmp}}/.cache}/golang-skills"
if ! mkdir -p "$CACHE_ROOT"; then
    CACHE_ROOT="${TMPDIR:-/tmp}/golang-skills-cache"
    mkdir -p "$CACHE_ROOT"
fi

SRC="$SCRIPT_DIR/check-docs-ast.go"
STAMP="$(cksum "$SRC" | awk '{print $1 "-" $2}')"
BIN="$CACHE_ROOT/check-docs-ast-$STAMP"

if [[ ! -x "$BIN" ]]; then
    GOCACHE="${GOCACHE:-$CACHE_ROOT/go-build}" go build -o "$BIN" "$SRC"
fi

exec "$BIN" "$@"
