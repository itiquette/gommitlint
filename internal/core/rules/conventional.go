// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

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
	*BaseRule
	allowedTypes     []string
	allowedScopes    []string
	requireScope     bool
	validateBreaking bool
	maxDescLength    int

	// Captured values for better reporting
	commitType  string
	scope       string
	hasBreaking bool
}

// ConventionalCommitOption is a function that configures a ConventionalCommitRule.
type ConventionalCommitOption func(*ConventionalCommitRule)

// WithAllowedTypes sets the allowed commit types.
func WithAllowedTypes(types []string) ConventionalCommitOption {
	return func(r *ConventionalCommitRule) {
		r.allowedTypes = types
	}
}

// WithAllowedScopes sets the allowed commit scopes.
func WithAllowedScopes(scopes []string) ConventionalCommitOption {
	return func(r *ConventionalCommitRule) {
		r.allowedScopes = scopes
	}
}

// WithRequiredScope makes the scope mandatory in commit messages.
func WithRequiredScope() ConventionalCommitOption {
	return func(r *ConventionalCommitRule) {
		r.requireScope = true
	}
}

// WithBreakingChangeValidation enables validation of the breaking change marker.
func WithBreakingChangeValidation() ConventionalCommitOption {
	return func(r *ConventionalCommitRule) {
		r.validateBreaking = true
	}
}

// WithMaxDescLength sets the maximum description length.
func WithMaxDescLength(maxLength int) ConventionalCommitOption {
	return func(r *ConventionalCommitRule) {
		// Skip if the value is 0 or negative
		if maxLength > 0 {
			r.maxDescLength = maxLength
		}
	}
}

