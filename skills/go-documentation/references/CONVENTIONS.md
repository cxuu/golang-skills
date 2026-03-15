# Documentation Conventions Reference

## Parameters and Configuration

> **Advisory**: Document error-prone or non-obvious parameters, not everything.

```go
// Bad: Restates the obvious
// Sprintf formats according to a format specifier and returns the resulting string.
//
// format is the format, and data is the interpolation data.
func Sprintf(format string, data ...any) string

// Good: Documents non-obvious behavior
// Sprintf formats according to a format specifier and returns the resulting string.
//
// The provided data is used to interpolate the format string. If the data does
// not match the expected format verbs or the amount of data does not satisfy
// the format specification, the function will inline warnings about formatting
// errors into the output string.
func Sprintf(format string, data ...any) string
```

---

## Contexts

> **Advisory**: Don't restate implied context behavior; document exceptions.

Context cancellation is implied to interrupt the function and return
`ctx.Err()`. Don't document this.

```go
// Bad: Restates implied behavior
// Run executes the worker's run loop.
//
// The method will process work until the context is cancelled.
func (Worker) Run(ctx context.Context) error

// Good: Just the essential
// Run executes the worker's run loop.
func (Worker) Run(ctx context.Context) error
```

**Document when behavior differs:**

```go
// Good: Non-standard cancellation behavior
// Run executes the worker's run loop.
//
// If the context is cancelled, Run returns a nil error.
func (Worker) Run(ctx context.Context) error

// Good: Special context requirements
// NewReceiver starts receiving messages sent to the specified queue.
// The context should not have a deadline.
func NewReceiver(ctx context.Context) *Receiver
```

---

## Concurrency

> **Advisory**: Document non-obvious thread safety characteristics.

Read-only operations are assumed safe; mutating operations are assumed unsafe.
Don't restate this.

**Document when:**

```go
// Ambiguous operation (looks read-only but mutates internally)
// Lookup returns the data associated with the key from the cache.
//
// This operation is not safe for concurrent use.
func (*Cache) Lookup(key string) (data []byte, ok bool)

// API provides synchronization
// NewFortuneTellerClient returns an *rpc.Client for the FortuneTeller service.
// It is safe for simultaneous use by multiple goroutines.
func NewFortuneTellerClient(cc *rpc.ClientConn) *FortuneTellerClient

// Interface has concurrency requirements
// A Watcher reports the health of some entity (usually a backend service).
//
// Watcher methods are safe for simultaneous use by multiple goroutines.
type Watcher interface {
    Watch(changed chan<- bool) (unwatch func())
    Health() error
}
```

---

## Cleanup

> **Advisory**: Always document explicit cleanup requirements.

```go
// Good:
// NewTicker returns a new Ticker containing a channel that will send the
// current time on the channel after each tick.
//
// Call Stop to release the Ticker's associated resources when done.
func NewTicker(d Duration) *Ticker

// Good: Show how to clean up
// Get issues a GET to the specified URL.
//
// When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
//
//    resp, err := http.Get("http://example.com/")
//    if err != nil {
//        // handle error
//    }
//    defer resp.Body.Close()
//    body, err := io.ReadAll(resp.Body)
func (c *Client) Get(url string) (resp *Response, err error)
```

---

## Errors

> **Advisory**: Document significant error sentinel values and types.

```go
// Good: Document sentinel values
// Read reads up to len(b) bytes from the File and stores them in b.
//
// At end of file, Read returns 0, io.EOF.
func (*File) Read(b []byte) (n int, err error)

// Good: Document error types (include pointer receiver)
// Chdir changes the current working directory to the named directory.
//
// If there is an error, it will be of type *PathError.
func Chdir(dir string) error
```

Noting `*PathError` (not `PathError`) enables correct use of `errors.Is` and
`errors.As`.

For package-wide error conventions, document in the package comment.

---

## Named Result Parameters

> **Advisory**: Use for documentation when types alone aren't clear enough.

```go
// Good: Multiple params of same type
func (n *Node) Children() (left, right *Node, err error)

// Good: Action-oriented name clarifies usage
// The caller must arrange for the returned cancel function to be called.
func WithTimeout(parent Context, d time.Duration) (ctx Context, cancel func())

// Bad: Type already clear, name adds nothing
func (n *Node) Parent1() (node *Node)
func (n *Node) Parent2() (node *Node, err error)

// Good: Type is sufficient
func (n *Node) Parent1() *Node
func (n *Node) Parent2() (*Node, error)
```

Don't name results just to enable naked returns. Clarity > brevity.

---

## Deprecation Notices

> **Advisory**: Use the `// Deprecated:` comment to mark symbols as deprecated.

The `Deprecated:` paragraph must appear in the doc comment immediately before
the symbol. It should state what to use instead.

**Standard format:**

```
// Deprecated: Use NewThing instead.
```

Godoc renders `Deprecated:` comments with special visual styling, making them
easy to spot.

**Function deprecation:**

```go
// EstimateSize returns an approximate byte count.
//
// Deprecated: Use [Size] instead, which returns an exact count.
func EstimateSize(r io.Reader) (int64, error)
```

**Type deprecation:**

```go
// LegacyClient talks to the v1 API.
//
// Deprecated: Use [Client] instead, which supports v2.
type LegacyClient struct{ /* ... */ }
```

**Package deprecation** — add `Deprecated:` to the package doc comment:

```go
// Package old provides the original implementation.
//
// Deprecated: Use package example/new instead.
package old
```

Always suggest a concrete alternative so callers know what to migrate to.

---

## Comment Sentences — Detailed

> **Normative**: Documentation comments must be complete sentences.

- Capitalize the first word, end with punctuation
- Exception: may begin with uncapitalized identifier if clear
- End-of-line comments for struct fields can be phrases:

```go
// Good:
// A Server handles serving quotes from Shakespeare.
type Server struct {
    // BaseDir points to the base directory for Shakespeare's works.
    //
    // Expected structure:
    //   {BaseDir}/manifest.json
    //   {BaseDir}/{name}/{name}-part{number}.txt
    BaseDir string

    WelcomeMessage  string // displayed when user logs in
    ProtocolVersion string // checked against incoming requests
    PageLength      int    // lines per page (optional; default: 20)
}
```

