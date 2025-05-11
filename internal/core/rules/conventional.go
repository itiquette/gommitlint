// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
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

// NewConventionalCommitRule creates a new rule with the specified options.
func NewConventionalCommitRule(options ...ConventionalCommitOption) ConventionalCommitRule {
	// Create initial rule with default values
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

	// Apply options using Reduce for a more functional approach
	return contextx.Reduce(
		options,
		rule,
		func(currentRule ConventionalCommitRule, option ConventionalCommitOption) ConventionalCommitRule {
			return option(currentRule)
		},
	)
}

// NewConventionalCommitRuleWithConfig creates a new rule using configuration.
func NewConventionalCommitRuleWithConfig(config domain.ConventionalConfigProvider) ConventionalCommitRule {
	// Create option collectors using pure functions
	optionCollectors := []func(domain.ConventionalConfigProvider) []ConventionalCommitOption{
		// Collect type options
		func(c domain.ConventionalConfigProvider) []ConventionalCommitOption {
			types := c.ConventionalTypes()
			if len(types) > 0 {
				return []ConventionalCommitOption{WithAllowedTypes(types)}
			}

			return nil
		},
		// Collect scope options
		func(c domain.ConventionalConfigProvider) []ConventionalCommitOption {
			scopes := c.ConventionalScopes()
			if len(scopes) > 0 {
				return []ConventionalCommitOption{WithAllowedScopes(scopes)}
			}

			return nil
		},
		// Collect max description length options
		func(c domain.ConventionalConfigProvider) []ConventionalCommitOption {
			maxLength := c.ConventionalMaxDescriptionLength()
			if maxLength > 0 {
				return []ConventionalCommitOption{WithMaxDescLength(maxLength)}
			}

			return nil
		},
	}

	// Use Map to collect options, then flatten the result
	nestedOptions := contextx.Map(optionCollectors, func(collector func(domain.ConventionalConfigProvider) []ConventionalCommitOption) []ConventionalCommitOption {
		return collector(config)
	})

	// Flatten the nested options
	options := contextx.Reduce(nestedOptions, []ConventionalCommitOption{}, func(acc []ConventionalCommitOption, opts []ConventionalCommitOption) []ConventionalCommitOption {
		if len(opts) > 0 {
			return append(acc, opts...)
		}

		return acc
	})

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

	// Use Reduce to accumulate errors into the BaseRule
	result.BaseRule = contextx.Reduce(
		errors,
		r.BaseRule.WithClearedErrors(),
		func(baseRule BaseRule, err appErrors.ValidationError) BaseRule {
			return baseRule.WithError(err)
		},
	)

	return result
}

// addError adds an error to the rule and returns a new rule instance.
func (r ConventionalCommitRule) addError(err appErrors.ValidationError) ConventionalCommitRule {
	result := r
	result.BaseRule = r.BaseRule.WithError(err)

	return result
}

