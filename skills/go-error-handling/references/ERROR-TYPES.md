# Error Types Reference

This reference covers structured error types, sentinel errors, and how to choose
the right error type for your use case.

---

## Error Structure

> The error-type decision table is in the parent skill (SKILL.md § Error Types).
> This reference covers: expanded code examples, sentinel errors, error checking
> with `errors.Is`/`errors.As`, and structured error types.

**Key considerations**:

- Does the caller need to match the error with `errors.Is` or `errors.As`?
- Is the error message static or does it require runtime values?
- Exported error variables/types become part of your public API

```go
// No matching needed, static message
func Open() error {
    return errors.New("could not open")
}

// Matching needed, static message - export a sentinel
var ErrCouldNotOpen = errors.New("could not open")

func Open() error {
    return ErrCouldNotOpen
}

// Matching needed, dynamic message - use custom type
type NotFoundError struct {
    File string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("file %q not found", e.File)
}

func Open(file string) error {
    return &NotFoundError{File: file}
}
```

---

## Sentinel Errors

The simplest structured errors are unparameterized global values:

```go
// Good: Sentinel errors for programmatic checking
var (
    // ErrDuplicate occurs if this animal has already been seen.
    ErrDuplicate = errors.New("duplicate")

    // ErrMarsupial occurs because we're allergic to marsupials.
    ErrMarsupial = errors.New("marsupials are not supported")
)

func process(animal Animal) error {
    switch {
    case seen[animal]:
        return ErrDuplicate
    case marsupial(animal):
        return ErrMarsupial
    }
    seen[animal] = true
    return nil
}
```

---

## Checking Errors

For direct comparison (when errors are not wrapped):

```go
// Good: Direct comparison with sentinel
switch err := process(an); err {
case ErrDuplicate:
    return fmt.Errorf("feed %q: %v", an, err)
case ErrMarsupial:
    alternate := an.BackupAnimal()
    return handlePet(alternate)
}
```

When errors may be wrapped, use `errors.Is`:

```go
// Good: Works with wrapped errors
switch err := process(an); {
case errors.Is(err, ErrDuplicate):
    return fmt.Errorf("feed %q: %v", an, err)
case errors.Is(err, ErrMarsupial):
    // Try to recover...
}
```

**Never** match errors based on string content:

```go
// Bad: Fragile string matching
if regexp.MatchString(`duplicate`, err.Error()) {...}
if regexp.MatchString(`marsupial`, err.Error()) {...}
```

---

## Structured Error Types

For errors needing additional programmatic information, use struct types:

```go
// Good: Structured error with accessible fields
type PathError struct {
    Op   string
    Path string
    Err  error
}

func (e *PathError) Error() string {
    return e.Op + " " + e.Path + ": " + e.Err.Error()
}

func (e *PathError) Unwrap() error { return e.Err }
```

Callers can use `errors.As` to extract the structured error:

```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("Failed path:", pathErr.Path)
}
```

---

## Quick Reference

| Scenario | Error Type |
|----------|------------|
| No matching needed, static message | `errors.New("message")` |
| No matching needed, dynamic message | `fmt.Errorf("msg: %v", val)` |
| Matching needed, static message | `var ErrFoo = errors.New(...)` |
| Matching needed, dynamic message | custom struct type |
| Checking sentinel errors | `errors.Is(err, ErrFoo)` |
| Extracting structured errors | `errors.As(err, &target)` |
