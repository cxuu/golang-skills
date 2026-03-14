---
name: go-code-review
description: Quick-reference checklist for Go code review based on the Go Wiki CodeReviewComments. Maps to detailed skills for comprehensive guidance. Use when reviewing Go code or checking code against community style standards.
sources: [Go Wiki CodeReviewComments, Uber Style Guide]
---

# Go Code Review Checklist

## Review Procedure

1. Run `gofmt -d .` and `go vet ./...` to catch mechanical issues first
2. Read the diff file-by-file; for each file, check the categories below in order
3. Flag issues with specific line references and the rule name
4. After reviewing all files, re-read flagged items to verify they're genuine issues
5. Summarize findings grouped by severity (must-fix, should-fix, nit)

---

## Formatting

- [ ] **gofmt**: Code is formatted with `gofmt` or `goimports` → [go-linting](../go-linting/SKILL.md)

---

## Documentation

- [ ] **Comment sentences**: Comments are full sentences starting with the name being described, ending with a period → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Doc comments**: All exported names have doc comments; non-trivial unexported declarations too → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Package comments**: Package comment appears adjacent to package clause with no blank line → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Named result parameters**: Only used when they clarify meaning (e.g., multiple same-type returns), not just to enable naked returns → [go-documentation](../go-documentation/SKILL.md)

---

## Error Handling

- [ ] **Handle errors**: No discarded errors with `_`; handle, return, or (exceptionally) panic → [go-error-handling](../go-error-handling/SKILL.md)
- [ ] **Error strings**: Lowercase, no punctuation (unless starting with proper noun/acronym) → [go-error-handling](../go-error-handling/SKILL.md)
- [ ] **In-band errors**: No magic values (-1, "", nil); use multiple returns with error or ok bool → [go-error-handling](../go-error-handling/SKILL.md)
- [ ] **Indent error flow**: Handle errors first and return; keep normal path at minimal indentation → [go-error-handling](../go-error-handling/SKILL.md)

---

## Naming

- [ ] **MixedCaps**: Use `MixedCaps` or `mixedCaps`, never underscores; unexported is `maxLength` not `MAX_LENGTH` → [go-naming](../go-naming/SKILL.md)
- [ ] **Initialisms**: Keep consistent case: `URL`/`url`, `ID`/`id`, `HTTP`/`http` (e.g., `ServeHTTP`, `xmlHTTPRequest`) → [go-naming](../go-naming/SKILL.md)
- [ ] **Variable names**: Short names for limited scope (`i`, `r`, `c`); longer names for wider scope → [go-naming](../go-naming/SKILL.md)
- [ ] **Receiver names**: One or two letter abbreviation of type (`c` for `Client`); no `this`, `self`, `me`; consistent across methods → [go-naming](../go-naming/SKILL.md)
- [ ] **Package names**: No stuttering (use `chubby.File` not `chubby.ChubbyFile`); avoid `util`, `common`, `misc` → [go-packages](../go-packages/SKILL.md)
- [ ] **Avoid built-in names**: Don't shadow `error`, `string`, `len`, `cap`, `append`, `copy`, `new`, `make` → [go-declarations](../go-declarations/SKILL.md)

---

## Concurrency

- [ ] **Goroutine lifetimes**: Clear when/whether goroutines exit; document if not obvious → [go-concurrency](../go-concurrency/SKILL.md)
- [ ] **Synchronous functions**: Prefer sync over async; let callers add concurrency if needed → [go-concurrency](../go-concurrency/SKILL.md)
- [ ] **Contexts**: First parameter; not in structs; no custom Context types; pass even if you think you don't need to → [go-context](../go-context/SKILL.md)

---

## Interfaces

- [ ] **Interface location**: Define in consumer package, not implementor; return concrete types from producers → [go-interfaces](../go-interfaces/SKILL.md)
- [ ] **No premature interfaces**: Don't define before used; don't define "for mocking" on implementor side → [go-interfaces](../go-interfaces/SKILL.md)
- [ ] **Receiver type**: Use pointer if mutating, has sync fields, or is large; value for small immutable types; don't mix → [go-interfaces](../go-interfaces/SKILL.md)

---

## Data Structures

- [ ] **Empty slices**: Prefer `var t []string` (nil) over `t := []string{}` (non-nil zero-length) → [go-data-structures](../go-data-structures/SKILL.md)
- [ ] **Copying**: Be careful copying structs with pointer/slice fields; don't copy `*T` methods' receivers by value → [go-data-structures](../go-data-structures/SKILL.md)

---

## Security

- [ ] **Crypto rand**: Use `crypto/rand` for keys, not `math/rand` → [go-defensive](../go-defensive/SKILL.md)
- [ ] **Don't panic**: Use error returns for normal error handling; panic only for truly exceptional cases → [go-defensive](../go-defensive/SKILL.md)

---

## Declarations and Initialization

