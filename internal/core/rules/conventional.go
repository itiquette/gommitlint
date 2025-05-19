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

		config.ScopeRequired = cfg.GetBool("conventional.scope_required")

		if maxDescLen := cfg.GetInt("conventional.max_description_length"); maxDescLen > 0 {
			config.MaxDescLength = maxDescLen
		} else if config.MaxDescLength == 0 && cfg.GetInt("subject.max_length") > 0 {
			// If maxDescLength is not set, use the subject max length from config
			config.MaxDescLength = cfg.GetInt("subject.max_length")
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
			appErrors.New(
				"ConventionalCommit",
				appErrors.ValidationErrorCode("invalid_format"),
				"commit message doesn't follow conventional format: type(scope)!: description",
			).WithContext("subject", subject),
		}

		return errors
	}

	// Validate type
	if !isValidType(parts.Type, config.Types) {
		errors := []appErrors.ValidationError{
			appErrors.New(
				"ConventionalCommit",
				appErrors.ValidationErrorCode("invalid_type"),
				fmt.Sprintf("commit type '%s' is not in allowed types", parts.Type),
			).WithContext("type", parts.Type).
				WithContext("allowed_types", strings.Join(config.Types, ", ")),
		}

		return errors
	}

	// Validate scope if required
	if config.ScopeRequired && parts.Scope == "" {
		errors := []appErrors.ValidationError{
			appErrors.New(
				"ConventionalCommit",
				appErrors.ValidationErrorCode("missing_scope"),
				"commit message must include a scope",
			),
		}

		return errors
	}

	// Validate allowed scopes if specified
	if parts.Scope != "" && len(config.Scopes) > 0 && !isValidScope(parts.Scope, config.Scopes) {
		errors := []appErrors.ValidationError{
			appErrors.New(
				"ConventionalCommit",
				appErrors.ValidationErrorCode("invalid_scope"),
				fmt.Sprintf("commit scope '%s' is not in allowed scopes", parts.Scope),
			).WithContext("scope", parts.Scope).
				WithContext("allowed_scopes", strings.Join(config.Scopes, ", ")),
		}

		return errors
	}

	// Validate description length
	if config.MaxDescLength > 0 && len(parts.Description) > config.MaxDescLength {
		errors := []appErrors.ValidationError{
			appErrors.New(
				"ConventionalCommit",
				appErrors.ValidationErrorCode("description_too_long"),
				fmt.Sprintf("description length %d exceeds maximum %d", len(parts.Description), config.MaxDescLength),
			).WithContext("length", strconv.Itoa(len(parts.Description))).
				WithContext("max_length", strconv.Itoa(config.MaxDescLength)),
		}

		return errors
	}

	// All validations passed
	errors := []appErrors.ValidationError{}

	// Convert the generic validation errors to rich errors with help messages
	if len(errors) > 0 {
		err := errors[0]
		subject := err.Context["subject"]

		switch err.Code {
		case "invalid_format":
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

			return []appErrors.ValidationError{
				appErrors.NewFormatValidationError(r.Name(), err.Message, helpMessage, subject),
			}

		case "too_many_spaces":
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

			richErr := appErrors.NewFormatValidationError(r.Name(), err.Message, helpMessage, subject)
			richErr = richErr.WithContext("suggested_form", suggestedForm)

			return []appErrors.ValidationError{richErr}

		case "invalid_type":
			commitType := err.Context["type"]
			allowedTypesStr := err.Context["allowed_types"]

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
				commitType, allowedTypesStr, commitType)

			richErr := appErrors.NewFormatValidationError(r.Name(), err.Message, helpMessage, subject)

			// Find the closest valid type for suggestion
			closestType := findClosestType(commitType, strings.Split(allowedTypesStr, ", "))
			if closestType != "" {
				// Replace only the type part while keeping the rest intact
				suggestedForm := strings.Replace(subject, commitType, closestType, 1)

				// Add suggestion to the context
				richErr = richErr.WithContext("suggested_type", closestType)
				richErr = richErr.WithContext("suggested_form", suggestedForm)

				// Update help message with suggestion
				helpText := fmt.Sprintf("Did you mean '%s' instead of '%s'?", closestType, commitType)
				richErr = richErr.WithContext("suggestion_text", helpText)
			}

			richErr = richErr.WithContext("type", commitType)
			richErr = richErr.WithContext("allowed_types", allowedTypesStr)

			return []appErrors.ValidationError{richErr}

		case "missing_scope":
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

			return []appErrors.ValidationError{
				appErrors.NewFormatValidationError(r.Name(), err.Message, helpMessage, subject),
			}

		case "invalid_scope":
			scope := err.Context["scope"]
			allowedScopesStr := err.Context["allowed_scopes"]

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
				scope, allowedScopesStr, scope)

			richErr := appErrors.NewFormatValidationError(r.Name(), err.Message, helpMessage, subject)
			richErr = richErr.WithContext("scope", scope)
			richErr = richErr.WithContext("allowed_scopes", allowedScopesStr)

			return []appErrors.ValidationError{richErr}

		case "empty_description":
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

			return []appErrors.ValidationError{
				appErrors.NewFormatValidationError(r.Name(), err.Message, helpMessage, subject),
			}

		case "description_too_long":
			length := err.Context["length"]
			maxLength := err.Context["max_length"]
			description := subject[strings.Index(subject, ":")+1:]
			description = strings.TrimSpace(description)

			helpMessage := fmt.Sprintf(`Description Too Long Error: Commit description exceeds maximum length.

Your commit description is %s characters long, but the maximum allowed is %s characters.

✅ CORRECT FORMAT:
- Keep descriptions concise and under %s characters
- Put additional details in the commit body

❌ INCORRECT FORMAT:
- Your description: "%s" (%s characters)

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
				length, maxLength, maxLength,
				description, length)

			richErr := appErrors.NewFormatValidationError(r.Name(), err.Message, helpMessage, subject)
			richErr = richErr.WithContext("length", length)
			richErr = richErr.WithContext("max_length", maxLength)

			return []appErrors.ValidationError{richErr}
		}

		// Default case: return the error as-is
		return errors
	}

	// With functional approach, we don't modify the rule's state
	// Just return an empty error slice to indicate success
	return []appErrors.ValidationError{}
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

// findClosestType finds the closest matching valid type from the allowed types list
// using Levenshtein distance. Returns an empty string if no good match is found.
func findClosestType(inputType string, allowedTypes []string) string {
	if len(allowedTypes) == 0 {
		return ""
	}

	inputType = strings.ToLower(inputType)
	minDistance := 3 // Maximum edit distance to consider a good match

	// Filter types by similar length and map to type-distance pairs in a single pass
	typeDistancePairs := slices.FilterMap(allowedTypes,
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
	return slices.Reduce(typeDistancePairs, struct {
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
	rowIndices := slices.Range(len(str1) + 1)
	colIndices := slices.Range(len(str2) + 1)

	// Initialize matrix with rows
	matrix := make([][]int, len(str1)+1)

	// Create and initialize each row with first column values
	slices.ForEach(rowIndices, func(rowIdx int) {
		// Create the row
		row := make([]int, len(str2)+1)
		// Set first column value
		row[0] = rowIdx
		// Store the row
		matrix[rowIdx] = row
	})

	// Initialize first row
	slices.ForEach(colIndices, func(colIdx int) {
		matrix[0][colIdx] = colIdx
	})

	// Fill in the matrix using a more functional-style approach
	// Create indices for rows and columns, excluding the first row/column (already initialized)
	innerRowIndices := slices.Range(len(str1))
	innerColIndices := slices.Range(len(str2))

	// Fill the matrix row by row
	slices.ForEach(innerRowIndices, func(i int) {
		rowIdx := i + 1 // Adjust index (skip first row)

		// Fill each cell in this row
		slices.ForEach(innerColIndices, func(j int) {
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
	return slices.Reduce([]int{first, second, third}, first, func(minVal int, current int) int {
		if current < minVal {
			return current
		}

		return minVal
	})
}

// Helper function for deep copying string slices.
func deepCopyStringSlice(src []string) []string {
	return slices.DeepCopy(src)
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
