---
name: go-functions
description: Go function design patterns including multiple return values, file organization, signature formatting, and Printf conventions. Use when writing functions, organizing Go source files, or formatting function signatures.
---

# Go Function Design

> **Sources**: Effective Go, Uber Go Style Guide, Google Go Style Guide

---

## Multiple Return Values

> **Source**: Effective Go

Functions and methods can return multiple values. This eliminates the need for
in-band error returns (like `-1` for EOF) or reference parameters:

```go
func (file *File) Write(b []byte) (n int, err error)
```

Common multi-return patterns:

| Pattern | Use Case |
|---------|----------|
| `(T, error)` | Operation that can fail |
| `(T, bool)` | Lookup with presence check (comma-ok) |
| `(T1, T2)` | Two meaningful results (e.g., position + value) |

```go
func nextInt(b []byte, i int) (int, int) {
    for ; i < len(b) && !isDigit(b[i]); i++ {
    }
    x := 0
    for ; i < len(b) && isDigit(b[i]); i++ {
        x = x*10 + int(b[i]) - '0'
    }
    return x, i
}
```

See **go-error-handling** for detailed error return patterns.

---

## Function Grouping and Ordering

> **Source**: Uber Go Style Guide

Organize functions in a file by these rules:

1. Functions sorted in **rough call order**
2. Functions **grouped by receiver**
3. **Exported** functions appear first, after `struct`/`const`/`var` definitions
4. `NewXxx`/`newXxx` constructors appear right after the type definition
5. Plain utility functions appear toward the end of the file

**Bad:**

```go
func (s *something) Cost() int {
    return calcCost(s.weights)
}

type something struct{ ... }

func calcCost(n []int) int { ... }

func (s *something) Stop() { ... }

func newSomething() *something {
    return &something{}
}
```

**Good:**

```go
type something struct{ ... }

func newSomething() *something {
    return &something{}
}

func (s *something) Cost() int {
    return calcCost(s.weights)
}

func (s *something) Stop() { ... }

func calcCost(n []int) int { ... }
```

---

## Function Signature Formatting

> **Source**: Google Go Style Guide (Normative)

Keep the signature on a single line when possible. When it must wrap, put **all
arguments on their own lines** with a trailing comma:

**Bad:**

```go
func (r *SomeType) SomeLongFunctionName(foo1, foo2, foo3 string,
    foo4, foo5, foo6 int) {
    foo7 := bar(foo1)
    // ...
}
```

**Good:**

```go
func (r *SomeType) SomeLongFunctionName(
    foo1, foo2, foo3 string,
    foo4, foo5, foo6 int,
) {
    foo7 := bar(foo1)
    // ...
}
```

Shorten call sites by factoring out local variables instead of splitting
arbitrarily:

```go
// Good: factor out locals
local := helper(some, parameters, here)
result := foo.Call(list, of, parameters, local)

// Bad: arbitrary line breaks
result := foo.Call(long, list, of, parameters,
    with, arbitrary, line, breaks)
```

For long string literals inside function calls, break after the format string
and group arguments semantically:

```go
// Good: break after format string, group by semantic meaning
log.Warningf("Database key (%q, %d, %q) incompatible in transaction started by (%q, %d, %q)",
    currentCustomer, currentOffset, currentKey,
    txCustomer, txOffset, txKey)
```

---

## Avoid Naked Parameters

> **Source**: Uber Go Style Guide

Naked parameters in function calls hurt readability. Add C-style comments for
ambiguous arguments:

```go
// Bad
printInfo("foo", true, true)

// Good
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
    StatusReady Status = iota + 1
    StatusDone
)

func printInfo(name string, region Region, status Status)
```

---

## Pointers to Interfaces

> **Source**: Uber Go Style Guide

You almost never need a pointer to an interface. Pass interfaces as values—the
underlying data can still be a pointer.

An interface value is two words:

1. A pointer to type-specific information ("type descriptor")
2. A data pointer (stores the value directly if it's a pointer, or a pointer to
   the value otherwise)

```go
// Bad: pointer to interface is almost always wrong
func process(r *io.Reader) { ... }

// Good: pass the interface value directly
func process(r io.Reader) { ... }
```

If you need interface methods to modify the underlying data, the concrete type
stored in the interface must be a pointer—not the interface itself.

---

## Use `%q` for Strings

> **Source**: Google Go Style Guide (Normative)

The `%q` verb prints strings inside double quotes, making empty strings and
control characters visible:

```go
// Good: %q makes boundaries and special chars visible
fmt.Printf("value %q looks like English text", someText)

// Bad: manually adding quotes
fmt.Printf("value \"%s\" looks like English text", someText)
// Also bad:
fmt.Printf("value '%s' looks like English text", someText)
```

Prefer `%q` in output intended for humans where the value could be empty or
contain control characters. `""` stands out clearly; an empty `%s` is invisible.

---

## Format Strings Outside Printf

> **Source**: Uber Go Style Guide

When declaring format strings outside a `Printf`-style call, use `const`. This
enables `go vet` to perform static analysis:

```go
// Bad: variable format string — go vet can't check it
msg := "unexpected values %v, %v\n"
fmt.Printf(msg, 1, 2)

// Good: const format string — go vet can validate
const msg = "unexpected values %v, %v\n"
fmt.Printf(msg, 1, 2)
```

---

## Naming Printf-style Functions

> **Source**: Uber Go Style Guide

Functions that accept a format string should end in `f`. This lets `go vet`
check format strings automatically:

```go
// Good: go vet checks Wrapf format strings by default
func Wrapf(err error, format string, args ...any) error

// Bad: go vet won't check Wrap's format string
func Wrap(err error, format string, args ...any) error
```

If using a non-standard name, you can tell `go vet` to check it:

```bash
go vet -printfuncs=wrapf,statusf
```

---

## Quick Reference

| Topic | Rule | Source |
|-------|------|--------|
| Multiple returns | `(T, error)`, `(T, bool)` patterns | Effective Go |
| File ordering | Type → constructor → exported → unexported → utils | Uber |
| Signature wrapping | All args on own lines with trailing comma | Google |
| Naked parameters | Add `/* name */` comments or use custom types | Uber |
| Pointers to interfaces | Almost never needed; pass interfaces by value | Uber |
| `%q` | Use for human-readable string output | Google |
| Format string storage | Declare as `const` outside Printf calls | Uber |
| Printf function names | End with `f` for `go vet` support | Uber |

---

## See Also

- **go-error-handling**: Error return patterns and wrapping
- **go-style-core**: Line length and formatting principles
- **go-declarations**: Variable declaration and initialization patterns
- **go-naming**: Function and method naming conventions
- **go-interfaces**: Interface design and type assertions
