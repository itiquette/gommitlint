// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
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
	errors []error
}

// Name returns the rule identifier.
func (c ConventionalCommit) Name() string {
	return "ConventionalCommit"
}

// Result returns a string representation of the validation result.
func (c ConventionalCommit) Result() string {
	if len(c.errors) > 0 {
		return c.errors[0].Error()
	}

	return "Commit message is a valid conventional commit"
}

// Errors returns all validation errors.
func (c ConventionalCommit) Errors() []error {
	return c.errors
}

// Help returns guidance for fixing rule violations.
func (c ConventionalCommit) Help() string {
	if len(c.errors) == 0 {
		return "No errors to fix"
	}

	errMsg := c.errors[0].Error()

	if strings.Contains(errMsg, "invalid conventional commit format") {
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
	}

	if strings.Contains(errMsg, "invalid type") {
		return `The commit type you used is not in the allowed list of types.
Your commit should use one of the approved types from the allowed list.
Check your project documentation or configuration for the full list of allowed types.`
	}

	if strings.Contains(errMsg, "invalid scope") {
		return `The scope you specified is not in the allowed list of scopes.
Scopes define the section of the codebase your change affects.
Check your project documentation or configuration for the full list of allowed scopes.`
	}

	if strings.Contains(errMsg, "empty description") {
		return `Your commit message is missing a description.
After the type(scope): prefix, you must include a description that explains what the commit does.
Example: feat(ui): add new button component`
	}

	if strings.Contains(errMsg, "description too long") {
		return `Your commit description exceeds the maximum allowed length.
Keep your commit description concise while still being descriptive.
Consider breaking down large changes into multiple smaller commits if possible.`
	}

	if strings.Contains(errMsg, "spacing error") {
		return `There should be exactly one space after the colon in your commit message.
Correct: feat: add feature
Incorrect: feat:add feature or feat:  add feature`
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

// addErrorf adds an error to the rule's errors slice.
func (c *ConventionalCommit) addErrorf(format string, args ...interface{}) {
	c.errors = append(c.errors, fmt.Errorf(format, args...))
}

// AddTestError adds an error to the rule's errors slice (for testing only).
func (c *ConventionalCommit) AddTestError(err error) {
	c.errors = append(c.errors, err)
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
		rule.addErrorf("invalid conventional commit format: empty message")

		return rule
	}

	// Default description length if not specified
	if descLength == 0 {
		descLength = 72
	}

	// Validate basic format first
	if !SubjectRegex.MatchString(subject) {
		rule.addErrorf("invalid conventional commit format: %q", subject)

		return rule
	}

	//Simple check for ": " vs ":  " (one space vs multiple spaces)
	if strings.Contains(subject, ":  ") {
		rule.addErrorf("spacing error: must have exactly one space after colon")

		return rule
	}

	// Parse the subject according to conventional commit format
	matches := SubjectRegex.FindStringSubmatch(subject)
	if len(matches) != 5 {
		rule.addErrorf("invalid conventional commit format: %q", subject)

		return rule
	}

	// Extract components
	commitType := matches[1]
	scope := matches[2]
	description := matches[4]

	// Validate type
	if len(types) > 0 && !slices.Contains(types, commitType) {
		rule.addErrorf("invalid type %q: allowed types are %v", commitType, types)

		return rule
	}

	// Validate scope if provided and scope list is defined
	if scope != "" && len(scopes) > 0 {
		scopesList := strings.Split(scope, ",")
		for _, s := range scopesList {
			if !slices.Contains(scopes, s) {
				rule.addErrorf("invalid scope %q: allowed scopes are %v", s, scopes)

				return rule
			}
		}
	}

	// Validate description content
	if strings.TrimSpace(description) == "" {
		rule.addErrorf("empty description: description must contain non-whitespace characters")

		return rule
	}

	// Validate description length
	if len(description) > descLength {
		rule.addErrorf("description too long: %d characters (max: %d)", len(description), descLength)

		return rule
	}

	return rule
}