// validateConventionalWithState validates a commit and returns errors.
// The second return value is for state tracking in more complex implementations.
func validateConventionalWithState(rule ConventionalCommitRule, commit domain.CommitInfo) ([]appErrors.ValidationError, ConventionalCommitRule) { //nolint:unparam
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
		// Generate a rich help message with examples
		helpMessage := `Format Error: Commit message doesn't follow conventional format.

Your commit message doesn't follow the conventional commit format required by this project.

✅ CORRECT FORMAT:
<type>[optional scope][optional !]: <description>

Examples:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions
- perf(api): optimize database queries
- feat(user)!: change user API response format

❌ INCORRECT FORMAT:
- ` + subject + `

WHY THIS MATTERS:
- Conventional commits provide a structured commit history
- They enable automated tools to generate changelogs
- They make it easier to understand the purpose of each commit
- They help categorize changes by type (feature, bugfix, etc.)

NEXT STEPS:
1. Rewrite your commit message following the conventional format
   - Choose an appropriate type from: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
   - Add optional scope in parentheses if relevant (e.g., (auth), (api))
   - Add optional breaking change marker (!) if needed
   - Add colon and a single space
   - Write a clear, concise description in imperative mood

2. Use 'git commit --amend' to edit your most recent commit`

		// Create an enhanced validation error using the helper function
		errorMessage := "commit message doesn't follow conventional format: type(scope)!: description"

		err := appErrors.FormatError(
			updatedRule.Name(),
			errorMessage,
			helpMessage,
			subject,
		)

		updatedRule = updatedRule.addError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Check for spacing issues
	if strings.Contains(subject, ":  ") {
		// Create a suggested correction with proper spacing
		suggestionPattern := `:\s+`
		suggestedForm := regexp.MustCompile(suggestionPattern).ReplaceAllString(subject, ": ")

		helpMessage := fmt.Sprintf(`Spacing Error: Too many spaces after colon in commit message.

Your commit message has too many spaces after the colon. Conventional commits require exactly one space.

✅ CORRECT FORMAT:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions

✅ SUGGESTED CORRECTION:
%s

❌ INCORRECT FORMAT:
- feat:  add user authentication (two spaces after colon)
- fix(auth):   resolve login issue (multiple spaces after colon)

WHY THIS MATTERS:
- Consistent spacing ensures proper parsing by tools
- It maintains readability and uniformity in commit history
- Many automation tools rely on exact spacing in conventional commits
- It helps maintain a professional and organized commit history

NEXT STEPS:
1. Edit your commit message to use exactly one space after the colon
2. Use 'git commit --amend' to modify your most recent commit
3. Check for and remove any extra spaces before saving`, suggestedForm)

		// Create an enhanced validation error using the helper function
		err := appErrors.FormatError(
			updatedRule.Name(),
			"commit message has too many spaces after colon (should be exactly one)",
			helpMessage,
			subject,
		)

		// Add suggested correction to the error context
		err = err.WithContext("suggested_form", suggestedForm)

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
			allowedTypes := strings.Join(updatedRule.allowedTypes, ", ")
			helpMessage := fmt.Sprintf(`Invalid Commit Type Error: "%s" is not an allowed type.

✅ CORRECT TYPES: %s

❌ INCORRECT TYPE: %s

WHY THIS MATTERS:
- Conventional commits require specific, standardized types
- Each type has a specific meaning (feat for features, fix for bugfixes, etc.)
- Consistent types enable automated changelog generation
- They help categorize changes for better organization

NEXT STEPS:
1. Choose an appropriate type from the allowed list
2. Update your commit message with a valid type
3. Use 'git commit --amend' to modify your most recent commit`,
				commitType, allowedTypes, commitType)

			// Use FormatError with the appropriate context
			err := appErrors.FormatError(
				updatedRule.Name(),
				"invalid commit type: "+commitType,
				helpMessage,
				subject,
			)

			// Find the closest valid type for suggestion
			closestType := findClosestType(commitType, updatedRule.allowedTypes)

			// Create a suggested correction if we found a close match
			if closestType != "" {
				// Replace only the type part while keeping the rest intact
				suggestedForm := strings.Replace(subject, commitType, closestType, 1)

				// Add suggestion to the context
				err = err.WithContext("suggested_type", closestType)
				err = err.WithContext("suggested_form", suggestedForm)

				// Update help message with suggestion
				helpText := fmt.Sprintf("Did you mean '%s' instead of '%s'?", closestType, commitType)
				err = err.WithContext("suggestion_text", helpText)
			}

			// Add additional context
			err = err.WithContext("type", commitType)
			err = err.WithContext("allowed_types", allowedTypes)

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}
	}

	// Extract and validate scope
	if scopeIdx >= 0 && scopeIdx < len(matches) {
		scope := matches[scopeIdx]
		updatedRule.scope = scope

		if scope != "" && !updatedRule.isValidScope(scope) {
			allowedScopes := strings.Join(updatedRule.allowedScopes, ", ")
			helpMessage := fmt.Sprintf(`Invalid Scope Error: "%s" is not an allowed scope.

✅ CORRECT SCOPES: %s

❌ INCORRECT SCOPE: %s

WHY THIS MATTERS:
- Scopes indicate which part of the project is affected by changes
- Consistent scopes improve organization and searchability
- They help reviewers understand the affected components
- They provide useful categorization for changelog generation

NEXT STEPS:
1. Choose an appropriate scope from the allowed list
2. Update your commit message with a valid scope
3. Use 'git commit --amend' to modify your most recent commit`,
				scope, allowedScopes, scope)

			// Use FormatError with the appropriate context
			err := appErrors.FormatError(
				updatedRule.Name(),
				"invalid commit scope: "+scope,
				helpMessage,
				subject,
			)

			// Add additional context
			err = err.WithContext("scope", scope)
			err = err.WithContext("allowed_scopes", allowedScopes)

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}
	}

	// Check if scope is required but missing
	if updatedRule.requireScope && (scopeIdx < 0 || scopeIdx >= len(matches) || matches[scopeIdx] == "") {
		helpMessage := `Missing Scope Error: Commit scope is required.

Your commit message is missing a required scope. Scopes must be included in parentheses after the type.

✅ CORRECT FORMAT:
- feat(auth): add user authentication
- fix(api): resolve timeout issue
- docs(readme): update installation instructions

❌ INCORRECT FORMAT:
- feat: add user authentication (missing scope)
- fix: resolve timeout issue (missing scope)

WHY THIS MATTERS:
- Scopes indicate which part of the project is affected by changes
- Required scopes help with organization and categorization
- They provide critical context for code reviewers
- They improve searchability and filtering of commits

NEXT STEPS:
1. Add an appropriate scope in parentheses after the commit type
2. The scope should indicate which component or module is affected
3. Use 'git commit --amend' to modify your most recent commit`

		// Use FormatError with the appropriate context
		err := appErrors.FormatError(
			updatedRule.Name(),
			"commit scope is required but not provided",
			helpMessage,
			subject,
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
			helpMessage := `Empty Description Error: Commit message has no description.

Your commit message is missing a description after the type and scope.

✅ CORRECT FORMAT:
- feat(auth): add user authentication system
- fix(api): resolve timeout issue
- docs: update installation instructions

❌ INCORRECT FORMAT:
- feat(auth): 
- fix: 
- docs(readme): 

WHY THIS MATTERS:
- The description explains what changes were made
- Without a description, the purpose of the commit is unclear
- It's a required part of the conventional commit format
- Clear descriptions improve commit history readability

NEXT STEPS:
1. Add a clear, concise description after the colon
2. Use imperative mood (e.g., "add feature" not "added feature")
3. Keep it under 72 characters if possible
4. Use 'git commit --amend' to modify your most recent commit`

			// Use FormatError with the appropriate context
			err := appErrors.FormatError(
				updatedRule.Name(),
				"commit description cannot be empty",
				helpMessage,
				subject,
			)

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}

		// Check description length
		descriptionLength := len(description)

		// Check description length
		if updatedRule.maxDescLength > 0 && descriptionLength > updatedRule.maxDescLength {
			helpMessage := fmt.Sprintf(`Description Too Long Error: Commit description exceeds maximum length.

Your commit description is %d characters long, but the maximum allowed is %d characters.

✅ CORRECT FORMAT:
- Keep descriptions concise and under %d characters
- Put additional details in the commit body

❌ INCORRECT FORMAT:
- Your description: "%s" (%d characters)

WHY THIS MATTERS:
- Short descriptions improve readability in git logs
- Many tools truncate longer descriptions
- Concise descriptions force focus on the core change
- Details can be added to the commit body

NEXT STEPS:
1. Shorten your description to be more concise
2. Move details to the commit body (after a blank line)
3. Focus on what changed rather than how or why
4. Use 'git commit --amend' to modify your most recent commit`,
				descriptionLength, updatedRule.maxDescLength, updatedRule.maxDescLength,
				description, descriptionLength)

			// Use FormatError with the appropriate context
			err := appErrors.FormatError(
				updatedRule.Name(),
				fmt.Sprintf("commit description is too long (%d chars, max is %d)",
					descriptionLength, updatedRule.maxDescLength),
				helpMessage,
				subject,
			)

			// Add additional context
			err = err.WithContext("length", strconv.Itoa(descriptionLength))
			err = err.WithContext("max_length", strconv.Itoa(updatedRule.maxDescLength))

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}
	}

	return []appErrors.ValidationError{}, updatedRule
}

