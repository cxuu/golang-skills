---
name: go-generics
description: Go generics guidance including when to use type parameters, constraints, and type aliases vs type definitions. Use when deciding whether to use generics, writing generic functions or types, or choosing between type aliases and definitions.
---

# Go Generics and Type Parameters

> **Source**: Google Go Style Guide (Normative)

---

## When to Use Generics

Generics (formally "[type parameters](https://go.dev/design/43651-type-parameters)")
are allowed where they fulfill your business requirements. However, in many
applications a conventional approach using slices, maps, interfaces, and
concrete types works just as well without the added complexity.

### Prefer Generics When

- Multiple types share a useful algorithm and the logic is identical across
  types (e.g., sorting, filtering, map/reduce operations)
- You would otherwise rely on `any` and excessive type switching
- You are building a reusable data structure (e.g., a concurrent-safe set,
  ordered map, or ring buffer)

### Avoid Generics When

- Only one type is being instantiated in practice—start with the concrete type
  and add polymorphism later when needed
- Interfaces already model the shared behavior cleanly
- You are inventing a domain-specific language or error-handling framework that
  burdens readers
- The generic code is harder to read than the type-specific alternative

> "Write code, don't design types." — Robert Griesemer and Ian Lance Taylor,
> GopherCon

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
// Start concrete; generalize only when a second type appears
func SumInts(vals []int) int {
    var total int
    for _, v := range vals {
        total += v
    }
    return total
}
```

---

## Type Parameter Conventions

### Naming

Use single uppercase letters for type parameters by convention:

| Name | Typical Use |
|------|-------------|
| `T` | General type parameter |
| `K` | Map key type |
| `V` | Map value type |
| `E` | Element/item type |
| `S` | Slice type |

For more complex constraints, a short descriptive name is acceptable:

```go
func Marshal[Opts encoding.MarshalOptions](v any, opts Opts) ([]byte, error)
```

### Constraints

Use the built-in and `constraints` package types:

```go
// Built-in constraint
func Print[T any](v T) { fmt.Println(v) }

// Comparable constraint (supports == and !=)
func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target {
            return true
        }
    }
    return false
}

// Union constraint
func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}
```

---

## Documenting Generic APIs

When introducing an exported generic API, document it thoroughly and include
motivating runnable examples:

```go
// Filter returns a new slice containing only elements for which keep returns
// true. It does not modify the original slice.
//
// Example:
//
//	evens := Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
//	// evens == []int{2, 4}
func Filter[T any](slice []T, keep func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if keep(v) {
            result = append(result, v)
        }
    }
    return result
}
```

---

## Type Aliases vs Type Definitions

> **Source**: Google Go Style Guide (Normative)

A **type definition** creates a new distinct type:

```go
type UserID string // new type — has its own method set
```

A **type alias** creates an alternate name for an existing type:

```go
type UserID = string // alias — same type, same method set
```

### When to Use Each

| Use | Mechanism | Example |
|-----|-----------|---------|
| New behavior (methods, validation) | Type definition (`type T1 T2`) | `type Celsius float64` |
| Package migration / renaming | Type alias (`type T1 = T2`) | `type OldName = newpkg.Name` |
| Gradual API refactoring | Type alias | Moving types between packages |

Type aliases are rare. Their primary use is to aid migrating packages to new
source code locations. Don't use type aliasing when it is not needed.

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
