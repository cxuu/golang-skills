# Struct Initialization

> **Source**: Uber Style Guide, Google Style Guide

Detailed rules and patterns for initializing Go structs.

---

## Always Use Field Names

Specify field names when initializing structs. This is enforced by `go vet` for
external package types:

```go
// Bad: positional — fragile, breaks when fields are added/reordered
k := User{"John", "Doe", true}

// Good: named fields — clear and resilient to changes
k := User{
    FirstName: "John",
    LastName:  "Doe",
    Admin:     true,
}
```

**Exception**: Field names may be omitted in test tables with 3 or fewer fields:

```go
tests := []struct {
    input    string
    expected int
}{
    {"abc", 3},
    {"", 0},
}
```

---

## Omit Zero-Value Fields

Let Go set zero values automatically. Only include fields that provide
meaningful context:

```go
// Bad: noise from zero fields
user := User{
    FirstName:  "John",
    LastName:   "Doe",
    MiddleName: "",
    Admin:      false,
}

// Good: zero fields omitted, important ones stand out
user := User{
    FirstName: "John",
    LastName:  "Doe",
}
```

**Exception**: Table-driven test structs often benefit from explicit field names
even for zero values to clarify the test case.

---

## Use `var` for Zero-Value Structs

Signal that a zero-value struct is intentional with `var`:

```go
// Bad: empty literal is ambiguous — forgot fields or intentional?
user := User{}

// Good: var clearly signals "zero value on purpose"
var user User
```

---

## Use `&T{}` for Struct References

Prefer `&T{}` over `new(T)` for consistency with struct initialization:

```go
// Bad: new() then set fields separately
sptr := new(T)
sptr.Name = "bar"

// Good: initialize inline
sptr := &T{Name: "bar"}
```

Both `&T{}` and `new(T)` produce a pointer to a zero-value `T`, but `&T{}`
allows inline field initialization.

---

## Multi-Line vs Single-Line

Use single-line for structs with 1-2 short fields:

```go
p := Point{X: 1, Y: 2}
```

Use multi-line for 3+ fields or long values:

```go
cfg := Config{
    Host:    "localhost",
    Port:    8080,
    Timeout: 30 * time.Second,
}
```

---

## Pointer to Struct Literals in Slices

When building slices of struct pointers, omit the repeated type name:

```go
// Good
items := []*Item{
    {Name: "a", Value: 1},
    {Name: "b", Value: 2},
}

// Bad: redundant type names
items := []*Item{
    &Item{Name: "a", Value: 1},
    &Item{Name: "b", Value: 2},
}
```

Run `gofmt -s` to clean these up automatically.
