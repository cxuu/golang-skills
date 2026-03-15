---
name: go-defensive
description: Use when hardening Go code at API boundaries — copying slices and maps to prevent mutation, verifying interface compliance at compile time, using defer for cleanup, handling time correctly with time.Time and time.Duration, avoiding mutable globals, or designing enums that start at one. Also use when reviewing Go code for robustness concerns like missing cleanup, shared mutable state, or unsafe crypto usage — even if the user doesn't mention "defensive programming" or "hardening" explicitly.
license: Apache-2.0
metadata:
  sources: "Effective Go, Uber Style Guide, Go Wiki CodeReviewComments"
---

# Go Defensive Programming Patterns

## Defensive Checklist Priority

When hardening code at API boundaries, check in this order:

```
Reviewing an API boundary?
├─ 1. Error handling     → Return errors; don't panic (see go-error-handling)
├─ 2. Input validation   → Copy slices/maps received from callers
├─ 3. Output safety      → Copy slices/maps before returning to callers
├─ 4. Resource cleanup   → Use defer for Close/Unlock/Cancel
├─ 5. Interface checks   → var _ Interface = (*Type)(nil) for compile-time verification
├─ 6. Time correctness   → Use time.Time and time.Duration, not int/float
├─ 7. Enum safety        → Start iota at 1 so zero-value is invalid
└─ 8. Crypto safety      → crypto/rand for keys, never math/rand
```

---

## Verify Interface Compliance

Use compile-time checks to verify interface implementation. See **go-interfaces**: Interface Satisfaction Checks for the full pattern and guidance on when to use it.

```go
var _ http.Handler = (*Handler)(nil)
```

## Copy Slices and Maps at Boundaries

Slices and maps contain pointers. Copy at API boundaries to prevent unintended modifications.

### Receiving

**Bad**
```go
func (d *Driver) SetTrips(trips []Trip) {
  d.trips = trips  // caller can still modify d.trips
}
```

**Good**
```go
func (d *Driver) SetTrips(trips []Trip) {
  d.trips = make([]Trip, len(trips))
  copy(d.trips, trips)
}
```

### Returning

**Bad**
```go
func (s *Stats) Snapshot() map[string]int {
  s.mu.Lock()
  defer s.mu.Unlock()
  return s.counters  // exposes internal state!
}
```

**Good**
```go
func (s *Stats) Snapshot() map[string]int {
  s.mu.Lock()
  defer s.mu.Unlock()
  result := make(map[string]int, len(s.counters))
  for k, v := range s.counters {
    result[k] = v
  }
  return result
}
```

## Defer to Clean Up

Use `defer` to clean up resources (files, locks). Avoids missed cleanup on multiple returns.

**Bad**
```go
p.Lock()
if p.count < 10 {
  p.Unlock()
  return p.count
}
p.count++
newCount := p.count
p.Unlock()
return newCount  // easy to miss unlocks
```

**Good**
```go
p.Lock()
defer p.Unlock()

if p.count < 10 {
  return p.count
}
p.count++
return p.count
```

Defer overhead is negligible. Only avoid in nanosecond-critical paths.

### Defer for File Operations

Place `defer f.Close()` immediately after opening a file for clarity:

```go
func Contents(filename string) (string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    defer f.Close()  // Close sits near Open - much clearer

    var result []byte
    buf := make([]byte, 100)
    for {
        n, err := f.Read(buf[0:])
        result = append(result, buf[0:n]...)
        if err != nil {
            if err == io.EOF {
                break
            }
            return "", err  // f will be closed
        }
    }
    return string(result), nil  // f will be closed
}
```

### Defer Argument Evaluation

Arguments to deferred functions are evaluated when `defer` executes, not when the
deferred function runs:

```go
for i := 0; i < 5; i++ {
    defer fmt.Printf("%d ", i)
}
// Prints: 4 3 2 1 0 (LIFO order, values captured at defer time)
```

### Defer LIFO Order

Multiple defers execute in Last-In-First-Out order:

```go
func trace(s string) string {
    fmt.Println("entering:", s)
    return s
}

func un(s string) {
    fmt.Println("leaving:", s)
}

func a() {
    defer un(trace("a"))  // trace() runs now, un() runs at return
    fmt.Println("in a")
}
// Output: entering: a, in a, leaving: a
```