// Validate validates a commit against the conventional commit rules.
// This method follows functional programming principles and does not modify the rule's state.
func (r ConventionalCommitRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Get the logger from context
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Entering ConventionalCommitRule.Validate")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Use the pure functional approach
	errors, _ := validateConventionalWithState(rule, commit)

	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r ConventionalCommitRule) withContextConfig(ctx context.Context) ConventionalCommitRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Create a copy of the rule
	result := r

	// Only override settings if they are specified in the context configuration
	if len(cfg.Conventional.Types) > 0 {
		result.allowedTypes = deepCopyStringSlice(cfg.Conventional.Types)
	}

	if len(cfg.Conventional.Scopes) > 0 {
		result.allowedScopes = deepCopyStringSlice(cfg.Conventional.Scopes)
	}

	result.requireScope = cfg.Conventional.RequireScope
	result.validateBreaking = cfg.Conventional.AllowBreakingChanges

	if cfg.Conventional.MaxDescriptionLength > 0 {
		result.maxDescLength = cfg.Conventional.MaxDescriptionLength
	} else if result.maxDescLength == 0 && cfg.Subject.MaxLength > 0 {
		// If maxDescLength is not set, use the subject max length from config
		result.maxDescLength = cfg.Subject.MaxLength
	}

	return result
}

