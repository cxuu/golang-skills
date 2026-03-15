---
name: go-error-handling
description: Use when writing Go code that returns, wraps, or handles errors — including choosing between sentinel errors, custom error types, and fmt.Errorf wrapping (%w vs %v), structuring error flow with early returns, and deciding whether to log, return, or match an error. Also use when propagating errors across package boundaries, choosing errors.Is/As vs type assertions, or handling errors from library calls, even if the user doesn't explicitly ask about error strategy.
license: Apache-2.0
compatibility: Requires Go 1.13+ for errors.Is/errors.As and fmt.Errorf %w wrapping
metadata:
  sources: "Google Style Guide, Uber Style Guide"
---

# Go Error Handling

In Go, [errors are values](https://go.dev/blog/errors-are-values) — they are
created by code and consumed by code.

## Choosing an Error Strategy

1. Is this a system boundary (RPC, IPC, storage)? → Wrap with `%v` to avoid leaking internals
2. Does the caller need to match specific error conditions? → Define sentinel or typed error, wrap with `%w`
3. Does the caller just need context for debugging? → Wrap with `fmt.Errorf("...: %w", err)`
4. Is this a leaf function with no wrapping needed? → Return the error directly

Default: wrap with `%w` and place it at the end of the format string.

---

## Returning Errors

### Use the `error` Type

Use `error` to signal that a function can fail. By convention, `error` is the
last result parameter.

```go
// Good:
func Good() error { /* ... */ }

func GoodLookup() (*Result, error) {
    // ...
    if err != nil {
        return nil, err
    }
    return res, nil
}
```

**Never return concrete error types from exported functions** - a concrete `nil`
pointer can become a non-nil interface value:

```go
// Bad: Concrete error type can cause subtle bugs
func Bad() *os.PathError { /*...*/ }

// Good: Always return the error interface
func Good() error { /*...*/ }
```

### Return Values on Error

When a function returns an error, callers must treat all non-error return values
as unspecified unless explicitly documented. Commonly, non-error return values
are their zero values.

**Tip**: Functions taking a `context.Context` should usually return an `error`
so callers can determine if the context was cancelled.

---

## Error Strings

Error strings should **not** be capitalized and should **not** end with
punctuation:

```go
// Bad:
err := fmt.Errorf("Something bad happened.")

// Good:
err := fmt.Errorf("something bad happened")
```

**Exception**: Error strings may start with a capital letter if they begin with
an exported name, proper noun, or acronym.

**Rationale**: Error strings usually appear within other context before being
printed.

For **displayed messages** (logs, test failures, API responses), capitalization
is appropriate:

```go
// Good:
log.Infof("Operation aborted: %v", err)
log.Errorf("Operation aborted: %v", err)
t.Errorf("Op(%q) failed unexpectedly; err=%v", args, err)
```

---

## Handling Errors

Code that encounters an error must make a **deliberate choice** about how to
handle it. Do not discard errors using `_` variables.

When a function returns an error, do one of:

1. **Handle and address the error immediately**
2. **Return the error to the caller**
3. **In exceptional situations**: call `log.Fatal` or (if absolutely necessary)
   `panic`

### Intentionally Ignoring Errors

In rare cases where ignoring an error is appropriate, add a comment explaining
why:

```go
// Good:
var b *bytes.Buffer
n, _ := b.Write(p) // never returns a non-nil error
```

### Using errgroup for Related Operations

When orchestrating related operations where only the first error is useful,
[`errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) provides a
convenient abstraction:

```go
// Good: errgroup handles cancellation and first-error semantics
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return task1(ctx) })
g.Go(func() error { return task2(ctx) })
if err := g.Wait(); err != nil {
    return err
}
```

---

## Avoid In-Band Errors

Do not return special values like `-1`, `nil`, or empty string to signal errors:

```go
// Bad: In-band error value
// Lookup returns the value for key or -1 if there is no mapping for key.
func Lookup(key string) int

// Bad: Caller mistakes can attribute errors to wrong function
return Parse(Lookup(missingKey))
```

Use multiple return values instead:

```go
// Good: Explicit error or ok value
func Lookup(key string) (value string, ok bool)

// Good: Forces caller to handle the error case
value, ok := Lookup(key)
if !ok {
    return fmt.Errorf("no value for %q", key)
}
return Parse(value)
```

This prevents callers from writing `Parse(Lookup(key))` - it causes a
compile-time error since `Lookup(key)` has 2 outputs.

---

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

## Error Types

> **Advisory**: Recommended best practice.

> **Default**: Wrap with `fmt.Errorf("...: %w", err)`. Escalate to sentinel
> errors only when callers need `errors.Is()`. Escalate to custom error types
> only when callers need `errors.As()` to extract structured fields.

If callers need to distinguish different error conditions programmatically, give
errors structure rather than relying on string matching. Choose the right error
type based on whether callers need to match errors and whether messages are
static or dynamic.

**Quick decision table**:

| Caller needs to match? | Message type | Use |
|------------------------|--------------|-----|
| No | static | `errors.New("message")` |
| No | dynamic | `fmt.Errorf("msg: %v", val)` |
| Yes | static | `var ErrFoo = errors.New("...")` |
| Yes | dynamic | custom `error` type |

For detailed coverage of sentinel errors, structured error types, and error
checking patterns, see [references/ERROR-TYPES.md](references/ERROR-TYPES.md).

> Read [references/ERROR-TYPES.md](references/ERROR-TYPES.md) when you need to define a new sentinel error, create a custom error type, or choose between error strategies for a package API.

---

## Error Wrapping

> **Advisory**: Recommended best practice.

> **Default**: Use `%w` within your application code. Use `%v` at system
> boundaries (RPC, IPC, storage) to avoid leaking internal error types to
> external callers.

The choice between `%v` and `%w` significantly impacts how errors are propagated
and inspected:

- **Use `%v`**: At system boundaries, for logging, to hide internal details
- **Use `%w`**: To preserve error chain for `errors.Is`/`errors.As` inspection

**Key rules**:
- Place `%w` at the **end**: `"context message: %w"`
- Add context that callers don't have; don't duplicate existing info
- If annotation adds nothing, just return `err` directly

For detailed coverage of wrapping patterns, placement, adding context, and
logging best practices, see [references/WRAPPING.md](references/WRAPPING.md).

> Read [references/WRAPPING.md](references/WRAPPING.md) when deciding between %v and %w, or when wrapping errors across package boundaries.

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

> **Normative**: Handle an error exactly once — either log it or return it, never both.

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

Prefer structured logging (`slog` in Go 1.21+, or `log/slog`-compatible libraries) over `log.Printf` for production code:

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

---

## See Also

- [go-style-core](../go-style-core/SKILL.md): Core Go style principles and formatting
- [go-naming](../go-naming/SKILL.md): Naming conventions including error naming (ErrFoo)
- [go-testing](../go-testing/SKILL.md): Testing patterns including error testing
- [go-defensive](../go-defensive/SKILL.md): Defensive programming including panic handling
- [go-linting](../go-linting/SKILL.md): Linting tools that catch error handling issues

### Reference Files

- [references/ERROR-TYPES.md](references/ERROR-TYPES.md) - Sentinel errors,
  structured error types, choosing error types, and checking errors
- [references/WRAPPING.md](references/WRAPPING.md) - Error wrapping with %v vs
  %w, placement, adding context, and logging
