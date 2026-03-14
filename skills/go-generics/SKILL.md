---
name: go-generics
description: Go generics guidance including when to use type parameters, constraints, and type aliases vs type definitions. Use when deciding whether to use generics, writing generic functions or types, or choosing between type aliases and definitions.
sources: [Google Style Guide]
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

## Quick Reference

| Topic | Guidance |
|-------|----------|
| When to use generics | Only when multiple types share identical logic and interfaces don't suffice |
| Starting point | Write concrete code first; generalize later |
| Naming | Single uppercase letter (`T`, `K`, `V`, `E`) |
| Documentation | Thorough docs + runnable examples for exported generic APIs |
| Type definitions | New distinct type with own method set |
| Type aliases | Same type, alternate name; use only for migration |

---

## See Also

- **go-interfaces**: Interface design and when interfaces suffice
- **go-declarations**: Variable and type declaration patterns
- **go-documentation**: Documenting APIs and writing examples
- **go-naming**: Naming conventions for types and functions
