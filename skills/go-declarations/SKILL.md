---
name: go-declarations
description: Use when declaring or initializing Go variables, constants, structs, or maps — including choosing between var and :=, reducing variable scope with if-init, formatting composite literals, designing iota enums, and using modern idioms like any instead of interface{}. Also use when a user is writing a new struct or const block, even if they don't ask about declaration style. For allocation (new vs make) and slice/map operations, see go-data-structures.
sources: [Google Style Guide, Uber Style Guide]
---

# Go Declarations and Initialization

---

## Group Similar Declarations

Group related `var`, `const`, and `type` declarations in parenthesized blocks.

**Bad:**

```go
import "a"
import "b"

const a = 1
const b = 2

var a = 1
var b = 2

type Area float64
type Volume float64
```

**Good:**

```go
import (
    "a"
    "b"
)

const (
    a = 1
    b = 2
)

var (
    a = 1
    b = 2
)

type (
    Area   float64
    Volume float64
)
```

Only group **related** declarations. Separate unrelated ones into distinct
blocks:

```go
// Good: related constants grouped, unrelated constant separate
type Operation int

const (
    Add Operation = iota + 1
    Subtract
    Multiply
)

const EnvVar = "MY_ENV"
```

Groups work inside functions too. Group adjacent variable declarations even if
unrelated:

```go
func (c *client) request() {
    var (
        caller  = c.name
        format  = "json"
        timeout = 5 * time.Second
        err     error
    )
    // ...
}
```

---

## Constants and iota

`iota` creates enumerated constants. Start enums at one so the zero value
represents an invalid/unset state:

```go
type Operation int

const (
    Add Operation = iota + 1
    Subtract
    Multiply
)
// Add=1, Subtract=2, Multiply=3
```

There are cases where the zero value makes sense as a default — use zero when
the default behavior is desirable:

```go
type LogOutput int

const (
    LogToStdout LogOutput = iota  // zero value = default
    LogToFile
    LogToRemote
)
```

Advanced `iota` patterns (e.g., bit-shifting for `ByteSize`):

```go
type ByteSize float64

const (
    _           = iota // ignore first value (0)
    KB ByteSize = 1 << (10 * iota)
    MB
    GB
    TB
)
```

---

## Top-level Variable Declarations

At the top level, use `var`. Do not specify the type unless it differs from the
expression's type:

**Bad:**

```go
var _s string = F()

func F() string { return "A" }
```

**Good:**

```go
var _s = F()

func F() string { return "A" }
```

Specify the type when the desired type differs from the expression:

```go
type myError struct{}

func (myError) Error() string { return "error" }
func F() myError              { return myError{} }

var _e error = F()
// F returns myError but we want the error interface.
```

---

## Local Variable Declarations

> **Default**: Use `:=` for local variables. Use `var` only for: zero-value
> initialization where the zero value matters, or when the type isn't clear from
> the right-hand side.

Use `:=` when explicitly assigning a value:

```go
// Bad
var s = "foo"

// Good
s := "foo"
```

Use `var` when the zero value is intentional—it signals "this starts empty on
purpose":

```go
// Bad: empty literal hides the intent
func f(list []int) {
    filtered := []int{}
    for _, v := range list {
        if v > 10 {
            filtered = append(filtered, v)
        }
    }
}

// Good: var signals intentional nil slice
func f(list []int) {
    var filtered []int
    for _, v := range list {
        if v > 10 {
            filtered = append(filtered, v)
        }
    }
}
```

---

## Reduce Scope of Variables

Move declarations as close to usage as possible. Use if-init to limit scope:

**Bad:**

```go
err := os.WriteFile(name, data, 0644)
if err != nil {
    return err
}
```

**Good:**

```go
if err := os.WriteFile(name, data, 0644); err != nil {
    return err
}
```

Don't reduce scope if it forces deeper nesting. When you need a result outside
the `if`, declare it before:

```go
// Good: data used after the error check
data, err := os.ReadFile(name)
if err != nil {
    return err
}

if err := cfg.Decode(data); err != nil {
    return err
}

fmt.Println(cfg)
```

Move constants into functions when only used there:

```go
// Good: constants scoped to the function that uses them
func Bar() {
    const (
        defaultPort = 8080
        defaultUser = "user"
    )
    fmt.Println("Default port", defaultPort)
}
```

---

## Initializing Structs

### Always Use Field Names

Specify field names when initializing structs. Enforced by `go vet`:

```go
// Bad
k := User{"John", "Doe", true}

// Good
k := User{
    FirstName: "John",
    LastName:  "Doe",
    Admin:     true,
}
```

