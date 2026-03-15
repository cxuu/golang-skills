#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Run Go benchmarks with optional comparison

USAGE
    bash $SCRIPT_NAME [options] [package]

DESCRIPTION
    Wrapper around 'go test -bench' that runs benchmarks multiple times and
    optionally compares results against a saved baseline using benchstat.

    Results can be saved to a file for future comparison. If benchstat is
    installed and a baseline is provided, a statistical comparison is shown.

    Always exits 0 (informational tool).

OPTIONS
    -h, --help           Show this help message
    -v, --version        Show version
    -n, --count N        Number of benchmark iterations (default: 5)
    -b, --baseline FILE  Compare results against this baseline file
    -s, --save FILE      Save benchmark results to this file
    -f, --filter REGEX   Benchmark filter regex (default: ".")
    --json               Output metadata as JSON
    --benchmem           Include memory allocation stats (default: on)
    --no-benchmem        Disable memory allocation stats

ARGUMENTS
    package              Go package to benchmark (default: ./...)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME -n 10 ./pkg/parser
    bash $SCRIPT_NAME --save baseline.txt ./...
    bash $SCRIPT_NAME --baseline baseline.txt --save current.txt ./...
    bash $SCRIPT_NAME --filter BenchmarkSort -n 3
EOF
}

COUNT=5
BASELINE=""
SAVE=""
FILTER="."
PACKAGE=""
JSON_OUTPUT=false
BENCHMEM=true

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)      usage; exit 0 ;;
        -v|--version)   echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        -n|--count)     COUNT="${2:?error: --count requires a number}"; shift 2 ;;
        -b|--baseline)  BASELINE="${2:?error: --baseline requires a file path}"; shift 2 ;;
        -s|--save)      SAVE="${2:?error: --save requires a file path}"; shift 2 ;;
        -f|--filter)    FILTER="${2:?error: --filter requires a regex}"; shift 2 ;;
        --json)         JSON_OUTPUT=true; shift ;;
        --benchmem)     BENCHMEM=true; shift ;;
        --no-benchmem)  BENCHMEM=false; shift ;;
        -*)             echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)              PACKAGE="$1"; shift ;;
    esac
done

PACKAGE="${PACKAGE:-./...}"

if ! command -v go &>/dev/null; then
    echo "error: 'go' command not found in PATH" >&2
    exit 2
fi

# Validate count
if ! [[ "$COUNT" =~ ^[1-9][0-9]*$ ]]; then
    echo "error: --count must be a positive integer, got: $COUNT" >&2
    exit 2
fi

# Validate baseline exists if specified
if [[ -n "$BASELINE" && ! -f "$BASELINE" ]]; then
    echo "error: baseline file not found: $BASELINE" >&2
    exit 2
fi

HAS_BENCHSTAT=false
if command -v benchstat &>/dev/null; then
    HAS_BENCHSTAT=true
fi

BENCH_ARGS=(-bench "$FILTER" -count "$COUNT" -run '^$')
if $BENCHMEM; then
    BENCH_ARGS+=(-benchmem)
fi

TMPFILE=$(mktemp "${TMPDIR:-/tmp}/bench-XXXXXX.txt")
trap 'rm -f "$TMPFILE"' EXIT

echo "Running benchmarks: go test ${BENCH_ARGS[*]} $PACKAGE"
echo "Iterations: $COUNT"
echo ""

if ! go test "${BENCH_ARGS[@]}" "$PACKAGE" 2>&1 | tee "$TMPFILE"; then
    echo ""
    echo "warning: 'go test' returned non-zero exit code" >&2
fi

# Save results if requested
if [[ -n "$SAVE" ]]; then
    cp "$TMPFILE" "$SAVE"
    echo ""
    echo "Results saved to: $SAVE"
fi

# Compare with baseline if provided
if [[ -n "$BASELINE" ]]; then
    echo ""
    echo "=== Comparison with baseline: $BASELINE ==="
    echo ""
    if $HAS_BENCHSTAT; then
        benchstat "$BASELINE" "$TMPFILE" || true
    else
        echo "note: install benchstat for statistical comparison:"
        echo "  go install golang.org/x/perf/cmd/benchstat@latest"
        echo ""
        echo "--- Baseline ---"
        grep -E '^Benchmark' "$BASELINE" || true
        echo ""
        echo "--- Current ---"
        grep -E '^Benchmark' "$TMPFILE" || true
    fi
fi

if $JSON_OUTPUT; then
    bench_count=$(grep -cE '^Benchmark' "$TMPFILE" || true)
    echo ""
    printf '{"count":%d,"package":"%s","filter":"%s","benchmarks_found":%s,"baseline":"%s","save":"%s"}\n' \
        "$COUNT" "$PACKAGE" "$FILTER" "$bench_count" "$BASELINE" "$SAVE"
fi

exit 0
