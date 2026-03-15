#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Check Go code for common naming anti-patterns

USAGE
    bash $SCRIPT_NAME [options] [path]

DESCRIPTION
    Scans Go source files for naming violations based on Go style guidelines:
      - SCREAMING_SNAKE_CASE constants (should be MixedCaps)
      - Get-prefixed getter methods (should omit Get)
      - Packages named util/helper/common/misc
      - Receivers named "this" or "self"

    Exits 0 if no violations found, 1 if violations found, 2 on error.

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --limit N        Show at most N results (default: all)

ARGUMENTS
    path             Directory or Go file to check (default: current directory)

EXAMPLES
    bash $SCRIPT_NAME
    bash $SCRIPT_NAME ./cmd/server
    bash $SCRIPT_NAME --json ./pkg/...
    bash $SCRIPT_NAME myfile.go
EOF
}

JSON_OUTPUT=false
LIMIT=0
TARGET=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -v|--version) echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --json)       JSON_OUTPUT=true; shift ;;
        --limit)      LIMIT="${2:?error: --limit requires a number}"; shift 2 ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)            TARGET="$1"; shift ;;
    esac
done

TARGET="${TARGET:-.}"

json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\t'/\\t}"
    s="${s//$'\r'/}"
    s="${s//$'\n'/\\n}"
    printf '%s' "$s"
}

# Resolve target to a list of .go files (exclude _test.go and vendor)
find_go_files() {
    local t="$1"
    if [[ -f "$t" ]]; then
        echo "$t"
    elif [[ -d "$t" ]]; then
        find "$t" -name '*.go' ! -name '*_test.go' ! -path '*/vendor/*' ! -path '*/.git/*' 2>/dev/null
    else
        # Handle ./... style patterns
        local dir="${t%%/...}"
        dir="${dir:-.}"
        if [[ -d "$dir" ]]; then
            find "$dir" -name '*.go' ! -name '*_test.go' ! -path '*/vendor/*' ! -path '*/.git/*' 2>/dev/null
        else
            echo "error: path not found: $t" >&2
            exit 2
        fi
    fi
}

VIOLATIONS=()

add_violation() {
    local file="$1" line="$2" rule="$3" message="$4"
    VIOLATIONS+=("${file}:${line}|${rule}|${message}")
}

# Rule 1: SCREAMING_SNAKE_CASE constants
check_screaming_constants() {
    local file="$1"
    local line_num=0
    while IFS= read -r line; do
        line_num=$((line_num + 1))
        # Match const declarations with ALL_CAPS_SNAKE names (2+ uppercase segments with underscore)
        pat='^[[:space:]]*(const[[:space:]]+)[A-Z][A-Z0-9]*_[A-Z0-9_]+[[:space:]]'
        if [[ "$line" =~ $pat ]]; then
            local name
            name=$(echo "$line" | sed -E -n 's/^[[:space:]]*const[[:space:]]+([A-Z][A-Z0-9]*_[A-Z0-9_]*).*/\1/p')
            if [[ -n "$name" ]]; then
                add_violation "$file" "$line_num" "screaming-const" "constant '$name' uses SCREAMING_SNAKE_CASE; use MixedCaps instead"
            fi
        fi
    done < "$file"
}

# Rule 2: Get-prefixed getter methods
check_get_prefix() {
    local file="$1"
    local line_num=0
    while IFS= read -r line; do
        line_num=$((line_num + 1))
        # Match: func (r Type) GetFoo(...) — exported getter with Get prefix
        local re_get='^[[:space:]]*func[[:space:]]+\([^)]+\)[[:space:]]+Get([A-Z][a-zA-Z0-9]*)\('
        if [[ "$line" =~ $re_get ]]; then
            local method_name="Get${BASH_REMATCH[1]}"
            # Skip GetX where X could be legitimate (e.g., GetByID is not a simple getter)
            # Only flag simple GetField patterns (no preposition after Get)
            case "${BASH_REMATCH[1]}" in
                By*|From*|Or*|With*|All*) continue ;;
            esac
            add_violation "$file" "$line_num" "get-prefix" "method '$method_name' has Get prefix; Go getters should omit Get (use '${BASH_REMATCH[1]}')"
        fi
    done < "$file"
}

