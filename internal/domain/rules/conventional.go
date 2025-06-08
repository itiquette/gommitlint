// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// conventionalParts represents the parts of a conventional commit.
// This is an internal type used only by the parseConventionalFormat function.
type conventionalParts struct {
	Type        string
	Scopes      []string // Support for multiple scopes: feat(ui,api): description
	Breaking    bool
	Description string
	RawScope    string // Original scope string for error reporting
}

// ConventionalCommitRule validates that commit messages follow the Conventional Commits specification.
//
// Conventional Commits format provides a standard way to structure commit messages,
// making them machine-readable and establishing a clear connection between commits
// and project features or fixes. This rule enforces that commit messages follow this format.
//
// The standard format is:
//
//	<type>[optional scope][optional !]: <description>
//
// Example: feat(auth): add login functionality
// Example with breaking change: feat(api)!: change auth endpoint structure
// Example with multiple scopes: feat(ui,api): add login functionality
//
// See https://www.conventionalcommits.org/ for more information.
type ConventionalCommitRule struct {
	allowedTypes     []string
	allowedScopes    []string
	requireScope     bool
	validateBreaking bool
	maxDescLength    int
	allowMultiScope  bool // Enable multi-scope support
}

// NewConventionalCommitRule creates a new rule for validating conventional commits from config.
func NewConventionalCommitRule(cfg config.Config) ConventionalCommitRule {
	// Set default types if not configured
	// Per Conventional Commits spec: "any casing may be used, but it's best to be consistent"
	// When no types are explicitly configured, we validate against semantic meaning
	// but allow any case variation of the standard types
	allowedTypes := cfg.Conventional.Types

	// Set max description length with default
	maxDescLength := cfg.Conventional.MaxDescriptionLength
	if maxDescLength <= 0 {
		maxDescLength = 72
	}

	return ConventionalCommitRule{
		allowedTypes:     allowedTypes,
		allowedScopes:    cfg.Conventional.Scopes,
		requireScope:     cfg.Conventional.RequireScope,
		validateBreaking: cfg.Conventional.AllowBreaking,
		maxDescLength:    maxDescLength,
		allowMultiScope:  true, // Enable multi-scope support by default
	}
}

// Name returns the name of the rule.
func (r ConventionalCommitRule) Name() string {
	return "ConventionalCommit"
}

// Validate validates a commit against the conventional commit rules.
func (r ConventionalCommitRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	var failures []domain.ValidationError

	// Parse conventional format with strict spacing if enabled
	parts, err := r.parseConventionalFormat(commit.Subject)
	if err != nil {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrInvalidConventionalFormat, "Must follow format: type(scope): description").
				WithContextMap(map[string]string{
					"actual":   commit.Subject,
					"expected": "type(scope): description",
				}).
				WithHelp("Use format: type(scope): description (e.g., 'feat: add login')"))

		return failures
	}

	// Always validate spacing format (spec requires exactly one space after colon)
	spacingErrors := r.validateSpacing(commit.Subject, parts)
	failures = append(failures, spacingErrors...)

	// Validate description content
	descriptionErrors := r.validateDescription(parts)
	failures = append(failures, descriptionErrors...)

	// Validate type - enforce case-sensitive validation per conventional commit spec
	if !isValidType(parts.Type, r.allowedTypes) {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrInvalidConventionalType,
				fmt.Sprintf("Invalid type '%s'", parts.Type)).
				WithContextMap(map[string]string{
					"actual":   parts.Type,
					"expected": strings.Join(r.allowedTypes, ", "),
				}).
				WithHelp("Use one of: "+strings.Join(r.allowedTypes, ", ")))
	}

	// Validate scope requirements
	scopeErrors := r.validateScopes(parts)
	failures = append(failures, scopeErrors...)

	// Validate description length
	if r.maxDescLength > 0 && len(parts.Description) > r.maxDescLength {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrConventionalDescTooLong,
				fmt.Sprintf("Description too long (%d > %d)", len(parts.Description), r.maxDescLength)).
				WithContextMap(map[string]string{
					"actual":   strconv.Itoa(len(parts.Description)),
					"expected": fmt.Sprintf("max %d", r.maxDescLength),
				}).
				WithHelp(fmt.Sprintf("Keep description under %d characters", r.maxDescLength)))
	}

	return failures
}

