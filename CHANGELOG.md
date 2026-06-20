# Changelog

All notable changes to this repository are documented here.

## [Unreleased]

## [1.8.0] - 2026-06-20

### Added

- Added repository-level third-party notices for bundled `source/` snapshots.
- Added a Go compatibility policy for version-sensitive standard-library
  guidance, eval harness expectations, and golangci-lint config verification.
- Added a release checklist covering changelog, provenance, compatibility,
  validation, and tagging steps.

### Changed

- Expanded the validation workflow to run on pull requests, pushes to `main`,
  `v*` tags, and manual dispatch.
- Pinned skill validation to `agentskills-validate@1.0.1`.
- Added Go setup, eval tests, and golangci-lint config verification to CI.
- Clarified README provenance and license wording.

### Fixed

- Corrected the README project tree so `evals/` and `source/` are shown as
  top-level directories and `evals/fixtures/` is included.
