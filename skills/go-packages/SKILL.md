---
name: go-packages
description: Use when creating Go packages, organizing imports, managing dependencies, or deciding how to structure Go code into packages. Also use when starting a new Go project or splitting a growing codebase into packages, even if the user doesn't explicitly ask about package organization. Helps avoid util packages, organize import groups, and teaches why to avoid init() and the pattern of exiting only from main().
license: Apache-2.0
metadata:
  sources: "Google Style Guide, Uber Style Guide, Go Wiki CodeReviewComments"
---

# Go Packages and Imports

## Package Organization

### Avoid Util Packages

> **Advisory**: This is a best practice recommendation.

Package names should describe what the package provides. Avoid generic names
like `util`, `helper`, `common`, or similar—they make code harder to read and
cause import conflicts.

```go
// Good: Meaningful package names
db := spannertest.NewDatabaseFromFile(...)
_, err := f.Seek(0, io.SeekStart)

// Bad: Vague package names obscure meaning
db := test.NewDatabaseFromFile(...)
_, err := f.Seek(0, common.SeekStart)
```

Generic names like `util` can be used as *part* of a name (e.g., `stringutil`)
but should not be the entire package name.

### Package Size

> **Advisory**: This is best practice guidance.

**When to combine packages:**
- If client code likely needs two types to interact, keep them together
- If types have tightly coupled implementations
- If users would need to import both packages to use either meaningfully

**When to split packages:**
- When something is conceptually distinct
- The short package name + exported type creates a meaningful identifier:
  `bytes.Buffer`, `ring.New`

**File organization:** No "one type, one file" convention in Go. Files should be
focused enough to know which file contains something and small enough to find
things easily.

---

## Imports

### Import Organization

Imports are organized in groups, with blank lines between them. The standard
library packages are always in the first group.

```go
package main

import (
	"fmt"
	"hash/adler32"
	"os"

	"github.com/foo/bar"
	"rsc.io/goversion/version"
)
```

Use [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) to manage
this automatically.

### Import Grouping (Extended)

**Minimal grouping (Uber):** stdlib, then everything else.

**Extended grouping (Google):** stdlib → other → protocol buffers → side-effects.

```go
// Good: Standard library separate from external packages
import (
    "fmt"
    "os"

    "go.uber.org/atomic"
    "golang.org/x/sync/errgroup"
)
```

```go
// Good: Full grouping with protos and side-effects
import (
    "fmt"
    "os"

    "github.com/dsnet/compress/flate"
    "golang.org/x/text/encoding"

    foopb "myproj/foo/proto/proto"

    _ "myproj/rpc/protocols/dial"
)
```

### Import Renaming

Avoid renaming imports except to avoid a name collision; good package names
should not require renaming. In the event of collision, **prefer to rename the
most local or project-specific import**.

**Must rename:** collision with other imports, generated protocol buffer packages
(remove underscores, add `pb` suffix).

**May rename:** uninformative names (e.g., `v1`), collision with local variable.

```go
// Good: Proto packages renamed with pb suffix
import (
    foosvcpb "path/to/package/foo_service_go_proto"
)

// Good: urlpkg when url variable is needed
import (
    urlpkg "net/url"
)

func parseEndpoint(url string) (*urlpkg.URL, error) {
    return urlpkg.Parse(url)
}
```

### Blank Imports (`import _`)

Packages that are imported only for their side effects (using `import _ "pkg"`)
should only be imported in the main package of a program, or in tests that
require them.

```go
// Good: Blank import in main package
package main

import (
    _ "time/tzdata"
    _ "image/jpeg"
)
```

### Dot Imports (`import .`)

**Do not** use dot imports. They make programs much harder to read because it is
unclear whether a name like `Quux` is a top-level identifier in the current
package or in an imported package.

**Exception:** The `import .` form can be useful in tests that, due to circular
dependencies, cannot be made part of the package being tested:

```go
package foo_test

import (
	"bar/testutil" // also imports "foo"
	. "foo"
)
```

In this case, the test file cannot be in package `foo` because it uses
`bar/testutil`, which imports `foo`. So the `import .` form lets the file
pretend to be part of package `foo` even though it is not.

**Except for this one case, do not use `import .` in your programs.**

```go
// Bad: Dot import hides origin
import . "foo"
var myThing = Bar() // Where does Bar come from?

// Good: Explicit qualification
import "foo"
var myThing = foo.Bar()
```

---

## Avoid init()

Avoid `init()` where possible. When `init()` is unavoidable, code should:

1. Be completely deterministic, regardless of program environment
2. Avoid depending on ordering or side-effects of other `init()` functions
3. Avoid global/environment state (env vars, working directory, args)
4. Avoid I/O (filesystem, network, system calls)

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

### Exit Once

Prefer to call `os.Exit` or `log.Fatal` **at most once** in `main()`. Extract
business logic into a separate function that returns errors.

```go
// Good: Single exit point with run() pattern
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

## Package Size

### When to Split a Package

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

- Don't split just because a file is long — large files in a focused package are fine
- Don't create packages with only one type or function
- Don't split if it would create circular dependencies
- Avoid splitting internal helpers into a `util` or `internal/helpers` package

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

For larger CLIs, consider libraries like `cobra` or `urfave/cli`. Exit only from `main()`.

---

## See Also

- [go-style-core](../go-style-core/SKILL.md): Core style principles
- [go-naming](../go-naming/SKILL.md): Naming conventions
- [go-error-handling](../go-error-handling/SKILL.md): Error handling patterns
- [go-defensive](../go-defensive/SKILL.md): Defensive coding
- [go-linting](../go-linting/SKILL.md): Linting tools
