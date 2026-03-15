# Must Functions

> **Source**: Uber Style Guide, Go standard library conventions

`Must` functions wrap a fallible function and panic on error. Use them **only**
during program initialization where failure means the program cannot run.

## Standard Library Examples

```go
// regexp.MustCompile panics if the pattern is invalid
var validID = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

// template.Must panics if template parsing fails
var tmpl = template.Must(template.ParseFiles("index.html"))
```

These are safe because they run at package init time — if they fail, the
program cannot function correctly.

## When to Use Must

```
Is this called during program initialization (package-level var, init, main setup)?
├─ Yes → Is failure unrecoverable (config, regex, template)?
│        ├─ Yes → Must is appropriate
│        └─ No  → Return error instead
└─ No  → Never use Must — return error
```

### Appropriate Uses

- **Package-level `var`**: Compiling regexps, parsing templates, loading
  required config
- **`init()` or early `main()`**: Setting up resources that must exist for
  the program to run
- **Test helpers**: `t.Fatal` is preferred in tests, but Must can be
  acceptable for test fixtures

### Never Use Must For

- Runtime request handling
- User-supplied input
- Network or file operations that can legitimately fail
- Anything called after program startup

## Writing a Must Function

Follow the naming convention `MustX` where `X` is the fallible function name:

```go
func MustParseConfig(path string) *Config {
    cfg, err := ParseConfig(path)
    if err != nil {
        panic(fmt.Sprintf("parsing config %s: %v", path, err))
    }
    return cfg
}
```

### Guidelines

- **Name**: `Must` prefix + the fallible function name (e.g., `MustParse`,
  `MustNew`, `MustCompile`)
- **Panic message**: Include the input and the error for debuggability
- **Document**: Always document that the function panics on error

```go
// MustParseConfig parses the config file at path.
// It panics if the file cannot be read or contains invalid configuration.
func MustParseConfig(path string) *Config { ... }
```

### Generic Must Helper

For one-off uses, a generic Must helper avoids boilerplate:

```go
func Must[T any](v T, err error) T {
    if err != nil {
        panic(err)
    }
    return v
}

// Usage at package level
var cfg = Must(ParseConfig("app.yaml"))
```

## Relationship to Panic/Recover

Must functions are a controlled use of `panic`. They should:

- Only run during initialization (so recover is unnecessary)
- Produce clear, actionable panic messages
- Never be used where returning an error is possible

See [PANIC-RECOVER.md](PANIC-RECOVER.md) for the full panic/recover pattern.
