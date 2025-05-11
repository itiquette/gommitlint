// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// JiraReferenceRule validates that commit messages reference JIRA issues.
type JiraReferenceRule struct {
	baseRule              BaseRule
	pattern               string
	prefixes              []string
	excludedTypes         []string
	references            []string
	searchInBody          bool
	checkConventionalOnly bool
}

// Name returns the rule name.
func (r JiraReferenceRule) Name() string {
	return r.baseRule.Name()
}

// Errors returns all validation errors found by this rule.
func (r JiraReferenceRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r JiraReferenceRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// JiraReferenceOption configures a JiraReferenceRule.
type JiraReferenceOption func(JiraReferenceRule) JiraReferenceRule

// WithJiraPattern sets the regex pattern for JIRA references.
func WithJiraPattern(pattern string) JiraReferenceOption {
	return func(r JiraReferenceRule) JiraReferenceRule {
		result := r
		result.pattern = pattern

		return result
	}
}

// WithJiraPrefixes sets the allowed JIRA prefixes.
func WithJiraPrefixes(prefixes []string) JiraReferenceOption {
	return func(r JiraReferenceRule) JiraReferenceRule {
		result := r
		// Create a deep copy of the prefixes slice
		if len(prefixes) > 0 {
			result.prefixes = make([]string, len(prefixes))
			copy(result.prefixes, prefixes)
		}

		return result
	}
}

// WithJiraExcludedTypes sets commit types that don't require JIRA references.
func WithJiraExcludedTypes(types []string) JiraReferenceOption {
	return func(r JiraReferenceRule) JiraReferenceRule {
		result := r
		// Create a deep copy of the types slice
		if len(types) > 0 {
			result.excludedTypes = make([]string, len(types))
			copy(result.excludedTypes, types)
		}

		return result
	}
}

// WithJiraBodySearch enables searching for JIRA references in the commit body.
func WithJiraBodySearch(search bool) JiraReferenceOption {
	return func(r JiraReferenceRule) JiraReferenceRule {
		result := r
		result.searchInBody = search

		return result
	}
}

// WithConventionalCommit configures the rule to work with conventional commits.
func WithConventionalCommit() JiraReferenceOption {
	return func(r JiraReferenceRule) JiraReferenceRule {
		result := r
		result.checkConventionalOnly = true

		return result
	}
}

// WithBodyRefChecking is an alias for WithJiraBodySearch for backward compatibility.
func WithBodyRefChecking() JiraReferenceOption {
	return WithJiraBodySearch(true)
}

// WithValidProjects sets the valid JIRA project prefixes.
func WithValidProjects(projects []string) JiraReferenceOption {
	return WithJiraPrefixes(projects)
}

// NewJiraReferenceRule creates a new rule for validating JIRA references.
func NewJiraReferenceRule(options ...JiraReferenceOption) JiraReferenceRule {
	// Create rule with default values
	rule := JiraReferenceRule{
		baseRule:      NewBaseRule("JiraReference"),
		pattern:       `[A-Z]+-\d+`,
		prefixes:      []string{},
		excludedTypes: []string{"docs", "chore", "style", "refactor", "test"},
		searchInBody:  true,
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Helper function to check if a commit message follows conventional commit format.
func isConventionalCommit(message string) bool {
	conventionalRegex := regexp.MustCompile(`^(?P<type>\w+)(?:\([^)]*\))?(?:!)?:`)

	return conventionalRegex.MatchString(message)
}

// Helper function to check if a JIRA reference is in the scope part of a conventional commit.
func isScopeJiraReference(message string) bool {
	// This matches patterns like feat(PROJ-123): message or feat(scope PROJ-123): message
	// or feat(scope-PROJ-123): message - any pattern where JIRA is in the scope
	scopeRegex := regexp.MustCompile(`^(?P<type>\w+)\((?:[^)]*)?([A-Z]+-\d+)(?:[^)]*)\):`)

	return scopeRegex.MatchString(message)
}

// Validate checks a commit for Jira reference compliance using context-based configuration.
func (r JiraReferenceRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating Jira references using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Add check for JIRA in middle of conventional commit if both are required
	if rule.checkConventionalOnly && isConventionalCommit(commit.Subject) {
		// Check if JIRA reference is in the middle (scope) rather than at the end
		if isScopeJiraReference(commit.Subject) {
			tempRule := rule
			validationErr := appErrors.CreateBasicError(
				tempRule.baseRule.Name(),
				appErrors.ErrMisplacedJira,
				"JIRA reference should be at the end of the commit message, not in the scope",
			).WithContext("subject", commit.Subject)
			tempRule.baseRule = tempRule.baseRule.WithError(validationErr)

			return tempRule.baseRule.Errors()
		}
	}

	// Use the existing validation logic
	errors, _ := validateJiraWithState(rule, commit)

	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r JiraReferenceRule) withContextConfig(ctx context.Context) JiraReferenceRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Extract configuration values
	isConventional := cfg.Conventional.Required
	validateBodyRef := cfg.Jira.BodyRef
	validProjects := cfg.Jira.Projects

	// Create a copy of the rule
	result := r

	// Apply configuration settings
	if isConventional {
		result.checkConventionalOnly = isConventional
	}

	if validateBodyRef {
		result.searchInBody = validateBodyRef
	}

	if len(validProjects) > 0 {
		result.prefixes = make([]string, len(validProjects))
		copy(result.prefixes, validProjects)
	}

	// Log configuration at debug level
	logger := log.Logger(ctx)
	logger.Debug().
		Bool("conventional_required", isConventional).
		Bool("body_ref_checking", validateBodyRef).
		Strs("valid_projects", validProjects).
		Msg("Jira reference rule configuration from context")

	return result
}

// validateJiraWithState validates the JIRA references and returns both the errors and an updated rule.
func validateJiraWithState(rule JiraReferenceRule, commit domain.CommitInfo) ([]appErrors.ValidationError, JiraReferenceRule) {
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Skip validation for merge commits
	if commit.IsMergeCommit {
		return result.baseRule.Errors(), result
	}

	// Check if this is a conventional commit type that's excluded
	commitType := extractCommitType(commit.Subject)
	if isExcludedType(commitType, rule.excludedTypes) {
		return result.baseRule.Errors(), result
	}

	// Find JIRA references
	var references []string

	// Check subject
	subjectRefs := findJiraReferences(commit.Subject, rule.pattern, rule.prefixes)
	references = append(references, subjectRefs...)

	// Check body if enabled
	if rule.searchInBody && commit.Body != "" {
		bodyRefs := findJiraReferences(commit.Body, rule.pattern, rule.prefixes)
		references = append(references, bodyRefs...)
	}

	// Update the rule with found references
	result.references = references

	// Validate that at least one reference was found
	if len(references) == 0 {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrMissingJira,
			"Commit message does not reference a JIRA ticket",
		)

		if len(rule.prefixes) > 0 {
			validationErr = validationErr.WithContext("allowed_prefixes", strings.Join(rule.prefixes, ", "))
		}

		result.baseRule = result.baseRule.WithError(validationErr)
	}

	return result.baseRule.Errors(), result
}

// extractCommitType extracts the type from a conventional commit message.
func extractCommitType(subject string) string {
	// Conventional commit format: type(scope): description
	conventionalRegex := regexp.MustCompile(`^(?P<type>\w+)(?:\([^)]*\))?(?:!)?:`)

	matches := conventionalRegex.FindStringSubmatch(subject)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// isExcludedType checks if a commit type is in the excluded list.
func isExcludedType(commitType string, excludedTypes []string) bool {
	if commitType == "" {
		return false
	}

	for _, excluded := range excludedTypes {
		if commitType == excluded {
			return true
		}
	}

	return false
}

// findJiraReferences finds all JIRA references in a text.
func findJiraReferences(text string, pattern string, prefixes []string) []string {
	// Compile regex
	regex := regexp.MustCompile(pattern)

	// Find all matches
	allMatches := regex.FindAllString(text, -1)

	// If no prefixes specified, return all matches
	if len(prefixes) == 0 {
		return allMatches
	}

	// Filter by prefixes
	var filtered []string

	for _, match := range allMatches {
		for _, prefix := range prefixes {
			if strings.HasPrefix(match, prefix+"-") {
				filtered = append(filtered, match)

				break
			}
		}
	}

	return filtered
}

// SetErrors sets the errors for this rule and returns an updated rule.
func (r JiraReferenceRule) SetErrors(errors []appErrors.ValidationError) JiraReferenceRule {
	result := r
	result.baseRule = result.baseRule.WithClearedErrors()

	for _, err := range errors {
		result.baseRule = result.baseRule.WithError(err)
	}

	return result
}

// Result returns a concise validation result.
func (r JiraReferenceRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		return "❌ No JIRA reference found"
	}

	return "✓ Found JIRA: " + strings.Join(r.references, ", ")
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r JiraReferenceRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		msg := "❌ Commit message does not reference a JIRA ticket"
		if len(r.prefixes) > 0 {
			msg += " with allowed prefix(es): " + strings.Join(r.prefixes, ", ")
		}

		return msg
	}

	return "✓ Found JIRA reference(s): " + strings.Join(r.references, ", ")
}

// Help returns guidance for fixing rule violations.
func (r JiraReferenceRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	message := "Your commit message should reference a JIRA ticket.\n\n"

	if len(r.prefixes) > 0 {
		message += fmt.Sprintf("The JIRA reference should start with one of these prefixes: %s\n",
			strings.Join(r.prefixes, ", "))
	} else {
		message += "The JIRA reference should follow the pattern: PROJECT-123\n"
	}

	message += "\nExamples:\n"
	message += "- \"fix: handle null pointer exception (ABC-123)\"\n"
	message += "- \"ABC-123: update documentation\"\n"
	message += "- \"feat: add new feature\n\nImplements ABC-123\"\n"

	if len(r.excludedTypes) > 0 {
		message += "\nNote: These commit types don't require JIRA references: " + strings.Join(r.excludedTypes, ", ")
	}

	return message
}
