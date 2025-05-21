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

	"github.com/itiquette/gommitlint/internal/common/contextx"
	commonSlices "github.com/itiquette/gommitlint/internal/common/slices"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
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

// ConventionalCommitOption is a function that configures a ConventionalCommitRule.
type ConventionalCommitOption func(ConventionalCommitRule) ConventionalCommitRule

// WithAllowedTypes sets the allowed commit types.
func WithAllowedTypes(types []string) ConventionalCommitOption {
	return func(r ConventionalCommitRule) ConventionalCommitRule {
		newRule := r
		newRule.allowedTypes = deepCopyStringSlice(types)

		return newRule
	}
}

// WithAllowedScopes sets the allowed commit scopes.
func WithAllowedScopes(scopes []string) ConventionalCommitOption {
	return func(r ConventionalCommitRule) ConventionalCommitRule {
		newRule := r
		newRule.allowedScopes = deepCopyStringSlice(scopes)

		return newRule
	}
}

// WithRequiredScope makes the scope mandatory in commit messages.
func WithRequiredScope() ConventionalCommitOption {
	return func(r ConventionalCommitRule) ConventionalCommitRule {
		newRule := r
		newRule.requireScope = true

		return newRule
	}
}

// WithBreakingChangeValidation enables validation of the breaking change marker.
func WithBreakingChangeValidation() ConventionalCommitOption {
	return func(r ConventionalCommitRule) ConventionalCommitRule {
		newRule := r
		newRule.validateBreaking = true

		return newRule
	}
}

// WithMaxDescLength sets the maximum description length.
func WithMaxDescLength(maxLength int) ConventionalCommitOption {
	return func(r ConventionalCommitRule) ConventionalCommitRule {
		newRule := r
		// Skip if the value is 0 or negative
		if maxLength > 0 {
			newRule.maxDescLength = maxLength
		}

		return newRule
	}
}

// WithRequireScope sets whether a scope is required in commit messages.
func WithRequireScope(require bool) ConventionalCommitOption {
	return func(r ConventionalCommitRule) ConventionalCommitRule {
		newRule := r
		newRule.requireScope = require

		return newRule
	}
}

// NewConventionalCommitRule creates a new rule with the specified options.
func NewConventionalCommitRule(options ...ConventionalCommitOption) ConventionalCommitRule {
	// Create initial rule with default values
	rule := ConventionalCommitRule{
		name: "ConventionalCommit",
		allowedTypes: []string{
			"feat", "fix", "docs", "style", "refactor", "perf",
			"test", "build", "ci", "chore", "revert",
		},
		allowedScopes:    []string{}, // Empty means all scopes are allowed
		requireScope:     false,      // Default to not requiring scope
		validateBreaking: false,      // Default to not validating breaking changes
		maxDescLength:    72,         // Default max length for description
	}

	// Apply options using Reduce for a more functional approach
	return commonSlices.Reduce(
		options,
		rule,
		func(currentRule ConventionalCommitRule, option ConventionalCommitOption) ConventionalCommitRule {
			return option(currentRule)
		},
	)
}

// Name returns the name of the rule.
func (r ConventionalCommitRule) Name() string {
	return r.name
}

// WithContext implements the ConfigurableRule interface for ConventionalCommitRule.
// It returns a new rule with configuration from the provided context.
func (r ConventionalCommitRule) WithContext(ctx context.Context) domain.Rule {
	// Get configuration directly from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return r
	}

	// Create a copy of the rule
	result := r

	// Only override settings if they are specified in the context configuration
	if types := cfg.GetStringSlice("conventional.types"); len(types) > 0 {
		result.allowedTypes = deepCopyStringSlice(types)
	}

	if scopes := cfg.GetStringSlice("conventional.scopes"); len(scopes) > 0 {
		result.allowedScopes = deepCopyStringSlice(scopes)
	}

	// Update scope requirement if explicitly set in config
	result.requireScope = cfg.GetBool("conventional.require_scope")

	if maxDescLen := cfg.GetInt("conventional.max_description_length"); maxDescLen > 0 {
		result.maxDescLength = maxDescLen
	} else if result.maxDescLength == 0 && cfg.GetInt("message.subject.max_length") > 0 {
		// If maxDescLength is not set, use the subject max length from config
		result.maxDescLength = cfg.GetInt("message.subject.max_length")
	}

	return result
}