Exception: field names may be omitted in test tables with 3 or fewer fields.

### Omit Zero-Value Fields

Let Go set zero values automatically. Only include fields that provide
meaningful context:

```go
// Bad
user := User{
    FirstName:  "John",
    LastName:   "Doe",
    MiddleName: "",
    Admin:      false,
}

// Good
user := User{
    FirstName: "John",
    LastName:  "Doe",
}
```

### Use `var` for Zero-Value Structs

```go
// Bad
user := User{}

// Good
var user User
```

### Use `&T{}` for Struct References

Prefer `&T{}` over `new(T)` for consistency with struct initialization:

```go
// Bad
sval := T{Name: "foo"}
sptr := new(T)
sptr.Name = "bar"

// Good
sval := T{Name: "foo"}
sptr := &T{Name: "bar"}
```

---

## Composite Literal Formatting

Use field names for external package types. Match closing brace indentation with
the opening line. Omit repeated type names in slice/map literals (`gofmt -s`).

> Read [references/INITIALIZATION.md](references/INITIALIZATION.md) when working with complex struct initialization, factory patterns, or builder patterns.

---

## Initializing Maps

Use `make()` for empty maps that will be populated programmatically—it visually
distinguishes initialization from a nil declaration and allows size hints:

```go
// Bad: empty literal looks too similar to nil declaration
var (
    m1 = map[T1]T2{}
    m2 map[T1]T2
)

// Good: make() is visually distinct
var (
    m1 = make(map[T1]T2)
    m2 map[T1]T2
)
```

Use map literals for fixed entries:

```go
// Bad: programmatic insertion of static data
m := make(map[string]int, 3)
m["one"] = 1
m["two"] = 2
m["three"] = 3

// Good: literal for fixed entries
m := map[string]int{
    "one":   1,
    "two":   2,
    "three": 3,
}
```

Rule of thumb: **literals** for fixed data at init time, **`make`** for maps
populated later (with a size hint if known).

---

## Use Raw String Literals

Use backtick strings to avoid hand-escaped characters:

```go
// Bad
wantError := "unknown name:\"test\""

// Good
wantError := `unknown name:"test"`
```

Raw string literals can span multiple lines and include quotes, making them
ideal for regex patterns, SQL, JSON, and multi-line text.

---

## Use `any` Instead of `interface{}`

Go 1.18 introduced `any` as an alias for `interface{}`. Prefer `any` in new
code:

```go
// Bad
func process(v interface{}) {}

// Good
func process(v any) {}
```

---

## Avoid Using Built-In Names

Never use Go's [predeclared identifiers](https://go.dev/ref/spec#Predeclared_identifiers)
as variable, function, or type names. Shadowing built-ins creates subtle bugs
that the compiler won't catch.

**Predeclared identifiers to avoid**: `error`, `string`, `bool`, `int`,
`float64`, `len`, `cap`, `append`, `copy`, `new`, `make`, `close`, `delete`,
`panic`, `recover`, `any`, `true`, `false`, `nil`, `iota`.

**Bad:**

```go
var error string
// error now shadows the builtin

func handleErrorMessage(error string) {
    // error shadows the builtin inside this function
}

type Foo struct {
    error  error
    string string  // grepping for "error" or "string" becomes ambiguous
}
```

**Good:**

```go
var errorMessage string

func handleErrorMessage(msg string) {
    // error still refers to the builtin
}

type Foo struct {
    err error
    str string
}
```

Tools such as `go vet` can detect shadowing of predeclared identifiers.

---

## Quick Reference

| Topic | Rule | Source |
|-------|------|--------|
| Grouping | Group related `var`/`const`/`type`; separate unrelated | Uber |
| `iota` enums | Start at one unless zero value is a meaningful default | Uber |
| Top-level vars | Use `var`; omit type unless it differs | Uber |
| Local vars | `:=` for explicit values; `var` for intentional zero | Uber |
| Variable scope | Move close to usage; use if-init | Uber |
| Struct init | Field names always; omit zero fields; `var` for zero struct | Uber |
| Map init | `make()` for dynamic; literal for fixed | Uber |
| Raw strings | Backticks for escapes, regex, multi-line | Uber |
| `any` vs `interface{}` | Prefer `any` in new code | Google |
| Built-in names | Never shadow predeclared identifiers | Uber |

---

## See Also

- **go-style-core**: Foundational style principles
- **go-naming**: Naming conventions including variable name length
- **go-data-structures**: Allocation with `new` vs `make`, slices, maps
- **go-control-flow**: If-init patterns, `:=` redeclaration rules
- **go-performance**: Container capacity hints for maps and slices
