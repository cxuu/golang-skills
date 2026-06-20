---
name: go-naming
description: Use when naming any Go identifier — packages, types, functions, methods, variables, constants, or receivers — to ensure idiomatic, clear names. Also use when a user is creating new types, packages, or exported APIs, even if they don't explicitly ask about naming conventions. Does not cover package organization (see go-packages).
allowed-tools: Bash(bash:*)
---

# Go Naming Conventions

## Resource Routing

- `scripts/check-naming.sh` - Run when checking SCREAMING_SNAKE_CASE constants, Get-prefixed getters, generic package names, or receivers named `this`/`self`.
- `references/IDENTIFIERS.md` - Read when choosing names for initialisms, exported identifiers, or package-level symbols.
- `references/REPETITION.md` - Read when names repeat package, receiver, type, or local context.
- `references/VARIABLES.md` - Read when choosing local variable names, receiver names, or loop identifiers.

## Core Principle

Names should:
- Not feel repetitive when used
- Take context into consideration
- Not repeat concepts that are already clear

Naming is more art than science—Go names tend to be shorter than in other
languages.

---

## Naming Decision Flow

```
What are you naming?
├─ Package       → Short, lowercase, singular noun (no underscores, no mixedCaps)
├─ Interface     → Method name + "-er" suffix when single-method (Reader, Writer)
├─ Receiver      → 1-2 letter abbreviation of type (c for Client); consistent across methods
├─ Constant      → MixedCaps; use iota for enums; no ALL_CAPS
├─ Exported func → Verb or verb-phrase in MixedCaps; no Get prefix for getters
├─ Variable      → Length proportional to scope distance
│                  ├─ Tiny scope (1-7 lines) → single letter (i, n, r)
│                  ├─ Medium scope           → short word (count, buf)
│                  └─ Package-level / wide   → descriptive (userAccountCount)
└─ Any name      → Check: does it repeat package name or context? If yes, shorten it
```

---

## MixedCaps (Required)

> **Normative**: All Go identifiers must use MixedCaps.

Underscores are allowed only in: test functions (`TestFoo_InvalidInput`),
generated code, and OS/cgo interop.

---

## Package Names

> **Normative**: Packages must be lowercase with no underscores.

Short, lowercase, singular nouns. Avoid generic names like `util`, `common`,
`helper` — prefer specific names: `stringutil`, `httpauth`, `configloader`.

```go
// Good: user, oauth2, tabwriter
// Bad:  user_service, UserService, count (shadows var)
```

---

## Interface Names

> **Advisory**: One-method interfaces use "-er" suffix.

Name one-method interfaces by the method plus `-er`: `Reader`, `Writer`,
`Formatter`. Honor canonical method names (`Read`, `Write`, `Close`, `String`)
and their signatures.

---

## Receiver Names

> **Normative**: Receivers must be short abbreviations, used consistently.

One or two letters abbreviating the type, consistent across all methods:
`func (c *Client) Connect()`, `func (c *Client) Send()`.
Never use `this` or `self`.

---

## Constant Names

> **Normative**: Constants use MixedCaps, never ALL_CAPS or K prefix.

Name constants by role, not value: `MaxRetries` not `Three`,
`DefaultPort` not `Port8080`.

```go
const MaxPacketSize = 512
const defaultTimeout = 30 * time.Second
```

---

## Initialisms and Acronyms

> **Normative**: Initialisms maintain consistent case throughout.

Initialisms (URL, ID, HTTP, API) must be all uppercase or all lowercase:
`HTTPClient`, `userID`, `ParseURL()` — not `HttpClient`, `orderId`, `ParseUrl()`.

---

## Function and Method Names

> **Advisory**: No `Get` prefix for simple accessors; use verb-like names for actions.

Getter for field `owner` is `Owner()`, not `GetOwner()`. Setter is
`SetOwner()`. Use `Compute` or `Fetch` for expensive operations.

When functions differ only by type, include type at the end:
`ParseInt()`, `ParseInt64()`.

---

## Variable Names

Variable naming balances brevity with clarity. Key principles:

- **Scope-based length**: Short names (`i`, `v`) for small scopes; longer,
  descriptive names for larger scopes
- **Single-letter conventions**: Use familiar patterns (`i` for index,
  `r`/`w` for reader/writer)
- **Avoid type in name**: Use `users` not `userSlice`, `name` not `nameString`
- **Prefix unexported globals**: Use `_` prefix for package-level unexported
  vars/consts to prevent shadowing

```go
for i, v := range items { ... }           // small scope
pendingOrders := filterPending(orders)    // larger scope
const _defaultPort = 8080                 // unexported global
```

---

## Avoiding Repetition

Go names should not feel repetitive when used. Consider the full context:

- **Package + symbol**: `widget.New()` not `widget.NewWidget()`
- **Receiver + method**: `p.Name()` not `p.ProjectName()`
- **Context + type**: In package `sqldb`, use `Connection` not `DBConnection`

---

## Avoid Built-In Names

Never shadow Go's predeclared identifiers (`error`, `string`, `len`, `cap`,
`append`, `copy`, `new`, `make`, etc.) as variable, parameter, or type names.

**For detailed guidance**: See `go-declarations` — "Avoid Using Built-In Names"
section.

---

## Quick Reference

| Element | Rule | Example |
|---------|------|---------|
| Package | lowercase, no underscores | `package httputil` |
| Exported | MixedCaps, starts uppercase | `func ParseURL()` |
| Unexported | mixedCaps, starts lowercase | `func parseURL()` |
| Receiver | 1-2 letter abbreviation | `func (c *Client)` |
| Constant | MixedCaps, never ALL_CAPS | `const MaxSize = 100` |
| Initialism | consistent case | `userID`, `XMLAPI` |
| Variable | length ~ scope size | `i` (small), `userCount` (large) |
| Built-in names | Never shadow predeclared identifiers | See `go-declarations` |

> **Validation**: After renaming identifiers, run `bash scripts/check-naming.sh` to verify no naming anti-patterns remain. Then run `go build ./...` to confirm the rename didn't break anything.

## Related Skills

- **Interface naming**: See [go-interfaces](../go-interfaces/SKILL.md) when naming interfaces with the `-er` suffix or choosing receiver types
- **Package naming**: See [go-packages](../go-packages/SKILL.md) when naming packages, avoiding `util`/`common`, or resolving import collisions
- **Error naming**: See [go-error-handling](../go-error-handling/SKILL.md) when naming sentinel errors (`ErrFoo`) or custom error types
- **Declaration scope**: See [go-declarations](../go-declarations/SKILL.md) when variable name length depends on scope or when avoiding built-in shadowing
- **Style principles**: See [go-style-core](../go-style-core/SKILL.md) when balancing clarity vs concision in identifier names
