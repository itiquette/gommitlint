// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/model"
)

// SubjectRegex Format: type(scope)!: description.
var SubjectRegex = regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ](.+)$`)

// ConventionalCommit enforces the Conventional Commits specification format for commit messages.
//
// This rule validates that commit messages follow the structured format defined by the
// Conventional Commits specification (https://www.conventionalcommits.org/), which provides
// a standardized way to communicate the purpose and scope of changes. This format makes
// commit messages more readable and enables automated tools to parse commit history for
// generating changelogs and determining semantic versioning.
//
// The rule validates that commit messages follow the format:
//
//	type(scope)!: description
//
// Where:
//   - type: Indicates the kind of change (e.g., feat, fix, docs)
//   - scope: Optional field specifying the section of the codebase affected
//   - !: Optional indicator for breaking changes
//   - description: A concise explanation of the changes
//
// Examples:
//
//   - Valid conventional commits:
//     "feat: add user authentication" would pass
//     "fix(auth): resolve login timeout issue" would pass
//     "docs(readme): update installation instructions" would pass
//     "chore!: drop support for Node 8" would pass (breaking change)
//
//   - Invalid conventional commits:
//     "Add new feature" would fail (missing type prefix)
//     "feat:add user authentication" would fail (missing space after colon)
//     "FIX: resolve bug" would fail (type not lowercase)
//     "feat(auth):" would fail (empty description)
//
// If configured with allowed types or scopes, the rule also validates that the
// commit uses only approved types and scopes according to project conventions.
type ConventionalCommit struct {
	errors      []*model.ValidationError
	commitType  string // Store for verbose output
	scope       string // Store for verbose output
	hasBreaking bool   // Store for verbose output
}

// Name returns the rule identifier.
func (c ConventionalCommit) Name() string {
	return "ConventionalCommit"
}

// Result returns a concise string representation of the validation result.
func (c ConventionalCommit) Result() string {
	if len(c.errors) > 0 {
		return "Invalid conventional commit format"
	}

	return "Valid conventional commit format"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (c ConventionalCommit) VerboseResult() string {
	if len(c.errors) > 0 {
		// Return a more detailed error message in verbose mode
		switch c.errors[0].Code {
		case "invalid_format":
			return "Invalid format: doesn't follow conventional format 'type(scope)!: description'"
		case "invalid_type":
			var allowedTypes string

			for k, v := range c.errors[0].Context {
				if k == "allowed_types" {
					allowedTypes = strings.ReplaceAll(v, ",", ", ")

					break
				}
			}

			return "Invalid type '" + c.commitType + "'. Must be one of: " + allowedTypes
		case "invalid_scope":
			var allowedScopes string

			for k, v := range c.errors[0].Context {
				if k == "allowed_scopes" {
					allowedScopes = strings.ReplaceAll(v, ",", ", ")

					break
				}
			}

			return "Invalid scope '" + c.scope + "'. Must be one of: " + allowedScopes
		case "empty_description":
			return "Missing description after type/scope"
		case "description_too_long":
			var actualLength, maxLength string

			for k, v := range c.errors[0].Context {
				if k == "actual_length" {
					actualLength = v
				} else if k == "max_length" {
					maxLength = v
				}
			}

			return "Description too long (" + actualLength + " chars). Maximum length is " + maxLength + " characters"
		case "spacing_error":
			return "Spacing error: Must have exactly one space after colon"
		default:
			return c.errors[0].Error()
		}
	}

	// Success verbose message with more details
	result := "Valid conventional commit with type '" + c.commitType + "'"
	if c.scope != "" {
		result += " and scope '" + c.scope + "'"
	}

	if c.hasBreaking {
		result += " (breaking change)"
	}

	return result
}

// Errors returns all validation errors.
func (c ConventionalCommit) Errors() []*model.ValidationError {
	return c.errors
}

// Help returns guidance for fixing rule violations.
func (c ConventionalCommit) Help() string {
	if len(c.errors) == 0 {
		return "No errors to fix"
	}

	// Check for specific error codes if available
	if len(c.errors) > 0 {
		switch c.errors[0].Code {
		case "invalid_format":
			return `Your commit message does not follow the conventional commit format.
The correct format is: type(scope)!: description
Examples:
- feat: add new feature
- fix(auth): resolve login issue
- chore!: drop support for Node 8
Make sure:
- The type is lowercase (feat, fix, docs, etc.)
- The scope is in parentheses (if provided)
- There's a colon followed by a single space
- Include a description after the space`

		case "invalid_type":
			return `The commit type you used is not in the allowed list of types.
Your commit should use one of the approved types from the allowed list.
Check your project documentation or configuration for the full list of allowed types.`

		case "invalid_scope":
			return `The scope you specified is not in the allowed list of scopes.
Scopes define the section of the codebase your change affects.
Check your project documentation or configuration for the full list of allowed scopes.`

		case "empty_description":
			return `Your commit message is missing a description.
