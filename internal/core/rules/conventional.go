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
	BaseRule
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
type ConventionalCommitOption func(ConventionalCommitRule) ConventionalCommitRule

// WithAllowedTypes sets the allowed commit types.
func WithAllowedTypes(types []string) ConventionalCommitOption {
	return func(r ConventionalCommitRule) ConventionalCommitRule {
		newRule := r
		newRule.allowedTypes = types

		return newRule
	}
}

// WithAllowedScopes sets the allowed commit scopes.
func WithAllowedScopes(scopes []string) ConventionalCommitOption {
	return func(r ConventionalCommitRule) ConventionalCommitRule {
		newRule := r
		newRule.allowedScopes = scopes

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

// NewConventionalCommitRule creates a new rule with the specified options.
func NewConventionalCommitRule(options ...ConventionalCommitOption) ConventionalCommitRule {
	rule := ConventionalCommitRule{
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
		rule = option(rule)
	}

	return rule
}

// NewConventionalCommitRuleWithConfig creates a new rule using configuration.
func NewConventionalCommitRuleWithConfig(config domain.ConventionalConfigProvider) ConventionalCommitRule {
	var options []ConventionalCommitOption

	// Apply the allowed types if provided
	if types := config.ConventionalTypes(); len(types) > 0 {
		options = append(options, WithAllowedTypes(types))
	}

	// Apply the allowed scopes if provided
	if scopes := config.ConventionalScopes(); len(scopes) > 0 {
		options = append(options, WithAllowedScopes(scopes))
	}

	// Apply the max description length if provided
	if maxLength := config.ConventionalMaxDescriptionLength(); maxLength > 0 {
		options = append(options, WithMaxDescLength(maxLength))
	}

	return NewConventionalCommitRule(options...)
}

// Name returns the name of the rule.
func (r ConventionalCommitRule) Name() string {
	return r.BaseRule.Name()
}

// Errors returns the validation errors.
func (r ConventionalCommitRule) Errors() []appErrors.ValidationError {
	return r.BaseRule.Errors()
}

// HasErrors checks if there are any validation errors.
func (r ConventionalCommitRule) HasErrors() bool {
	return r.BaseRule.HasErrors()
}

// SetErrors sets the validation errors for this rule.
// This method supports value semantics by returning a new instance.
func (r ConventionalCommitRule) SetErrors(errors []appErrors.ValidationError) ConventionalCommitRule {
	result := r
	baseRule := r.BaseRule.WithClearedErrors()

	for _, err := range errors {
		baseRule = baseRule.WithError(err)
	}

	result.BaseRule = baseRule

	return result
}

// addError adds an error to the rule and returns a new rule instance.
func (r ConventionalCommitRule) addError(err appErrors.ValidationError) ConventionalCommitRule {
	result := r
	result.BaseRule = r.BaseRule.WithError(err)

	return result
}

// validateConventionalWithState validates a commit and returns both errors and an updated rule state.
func validateConventionalWithState(rule ConventionalCommitRule, commit domain.CommitInfo) ([]appErrors.ValidationError, ConventionalCommitRule) {
	updatedRule := rule
	// Mark as run
	updatedRule.BaseRule = updatedRule.BaseRule.WithRun()

	// Reset captured values
	updatedRule.commitType = ""
	updatedRule.scope = ""
	updatedRule.hasBreaking = false

	// Get the subject
	subject := strings.TrimSpace(commit.Subject)

	// Compile the regex
	pattern := `^(?P<type>[a-z]+)(?:\((?P<scope>[a-z0-9/-]+)\))?(?P<breaking>!)?:\s?(?P<description>.*)`
	regex := regexp.MustCompile(pattern)

	// Check if the subject follows the pattern
	matches := regex.FindStringSubmatch(subject)
	if len(matches) == 0 {
		// Create error context with rich information
		ctx := appErrors.NewContext().WithCommit(
			commit.Hash,    // commit hash
			commit.Message, // full commit message
			commit.Subject, // subject line
			commit.Body,    // body text
		)

		// Create a rich error
		err := appErrors.CreateRichError(
			updatedRule.Name(),
			appErrors.ErrInvalidFormat,
			"commit message doesn't follow conventional format: type(scope)!: description",
			`Your commit message doesn't follow the conventional format: type(scope)!: description

The Conventional Commits specification is a lightweight convention for creating commit messages.
This format makes commits readable and machine-parsable, allowing automated tools to generate changelogs.

Required format:
<type>[optional scope][optional !]: <description>

Examples of valid conventional commits:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions
- perf(api): optimize database queries
- feat(user)!: change user API response format

Common types:
- feat: a new feature
- fix: a bug fix
- docs: documentation changes
- style: changes that don't affect code meaning (whitespace, formatting)
- refactor: code change that neither fixes a bug nor adds a feature
- perf: code change that improves performance
- test: adding or correcting tests
- build: changes to build system or dependencies
- ci: changes to CI configuration
- chore: other changes that don't modify src or test files
- revert: reverts a previous commit

The format is strict and requires the specific characters shown above.`,
			ctx,
		)

		updatedRule = updatedRule.addError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Check for spacing issues
	if strings.Contains(subject, ":  ") {
		// Create error context
		ctx := appErrors.NewContext().WithCommit(
			commit.Hash,    // commit hash
			commit.Message, // full commit message
			commit.Subject, // subject line
			commit.Body,    // body text
		)

		// Create a rich error
		err := appErrors.CreateRichError(
			updatedRule.Name(),
			appErrors.ErrSpacing,
			"commit message has too many spaces after colon (should be exactly one)",
			`Use exactly one space after the colon in commit messages.
Correct:
feat: add new feature
Incorrect:
feat:  add new feature
The conventional commit format requires exactly one space between the colon and the description.`,
			ctx,
		)

		updatedRule = updatedRule.addError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Extract capture groups
	typeIdx := regex.SubexpIndex("type")
	scopeIdx := regex.SubexpIndex("scope")
	breakingIdx := regex.SubexpIndex("breaking")
	descIdx := regex.SubexpIndex("description")

	// Extract and validate type
	if typeIdx >= 0 && typeIdx < len(matches) {
		commitType := matches[typeIdx]
		updatedRule.commitType = commitType

		if !updatedRule.isValidType(commitType) {
			allowedTypes := strings.Join(updatedRule.allowedTypes, ",")
			err := appErrors.New(
				updatedRule.Name(),
				appErrors.ErrInvalidType,
				"invalid commit type: "+commitType,
				appErrors.WithContextMap(map[string]string{
					"type":          commitType,
					"allowed_types": allowedTypes,
				}),
			)

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}
	}

	// Extract and validate scope
	if scopeIdx >= 0 && scopeIdx < len(matches) {
		scope := matches[scopeIdx]
		updatedRule.scope = scope

		if scope != "" && !updatedRule.isValidScope(scope) {
			allowedScopes := strings.Join(updatedRule.allowedScopes, ",")
			err := appErrors.New(
				updatedRule.Name(),
				appErrors.ErrInvalidScope,
				"invalid commit scope: "+scope,
				appErrors.WithContextMap(map[string]string{
					"scope":          scope,
					"allowed_scopes": allowedScopes,
				}),
			)

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}
	}

	// Check if scope is required but missing
	if updatedRule.requireScope && (scopeIdx < 0 || scopeIdx >= len(matches) || matches[scopeIdx] == "") {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrInvalidScope,
			"commit scope is required but not provided",
			appErrors.WithContextMap(map[string]string{}),
		)

		updatedRule = updatedRule.addError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Extract breaking change marker
	if breakingIdx >= 0 && breakingIdx < len(matches) {
		updatedRule.hasBreaking = matches[breakingIdx] != ""
	}

	// Validate description
	if descIdx >= 0 && descIdx < len(matches) {
		description := matches[descIdx]
		if strings.TrimSpace(description) == "" {
			err := appErrors.New(
				updatedRule.Name(),
				appErrors.ErrEmptyDescription,
				"commit description cannot be empty",
				appErrors.WithContextMap(map[string]string{}),
			)

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}

		// Check description length
		descriptionLength := len(description)

		// Check description length
		if updatedRule.maxDescLength > 0 && descriptionLength > updatedRule.maxDescLength {
			err := appErrors.New(
				updatedRule.Name(),
				appErrors.ErrDescriptionTooLong,
				fmt.Sprintf("commit description is too long (%d chars, max is %d)", descriptionLength, updatedRule.maxDescLength),
				appErrors.WithContextMap(map[string]string{
					"length":     strconv.Itoa(descriptionLength),
					"max_length": strconv.Itoa(updatedRule.maxDescLength),
				}),
			)

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}
	}

	return []appErrors.ValidationError{}, updatedRule
}

// Validate validates a commit against the conventional commit rules.
// This method follows functional programming principles and does not modify the rule's state.
func (r ConventionalCommitRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	// Use the pure functional approach
	errors, _ := validateConventionalWithState(r, commit)

	return errors
}

// isValidType checks if the commit type is in the list of allowed types.
func (r ConventionalCommitRule) isValidType(commitType string) bool {
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
func (r ConventionalCommitRule) isValidScope(scope string) bool {
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
func (r ConventionalCommitRule) Result() string {
	if r.HasErrors() {
		return "Invalid conventional commit format"
	}

	return "Valid conventional commit format"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r ConventionalCommitRule) VerboseResult() string {
	if r.HasErrors() {
		errors := r.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		// Use the enhanced error formatter for rich output
		formatter := appErrors.NewTextFormatter(true) // verbose mode

		return formatter.FormatError(errors[0])
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
func (r ConventionalCommitRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix. This rule checks that commits follow the conventional commit format with proper type, structure, and description (e.g., feat: add new login feature, fix(auth): resolve timeout issue)."
	}

	errors := r.Errors()
	if len(errors) > 0 {
		// Get help text from the enhanced error
		helpText := errors[0].GetHelp()
		if helpText != "" {
			return helpText
		}

		// Fallback to default help by error code
		validationErr := errors[0]

		switch validationErr.Code {
		case string(appErrors.ErrInvalidFormat):
			return `Your commit message doesn't follow the conventional format: type(scope)!: description

The Conventional Commits specification is a lightweight convention for creating commit messages.
This format makes commits readable and machine-parsable, allowing automated tools to generate changelogs.

Required format:
<type>[optional scope][optional !]: <description>

Examples of valid conventional commits:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions
- perf(api): optimize database queries
- feat(user)!: change user API response format

Common types:
- feat: a new feature
- fix: a bug fix
- docs: documentation changes
- style: changes that don't affect code meaning (whitespace, formatting)
- refactor: code change that neither fixes a bug nor adds a feature
- perf: code change that improves performance
- test: adding or correcting tests
- build: changes to build system or dependencies
- ci: changes to CI configuration
- chore: other changes that don't modify src or test files
- revert: reverts a previous commit

The format is strict and requires the specific characters shown above.`
		case string(appErrors.ErrInvalidType):
			allowedTypes := strings.Join(r.allowedTypes, ", ")

			return fmt.Sprintf(`Invalid commit type - Use only allowed types: %s

Each commit type has a specific purpose:

- feat: adds a new feature to the application or library
  Example: feat: add user authentication
  
- fix: patches a bug in the codebase
  Example: fix: prevent crash when invalid data is submitted
  
- docs: documentation only changes
  Example: docs: update API documentation with new endpoints
  
- style: changes that don't affect code meaning (white-space, formatting, missing semi-colons)
  Example: style: format code according to linting rules
  
- refactor: code change that neither fixes a bug nor adds a feature
  Example: refactor: simplify authentication logic
  
- perf: code changes that improve performance
  Example: perf: optimize database queries
  
- test: adding missing tests or correcting existing tests
  Example: test: add unit tests for login function
  
- build: changes that affect the build system or external dependencies
  Example: build: update webpack configuration
  
- ci: changes to CI configuration files and scripts
  Example: ci: add new GitHub Actions workflow
  
- chore: other changes that don't modify src or test files
  Example: chore: update dependencies
  
- revert: reverts a previous commit
  Example: revert: feat: add login feature

Choose one of the allowed types for your commit.`, allowedTypes)
		case string(appErrors.ErrInvalidScope):
			allowedScopes := strings.Join(r.allowedScopes, ", ")
			if r.requireScope {
				if len(r.allowedScopes) > 0 {
					return fmt.Sprintf(`Scope Error: A scope is required and must be one of: %s

The scope indicates which part of the project is affected by your change.
It must be placed in parentheses immediately after the commit type.

Correct format:
<type>(<scope>): <description>

Examples with valid scopes:
- feat(%s): add login functionality
- fix(%s): resolve data validation error
- docs(%s): update API documentation

Each scope represents a specific area of the codebase. Using the correct scope
helps maintainers understand which areas are being modified.`,
						allowedScopes,
						r.allowedScopes[0],
						r.allowedScopes[0],
						r.allowedScopes[0])
				}

				return `Scope Error: A scope is required but was not provided.

The scope indicates which part of the project is affected by your change.
It must be placed in parentheses immediately after the commit type.

Correct format:
<type>(<scope>): <description>

Examples with scopes:
- feat(auth): add login functionality
- fix(api): resolve data validation error
- docs(readme): update installation instructions

Each scope represents a specific area of the codebase. Using an appropriate scope
helps maintainers understand which areas are being modified.`
			}

			return fmt.Sprintf(`Scope Error: Invalid scope used. Use only allowed scopes: %s

The scope indicates which part of the project is affected by your change.
It must be placed in parentheses immediately after the commit type.

Correct format:
<type>(<scope>): <description>

Examples with valid scopes:
- feat(%s): add login functionality
- fix(%s): resolve data validation error
- docs(%s): update API documentation

Each scope represents a specific area of the codebase. Using the correct scope
helps maintainers understand which areas are being modified.`,
				allowedScopes,
				r.allowedScopes[0],
				r.allowedScopes[0],
				r.allowedScopes[0])
		case string(appErrors.ErrEmptyDescription):
			return `Description Error: Your commit message is missing a description.

The description is a required part of a conventional commit message that explains what changes were made.
It should be concise but descriptive enough to understand the purpose of the commit.

Correct format:
<type>[optional scope][optional !]: <description>

Examples of good descriptions:
- feat: add user authentication system
- fix(auth): resolve login timeout issue
- docs: update installation instructions
- refactor(api): simplify error handling
- test: add unit tests for user service

A good description:
- Starts with a lowercase letter
- Uses imperative, present tense ("add", not "added" or "adds")
- Does not end with a period
- Is clear and specific about what was changed
- Is concise but descriptive (aim for 50 characters or less)

Without a proper description, your commit history will be difficult to understand and navigate.`
		case string(appErrors.ErrDescriptionTooLong):
			return fmt.Sprintf(`Description Length Error: Your commit description exceeds the maximum length of %d characters.

The subject line of a commit should be concise and focused on the core change.
Longer explanations should be placed in the commit body after a blank line.

Instead of using a long description like this:
❌ feat: implement comprehensive user authentication system with social login integration and multi-factor authentication options

Split it into a concise subject and detailed body:
✅ feat: add user authentication system

This commit implements:
- Email/password login
- Social login integration
- Multi-factor authentication
- Session management
- Password recovery

Keeping your subject line under %d characters improves readability in:
- Git history logs
- GitHub/GitLab commit lists
- CLI outputs
- Changelogs
- Email notifications

Remember that the commit body (separated by a blank line) can contain all the details needed to fully understand the change.`, r.maxDescLength, r.maxDescLength)
		case string(appErrors.ErrSpacing):
			return `Spacing Error: You must use exactly one space after the colon in commit messages.

The conventional commit format is very specific about spacing:
- Exactly ONE space must follow the colon
- No spaces before the colon
- Additional spaces change the parsing of the message

Examples:

✅ Correct:
feat: add new feature
fix(auth): resolve login issue

❌ Incorrect:
feat:  add new feature  (too many spaces after colon)
feat : add new feature  (space before colon)
feat:add new feature    (no space after colon)

This strict spacing requirement ensures consistent parsing across tools and
helps maintain a clean, readable commit history. Even a single extra space
can cause problems with automated tools that generate changelogs or analyze
commit history.`
		}
	}

	// Default help
	return `Your commit message should follow the conventional commit format

The Conventional Commits specification is a lightweight convention for creating commit messages.
This format makes commits readable and machine-parsable, allowing automated tools to generate changelogs.

Required format:
<type>[optional scope][optional !]: <description>

Examples of valid conventional commits:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions
- perf(api): optimize database queries
- feat(user)!: change user API response format

Common types:
- feat: a new feature
- fix: a bug fix
- docs: documentation changes
- style: changes that don't affect code meaning (whitespace, formatting)
- refactor: code change that neither fixes a bug nor adds a feature
- perf: code change that improves performance
- test: adding or correcting tests
- build: changes to build system or dependencies
- ci: changes to CI configuration
- chore: other changes that don't modify src or test files
- revert: reverts a previous commit

For more information, see https://www.conventionalcommits.org/`
}
