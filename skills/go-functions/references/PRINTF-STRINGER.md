# Printf, Stringer, and Custom Formatting

Deep reference for Go's `fmt` printing verbs, the `Stringer` and `GoStringer`
interfaces, custom `Format()` methods, and common pitfalls.

---

## Printf Verbs

### General Verbs

| Verb | Use |
|------|-----|
| `%v` | Default format (struct fields, slice elements) |
| `%+v` | Struct fields with names: `{Name:alice Age:30}` |
| `%#v` | Go-syntax representation: `main.User{Name:"alice", Age:30}` |
| `%T` | Type of the value: `main.User` |
| `%%` | Literal percent sign |

### String and Byte Verbs

| Verb | Use |
|------|-----|
| `%s` | Plain string or byte slice |
| `%q` | Quoted string with Go syntax escaping: `"hello\n"` |
| `%x` | Hex encoding, lowercase: `68656c6c6f` |
| `%X` | Hex encoding, uppercase: `68656C6C6F` |

### Integer Verbs

| Verb | Use |
|------|-----|
| `%d` | Decimal integer |
| `%b` | Binary |
| `%o` | Octal |
| `%O` | Octal with `0o` prefix |
| `%x` | Hex, lowercase |
| `%X` | Hex, uppercase |

### Float Verbs

| Verb | Use |
|------|-----|
| `%f` | Decimal point, no exponent: `123.456` |
| `%e` | Scientific notation: `1.23456e+02` |
| `%g` | Compact: `%e` for large exponents, `%f` otherwise |

### Width and Precision

```go
fmt.Sprintf("%10d", 42)     // "        42"  (width 10, right-aligned)
fmt.Sprintf("%-10d", 42)    // "42        "  (width 10, left-aligned)
fmt.Sprintf("%.2f", 3.14159) // "3.14"       (2 decimal places)
fmt.Sprintf("%010d", 42)    // "0000000042"  (zero-padded)
```

---

## Use `%q` for Strings

The `%q` verb prints strings inside double quotes, making empty strings and
control characters visible:

```go
fmt.Printf("value %q looks like English text", someText)

// Bad: manually adding quotes
fmt.Printf("value \"%s\" looks like English text", someText)
```

Prefer `%q` in output intended for humans where the value could be empty or
contain control characters.

---

## Format Strings Outside Printf

When declaring format strings outside a `Printf`-style call, use `const`. This
enables `go vet` to perform static analysis:

```go
// Bad: variable format string — go vet can't check it
msg := "unexpected values %v, %v\n"
fmt.Printf(msg, 1, 2)

// Good: const format string — go vet can validate
const msg = "unexpected values %v, %v\n"
fmt.Printf(msg, 1, 2)
```

---

## Naming Printf-style Functions

Functions that accept a format string should end in `f`. This lets `go vet`
check format strings automatically:

```go
func Wrapf(err error, format string, args ...any) error
```

If using a non-standard name, tell `go vet`:

```bash
go vet -printfuncs=wrapf,statusf
```

---

## The `fmt.Stringer` Interface

Implement `fmt.Stringer` to control how your type appears with `%v` and `%s`:

```go
type fmt.Stringer interface {
    String() string
}
```

```go
type Point struct{ X, Y int }

func (p Point) String() string {
    return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

// fmt.Println(Point{1, 2})           → "(1, 2)"
// fmt.Sprintf("point: %v", p)        → "point: (1, 2)"
// fmt.Sprintf("point: %s", p)        → "point: (1, 2)"
```

### When to Implement Stringer

- Your type will appear in log messages or user-facing output
- The default `%v` output (field values only) isn't meaningful
- You need a human-friendly representation separate from serialization

---

## The `fmt.GoStringer` Interface

Implement `fmt.GoStringer` to control `%#v` output. This is useful for types
where the default Go-syntax representation is misleading or too verbose:

```go
type fmt.GoStringer interface {
    GoString() string
}
```

