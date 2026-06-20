# Release Checklist

Use this checklist before tagging a release.

## Documentation

- Update `CHANGELOG.md` and move relevant entries from `[Unreleased]` to the
  release version.
- Confirm `README.md` install, project structure, license, and provenance
  sections still match the repository.
- Confirm `THIRD_PARTY_NOTICES.md` covers every file under `source/`.

## Validation

Run the release validation commands from the repository root:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1

for skill_dir in skills/*/; do
  npx --yes agentskills-validate@1.0.1 "$skill_dir"
done

(cd evals && go test -count=1 ./...)

golangci-lint config verify --config skills/go-linting/assets/golangci.yml
```

## Versioning

- Update plugin and marketplace version metadata when cutting a release.
- Use a `vX.Y.Z` tag.
- Push the tag and confirm the `Validate Skills` workflow passes for the tag.
