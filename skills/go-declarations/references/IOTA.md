# Constants and iota Patterns

> **Source**: Uber Style Guide, Google Style Guide

Detailed patterns for designing enumerated constants with `iota` in Go.

---

## Start Enums at One

Start enums at one so the zero value represents an invalid/unset state. This
catches uninitialized variables:

```go
type Operation int

const (
    Add Operation = iota + 1
    Subtract
    Multiply
)
// Add=1, Subtract=2, Multiply=3
```

### When Zero Makes Sense

Use zero when the default behavior is desirable:

```go
type LogOutput int

const (
    LogToStdout LogOutput = iota  // zero value = default
    LogToFile
    LogToRemote
)
```

The key question: **is the zero value a valid, useful default?** If yes, start
at zero. If no, start at one.

---

## Bitmask Patterns

Use bit-shifting with `iota` for flag/bitmask enums:

```go
type Permission int

const (
    Read    Permission = 1 << iota  // 1
    Write                           // 2
    Execute                         // 4
)

// Combine with bitwise OR
perms := Read | Write  // 3
```

---

## Byte Size Pattern

A common pattern for byte size constants:

```go
type ByteSize float64

const (
    _           = iota // ignore first value (0)
    KB ByteSize = 1 << (10 * iota)
    MB
    GB
    TB
    PB
)
```

---

## String Representation

Always implement `String()` for enum types to aid debugging:

```go
func (o Operation) String() string {
    switch o {
    case Add:
        return "Add"
    case Subtract:
        return "Subtract"
    case Multiply:
        return "Multiply"
    default:
        return fmt.Sprintf("Operation(%d)", o)
    }
}
```

Consider using `go generate` with `stringer` for automatic string methods on
large enums.

---

## Grouping Rules

- Each enum type gets its own `const` block — `iota` resets to 0 in each block
- Unrelated constants go in separate blocks
- Document the enum type, not each individual constant (unless behavior is
  non-obvious)
