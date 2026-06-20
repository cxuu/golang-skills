---
name: go-linting
description: Use when setting up linting for a Go project, configuring golangci-lint, or adding Go checks to a CI/CD pipeline. Also use when starting a new Go project and deciding which linters to enable, even if the user only asks about "code quality" or "static analysis" without mentioning specific linter names. Does not cover code review process (see go-code-review).
allowed-tools: Bash(bash:*)
---

# Go Linting

## Core Principle

More important than any "blessed" set of linters: **lint consistently across a codebase**.

Consistent linting helps catch common issues and establishes a high bar for code quality without being unnecessarily prescriptive.

## Resource Routing

- `scripts/setup-lint.sh` - Run when generating a `.golangci.yml`, validating the first lint pass, or producing JSON metadata.
- `assets/golangci.yml` - Use as the v2 golangci-lint baseline for established projects.

## Setup Procedure

1. Create `.golangci.yml` with `scripts/setup-lint.sh` or copy `assets/golangci.yml`
2. Run `golangci-lint run ./...`
3. If errors appear, fix them category by category (formatting first, then vet, then style)
4. Re-run until clean

After generating `.golangci.yml`, run `golangci-lint config verify --config .golangci.yml`
to verify the configuration schema before relying on lint results.

---

## Minimum Recommended Linters

These linters catch the most common issues while maintaining a high quality bar:

| Linter | Purpose |
|--------|---------|
| [errcheck](https://github.com/kisielk/errcheck) | Ensure errors are handled |
| [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) | Format code and manage imports |
| [revive](https://github.com/mgechev/revive) | Common style mistakes (modern replacement for golint) |
| [govet](https://pkg.go.dev/cmd/vet) | Analyze code for common mistakes |
| [staticcheck](https://staticcheck.dev) | Various static analysis checks |

> **Note**: `revive` is the modern, faster successor to the now-deprecated `golint`.

---

## Lint Runner: golangci-lint

Use [golangci-lint](https://github.com/golangci/golangci-lint) as your lint runner. See the [example .golangci.yml](https://github.com/uber-go/guide/blob/master/.golangci.yml) from uber-go/guide.

---

## Example Configuration

Use `assets/golangci.yml` as the maintained example. It targets
golangci-lint v2 (verified with 2.10.1 on 2026-06-19), keeps `goimports`
under `formatters`, and enables the core linters plus common production
additions.

### Running

```bash
# Install the version this skill's config is verified against
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1

# Run all linters
golangci-lint run

# Run on specific paths
golangci-lint run ./pkg/...
```

---

## Additional Recommended Linters

Beyond the minimum set, consider these for production projects:

| Linter | Purpose | When to enable |
|--------|---------|----------------|
| [gosec](https://github.com/securego/gosec) | Security vulnerability detection | Always for services handling user input |
| [ineffassign](https://github.com/gordonklaus/ineffassign) | Detect ineffectual assignments | Always — catches dead code |
| [misspell](https://github.com/client9/misspell) | Correct common misspellings in comments/strings | Always |
| [gocyclo](https://github.com/fzipp/gocyclo) | Cyclomatic complexity threshold | When functions exceed ~15 complexity |
| [exhaustive](https://github.com/nishanths/exhaustive) | Ensure switch covers all enum values | When using iota enums |
| [bodyclose](https://github.com/timakin/bodyclose) | Detect unclosed HTTP response bodies | Always for HTTP client code |

---

## Nolint Directives

When suppressing a lint finding, always explain why:

```go
//nolint:errcheck // fire-and-forget logging; error is not actionable
_ = logger.Sync()
```

Rules:
- Use `//nolint:lintername` — never bare `//nolint`
- Place the comment on the same line as the finding
- Include a justification after `//`

---

## CI/CD Integration

Run `golangci-lint run ./...` in CI after tests. Pin the golangci-lint version
used by CI so local and release behavior do not drift.

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit
golangci-lint run --new-from-rev=HEAD~1
```

Use `--new-from-rev` to lint only changed code, keeping the feedback loop fast.

---

## Quick Reference

| Task | Command/Action |
|------|----------------|
| Install golangci-lint | `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1` |
| Run linters | `golangci-lint run` |
| Run on path | `golangci-lint run ./pkg/...` |
| Config file | `.golangci.yml` in project root |
| CI integration | Run `golangci-lint run` in pipeline |
| Nolint directives | `//nolint:name // reason` — never bare `//nolint` |
| CI integration | Use `golangci/golangci-lint-action` for GitHub Actions |
| Pre-commit | `golangci-lint run --new-from-rev=HEAD~1` |

### Linter Selection Guidelines

| When you need... | Use |
|------------------|-----|
| Error handling coverage | errcheck |
| Import formatting | goimports |
| Style consistency | revive |
| Bug detection | govet, staticcheck |
| All of the above | golangci-lint with config |

---

## Related Skills

- **Style foundations**: See [go-style-core](../go-style-core/SKILL.md) when resolving style questions that linters enforce (formatting, nesting, naming)
- **Code review**: See [go-code-review](../go-code-review/SKILL.md) when combining linter output with a manual review checklist
- **Error handling**: See [go-error-handling](../go-error-handling/SKILL.md) when errcheck flags unhandled errors and you need to decide how to handle them
- **Testing**: See [go-testing](../go-testing/SKILL.md) when running linters alongside tests in CI pipelines
