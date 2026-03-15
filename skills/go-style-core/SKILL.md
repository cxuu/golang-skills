---
name: go-style-core
description: Use for questions about general Go formatting, line length, nesting reduction, naked returns, semicolon rules, or foundational style principles like clarity vs simplicity vs concision. Also use as a fallback when a Go style question isn't covered by a more specific skill (naming, error handling, testing, etc.). Helps resolve formatting debates and teaches the priority order of Go style principles (clarity > simplicity > concision > maintainability > consistency).
license: Apache-2.0
metadata:
  sources: "Effective Go, Google Style Guide, Uber Style Guide, Go Wiki CodeReviewComments"
---

# Go Style Core Principles

## Style Principles (Priority Order)

When writing readable Go code, apply these principles in order of importance:

### Priority Order

1. **Clarity** — Can a reader understand the code without extra context?
2. **Simplicity** — Is this the simplest way to accomplish the goal?
3. **Concision** — Does every line earn its place?
4. **Maintainability** — Will this be easy to modify later?
5. **Consistency** — Does it match surrounding code and project conventions?

> For detailed explanations and examples of each principle, see [references/PRINCIPLES.md](references/PRINCIPLES.md).

---

## Formatting

### gofmt is Required

All Go source files **must** conform to `gofmt` output. No exceptions.

```bash
# Format a file
gofmt -w myfile.go

# Format all files in directory
gofmt -w .
```

### Parentheses

Go needs fewer parentheses than C and Java. Control structures (`if`, `for`, `switch`) don't have parentheses in their syntax. The operator precedence hierarchy is shorter and clearer, so `x<<8 + y<<16` means what the spacing suggests—unlike in other languages.

### MixedCaps (Camel Case)

Go uses `MixedCaps` or `mixedCaps`, never underscores:

```go
// Good
MaxLength    // exported constant
maxLength    // unexported constant
userID       // variable

// Bad
MAX_LENGTH   // no snake_case
max_length   // no underscores
```

Exceptions:
- Test function names may use underscores: `TestFoo_Bar`
- Generated code interoperating with OS/cgo

### Line Length

There is **no rigid line length limit** in Go, but avoid uncomfortably long
lines. Uber suggests a soft limit of 99 characters.

Guidelines:
- If a line feels too long, **refactor** rather than just wrap
- Don't split before indentation changes (function declarations, conditionals)
- Don't split long strings (URLs) into multiple lines
- When splitting, put all arguments on their own lines
- If it's already as short as practical, let it remain long

**Break by semantics, not length**:

Don't add line breaks just to keep lines short when they are more readable long
(e.g., repetitive lines). Break lines because of what you're writing, not
because of line length.

Long lines often correlate with long names. If you find lines are too long,
consider whether the names could be shorter. Getting rid of long names often
helps more than wrapping lines.

This advice applies equally to function length—there's no rule "never have a
function more than N lines", but there is such a thing as too long. The solution
is to change where function boundaries are, not to count lines.

```go
// Bad: Arbitrary mid-line break
func (s *Store) GetUser(ctx context.Context,
    id string) (*User, error) {

// Good: All arguments on own lines
func (s *Store) GetUser(
    ctx context.Context,
    id string,
) (*User, error) {
```

### Local Consistency

When the style guide is silent, be consistent with nearby code:

**Valid** local choices:
- `%s` vs `%v` for error formatting
- Buffered channels vs mutexes

**Invalid** local overrides:
- Line length restrictions
- Assertion-based testing libraries

---

## Reduce Nesting

Handle error cases and special conditions first. Return early or continue the loop to keep the "happy path" unindented.

```go
// Bad: Deeply nested
for _, v := range data {
    if v.F1 == 1 {
        v = process(v)
        if err := v.Call(); err == nil {
            v.Send()
        } else {
            return err
        }
    } else {
        log.Printf("Invalid v: %v", v)
    }
}

// Good: Flat structure with early returns
for _, v := range data {
    if v.F1 != 1 {
        log.Printf("Invalid v: %v", v)
        continue
    }

    v = process(v)
    if err := v.Call(); err != nil {
        return err
    }
    v.Send()
}
```

### Unnecessary Else

If a variable is set in both branches of an if, use default + override pattern.

```go
// Bad: Setting in both branches
var a int
if b {
    a = 100
} else {
    a = 10
}

// Good: Default + override
a := 10
if b {
    a = 100
}
```

---

## Naked Returns

A `return` statement without arguments returns the named return values. This is
known as a "naked" return.

```go
func split(sum int) (x, y int) {
    x = sum * 4 / 9
    y = sum - x
    return  // returns x, y
}
```

### Guidelines for Naked Returns

- **OK in small functions**: Naked returns are fine in functions that are just a
  handful of lines
- **Be explicit in medium+ functions**: Once a function grows to medium size, be
  explicit with return values for clarity
- **Don't name results just for naked returns**: Clarity of documentation is
  always more important than saving a line or two. Don't name result parameters
  just because it enables naked returns

```go
// Good: Small function, naked return is clear
func minMax(a, b int) (min, max int) {
    if a < b {
        min, max = a, b
    } else {
        min, max = b, a
    }
    return
}

// Good: Larger function, explicit return
func processData(data []byte) (result []byte, err error) {
    result = make([]byte, 0, len(data))

    for _, b := range data {
        if b == 0 {
            return nil, errors.New("null byte in data")
        }
        result = append(result, transform(b))
    }

    return result, nil  // explicit: clearer in longer functions
}
```

See **go-documentation** for guidance on Named Result Parameters.

---

## Semicolons

Go's lexer automatically inserts semicolons after any line whose last token is
an identifier, literal, or one of: `break continue fallthrough return ++ -- ) }`.

This means **opening braces must be on the same line** as the control structure:

```go
// Good: brace on same line
if i < f() {
    g()
}

// Bad: brace on next line — lexer inserts semicolon after f()
if i < f()  // wrong!
{           // wrong!
    g()
}
```

Idiomatic Go only has explicit semicolons in `for` loop clauses and to separate
multiple statements on a single line.

---

## Quick Reference

| Principle | Key Question |
|-----------|--------------|
| Clarity | Can a reader understand what and why? |
| Simplicity | Is this the simplest approach? |
| Concision | Is the signal-to-noise ratio high? |
| Maintainability | Can this be safely modified later? |
| Consistency | Does this match surrounding code? |

## See Also

- [go-naming](../go-naming/SKILL.md): Naming conventions
- [go-error-handling](../go-error-handling/SKILL.md): Error handling patterns
- [go-documentation](../go-documentation/SKILL.md): Documentation guidelines
- [go-testing](../go-testing/SKILL.md): Testing best practices
- [go-defensive](../go-defensive/SKILL.md): Defensive programming
- [go-performance](../go-performance/SKILL.md): Performance optimization
- [go-linting](../go-linting/SKILL.md): Linting and static analysis
