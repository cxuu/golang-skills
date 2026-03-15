# Agent Skills For Go

AI [Agent Skills](https://agentskills.io/) for writing idiomatic,
production-quality Go code. 19 modular skills teach AI coding assistants Go
best practices derived from:

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Wiki CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)

Skills are tuned following
[agentskills.io best practices](https://agentskills.io/skill-creation/best-practices):
content the agent already knows is omitted, procedural decision trees guide
multi-step tasks, 46 reference files load on demand via progressive disclosure,
8 bundled scripts automate common checks, and 4 asset templates ensure
consistent output.

## Skills Included

| Skill | Description |
|-------|-------------|
| **go-code-review** | Systematic checklist for reviewing Go code and PR submissions |
| **go-concurrency** | Goroutine lifecycle, channels, mutexes, parallelization, thread-safety |
| **go-context** | Context.Context placement, cancellation, deadlines, request-scoped data |
| **go-control-flow** | Idiomatic conditionals, loops, switch/break behavior, guard clauses |
| **go-data-structures** | Slices, maps, arrays — allocation with new vs make, append, copying |
| **go-declarations** | Variable/const/type declarations, var vs :=, iota enums, shadowing |
| **go-defensive** | API boundary hardening, defer cleanup, Must functions, time handling |
| **go-documentation** | Doc comments, package docs, godoc formatting, runnable examples |
| **go-error-handling** | Error strategy decisions, wrapping (%v vs %w), sentinels, logging patterns |
| **go-functional-options** | Functional options pattern for constructors with optional config |
| **go-functions** | Function ordering, signature formatting, Printf verbs, Stringer interface |
| **go-generics** | When to use generics, constraints, common pitfalls, type aliases |
| **go-interfaces** | Interface design, abstractions, embedding, "accept interfaces return structs" |
| **go-linting** | Linters, golangci-lint setup, nolint directives, CI/CD integration |
| **go-naming** | Naming decision flow for packages, types, functions, variables, receivers |
| **go-packages** | Package organization, imports, package size, CLI/flag patterns |
| **go-performance** | String optimization, capacity hints, benchmarking, strconv over fmt |
| **go-style-core** | Formatting, nesting reduction, style principles, fallback style guide |
| **go-testing** | Table-driven tests, subtests, test helpers, assertions, test organization |

## Bundled Scripts

8 scripts automate common Go checks. Each supports `--help` and uses meaningful
exit codes (0 = clean, 1 = issues found).

| Script | Skill | Purpose |
|--------|-------|---------|
| `pre-review.sh` | go-code-review | Run gofmt + go vet + golangci-lint before review |
| `check-naming.sh` | go-naming | Detect SCREAMING_SNAKE, Get-prefixed getters, bad package names |
| `check-docs.sh` | go-documentation | Find exported symbols missing doc comments |
| `check-errors.sh` | go-error-handling | Catch bare returns, string comparison on errors, log-and-return |
| `check-interface-compliance.sh` | go-interfaces | Find interfaces missing compile-time verification |
| `bench-compare.sh` | go-performance | Run benchmarks with optional benchstat comparison |
| `setup-lint.sh` | go-linting | Generate .golangci.yml with recommended linters |
| `gen-table-test.sh` | go-testing | Scaffold a table-driven test file |

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
3. **Progressive disclosure**: Core rules load immediately; 46 reference files
   load on demand when specific situations arise
4. **Automation**: 8 bundled scripts handle repetitive checks so the agent
   focuses on higher-level guidance
5. **Conditional cross-references**: Skills link to each other with "when"
   conditions to avoid unnecessary context loading

## Project Structure

```
skills/
├── go-*/
│   ├── SKILL.md          # Core rules (< 210 lines each)
│   ├── references/       # Detailed guidance, loaded on demand
│   ├── scripts/          # Automation scripts (8 skills)
│   └── assets/           # Output templates (4 skills)
├── evals/
│   └── evals.json        # 45 trigger evals + 12 quality evals
└── source/               # Original style guide sources
```

## License

This project is licensed under the Apache License, Version 2.0. See the
[LICENSE](LICENSE) file for details.
