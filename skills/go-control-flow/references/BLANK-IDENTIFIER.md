# Blank Identifier Patterns

The blank identifier `_` serves multiple roles in Go: discarding unwanted
values, importing packages for side effects, and verifying interface compliance
at compile time.

---

## Multiple Assignment

Use `_` to discard unwanted values from multi-value expressions:

```go
if _, err := os.Stat(path); os.IsNotExist(err) {
    fmt.Printf("%s does not exist\n", path)
}
```

### Never Discard Errors Carelessly

Silently discarding an error invites nil-pointer panics:

```go
// Bad: ignoring error will crash if path doesn't exist
fi, _ := os.Stat(path)
if fi.IsDir() { ... }  // nil pointer dereference
```

If you truly don't need the error, document why:

```go
_ = logger.Sync() // best-effort flush; error is non-actionable
```

---

## Import for Side Effect

Import a package solely for its `init()` side effects using the blank
identifier:

```go
import _ "net/http/pprof"  // registers HTTP handlers
import _ "image/png"       // registers PNG decoder
```

This is commonly used to register drivers, codecs, or debug handlers that
wire themselves into a registry during `init()`.

---

## Interface Compliance Check

Verify at compile time that a type implements an interface by assigning a
nil pointer to a blank-identifier variable of the interface type:

```go
var _ io.Writer = (*MyType)(nil)
```

This produces a compile error if `*MyType` does not satisfy `io.Writer`,
catching missing methods before runtime.

**When to use**: Place this check in the same file that defines the type,
typically right after the type declaration. It is especially useful when a
type must satisfy an interface defined in another package.

See [go-interfaces](../../go-interfaces/SKILL.md): Interface Satisfaction
Checks for full guidance on when and where to use this pattern.

---

## Quick Reference

| Pattern | Syntax |
|---------|--------|
| Discard value | `_, err := f()` |
| Discard in if-init | `if _, err := f(); err != nil { }` |
| Side-effect import | `import _ "pkg"` |
| Interface check | `var _ Interface = (*Type)(nil)` |
