---
name: go-control-flow
description: Use when writing conditionals, loops, or switch statements in Go — including if with initialization, early returns to omit else, for loop forms, range, switch without fallthrough, type switches, and blank identifier patterns. Also use when writing even a simple if/else or for loop in Go, to ensure idiomatic structure like guard clauses and proper variable scoping.
license: Apache-2.0
metadata:
  sources: "Effective Go, Google Style Guide"
---

# Go Control Flow

---

## If with Initialization

`if` and `switch` accept an optional initialization statement. Use it to scope
variables to the conditional block:

```go
if err := file.Chmod(0664); err != nil {
    log.Print(err)
    return err
}
```

**Tip**: If you use the variable for more than a few lines after the `if`, move
the declaration out and use a standard `if` statement instead:

```go
x, err := f()
if err != nil {
    return err
}
// lots of code that uses x
```

## Indent Error Flow (Guard Clauses)

When an `if` body ends with `break`, `continue`, `goto`, or `return`, omit the
unnecessary `else`. Keep the success path unindented:

```go
// Good: no else, success path at left margin
f, err := os.Open(name)
if err != nil {
    return err
}
d, err := f.Stat()
if err != nil {
    f.Close()
    return err
}
codeUsing(f, d)
```

```go
// Bad: else clause buries normal flow
f, err := os.Open(name)
if err != nil {
    return err
} else {
    codeUsing(f)  // unnecessarily indented
}
```

---

## Redeclaration and Reassignment

The `:=` short declaration allows redeclaring variables in the same scope:

```go
f, err := os.Open(name)  // declares f and err
d, err := f.Stat()       // declares d, reassigns err (not a new err)
```

A variable `v` may appear in a `:=` declaration even if already declared,
provided:

1. The declaration is in the **same scope** as the existing `v`
2. The value is **assignable** to `v`
3. At least **one other variable** is newly created by the declaration

### Variable Shadowing

**Warning**: If `v` is declared in an outer scope, `:=` creates a **new**
variable that shadows it. This is a common source of bugs:

```go
// Bug: ctx inside the if block shadows the outer ctx
func (s *Server) innerHandler(ctx context.Context, req *pb.MyRequest) *pb.MyResponse {
    if *shortenDeadlines {
        ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
        defer cancel()
    }
    // BUG: ctx here is still the original — the shadowed ctx didn't escape the if block
}

// Fix: use = instead of :=
func (s *Server) innerHandler(ctx context.Context, req *pb.MyRequest) *pb.MyResponse {
    if *shortenDeadlines {
        var cancel func()
        ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
        defer cancel()
    }
    // ctx here is correctly the deadline-capped context
}
```

---

## For Loops

Go unifies `for` and `while` into a single construct. Use `range` to iterate
over arrays, slices, strings, maps, and channels.

### Parallel Assignment in For

Go has no comma operator. Use parallel assignment for multiple loop variables:

```go
for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
    a[i], a[j] = a[j], a[i]
}
```

`++` and `--` are statements, not expressions — they cannot be used in parallel
assignment.

---

## Switch

Go's `switch` has **no automatic fall through** — no need for `break` in each
case. Use explicit `fallthrough` if needed (rare).

### Expression-less Switch

A `switch` with no expression switches on `true`. Use it for clean if-else-if
chains:

```go
func unhex(c byte) byte {
    switch {
    case '0' <= c && c <= '9':
        return c - '0'
    case 'a' <= c && c <= 'f':
        return c - 'a' + 10
    case 'A' <= c && c <= 'F':
        return c - 'A' + 10
    }
    return 0
}
```

### Comma-Separated Cases

Multiple cases can be combined with commas:

```go
func shouldEscape(c byte) bool {
    switch c {
    case ' ', '?', '&', '=', '#', '+', '%':
        return true
    }
    return false
}
```

### Break with Labels

`break` inside a `switch` terminates only the switch, not an enclosing `for`
loop. Use a label to break out of the loop:

```go
Loop:
    for n := 0; n < len(src); n += size {
        switch {
        case src[n] < sizeOne:
            break        // breaks switch only
        case src[n] < sizeTwo:
            if n+1 >= len(src) {
                break Loop   // breaks out of for loop
            }
        }
    }
```

For type switches, see **go-interfaces**: Type Switch.

---

## The Blank Identifier

### Multiple Assignment

Discard unwanted values from multi-value expressions:

```go
if _, err := os.Stat(path); os.IsNotExist(err) {
    fmt.Printf("%s does not exist\n", path)
}
```

**Never discard errors carelessly** — a nil dereference panic will follow:

```go
// Bad: ignoring error will crash if path doesn't exist
fi, _ := os.Stat(path)
if fi.IsDir() { ... }  // nil pointer dereference
```

### Import for Side Effect

Import a package only for its `init()` side effects:

```go
import _ "net/http/pprof"  // registers HTTP handlers
import _ "image/png"       // registers PNG decoder
```

### Interface Compliance Check

Verify at compile time that a type implements an interface:

```go
var _ io.Writer = (*MyType)(nil)
```

See **go-interfaces**: Interface Satisfaction Checks for when to use this pattern.

---

## Switch and Break

Go `switch` cases do **not** fall through by default (unlike C/Java). Each case
body implicitly breaks. Use `fallthrough` only when explicitly needed.

```go
switch n {
case 1:
    fmt.Println("one")
    // no fallthrough — next case is NOT executed
case 2:
    fmt.Println("two")
}
```

A `break` inside a `switch` that is inside a `for` loop breaks out of the
**switch**, not the loop. Use a labeled break to exit the loop:

```go
Loop:
    for _, v := range items {
        switch v.Type {
        case "done":
            break Loop  // breaks the for loop
        case "skip":
            break  // breaks only the switch
        }
    }
```

---

## Quick Reference

| Pattern | Go Idiom |
|---------|----------|
| If initialization | `if err := f(); err != nil { }` |
| Early return | Omit `else` when if body returns |
| Redeclaration | `:=` reassigns if same scope + new var |
| Shadowing trap | `:=` in inner scope creates new variable |
| Parallel assignment | `i, j = i+1, j-1` |
| Expression-less switch | `switch { case cond: }` |
| Comma cases | `case 'a', 'b', 'c':` |
| No fallthrough | Default behavior (explicit `fallthrough` if needed) |
| Break from loop in switch | `break Label` |
| Discard value | `_, err := f()` |
| Side-effect import | `import _ "pkg"` |
| Interface check | `var _ Interface = (*Type)(nil)` |
| switch/break | No fallthrough by default; labeled `break` to exit enclosing loop |

---

## See Also

- [go-style-core](../go-style-core/SKILL.md): Core Go style principles and formatting
- [go-error-handling](../go-error-handling/SKILL.md): Error handling patterns including guard clauses
- [go-naming](../go-naming/SKILL.md): Naming conventions for loop variables and labels
- [go-concurrency](../go-concurrency/SKILL.md): Goroutines, channels, and select statements
- [go-interfaces](../go-interfaces/SKILL.md): Type switches and interface satisfaction checks
