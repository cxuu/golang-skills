# Type Constraints in Go Generics

> **Sources**: Google Go Style Guide, Go language specification

Constraints define what operations a type parameter supports. Choose the
narrowest constraint that satisfies your function's needs — no more.

---

## Built-in Constraints

> **Normative**: Use standard constraints before writing your own.

| Constraint | Meaning |
|------------|---------|
| `any` | Alias for `interface{}`; no requirements on the type |
| `comparable` | Supports `==` and `!=`; required for map keys |
| `cmp.Ordered` | Supports `<`, `<=`, `>=`, `>` (Go 1.21+, replaces `constraints.Ordered`) |

Prefer `cmp.Ordered` (from `cmp` package) over the deprecated
`golang.org/x/exp/constraints.Ordered` for new code.

---

## The `~` Operator (Underlying Types)

> **Advisory**: Use `~` when you want to accept named types built on a
> primitive.

The `~T` syntax matches any type whose **underlying type** is `T`. Without `~`,
only the exact type matches.

```go
type Celsius float64

type ExactFloat interface{ float64 }   // rejects Celsius
type AnyFloat64 interface{ ~float64 }  // accepts Celsius
```

Use `~` when callers are likely to define named types over the base type.
Omit `~` only when you need to restrict to the exact built-in type.

---

## Composing and Writing Constraints

> **Advisory**: Define a custom constraint only when no standard one fits.

Combine types with `|` and embed constraints to compose them:

```go
type Numeric interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~float32 | ~float64
}

type Addable interface {
    Numeric | ~string  // numbers and string concatenation
}
```

Constraints can require methods alongside type elements:

```go
type Stringer interface {
    comparable
    String() string
}
```

A type satisfying `Stringer` must be comparable **and** have a `String()` method.

---

## Avoiding Over-Constraining

> **Normative**: Use the minimal constraint that supports the operations
> you perform.

**Bad**
```go
// Only uses == but restricts to int and string
func Contains[T interface{ ~int | ~string }](s []T, v T) bool { ... }
```

**Good**
```go
// comparable is the minimal constraint for ==
func Contains[T comparable](s []T, v T) bool { ... }
```

Over-constraining limits reuse and forces callers to work around restrictions
that the implementation never needed.

## Type Inference

> **Advisory**: Let the compiler infer type arguments when unambiguous.

The compiler infers type parameters from function arguments:

```go
result := slices.Contains[string](names, "alice")  // explicit — unnecessary
result := slices.Contains(names, "alice")           // inferred — preferred
```

Supply type arguments explicitly only when there are no function arguments to
infer from, the inferred type is wrong (e.g., untyped constant promotes to the
wrong type), or readability benefits from making the type visible.

---

## Common Pitfalls

### Don't Use Generics When Interfaces Suffice

> **Normative**: From Google Style Guide — prefer interfaces when types
> share a useful unifying interface.

**Bad**
```go
// T is only used to satisfy io.Reader — just use the interface
func Process[T io.Reader](r T) error { ... }
```

**Good**
```go
func Process(r io.Reader) error { ... }
```

If the constraint is a single existing interface, accept the interface directly.

### Don't Wrap Standard Library Types Generically

> **Advisory**: A single-use generic is just indirection.

**Bad**
```go
type Set[T comparable] struct{ m map[T]struct{} }  // only ever Set[string]
```

**Good**
```go
seen := map[string]struct{}{}  // use map directly for a single instantiation
```

Generics justify complexity when they eliminate duplication across **multiple
call sites**. If only one type is ever used, start concrete.

### Method Sets and Type Constraints

You can only call operations the constraint allows:

**Bad**
```go
func Stringify[T any](v T) string {
    return v.String()  // compile error: any does not have String()
}
```

**Good**
```go
func Stringify[T fmt.Stringer](v T) string {
    return v.String()
}
```

---

## Quick Reference

| Topic | Guidance |
|-------|----------|
| Default constraint | `any` — use when no operations on T are needed |
| Equality checks | `comparable` — required for `==`, `!=`, and map keys |
| Ordering | `cmp.Ordered` (Go 1.21+) for `<`, `>` comparisons |
| Named types | Use `~T` to accept types whose underlying type is T |
| Union types | Combine with `\|` — e.g., `~int \| ~float64` |
| Custom constraints | Define as interface with type elements and/or methods |
| Type inference | Omit type args when the compiler can infer them |
| Minimal constraint | Use the narrowest constraint the function actually needs |
