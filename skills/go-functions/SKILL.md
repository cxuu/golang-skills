---
name: go-functions
description: Use when organizing functions within a Go file, formatting function signatures, designing return values, or following Printf-style naming conventions — even if the user is just adding a new function. Helps with function grouping and ordering, multi-line signature formatting, avoiding naked parameters, Printf verbs and Stringer interface, and when to use pointers to interfaces (almost never).
license: Apache-2.0
metadata:
  sources: "Effective Go, Google Style Guide, Uber Style Guide"
---

# Go Function Design

---

## Function Grouping and Ordering

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

Keep the signature on a single line when possible. When it must wrap, put **all
arguments on their own lines** with a trailing comma:

**Bad:**

```go
func (r *SomeType) SomeLongFunctionName(foo1, foo2, foo3 string,
    foo4, foo5, foo6 int) {
    foo7 := bar(foo1)
}
```

**Good:**

```go
func (r *SomeType) SomeLongFunctionName(
    foo1, foo2, foo3 string,
    foo4, foo5, foo6 int,
) {
    foo7 := bar(foo1)
}
```

Shorten call sites by factoring out local variables instead of splitting
arbitrarily:

```go
// Good: factor out locals
local := helper(some, parameters, here)
result := foo.Call(list, of, parameters, local)
```

---

## Avoid Naked Parameters

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

func printInfo(name string, region Region, status Status)
```

---

## Pointers to Interfaces

You almost never need a pointer to an interface. Pass interfaces as values — the
underlying data can still be a pointer.

```go
// Bad: pointer to interface is almost always wrong
func process(r *io.Reader) { ... }

// Good: pass the interface value directly
func process(r io.Reader) { ... }
```

If you need interface methods to modify the underlying data, the concrete type
stored in the interface must be a pointer — not the interface itself.

---

## Use `%q` for Strings

The `%q` verb prints strings inside double quotes, making empty strings and
control characters visible:

```go
// Good: %q makes boundaries and special chars visible
fmt.Printf("value %q looks like English text", someText)

// Bad: manually adding quotes
fmt.Printf("value \"%s\" looks like English text", someText)
```

Prefer `%q` in output intended for humans where the value could be empty or
contain control characters.

---

## Format Strings Outside Printf

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

Functions that accept a format string should end in `f`. This lets `go vet`
check format strings automatically:

```go
// Good: go vet checks Wrapf format strings by default
func Wrapf(err error, format string, args ...any) error

// Bad: go vet won't check Wrap's format string
func Wrap(err error, format string, args ...any) error
```

If using a non-standard name:

```bash
go vet -printfuncs=wrapf,statusf
```

See **go-functional-options** when designing a constructor with 3+ optional
parameters.

---

## Printf Verbs and Stringer

### Common Verbs

| Verb | Use |
|------|-----|
| `%v` | Default format (struct fields, slice elements) |
| `%+v` | Struct fields with names: `{Name:alice Age:30}` |
| `%#v` | Go-syntax representation: `main.User{Name:"alice", Age:30}` |
| `%T` | Type of the value: `main.User` |
| `%q` | Quoted string or byte slice: `"hello"` |
| `%x` | Hex encoding: `68656c6c6f` |

### Custom String Representation

Implement `fmt.Stringer` to control how your type appears in `%v` and `%s`:

```go
type Point struct{ X, Y int }

func (p Point) String() string {
    return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

// fmt.Println(Point{1, 2}) → "(1, 2)"
```

### Custom Formatting with fmt.Formatter

For full control over all format verbs, implement `fmt.Formatter`:

```go
func (p Point) Format(f fmt.State, verb rune) {
    switch verb {
    case 'v':
        if f.Flag('+') {
            fmt.Fprintf(f, "X:%d Y:%d", p.X, p.Y)
            return
        }
        fmt.Fprintf(f, "(%d, %d)", p.X, p.Y)
    case 's':
        fmt.Fprintf(f, "(%d, %d)", p.X, p.Y)
    }
}
```

Only implement `fmt.Formatter` when `String()` isn't sufficient — it's rarely needed.

---

## Quick Reference

| Topic | Rule |
|-------|------|
| File ordering | Type -> constructor -> exported -> unexported -> utils |
| Signature wrapping | All args on own lines with trailing comma |
| Naked parameters | Add `/* name */` comments or use custom types |
| Pointers to interfaces | Almost never needed; pass interfaces by value |
| `%q` | Use for human-readable string output |
| Format string storage | Declare as `const` outside Printf calls |
| Printf function names | End with `f` for `go vet` support |
| Printf/Stringer | `%v` default, `%+v` with names, `%#v` Go-syntax; implement `Stringer` for custom output |

---

## See Also

- [go-error-handling](../go-error-handling/SKILL.md): Error return patterns and wrapping
- [go-style-core](../go-style-core/SKILL.md): Line length and formatting principles
- [go-declarations](../go-declarations/SKILL.md): Variable declaration and initialization patterns
- [go-naming](../go-naming/SKILL.md): Function and method naming conventions
- [go-interfaces](../go-interfaces/SKILL.md): Interface design and type assertions
