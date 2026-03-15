---
name: go-data-structures
description: Use when working with Go slices, maps, or arrays — including choosing between new and make, using append and 2D slices, declaring empty slices (nil vs literal for JSON), implementing sets with maps, and copying data at boundaries. Also use when building or manipulating collections, even without asking about allocation idioms.
license: Apache-2.0
metadata:
  sources: "Effective Go, Google Style Guide, Uber Style Guide, Go Wiki CodeReviewComments"
---

# Go Data Structures

---

## Choosing a Data Structure

```
What do you need?
├─ Ordered collection of items
│  ├─ Fixed size known at compile time → Array [N]T
│  └─ Dynamic size → Slice []T
│     ├─ Know approximate size? → make([]T, 0, capacity)
│     └─ Unknown size or nil-safe for JSON? → var s []T (nil)
├─ Key-value lookup
│  └─ Map map[K]V
│     ├─ Know approximate size? → make(map[K]V, capacity)
│     └─ Need a set? → map[T]struct{} (zero-size values)
└─ Need to pass to a function?
   └─ Copy at the boundary if the caller might mutate it
```

---

## Allocation: new vs make

- `new(T)` returns `*T`, zeroed. Useful when the zero value is ready to use
  (e.g., `bytes.Buffer`, `sync.Mutex`).
- `make(T, args)` creates slices, maps, and channels only. Returns an
  initialized (not zeroed) value of type `T` (not `*T`).

Design data structures so the zero value is useful without further
initialization.

---

## Composite Literals

Create and initialize structs, arrays, slices, and maps in one expression.
Always use **field names** for types defined outside the current package:

```go
// Good: named fields — order-independent, resilient to struct changes
f := &File{fd: fd, name: name}

// Bad: positional fields for external types — fragile
r := csv.Reader{',', '#', 4, false, false, false, false}
```

Closing braces must match indentation of the opening line:

```go
// Good: cuddled braces
good := []*Type{{
    Field: "value",
}, {
    Field: "value",
}}

// Good: non-cuddled
good := []*Type{
    {Field: "multi"},
    {Field: "line"},
}

// Bad: closing brace on same line as value
bad := []*Type{
    {Key: "multi"},
    {Key: "line"}}
```

It's safe to return the address of a local variable — the storage survives
after the function returns.

---

## Slices

### The append Function

**Always assign the result** — the underlying array may change:

```go
x := []int{1, 2, 3}
x = append(x, 4, 5, 6)

// Append a slice to a slice
x = append(x, y...)  // Note the ...
```

### Two-Dimensional Slices

**Independent inner slices** (can grow/shrink independently):

```go
picture := make([][]uint8, YSize)
for i := range picture {
    picture[i] = make([]uint8, XSize)
}
```

**Single allocation** (more efficient for fixed sizes):

```go
picture := make([][]uint8, YSize)
pixels := make([]uint8, XSize*YSize)
for i := range picture {
    picture[i], pixels = pixels[:XSize], pixels[XSize:]
}
```

> Read [references/SLICES.md](references/SLICES.md) when debugging unexpected slice behavior, sharing slices across goroutines, or working with slice headers.

### Declaring Empty Slices

Prefer nil slices over empty literals:

```go
// Good: nil slice
var t []string

// Avoid: non-nil but zero-length
t := []string{}
```

Both have `len` and `cap` of zero, but the nil slice is the preferred style.

**Exception for JSON**: A nil slice encodes to `null`, while `[]string{}`
encodes to `[]`. Use non-nil when you need a JSON array.

When designing interfaces, avoid distinguishing between nil and non-nil
zero-length slices.

---

## Maps

### Implementing a Set

Use `map[T]bool` — idiomatic and reads naturally:

```go
attended := map[string]bool{"Ann": true, "Joe": true}
if attended[person] {  // false if not in map
    fmt.Println(person, "was at the meeting")
}
```

---

## Copying

Be careful when copying a struct from another package. If the type has methods
on its pointer type (`*T`), copying the value can cause aliasing bugs.

**General rule:** Do not copy a value of type `T` if its methods are associated
with the pointer type `*T`. This applies to `bytes.Buffer`, `sync.Mutex`,
`sync.WaitGroup`, and types containing them.

```go
// Bad: copying a mutex
var mu sync.Mutex
mu2 := mu  // almost always a bug

// Good: pass by pointer
func increment(sc *SafeCounter) {
    sc.mu.Lock()
    sc.count++
    sc.mu.Unlock()
}
```

---

## Quick Reference

| Topic | Key Point |
|-------|-----------|
| `new(T)` | Returns `*T`, zeroed |
| `make(T)` | Slices, maps, channels only; returns `T`, initialized |
| Composite literals | Use field names for external types; match brace indentation |
| Slices | Always assign `append` result; `nil` slice preferred over `[]T{}` |
| Sets | `map[T]bool` is idiomatic |
| Copying | Don't copy `T` if methods are on `*T`; beware aliasing |

## See Also

- [go-style-core](../go-style-core/SKILL.md): Core Go style principles
- [go-control-flow](../go-control-flow/SKILL.md): Control structures including range
- [go-interfaces](../go-interfaces/SKILL.md): Interface patterns and embedding
- [go-concurrency](../go-concurrency/SKILL.md): Channels and goroutines
