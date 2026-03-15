# Context Patterns

Common patterns for deriving, checking, and propagating `context.Context`.

---

## Context Immutability

Contexts are immutable. It's safe to pass the same `ctx` to multiple calls that
share the same deadline, cancellation signal, credentials, and parent trace:

```go
// Safe: same context to sequential calls
func ProcessBatch(ctx context.Context, items []Item) error {
    for _, item := range items {
        if err := process(ctx, item); err != nil {
            return err
        }
    }
    return nil
}

// Safe: same context to concurrent calls
func ProcessConcurrently(ctx context.Context, a, b *Data) error {
    g, ctx := errgroup.WithContext(ctx)
    g.Go(func() error { return processA(ctx, a) })
    g.Go(func() error { return processB(ctx, b) })
    return g.Wait()
}
```

---

## When to Use context.Background()

Use `context.Background()` only for functions that are **never request-specific**:

```go
func main() {
    ctx := context.Background()
    if err := run(ctx); err != nil {
        log.Fatal(err)
    }
}

func startBackgroundWorker() {
    ctx := context.Background()
    go worker(ctx)
}
```

**Default to passing a Context** even if you think you don't need to. Only use
`context.Background()` directly if you have a good reason why passing a context
would be a mistake:

```go
func LoadConfig(ctx context.Context) (*Config, error) {
    // Even if not using ctx now, accepting it allows future
    // additions without API changes
}
```

---

## Deriving Contexts

```go
// Add timeout — cancel fires after duration elapses
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

// Add cancellation — caller controls when to cancel
ctx, cancel := context.WithCancel(ctx)
defer cancel()

// Add deadline — cancel fires at a specific wall-clock time
ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Hour))
defer cancel()

// Add value (use sparingly — only for request-scoped data)
ctx = context.WithValue(ctx, requestIDKey, reqID)
```

**Always `defer cancel()`** immediately after creating a derived context. This
ensures resources are released even if the function returns early.

### Nested Derivation

Derived contexts form a tree. Cancelling a parent cancels all its children:

```go
func handleRequest(ctx context.Context) error {
    // Parent timeout for the whole request
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Tighter timeout for the database call
    dbCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
    defer dbCancel()

    data, err := queryDB(dbCtx)
    if err != nil {
        return err
    }

    // Remaining time from parent context applies here
    return sendResponse(ctx, data)
}
```

---

## Checking Cancellation

### In Long-Running Loops

```go
func LongRunningOperation(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Do work
        }
    }
}
```

### Before Expensive Operations

Check cancellation before starting work that can't be interrupted:

```go
func ProcessItems(ctx context.Context, items []Item) error {
    for _, item := range items {
        if ctx.Err() != nil {
            return ctx.Err()
        }
        if err := expensiveProcess(item); err != nil {
            return err
        }
    }
    return nil
}
```

### Distinguishing Cancellation Causes

```go
if err := ctx.Err(); err != nil {
    switch {
    case errors.Is(err, context.Canceled):
        // Caller explicitly cancelled (e.g., client disconnected)
    case errors.Is(err, context.DeadlineExceeded):
        // Timeout or deadline passed
    }
}
```

---

## Respecting Cancellation in HTTP Handlers

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    result, err := slowOperation(ctx)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            // Client disconnected — nothing to write
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(result)
}
```

The `r.Context()` is cancelled when:
- The client closes the connection
- The `http.Server`'s `ReadTimeout` or `WriteTimeout` fires
- The `ServeHTTP` method returns

---

## Context Value Best Practices

### Use Unexported Key Types

```go
type contextKey struct{}

var userIDKey contextKey

func WithUserID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(userIDKey).(string)
    return id, ok
}
```

Using an unexported struct type as the key prevents collisions with keys from
other packages — even if they use the same string or int value.

### Provide Accessor Functions

Always wrap `context.WithValue` and `ctx.Value` in typed helper functions (as
shown above) rather than exposing keys. This gives you type safety and a single
place to change the implementation.

---

## Quick Reference

| Pattern | Guidance |
|---------|----------|
| Parameter position | Always first: `func F(ctx context.Context, ...)` |
| Struct storage | Don't store in structs; pass to methods |
| Custom types | Don't create; use `context.Context` interface |
| Application data | Prefer parameters > receiver > globals > context values |
| Request-scoped data | Appropriate for context values |
| Sharing context | Safe — contexts are immutable |
| `context.Background()` | Only for non-request-specific code |
| Default | Pass context even if you think you don't need it |
| `defer cancel()` | Always defer immediately after `WithTimeout`/`WithCancel`/`WithDeadline` |
| Value keys | Use unexported struct types, provide accessor functions |
| Cancellation check | `ctx.Err()` before expensive ops; `select` on `ctx.Done()` in loops |
