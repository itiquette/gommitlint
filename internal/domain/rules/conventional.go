// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
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
	name             string
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
		name:             "ConventionalCommit",
		allowedTypes:     allowedTypes,
		allowedScopes:    cfg.Conventional.Scopes,
		requireScope:     cfg.Conventional.RequireScope,
		validateBreaking: cfg.Conventional.AllowBreaking,
		maxDescLength:    maxDescLength,
	}
}

// Name returns the name of the rule.
func (r ConventionalCommitRule) Name() string {
	return r.name
}

// Validate validates a commit against the conventional commit rules.
// This method follows functional programming principles and does not modify the rule's state.
func (r ConventionalCommitRule) Validate(_ context.Context, commit domain.CommitInfo) []domain.ValidationError {
	// Validate conventional commit format
	// Parse subject from commit
	subject := commit.Subject
	if subject == "" && commit.Message != "" {
		subject, _ = domain.SplitCommitMessage(commit.Message)
	}

	// Parse conventional format
	parts, err := parseConventionalFormat(subject)
	if err != nil {
		errors := []domain.ValidationError{
			domain.New(
				"ConventionalCommit",
				domain.ErrInvalidConventionalFormat,
				"Commit message doesn't follow conventional format",
			).WithHelp("type(scope): description (e.g., 'feat: add login')").WithContextMap(map[string]string{
				"subject": subject,
			}),
		}

		return errors
	}

	// Validate type
	if !isValidType(parts.Type, r.allowedTypes) {
		errors := []domain.ValidationError{
			domain.New(
				"ConventionalCommit",
				domain.ErrInvalidConventionalType,
				fmt.Sprintf("Invalid type '%s'; allowed types: %s", parts.Type, strings.Join(r.allowedTypes, ", ")),
			).WithHelp("Use one of: " + strings.Join(r.allowedTypes, ", ")).WithContextMap(map[string]string{
				"type":          parts.Type,
				"allowed_types": strings.Join(r.allowedTypes, ", "),
			}),
		}

		return errors
	}

	// Validate scope if required
	if r.requireScope && parts.Scope == "" {
		errors := []domain.ValidationError{
			domain.New(
				"ConventionalCommit",
				domain.ErrMissingConventionalScope,
				"Scope is required but not provided",
			).WithHelp("type(scope): description"),
		}

		return errors
	}

	// Validate allowed scopes if specified
	if parts.Scope != "" && len(r.allowedScopes) > 0 && !isValidScope(parts.Scope, r.allowedScopes) {
		errors := []domain.ValidationError{
			domain.New(
				"ConventionalCommit",
				domain.ErrInvalidConventionalScope,
				fmt.Sprintf("Invalid scope '%s'; allowed scopes: %s", parts.Scope, strings.Join(r.allowedScopes, ", ")),
			).WithHelp("Use one of: " + strings.Join(r.allowedScopes, ", ")).WithContextMap(map[string]string{
				"scope":          parts.Scope,
				"allowed_scopes": strings.Join(r.allowedScopes, ", "),
			}),
		}

		return errors
	}

	// Validate description length
	if r.maxDescLength > 0 && len(parts.Description) > r.maxDescLength {
		errors := []domain.ValidationError{
			domain.New(
				"ConventionalCommit",
				domain.ErrConventionalDescTooLong,
				fmt.Sprintf("Description too long (%d > %d)", len(parts.Description), r.maxDescLength),
			).WithHelp(fmt.Sprintf("Keep description under %d characters", r.maxDescLength)).WithContextMap(map[string]string{
				"length":     strconv.Itoa(len(parts.Description)),
				"max_length": strconv.Itoa(r.maxDescLength),
			}),
		}

		return errors
	}

	// All validations passed
	return []domain.ValidationError{}
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
