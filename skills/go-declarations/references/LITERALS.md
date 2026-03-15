# Composite Literal Formatting

Sources: Google Go Style Guide (decisions), Effective Go.

---

## Struct Literal Rules

**Always use field names** for struct literals. This is enforced by `go vet` for
types from other packages and prevents breakage when fields are reordered.

```go
// Good: Named fields
cfg := Config{
    Host:    "localhost",
    Port:    8080,
    Timeout: 30 * time.Second,
}

// Bad: Positional — fragile and unreadable
cfg := Config{"localhost", 8080, 30 * time.Second}
```

**When field names may be omitted:**

- Test table rows with 3 or fewer fields where meaning is obvious
- Coordinate-like types: `image.Point{0, 0}`, `color.RGBA{255, 0, 0, 255}`
- Types where field order is part of the documented API

---

## Multi-Line vs Single-Line Structs

**Single-line** for 1-2 short fields:

```go
p := Point{X: 1, Y: 2}
e := Entry{Key: "name", Value: "Alice"}
```

**Multi-line** for 3+ fields, long values, or when readability benefits:

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
}
```

The closing brace sits on its own line, aligned with the opening identifier.

---

## Cuddled Braces

When a composite literal is used as a function argument or assignment, the
opening brace "cuddles" with the preceding token — no line break between them:

```go
// Good: Brace cuddles with the function call
db.SetConnMaxLifetime(ConnConfig{
    MaxOpen:     25,
    MaxIdle:     5,
    MaxLifetime: 5 * time.Minute,
})

// Bad: Unnecessary line break before brace
cfg :=
    Config{
        Verbose: true,
        Output:  os.Stdout,
    }
```

---

## Slice Literal Formatting

**Short slices** fit on one line:

```go
primes := []int{2, 3, 5, 7, 11}
```

**Long slices** use one element per line with a trailing comma:

```go
endpoints := []string{
    "/api/users",
    "/api/posts",
    "/api/comments",
    "/healthz",
}
```

### Omitting Repeated Type Names

Omit the element type name in slice literals — `gofmt -s` removes them:

```go
// Good: Type name omitted
items := []Item{
    {Name: "widget", Price: 9.99},
    {Name: "gadget", Price: 19.99},
}

// Bad: Redundant type names
items := []Item{
    Item{Name: "widget", Price: 9.99},
    Item{Name: "gadget", Price: 19.99},
}
```

The same applies to pointer slices — use `{...}` not `&Node{...}`.

## Map Literal Formatting

**Short maps** fit on one line:

```go
counts := map[string]int{"a": 1, "b": 2}
```

**Multi-line maps** use one entry per line with trailing commas:

```go
headers := map[string]string{
    "Content-Type":  "application/json",
    "Authorization": "Bearer " + token,
    "X-Request-ID":  reqID,
}
```

---

## Function Literal Formatting

**Short closures** can stay on one line when used as arguments:

```go
sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
```

**Multi-line closures** follow standard indentation:

```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "ok")
})
```

If a closure exceeds ~15 lines or captures many variables, extract it into a
named function for readability and testability.

---

## Multi-Line Wrapping Rules

When a composite literal doesn't fit on one line, follow these rules:

1. Opening brace stays on the same line as the declaration or call
2. Each element on its own line, indented one level, with trailing comma
3. Closing brace on its own line at the original indentation level

```go
// Good: Follows all wrapping rules
resp := &Response{
    StatusCode: http.StatusOK,
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
    Body: mustMarshal(data),
}
```

Go requires the trailing comma when the closing brace is on a separate line.
