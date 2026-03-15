# Switch Patterns

Detailed patterns for Go `switch` statements, including expression-less
switches, comma cases, break behavior, and labeled breaks.

---

## No Automatic Fallthrough

Go `switch` cases do **not** fall through by default (unlike C/Java). Each case
body implicitly breaks. Use `fallthrough` only when explicitly needed — it is
rare in idiomatic Go.

```go
switch n {
case 1:
    fmt.Println("one")
    // no fallthrough — next case is NOT executed
case 2:
    fmt.Println("two")
}
```

---

## Expression-less Switch

A `switch` with no expression switches on `true`. Use it for clean if-else-if
chains when comparing a single variable against multiple conditions:

```go
func unhex(c byte) byte {
    switch {
    case '0' <= c && c <= '9':
        return c - '0'
    case 'a' <= c && c <= 'f':
        return c - 'a' + 10
    case 'A' <= c && c <= 'F':
        return c - 'A' + 10
    }
    return 0
}
```

---

## Comma-Separated Cases

Multiple values can share a single case body using commas — no need for
`fallthrough`:

```go
func shouldEscape(c byte) bool {
    switch c {
    case ' ', '?', '&', '=', '#', '+', '%':
        return true
    }
    return false
}
```

---

## Break with Labels

`break` inside a `switch` terminates only the switch, **not** an enclosing
`for` loop. Use a label to break out of the loop:

```go
Loop:
    for n := 0; n < len(src); n += size {
        switch {
        case src[n] < sizeOne:
            break        // breaks switch only
        case src[n] < sizeTwo:
            if n+1 >= len(src) {
                break Loop   // breaks out of for loop
            }
        }
    }
```

Another common pattern — breaking a range loop from inside a switch:

```go
Loop:
    for _, v := range items {
        switch v.Type {
        case "done":
            break Loop  // breaks the for loop
        case "skip":
            break  // breaks only the switch
        }
    }
```

**Rule of thumb**: Whenever you have a `switch` inside a `for` and need to
exit the loop from a case, always use a labeled break.

---

## Type Switches

For type switches (`switch v := x.(type)`), see
[go-interfaces](../../go-interfaces/SKILL.md): Type Switch.

---

## Quick Reference

| Pattern | Syntax |
|---------|--------|
| Expression-less switch | `switch { case cond: }` |
| Comma cases | `case 'a', 'b', 'c':` |
| No fallthrough | Default; use `fallthrough` keyword if needed |
| Break switch only | `break` inside case |
| Break enclosing loop | `break Label` with labeled `for` |