After the type(scope): prefix, you must include a description that explains what the commit does.
Example: feat(ui): add new button component`

		case "description_too_long":
			return `Your commit description exceeds the maximum allowed length.
Keep your commit description concise while still being descriptive.
Consider breaking down large changes into multiple smaller commits if possible.`

		case "spacing_error":
			return `There should be exactly one space after the colon in your commit message.
Correct: feat: add feature
Incorrect: feat:add feature or feat:  add feature`
		}
	}

	// Default help message
	return `Ensure your commit message follows the conventional commit format:
type(scope)!: description
Examples:
- feat: add user authentication
- fix(api): resolve timeout issue
- docs(readme): update installation instructions
- chore!: drop support for legacy systems`
}

// addError adds a structured validation error.
func (c *ConventionalCommit) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("ConventionalCommit", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	c.errors = append(c.errors, err)
}

func (c *ConventionalCommit) AddTestError(code, message string, context map[string]string) {
	c.addError(code, message, context)
}

// ValidateConventionalCommit checks if a commit subject follows conventional format.
//
// Parameters:
//   - subject: The commit subject line to validate
//   - types: Optional list of allowed commit types (e.g., feat, fix, docs)
//   - scopes: Optional list of allowed commit scopes (e.g., auth, ui, api)
//   - descLength: Maximum allowed description length (0 means use default of 72)
//
// The function validates several aspects of the conventional commit format:
//  1. Basic format compliance (type(scope)!: description)
//  2. Correct spacing after the colon
//  3. Valid commit type (if allowed types are specified)
//  4. Valid scope (if allowed scopes are specified)
//  5. Non-empty description
//  6. Description length within limits
//
// For multi-scope commits using comma separators (e.g., "feat(ui,api)"), each scope
// is individually validated against the allowed scopes list.
//
// Returns:
//   - A ConventionalCommit instance with validation results
func ValidateConventionalCommit(subject string, types []string, scopes []string, descLength int) ConventionalCommit {
	rule := ConventionalCommit{}

	// Handle empty subject early
	if strings.TrimSpace(subject) == "" {
		rule.addError(
			"invalid_format",
			"invalid conventional commit format: empty message",
			map[string]string{
				"subject": subject,
			},
		)

		return rule
	}

	// Default description length if not specified
	if descLength == 0 {
		descLength = 72
	}

	// Validate basic format first
	if !SubjectRegex.MatchString(subject) {
		rule.addError(
			"invalid_format",
			"invalid conventional commit format: "+subject,
			map[string]string{
				"subject": subject,
			},
		)

		return rule
	}

	//Simple check for ": " vs ":  " (one space vs multiple spaces)
	if strings.Contains(subject, ":  ") {
		rule.addError(
			"spacing_error",
			"spacing error: must have exactly one space after colon",
			map[string]string{
				"subject": subject,
			},
		)

		return rule
	}

	// Parse the subject according to conventional commit format
	matches := SubjectRegex.FindStringSubmatch(subject)
	if len(matches) != 5 {
		rule.addError(
			"invalid_format",
			"invalid conventional commit format: "+subject,
			map[string]string{
				"subject": subject,
			},
		)

		return rule
	}

	// Extract components
	commitType := matches[1]
	scope := matches[2]
	hasBreaking := matches[3] == "!"
	description := matches[4]

	// Store for verbose output
	rule.commitType = commitType
	rule.scope = scope
	rule.hasBreaking = hasBreaking

	// Validate type
	if len(types) > 0 && !slices.Contains(types, commitType) {
		rule.addError(
			"invalid_type",
			"invalid type \""+commitType+"\": allowed types are "+strings.Join(types, ", "),
			map[string]string{
				"type":          commitType,
				"allowed_types": strings.Join(types, ","),
			},
		)

		return rule
	}

	// Validate scope if provided and scope list is defined
	if scope != "" && len(scopes) > 0 {
		scopesList := strings.Split(scope, ",")
		for _, scope := range scopesList {
			if !slices.Contains(scopes, scope) {
				rule.addError(
					"invalid_scope",
					"invalid scope \""+scope+"\": allowed scopes are "+strings.Join(scopes, ", "),
					map[string]string{
						"scope":          scope,
						"allowed_scopes": strings.Join(scopes, ","),
					},
				)

				return rule
			}
		}
	}

	// Validate description content
	if strings.TrimSpace(description) == "" {
		rule.addError(
			"empty_description",
			"empty description: description must contain non-whitespace characters",
			nil,
		)

		return rule
	}

	// Validate description length
	if len(description) > descLength {
		rule.addError(
			"description_too_long",
			"description too long: "+strconv.Itoa(len(description))+" characters (max: "+strconv.Itoa(descLength)+")",
			map[string]string{
				"actual_length": strconv.Itoa(len(description)),
				"max_length":    strconv.Itoa(descLength),
			},
		)

		return rule
	}

	return rule
}
