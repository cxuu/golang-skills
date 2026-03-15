# Identifier Naming Rules

Detailed rules and examples for naming Go packages, interfaces, receivers,
constants, initialisms, and functions.

## Package Names

> **Normative**: Packages must be lowercase with no underscores.

Package names must be:
- Concise and lowercase only
- No underscores (e.g., `tabwriter` not `tab_writer`)
- Not likely to shadow common variables

```go
// Good: user, oauth2, k8s, tabwriter
// Bad: user_service (underscores), UserService (uppercase), count (shadows var)
```

### Avoid Uninformative Names

> **Advisory**: Don't use generic package names.

Avoid names that tempt users to rename on import: `util`, `common`, `helper`,
`model`, `base`. Prefer specific names: `stringutil`, `httpauth`, `configloader`.

### Import Renaming

When renaming imports, the local name must follow package naming rules:
`import foopb "path/to/foo_go_proto"` (not `foo_pb` with underscore).

---

## Interface Names

> **Advisory**: One-method interfaces use "-er" suffix.

By convention, one-method interfaces are named by the method name plus an `-er`
suffix to construct an agent noun:

```go
// Standard library examples
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Formatter interface { Format(f State, verb rune) }
type CloseNotifier interface { CloseNotify() <-chan bool }
```

Honor canonical method names (`Read`, `Write`, `Close`, `String`) and their
signatures. If your type implements a method with the same meaning as a
well-known type, use the same name—call it `String` not `ToString`.

---

## Receiver Names

> **Normative**: Receivers must be short abbreviations, used consistently.

Receiver variable names must be:
- Short (one or two letters)
- Abbreviations for the type itself
- Consistent across all methods of that type

| Long Name (Bad)             | Better Name              |
|-----------------------------|--------------------------|
| `func (tray Tray)`          | `func (t Tray)`          |
| `func (info *ResearchInfo)` | `func (ri *ResearchInfo)`|
| `func (this *ReportWriter)` | `func (w *ReportWriter)` |
| `func (self *Scanner)`      | `func (s *Scanner)`      |

```go
// Good - consistent short receiver
func (c *Client) Connect() error
func (c *Client) Send(msg []byte) error
func (c *Client) Close() error

// Bad - inconsistent or long receivers
func (client *Client) Connect() error
func (cl *Client) Send(msg []byte) error
func (this *Client) Close() error
```

---

## Constant Names

> **Normative**: Constants use MixedCaps, never ALL_CAPS or K prefix.

```go
// Good
const MaxPacketSize = 512
const defaultTimeout = 30 * time.Second

// Bad
const MAX_PACKET_SIZE = 512    // no snake_case
const kMaxBufferSize = 1024    // no K prefix
```

### Name by Role, Not Value

> **Advisory**: Constants should explain what the value denotes.

```go
// Good - names explain the role
const MaxRetries = 3
const DefaultPort = 8080

// Bad - names just describe the value
const Three = 3
const Port8080 = 8080
```

---

## Initialisms and Acronyms

> **Normative**: Initialisms maintain consistent case throughout.

Initialisms (URL, ID, HTTP, API) should be all uppercase or all lowercase:

| English   | Exported  | Unexported |
|-----------|-----------|------------|
| URL       | `URL`     | `url`      |
| ID        | `ID`      | `id`       |
| HTTP/API  | `HTTP`    | `http`     |
| gRPC/iOS  | `GRPC`/`IOS` | `gRPC`/`iOS` |

```go
// Good: HTTPClient, userID, ParseURL()
// Bad: HttpClient, orderId, ParseUrl()
```

---

## Function and Method Names

### Getters and Setters

> **Advisory**: Don't use `Get` prefix for simple accessors.

If you have a field called `owner` (unexported), the getter should be `Owner()`
(exported), not `GetOwner()`. The setter, if needed, is `SetOwner()`:

```go
// Good
owner := obj.Owner()
if owner != user {
    obj.SetOwner(user)
}

// Bad: c.GetName(), u.GetEmail(), p.GetID()
```

Use `Compute` or `Fetch` for expensive operations:
`db.FetchUser(id)`, `stats.ComputeAverage()`.

### Naming Conventions

> **Advisory**: Use noun-like names for getters, verb-like names for actions.

```go
// Noun-like for returning values
func (c *Config) JobName(key string) string
func (u *User) Permissions() []Permission

// Verb-like for actions
func (c *Config) WriteDetail(w io.Writer) error
```

### Type Suffixes

When functions differ only by type, include type at the end:
`ParseInt()`, `ParseInt64()`, `AppendInt()`, `AppendInt64()`.

For a clear "primary" version, omit the type:
`Marshal()` (primary), `MarshalText()` (variant).
