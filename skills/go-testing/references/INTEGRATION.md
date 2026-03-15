# Go Testing: Integration and Advanced Patterns

Detailed reference for TestMain, acceptance testing, and real transport testing.
Sources: Google Go Style Guide (best-practices).

---

## TestMain

> **Source**: Google Go Style Guide (best-practices)

Use `func TestMain(m *testing.M)` when **all tests in the package** require
common setup that needs teardown (e.g., a shared database). This should **not be
your first choice**---prefer scoped test helpers or `t.Cleanup` when possible.

```go
var db *sql.DB

func TestInsert(t *testing.T) { /* uses db */ }
func TestSelect(t *testing.T) { /* uses db */ }

func runMain(ctx context.Context, m *testing.M) (code int, err error) {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    d, err := setupDatabase(ctx)
    if err != nil {
        return 0, err
    }
    defer d.Close()
    db = d

    return m.Run(), nil
}

func TestMain(m *testing.M) {
    code, err := runMain(context.Background(), m)
    if err != nil {
        log.Fatal(err)
    }
    // defer statements do not run past os.Exit
    os.Exit(code)
}
```

Key points:
- Extract setup into a helper function (`runMain`) so `defer` works correctly
- Write failure messages to stderr via `log.Fatal`
- Ensure individual test cases remain hermetic---reset any global state they modify

---

## Acceptance Testing

> **Source**: Google Go Style Guide (best-practices)

Acceptance testing validates that an implementation upholds a contract, treating
it as a black box. This pattern is useful when users implement your interfaces
and you want to provide a reusable validation suite.

### Structure

1. Create a test helper package (e.g., `chesstest` for package `chess`)
2. Export a validation function that accepts the implementation under test:

```go
// Package chesstest provides acceptance tests for chess.Player implementations.
package chesstest

// ExercisePlayer tests a Player implementation in a single turn.
// Returns nil if the player makes a correct move, or an error describing
// the violation.
func ExercisePlayer(b *chess.Board, p chess.Player) error {
    move := p.Move()
    if putsOwnKingIntoCheck(b, move) {
        return &IllegalMoveError{Move: move, Reason: "puts own king in check"}
    }
    return nil
}
```

3. End users write simple tests against the validation function:

```go
func TestAcceptance(t *testing.T) {
    player := deepblue.New()
    if err := chesstest.ExerciseGame(t, chesstest.SimpleGame, player); err != nil {
        t.Errorf("Deep Blue player failed acceptance test: %v", err)
    }
}
```

Reserve `t.Fatal` for setup failures only---validation errors should be returned,
not fataled.

---

## Use Real Transports

> **Source**: Google Go Style Guide (best-practices)

When testing component integrations over HTTP or RPC, prefer real transport
round-trips over hand-implemented client mocks:

```go
func TestAPIIntegration(t *testing.T) {
    // Start a test server with a fake backend
    srv := httptest.NewServer(newFakeHandler())
    t.Cleanup(srv.Close)

    // Use a real HTTP client against the test server
    client := api.NewClient(srv.URL)
    result, err := client.GetUser(context.Background(), "user-123")
    if err != nil {
        t.Fatalf("GetUser() error: %v", err)
    }
    if result.Name != "Test User" {
        t.Errorf("GetUser().Name = %q, want %q", result.Name, "Test User")
    }
}
```

Using the production client with a test server ensures your test exercises as
much real code as possible, avoiding the complexity of imitating client behavior.

---

## Common Mistakes

### Calling os.Exit directly in TestMain

`os.Exit` terminates the process immediately — deferred cleanup functions never
run. Extract setup/teardown into a helper so `defer` works correctly:

```go
// Bad: defers won't run
func TestMain(m *testing.M) {
    setup()
    defer cleanup()
    os.Exit(m.Run()) // cleanup() never executes
}

// Good: Extract to a helper function so defer runs before os.Exit
func runTests(m *testing.M) int {
    setup()
    defer cleanup()
    return m.Run()
}

func TestMain(m *testing.M) {
    os.Exit(runTests(m))
}
```