// parseConventionalFormat parses a commit subject into conventional commit parts.
func (r ConventionalCommitRule) parseConventionalFormat(subject string) (conventionalParts, error) {
	// Use shared conventional commit parser for consistency
	parsed := domain.ParseConventionalCommit(subject)
	if !parsed.IsValid {
		return conventionalParts{}, errors.New("subject does not match conventional format")
	}

	// Handle multi-scope configuration
	scopes := parsed.Scopes
	if !r.allowMultiScope && len(parsed.Scopes) > 1 {
		// If multi-scope is disabled but multiple scopes found, use only the first one
		scopes = []string{parsed.Scopes[0]}
	}

	return conventionalParts{
		Type:        parsed.Type,
		Scopes:      scopes,
		Breaking:    parsed.Breaking,
		Description: parsed.Description,
		RawScope:    parsed.RawScope,
	}, nil
}

// isValidType checks if the commit type is in the list of allowed types.
// When no explicit types are configured, validates against standard conventional commit types
// with case-sensitive matching per the specification.
func isValidType(commitType string, allowedTypes []string) bool {
	// If no allowed types are specified, validate against standard types case-sensitively
	if len(allowedTypes) == 0 {
		standardTypes := []string{
			"feat", "fix", "docs", "style", "refactor", "perf",
			"test", "build", "ci", "chore", "revert",
		}

		return slices.Contains(standardTypes, commitType)
	}

	// When explicit types are configured, match exactly (case-sensitive)
	return slices.Contains(allowedTypes, commitType)
}

// validateSpacing validates that spacing after colon follows the specification.
// The Conventional Commits spec requires exactly one space after the colon.
func (r ConventionalCommitRule) validateSpacing(subject string, _ conventionalParts) []domain.ValidationError {
	var failures []domain.ValidationError

	// Look for the colon and check what follows
	colonIndex := strings.Index(subject, ":")
	if colonIndex != -1 && colonIndex < len(subject)-1 {
		afterColon := subject[colonIndex+1:]

		// Must be exactly one space followed by non-whitespace content
		if len(afterColon) == 0 || afterColon[0] != ' ' {
			failures = append(failures,
				domain.New(r.Name(), domain.ErrInvalidSpacing, "Missing space after colon").
					WithContextMap(map[string]string{
						"actual":   subject,
						"expected": "type: description",
					}).
					WithHelp("Add exactly one space after the colon"))
		} else if len(afterColon) > 1 && (afterColon[1] == ' ' || afterColon[1] == '\t' || afterColon[1] == '\n' || afterColon[1] == '\r') {
			// Invalid whitespace after the required single space
			failures = append(failures,
				domain.New(r.Name(), domain.ErrInvalidSpacing, "Invalid whitespace after colon").
					WithContextMap(map[string]string{
						"actual":   subject,
						"expected": "type: description",
					}).
					WithHelp("Use exactly one space after the colon followed by description text"))
		}
	}

	return failures
}

// validateDescription validates the description content.
func (r ConventionalCommitRule) validateDescription(parts conventionalParts) []domain.ValidationError {
	var failures []domain.ValidationError

	// Enhanced empty description detection
	trimmedDesc := strings.TrimSpace(parts.Description)
	if trimmedDesc == "" {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrEmptyConventionalDesc, "Description cannot be empty").
				WithContextMap(map[string]string{
					"actual":   "empty",
					"expected": "meaningful description",
				}).
				WithHelp("Add a meaningful description explaining what the commit does"))
	}

	return failures
}

