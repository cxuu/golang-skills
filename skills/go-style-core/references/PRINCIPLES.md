# Style Principles Reference

## 1. Clarity

The code's purpose and rationale must be clear to the reader.

- **What**: Use descriptive names, helpful comments, and efficient organization
- **Why**: Add commentary explaining rationale, especially for nuances
- View clarity through the reader's lens, not the author's
- Code should be easy to read, not easy to write

```go
// Good: Clear purpose
func (c *Config) WriteTo(w io.Writer) (int64, error)

// Bad: Unclear, repeats receiver
func (c *Config) WriteConfigTo(w io.Writer) (int64, error)
```

## 2. Simplicity

Code should accomplish goals in the simplest way possible.

Simple code:
- Is easy to read top to bottom
- Does not assume prior knowledge
- Has no unnecessary abstraction levels
- Has comments explaining "why", not "what"
- May be mutually exclusive with "clever" code

### Least Mechanism

Where there are several ways to express the same idea, prefer the most standard
tool:

1. Core language constructs (channel, slice, map, loop, struct)
2. Standard library (HTTP client, template engine)
3. Third-party library — only when (1) and (2) don't suffice

## 3. Concision

Code should have high signal-to-noise ratio.

- Avoid repetitive code
- Avoid extraneous syntax
- Avoid unnecessary abstraction
- Use table-driven tests to factor out common code

```go
// Good: Common idiom, high signal
if err := doSomething(); err != nil {
    return err
}

// Good: Signal boost for unusual case
if err := doSomething(); err == nil { // if NO error
    // ...
}
```

## 4. Maintainability

Code is edited many more times than written.

Maintainable code:
- Is easy for future programmers to modify correctly
- Has APIs that grow gracefully
- Uses predictable names (same concept = same name)
- Minimizes dependencies
- Has comprehensive tests with clear diagnostics

```go
// Bad: Critical detail hidden
if user, err = db.UserByID(userID); err != nil { // = vs :=

// Good: Explicit and clear
u, err := db.UserByID(userID)
if err != nil {
    return fmt.Errorf("invalid origin user: %s", err)
}
user = u
```

## 5. Consistency

Code should look and behave like similar code in the codebase.

- Package-level consistency is most important
- When ties occur, break in favor of consistency
- Never override documented style principles for consistency
