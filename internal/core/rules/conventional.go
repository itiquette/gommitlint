// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/common/slices"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// ConventionalParts represents the parts of a conventional commit.
type ConventionalParts struct {
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
	return slices.Reduce(
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

// Validate validates a commit against the conventional commit rules.
// This method follows functional programming principles and does not modify the rule's state.
func (r ConventionalCommitRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Get the logger from context
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating conventional commit format", "rule", r.Name(), "commit_hash", commit.Hash)

	// Build configuration from rule and context using a local struct for validation
	// We use a custom struct here because the rule's internal state needs different
	// fields than the standard types.ConventionalConfig
	config := struct {
		Required         bool
		Types            []string
		Scopes           []string
		ScopeRequired    bool
		ValidateBreaking bool
		MaxDescLength    int
	}{
		Required:         true, // Conventional rule is always required if enabled
		Types:            r.allowedTypes,
		Scopes:           r.allowedScopes,
		ScopeRequired:    r.requireScope,
		ValidateBreaking: r.validateBreaking,
		MaxDescLength:    r.maxDescLength,
	}

	// Get configuration directly from context
	cfg := contextx.GetConfig(ctx)
	if cfg != nil {
		// Only override settings if they are specified in the context configuration
		if types := cfg.GetStringSlice("conventional.types"); len(types) > 0 {
			config.Types = deepCopyStringSlice(types)
		}

		if scopes := cfg.GetStringSlice("conventional.scopes"); len(scopes) > 0 {
			config.Scopes = deepCopyStringSlice(scopes)
		}

		config.ScopeRequired = cfg.GetBool("conventional.require_scope")

		if maxDescLen := cfg.GetInt("conventional.max_description_length"); maxDescLen > 0 {
			config.MaxDescLength = maxDescLen
		} else if config.MaxDescLength == 0 && cfg.GetInt("message.subject.max_length") > 0 {
			// If maxDescLength is not set, use the subject max length from config
			config.MaxDescLength = cfg.GetInt("message.subject.max_length")
		}
	}

	// Log configuration for debugging
	logger.Debug("Conventional commit rule configuration",
		"allowed_types", config.Types,
		"allowed_scopes", config.Scopes,
		"require_scope", config.ScopeRequired,
		"validate_breaking", config.ValidateBreaking,
		"max_desc_length", config.MaxDescLength)

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
	if !isValidType(parts.Type, config.Types) {
		errors := []appErrors.ValidationError{
			appErrors.NewConventionalCommitError(
				appErrors.ErrInvalidConventionalType,
				"ConventionalCommit",
				fmt.Sprintf("Invalid type '%s'; allowed types: %s", parts.Type, strings.Join(config.Types, ", ")),
				strings.Join(config.Types, ", "),
			).WithContextMap(map[string]string{
				"type":          parts.Type,
				"allowed_types": strings.Join(config.Types, ", "),
			}),
		}

		return errors
	}

	// Validate scope if required
	if config.ScopeRequired && parts.Scope == "" {
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
	if parts.Scope != "" && len(config.Scopes) > 0 && !isValidScope(parts.Scope, config.Scopes) {
		errors := []appErrors.ValidationError{
			appErrors.NewConventionalCommitError(
				appErrors.ErrInvalidConventionalScope,
				"ConventionalCommit",
				fmt.Sprintf("Invalid scope '%s'; allowed scopes: %s", parts.Scope, strings.Join(config.Scopes, ", ")),
				strings.Join(config.Scopes, ", "),
			).WithContextMap(map[string]string{
				"scope":          parts.Scope,
				"allowed_scopes": strings.Join(config.Scopes, ", "),
			}),
		}

		return errors
	}

	// Validate description length
	if config.MaxDescLength > 0 && len(parts.Description) > config.MaxDescLength {
		errors := []appErrors.ValidationError{
			appErrors.NewConventionalCommitError(
				appErrors.ErrConventionalDescTooLong,
				"ConventionalCommit",
				fmt.Sprintf("Description too long (%d > %d)", len(parts.Description), config.MaxDescLength),
				fmt.Sprintf("type[(scope)]: description (max %d chars)", config.MaxDescLength),
			).WithContextMap(map[string]string{
				"length":     strconv.Itoa(len(parts.Description)),
				"max_length": strconv.Itoa(config.MaxDescLength),
			}),
		}

		return errors
	}

	// All validations passed
	return []appErrors.ValidationError{}
}

// parseConventionalFormat parses a commit subject into conventional commit parts.
func parseConventionalFormat(subject string) (ConventionalParts, error) {
	// Conventional commit format: <type>[(scope)][!]: <description>
	// Example: feat(api)!: add new endpoint
	pattern := `^(?P<type>[a-z]+)(?:\((?P<scope>[a-z0-9/-]+)\))?(?P<breaking>!)?:\s?(?P<description>.*)`
	regex := regexp.MustCompile(pattern)

	match := regex.FindStringSubmatch(subject)
	if match == nil {
		return ConventionalParts{}, errors.New("subject does not match conventional format")
	}

	// Extract named groups
	groups := make(map[string]string)

	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" && i < len(match) {
			groups[name] = match[i]
		}
	}

	return ConventionalParts{
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

// Helper function for deep copying string slices.
func deepCopyStringSlice(src []string) []string {
	return slices.DeepCopy(src)
}
