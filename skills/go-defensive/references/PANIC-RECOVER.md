# Panic and Recover Patterns

> **Source**: Effective Go

## Panic Guidelines

`panic` creates a run-time error that stops the program. Use it only for truly
unrecoverable situations.

### When to Panic

Real library functions should **avoid panic**. If the problem can be masked or
worked around, let things continue rather than taking down the whole program.

```go
// Acceptable: Truly impossible situation
func CubeRoot(x float64) float64 {
    z := x/3
    for i := 0; i < 1e6; i++ {
        prevz := z
        z -= (z*z*z-x) / (3*z*z)
        if veryClose(z, prevz) {
            return z
        }
    }
    // A million iterations has not converged; something is wrong.
    panic(fmt.Sprintf("CubeRoot(%g) did not converge", x))
}
```

### Panic in Initialization

Exception: If a library truly cannot set itself up during `init()`, it may be
reasonable to panic:

```go
var user = os.Getenv("USER")

func init() {
    if user == "" {
        panic("no value for $USER")
    }
}
```

### When Panic IS Acceptable

Beyond initialization, panic is acceptable in these narrow cases:

1. **API misuse** — analogous to how core language panics on out-of-bounds
   access. The `reflect` package uses this approach.
2. **Internal implementation detail with matching `recover`** at the package
   boundary. Panic simplifies deeply-nested control flow while the public API
   still returns errors (the Parse/parseInt pattern below).
3. **`panic("unreachable")`** after `log.Fatal` when the compiler can't detect
   unreachable code.

#### Parse/parseInt Pattern

Use panic internally to unwind complex recursion, but always convert to an error
at the package boundary:

```go
func parseInt(in string) int {
    n, err := strconv.Atoi(in)
    if err != nil {
        panic(&syntaxError{"not a valid integer"})
    }
    return n
}

func Parse(in string) (_ *Node, err error) {
    defer func() {
        if p := recover(); p != nil {
            sErr, ok := p.(*syntaxError)
            if !ok {
                panic(p)  // not ours — re-panic
            }
            err = fmt.Errorf("syntax error: %v", sErr.msg)
        }
    }()
    // ... calls parseInt internally
}
```

**Key**: The type check `p.(*syntaxError)` ensures only *our* panics are caught.
Unexpected panics (nil pointer, etc.) propagate normally.

---

## Recover Patterns

`recover` regains control of a panicking goroutine. It only works inside
deferred functions.

### Basic Recovery Pattern

```go
func safelyDo(work *Work) {
    defer func() {
        if err := recover(); err != nil {
            log.Println("work failed:", err)
        }
    }()
    do(work)
}
```

### Server Goroutine Protection

Isolate panics to individual goroutines in servers:

```go
func server(workChan <-chan *Work) {
    for work := range workChan {
        go safelyDo(work)  // Each worker is protected
    }
}
```

If `do(work)` panics, the result is logged and the goroutine exits cleanly
without disturbing others.

### Package-Internal Panic/Recover

Use panic internally but convert to errors at API boundaries:

```go
// Error is a parse error type
type Error string
func (e Error) Error() string { return string(e) }

// Internal: panic with Error type
func (regexp *Regexp) error(err string) {
    panic(Error(err))
}

// External API: converts panic to error return
func Compile(str string) (regexp *Regexp, err error) {
    regexp = new(Regexp)
    defer func() {
        if e := recover(); e != nil {
            regexp = nil
            err = e.(Error)  // Re-panics if not our Error type
        }
    }()
    return regexp.doParse(str), nil
}
```

**Key points:**

- Deferred functions can modify named return values
- Type assertion `e.(Error)` re-panics on unexpected error types
- Never expose panics to clients—always convert at API boundary

---

## Quick Reference

| Pattern | Description |
|---------|-------------|
| Basic recovery | `defer func() { if err := recover(); err != nil { ... } }()` |
| Server protection | Wrap each goroutine handler in safelyDo |
| Package-internal | Panic internally, recover and return error at API boundary |
| Type-safe recovery | Use type assertion to re-panic on unexpected errors |

## When to Use

- **Panic**: Only for truly unrecoverable situations or init failures
- **Recover**: Server handlers, package-internal error simplification
- **Never**: Expose panics across package boundaries—always convert to errors
