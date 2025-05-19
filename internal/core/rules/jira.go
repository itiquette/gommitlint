// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// JiraReferenceRule validates that commit messages reference JIRA issues.
type JiraReferenceRule struct {
	name                  string
	pattern               string
	prefixes              []string
	excludedTypes         []string
	searchInBody          bool
	checkConventionalOnly bool
	requiredForTypes      []string
}

// Name returns the rule name.
func (r JiraReferenceRule) Name() string {
	return r.name
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

// WithValidProjects sets the valid JIRA project prefixes.
func WithValidProjects(projects []string) JiraReferenceOption {
	return WithJiraPrefixes(projects)
}

// NewJiraReferenceRule creates a new rule for validating JIRA references.
func NewJiraReferenceRule(options ...JiraReferenceOption) JiraReferenceRule {
	// Create rule with default values
	rule := JiraReferenceRule{
		name:          "JiraReference",
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

// Helper function to check if a JIRA reference is in the scope part of a conventional commit.

// Validate checks a commit for Jira reference compliance using context-based configuration.
func (r JiraReferenceRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating Jira references using context configuration", "rule", r.Name(), "commit_hash", commit.Hash)

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Check if this commit type should be excluded from JIRA validation
	if shouldExcludeCommitType(commit.Subject, rule.excludedTypes) {
		return nil
	}

	// Check if JIRA is required for this commit type
	if !isJiraRequiredForType(commit.Subject, rule.requiredForTypes) && rule.checkConventionalOnly {
		return nil
	}

	// Prepare the text to search
	textToSearch := commit.Subject
	if rule.searchInBody && commit.Body != "" {
		textToSearch = fmt.Sprintf("%s\n%s", commit.Subject, commit.Body)
	}

	// Extract JIRA references
	references := extractJiraReferences(textToSearch, rule.pattern, rule.prefixes)

	// If no references found and they're required
	if len(references) == 0 {
		return []appErrors.ValidationError{
			appErrors.New(
				"JiraReference",
				appErrors.ErrMissingJira,
				"commit message must reference a JIRA issue",
			),
		}
	}

	// Validate projects if configured
	if len(rule.prefixes) > 0 {
		for _, ref := range references {
			project := extractProjectFromReference(ref)
			if !isValidProject(project, rule.prefixes) {
				return []appErrors.ValidationError{
					appErrors.New(
						"JiraReference",
						appErrors.ErrInvalidProject,
						fmt.Sprintf("invalid JIRA project '%s'", project),
					).WithContext("project", project).
						WithContext("valid_projects", strings.Join(rule.prefixes, ", ")),
				}
			}
		}
	}

	// Check reference placement in conventional commits
	if rule.checkConventionalOnly && isConventionalCommit(commit.Subject) {
		conventionalType, _, _ := parseConventionalCommit(commit.Subject)

		// For merge commits, we're more lenient
		if conventionalType == "merge" {
			return nil
		}

		// Check if JIRA is in the correct position (not in description)
		if hasJiraInDescription(commit.Subject, rule.pattern) {
			return []appErrors.ValidationError{
				appErrors.New(
					"JiraReference",
					appErrors.ErrMisplacedJira,
					"JIRA reference should be in scope, not description",
				),
			}
		}
	}

	return nil
}

// withContextConfig creates a new rule with configuration from context.
func (r JiraReferenceRule) withContextConfig(ctx context.Context) JiraReferenceRule {
	// Get configuration directly from context
	cfg := contextx.GetConfig(ctx)

	// Extract configuration values
	isConventional := cfg.GetBool("conventional.required")
	validateBodyRef := cfg.GetBool("jira.body_ref")
	validProjects := cfg.GetStringSlice("jira.projects")
	pattern := cfg.GetString("jira.pattern")

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

	if pattern != "" {
		result.pattern = pattern
	}

	// Log configuration at debug level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Jira reference rule configuration from context",
		"conventional_required", isConventional,
		"body_ref_checking", validateBodyRef,
		"valid_projects", validProjects,
		"pattern", pattern)

	return result
}

// validateJiraWithState validates the JIRA references and returns both the errors and an updated rule.

// extractCommitType extracts the type from a conventional commit message.

// isExcludedType checks if a commit type is in the excluded list.

// findJiraReferences finds all JIRA references in a text.

// SetErrors is no longer used since we don't have baseRule.
// Validation errors are returned directly from the Validate method.

// shouldExcludeCommitType checks if a commit type should be excluded from JIRA validation.
func shouldExcludeCommitType(subject string, excludedTypes []string) bool {
	if len(excludedTypes) == 0 {
		return false
	}

	// Extract type from conventional commit format
	regex := regexp.MustCompile(`^(\w+)(?:\([^)]*\))?!?:`)
	matches := regex.FindStringSubmatch(subject)

	if len(matches) > 1 {
		commitType := matches[1]
		for _, excluded := range excludedTypes {
			if commitType == excluded {
				return true
			}
		}
	}

	return false
}

// isJiraRequiredForType checks if JIRA reference is required for a commit type.
func isJiraRequiredForType(subject string, requiredTypes []string) bool {
	if len(requiredTypes) == 0 {
		return true // Required for all types if not specified
	}

	// Extract type from conventional commit format
	regex := regexp.MustCompile(`^(\w+)(?:\([^)]*\))?!?:`)
	matches := regex.FindStringSubmatch(subject)

	if len(matches) > 1 {
		commitType := matches[1]
		for _, required := range requiredTypes {
			if commitType == required {
				return true
			}
		}
	}

	return false
}

// extractJiraReferences extracts JIRA references from text.
func extractJiraReferences(text, pattern string, prefixes []string) []string {
	var references []string

	// Use custom pattern if provided
	if pattern != "" {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllString(text, -1)
		references = append(references, matches...)
	} else {
		// Default pattern: PROJECT-123 format
		defaultPattern := `\b[A-Z]+-\d+\b`
		regex := regexp.MustCompile(defaultPattern)
		matches := regex.FindAllString(text, -1)

		// Filter by prefixes if provided
		if len(prefixes) > 0 {
			for _, match := range matches {
				for _, prefix := range prefixes {
					if strings.HasPrefix(match, prefix+"-") {
						references = append(references, match)

						break
					}
				}
			}
		} else {
			references = matches
		}
	}

	// Also check for body references with explicit markers
	bodyRefPattern := `(?i)(?:refs?|references?|see|fixes?|closes?):\s*([A-Z]+-\d+(?:\s*,\s*[A-Z]+-\d+)*)`
	bodyRe := regexp.MustCompile(bodyRefPattern)
	bodyMatches := bodyRe.FindAllStringSubmatch(text, -1)

	for _, match := range bodyMatches {
		if len(match) > 1 {
			// Split multiple references
			refs := regexp.MustCompile(`[A-Z]+-\d+`).FindAllString(match[1], -1)
			references = append(references, refs...)
		}
	}

	// Deduplicate references
	seen := make(map[string]bool)

	var unique []string

	for _, ref := range references {
		if !seen[ref] {
			seen[ref] = true

			unique = append(unique, ref)
		}
	}

	return unique
}

// extractProjectFromReference extracts the project key from a JIRA reference.
func extractProjectFromReference(reference string) string {
	parts := strings.Split(reference, "-")
	if len(parts) >= 2 {
		return parts[0]
	}

	return ""
}

// isValidProject checks if a project key is in the valid projects list.
func isValidProject(project string, validProjects []string) bool {
	for _, valid := range validProjects {
		if project == valid {
			return true
		}
	}

	return false
}

// isConventionalCommit checks if a subject follows conventional commit format.
func isConventionalCommit(subject string) bool {
	regex := regexp.MustCompile(`^[a-z]+(?:\([^)]*\))?!?:`)

	return regex.MatchString(subject)
}

// parseConventionalCommit parses a conventional commit and returns type, scope, and description.
func parseConventionalCommit(subject string) (string, string, string) {
	regex := regexp.MustCompile(`^(\w+)(?:\(([^)]*)\))?!?:\s*(.*)`)
	matches := regex.FindStringSubmatch(subject)

	if len(matches) >= 4 {
		return matches[1], matches[2], matches[3]
	}

	return "", "", ""
}

// hasJiraInDescription checks if JIRA reference is ONLY in the description and NOT at the end.
// Returns true if JIRA is misplaced (not at the end).
func hasJiraInDescription(subject string, pattern string) bool {
	conventionalType, scope, description := parseConventionalCommit(subject)

	if conventionalType == "" {
		return false
	}

	// Check if JIRA is in the scope - this is misplaced
	regex := regexp.MustCompile(pattern)
	if scope != "" && regex.MatchString(scope) {
		return true
	}

	// Check if JIRA is in the middle of the description (not at the end)
	if regex.MatchString(description) {
		matches := regex.FindAllStringIndex(description, -1)
		for _, match := range matches {
			// If there's text after the JIRA reference, it's misplaced
			if match[1] < len(description) && strings.TrimSpace(description[match[1]:]) != "" {
				return true
			}
		}
	}

	return false
}
