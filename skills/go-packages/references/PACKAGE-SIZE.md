# Package Size, Program Structure, and CLIs

Detailed guidance on package splitting, avoiding init(), the run() pattern, and
CLI structure.

## When to Split a Package

```
Is the package getting too large?
├─ Can you describe its purpose in one sentence?
│  ├─ No → Split by responsibility
│  └─ Yes → Keep it, but check below
├─ Do files in the package never import each other's unexported symbols?
│  └─ Yes → Those files could be separate packages
├─ Does the package have distinct user groups using different parts?
│  └─ Yes → Split along user boundaries
└─ Is the godoc page overwhelming?
   └─ Yes → Split to improve discoverability
```

### When NOT to Split

- Don't split just because a file is long — large files in a focused package are
  fine
- Don't create packages with only one type or function
- Don't split if it would create circular dependencies
- Avoid splitting internal helpers into a `util` or `internal/helpers` package

### When to Combine Packages

- If client code likely needs two types to interact, keep them together
- If types have tightly coupled implementations
- If users would need to import both packages to use either meaningfully

### File Organization

No "one type, one file" convention in Go. Files should be focused enough to know
which file contains something and small enough to find things easily.

---

## Avoiding init()

Prefer explicit functions over `init()`:

```go
// Bad: init() with I/O and environment dependencies
var _config Config

func init() {
    cwd, _ := os.Getwd()
    raw, _ := os.ReadFile(path.Join(cwd, "config.yaml"))
    yaml.Unmarshal(raw, &_config)
}
```

```go
// Good: Explicit function for loading config
func loadConfig() (Config, error) {
    cwd, err := os.Getwd()
    if err != nil {
        return Config{}, err
    }

    raw, err := os.ReadFile(path.Join(cwd, "config.yaml"))
    if err != nil {
        return Config{}, err
    }

    var config Config
    if err := yaml.Unmarshal(raw, &config); err != nil {
        return Config{}, err
    }
    return config, nil
}
```

**Acceptable uses of init():**
- Complex expressions that cannot be single assignments
- Pluggable hooks (e.g., `database/sql` dialects, encoding registries)
- Deterministic precomputation

---

## Exit in Main

Call `os.Exit` or `log.Fatal*` **only in `main()`**. All other functions should
return errors to signal failure.

**Why this matters:**
- Non-obvious control flow: Any function can exit the program
- Difficult to test: Functions that exit also exit the test
- Skipped cleanup: `defer` statements are skipped

```go
// Bad: log.Fatal in helper function
func readFile(path string) string {
    f, err := os.Open(path)
    if err != nil {
        log.Fatal(err)  // Exits program, skips defers
    }
    b, err := io.ReadAll(f)
    if err != nil {
        log.Fatal(err)
    }
    return string(b)
}
```

```go
// Good: Return errors, let main() decide to exit
func main() {
    body, err := readFile(path)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(body)
}

func readFile(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    b, err := io.ReadAll(f)
    if err != nil {
        return "", err
    }
    return string(b), nil
}
```

### The run() Pattern

Prefer to call `os.Exit` or `log.Fatal` **at most once** in `main()`. Extract
business logic into a separate function that returns errors.

```go
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}

func run() error {
    args := os.Args[1:]
    if len(args) != 1 {
        return errors.New("missing file")
    }

    f, err := os.Open(args[0])
    if err != nil {
        return err
    }
    defer f.Close()  // Will always run

    b, err := io.ReadAll(f)
    if err != nil {
        return err
    }

    // Process b...
    return nil
}
```

**Benefits of the `run()` pattern:**
- Short `main()` function with single exit point
- All business logic is testable
- `defer` statements always execute

---

## Command-Line Interfaces

### Flag Naming

Use lowercase, hyphen-separated flag names:

```go
// Good
flag.String("output-dir", ".", "directory for output files")
flag.Bool("dry-run", false, "print actions without executing")

// Bad
flag.String("outputDir", ".", "")    // camelCase
flag.String("output_dir", ".", "")   // underscores
```

### Subcommands

For complex CLIs with subcommands, use `flag.NewFlagSet` per subcommand:

```go
func main() {
    serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
    port := serveCmd.Int("port", 8080, "listen port")

    migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
    dryRun := migrateCmd.Bool("dry-run", false, "preview changes")

    switch os.Args[1] {
    case "serve":
        serveCmd.Parse(os.Args[2:])
        runServe(*port)
    case "migrate":
        migrateCmd.Parse(os.Args[2:])
        runMigrate(*dryRun)
    default:
        fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
        os.Exit(1)
    }
}
```

For larger CLIs, consider libraries like `cobra` or `urfave/cli`. Exit only from
`main()`.
