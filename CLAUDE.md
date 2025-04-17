# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `make build/plain`
- Test all: `make test`
- Single test: `go test -v -count=1 ./internal/path/to/package/file_test.go -run TestSpecificFunction`
- Lint: `make quality/golangcilint` or `make quality` (all linters)
- Format: `make quality/tidy`
- Run with quality: `make build`

## Code Style
- Follow Go standard formatting (`go fmt`)
- Imports: stdlib first, external next, internal last (with blank lines between groups)
- Error handling:
  - Use `fmt.Errorf("context: %w", err)` pattern
  - Custom errors: `internal.NewValidationError(err, map[string]string{"key": "value"})`
  - Validation errors: `model.NewValidationError("RuleName", "error_code", "message")`
- Testing: Use table-driven tests with testify/require and assert
- Documentation: Full godoc comments for all exported functions/types
- Naming: PascalCase for exported, camelCase for non-exported identifiers

## Commit Style
- Follow conventional commits format: `type(scope): description`
- Commits require sign-off: `git commit --signoff -m "..."`
- GPG signing recommended: `git commit --signoff --gpg-sign -m "..."`
- Commit linting: `make quality/commit`