// isValidType checks if the commit type is in the list of allowed types.
func (r ConventionalCommitRule) isValidType(commitType string) bool {
	// If no allowed types are specified, all types are allowed
	if len(r.allowedTypes) == 0 {
		return true
	}

	return contextx.Contains(r.allowedTypes, commitType)
}

// isValidScope checks if the commit scope is in the list of allowed scopes.
func (r ConventionalCommitRule) isValidScope(scope string) bool {
	// If no allowed scopes are specified, all scopes are allowed
	if len(r.allowedScopes) == 0 {
		return true
	}

	return contextx.Contains(r.allowedScopes, scope)
}

// Result returns a concise validation result.
func (r ConventionalCommitRule) Result(_ []appErrors.ValidationError) string {
	if r.HasErrors() {
		return "Invalid conventional commit format"
	}

	return "Valid conventional commit format"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r ConventionalCommitRule) VerboseResult(_ []appErrors.ValidationError) string {
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
func (r ConventionalCommitRule) Help(errors []appErrors.ValidationError) string {
	if !r.HasErrors() {
		return "No errors to fix. This rule checks that commits follow the conventional commit format with proper type, structure, and description (e.g., feat: add new login feature, fix(auth): resolve timeout issue)."
	}

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

// findClosestType finds the closest matching valid type from the allowed types list
// using Levenshtein distance. Returns an empty string if no good match is found.
func findClosestType(inputType string, allowedTypes []string) string {
	if len(allowedTypes) == 0 {
		return ""
	}

	inputType = strings.ToLower(inputType)
	minDistance := 3 // Maximum edit distance to consider a good match

	// Filter types by similar length and map to type-distance pairs in a single pass
	typeDistancePairs := contextx.FilterMap(allowedTypes,
		// Filter by similar length
		func(validType string) bool {
			return abs(len(validType)-len(inputType)) <= 2
		},
		// Map to type+distance pairs
		func(validType string) struct {
			typeName string
			distance int
		} {
			return struct {
				typeName string
				distance int
			}{
				typeName: validType,
				distance: levenshteinDistance(inputType, validType),
			}
		})

	if len(typeDistancePairs) == 0 {
		return ""
	}

	// Find the pair with minimum distance
	return contextx.Reduce(typeDistancePairs, struct {
		typeName  string
		minDist   int
		foundGood bool
	}{
		typeName:  "",
		minDist:   minDistance + 1, // Start with a value larger than our threshold
		foundGood: false,
	}, func(acc struct {
		typeName  string
		minDist   int
		foundGood bool
	}, pair struct {
		typeName string
		distance int
	}) struct {
		typeName  string
		minDist   int
		foundGood bool
	} {
		if pair.distance < acc.minDist {
			return struct {
				typeName  string
				minDist   int
				foundGood bool
			}{
				typeName:  pair.typeName,
				minDist:   pair.distance,
				foundGood: pair.distance < minDistance,
			}
		}

		return acc
	}).typeName
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

// levenshteinDistance calculates the Levenshtein (edit) distance between two strings.
func levenshteinDistance(str1, str2 string) int {
	if len(str1) == 0 {
		return len(str2)
	}

	if len(str2) == 0 {
		return len(str1)
	}

	// Create row indices functionally
	rowIndices := contextx.Range(len(str1) + 1)
	colIndices := contextx.Range(len(str2) + 1)

	// Initialize matrix with rows
	matrix := make([][]int, len(str1)+1)

	// Create and initialize each row with first column values
	contextx.ForEach(rowIndices, func(rowIdx int) {
		// Create the row
		row := make([]int, len(str2)+1)
		// Set first column value
		row[0] = rowIdx
		// Store the row
		matrix[rowIdx] = row
	})

	// Initialize first row
	contextx.ForEach(colIndices, func(colIdx int) {
		matrix[0][colIdx] = colIdx
	})

	// Fill in the matrix using a more functional-style approach
	// Create indices for rows and columns, excluding the first row/column (already initialized)
	innerRowIndices := contextx.Range(len(str1))
	innerColIndices := contextx.Range(len(str2))

	// Fill the matrix row by row
	contextx.ForEach(innerRowIndices, func(i int) {
		rowIdx := i + 1 // Adjust index (skip first row)

		// Fill each cell in this row
		contextx.ForEach(innerColIndices, func(j int) {
			colIdx := j + 1 // Adjust index (skip first column)

			// Calculate cost based on character comparison
			cost := 1
			if str1[rowIdx-1] == str2[colIdx-1] {
				cost = 0
			}

			// Calculate the minimum of three operations
			matrix[rowIdx][colIdx] = min3(
				matrix[rowIdx-1][colIdx]+1,      // deletion
				matrix[rowIdx][colIdx-1]+1,      // insertion
				matrix[rowIdx-1][colIdx-1]+cost, // substitution
			)
		})
	})

	return matrix[len(str1)][len(str2)]
}

// min3 returns the minimum of three integers.
func min3(first, second, third int) int {
	return contextx.Reduce([]int{first, second, third}, first, func(minVal int, current int) int {
		if current < minVal {
			return current
		}

		return minVal
	})
}

// Helper function for deep copying string slices.
func deepCopyStringSlice(src []string) []string {
	return contextx.DeepCopy(src)
}

// Note: This file focuses on the production implementation, with any test-specific code moved to test helper files

// ConventionalCommitRuleCtx is a context-aware version of ConventionalCommitRule.
type ConventionalCommitRuleCtx struct {
	BaseRule
}

// NewConventionalCommitRuleCtx creates a new context-aware ConventionalCommitRule.
func NewConventionalCommitRuleCtx() ConventionalCommitRuleCtx {
	return ConventionalCommitRuleCtx{
		BaseRule: NewBaseRule("ConventionalCommit"),
	}
}

// Validate validates that commit messages follow the Conventional Commits specification
// using configuration from context.
func (r ConventionalCommitRuleCtx) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating conventional commit using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Use the state-based validation logic
	errors, _ := validateConventionalWithState(rule, commit)

	return errors
}

// withContextConfig creates a new ConventionalCommitRule with configuration from context.
func (r ConventionalCommitRuleCtx) withContextConfig(ctx context.Context) ConventionalCommitRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Create a rule with configuration from context
	rule := ConventionalCommitRule{
		BaseRule:         r.BaseRule,
		allowedTypes:     deepCopyStringSlice(cfg.Conventional.Types),
		allowedScopes:    deepCopyStringSlice(cfg.Conventional.Scopes),
		requireScope:     cfg.Conventional.RequireScope,
		validateBreaking: cfg.Conventional.AllowBreakingChanges,
		maxDescLength:    cfg.Conventional.MaxDescriptionLength,
	}

	// If maxDescLength is not set, use the subject max length from config
	if rule.maxDescLength == 0 {
		rule.maxDescLength = cfg.Subject.MaxLength
	}

	return rule
}

// Name returns the rule name.
func (r ConventionalCommitRuleCtx) Name() string {
	return r.BaseRule.Name()
}

// Result returns a concise validation result.
func (r ConventionalCommitRuleCtx) Result(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return "✓ Valid conventional format"
	}

	return "Invalid conventional commit format"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r ConventionalCommitRuleCtx) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return "Valid conventional commit format"
	}

	return "Commit message does not follow the Conventional Commits specification"
}

// Help returns guidance for fixing rule violations.
func (r ConventionalCommitRuleCtx) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	return "Your commit message should follow the Conventional Commits format:\n\n" +
		"  <type>[optional scope][!]: <description>\n\n" +
		"Examples:\n" +
		"  feat: add new feature\n" +
		"  fix(api): resolve null pointer exception\n" +
		"  chore!: drop support for Node 6"
}
