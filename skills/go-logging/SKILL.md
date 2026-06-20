---
name: go-logging
description: Use when choosing a logging approach, configuring slog, writing structured log statements, or deciding log levels in Go. Also use when setting up production logging, adding request-scoped context to logs, or migrating from log to slog, even if the user doesn't explicitly mention logging. Does not cover error handling strategy (see go-error-handling).
---

# Go Logging

> Compatibility: `log/slog` requires Go 1.21+; `testing/slogtest` requires Go 1.22+.

## Resource Routing

- `references/LEVELS-AND-CONTEXT.md` - Read when choosing log levels, deciding logger-in-context versus explicit parameters, or excluding sensitive fields.
- `references/LOGGING-PATTERNS.md` - Read when configuring slog handlers, logging HTTP requests, testing handlers, or migrating from `log.Printf`.

## Core Principle

Logs are for **operators**, not developers. Every log line should help someone
diagnose a production issue. If it doesn't serve that purpose, it's noise.

---

## Choosing a Logger

> **Normative**: Use `log/slog` for new Go code.

`slog` is structured, leveled, and in the standard library (Go 1.21+). It
covers the vast majority of production logging needs.

```
Which logger?
├─ New production code      → log/slog
├─ Trivial CLI / one-off    → log (standard)
└─ Measured perf bottleneck → zerolog or zap (benchmark first)
```

Do not introduce a third-party logging library unless profiling shows `slog`
is a bottleneck in your hot path. When you do, keep the same structured
key-value style.

---

## Structured Logging

> **Normative**: Always use key-value pairs. Never interpolate values into the message string.

The message is a **static description** of what happened. Dynamic data goes in
key-value attributes:

```go
// Good: static message, structured fields
slog.Info("order placed", "order_id", orderID, "total", total)

// Bad: dynamic data baked into the message string
slog.Info(fmt.Sprintf("order %d placed for $%.2f", orderID, total))
```

### Key Naming

> **Advisory**: Use `snake_case` for log attribute keys.

Keys should be lowercase, underscore-separated, and consistent across the
codebase: `user_id`, `request_id`, `elapsed_ms`.

### Typed Attributes

For performance-critical paths, use typed constructors to avoid allocations:

```go
slog.LogAttrs(ctx, slog.LevelInfo, "request handled",
    slog.String("method", r.Method),
    slog.Int("status", code),
    slog.Duration("elapsed", elapsed),
)
```

---

## Log Levels

> **Advisory**: Follow these level semantics consistently.

| Level | When to use | Production default |
|-------|-------------|--------------------|
| Debug | Developer-only diagnostics, tracing internal state | Disabled |
| Info  | Notable lifecycle events: startup, shutdown, config loaded | Enabled |
| Warn  | Unexpected but recoverable: deprecated feature used, retry succeeded | Enabled |
| Error | Operation failed, requires operator attention | Enabled |

**Rules of thumb**:
- If nobody should act on it, it's not Error — use Warn or Info
- If it's only useful with a debugger attached, it's Debug
- `slog.Error` should always include an `"err"` attribute

```go
slog.Error("payment failed", "err", err, "order_id", id)
slog.Warn("retry succeeded", "attempt", n, "endpoint", url)
slog.Info("server started", "addr", addr)
slog.Debug("cache lookup", "key", key, "hit", hit)
```

---

## Request-Scoped Logging

> **Advisory**: Derive loggers from context to carry request-scoped fields.

Use middleware to enrich a logger with request ID, user ID, or trace ID, then
pass the enriched logger downstream via context or as an explicit parameter.
Keep the full context-key and middleware implementation in the logging patterns
reference so request-scoped logging has one owner.

---

## Log or Return, Not Both

The handle-once rule belongs to [go-error-handling](../go-error-handling/SKILL.md).
In logging work, apply it by choosing either a local log-and-recover path or a
return path with context, not both for the same error.

**Exception**: HTTP handlers and other top-of-stack boundaries may log detailed
errors server-side while returning a sanitized message to the client:

```go
if err != nil {
    slog.Error("checkout failed", "err", err, "user_id", uid)
    http.Error(w, "internal error", http.StatusInternalServerError)
    return
}
```

See [go-error-handling](../go-error-handling/SKILL.md) for the full
handle-once pattern and error wrapping guidance.

---

## What NOT to Log

> **Normative**: Never log secrets, credentials, PII, or high-cardinality unbounded data.

- Passwords, API keys, tokens, session IDs
- Full credit card numbers, SSNs
- Request/response bodies that may contain user data
- Entire slices or maps of unbounded size

---

## Quick Reference

| Do | Don't |
|----|-------|
| `slog.Info("msg", "key", val)` | `log.Printf("msg %v", val)` |
| Static message + structured fields | `fmt.Sprintf` in message |
| `snake_case` keys | camelCase or inconsistent keys |
| Log OR return errors | Log AND return the same error |
| Derive logger from context | Create a new logger per call |
| Use `slog.Error` with `"err"` attr | `slog.Info` for errors |
| Pre-check `Enabled()` on hot paths | Always allocate log args |

---

## Related Skills

- **Error handling**: See [go-error-handling](../go-error-handling/SKILL.md) when deciding whether to log or return an error, or for the handle-once pattern
- **Context propagation**: See [go-context](../go-context/SKILL.md) when passing request-scoped values (including loggers) through context
- **Performance**: See [go-performance](../go-performance/SKILL.md) when optimizing hot-path logging or reducing allocations in log calls
- **Code review**: See [go-code-review](../go-code-review/SKILL.md) when reviewing logging practices in Go PRs
