#!/usr/bin/env bash
set -euo pipefail

# Generates a table-driven test scaffold for a Go function.
# Usage: bash scripts/gen-table-test.sh <FuncName> <package>
# Example: bash scripts/gen-table-test.sh ParseConfig config
#
# Output: writes to stdout. Redirect to a file:
#   bash scripts/gen-table-test.sh ParseConfig config > config/parse_config_test.go

FUNC="${1:?Usage: gen-table-test.sh <FuncName> <package>}"
PKG="${2:?Usage: gen-table-test.sh <FuncName> <package>}"

cat << EOF
package ${PKG}

import (
	"testing"
)

func Test${FUNC}(t *testing.T) {
	tests := []struct {
		name string
		give string // TODO: replace with actual input type
		want string // TODO: replace with actual output type
	}{
		{
			name: "basic case",
			give: "",
			want: "",
		},
		// TODO: add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ${FUNC}(tt.give)
			if got != tt.want {
				t.Errorf("${FUNC}(%q) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}
EOF