// validateScopes validates scope requirements and allowed scopes.
func (r ConventionalCommitRule) validateScopes(parts conventionalParts) []domain.ValidationError {
	var failures []domain.ValidationError

	// Check if scope is required but missing
	if r.requireScope && len(parts.Scopes) == 0 {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrMissingConventionalScope, "Scope is required but not provided").
				WithContextMap(map[string]string{
					"actual":   fmt.Sprintf("%s: %s", parts.Type, parts.Description),
					"expected": fmt.Sprintf("%s(scope): %s", parts.Type, parts.Description),
				}).
				WithHelp("Use format: type(scope): description"))
	}

	// Validate each scope against allowed scopes if specified
	if len(parts.Scopes) > 0 && len(r.allowedScopes) > 0 {
		for _, scope := range parts.Scopes {
			if !isValidScope(scope, r.allowedScopes) {
				failures = append(failures,
					domain.New(r.Name(), domain.ErrInvalidConventionalScope,
						fmt.Sprintf("Invalid scope '%s'", scope)).
						WithContextMap(map[string]string{
							"actual":   scope,
							"expected": strings.Join(r.allowedScopes, ", "),
						}).
						WithHelp("Use one of: "+strings.Join(r.allowedScopes, ", ")))
			}
		}
	}

	// Validate multi-scope format if enabled and scopes contain commas (regardless of scope restrictions)
	if r.allowMultiScope && parts.RawScope != "" && strings.Contains(parts.RawScope, ",") {
		// Check for proper comma separation format
		if !isValidMultiScopeFormat(parts.RawScope) {
			failures = append(failures,
				domain.New(r.Name(), domain.ErrInvalidMultiScope, "Invalid multi-scope format").
					WithContextMap(map[string]string{
						"actual":   parts.RawScope,
						"expected": "scope1,scope2",
					}).
					WithHelp("Use comma-separated scopes without spaces: (scope1,scope2)"))
		}
	}

	return failures
}

// isValidScope checks if the commit scope is in the list of allowed scopes.
func isValidScope(scope string, allowedScopes []string) bool {
	// If no allowed scopes are specified, all scopes are allowed
	if len(allowedScopes) == 0 {
		return true
	}

	return slices.Contains(allowedScopes, scope)
}

// isValidMultiScopeFormat checks if multi-scope format is valid.
func isValidMultiScopeFormat(rawScope string) bool {
	// Valid: "ui,api", "frontend,backend"
	// Invalid: "ui, api", "ui ,api", "ui , api"
	// Should not have spaces around commas
	if strings.Contains(rawScope, " ,") || strings.Contains(rawScope, ", ") {
		return false
	}

	// Should not start or end with comma
	if strings.HasPrefix(rawScope, ",") || strings.HasSuffix(rawScope, ",") {
		return false
	}

	// Should not have consecutive commas
	if strings.Contains(rawScope, ",,") {
		return false
	}

	return true
}

// Help returns context-aware help text for conventional commits.
func (r ConventionalCommitRule) Help() string {
	help := "Use conventional commit format: type(scope): description"

	// Show allowed types if explicitly configured
	if len(r.allowedTypes) > 0 {
		help += "\nValid types: " + strings.Join(r.allowedTypes, ", ")
	} else {
		help += "\nStandard types (any case): feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert"
	}

	// Show allowed scopes if configured
	if len(r.allowedScopes) > 0 {
		help += "\nValid scopes: " + strings.Join(r.allowedScopes, ", ")
	}

	// Multi-scope information
	if r.allowMultiScope {
		help += "\nMulti-scope format: type(scope1,scope2): description"
	}

	// Breaking change information
	if r.validateBreaking {
		help += "\nBreaking changes: type(scope)!: description"
	}

	// Description length information
	if r.maxDescLength > 0 {
		help += fmt.Sprintf("\nMax description length: %d characters", r.maxDescLength)
	}

	// Spacing requirement
	help += "\nSpacing: exactly one space after colon (required by spec)"

	return help
}