# Rule 3: Packages named util/helper/common/misc
check_bad_package_names() {
    local file="$1"
    local line_num=0
    while IFS= read -r line; do
        line_num=$((line_num + 1))
        pat='^package[[:space:]]+(util|utils|helper|helpers|common|misc|shared|base|lib)$'
        if [[ "$line" =~ $pat ]]; then
            local pkg_name="${BASH_REMATCH[1]}"
            add_violation "$file" "$line_num" "bad-package-name" "package '$pkg_name' is too generic; use a specific, descriptive name"
        fi
        # Only check the first package line
        if [[ "$line" =~ ^package[[:space:]] ]]; then
            break
        fi
    done < "$file"
}

# Rule 4: Receivers named "this" or "self"
check_bad_receivers() {
    local file="$1"
    local line_num=0
    while IFS= read -r line; do
        line_num=$((line_num + 1))
        # Match: func (this *Type) or func (self Type)
        pat='^[[:space:]]*func[[:space:]]+\([[:space:]]*(this|self)[[:space:]]'
        if [[ "$line" =~ $pat ]]; then
            local recv="${BASH_REMATCH[1]}"
            add_violation "$file" "$line_num" "bad-receiver" "receiver named '$recv'; use a short 1-2 letter abbreviation of the type instead"
        fi
    done < "$file"
}

FILES=()
while IFS= read -r f; do
    [[ -n "$f" ]] && FILES+=("$f")
done < <(find_go_files "$TARGET")

if [[ ${#FILES[@]} -eq 0 ]]; then
    if $JSON_OUTPUT; then
        echo '{"violations":[],"count":0,"status":"no_go_files"}'
    else
        echo "No Go files found in: $TARGET"
    fi
    exit 0
fi

for file in "${FILES[@]}"; do
    check_screaming_constants "$file"
    check_get_prefix "$file"
    check_bad_package_names "$file"
    check_bad_receivers "$file"
done

# Truncation
TOTAL=${#VIOLATIONS[@]}
TRUNCATED=false
if [[ $LIMIT -gt 0 && $TOTAL -gt $LIMIT ]]; then
    VIOLATIONS=("${VIOLATIONS[@]:0:$LIMIT}")
    TRUNCATED=true
fi

# Output results
if $JSON_OUTPUT; then
    echo "{"
    echo '  "violations": ['
    first=true
    for v in "${VIOLATIONS[@]+"${VIOLATIONS[@]}"}"; do
        IFS='|' read -r location rule message <<< "$v"
        file="${location%%:*}"
        line="${location#*:}"
        $first || echo ","
        first=false
        printf '    {"file":"%s","line":%s,"rule":"%s","message":"%s"}' \
            "$(json_escape "$file")" "$line" "$(json_escape "$rule")" "$(json_escape "$message")"
    done
    echo ""
    echo "  ],"
    printf '  "total": %d,\n' "$TOTAL"
    printf '  "truncated": %s\n' "$TRUNCATED"
    echo "}"
else
    if [[ $TOTAL -eq 0 ]]; then
        echo "No naming violations found."
        exit 0
    fi

    echo "Naming violations found:"
    echo ""
    for v in "${VIOLATIONS[@]}"; do
        IFS='|' read -r location rule message <<< "$v"
        printf "  %s  [%s] %s\n" "$location" "$rule" "$message"
    done
    if $TRUNCATED; then
        echo "  ... and $((TOTAL - LIMIT)) more (use --limit to adjust)"
    fi
    echo ""
    echo "Total: $TOTAL violation(s)"
fi

if [[ $TOTAL -gt 0 ]]; then
    exit 1
fi
exit 0
