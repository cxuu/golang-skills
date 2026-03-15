#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Check for missing compile-time interface compliance verifications

USAGE
    bash $SCRIPT_NAME [options] [path]

DESCRIPTION
    Scans Go files for exported interface definitions and checks whether each
    has a corresponding compile-time compliance assertion like:

        var _ MyInterface = (*MyImpl)(nil)
        var _ MyInterface = MyImpl{}

    Reports interfaces that lack such compile-time checks. This helps catch
    interface drift at compile time instead of runtime.

    Exits 0 if all interfaces are verified, 1 if missing checks found, 2 on error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --include-test   Also scan _test.go files for compliance checks

ARGUMENTS
    path             Directory to scan (default: current directory)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME ./pkg/storage
    bash $SCRIPT_NAME --json .
    bash $SCRIPT_NAME --include-test ./internal
EOF
}

JSON_OUTPUT=false
INCLUDE_TEST=false
TARGET=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)       usage; exit 0 ;;
        -v|--version)    echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --json)          JSON_OUTPUT=true; shift ;;
        --include-test)  INCLUDE_TEST=true; shift ;;
        -*)              echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)               TARGET="$1"; shift ;;
    esac
done

TARGET="${TARGET:-.}"

if [[ ! -d "$TARGET" && ! -f "$TARGET" ]]; then
    # Handle ./... patterns
    dir="${TARGET%%/...}"
    dir="${dir:-.}"
    if [[ ! -d "$dir" ]]; then
        echo "error: path not found: $TARGET" >&2
        exit 2
    fi
    TARGET="$dir"
fi

# Collect all Go source files
find_go_files() {
    local t="$1"
    if $INCLUDE_TEST; then
        find "$t" -name '*.go' ! -path '*/vendor/*' ! -path '*/.git/*' 2>/dev/null
    else
        find "$t" -name '*.go' ! -name '*_test.go' ! -path '*/vendor/*' ! -path '*/.git/*' 2>/dev/null
    fi
}

# Collect all Go files (including tests) for checking compliance vars
find_all_go_files() {
    find "$1" -name '*.go' ! -path '*/vendor/*' ! -path '*/.git/*' 2>/dev/null
}

# Step 1: Find all exported interface definitions
declare -A INTERFACES  # key: "InterfaceName" value: "file:line"

while IFS= read -r file; do
    [[ -n "$file" ]] || continue
    line_num=0
    while IFS= read -r line; do
        line_num=$((line_num + 1))
        # Match: type ExportedName interface {
        if [[ "$line" =~ ^[[:space:]]*type[[:space:]]+([A-Z][a-zA-Z0-9]*)[[:space:]]+interface[[:space:]]*\{ ]]; then
            iface_name="${BASH_REMATCH[1]}"
            INTERFACES["$iface_name"]="$file:$line_num"
        fi
    done < "$file"
done < <(find_go_files "$TARGET")

if [[ ${#INTERFACES[@]} -eq 0 ]]; then
    if $JSON_OUTPUT; then
        echo '{"interfaces":[],"missing":[],"count_interfaces":0,"count_missing":0}'
    else
        echo "No exported interfaces found in: $TARGET"
    fi
    exit 0
fi

# Step 2: Scan all Go files (including tests) for compliance checks
# Pattern: var _ InterfaceName = ...
ALL_CONTENT=""
while IFS= read -r file; do
    [[ -n "$file" ]] || continue
    ALL_CONTENT+="$(cat "$file")"$'\n'
done < <(find_all_go_files "$TARGET")

MISSING=()

for iface_name in "${!INTERFACES[@]}"; do
    location="${INTERFACES[$iface_name]}"
    # Look for: var _ InterfaceName = (various patterns)
    if ! echo "$ALL_CONTENT" | grep -qE "var[[:space:]]+_[[:space:]]+${iface_name}[[:space:]]*="; then
        MISSING+=("${iface_name}|${location}")
    fi
done

# Sort for stable output
IFS=$'\n' MISSING=($(sort <<<"${MISSING[*]}")); unset IFS

# Output results
if $JSON_OUTPUT; then
    echo "{"
    echo '  "interfaces": ['
    first=true
    for iface_name in $(echo "${!INTERFACES[@]}" | tr ' ' '\n' | sort); do
        location="${INTERFACES[$iface_name]}"
        file="${location%%:*}"
        line="${location#*:}"
        $first || echo ","
        first=false
        printf '    {"name":"%s","file":"%s","line":%s}' "$iface_name" "$file" "$line"
    done
    echo ""
    echo "  ],"
    echo '  "missing": ['
    first=true
    for entry in "${MISSING[@]+"${MISSING[@]}"}"; do
        IFS='|' read -r name location <<< "$entry"
        file="${location%%:*}"
        line="${location#*:}"
        $first || echo ","
        first=false
        printf '    {"name":"%s","file":"%s","line":%s}' "$name" "$file" "$line"
    done
    echo ""
    echo "  ],"
    printf '  "count_interfaces": %d,\n' "${#INTERFACES[@]}"
    printf '  "count_missing": %d\n' "${#MISSING[@]}"
    echo "}"
else
    echo "Exported interfaces found: ${#INTERFACES[@]}"
    echo ""

    if [[ ${#MISSING[@]} -eq 0 ]]; then
        echo "All interfaces have compile-time compliance checks."
        exit 0
    fi

    echo "Missing compile-time compliance checks:"
    echo ""
    for entry in "${MISSING[@]}"; do
        IFS='|' read -r name location <<< "$entry"
        printf "  %s  interface '%s' has no 'var _ %s = ...' assertion\n" "$location" "$name" "$name"
    done
    echo ""
    echo "Add compile-time checks like:"
    echo "  var _ MyInterface = (*MyImpl)(nil)"
    echo ""
    echo "Total: ${#MISSING[@]} interface(s) missing verification"
fi

if [[ ${#MISSING[@]} -gt 0 ]]; then
    exit 1
fi
exit 0
