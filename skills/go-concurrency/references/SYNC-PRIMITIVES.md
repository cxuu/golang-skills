# Sync Primitives Patterns

Detailed patterns for mutexes and atomic operations — covering mutex embedding
pitfalls and type-safe atomic access.

---

## Don't Embed Mutexes

If you use a struct by pointer, the mutex should be a non-pointer field. Do not
embed the mutex on the struct, even if the struct is not exported.

```go
// Bad: Embedded mutex exposes Lock/Unlock as part of API
type SMap struct {
    sync.Mutex // Lock() and Unlock() become methods of SMap
    data map[string]string
}

func (m *SMap) Get(k string) string {
    m.Lock()
    defer m.Unlock()
    return m.data[k]
}
```

```go
// Good: Named field keeps mutex as implementation detail
type SMap struct {
    mu   sync.Mutex
    data map[string]string
}

func (m *SMap) Get(k string) string {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.data[k]
}
```

With the bad example, `Lock` and `Unlock` methods are unintentionally part of
the exported API. With the good example, the mutex is an implementation detail
hidden from callers.

---

## Atomic Operations: Full Example

The standard `sync/atomic` package operates on raw types (`int32`, `int64`,
etc.), making it easy to forget to use atomic operations consistently.

```go
// Bad: Easy to forget atomic operation
type foo struct {
    running int32 // atomic
}

func (f *foo) start() {
    if atomic.SwapInt32(&f.running, 1) == 1 {
        return // already running
    }
    // start the Foo
}

func (f *foo) isRunning() bool {
    return f.running == 1 // race! forgot atomic.LoadInt32
}
```

```go
// Good: Type-safe atomic operations
type foo struct {
    running atomic.Bool
}

func (f *foo) start() {
    if f.running.Swap(true) {
        return // already running
    }
    // start the Foo
}

func (f *foo) isRunning() bool {
    return f.running.Load() // can't accidentally read non-atomically
}
```

The `atomic.Bool`, `atomic.Int64`, etc. types (available in stdlib `sync/atomic`
since Go 1.19, or via [go.uber.org/atomic](https://pkg.go.dev/go.uber.org/atomic))
add type safety by hiding the underlying type.

---

## Channel Direction Examples

Specifying direction prevents accidental misuse:

```go
// Good: Direction specified - clear ownership
func sum(values <-chan int) int {
    total := 0
    for v := range values {
        total += v
    }
    return total
}
```

```go
// Bad: No direction - allows accidental misuse
func sum(values chan int) (out int) {
    for v := range values {
        out += v
    }
    close(values) // Bug! This compiles but shouldn't happen.
}
```
