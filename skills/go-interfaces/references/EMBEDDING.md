# Embedding Patterns in Go

> **Sources**: Effective Go, Uber Style Guide

Go uses embedding for composition instead of inheritance. Embedding promotes
the inner type's methods to the outer type, satisfying interfaces automatically.

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

## Struct Embedding

Embedding promotes methods from the inner type to the outer type without
explicit forwarding.

```go
type ReadWriter struct {
    *Reader  // *bufio.Reader
    *Writer  // *bufio.Writer
}
```

With embedding, `bufio.ReadWriter` satisfies `io.Reader`, `io.Writer`, and
`io.ReadWriter` automatically.

Mix embedded and named fields:

```go
type Job struct {
    Command string
    *log.Logger
}

job.Println("starting now...")
job.Logger.SetPrefix("Job: ")
```

## Method Overriding

Define a method on the outer type to override the promoted method:

```go
func (job *Job) Printf(format string, args ...any) {
    job.Logger.Printf("%q: %s", job.Command, fmt.Sprintf(format, args...))
}
```

The outer method takes precedence — calls to `job.Printf(...)` invoke the
outer method, while the embedded method is still accessible via
`job.Logger.Printf(...)`.

## Embedding vs Subclassing

When an embedded method is invoked, the receiver is the **inner** type, not the
outer one. The embedded type has no knowledge that it is embedded — there is no
equivalent to `this` or `super` referencing the containing type.

```go
type Base struct{}
func (b *Base) Name() string { return "Base" }

type Derived struct{ Base }

d := Derived{}
d.Name() // returns "Base", not "Derived"
```

## Name Conflict Resolution

1. **Outer hides inner** — Fields or methods on the outer type shadow those
   promoted from an embedded type at the same name
2. **Same-level conflicts are errors** — If two embedded types at the same
   depth promote the same name, it is a compile error (unless the name is
   never accessed)

```go
type A struct{}
func (A) Hello() string { return "A" }

type B struct{}
func (B) Hello() string { return "B" }

type C struct {
    A
    B
}

// c.Hello()  // compile error: ambiguous selector
c.A.Hello()   // OK: explicit disambiguation
```

## Don't Embed in Public Structs

Embedding exposes the inner type's full method set as part of your public API.
This creates a maintenance burden: changes to the embedded type's methods
break your API's compatibility guarantees.

**Bad**
```go
type SMap struct {
    sync.Mutex  // Lock and Unlock are now part of SMap's API
    data map[string]string
}
```

**Good**
```go
type SMap struct {
    mu   sync.Mutex  // unexported field — implementation detail
    data map[string]string
}

func (m *SMap) Get(k string) string {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.data[k]
}
```

Exception: Embedding is acceptable in test types and internal structs where
API stability is not a concern.

## The HandlerFunc Adapter Pattern

Methods can be defined on any named type, not just structs. The
`http.HandlerFunc` pattern converts an ordinary function into an interface
implementation:

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

This adapter pattern is useful whenever you need a single-method interface
satisfied by a standalone function.
