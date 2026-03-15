---
name: go-generics
description: Use when deciding whether to use Go generics, writing generic functions or types, choosing constraints, or picking between type aliases and type definitions. Also use when a user is writing a utility function that could work with multiple types, or is unsure whether generics are appropriate for their use case — even if they don't mention generics explicitly. Helps decide when to prefer generics vs interfaces, how to name type parameters, and how to document generic APIs.
license: Apache-2.0
compatibility: Requires Go 1.18+ (generics were introduced in Go 1.18)
metadata:
  sources: "Google Style Guide"
---

# Go Generics and Type Parameters

---

## When to Use Generics

Start with concrete types. Generalize only when a second type appears.

### Prefer Generics When

- Multiple types share identical logic (sorting, filtering, map/reduce)
- You would otherwise rely on `any` and excessive type switching
- You are building a reusable data structure (concurrent-safe set, ordered map)

### Avoid Generics When

- Only one type is being instantiated in practice
- Interfaces already model the shared behavior cleanly
- The generic code is harder to read than the type-specific alternative

> "Write code, don't design types." — Robert Griesemer and Ian Lance Taylor

### Decision Flow

```
Do multiple types share identical logic?
├─ No  → Use concrete types
├─ Yes → Do they share a useful interface?
│        ├─ Yes → Use an interface
│        └─ No  → Use generics
```

**Bad:**

```go
// Premature generics: only ever called with int
func Sum[T constraints.Integer | constraints.Float](vals []T) T {
    var total T
    for _, v := range vals {
        total += v
    }
    return total
}
```

**Good:**

```go
func SumInts(vals []int) int {
    var total int
    for _, v := range vals {
        total += v
    }
    return total
}
```

---

## Type Parameter Naming

| Name | Typical Use |
|------|-------------|
| `T` | General type parameter |
| `K` | Map key type |
| `V` | Map value type |
| `E` | Element/item type |

For complex constraints, a short descriptive name is acceptable:

```go
func Marshal[Opts encoding.MarshalOptions](v any, opts Opts) ([]byte, error)
```

---

## Documenting Generic APIs

Document exported generic APIs thoroughly with motivating runnable examples:

```go
// Filter returns a new slice containing only elements for which keep returns
// true. It does not modify the original slice.
//
// Example:
//
//	evens := Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
//	// evens == []int{2, 4}
func Filter[T any](slice []T, keep func(T) bool) []T { ... }
```

---

## Type Aliases vs Type Definitions

A **type definition** creates a new distinct type with its own method set:

```go
type Celsius float64  // new type — can have methods
```

A **type alias** creates an alternate name for an existing type:

```go
type OldName = newpkg.Name  // alias — same type, same method set
```

Type aliases are rare. Use them only for package migration or gradual API
refactoring. Don't use aliasing when it is not needed.

---

## Constraint Composition

Combine constraints with `~` (underlying type) and `|` (union):

```go
type Numeric interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~float32 | ~float64
}

func Sum[T Numeric](vals []T) T {
    var total T
    for _, v := range vals {
        total += v
    }
    return total
}
```

Use the `constraints` package or `cmp` package (Go 1.21+) for standard constraints
like `cmp.Ordered` instead of writing your own.

---

## Common Pitfalls

### Don't Wrap Standard Library Types

```go
// Bad: generic wrapper adds complexity without value
type Set[T comparable] struct {
    m map[T]struct{}
}

// Better: use map[T]struct{} directly when the usage is simple
seen := map[string]struct{}{}
```

Generics justify their complexity when they eliminate duplication across
**multiple call sites**. A single-use generic is just indirection.

### Don't Use Generics for Interface Satisfaction

```go
// Bad: T is only used to satisfy an interface — just use the interface
func Process[T io.Reader](r T) error { ... }

// Good: accept the interface directly
func Process(r io.Reader) error { ... }
```

### Avoid Over-Constraining

```go
// Bad: constraint is more restrictive than needed
func Contains[T interface{ ~int | ~string }](slice []T, target T) bool { ... }

// Good: comparable is sufficient
func Contains[T comparable](slice []T, target T) bool { ... }
```

---

## Quick Reference

| Topic | Guidance |
|-------|----------|
| When to use generics | Only when multiple types share identical logic and interfaces don't suffice |
| Starting point | Write concrete code first; generalize later |
| Naming | Single uppercase letter (`T`, `K`, `V`, `E`) |
| Documentation | Thorough docs + runnable examples for exported generic APIs |
| Type definitions | New distinct type with own method set |
| Type aliases | Same type, alternate name; use only for migration |
| Constraint composition | Use `~` for underlying types, `|` for unions; prefer `cmp.Ordered` over custom |
| Common pitfall | Don't genericize single-use code or when interfaces suffice |

---

## See Also

- [go-interfaces](../go-interfaces/SKILL.md): Interface design and when interfaces suffice
- [go-declarations](../go-declarations/SKILL.md): Variable and type declaration patterns
- [go-documentation](../go-documentation/SKILL.md): Documenting APIs and writing examples
- [go-naming](../go-naming/SKILL.md): Naming conventions for types and functions
