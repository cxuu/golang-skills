---
name: go-declarations
description: Use when declaring or initializing Go variables, constants, structs, or maps — including var vs :=, reducing scope with if-init, formatting composite literals, designing iota enums, and using any instead of interface{}. Also use when writing a new struct or const block, even if the user doesn't ask about declaration style. Does not cover naming conventions (see go-naming).
---

# Go Declarations and Initialization

> Compatibility: Examples may use `any`, which requires Go 1.18+.

## Resource Routing

- `references/SCOPE.md` - Read when deciding between `var`, `:=`, if-init, and narrow variable scope.
- `references/IOTA.md` - Read when designing constants or enum-like values.
- `references/INITIALIZATION.md` - Read when initializing structs, maps, zero values, or pointers.
- `references/LITERALS.md` - Read for composite literal formatting and keyed-field tradeoffs.
- `references/STRUCTS.md` - Read when designing or initializing structs.
- `references/SHADOWING.md` - Read when a declaration may shadow a builtin or outer variable.

## Quick Reference: var vs :=

| Context | Use | Example |
|---------|-----|---------|
| Top-level | `var` (always) | `var _s = F()` |
| Local with value | `:=` | `s := "foo"` |
| Local zero-value (intentional) | `var` | `var filtered []int` |
| Type differs from expression | `var` with type | `var _e error = F()` |

---

## Group Similar Declarations

Group related `var`, `const`, `type` in parenthesized blocks. Separate
**unrelated** declarations into distinct blocks.

```go
// Bad
const a = 1
const b = 2

// Good
const (
    a = 1
    b = 2
)
```

Inside functions, group adjacent vars even if unrelated:

```go
var (
    caller  = c.name
    format  = "json"
    timeout = 5 * time.Second
)
```

---

## Constants and iota

Start enums at one so the zero value represents invalid/unset:

```go
const (
    Add Operation = iota + 1
    Subtract
    Multiply
)
```

Use zero when the default behavior is desirable (e.g., `LogToStdout`).

---

## Variable Scope

Use if-init to limit scope when the result is only needed for the error check:

```go
if err := os.WriteFile(name, data, 0644); err != nil {
    return err
}
```

Don't reduce scope if it forces deeper nesting or you need the result outside
the `if`. Move constants into functions when only used there.

---

## Initializing Structs

- **Always use field names** (enforced by `go vet`). Exception: test tables
  with ≤3 fields.
- **Omit zero-value fields** — let Go set defaults.
- **Use `var` for zero-value structs**: `var user User` not `user := User{}`
- **Use `&T{}` over `new(T)`**: `sptr := &T{Name: "bar"}`

---

## Composite Literal Formatting

Use field names for external package types. Match closing brace indentation
with the opening line. Omit repeated type names in slice/map literals
(`gofmt -s`).

---

## Initializing Maps

| Scenario | Use | Example |
|----------|-----|---------|
| Empty, populated later | `make(map[K]V)` | `m := make(map[string]int)` |
| Nil declaration | `var` | `var m map[string]int` |
| Fixed entries at init | Literal | `m := map[string]int{"a": 1}` |

`make()` visually distinguishes empty-but-initialized from nil. Use size hints
when the count is known.

---

## Raw String Literals

Use backtick strings to avoid hand-escaped characters:

```go
// Bad
wantError := "unknown name:\"test\""

// Good
wantError := `unknown name:"test"`
```

Ideal for regex, SQL, JSON, and multi-line text.

---

## Prefer `any` Over `interface{}`

Go 1.18+: use `any` instead of `interface{}` in all new code.

---

## Avoid Shadowing Built-In Names

Never use predeclared identifiers (`error`, `string`, `len`, `cap`, `append`,
`copy`, `new`, `make`, `close`, `delete`, `panic`, `recover`, `any`, `true`,
`false`, `nil`, `iota`) as names. Use `go vet` to detect.

```go
// Bad — shadows the builtin
var error string

// Good
var errorMessage string
```

---

## Related Skills

- **Naming conventions**: See [go-naming](../go-naming/SKILL.md) when choosing variable names, constant names, or deciding name length by scope
- **Data structures**: See [go-data-structures](../go-data-structures/SKILL.md) when choosing between `new` and `make`, or initializing slices and maps
- **Control flow scoping**: See [go-control-flow](../go-control-flow/SKILL.md) when using if-init, `:=` redeclaration, or avoiding variable shadowing
- **Capacity hints**: See [go-performance](../go-performance/SKILL.md) when pre-allocating maps or slices with known sizes
