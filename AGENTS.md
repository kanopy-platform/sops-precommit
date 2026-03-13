# AGENTS.md

## Build & Test

Run `make test` to lint and test. See `Makefile` for available targets and variables.

## Architecture

Go CLI tool that validates SOPS encryption in git pre-commit hooks. Accepts changed files via args or stdin, then decrypts each to verify encryption.

- **Entry point:** `cmd/sops-precommit/main.go` → `cli.NewRootCommand().Execute()`
- **Core logic:** `internal/cli/cli.go` — if no `.sops.yaml` config exists, exits early with success. Otherwise filters files to those matching `creation_rules` and validates decryption.
- **Testability:** `decrypter` and `sopsRuleMatcher` interfaces abstract the SOPS library. Tests use `decryptmock`.
