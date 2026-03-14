# Agent Skills For Go

AI [Agent Skills](https://agentskills.io/) for writing idiomatic,
production-quality Go code. These modular skills teach AI coding assistants Go
best practices derived from:

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Wiki
  CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)

## Skills Included

| Skill | Description |
|-------|-------------|
| **go-code-review** | Quick-reference checklist for Go code review based on CodeReviewComments |
| **go-concurrency** | Goroutine lifecycle, channels, mutex handling, sync primitives |
| **go-context** | Context usage patterns, parameter placement, proper propagation |
| **go-control-flow** | If with initialization, for loops, range, switch, type switch, blank identifier |
| **go-data-structures** | Allocation (new vs make), arrays, slices, maps, printing, iota constants |
| **go-declarations** | Variable/const/type declarations, initialization, scope, `any` vs `interface{}` |
| **go-defensive** | Interface verification, slice/map copying, time handling, defer for cleanup |
| **go-documentation** | Doc comments, package docs, godoc formatting, examples |
| **go-error-handling** | Error handling patterns, wrapping, sentinel errors, error flow |
| **go-functional-options** | Functional options pattern for constructors and APIs with optional config |
| **go-functions** | Multiple return values, file organization, signature formatting, Printf conventions |
| **go-generics** | When to use type parameters, constraints, type aliases vs definitions |
| **go-interfaces** | Interface design, type assertions, type switches, embedding for composition |
| **go-linting** | Recommended linters and golangci-lint configuration |
| **go-naming** | Naming conventions for packages, functions, methods, variables, constants |
| **go-packages** | Package organization, imports, avoiding util packages |
| **go-performance** | String handling, type conversions, container capacity, string concatenation |
| **go-style-core** | Core principles (clarity, simplicity, consistency) and formatting rules |
| **go-testing** | Table-driven tests, test helpers, subtests, parallel tests, TestMain |

## Quick Install

### Using npx skills (Recommended)

The easiest way to install across **any** AI coding agent. Supports Cursor,
Codex, OpenCode, Cline, GitHub Copilot, Windsurf, Roo Code, and [25+ more
agents](https://github.com/vercel-labs/skills#supported-agents).

```bash
npx skills add cxuu/golang-skills --all
```

### Claude Code

```bash
# Add the marketplace (one time)
/plugin marketplace add cxuu/golang-skills

# Install the skills
/plugin install golang-skills@cxuu-golang-skills
```

### Cursor (Native Remote Rule)

1. Open **Cursor Settings** (Cmd+Shift+J on Mac, Ctrl+Shift+J on Windows/Linux)
2. Navigate to **Rules** → **Add Rule** → **Remote Rule (Github)**
3. Enter: `https://github.com/cxuu/golang-skills`

## How It Works

These skills follow the [Agent Skills open standard](https://agentskills.io/),
which works across multiple AI coding tools. When you're writing Go code:

1. **Automatic activation**: The AI agent loads relevant skills based on context
   (e.g., `go-naming` when you're writing a new function)
2. **On-demand guidance**: Skills provide actionable rules with good/bad code
   examples
3. **Cross-references**: Skills link to each other for comprehensive coverage

## License

This project is licensed under the Apache License, Version 2.0. See the
[LICENSE](LICENSE) file for details.
