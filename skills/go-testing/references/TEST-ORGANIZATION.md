# Test Organization Reference

Sources: Google Go Style Guide (best-practices, decisions).

---

## Test Double Types

| Double | Purpose | State? | Verifies calls? |
|--------|---------|--------|-----------------|
| Stub   | Returns canned data | No | No |
| Fake   | Working but simplified implementation | Yes | No |
| Spy    | Records calls for later inspection | Yes | Yes |

**Prefer fakes over mocks.** Fakes are more readable and don't require mock
frameworks. Reserve spies for verifying side effects (e.g., an analytics event).

```go
// Fake: Working in-memory implementation
type FakeUserStore struct {
    users map[string]*User
}

func (f *FakeUserStore) GetUser(id string) (*User, error) {
    u, ok := f.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return u, nil
}

// Spy: Records calls for later assertion
type SpyEmailSender struct{ Sent []string }

func (s *SpyEmailSender) Send(to, body string) error {
    s.Sent = append(s.Sent, to)
    return nil
}
```

---

## Test Double Naming Conventions

> **Advisory**: Follow consistent naming for test doubles (stubs, fakes, spies).

**Package naming**: Create a `*test` package alongside production code (e.g.,
`creditcardtest` for package `creditcard`, `fakeauthservice` for a standalone
fake service).

```go
// Good: In package creditcardtest

// Single double — use simple name
type Stub struct{}
func (Stub) Charge(*creditcard.Card, money.Money) error { return nil }

// Multiple behaviors — name by behavior
type AlwaysCharges struct{}
type AlwaysDeclines struct{}

// Multiple types — include type name
type StubService struct{}
type StubStoredValue struct{}
```

**Local variables**: Prefix test double variables with the double type for
clarity at the call site:

```go
// Good: Double type is immediately visible
spyCC := &creditcardtest.Spy{}
stubDB := &dbtest.Stub{Balance: 100}

// Bad: Ambiguous — is this real or a double?
cc := &creditcardtest.Spy{}
db := &dbtest.Stub{Balance: 100}
```

---

## Standalone Test Helper Packages

Create a standalone test helper package when multiple packages need the same
double, the helper has enough logic to warrant its own tests, or you want to
provide an acceptance test suite for interface implementers.

| Pattern | When to use | Example |
|---------|-------------|---------|
| `footest` | General test helpers for package `foo` | `creditcardtest`, `usertest` |
| `fakeX` | Standalone fake service package | `fakeauthservice`, `fakestorage` |

```go
package usertest

func NewFakeStore(t *testing.T, users ...*user.User) *FakeUserStore {
    t.Helper()
    store := &FakeUserStore{users: make(map[string]*user.User)}
    for _, u := range users {
        store.users[u.ID] = u
    }
    return store
}
```

Export constructors that accept `*testing.T` so they can call `t.Helper()` and
`t.Cleanup()`.

---

## Test Packages

| Package Declaration | Use Case |
|---------------------|----------|
| `package foo` | Same-package tests, can access unexported identifiers |
| `package foo_test` | Black-box tests, avoids circular dependencies |

Both go in `foo_test.go` files in the same directory.

**Use `package foo` (white-box)** when you need to test unexported functions or
internal state.

**Use `package foo_test` (black-box)** when testing only the public API, breaking
import cycles, or verifying external usability.

```go
package parser_test  // Black-box: only tests exported API

import "mymodule/parser"

func TestParse(t *testing.T) {
    got, err := parser.Parse("input")
    // ...
}
```

If a black-box test needs an unexported symbol, create `export_test.go` in
`package foo` (not `foo_test`) that exposes it. Use this sparingly.

---

## Setup Scoping

> **Advisory**: Keep setup scoped to tests that need it.

Explicit setup in each test is clearer and avoids penalizing unrelated tests:

```go
// Good: Explicit setup in tests that need it
func TestParseData(t *testing.T) {
    data := mustLoadDataset(t)
    // ...
}

func TestUnrelated(t *testing.T) {
    // Doesn't pay for dataset loading
}
```

**Avoid global `init` for test setup** — it runs for every test in the file,
even unrelated ones.

**Subtest setup**: Use a parent test with `t.Run` when a group of subtests
shares setup:

```go
func TestDatabase(t *testing.T) {
    db := setupTestDB(t)

    t.Run("Insert", func(t *testing.T) {
        // uses db
    })
    t.Run("Select", func(t *testing.T) {
        // uses db
    })
}
```

This scopes the database lifecycle to the subtests that need it. Use `TestMain`
only as a last resort (see [INTEGRATION.md](INTEGRATION.md)).
