# Variable Shadowing

> **Normative**: Be aware that `:=` in inner scopes creates a new variable that shadows the outer one.

### The Trap

```go
// Bug: err in the inner scope shadows the outer err
var err error
if condition {
    val, err := someFunc()  // new err — outer err stays nil
    use(val)
}
return err  // always nil!
```

### Fix: Assign to the Outer Variable

```go
var err error
if condition {
    var val int
    val, err = someFunc()  // assigns to outer err
    use(val)
}
return err  // correct
```

### Detection

Enable the `shadow` linter via `go vet`:

```bash
go vet -vettool=$(which shadow) ./...
```

Or add `govet` with shadow check enabled in `.golangci.yml`.
