# Composite Literal Formatting

> **Source**: Google Go Style Guide (decisions.md)

Detailed rules for formatting composite literals (struct, slice, map) in Go.

---

## Field Names

Struct literals must specify **field names** for types defined outside the
current package:

```go
// Good: external package type — use field names
r := csv.Reader{
    Comma:           ',',
    Comment:         '#',
    FieldsPerRecord: 4,
}

// Bad: positional — fragile and unreadable
r := csv.Reader{',', '#', 4, false, false, false, false}
```

For package-local types, field names are optional but recommended when the
struct has many fields:

```go
// Acceptable: small internal type
okay := Type{42}

// Recommended: many fields
okay := StructWithLotsOfFields{
    field1: 1,
    field2: "two",
    field3: 3.14,
    field4: true,
}
```

---

## Matching Braces

The closing brace must appear on a line with the same indentation as the
opening brace. Don't put the closing brace on the same line as a value in a
multi-line literal:

```go
// Good
[]*Type{
    {Key: "multi"},
    {Key: "line"},
}

// Bad: closing brace on value line
[]*Type{
    {Key: "multi"},
    {Key: "line"}}
```

---

## Cuddled Braces

Dropping whitespace between braces ("cuddling") is permitted only when both:

1. Indentation matches
2. Inner values are also literals (not variables or expressions)

```go
// Good: cuddled braces with literal inner values
[]*Type{{
    Field: "value",
}, {
    Field: "value",
}}

// Bad: cuddled with non-literal inner value
[]*Type{
    first,
    {
        Field: "second",
    }}
```

---

## Repeated Type Names

Repeated type names may be omitted in slice and map literals:

```go
// Good: type names omitted (cleaner)
[]*Type{
    {A: 42},
    {A: 43},
}

// Bad: redundant type names
[]*Type{
    &Type{A: 42},
    &Type{A: 43},
}
```

**Tip**: Run `gofmt -s` to remove repetitive type names automatically.

---

## Zero-Value Fields

Omit zero-value fields when doing so does not reduce clarity. Well-designed APIs
use zero-value construction to draw attention to the options being specified:

```go
// Good: zero fields omitted, important ones stand out
ldb := leveldb.Open("/my/table", &db.Options{
    BlockSize:       1 << 16,
    ErrorIfDBExists: true,
})

// Bad: noise from zero fields
ldb := leveldb.Open("/my/table", &db.Options{
    BlockSize:            1 << 16,
    ErrorIfDBExists:      true,
    BlockRestartInterval: 0,
    // ... all zero fields listed ...
})
```

Exception: table-driven test structs often benefit from explicit field names
even for zero values to clarify the test case.
