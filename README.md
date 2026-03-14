# Agent Skills For Go

AI [Agent Skills](https://agentskills.io/) for writing idiomatic,
production-quality Go code. These modular skills teach AI coding assistants Go
best practices derived from:

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Wiki
  CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)

Skills are tuned following
[agentskills.io best practices](https://agentskills.io/skill-creation/best-practices):
content the agent already knows is omitted, procedural decision trees guide
multi-step tasks, and detailed reference material loads on demand via
progressive disclosure.

## Skills Included

| Skill | Description |
|-------|-------------|
| **go-code-review** | Review procedure, checklist, automated pre-check with golangci-lint |
| **go-concurrency** | Goroutine lifecycle, channels, mutex handling, sync primitives |
| **go-context** | Context usage patterns, parameter placement, proper propagation |
| **go-control-flow** | Guard clauses, variable shadowing, expression-less switch, break-with-labels |
| **go-data-structures** | Composite literals, slice gotchas, nil vs empty slice, iota enums, copying |
| **go-declarations** | Variable/const/type declarations, initialization, scope reduction |
| **go-defensive** | Slice/map copying at boundaries, time handling, defer, panic/recover |
| **go-documentation** | Doc comments, package docs, godoc formatting, examples |
| **go-error-handling** | Error strategy decision tree, wrapping (%v vs %w), sentinels, error flow |
| **go-functional-options** | Functional options pattern for constructors with optional config |
| **go-functions** | File ordering, signature formatting, naked parameters, Printf conventions |
| **go-generics** | When to use generics (decision flow), type aliases vs definitions |
| **go-interfaces** | Accept interfaces/return concrete, embedding, satisfaction checks, type switch |
| **go-linting** | Recommended linters and golangci-lint configuration |
| **go-naming** | Naming conventions for packages, functions, methods, variables, constants |
| **go-packages** | Package organization, imports, avoiding util packages |
| **go-performance** | String handling, type conversions, container capacity hints |
| **go-style-core** | Clarity, simplicity, least mechanism, consistency, formatting rules |
| **go-testing** | Table-driven tests, test helpers, test double naming, subtests, TestMain |

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
2. **Procedural guidance**: Decision trees and step-by-step procedures for
   multi-step tasks like code review and error strategy selection
3. **Progressive disclosure**: Core rules load immediately; detailed reference
   material (`references/` files) loads on demand when specific situations arise
4. **Cross-references**: Skills link to each other for comprehensive coverage

## License

This project is licensed under the Apache License, Version 2.0. See the
[LICENSE](LICENSE) file for details.