- [ ] **Group similar**: Related `var`/`const`/`type` in parenthesized blocks; separate unrelated → [go-declarations](../go-declarations/SKILL.md)
- [ ] **var vs :=**: Use `var` for intentional zero values; `:=` for explicit assignments → [go-declarations](../go-declarations/SKILL.md)
- [ ] **Reduce scope**: Move declarations close to usage; use if-init to limit variable scope → [go-declarations](../go-declarations/SKILL.md)
- [ ] **Struct init**: Always use field names; omit zero fields; `var` for zero structs → [go-declarations](../go-declarations/SKILL.md)
- [ ] **Use `any`**: Prefer `any` over `interface{}` in new code → [go-declarations](../go-declarations/SKILL.md)

---

## Functions

- [ ] **File ordering**: Types → constructors → exported methods → unexported → utilities → [go-functions](../go-functions/SKILL.md)
- [ ] **Signature formatting**: All args on own lines with trailing comma when wrapping → [go-functions](../go-functions/SKILL.md)
- [ ] **Naked parameters**: Add `/* name */` comments for ambiguous bool/int args, or use custom types → [go-functions](../go-functions/SKILL.md)
- [ ] **Printf naming**: Functions accepting format strings end in `f` for `go vet` → [go-functions](../go-functions/SKILL.md)

---

## Style

- [ ] **Line length**: No rigid limit, but avoid uncomfortably long lines; break by semantics, not arbitrary length → [go-style-core](../go-style-core/SKILL.md)
- [ ] **Naked returns**: Only in short functions; explicit returns in medium/large functions → [go-style-core](../go-style-core/SKILL.md)
- [ ] **Pass values**: Don't use pointers just to save bytes; pass `string` not `*string` for small fixed-size types → [go-performance](../go-performance/SKILL.md)
- [ ] **String concatenation**: `+` for simple; `fmt.Sprintf` for formatting; `strings.Builder` for loops → [go-performance](../go-performance/SKILL.md)

---

## Imports

- [ ] **Import groups**: Standard library first, then blank line, then external packages → [go-packages](../go-packages/SKILL.md)
- [ ] **Import renaming**: Avoid unless collision; rename local/project-specific import on collision → [go-packages](../go-packages/SKILL.md)
- [ ] **Import blank**: `import _ "pkg"` only in main package or tests → [go-packages](../go-packages/SKILL.md)
- [ ] **Import dot**: Only for circular dependency workarounds in tests → [go-packages](../go-packages/SKILL.md)

---

## Generics

- [ ] **When to use**: Only when multiple types share identical logic and interfaces don't suffice → [go-generics](../go-generics/SKILL.md)
- [ ] **Type aliases**: Use definitions for new types; aliases only for package migration → [go-generics](../go-generics/SKILL.md)

---

## Testing

- [ ] **Examples**: Include runnable `Example` functions or tests demonstrating usage → [go-documentation](../go-documentation/SKILL.md)
- [ ] **Useful test failures**: Messages include what was wrong, inputs, got, and want; order is `got != want` → [go-testing](../go-testing/SKILL.md)
- [ ] **TestMain**: Use only when all tests need common setup with teardown; prefer scoped helpers first → [go-testing](../go-testing/SKILL.md)
- [ ] **Real transports**: Prefer `httptest.NewServer` + real client over mocking HTTP → [go-testing](../go-testing/SKILL.md)

---

## Automated Checks

More important than any "blessed" set of linters: **lint consistently across a codebase**.

### Minimum Recommended Linters

| Linter | Purpose |
|--------|---------|
| [errcheck](https://github.com/kisielk/errcheck) | Ensure errors are handled |
| [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) | Format code and manage imports |
| [revive](https://github.com/mgechev/revive) | Common style mistakes (modern replacement for golint) |
| [govet](https://pkg.go.dev/cmd/vet) | Analyze code for common mistakes |
| [staticcheck](https://staticcheck.dev) | Various static analysis checks |

### golangci-lint Configuration

Use [golangci-lint](https://github.com/golangci/golangci-lint) as your lint runner. Create `.golangci.yml` in your project root:

```yaml
linters:
  enable:
    - errcheck
    - goimports
    - revive
    - govet
    - staticcheck

linters-settings:
  goimports:
    local-prefixes: github.com/your-org/your-repo
  revive:
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: error-return
      - name: error-strings
      - name: exported

run:
  timeout: 5m
```

### Running

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all linters
golangci-lint run

# Run on specific paths
golangci-lint run ./pkg/...
```

---

## Automated Pre-Check

1. Run: `gofmt -l . && go vet ./... && golangci-lint run ./...`
2. If any output, fix those issues first before manual review
3. Re-run until clean, then proceed to the checklist above

---

## See Also

- **go-linting**: Automated tooling for style enforcement
- **go-style-core**: Core Go style principles
- **go-documentation**: Documentation and comment standards
- **go-error-handling**: Error handling patterns
- **go-naming**: Naming conventions
- **go-packages**: Package design and imports
- **go-interfaces**: Interface design patterns
- **go-concurrency**: Concurrency patterns
- **go-context**: Context usage patterns
- **go-data-structures**: Data structure idioms
- **go-defensive**: Defensive programming
- **go-testing**: Testing patterns
- **go-performance**: Performance considerations
- **go-control-flow**: Control flow idioms and blank identifier
- **go-functional-options**: Functional options pattern
- **go-declarations**: Declaration and initialization patterns
- **go-functions**: Function design and file organization
- **go-generics**: Generics and type parameters