## Start Enums at One

Start enums at non-zero to distinguish uninitialized from valid values.

**Bad**
```go
const (
  Add Operation = iota  // Add=0, zero value looks valid
  Subtract
  Multiply
)
```

**Good**
```go
const (
  Add Operation = iota + 1  // Add=1, zero value = uninitialized
  Subtract
  Multiply
)
```

**Exception**: When zero is the sensible default (e.g., `LogToStdout = iota`).

## Time, Struct Tags, and Embedding

> For detailed guidance on using `time.Time`/`time.Duration` instead of raw ints, field tags in marshaled structs, and avoiding embedded types in public structs, see [references/TIME-ENUMS-TAGS.md](references/TIME-ENUMS-TAGS.md).

## Avoid Mutable Globals

Use dependency injection instead of mutable globals.

**Bad**
```go
var _timeNow = time.Now

func sign(msg string) string {
  now := _timeNow()
  return signWithTime(msg, now)
}

// Test requires save/restore of global
func TestSign(t *testing.T) {
  oldTimeNow := _timeNow
  _timeNow = func() time.Time { return someFixedTime }
  defer func() { _timeNow = oldTimeNow }()
  assert.Equal(t, want, sign(give))
}
```

**Good**
```go
type signer struct {
  now func() time.Time
}

func newSigner() *signer {
  return &signer{now: time.Now}
}

func (s *signer) Sign(msg string) string {
  now := s.now()
  return signWithTime(msg, now)
}

// Test injects dependency cleanly
func TestSigner(t *testing.T) {
  s := newSigner()
  s.now = func() time.Time { return someFixedTime }
  assert.Equal(t, want, s.Sign(give))
}
```

---

## Crypto Rand

Do not use `math/rand` or `math/rand/v2` to generate keys, even throwaway ones. This is a **security concern**.

Unseeded or time-seeded random generators have predictable output:
- `Time.Nanoseconds()` provides only a few bits of entropy
- Keys generated this way can be guessed by attackers

**Use `crypto/rand` instead:**

```go
import (
	"crypto/rand"
)

func Key() string {
	return rand.Text()
}
```

For text output:
- Use `crypto/rand.Text` directly (preferred)
- Or encode random bytes with `encoding/hex` or `encoding/base64`

---

## Panic and Recover

Use `panic` only for truly unrecoverable situations. Library functions should avoid panic—if the problem can be worked around, let things continue rather than taking down the whole program.

Use `recover` to regain control of a panicking goroutine (only works inside deferred functions):

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

**Key rules:**
- Never expose panics across package boundaries—always convert to errors
- Acceptable to panic in `init()` if a library truly cannot set itself up
- Use recover to isolate panics in server goroutine handlers

For detailed patterns including server protection and package-internal panic/recover, see [references/PANIC-RECOVER.md](references/PANIC-RECOVER.md).

> Read [references/PANIC-RECOVER.md](references/PANIC-RECOVER.md) when writing panic recovery in HTTP servers, using panic as an internal control flow mechanism in parsers, or deciding between log.Fatal and panic.

---

## Must Functions

`Must` functions panic on error — use them **only** for program initialization
where failure means the program cannot run.

```go
// Good: MustCompile in package-level var — panics at startup if invalid
var validID = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

// Good: template.Must for compile-time template parsing
var tmpl = template.Must(template.ParseFiles("index.html"))
```

### When to Use Must

```
Is this called during program initialization (package-level var, init, main setup)?
├─ Yes → Is failure unrecoverable (config, regex, template)?
│        ├─ Yes → Must is appropriate
│        └─ No  → Return error instead
└─ No  → Never use Must — return error
```

### Writing a Must Function

```go
func MustParseConfig(path string) *Config {
    cfg, err := ParseConfig(path)
    if err != nil {
        panic(fmt.Sprintf("parsing config %s: %v", path, err))
    }
    return cfg
}
```

Name them `MustX` where `X` is the fallible function name. Document that they panic.

---

## See Also

- [go-style-core](../go-style-core/SKILL.md): Core Go style principles
- [go-concurrency](../go-concurrency/SKILL.md): Goroutine and channel patterns
- [go-error-handling](../go-error-handling/SKILL.md): Error handling best practices
