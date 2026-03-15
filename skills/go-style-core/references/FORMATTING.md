# Formatting Reference

## gofmt is Required

All Go source files **must** conform to `gofmt` output. No exceptions.

```bash
# Format a file
gofmt -w myfile.go

# Format all files in directory
gofmt -w .
```

Additional formatting tools:

| Tool | Purpose |
|------|---------|
| `gofmt` | Standard formatter (required) |
| `goimports` | gofmt + import management |
| `gofumpt` | Stricter superset of gofmt |

---

## Parentheses

Go needs fewer parentheses than C and Java. Control structures (`if`, `for`,
`switch`) don't have parentheses in their syntax. The operator precedence
hierarchy is shorter and clearer, so `x<<8 + y<<16` means what the spacing
suggests—unlike in other languages.

---

## MixedCaps (Camel Case)

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

---

## Line Length

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

---

## Local Consistency

When the style guide is silent, be consistent with nearby code:

**Valid** local choices:
- `%s` vs `%v` for error formatting
- Buffered channels vs mutexes

**Invalid** local overrides:
- Line length restrictions
- Assertion-based testing libraries
