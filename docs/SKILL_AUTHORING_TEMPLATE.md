# Skill Authoring Template

Use this structure for every Go skill unless a section is genuinely empty.
Keep core `SKILL.md` files decision-oriented; move long examples, tables, and
edge-case catalogs into `references/`, reusable commands into `scripts/`, and
copyable files into `assets/`.

Before duplicating a rule in more than one skill, check
[`RULE_OWNERSHIP.md`](RULE_OWNERSHIP.md) and route non-owner skills to the
canonical owner.

## Frontmatter

Use trigger-critical fields plus runtime-required tool grants:

```yaml
---
name: go-example
description: Use when ...
allowed-tools: Bash(bash:*) # only when the skill bundles scripts that need it
---
```

Keep license, compatibility, provenance, source authority, and validation
policy in the body or references. Keep `allowed-tools` in frontmatter for
script-backed skills when the target runtime uses it to pre-authorize bundled
commands.

## Body Structure

```md
# Skill Title

One short paragraph stating the core job.

## Resource Routing

- `references/FILE.md` - Read when ...
- `scripts/tool.sh` - Run when ...
- `assets/file.ext` - Use when ...

## Domain-Specific Guidance

Use headings that match the skill's domain. Keep them decision-oriented:
short rules, routing trees, checklists, and small examples.

## Related Skills

- `go-other-skill` - Route when ...
```

## Required Conformance

- Keep `SKILL.md` frontmatter to `name`, `description`, and runtime-required
  `allowed-tools` grants for script-backed skills.
- Include exactly one `## Resource Routing` section, and list every bundled
  file under `references/`, `scripts/`, and `assets`.
- Include a `## Related Skills` section for handoffs to owner skills.
- Keep core files at or below 225 lines; use references for long examples,
  edge cases, and source-sensitive details.
- Include validation guidance near the relevant command or rule when a
  deterministic check exists.

## Reference Headers

Start reference files with compact provenance when source authority matters:

```md
> Sources: source/path.md; official docs URL
> Authority: normative | advisory | historical | project policy
> Minimum Go: 1.xx, if version-sensitive
> Last verified: YYYY-MM-DD
```

Use the authority labels this way:

- `normative`: canonical Go or project-required guidance.
- `advisory`: style guidance that may vary by codebase.
- `historical`: useful context that may not reflect modern Go.
- `project policy`: this repository's chosen rule where sources differ.

## Long References

Any reference over 200 lines must include a `## Contents` section near the top.
Prefer splitting only when a TOC still leaves unrelated subtopics hard to route.
