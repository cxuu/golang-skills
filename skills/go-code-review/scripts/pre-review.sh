#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Run automated pre-review checks on Go code

USAGE
    bash $SCRIPT_NAME [options] [path]

DESCRIPTION
    Runs gofmt, go vet, and golangci-lint against the target path and
    reports any findings. Use before manual code review to catch
    mechanical issues early.

    Exits 0 if all checks pass, 1 if issues found, 2 on error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --force          Run even if golangci-lint is not installed (skip it)

ARGUMENTS
    path             Package pattern to check (default: ./...)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME ./pkg/...
    bash $SCRIPT_NAME --json ./cmd/server/...
    bash $SCRIPT_NAME --force ./...
EOF
}

JSON_OUTPUT=false
FORCE=false
TARGET=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -v|--version) echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --json)       JSON_OUTPUT=true; shift ;;
        --force)      FORCE=true; shift ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)            TARGET="$1"; shift ;;
    esac
done

TARGET="${TARGET:-./...}"

if ! command -v go &>/dev/null; then
    echo "error: go is not installed or not in PATH" >&2
    exit 2
fi

if ! command -v gofmt &>/dev/null; then
    echo "error: gofmt is not installed or not in PATH" >&2
    exit 2
fi

GOFMT_STATUS="pass"
GOFMT_FINDINGS=()
UNFORMATTED=$(gofmt -l . 2>&1) || true
if [[ -n "$UNFORMATTED" ]]; then
    GOFMT_STATUS="fail"
    while IFS= read -r f; do
        [[ -n "$f" ]] && GOFMT_FINDINGS+=("$f")
    done <<< "$UNFORMATTED"
fi

GOVET_STATUS="pass"
GOVET_OUTPUT=""
if ! GOVET_OUTPUT=$(go vet "$TARGET" 2>&1); then
    GOVET_STATUS="fail"
fi

LINT_STATUS="skip"
LINT_OUTPUT=""
if command -v golangci-lint &>/dev/null; then
    LINT_STATUS="pass"
    if ! LINT_OUTPUT=$(golangci-lint run "$TARGET" 2>&1); then
        LINT_STATUS="fail"
    fi
elif ! $FORCE; then
    echo "error: golangci-lint not installed (use --force to skip)" >&2
    exit 2
fi

FAILED=0
[[ "$GOFMT_STATUS" == "fail" ]] && FAILED=1
[[ "$GOVET_STATUS" == "fail" ]] && FAILED=1
[[ "$LINT_STATUS" == "fail" ]] && FAILED=1

if $JSON_OUTPUT; then
    GOFMT_JSON="["
    first=true
    for f in "${GOFMT_FINDINGS[@]+"${GOFMT_FINDINGS[@]}"}"; do
        $first || GOFMT_JSON+=","
        first=false
        GOFMT_JSON+="\"$f\""
    done
    GOFMT_JSON+="]"

    GOVET_ESC="${GOVET_OUTPUT//\"/\\\"}"
    GOVET_ESC="${GOVET_ESC//$'\n'/\\n}"
    LINT_ESC="${LINT_OUTPUT//\"/\\\"}"
    LINT_ESC="${LINT_ESC//$'\n'/\\n}"

    cat <<EOF
{"gofmt":{"status":"$GOFMT_STATUS","files":$GOFMT_JSON},"govet":{"status":"$GOVET_STATUS","output":"$GOVET_ESC"},"golangci_lint":{"status":"$LINT_STATUS","output":"$LINT_ESC"},"passed":$( [[ $FAILED -eq 0 ]] && echo true || echo false )}
EOF
else
    echo "=== gofmt ==="
    if [[ "$GOFMT_STATUS" == "fail" ]]; then
        echo "Unformatted files:"
        for f in "${GOFMT_FINDINGS[@]}"; do
            echo "  $f"
        done
    else
        echo "OK"
    fi

    echo ""
    echo "=== go vet ==="
    if [[ "$GOVET_STATUS" == "fail" ]]; then
        echo "$GOVET_OUTPUT"
    else
        echo "OK"
    fi

    echo ""
    echo "=== golangci-lint ==="
    if [[ "$LINT_STATUS" == "skip" ]]; then
        echo "Skipped (not installed)"
    elif [[ "$LINT_STATUS" == "fail" ]]; then
        echo "$LINT_OUTPUT"
    else
        echo "OK"
    fi

    echo ""
    if [[ $FAILED -eq 1 ]]; then
        echo "Pre-review checks FAILED — fix issues before manual review."
    else
        echo "All pre-review checks passed."
    fi
fi

exit $FAILED
