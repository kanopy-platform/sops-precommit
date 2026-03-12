# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make test                          # Lint (golangci-lint) + run all tests
go test -cover ./...               # Tests only, no linting
go test -cover -run=TestName ./... # Run a single test
go install ./cmd/sops-precommit/   # Build and install the binary
```

The `make test` target runs `golangci-lint` first, then `go test`. You can scope both with `PKG=./internal/cli/...` and `RUN=TestDecryptFiles`.

## Architecture

This is a Go CLI tool that validates SOPS encryption in git pre-commit hooks. It accepts a list of changed files (via args or stdin pipe) and attempts to decrypt each one to verify they are properly encrypted.

**Entry point:** `cmd/sops-precommit/main.go` - thin wrapper that calls `cli.NewRootCommand().Execute()`

**Core logic:** `internal/cli/cli.go` - all business logic lives here:
- `runE` - parses file list from args or stdin, loads `.sops.yaml` config, filters files, decrypts
- `getFilteredFiles` - if a `.sops.yaml` exists, filters files to only those matching `creation_rules`; otherwise passes all files through
- `decryptFiles` - validates each file by attempting SOPS decryption; collects all errors before returning

**Key interfaces:** `decrypter` and `sopsRuleMatcher` abstract the SOPS library for testability. Tests use `decryptmock` to avoid real crypto operations.

**Config:** Uses cobra/viper. Log level configurable via `--log-level` flag or `SOPS_LOG_LEVEL` env var.
