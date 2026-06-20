# Rule Ownership Map

Each rule should have one canonical owner. Other skills should route to that
owner with a short pointer instead of repeating a full explanation.

| Rule area | Canonical owner | Route from | Source basis |
|---|---|---|---|
| Interface placement and shape | `go-interfaces` | `go-code-review`, `go-control-flow`, `go-defensive` | Go CodeReviewComments `Interfaces`; Effective Go interface names |
| Compile-time interface assertions | `go-interfaces` | `go-defensive`, `go-control-flow` | Uber `Verify Interface Compliance`; Effective Go blank identifier |
| Context parameter placement and values | `go-context` | `go-concurrency`, `go-code-review`, `go-logging`, `go-testing` | Go CodeReviewComments `Contexts`; Google documentation conventions |
| Goroutine lifetime and synchronization | `go-concurrency` | `go-context`, `go-code-review`, `go-testing` | Go CodeReviewComments `Goroutine Lifetimes`; Uber goroutine guidance |
| Error matching, wrapping, and ownership | `go-error-handling` | `go-code-review`, `go-logging`, `go-defensive` | Uber `Errors`; Go CodeReviewComments `Handle Errors` |
| Log levels and structured logging | `go-logging` | `go-error-handling`, `go-code-review`, `go-context` | Google logging best practices; `log/slog` docs |
| Documentation comments and examples | `go-documentation` | all skills that add exported APIs | Google doc comments; Go CodeReviewComments `Doc Comments` |
| Naming, initialisms, receivers, packages | `go-naming` | `go-packages`, `go-interfaces`, `go-functions` | Effective Go naming; Go CodeReviewComments naming sections; Google naming decisions |
| Pointers to interfaces | `go-functions` | `go-interfaces`, `go-code-review` | Uber `Pointers to Interfaces`; Go CodeReviewComments `Pass Values` |
| Declarations, literals, initialization | `go-declarations` | `go-data-structures`, `go-style-core` | Google declarations decisions; Uber initialization guidance |
| Data structure selection | `go-data-structures` | `go-generics`, `go-performance` | Go CodeReviewComments slices/maps; Google style decisions |
| Functional options vs config structs | `go-functional-options` | `go-functions`, `go-interfaces` | Uber functional options; Google option struct guidance |
| Lint setup and static analysis | `go-linting` | `go-code-review`, `go-style-core` | Uber linting; golangci-lint v2 config schema |
| Benchmarks, profiling, hot-path changes | `go-performance` | `go-data-structures`, `go-functions` | Uber performance guidance; Go testing benchmark docs |
| Table tests, helpers, integration tests | `go-testing` | `go-code-review`, `go-documentation` | Google testing best practices; Uber test tables |
| Package structure, imports, main/run pattern | `go-packages` | `go-code-review`, `go-naming` | Go CodeReviewComments package names/imports; Uber exit-in-main guidance |

## Maintenance Rules

- Add new rule areas here before duplicating guidance in another skill.
- In non-owner skills, keep route text to one or two lines plus a link.
- If sources conflict, record the chosen repository policy in the owner
  reference and link to it from route-only skills.
