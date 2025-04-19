# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Best Practices for Cost-Effective Usage

## Build Commands

- Build: `make build/plain`
- Test all: `make test`
- Single test: `go test -v -count=1 ./internal/path/to/package/file_test.go -run TestSpecificFunction`
- Lint: `make quality/golangcilint` or `make quality` (all linters)
- Format: `make quality/tidy`
- Run with quality: `make build`

### Keep Conversations Focused

- Ask specific questions with clear objectives
- Break complex tasks into smaller, well-defined requests
- Start new conversations for unrelated topics

### Leverage Command Line Tools

- Use built-in tools like `ls`, `grep`, `find`, and `cat` for file operations
- Ask Claude to suggest relevant shell commands for tasks instead of custom scripts
- Combine simple commands with pipes (|) for more complex operations

### Avoid Overengineering

- Request simpler solutions first, then expand if needed
- Prefer existing tools over custom implementations
- Ask for small shell scripts instead of complex programs when appropriate
- Don't worry about backward compatibility

### Code Quality

- Always provide solutions that are simple, coherent, concise, and idiomatic
- Fix and update tests alongside code changes
- Prioritize readability and maintainability over clever solutions

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
- Keep it simple and maintainable
- Keep it idiomatic go lang

### Working with Code

- Specify exact file paths within your project directory
- Always mention file extensions (.js, .py, etc.)
- Request relative paths only (no git operations)

### File Management

- When sharing files, include only what's necessary
- Use tools like `head`, `tail`, and `wc` to summarize file contents
- Let Claude know if output should be redirected to files (> or >>)

### Project Understanding

- Always check for and read ARCHITECTURE.md in the project or docs/ directory first
- Reference project documentation when discussing implementation

### Debugging Help

- Include specific error messages
- Use tools like `grep` to isolate relevant log sections
- Share command output with context

### Error Handling

- Request appropriate error handling strategies based on context
- Specify how verbose error messages should be
- Consider logging needs for production vs. debugging

### Documentation

- Ask for inline comments only where truly needed
- Request updates to README or other docs when implementing new features
- Keep documentation concise, practical, and idiomatic
