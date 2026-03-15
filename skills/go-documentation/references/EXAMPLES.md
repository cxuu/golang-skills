# Package Comments and Examples Reference

## Package Comments

> **Normative**: Every package must have exactly one package comment.

```go
// Good:
// Package math provides basic constants and mathematical functions.
//
// This package does not guarantee bit-identical results across architectures.
package math
```

### Main Packages

Use the binary name (matching the BUILD file):

```go
// Good:
// The seed_generator command is a utility that generates a Finch seed file
// from a set of JSON study configs.
package main
```

Valid styles: `Binary seed_generator`, `Command seed_generator`, `The
seed_generator command`, `Seed_generator ...`

### doc.go

- For long package comments, use a `doc.go` file containing only the package
  comment and the `package` clause
- Maintainer comments placed after imports don't appear in Godoc
- Keep the doc.go file focused on user-facing documentation

```go
// Package complex provides advanced mathematical operations for
// complex number arithmetic, including polar form conversion,
// matrix operations, and numerical integration.
//
// Basic usage
//
// Create a complex number and perform operations:
//
//   z := complex.New(3, 4)
//   magnitude := z.Abs()    // 5.0
//   conjugate := z.Conj()   // (3, -4)
//
// Matrix operations
//
// The package supports complex-valued matrices:
//
//   m := complex.NewMatrix(2, 2)
//   m.Set(0, 0, complex.New(1, 0))
//   det := m.Det()
package complex
```

---

## Runnable Examples

> **Advisory**: Provide runnable examples to demonstrate package usage.

Place examples in test files (`*_test.go`):

```go
// Good:
func ExampleConfig_WriteTo() {
    cfg := &Config{
        Name: "example",
    }
    if err := cfg.WriteTo(os.Stdout); err != nil {
        log.Exitf("Failed to write config: %s", err)
    }
    // Output:
    // {
    //   "name": "example"
    // }
}
```

Examples appear in Godoc attached to the documented element.

### Naming Conventions

| Function Name | Documents |
|---------------|-----------|
| `Example()` | Package-level example |
| `ExampleFoo()` | Function `Foo` |
| `ExampleBar_Baz()` | Method `Bar.Baz` |
| `ExampleFoo_suffix()` | Named variant of `Foo` example |

### Tips

- Use `// Output:` comments to make examples testable and verifiable by `go test`
- Keep examples focused on demonstrating one concept
- Use realistic but minimal data
- For complex setup, use a `testMain` or helper to keep the example body clean
- Multiple examples for the same symbol use a lowercase `_suffix`:

```go
func ExampleNewClient_withTimeout() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    client := NewClient(ctx)
    // ...
}
```
