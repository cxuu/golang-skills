# Global State Patterns

> **Source**: Google Style Guide, Effective Go

Global state makes programs harder to test, reason about, and maintain.
Dependency injection is the preferred alternative, but some global state
is acceptable when used carefully.

## When Global State Is Acceptable

Not all package-level variables are harmful. Global state is appropriate when
it is **truly process-wide** and **not worth injecting**:

- **Default instances** — `http.DefaultClient`, `log.Default()`, `flag.CommandLine`
- **Compiled-once values** — `regexp.MustCompile(...)` at package level
- **Registries** — `database/sql.Register`, `image.RegisterFormat`
- **Singleton infrastructure** — a process-wide metric collector or trace exporter

## Litmus Test for Global Variables

Before adding a package-level variable, ask:

1. **Is it truly process-wide?** If two goroutines or tests might need
   different values, it should not be global
2. **Does it prevent testing?** If tests must save/restore the variable or
   cannot run in parallel because of it, inject it instead
3. **Could it be a constant?** If the value never changes after init, prefer
   `const` or an unexported `var` initialized once
4. **Does it carry mutable state?** Mutable globals are the most dangerous —
   only acceptable for well-documented, concurrency-safe singletons

## Package State API Pattern: New() + Default()

The standard library pattern provides both a customizable constructor and a
convenient default. This lets callers use the default for simple cases and
inject a custom instance for testing or specialized behavior.

**Good**
```go
package mylog

type Logger struct {
    prefix string
    out    io.Writer
}

func New(prefix string, out io.Writer) *Logger {
    return &Logger{prefix: prefix, out: out}
}

var defaultLogger = New("", os.Stderr)

func Default() *Logger { return defaultLogger }

func (l *Logger) Info(msg string) {
    fmt.Fprintf(l.out, "%s%s\n", l.prefix, msg)
}

// Package-level convenience functions delegate to the default instance.
func Info(msg string) { defaultLogger.Info(msg) }
```

```go
// Callers use the default for simple cases
mylog.Info("starting server")

// Tests or specialized code create custom instances
logger := mylog.New("[test] ", &buf)
logger.Info("test message")
```

Standard library examples of this pattern:
- `log.New()` + `log.Default()` + `log.Println()`
- `http.NewServeMux()` + `http.DefaultServeMux`
- `flag.NewFlagSet()` + `flag.CommandLine`

## Dependency Injection as the Preferred Alternative

When code needs configurable behavior, accept dependencies as constructor
parameters or struct fields instead of reading package-level variables.

**Bad**
```go
var db *sql.DB

func GetUser(id int) (*User, error) {
    return db.QueryRow("SELECT ...", id) // depends on global
}
```

**Good**
```go
type UserStore struct {
    db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
    return &UserStore{db: db}
}

func (s *UserStore) GetUser(id int) (*User, error) {
    return s.db.QueryRow("SELECT ...", id)
}
```

Benefits of injection:
- Tests provide mock or in-memory implementations
- Multiple instances can coexist (e.g., read replica vs primary)
- Dependencies are explicit in the constructor signature

## Injecting Time

A common case: replacing `time.Now` for deterministic tests.

**Bad**
```go
func IsExpired(expiry time.Time) bool {
    return time.Now().After(expiry) // untestable
}
```

**Good**
```go
type Checker struct {
    now func() time.Time
}

func NewChecker() *Checker {
    return &Checker{now: time.Now}
}

func (c *Checker) IsExpired(expiry time.Time) bool {
    return c.now().After(expiry)
}
```

Tests replace `now` with a fixed function:

```go
c := &Checker{now: func() time.Time {
    return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}}
```

## Summary

| Situation | Approach |
|-----------|----------|
| Process-wide singleton (logger, metrics) | Default instance + `New()` constructor |
| Compiled-once regex or template | Package-level `var` with `MustCompile` |
| Registry (database drivers, codecs) | Package-level `Register()` function |
| Configurable behavior | Dependency injection via constructor |
| Time-dependent logic | Inject `func() time.Time` |
| Anything tests need to vary | Do not use global state |
