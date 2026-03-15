# Levels and Context

Detailed guidance on log level semantics, context-based logging patterns,
performance considerations, and what to keep out of logs.

## Level Semantics

### Debug

Developer-only diagnostics. Disabled in production by default. Use for tracing
internal state that helps during development or troubleshooting:

```go
slog.Debug("cache lookup", "key", key, "hit", hit)
slog.Debug("parsed config", "fields", len(cfg.Fields))
slog.Debug("SQL query", "query", q, "args", args)
```

**When to use**: Internal state transitions, cache behavior, detailed
request/response data during development.

### Info

Notable events that confirm the system is working as expected. These should
be useful in production for understanding system behavior:

```go
slog.Info("server started", "addr", addr, "version", version)
slog.Info("config loaded", "path", cfgPath, "env", env)
slog.Info("migration completed", "version", v, "elapsed_ms", elapsed)
slog.Info("user registered", "user_id", uid)
```

**When to use**: Startup/shutdown, configuration changes, significant business
events, periodic health summaries.

### Warn

Something unexpected happened, but the system recovered or degraded
gracefully. An operator may want to investigate but no immediate action is
required:

```go
slog.Warn("retry succeeded", "attempt", n, "endpoint", url)
slog.Warn("deprecated endpoint called", "path", r.URL.Path, "user_id", uid)
slog.Warn("rate limit approaching", "current", rate, "limit", max)
slog.Warn("fallback to default config", "err", err)
```

**When to use**: Retries that eventually succeeded, deprecated code paths,
approaching resource limits, fallback behavior.

### Error

An operation failed and requires operator attention. The system could not
fulfill the request or complete the task:

```go
slog.Error("payment failed", "err", err, "order_id", id, "amount", amt)
slog.Error("database connection lost", "err", err, "host", dbHost)
slog.Error("message processing failed", "err", err, "msg_id", msgID)
```

**When to use**: Failed operations that affect users, lost connections,
data integrity issues, external service failures that weren't recovered.

**Always include the error**: `slog.Error` calls should always have an `"err"`
attribute with the actual error value.

### Choosing Between Warn and Error

```
Did the operation ultimately succeed?
├─ Yes (after retry/fallback) → Warn
└─ No (caller gets an error)  → Error
    ├─ Requires immediate attention → Error
    └─ Can wait for next review   → Warn
```

---

## Custom Verbosity Levels

slog levels are integers. Define custom sub-levels between the standard ones
for fine-grained control:

```go
const (
    LevelTrace = slog.Level(-8)  // below Debug
    LevelNotice = slog.Level(2)  // between Info and Warn
)

slog.Log(ctx, LevelTrace, "detailed trace", "span_id", spanID)
```

Use `HandlerOptions.Level` with a `slog.LevelVar` to control the minimum
level at runtime.

---

## Context-Based Logging

### Pattern 1: Logger in Context

Store an enriched `*slog.Logger` in the context. Each middleware layer adds
its own fields:

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := authenticate(r)
        logger := loggerFromCtx(r.Context()).With("user_id", userID)
        ctx := context.WithValue(r.Context(), loggerKey, logger)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Pros**: Simple, works with any handler chain.
**Cons**: Requires discipline to always use `loggerFromCtx`.

### Pattern 2: Explicit Logger Parameter

Pass `*slog.Logger` as a function parameter alongside context:

```go
func processOrder(ctx context.Context, logger *slog.Logger, order *Order) error {
    logger.Info("processing order", "order_id", order.ID)
    // ...
}
```

**Pros**: Explicit dependency, easier to test, no context key.
**Cons**: Extra parameter in every function signature.

### When to Use Each

| Situation | Recommendation |
|-----------|---------------|
| HTTP handlers / middleware chains | Logger in context |
| Library code with no HTTP dependency | Explicit parameter |
| Background workers / batch jobs | Explicit parameter |
| Deep call chains (5+ levels) | Logger in context |

---

## Performance Considerations

### Pre-Check with Enabled()

Avoid allocating log arguments when the level is disabled:

```go
// Expensive: args are always evaluated, even if Debug is disabled
slog.Debug("request details",
    "headers", fmt.Sprintf("%v", r.Header),
    "body", string(bodyBytes),
)

// Better: skip entirely when disabled
if slog.Default().Enabled(ctx, slog.LevelDebug) {
    slog.Debug("request details",
        "headers", fmt.Sprintf("%v", r.Header),
        "body", string(bodyBytes),
    )
}
```

This matters when argument construction is expensive (formatting, marshaling,
or reading data). For simple attributes (`slog.String`, `slog.Int`), the
overhead is negligible.

### Use LogAttrs on Hot Paths

`slog.LogAttrs` avoids the `[]any` allocation that the convenience methods
(`slog.Info`, etc.) incur:

```go
// Standard — allocates a []any for the key-value pairs
slog.Info("request handled", "method", r.Method, "status", code)

// Faster — typed attributes, no []any allocation
slog.LogAttrs(ctx, slog.LevelInfo, "request handled",
    slog.String("method", r.Method),
    slog.Int("status", code),
)
```

### Avoid Logging in Tight Loops

If a loop processes thousands of items, log a summary rather than each
iteration:

```go
// Bad: one log per item in a 10k-item batch
for _, item := range items {
    slog.Debug("processing item", "id", item.ID)
    process(item)
}

// Good: log summary
slog.Info("batch started", "count", len(items))
processed, failed := processBatch(items)
slog.Info("batch completed", "processed", processed, "failed", failed)
```

---

## What NOT to Log

### Secrets and Credentials

Never log:
- Passwords, API keys, tokens (OAuth, JWT, session)
- Private keys, certificates
- Database connection strings with credentials

```go
// Bad
slog.Info("connecting", "dsn", dsn) // may contain password

// Good
slog.Info("connecting", "host", dbHost, "database", dbName)
```

### Personally Identifiable Information (PII)

Avoid logging unless required for debugging and your retention policy
allows it:
- Email addresses, phone numbers
- Full names, physical addresses
- IP addresses (in some jurisdictions)
- Credit card numbers, SSNs

If you must log a user identifier, use an opaque ID rather than PII.

### High-Cardinality Unbounded Data

Don't log entire request bodies, full stack traces at Info level, or
unbounded collections:

```go
// Bad: unbounded data
slog.Info("received", "body", string(requestBody))
slog.Info("users loaded", "users", users) // could be 100k entries

// Good: bounded summary
slog.Info("received", "content_length", len(requestBody), "content_type", ct)
slog.Info("users loaded", "count", len(users))
```

### Decision Table

| Data type | Log it? | Alternative |
|-----------|---------|-------------|
| Request ID / trace ID | Yes | — |
| User ID (opaque) | Yes | — |
| HTTP method, path, status | Yes | — |
| Error messages | Yes | — |
| Passwords / tokens | **Never** | Log token prefix or "redacted" |
| Full request body | **No** | Log content length and type |
| PII (email, name) | **Avoid** | Log opaque user ID |
| Large collections | **No** | Log count or summary |
| Stack traces | Debug only | Use `slog.Debug` |
