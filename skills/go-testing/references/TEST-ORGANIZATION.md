# Test Organization Reference

## Test Doubles

> **Advisory**: Follow consistent naming for test doubles (stubs, fakes, mocks,
> spies).

**Package naming**: Append `test` to the production package (e.g.,
`creditcardtest`).

```go
// Good: In package creditcardtest

// Single double - use simple name
type Stub struct{}
func (Stub) Charge(*creditcard.Card, money.Money) error { return nil }

// Multiple behaviors - name by behavior
type AlwaysCharges struct{}
type AlwaysDeclines struct{}

// Multiple types - include type name
type StubService struct{}
type StubStoredValue struct{}
```

**Local variables**: Prefix test double variables for clarity (`spyCC` not
`cc`).

---

## Test Packages

| Package Declaration | Use Case |
|---------------------|----------|
| `package foo` | Same-package tests, can access unexported identifiers |
| `package foo_test` | Black-box tests, avoids circular dependencies |

Both go in `foo_test.go` files. Use `_test` suffix when testing only public API
or to break import cycles.

---

## Setup Scoping

> **Advisory**: Keep setup scoped to tests that need it.

```go
// Good: Explicit setup in tests that need it
func TestParseData(t *testing.T) {
    data := mustLoadDataset(t)
    // ...
}

func TestUnrelated(t *testing.T) {
    // Doesn't pay for dataset loading
}

// Bad: Global init loads data for all tests
var dataset []byte

func init() {
    dataset = mustLoadDataset()  // Runs even for unrelated tests
}
```
