# Error Flow Patterns

Detailed patterns for error flow, the handle-once principle, and logging
decisions.

## Indent Error Flow

Handle errors before proceeding with normal code. This improves readability by
enabling the reader to find the normal path quickly.

```go
// Good: Error handling first, normal code unindented
if err != nil {
    // error handling
    return // or continue, etc.
}
// normal code
```

```go
// Bad: Normal code hidden in else clause
if err != nil {
    // error handling
} else {
    // normal code that looks abnormal due to indentation
}
```

### Avoid If-with-Initializer for Long-Lived Variables

If you use a variable for more than a few lines, move the declaration out:

```go
// Good: Declaration separate from error check
x, err := f()
if err != nil {
    return err
}
// lots of code that uses x
// across multiple lines
```

```go
// Bad: Variable scoped to else block, hard to read
if x, err := f(); err != nil {
    return err
} else {
    // lots of code that uses x
    // across multiple lines
}
```

---

## Handle Errors Once

When a caller receives an error, it should handle each error **only once**.
Choose ONE response:

1. **Return the error** (wrapped or verbatim) for the caller to handle
2. **Log and degrade gracefully** (don't return the error)
3. **Match and handle** specific error cases, return others

**If you return an error, don't log it yourself** — let the caller handle it.
Logging and returning the same error is the most common "handle errors once"
violation, causing duplicate noise as callers up the stack also handle the error.

```go
// Bad: Logs AND returns - causes noise in logs
u, err := getUser(id)
if err != nil {
    log.Printf("Could not get user %q: %v", id, err)
    return err  // Callers will also log this!
}

// Good: Wrap and return - let caller decide how to handle
u, err := getUser(id)
if err != nil {
    return fmt.Errorf("get user %q: %w", id, err)
}

// Good: Log and degrade gracefully (don't return error)
if err := emitMetrics(); err != nil {
    // Failure to write metrics should not break the application
    log.Printf("Could not emit metrics: %v", err)
}
// Continue execution...

// Good: Match specific errors, return others
tz, err := getUserTimeZone(id)
if err != nil {
    if errors.Is(err, ErrUserNotFound) {
        // User doesn't exist. Use UTC.
        tz = time.UTC
    } else {
        return fmt.Errorf("get user %q: %w", id, err)
    }
}
```

---

## Logging vs Returning Errors

> Handle an error exactly once — either log it or return it, never both.

### Decision Flow

```
Error encountered?
├─ Can the caller act on it? → Return the error (with context via %w)
├─ Is this the top of the call chain? → Log and handle (return HTTP status, exit, etc.)
└─ Neither? → Log at appropriate level and continue
```

### Don't Log and Return

```go
// Bad: error is logged AND returned — appears twice in logs
func process(ctx context.Context, id string) error {
    result, err := fetch(ctx, id)
    if err != nil {
        log.Printf("failed to fetch %s: %v", id, err)
        return fmt.Errorf("fetching %s: %w", id, err)
    }
    return handle(result)
}

// Good: return with context — let the caller decide whether to log
func process(ctx context.Context, id string) error {
    result, err := fetch(ctx, id)
    if err != nil {
        return fmt.Errorf("fetching %s: %w", id, err)
    }
    return handle(result)
}
```

### Structured Logging

Prefer structured logging (`slog` in Go 1.21+, or `log/slog`-compatible
libraries) over `log.Printf` for production code:

```go
// Good: structured fields are machine-parseable
slog.Error("fetch failed", "id", id, "err", err)

// Avoid: unstructured string interpolation
log.Printf("fetch failed for %s: %v", id, err)
```

### Verbosity Levels

| Level | Use for |
|-------|---------|
| Error | Actionable failures that need attention |
| Warn  | Degraded behavior that doesn't require immediate action |
| Info  | Key lifecycle events (startup, shutdown, config loaded) |
| Debug | Diagnostic detail useful during development |
