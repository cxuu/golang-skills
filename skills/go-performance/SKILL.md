---
name: go-performance
description: Use when optimizing Go code, investigating slow performance, or writing performance-critical sections. Also use when a user mentions slow Go code, string concatenation in loops, or asks about benchmarking, even if the user doesn't explicitly mention performance patterns. Does not cover concurrent performance patterns (see go-concurrency).
allowed-tools: Bash(bash:*)
---

# Go Performance Patterns

## Resource Routing

- `scripts/bench-compare.sh` - Run when comparing benchmark results, saving baselines, or producing JSON benchmark metadata.
- `references/BENCHMARKS.md` - Read when writing benchmarks, using benchstat, or profiling with pprof.
- `references/STRING-OPTIMIZATION.md` - Read when optimizing string conversion, concatenation, or byte/string boundaries.

Performance-specific guidelines apply only to the **hot path**. Don't prematurely optimize—focus these patterns where they matter most.

---

## Prefer strconv over fmt

When converting primitives to/from strings, `strconv` is faster than `fmt`:

```go
s := strconv.Itoa(rand.Int()) // ~2x faster than fmt.Sprint()
```

| Approach | Speed | Allocations |
|----------|-------|-------------|
| `fmt.Sprint` | 143 ns/op | 2 allocs/op |
| `strconv.Itoa` | 64.2 ns/op | 1 allocs/op |

---

## Avoid Repeated String-to-Byte Conversions

Convert a fixed string to `[]byte` once outside the loop:

```go
data := []byte("Hello world")
for b.Loop() { // Go 1.24+; use b.N loops only for older Go
    w.Write(data) // ~7x faster than []byte("...") each iteration
}
```

---

## Prefer Specifying Container Capacity

Specify container capacity where possible to allocate memory up front. This minimizes subsequent allocations from copying and resizing as elements are added.

### Map Capacity Hints

Provide capacity hints when initializing maps with `make()`:

```go
m := make(map[string]os.DirEntry, len(files))
```

**Note**: Unlike slices, map capacity hints do not guarantee complete preemptive allocation—they approximate the number of hashmap buckets required.

### Slice Capacity

Provide capacity hints when initializing slices with `make()`, particularly when appending:

```go
data := make([]int, 0, size)
```

Unlike maps, slice capacity is **not a hint**—the compiler allocates exactly that much memory. Subsequent `append()` operations incur zero allocations until capacity is reached.

| Approach | Time (100M iterations) |
|----------|------------------------|
| No capacity | 2.48s |
| With capacity | 0.21s |

The capacity version is **~12x faster** due to zero reallocations during append.

---

## Pass Values

Don't pass pointers as function arguments just to save a few bytes. If a function refers to its argument `x` only as `*x` throughout, then the argument shouldn't be a pointer.

```go
func process(s string) { // not *string — strings are small fixed-size headers
    fmt.Println(s)
}
```

**Common pass-by-value types**: `string`, `io.Reader`, small structs.

**Exceptions**:
- Large structs where copying is expensive
- Small structs that might grow in the future

---

## String Concatenation

Choose the right strategy based on complexity:

| Method | Best For |
|--------|----------|
| `+` | Few strings, simple concat |
| `fmt.Sprintf` | Formatted output with mixed types |
| `strings.Builder` | Loop/piecemeal construction |
| `strings.Join` | Joining a slice |
| Backtick literal | Constant multi-line text |

---

## Benchmarking and Profiling

Always measure before and after optimizing. Use Go's built-in benchmark framework and profiling tools.

```bash
go test -bench=. -benchmem -count=10 ./...
```

> **Validation**: After applying optimizations, run `bash scripts/bench-compare.sh` to measure the actual impact. Only keep optimizations with measurable improvement.

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
| Loop string build | Repeated `+=` | `strings.Builder` | O(n) vs O(n²) |

---

## Related Skills

- **Data structures**: See [go-data-structures](../go-data-structures/SKILL.md) when choosing between slices, maps, and arrays, or understanding allocation semantics
- **Declaration patterns**: See [go-declarations](../go-declarations/SKILL.md) when using `make` with capacity hints or initializing maps and slices
- **Concurrency**: See [go-concurrency](../go-concurrency/SKILL.md) when parallelizing work across goroutines or using sync.Pool for buffer reuse
- **Style principles**: See [go-style-core](../go-style-core/SKILL.md) when deciding whether an optimization is worth the readability cost
