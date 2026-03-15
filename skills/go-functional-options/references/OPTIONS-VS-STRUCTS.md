# Functional Options vs Config Structs

> **Source**: Google Style Guide, Uber Style Guide

Both functional options and config structs solve the same problem — optional
configuration for constructors — but they have different trade-offs. Choose
based on API audience, extensibility needs, and complexity budget.

## Decision Framework

```
Need optional configuration?
├─ Internal or test-only API?
│   └─ Config struct (simpler, less boilerplate)
├─ Public API with 3+ options?
│   └─ Functional options (extensible, backward-compatible)
├─ Options need validation or interdependencies?
│   └─ Functional options (validate in apply or constructor)
├─ All options usually specified together?
│   └─ Config struct (one literal, no With* ceremony)
└─ Options likely to grow over time?
    └─ Functional options (add With* without breaking callers)
```

## Config Struct Pattern

A config struct groups optional parameters into a single struct passed to the
constructor. Zero values serve as defaults, or provide a `DefaultConfig()`.

**Good**
```go
type Config struct {
    Timeout  time.Duration // zero = no timeout
    MaxRetry int           // zero = no retries
    Logger   *log.Logger   // nil = discard
}

func NewClient(addr string, cfg Config) *Client {
    if cfg.Logger == nil {
        cfg.Logger = log.New(io.Discard, "", 0)
    }
    return &Client{addr: addr, cfg: cfg}
}
```

```go
c := NewClient("localhost:8080", Config{
    Timeout:  5 * time.Second,
    MaxRetry: 3,
})
```

**Bad** — Relying on unexported config fields in a public API:
```go
type config struct {  // unexported: callers can't construct it
    timeout time.Duration
}

func NewClient(addr string, cfg config) *Client { ... }
```

### When Zero Values Don't Work

If zero is a valid non-default value (e.g., timeout of 0 means "no timeout"
but the desired default is 30s), use a pointer field or a sentinel value:

```go
type Config struct {
    Timeout *time.Duration // nil = use default (30s), zero = no timeout
}
```

## Comparison

| Aspect | Functional Options | Config Struct |
|--------|-------------------|---------------|
| **Boilerplate** | High (type + apply + With* per option) | Low (one struct) |
| **Extensibility** | Add `With*` — no breaking changes | Add field — no breaking changes |
| **Backward compat** | Excellent for public APIs | Good (new fields get zero values) |
| **Defaults** | Built into constructor | Zero values or `DefaultConfig()` |
| **Validation** | In `apply` or constructor loop | In constructor after struct received |
| **Discoverability** | `With*` functions appear in godoc | All fields visible in one struct |
| **Testability** | Compare options or test constructor output | Compare struct literals |
| **Caller experience** | Only specify what differs from defaults | Must construct struct literal |
| **Zero-value ambiguity** | None — unset options not applied | May need pointer fields |

## When to Prefer Config Structs

- **Internal APIs** — less ceremony, easier to read at call sites
- **Few options (1-3)** — functional options overhead not justified
- **All options typically set together** — no benefit to variadic style
- **No validation needed** — simple field assignment suffices
- **Options are data, not behavior** — struct fields map naturally

```go
srv := NewServer(Config{
    Port:    8080,
    TLSCert: "/path/to/cert.pem",
    TLSKey:  "/path/to/key.pem",
})
```

## When to Prefer Functional Options

- **Public/library APIs** — callers shouldn't track internal config evolution
- **3+ options** that are individually optional
- **Complex defaults** — default computation depends on other options
- **Validation per option** — reject bad values at apply time
- **Options may grow** — new `With*` functions are purely additive

```go
srv := NewServer(
    WithPort(8080),
    WithTLS("/path/to/cert.pem", "/path/to/key.pem"),
    WithLogger(logger),
)
```

## Hybrid Approach

For APIs that need both convenience and extensibility, accept a config struct
for common settings and functional options for advanced overrides:

```go
func NewServer(cfg Config, opts ...Option) *Server {
    s := &Server{cfg: cfg}
    for _, o := range opts {
        o.apply(&s.cfg)
    }
    return s
}
```

Use this sparingly — it adds complexity. Prefer one approach per API.
