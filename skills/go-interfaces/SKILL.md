---
name: go-interfaces
description: Use when defining or implementing Go interfaces, designing abstractions, creating mockable boundaries for testing, or composing types through embedding. Also use when deciding whether to accept an interface or return a concrete type, using type assertions or type switches, or structuring code with the "accept interfaces, return structs" principle — even if the user doesn't explicitly mention interfaces. Helps with implicit satisfaction checks, the comma-ok idiom, embedding patterns, and choosing pointer vs value receivers.
sources: [Effective Go, Google Style Guide, Uber Style Guide]
---

# Go Interfaces and Composition

---

## Accept Interfaces, Return Concrete Types

Interfaces belong in the package that **consumes** values, not the package that
**implements** them. Return concrete (usually pointer or struct) types from
constructors so new methods can be added without refactoring.

```go
// Good: consumer defines the interface it needs
package consumer

type Thinger interface { Thing() bool }

func Foo(t Thinger) string { ... }
```

```go
// Good: producer returns concrete type
package producer

type Thinger struct{ ... }
func (t Thinger) Thing() bool { ... }
func NewThinger() Thinger { return Thinger{ ... } }
```

```go
// Bad: producer defines and returns its own interface
package producer

type Thinger interface { Thing() bool }
type defaultThinger struct{ ... }
func NewThinger() Thinger { return defaultThinger{ ... } }
```

**Do not define interfaces before they are used.** Without a realistic example
of usage, it is too difficult to see whether an interface is even necessary.

---

## Generality: Hide Implementation, Expose Interface

If a type exists only to implement an interface with no exported methods beyond
that interface, return the interface from constructors to hide the implementation:

```go
func NewHash() hash.Hash32 {
    return &myHash{}  // unexported type
}
```

Benefits: implementation can change without affecting callers, substituting
algorithms requires only changing the constructor call.

---

## Type Assertions: Comma-Ok Idiom

Without checking, a failed assertion causes a runtime panic. Always use the
comma-ok idiom to test safely:

```go
str, ok := value.(string)
if ok {
    fmt.Printf("string value is: %q\n", str)
}
```

To check if a value implements an interface:

```go
if _, ok := val.(json.Marshaler); ok {
    fmt.Printf("value %v implements json.Marshaler\n", val)
}
```

---

## Type Switch

A type switch discovers the dynamic type of an interface value:

```go
switch t := t.(type) {
case bool:
    fmt.Printf("boolean %t\n", t)             // t has type bool
case int:
    fmt.Printf("integer %d\n", t)             // t has type int
case *bool:
    fmt.Printf("pointer to boolean %t\n", *t) // t has type *bool
default:
    fmt.Printf("unexpected type %T\n", t)
}
```

It's idiomatic to reuse the variable name (`t := t.(type)`) — the variable has
the correct type in each case branch. When a case lists multiple types
(`case int, int64:`), the variable has the interface type.

Type switches can match both concrete types and interface types:

```go
switch str := value.(type) {
case string:
    return str
case Stringer:
    return str.String()
}
```

---

## Interface Embedding

Combine interfaces by embedding them:

```go
type ReadWriter interface {
    Reader
    Writer
}
```

A `ReadWriter` can do what a `Reader` does *and* what a `Writer` does. Only
interfaces can be embedded within interfaces.

---

## Struct Embedding

Go uses embedding for composition instead of inheritance. Embedding promotes
methods from the inner type to the outer type.

```go
type ReadWriter struct {
    *Reader  // *bufio.Reader
    *Writer  // *bufio.Writer
}
```

With embedding, methods are promoted automatically. `bufio.ReadWriter` satisfies
`io.Reader`, `io.Writer`, and `io.ReadWriter` without explicit forwarding.

Mix embedded and named fields:

```go
type Job struct {
    Command string
    *log.Logger
}

job.Println("starting now...")
job.Logger.SetPrefix("Job: ")
```

### Method Overriding

Define a method on the outer type to override the embedded method:

```go
func (job *Job) Printf(format string, args ...any) {
    job.Logger.Printf("%q: %s", job.Command, fmt.Sprintf(format, args...))
}
```

### Embedding vs Subclassing

When an embedded method is invoked, the receiver is the *inner* type, not the
outer one. The embedded type doesn't know it's embedded.

### Name Conflict Resolution

1. Outer fields/methods hide inner ones at the same name
2. Same-level conflicts are errors (unless never accessed)

---

## Interface Satisfaction Checks

Use a blank identifier assignment to verify a type implements an interface at
compile time:

```go
var _ json.Marshaler = (*RawMessage)(nil)
```

This causes a compile error if `*RawMessage` doesn't implement `json.Marshaler`.

Use this pattern when:
- There are no static conversions that would verify the interface automatically
- The type must satisfy an interface for correct behavior (e.g., custom JSON
  marshaling)
- Interface changes should break compilation, not silently degrade

**Don't** add these checks for every interface — only when no other static
conversion would catch the error.

---

## Methods on Any Named Type

Methods are not limited to structs. The `http.HandlerFunc` adapter pattern
converts an ordinary function into an `http.Handler`:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, req *Request) {
    f(w, req)
}
```

Any function with the right signature becomes an HTTP handler:

```go
http.Handle("/args", http.HandlerFunc(ArgServer))
```

---

## Receiver Type

If in doubt, use a pointer receiver. Don't mix receiver types on a single
type — if any method needs a pointer, use pointers for all methods. Use value
receivers only for small, immutable types (`Point`, `time.Time`) or basic types.

> Read [references/RECEIVER-TYPE.md](references/RECEIVER-TYPE.md) when deciding between pointer and value receivers for a new type, especially for types with sync primitives or large structs.

---

## Quick Reference

| Concept | Pattern | Notes |
|---------|---------|-------|
| Consumer owns interface | Define interfaces where used | Not in the implementing package |
| Safe type assertion | `v, ok := x.(Type)` | Returns zero value + false |
| Type switch | `switch v := x.(type)` | Variable has correct type per case |
| Interface embedding | `type RW interface { Reader; Writer }` | Union of methods |
| Struct embedding | `type S struct { *T }` | Promotes T's methods |
| Interface check | `var _ I = (*T)(nil)` | Compile-time verification |
| Generality | Return interface from constructor | Hide implementation |

---

## See Also

- **go-style-core**: Core Go style principles and formatting
- **go-naming**: Interface naming conventions (Reader, Writer, etc.)
- **go-error-handling**: Error interface and custom error types
- **go-functional-options**: Using interfaces for flexible APIs
- **go-defensive**: Defensive programming patterns
- **go-generics**: When generics suffice vs interfaces
