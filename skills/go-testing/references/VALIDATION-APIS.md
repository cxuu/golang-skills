# Extensible Validation APIs

Detailed reference for designing reusable test validation functions that callers
can use for acceptance testing. Sources: Google Go Style Guide (best-practices).

---

## The `*test` Package Export Pattern

When you own an interface that others implement, export a validation function in
a companion `*test` package. This lets implementers verify correctness without
duplicating your test logic.

```go
// Package storagetest provides acceptance tests for storage.Backend.
package storagetest

// Verify runs a validation suite against any storage.Backend.
// Returns an error describing the first violation, or nil on success.
func Verify(b storage.Backend) error {
    if err := verifyRoundTrip(b); err != nil {
        return fmt.Errorf("round-trip: %w", err)
    }
    if err := verifyNotFound(b); err != nil {
        return fmt.Errorf("not-found: %w", err)
    }
    return nil
}
```

Callers write a thin test that plugs in their implementation:

```go
func TestMyBackend(t *testing.T) {
    b := mybackend.New(t)
    if err := storagetest.Verify(b); err != nil {
        t.Errorf("MyBackend failed acceptance: %v", err)
    }
}
```

---

## Designing Extensible Validation Functions

**Return errors, not `*testing.T` failures.** This keeps validation functions
usable as plain Go functions — callers decide whether a violation is `t.Error`
or `t.Fatal`.

```go
// Good: Returns error — caller controls test flow
func ExercisePlayer(b *chess.Board, p chess.Player) error {
    move := p.Move()
    if putsOwnKingIntoCheck(b, move) {
        return &IllegalMoveError{Move: move, Reason: "puts own king in check"}
    }
    return nil
}

// Bad: Calls t.Fatal — caller loses control
func ExercisePlayer(t *testing.T, b *chess.Board, p chess.Player) {
    t.Helper()
    move := p.Move()
    if putsOwnKingIntoCheck(b, move) {
        t.Fatalf("illegal move: %v puts own king in check", move)
    }
}
```

**Use custom error types** for rich diagnostics when needed:

```go
type IllegalMoveError struct {
    Move   chess.Move
    Reason string
}

func (e *IllegalMoveError) Error() string {
    return fmt.Sprintf("illegal move %v: %s", e.Move, e.Reason)
}
```

---

## When to Use Validation APIs vs Simple Helpers

| Situation | Use |
|-----------|-----|
| Interface you own, others implement | Validation API in `*test` package |
| Shared setup across tests in one package | Test helper with `t.Helper()` |
| Complex assertion reused in 2-3 tests | Helper returning `error` or `bool` |
| One-off setup or comparison | Inline test code |

**Validation APIs** are worth the extra package when:
- Multiple external packages will implement your interface
- The contract has non-obvious invariants that are easy to get wrong
- You want a single source of truth for "correct behavior"

**Simple helpers** are better when:
- The helper is a straightforward setup or comparison function
- The reuse is incidental, not part of a published contract

---

## Naming Conventions

Name the function with a verb that signals scope: `Verify`, `Exercise`,
`RunConformance`. Accept the interface under test as a parameter — never
construct the implementation inside the validation package.

| Package | Function | Purpose |
|---------|----------|---------|
| `storagetest` | `Verify` | Validates a `storage.Backend` |
| `chesstest` | `ExercisePlayer` | Validates a `chess.Player` |
| `cachetest` | `RunConformance` | Full conformance suite for `cache.Cache` |
