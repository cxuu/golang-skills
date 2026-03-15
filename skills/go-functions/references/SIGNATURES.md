# Function Signatures

Detailed rules for formatting Go function signatures, avoiding naked parameters,
and keeping call sites readable.

---

## Single-Line vs Multi-Line

Keep the signature on a single line when it fits comfortably. When it must wrap,
put **all arguments on their own lines** with a trailing comma:

**Bad** — partial wrapping makes alignment brittle:

```go
func (r *SomeType) SomeLongFunctionName(foo1, foo2, foo3 string,
    foo4, foo5, foo6 int) {
    foo7 := bar(foo1)
}
```

**Good** — full wrapping, trailing comma:

```go
func (r *SomeType) SomeLongFunctionName(
    foo1, foo2, foo3 string,
    foo4, foo5, foo6 int,
) {
    foo7 := bar(foo1)
}
```

### Return Values

When return values also need wrapping, follow the same pattern:

```go
func (r *SomeType) LongName(
    foo1, foo2, foo3 string,
    foo4, foo5, foo6 int,
) (
    *Result,
    error,
) {
    // ...
}
```

For simpler cases, named return values can stay on the same line as the closing
paren of parameters:

```go
func (r *SomeType) LongName(
    foo1, foo2, foo3 string,
) (result *Result, err error) {
    // ...
}
```

---

## Shortening Call Sites

Factor out local variables instead of splitting function calls across lines:

```go
// Bad: long inline call
result := foo.Call(
    somePackage.ComplexFunction(arg1, arg2),
    anotherPackage.Transform(data),
    defaultOptions,
)

// Good: factor out locals for clarity
transformed := anotherPackage.Transform(data)
computed := somePackage.ComplexFunction(arg1, arg2)
result := foo.Call(computed, transformed, defaultOptions)
```

This improves readability and makes intermediate values available for debugging.

---

## Avoid Naked Parameters

Naked parameters in function calls hurt readability. Add C-style comments for
ambiguous arguments:

```go
// Bad: what do these booleans mean?
printInfo("foo", true, true)

// Good: inline comments clarify intent
printInfo("foo", true /* isLocal */, true /* done */)
```

Better yet, replace naked `bool` parameters with custom types:

```go
type Region int

const (
    UnknownRegion Region = iota
    Local
)

type Status int

const (
    Pending Status = iota
    Done
)

func printInfo(name string, region Region, status Status)
```

### When to Use Each Approach

| Approach | When |
|----------|------|
| C-style comments | Quick fix; few call sites; third-party API you can't change |
| Custom types | Multiple call sites; public API; more than one bool/int parameter |
| Functional options | 3+ optional parameters; see [go-functional-options](../../go-functional-options/SKILL.md) |

---

## Grouping Related Parameters

When a function takes several parameters of the same type, group them:

```go
// Acceptable: group same-type params
func Copy(dst, src string) error

// Acceptable: separate when meaning differs despite same type
func Move(source string, destination string) error
```

Use grouping when the parameter names make the roles obvious; use separate
declarations when they don't.

---

## Method Receiver Placement

The receiver goes before the function name, formatted like a parameter:

```go
// Short receiver — on the same line
func (s *Server) Start(ctx context.Context) error { ... }

// Long receiver type — consider wrapping if the whole line is too long
func (h *ComplicatedHandler) ServeHTTP(
    w http.ResponseWriter,
    r *http.Request,
) { ... }
```

See [go-naming](../../go-naming/SKILL.md) for receiver naming conventions
(short, one or two letter abbreviations).

---

## Quick Reference

| Topic | Rule |
|-------|------|
| Single-line | Keep on one line when it fits |
| Multi-line | All args on own lines, trailing comma |
| Return wrapping | Same pattern as parameters |
| Call sites | Factor out locals instead of splitting calls |
| Naked bools | Add `/* name */` comments or use custom types |
| Grouped params | Group same-type when names make roles obvious |
| Receiver | Before function name; short abbreviation |