```go
type Color struct{ R, G, B uint8 }

func (c Color) GoString() string {
    return fmt.Sprintf("Color(%#02x, %#02x, %#02x)", c.R, c.G, c.B)
}

// fmt.Sprintf("%#v", Color{255, 128, 0})
// → "Color(0xff, 0x80, 0x00)"   instead of  "main.Color{R:0xff, G:0x80, B:0x00}"
```

`GoString()` output should be valid Go syntax or close to it — it's meant for
debugging, not user-facing display.

---

## Custom Formatting with `fmt.Formatter`

For full control over all format verbs, implement `fmt.Formatter`:

```go
type fmt.Formatter interface {
    Format(f fmt.State, verb rune)
}
```

```go
type Point struct{ X, Y int }

func (p Point) Format(f fmt.State, verb rune) {
    switch verb {
    case 'v':
        if f.Flag('#') {
            // %#v — Go-syntax representation
            fmt.Fprintf(f, "Point{X: %d, Y: %d}", p.X, p.Y)
            return
        }
        if f.Flag('+') {
            // %+v — verbose with field names
            fmt.Fprintf(f, "X:%d Y:%d", p.X, p.Y)
            return
        }
        // %v — default
        fmt.Fprintf(f, "(%d, %d)", p.X, p.Y)
    case 's':
        fmt.Fprintf(f, "(%d, %d)", p.X, p.Y)
    case 'q':
        fmt.Fprintf(f, "%q", p.String())
    default:
        fmt.Fprintf(f, "%%!%c(Point=%d,%d)", verb, p.X, p.Y)
    }
}
```

### `fmt.State` Methods

| Method | Returns |
|--------|---------|
| `Flag(c int) bool` | Whether flag (`+`, `-`, `#`, `0`, ` `) is set |
| `Width() (int, bool)` | Width value and whether it was specified |
| `Precision() (int, bool)` | Precision value and whether it was specified |
| `Write(b []byte) (int, error)` | Writes output bytes |

Only implement `fmt.Formatter` when `String()` isn't sufficient — it's rarely
needed. Common reasons: different output for `%v` vs `%+v` vs `%#v`, or
respecting width/precision flags.

---

## The Infinite Recursion Trap

**Calling `fmt.Sprintf` with `%s` or `%v` on the receiver inside `String()`
causes infinite recursion:**

```go
type MyString string

// BUG: infinite recursion — Sprintf calls String(), which calls Sprintf...
func (m MyString) String() string {
    return fmt.Sprintf("MyString: %s", m)  // CRASH: stack overflow
}
```

The fix — convert the receiver to its underlying type to break the method set:

```go
func (m MyString) String() string {
    return fmt.Sprintf("MyString: %s", string(m))  // Safe: string has no String()
}
```

This trap also applies to:
- Types whose underlying type is a string, []byte, or another Stringer
- Any `String()` method that formats `self` using `%s` or `%v`
- `GoString()` methods that format `self` using `%#v`

```go
type IPAddr [4]byte

// BUG: %v calls String(), infinite recursion
func (ip IPAddr) String() string {
    return fmt.Sprintf("%v.%v.%v.%v", ip[0], ip[1], ip[2], ip[3])
    // Safe here — ip[0] is a byte (uint8), which has no String() method.
    // But if ip were a named type wrapping a Stringer, this would recurse.
}
```

**Rule of thumb**: inside `String()`, never pass the receiver (or the receiver
directly re-typed as its own type) to a `%s` or `%v` verb. Convert to the
underlying primitive type first.

---

## Quick Reference

| Topic | Rule |
|-------|------|
| `%q` | Use for human-readable string output |
| `%+v` | Struct fields with names |
| `%#v` | Go-syntax representation; customize via `GoStringer` |
| Format string storage | Declare as `const` outside Printf calls |
| Printf function names | End with `f` for `go vet` support |
| `Stringer` | Implement `String() string` for `%v`/`%s` output |
| `GoStringer` | Implement `GoString() string` for `%#v` output |
| `Formatter` | Implement `Format(fmt.State, rune)` for full verb control |
| Recursion trap | Never `Sprintf("%s", receiver)` inside `String()`; convert to underlying type |
