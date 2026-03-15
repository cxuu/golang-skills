# Benchmark Methodology

## Writing Benchmarks

Go benchmarks use the `testing.B` type and live in `_test.go` files. The
benchmark function name must start with `Benchmark`.

```go
func BenchmarkStrconv(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := strconv.Itoa(rand.Int())
        _ = s
    }
}

func BenchmarkFmtSprint(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := fmt.Sprint(rand.Int())
        _ = s
    }
}
```

Key rules:
- Use `b.N` as the loop bound — the framework adjusts it for stable timing
- Assign results to a variable (or `_`) to prevent the compiler from
  optimizing away the call
- Use `b.ResetTimer()` after expensive setup that shouldn't be measured
- Use `b.ReportAllocs()` or the `-benchmem` flag for allocation tracking

### Sub-benchmarks

```go
func BenchmarkConvert(b *testing.B) {
    for _, size := range []int{10, 100, 1000} {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            data := make([]byte, size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = string(data)
            }
        })
    }
}
```

---

## Running Benchmarks

```bash
# Run all benchmarks in a package
go test -bench=. ./...

# Run specific benchmark with memory stats
go test -bench=BenchmarkStrconv -benchmem ./...

# Run with count for statistical significance
go test -bench=. -benchmem -count=10 ./...
```

The `-benchmem` flag reports allocations per operation. The `-count` flag runs
each benchmark N times for statistical significance.

---

## Interpreting Results

```
BenchmarkStrconv-8     18705042    64.2 ns/op    16 B/op    1 allocs/op
BenchmarkFmtSprint-8    8249536   143.0 ns/op    16 B/op    2 allocs/op
```

| Field | Meaning |
|-------|---------|
| `-8` | GOMAXPROCS |
| `18705042` | Number of iterations |
| `64.2 ns/op` | Time per operation |
| `16 B/op` | Bytes allocated per operation |
| `1 allocs/op` | Heap allocations per operation |

---

## Using benchstat for Comparison

`benchstat` compares benchmark results statistically. Install it and save
benchmark output to files:

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run benchmarks and save results
go test -bench=. -benchmem -count=10 ./... > old.txt

# Make changes, then run again
go test -bench=. -benchmem -count=10 ./... > new.txt

# Compare results
benchstat old.txt new.txt
```

### Interpreting benchstat Output

```
name          old time/op    new time/op    delta
Strconv-8     64.2ns ± 2%    61.8ns ± 1%   -3.74%  (p=0.001 n=10+10)
```

- **delta**: Percentage change (negative = faster)
- **p-value**: Statistical significance (p < 0.05 is significant)
- **n**: Number of valid samples used

Tips:
- Always use `-count=10` or higher for reliable results
- A small p-value confirms the change is real, not noise
- If benchstat shows `~` (tilde), the difference is not statistically
  significant

---

## Benchmark Examples from Performance Patterns

### strconv vs fmt

| Approach | Speed | Allocations |
|----------|-------|-------------|
| `fmt.Sprint` | 143 ns/op | 2 allocs/op |
| `strconv.Itoa` | 64.2 ns/op | 1 allocs/op |

### Repeated Byte Conversions

```go
func BenchmarkRepeatedConversion(b *testing.B) {
    var buf bytes.Buffer
    for i := 0; i < b.N; i++ {
        buf.Write([]byte("Hello world"))
    }
}

func BenchmarkSingleConversion(b *testing.B) {
    var buf bytes.Buffer
    data := []byte("Hello world")
    for i := 0; i < b.N; i++ {
        buf.Write(data)
    }
}
```

| Approach | Speed |
|----------|-------|
| Repeated conversion | 22.2 ns/op |
| Single conversion | 3.25 ns/op |

### Slice Capacity

```go
func BenchmarkNoCapacity(b *testing.B) {
    for n := 0; n < b.N; n++ {
        data := make([]int, 0)
        for k := 0; k < 1000; k++ {
            data = append(data, k)
        }
    }
}

func BenchmarkWithCapacity(b *testing.B) {
    for n := 0; n < b.N; n++ {
        data := make([]int, 0, 1000)
        for k := 0; k < 1000; k++ {
            data = append(data, k)
        }
    }
}
```

| Approach | Time (100M iterations) |
|----------|------------------------|
| No capacity | 2.48s |
| With capacity | 0.21s |

---

## Profiling with pprof

Use `pprof` to identify bottlenecks before optimizing. Benchmarks measure
improvement; pprof finds where to improve.

### CPU Profiling

```bash
# Generate a CPU profile from benchmarks
go test -bench=BenchmarkHotPath -cpuprofile=cpu.prof ./...

# Analyze with pprof
go tool pprof cpu.prof
```

Common pprof commands:

```
(pprof) top10          # Top 10 functions by CPU time
(pprof) list funcName  # Annotated source for a function
(pprof) web            # Interactive graph in browser
```

### Memory Profiling

```bash
# Generate a memory profile
go test -bench=BenchmarkHotPath -memprofile=mem.prof ./...

# Analyze allocations
go tool pprof -alloc_space mem.prof
```

### HTTP Profiling for Running Services

```go
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    // ... application code ...
}
```

Access profiles at `http://localhost:6060/debug/pprof/`.

### Profiling Workflow

1. **Benchmark** the suspected hot path
2. **Profile** with pprof to confirm where time is spent
3. **Optimize** using patterns from this skill
4. **Re-benchmark** to verify improvement with benchstat
5. **Re-profile** to check for new bottlenecks

---

## Common Mistakes

### Ignoring b.N

The testing framework adjusts `b.N` to get stable timing. Using a fixed
iteration count produces meaningless results:

```go
// Bad: Ignores b.N — benchmark framework can't calibrate
func BenchmarkFixed(b *testing.B) {
    for i := 0; i < 1000; i++ {
        doWork()
    }
}

// Good: Use b.N as the loop bound
func BenchmarkCorrect(b *testing.B) {
    for i := 0; i < b.N; i++ {
        doWork()
    }
}
```

### Not preventing compiler elision

If the result of a function call is unused, the compiler may optimize the call
away entirely. Assign results to a package-level variable:

```go
// Bad: Compiler may optimize away the call
func BenchmarkElided(b *testing.B) {
    for i := 0; i < b.N; i++ {
        expensiveFunc()
    }
}

// Good: Assign to package-level var to prevent elision
var benchResult int

func BenchmarkKept(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = expensiveFunc()
    }
    benchResult = r
}
```
