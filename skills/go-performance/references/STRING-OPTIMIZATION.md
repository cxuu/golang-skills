# String Optimization Patterns

## strconv vs fmt

When converting primitives to/from strings, `strconv` is faster than `fmt`
because `fmt` uses reflection and handles arbitrary types.

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

Common conversions:

| Task | `fmt` | `strconv` |
|------|-------|-----------|
| Int → string | `fmt.Sprint(n)` | `strconv.Itoa(n)` |
| Int64 → string | `fmt.Sprint(n)` | `strconv.FormatInt(n, 10)` |
| Float → string | `fmt.Sprint(f)` | `strconv.FormatFloat(f, 'f', -1, 64)` |
| String → int | — | `strconv.Atoi(s)` |
| Bool → string | `fmt.Sprint(b)` | `strconv.FormatBool(b)` |

---

## Repeated String-to-Byte Conversions

Do not create byte slices from a fixed string repeatedly. Instead, perform the
conversion once and capture the result.

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

The good version is **~7x faster** because it avoids allocating a new byte slice
on each iteration.

---

## String Concatenation

Choose the right string building strategy based on complexity.

### Use `+` for Simple Cases

```go
key := "projectid: " + p
```

The `+` operator is efficient for a small, fixed number of strings. The compiler
can often optimize adjacent string literals.

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

### Strategy Summary

| Method | Best For | Performance |
|--------|----------|-------------|
| `+` | Few strings, simple concat | O(n) for small n |
| `fmt.Sprintf` | Formatted output | Slower, but clearer |
| `strings.Builder` | Loop/piecemeal construction | Amortized O(n) |
| `strings.Join` | Joining a slice | O(n) |
| Backtick literal | Constant multi-line text | Zero cost |
