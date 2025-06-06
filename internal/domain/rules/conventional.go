// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// conventionalParts represents the parts of a conventional commit.
// This is an internal type used only by the parseConventionalFormat function.
type conventionalParts struct {
	Type        string
	Scope       string
	Breaking    bool
	Description string
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
//
// See https://www.conventionalcommits.org/ for more information.
type ConventionalCommitRule struct {
	allowedTypes     []string
	allowedScopes    []string
	requireScope     bool
	validateBreaking bool
	maxDescLength    int
}

// NewConventionalCommitRule creates a new rule for validating conventional commits from config.
func NewConventionalCommitRule(cfg config.Config) ConventionalCommitRule {
	// Set default types if not configured
	allowedTypes := cfg.Conventional.Types
	if len(allowedTypes) == 0 {
		allowedTypes = []string{
			"feat", "fix", "docs", "style", "refactor", "perf",
			"test", "build", "ci", "chore", "revert",
		}
	}

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
	}
}

// Name returns the name of the rule.
func (r ConventionalCommitRule) Name() string {
	return "ConventionalCommit"
}

// Validate validates a commit against the conventional commit rules.
func (r ConventionalCommitRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	var failures []domain.ValidationError

	// Parse conventional format
	parts, err := parseConventionalFormat(commit.Subject)
	if err != nil {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrInvalidConventionalFormat, "Commit message doesn't follow conventional format").
				WithHelp("Use format: type(scope): description (e.g., 'feat: add login')"))

		return failures
	}

	// Validate type
	if !isValidType(parts.Type, r.allowedTypes) {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrInvalidConventionalType,
				fmt.Sprintf("Invalid type '%s'", parts.Type)).
				WithHelp("Use one of: "+strings.Join(r.allowedTypes, ", ")))
	}

	// Validate scope if required
	if r.requireScope && parts.Scope == "" {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrMissingConventionalScope, "Scope is required but not provided").
				WithHelp("Use format: type(scope): description"))
	}

	// Validate allowed scopes if specified
	if parts.Scope != "" && len(r.allowedScopes) > 0 && !isValidScope(parts.Scope, r.allowedScopes) {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrInvalidConventionalScope,
				fmt.Sprintf("Invalid scope '%s'", parts.Scope)).
				WithHelp("Use one of: "+strings.Join(r.allowedScopes, ", ")))
	}

	// Validate description length
	if r.maxDescLength > 0 && len(parts.Description) > r.maxDescLength {
		failures = append(failures,
			domain.New(r.Name(), domain.ErrConventionalDescTooLong,
				fmt.Sprintf("Description too long (%d > %d)", len(parts.Description), r.maxDescLength)).
				WithHelp(fmt.Sprintf("Keep description under %d characters", r.maxDescLength)))
	}

	return failures
}

// parseConventionalFormat parses a commit subject into conventional commit parts.
func parseConventionalFormat(subject string) (conventionalParts, error) {
	// Conventional commit format: <type>[(scope)][!]: <description>
	// Example: feat(api)!: add new endpoint
	pattern := `^(?P<type>[a-z]+)(?:\((?P<scope>[a-z0-9/-]+)\))?(?P<breaking>!)?:\s?(?P<description>.*)`
	regex := regexp.MustCompile(pattern)

	match := regex.FindStringSubmatch(subject)
	if match == nil {
		return conventionalParts{}, errors.New("subject does not match conventional format")
	}

	// Extract named groups
	groups := make(map[string]string)

	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" && i < len(match) {
			groups[name] = match[i]
		}
	}

	return conventionalParts{
		Type:        groups["type"],
		Scope:       groups["scope"],
		Breaking:    groups["breaking"] == "!",
		Description: groups["description"],
	}, nil
}

// isValidType checks if the commit type is in the list of allowed types.
func isValidType(commitType string, allowedTypes []string) bool {
	// If no allowed types are specified, all types are allowed
	if len(allowedTypes) == 0 {
		return true
	}

	return slices.Contains(allowedTypes, commitType)
}

// isValidScope checks if the commit scope is in the list of allowed scopes.
func isValidScope(scope string, allowedScopes []string) bool {
	// If no allowed scopes are specified, all scopes are allowed
	if len(allowedScopes) == 0 {
		return true
	}

	return slices.Contains(allowedScopes, scope)
}
