---
name: go-performance
description: Use when optimizing Go code, investigating slow performance, or writing performance-critical sections. Also use when a user mentions slow Go code, string concatenation in loops, or asks about benchmarking — even without explicitly requesting "performance patterns." Helps choose strconv over fmt, avoid repeated string-to-byte conversions, specify map and slice capacity hints, pick the right string concatenation strategy, and decide when to pass values vs pointers.
sources: [Uber Style Guide, Google Style Guide, Go Wiki CodeReviewComments]
---

# Go Performance Patterns

Performance-specific guidelines apply only to the **hot path**. Don't prematurely optimize—focus these patterns where they matter most.

---

## Prefer strconv over fmt

When converting primitives to/from strings, `strconv` is faster than `fmt`.

**Bad:**

```go
for i := 0; i < b.N; i++ {
    s := fmt.Sprint(rand.Int())
}
```

**Good:**

```go
for i := 0; i < b.N; i++ {
    s := strconv.Itoa(rand.Int())
}
```

**Benchmark comparison:**

| Approach | Speed | Allocations |
|----------|-------|-------------|
| `fmt.Sprint` | 143 ns/op | 2 allocs/op |
| `strconv.Itoa` | 64.2 ns/op | 1 allocs/op |

---

## Avoid Repeated String-to-Byte Conversions

Do not create byte slices from a fixed string repeatedly. Instead, perform the conversion once and capture the result.

**Bad:**

```go
for i := 0; i < b.N; i++ {
    w.Write([]byte("Hello world"))
}
```

**Good:**

```go
data := []byte("Hello world")
for i := 0; i < b.N; i++ {
    w.Write(data)
}
```

**Benchmark comparison:**

| Approach | Speed |
|----------|-------|
| Repeated conversion | 22.2 ns/op |
| Single conversion | 3.25 ns/op |

The good version is **~7x faster** because it avoids allocating a new byte slice on each iteration.

---

## Prefer Specifying Container Capacity

Specify container capacity where possible to allocate memory up front. This minimizes subsequent allocations from copying and resizing as elements are added.

### Map Capacity Hints

Provide capacity hints when initializing maps with `make()`.

```go
make(map[T1]T2, hint)
```

**Note**: Unlike slices, map capacity hints do not guarantee complete preemptive allocation—they approximate the number of hashmap buckets required.

**Bad:**

```go
files, _ := os.ReadDir("./files")

m := make(map[string]os.DirEntry)
for _, f := range files {
    m[f.Name()] = f
}
// Map resizes dynamically, causing multiple allocations
```

**Good:**

```go
files, _ := os.ReadDir("./files")

m := make(map[string]os.DirEntry, len(files))
for _, f := range files {
    m[f.Name()] = f
}
// Map is right-sized at initialization, fewer allocations
```

### Slice Capacity

Provide capacity hints when initializing slices with `make()`, particularly when appending.

```go
make([]T, length, capacity)
```

Unlike maps, slice capacity is **not a hint**—the compiler allocates exactly that much memory. Subsequent `append()` operations incur zero allocations until capacity is reached.

**Bad:**

```go
for n := 0; n < b.N; n++ {
    data := make([]int, 0)
    for k := 0; k < size; k++ {
        data = append(data, k)
    }
}
```

**Good:**

```go
for n := 0; n < b.N; n++ {
    data := make([]int, 0, size)
    for k := 0; k < size; k++ {
        data = append(data, k)
    }
}
```

**Benchmark comparison:**

| Approach | Time (100M iterations) |
|----------|------------------------|
| No capacity | 2.48s |
| With capacity | 0.21s |

The good version is **~12x faster** due to zero reallocations during append.

---

## Pass Values

Don't pass pointers as function arguments just to save a few bytes. If a function refers to its argument `x` only as `*x` throughout, then the argument shouldn't be a pointer.

**Common instances where values should be passed directly:**

- Pointer to a string (`*string`) — strings are already small fixed-size headers
- Pointer to an interface value (`*io.Reader`) — interfaces are fixed-size (type + data pointers)

**Bad:**

```go
func process(s *string) {
	fmt.Println(*s)  // only dereferences, never modifies
}
```

**Good:**

```go
func process(s string) {
	fmt.Println(s)
}
```

**Exceptions:**
- Large structs where copying is expensive
- Small structs that might grow in the future

---

## String Concatenation

Choose the right string building strategy based on complexity:

### Use `+` for Simple Cases

```go
key := "projectid: " + p
```

### Use `fmt.Sprintf` for Formatting

```go
// Good: clear formatting
str := fmt.Sprintf("%s [%s:%d]-> %s", src, qos, mtu, dst)

// Bad: + with manual conversions
str := src.String() + " [" + qos.String() + ":" + strconv.Itoa(mtu) + "]-> " + dst.String()
```

When writing to an `io.Writer`, use `fmt.Fprintf` directly instead of building a
temporary string with `fmt.Sprintf`.

### Use `strings.Builder` for Piecemeal Construction

`strings.Builder` takes amortized linear time, whereas repeated `+` or
`fmt.Sprintf` take quadratic time when building a large string:

```go
b := new(strings.Builder)
for i, d := range digitsOfPi {
    fmt.Fprintf(b, "the %d digit of pi is: %d\n", i, d)
}
str := b.String()
```

### Use Backticks for Constant Multi-line Strings

```go
// Good: raw string literal
usage := `Usage:

custom_tool [args]`

// Bad: concatenation with escape sequences
usage := "" +
    "Usage:\n" +
    "\n" +
    "custom_tool [args]"
```

| Method | Best For | Performance |
|--------|----------|-------------|
| `+` | Few strings, simple concat | O(n) for small n |
| `fmt.Sprintf` | Formatted output | Slower, but clearer |
| `strings.Builder` | Loop/piecemeal construction | Amortized O(n) |
| `strings.Join` | Joining a slice | O(n) |
| Backtick literal | Constant multi-line text | Zero cost |

---

## Quick Reference

| Pattern | Bad | Good | Improvement |
|---------|-----|------|-------------|
| Int to string | `fmt.Sprint(n)` | `strconv.Itoa(n)` | ~2x faster |
| Repeated `[]byte` | `[]byte("str")` in loop | Convert once outside | ~7x faster |
| Map initialization | `make(map[K]V)` | `make(map[K]V, size)` | Fewer allocs |
| Slice initialization | `make([]T, 0)` | `make([]T, 0, cap)` | ~12x faster |
| Small fixed-size args | `*string`, `*io.Reader` | `string`, `io.Reader` | No indirection |
| Simple string join | `s1 + " " + s2` | (already good) | Use `+` for few strings |
| Loop string build | Repeated `+=` | `strings.Builder` | O(n) vs O(n^2) |

---

## See Also

- For core style principles: `go-style-core`
- For naming conventions: `go-naming`
- For declaration patterns: `go-declarations`
