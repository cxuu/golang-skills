# Receiver Type: Pointer vs Value

> **Advisory**: Go Wiki CodeReviewComments

Choosing whether to use a value or pointer receiver on methods can be difficult.
**If in doubt, use a pointer**, but there are times when a value receiver makes
sense.

## When to Use Pointer Receiver

- **Method mutates receiver**: The receiver must be a pointer
- **Receiver contains sync.Mutex or similar**: Must be a pointer to avoid copying
- **Large struct or array**: A pointer receiver is more efficient. If passing all
  elements as arguments feels too large, it's too large for a value receiver
- **Concurrent or called methods might mutate**: If changes must be visible in
  the original receiver, it must be a pointer
- **Elements are pointers to something mutating**: Prefer pointer receiver to
  make the intention clearer

## When to Use Value Receiver

- **Small unchanging structs or basic types**: Value receiver for efficiency
- **Map, func, or chan**: Don't use a pointer to them
- **Slice without reslicing/reallocating**: Don't use a pointer if the method
  doesn't reslice or reallocate the slice
- **Small value types with no mutable fields**: Types like `time.Time` with no
  mutable fields and no pointers work well as value receivers
- **Simple basic types**: `int`, `string`, etc.

```go
// Value receiver: small, immutable type
type Point struct {
    X, Y float64
}

func (p Point) Distance(q Point) float64 {
    return math.Hypot(q.X-p.X, q.Y-p.Y)
}

// Pointer receiver: method mutates receiver
func (p *Point) ScaleBy(factor float64) {
    p.X *= factor
    p.Y *= factor
}

// Pointer receiver: contains sync.Mutex
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    c.count++
    c.mu.Unlock()
}
```

## Consistency Rule

**Don't mix receiver types**. Choose either pointers or struct types for all
available methods on a type. If any method needs a pointer receiver, use pointer
receivers for all methods.

```go
// Good: Consistent pointer receivers
type Buffer struct {
    data []byte
}

func (b *Buffer) Write(p []byte) (int, error) { /* ... */ }
func (b *Buffer) Read(p []byte) (int, error)  { /* ... */ }
func (b *Buffer) Len() int                     { return len(b.data) }

// Bad: Mixed receiver types
func (b Buffer) Len() int                      { return len(b.data) }  // inconsistent
```
