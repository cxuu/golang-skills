# Import Organization

Detailed rules and examples for organizing Go imports.

## Import Grouping

Imports are organized in groups, with blank lines between them. The standard
library packages are always in the first group.

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

## Import Renaming

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

## Blank Imports (`import _`)

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

## Dot Imports (`import .`)

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