// Validate validates a commit against the conventional commit rules.
// This method follows functional programming principles and does not modify the rule's state.
func (r ConventionalCommitRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Validate conventional commit format
	// Parse subject from commit
	subject := commit.Subject
	if subject == "" && commit.Message != "" {
		subject, _ = domain.SplitCommitMessage(commit.Message)
	}

	// Parse conventional format
	parts, err := parseConventionalFormat(subject)
	if err != nil {
		errors := []appErrors.ValidationError{
			appErrors.NewConventionalCommitError(
				appErrors.ErrInvalidConventionalFormat,
				"ConventionalCommit",
				"Commit message doesn't follow conventional format",
				"type(scope): description (e.g., 'feat: add login')",
			).WithContextMap(map[string]string{
				"subject": subject,
			}),
		}

		return errors
	}

	// Validate type
	if !isValidType(parts.Type, r.allowedTypes) {
		errors := []appErrors.ValidationError{
			appErrors.NewConventionalCommitError(
				appErrors.ErrInvalidConventionalType,
				"ConventionalCommit",
				fmt.Sprintf("Invalid type '%s'; allowed types: %s", parts.Type, strings.Join(r.allowedTypes, ", ")),
				strings.Join(r.allowedTypes, ", "),
			).WithContextMap(map[string]string{
				"type":          parts.Type,
				"allowed_types": strings.Join(r.allowedTypes, ", "),
			}),
		}

		return errors
	}

	// Validate scope if required
	if r.requireScope && parts.Scope == "" {
		errors := []appErrors.ValidationError{
			appErrors.NewConventionalCommitError(
				appErrors.ErrMissingConventionalScope,
				"ConventionalCommit",
				"Scope is required but not provided",
				"type(scope): description",
			),
		}

		return errors
	}

	// Validate allowed scopes if specified
	if parts.Scope != "" && len(r.allowedScopes) > 0 && !isValidScope(parts.Scope, r.allowedScopes) {
		errors := []appErrors.ValidationError{
			appErrors.NewConventionalCommitError(
				appErrors.ErrInvalidConventionalScope,
				"ConventionalCommit",
				fmt.Sprintf("Invalid scope '%s'; allowed scopes: %s", parts.Scope, strings.Join(r.allowedScopes, ", ")),
				strings.Join(r.allowedScopes, ", "),
			).WithContextMap(map[string]string{
				"scope":          parts.Scope,
				"allowed_scopes": strings.Join(r.allowedScopes, ", "),
			}),
		}

		return errors
	}

	// Validate description length
	if r.maxDescLength > 0 && len(parts.Description) > r.maxDescLength {
		errors := []appErrors.ValidationError{
			appErrors.NewConventionalCommitError(
				appErrors.ErrConventionalDescTooLong,
				"ConventionalCommit",
				fmt.Sprintf("Description too long (%d > %d)", len(parts.Description), r.maxDescLength),
				fmt.Sprintf("type[(scope)]: description (max %d chars)", r.maxDescLength),
			).WithContextMap(map[string]string{
				"length":     strconv.Itoa(len(parts.Description)),
				"max_length": strconv.Itoa(r.maxDescLength),
			}),
		}

		return errors
	}

	// All validations passed
	return []appErrors.ValidationError{}
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

	return commonSlices.Contains(allowedScopes, scope)
}

// Helper function for deep copying string slices.
func deepCopyStringSlice(src []string) []string {
	return slices.Clone(src)
}