// NewConventionalCommitRule creates a new rule with the specified options.
func NewConventionalCommitRule(options ...ConventionalCommitOption) *ConventionalCommitRule {
	rule := &ConventionalCommitRule{
		BaseRule: NewBaseRule("ConventionalCommit"),
		allowedTypes: []string{
			"feat", "fix", "docs", "style", "refactor", "perf",
			"test", "build", "ci", "chore", "revert",
		},
		allowedScopes:    []string{}, // Empty means all scopes are allowed
		requireScope:     false,      // Default to not requiring scope
		validateBreaking: false,      // Default to not validating breaking changes
		maxDescLength:    72,         // Default max length for description
	}

	// Apply options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// addError is a helper that adds errors using the new error system.
func (r *ConventionalCommitRule) addError(code appErrors.ValidationErrorCode, message string, context map[string]string) {
	r.AddErrorWithContext(code, message, context)
}

// Validate validates a commit against the conventional commit rules.
func (r *ConventionalCommitRule) Validate(commit *domain.CommitInfo) []appErrors.ValidationError {
	// Reset errors and state
	r.ClearErrors()
	r.commitType = ""
	r.scope = ""
	r.hasBreaking = false

	// Mark the rule as having been run
	r.MarkAsRun()

	// Get the subject
	subject := strings.TrimSpace(commit.Subject)

	// Compile the regex (optimally this would be compiled once at rule creation)
	pattern := `^(?P<type>[a-z]+)(?:\((?P<scope>[a-z0-9/-]+)\))?(?P<breaking>!)?:\s?(?P<description>.*)`
	regex := regexp.MustCompile(pattern)

	// Check if the subject follows the pattern
	matches := regex.FindStringSubmatch(subject)
	if len(matches) == 0 {
		r.addError(
			appErrors.ErrInvalidFormat,
			"commit message doesn't follow conventional format: type(scope)!: description",
			nil,
		)

		return r.Errors()
	}

	// Check for spacing issues
	if strings.Contains(subject, ":  ") {
		r.addError(
			appErrors.ErrSpacing,
			"commit message has too many spaces after colon (should be exactly one)",
			nil,
		)

		return r.Errors()
	}

	// Extract capture groups
	typeIdx := regex.SubexpIndex("type")
	scopeIdx := regex.SubexpIndex("scope")
	breakingIdx := regex.SubexpIndex("breaking")
	descIdx := regex.SubexpIndex("description")

	// Extract and validate type
	if typeIdx >= 0 && typeIdx < len(matches) {
		r.commitType = matches[typeIdx]
		if !r.isValidType(r.commitType) {
			allowedTypes := strings.Join(r.allowedTypes, ",")
			r.addError(
				appErrors.ErrInvalidType,
				"invalid commit type: "+r.commitType,
				map[string]string{
					"type":          r.commitType,
					"allowed_types": allowedTypes,
				},
			)

			return r.Errors()
		}
	}

	// Extract and validate scope
	if scopeIdx >= 0 && scopeIdx < len(matches) {
		r.scope = matches[scopeIdx]
		if r.scope != "" && !r.isValidScope(r.scope) {
			allowedScopes := strings.Join(r.allowedScopes, ",")
			r.addError(
				appErrors.ErrInvalidScope,
				"invalid commit scope: "+r.scope,
				map[string]string{
					"scope":          r.scope,
					"allowed_scopes": allowedScopes,
				},
			)

			return r.Errors()
		}
	}

	// Check if scope is required but missing
	if r.requireScope && (scopeIdx < 0 || scopeIdx >= len(matches) || matches[scopeIdx] == "") {
		r.addError(
			appErrors.ErrInvalidScope,
			"commit scope is required but not provided",
			nil,
		)

		return r.Errors()
	}

	// Extract breaking change marker
	if breakingIdx >= 0 && breakingIdx < len(matches) {
		r.hasBreaking = matches[breakingIdx] != ""
	}

	// Validate description
	if descIdx >= 0 && descIdx < len(matches) {
		description := matches[descIdx]
		if strings.TrimSpace(description) == "" {
			r.addError(
				appErrors.ErrEmptyDescription,
				"commit description cannot be empty",
				nil,
			)

			return r.Errors()
		}

		// Check description length - store it for error reporting
		descriptionLength := len(description)

		// Check description length
		if r.maxDescLength > 0 && descriptionLength > r.maxDescLength {
			r.addError(
				appErrors.ErrDescriptionTooLong,
				fmt.Sprintf("commit description is too long (%d chars, max is %d)", descriptionLength, r.maxDescLength),
				map[string]string{
					"length":     strconv.Itoa(descriptionLength),
					"max_length": strconv.Itoa(r.maxDescLength),
				},
			)

			return r.Errors()
		}
	}

	return r.Errors()
}

// isValidType checks if the commit type is in the list of allowed types.
func (r *ConventionalCommitRule) isValidType(commitType string) bool {
	// If no allowed types are specified, all types are allowed
	if len(r.allowedTypes) == 0 {
		return true
	}

	for _, t := range r.allowedTypes {
		if commitType == t {
			return true
		}
	}

	return false
}

// isValidScope checks if the commit scope is in the list of allowed scopes.
func (r *ConventionalCommitRule) isValidScope(scope string) bool {
	// If no allowed scopes are specified, all scopes are allowed
	if len(r.allowedScopes) == 0 {
		return true
	}

	for _, s := range r.allowedScopes {
		if scope == s {
			return true
		}
	}

	return false
}

// Result returns a concise validation result.
func (r *ConventionalCommitRule) Result() string {
	if r.HasErrors() {
		return "Invalid conventional commit format"
	}

	return "Valid conventional commit format"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *ConventionalCommitRule) VerboseResult() string {
	if r.HasErrors() {
		errors := r.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		// Return a more detailed error message in verbose mode
		switch validationErr.Code {
		case string(appErrors.ErrInvalidFormat):
			return "Invalid format: doesn't follow conventional format 'type(scope)!: description'"
		case string(appErrors.ErrInvalidType):
			var allowedTypes string
			if val, ok := validationErr.Context["allowed_types"]; ok {
				allowedTypes = strings.ReplaceAll(val, ",", ", ")
			}

			return fmt.Sprintf("Invalid type '%s'. Must be one of: %s", r.commitType, allowedTypes)
		case string(appErrors.ErrInvalidScope):
			var allowedScopes string
			if val, ok := validationErr.Context["allowed_scopes"]; ok {
				allowedScopes = strings.ReplaceAll(val, ",", ", ")
			}

			return fmt.Sprintf("Invalid scope '%s'. Must be one of: %s", r.scope, allowedScopes)
		case string(appErrors.ErrEmptyDescription):
			return "Commit description is empty. A description following the colon is required."
		case string(appErrors.ErrDescriptionTooLong):
			maxLength := "100"
			if ml, ok := validationErr.Context["max_length"]; ok {
				maxLength = ml
			}

			return fmt.Sprintf("Commit description is too long. Maximum is %s characters but got %s characters.",
				maxLength, validationErr.Context["length"])
		case string(appErrors.ErrSpacing):
			return "Spacing error: There should be exactly one space after the colon."
		}

		// Default to the error message
		return validationErr.Message
	}

	// Build a nice formatted success message
	result := "Valid conventional commit format: "
	if r.hasBreaking {
		result += "BREAKING CHANGE - "
	}

	result += r.commitType
	if r.scope != "" {
		result += fmt.Sprintf("(%s)", r.scope)
	}

	return result
}

// Help returns guidance on how to fix the rule violation.
func (r *ConventionalCommitRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix"
	}

	errors := r.Errors()
	if len(errors) > 0 {
		// Cast to ValidationError if possible
		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		switch validationErr.Code {
		case string(appErrors.ErrInvalidFormat):
			return `Follow the conventional commit format:
    <type>[optional scope][optional !]: <description>

Examples:
- feat: add new feature
- fix(auth): resolve login issue
- feat(api)!: change user API endpoints

The format is strict and requires the specific characters shown above.`

		case string(appErrors.ErrInvalidType):
			allowedTypes := ""
			if val, ok := validationErr.Context["allowed_types"]; ok {
				allowedTypes = strings.ReplaceAll(val, ",", ", ")
			}

			return fmt.Sprintf(`Use only allowed commit types: %s

Examples:
- feat: adds a new feature
- fix: fixes a bug
- docs: documentation only changes
- style: formatting changes
- refactor: code change that neither fixes a bug nor adds a feature
- perf: improves performance
- test: adds missing tests or corrects existing tests
- build: affects build system or external dependencies
- ci: changes CI configuration files and scripts
- chore: other changes that don't modify src or test files
- revert: reverts a previous commit`, allowedTypes)

		case string(appErrors.ErrInvalidScope):
			if r.requireScope {
				allowedScopes := ""
				if val, ok := validationErr.Context["allowed_scopes"]; ok {
					allowedScopes = strings.ReplaceAll(val, ",", ", ")
				}

				if allowedScopes != "" {
					return fmt.Sprintf(`A scope is required and must be one of: %s

Example:
- feat(%s): add new feature

The scope must be in parentheses and directly after the type.`, allowedScopes, r.allowedScopes[0])
				}

				return `A scope is required but was not provided.

Example:
- feat(auth): add new feature

The scope must be in parentheses and directly after the type.`
			}

			allowedScopes := ""
			if val, ok := validationErr.Context["allowed_scopes"]; ok {
				allowedScopes = strings.ReplaceAll(val, ",", ", ")
			}

			return fmt.Sprintf(`Use only allowed scopes: %s

Example:
- feat(%s): add new feature

The scope must be in parentheses and directly after the type.`, allowedScopes, r.allowedScopes[0])

		case string(appErrors.ErrEmptyDescription):
			return `Provide a description after the colon.

Examples:
- feat: add new user authentication feature
- fix(auth): resolve login issue with special characters

The description should be concise but descriptive.`

		case string(appErrors.ErrDescriptionTooLong):
			maxLength := "100"
			if ml, ok := validationErr.Context["max_length"]; ok {
				maxLength = ml
			}

			return fmt.Sprintf(`Keep the commit description under %s characters.

Long descriptions should be moved to the commit body, which comes after a blank line following the subject.

Example:
feat: add new authentication method

This commit introduces a new authentication method that allows users to log in with their social media accounts.`, maxLength)

		case string(appErrors.ErrSpacing):
			return `Use exactly one space after the colon in commit messages.

Correct:
feat: add new feature

Incorrect:
feat:  add new feature

The conventional commit format requires exactly one space between the colon and the description.`
		}
	}

	// Default help
	return `Follow the conventional commit format:
<type>[optional scope][optional !]: <description>

Examples:
- feat: add new feature
- fix(auth): resolve login issue
- feat(api)!: change user API endpoints

For more information, see https://www.conventionalcommits.org/`
}
