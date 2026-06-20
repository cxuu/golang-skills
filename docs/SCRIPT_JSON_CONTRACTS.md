# Script JSON Contracts

Shell scripts keep their current command-line UX and emit stable JSON when
called with `--json`. Exit codes are part of the contract: `0` means no
findings or successful generation, `1` means findings or tool-reported issues,
and `2` means usage/environment errors.

## Findings Scripts

`go-naming/scripts/check-naming.sh`:

```json
{"violations":[{"file":"path","line":1,"rule":"rule-id","message":"text"}],"total":1,"truncated":false}
```

`go-documentation/scripts/check-docs.sh`:

```json
{"missing":[{"file":"path","line":1,"kind":"type","name":"Name"}],"total":1,"truncated":false}
```

`go-error-handling/scripts/check-errors.sh`:

```json
{"findings":[{"file":"path","line":1,"rule":"rule-id","message":"text"}],"total":1,"truncated":false}
```

`go-interfaces/scripts/check-interface-compliance.sh`:

```json
{"interfaces":[{"name":"Reader","file":"path","line":1}],"missing":[{"name":"Reader","file":"path","line":1}],"count_interfaces":1,"count_missing":1,"truncated":false}
```

No-Go-file targets are successful empty scans and include a status marker:

```json
{"violations":[],"total":0,"truncated":false,"status":"no_go_files"}
{"missing":[],"total":0,"truncated":false,"status":"no_go_files"}
{"findings":[],"total":0,"truncated":false,"status":"no_go_files"}
{"interfaces":[],"missing":[],"count_interfaces":0,"count_missing":0,"truncated":false,"status":"no_go_files"}
```

## Tool Scripts

`go-code-review/scripts/pre-review.sh`:

```json
{"gofmt":{"status":"pass","files":[]},"govet":{"status":"pass","output":""},"golangci_lint":{"status":"skip","output":""},"passed":true}
```

`go-linting/scripts/setup-lint.sh`:

```json
{"config_path":".golangci.yml","local_prefix":"","created":true,"lint_issues":false,"lint_output":""}
```

`go-performance/scripts/bench-compare.sh`:

```json
{"count":1,"package":"./...","filter":".","benchmarks_found":1,"baseline":"","save":"","exit_code":0,"output":"Benchmark..."}
```

`go-testing/scripts/gen-table-test.sh`:

```json
{"func":"ParseConfig","package":"config","output_file":"","parallel":false,"written":false}
```

## Migration Note

Keep shell wrappers as the public interface. Replace regex-heavy internals with
Go AST helpers incrementally, starting with checks that need package-aware
analysis: documentation, interface compliance, declaration naming, and error
return flow.
