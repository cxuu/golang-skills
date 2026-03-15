# Logging Patterns

Detailed patterns for slog setup, handler configuration, testing, HTTP
middleware, and migration from the legacy `log` package.

## Setting Up slog

### Basic Configuration

```go
package main

import (
    "log/slog"
    "os"
)

func main() {
    // JSON handler for production (machine-parseable)
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    slog.Info("server started", "addr", ":8080")
    // Output: {"time":"...","level":"INFO","msg":"server started","addr":":8080"}
}
```

### Text Handler for Development

```go
// Human-readable output for local development
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
slog.SetDefault(logger)
// Output: time=... level=DEBUG msg="cache lookup" key=user:42 hit=true
```

### Dynamic Level Control

Use `slog.LevelVar` to change the minimum level at runtime (e.g., via an
admin endpoint or signal handler):

```go
var programLevel = new(slog.LevelVar) // default Info

func init() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: programLevel,
    }))
    slog.SetDefault(logger)
}

// Call from an admin endpoint or signal handler
func enableDebug() {
    programLevel.Set(slog.LevelDebug)
}
```

---

## Custom Handler Patterns

### Adding Source Location

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    AddSource: true,
    Level:     slog.LevelInfo,
}))
// Output includes: "source":{"function":"main.handleRequest","file":"server.go","line":42}
```

### Wrapping Handlers with Default Attributes

Use `slog.Handler` middleware to inject fields into every log record:

```go
type contextHandler struct {
    inner   slog.Handler
    attrs   []slog.Attr
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.inner.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
    r.AddAttrs(h.attrs...)
    return h.inner.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return &contextHandler{inner: h.inner.WithAttrs(attrs), attrs: h.attrs}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
    return &contextHandler{inner: h.inner.WithGroup(name), attrs: h.attrs}
}
```

### Multi-Handler (Fan-Out)

Write to multiple destinations (e.g., stdout + file):

```go
type multiHandler struct {
    handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
    for _, h := range m.handlers {
        if h.Enabled(ctx, level) {
            return true
        }
    }
    return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
    var errs []error
    for _, h := range m.handlers {
        if h.Enabled(ctx, r.Level) {
            if err := h.Handle(ctx, r); err != nil {
                errs = append(errs, err)
            }
        }
    }
    return errors.Join(errs...)
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    handlers := make([]slog.Handler, len(m.handlers))
    for i, h := range m.handlers {
        handlers[i] = h.WithAttrs(attrs)
    }
    return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
    handlers := make([]slog.Handler, len(m.handlers))
    for i, h := range m.handlers {
        handlers[i] = h.WithGroup(name)
    }
    return &multiHandler{handlers: handlers}
}
```

---

## Testing with slogtest

Go 1.22+ provides `testing/slogtest` to verify handler implementations:

```go
package myhandler_test

import (
    "testing"
    "testing/slogtest"
)

func TestHandler(t *testing.T) {
    // newHandler returns your custom slog.Handler and a func that
    // parses the output into []map[string]any for verification.
    results := func(t *testing.T) map[string]any {
        // parse your handler's output here
    }

    h := NewMyHandler(buf, nil)
    slogtest.Run(t, func(t *testing.T) slog.Handler { return h }, results)
}
```

### Capturing Logs in Tests

For unit tests that assert on log output, write to a buffer:

```go
func TestOrderProcessing(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewJSONHandler(&buf, nil))

    processOrder(logger, order)

    if !strings.Contains(buf.String(), `"order_id"`) {
        t.Error("expected order_id in log output")
    }
}
```

---

## HTTP Request Logging Middleware

A complete middleware that logs each request with timing, status, and
request-scoped fields:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        reqID := r.Header.Get("X-Request-ID")
        if reqID == "" {
            reqID = uuid.NewString()
        }

        logger := slog.With(
            "request_id", reqID,
            "method", r.Method,
            "path", r.URL.Path,
        )

        // Wrap the response writer to capture the status code
        rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        // Store logger in context for downstream handlers
        ctx := context.WithValue(r.Context(), loggerKey, logger)
        next.ServeHTTP(rw, r.WithContext(ctx))

        logger.Info("request completed",
            "status", rw.status,
            "elapsed_ms", time.Since(start).Milliseconds(),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}
```

### Retrieving the Logger from Context

```go
type ctxKey struct{}

var loggerKey = ctxKey{}

func loggerFromCtx(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
```

---

## Migration from log.Printf to slog

### Step 1: Replace Direct Calls

```go
// Before
log.Printf("user %s logged in from %s", userID, ip)

// After
slog.Info("user logged in", "user_id", userID, "ip", ip)
```

### Step 2: Replace log.Fatalf in main()

```go
// Before
log.Fatalf("failed to connect: %v", err)

// After — slog has no Fatal; use slog + os.Exit in main
slog.Error("failed to connect", "err", err)
os.Exit(1)
```

### Step 3: Bridge Legacy Code

If migrating incrementally, redirect the standard `log` package output
through slog:

```go
// In main(), after setting up slog:
slog.SetDefault(logger)

// The standard log package now writes through slog's default handler.
// This works because slog.SetDefault also updates log.Default().
```

### Step 4: Replace Logger Parameters

```go
// Before: passing *log.Logger around
func NewServer(addr string, logger *log.Logger) *Server

// After: pass *slog.Logger explicitly
func NewServer(addr string, logger *slog.Logger) *Server

// Or derive from context in handlers
func (s *Server) handleRequest(ctx context.Context) {
    logger := loggerFromCtx(ctx)
    logger.Info("handling request")
}
```

### Migration Checklist

| Step | What to change | Verify |
|------|---------------|--------|
| 1 | `log.Printf` → `slog.Info/Warn/Error` | `rg 'log\.Printf'` returns 0 hits |
| 2 | `log.Fatalf` → `slog.Error` + `os.Exit(1)` in main | Only in `main()` |
| 3 | Set `slog.SetDefault` early in main | Legacy `log` calls route through slog |
| 4 | `*log.Logger` params → `*slog.Logger` | All constructors updated |
| 5 | Remove `"log"` imports where replaced | `goimports` handles this |
