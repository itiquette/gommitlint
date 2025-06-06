// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// JiraReferenceRule validates that commit messages reference JIRA issues.
type JiraReferenceRule struct {
	pattern               string
	prefixes              []string
	excludedTypes         []string
	searchInBody          bool
	checkConventionalOnly bool
	requiredForTypes      []string
}

// Name returns the rule name.
func (r JiraReferenceRule) Name() string {
	return "JiraReference"
}

// NewJiraReferenceRule creates a new rule for validating JIRA references from config.
func NewJiraReferenceRule(cfg config.Config) JiraReferenceRule {
	// Set pattern from project prefixes or use default
	var pattern string
	if len(cfg.Jira.ProjectPrefixes) > 0 {
		// Build pattern from project prefixes
		pattern = fmt.Sprintf("(%s)-\\d+", strings.Join(cfg.Jira.ProjectPrefixes, "|"))
	} else {
		pattern = `[A-Z]+-\d+`
	}

	// Check if conventional commit is enabled
	isConventionalEnabled := domain.IsRuleActive("conventional", cfg.Rules.Enabled, cfg.Rules.Disabled)

	return JiraReferenceRule{
		pattern:               pattern,
		prefixes:              cfg.Jira.ProjectPrefixes,
		excludedTypes:         []string{"docs", "chore", "style", "refactor", "test"},
		searchInBody:          cfg.Jira.RequireInBody,
		checkConventionalOnly: isConventionalEnabled,
		requiredForTypes:      []string{},
	}
}

// Helper function to check if a commit message follows conventional commit format.

// Helper function to check if a JIRA reference is in the scope part of a conventional commit.

// Validate checks a commit for Jira reference compliance.
func (r JiraReferenceRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// Check if this commit type should be excluded from JIRA validation
	if shouldExcludeCommitType(commit.Subject, r.excludedTypes) {
		return nil
	}

	// Check if JIRA is required for this commit type
	if !isJiraRequiredForType(commit.Subject, r.requiredForTypes) && r.checkConventionalOnly {
		return nil
	}

	// Prepare the text to search
	textToSearch := commit.Subject
	if r.searchInBody && commit.Body != "" {
		textToSearch = fmt.Sprintf("%s\n%s", commit.Subject, commit.Body)
	}

	// Extract JIRA references
	references := extractJiraReferences(textToSearch, r.pattern, r.prefixes)

	// If no references found and they're required
	if len(references) == 0 {
		expectedPattern := "ABC-123"
		if len(r.prefixes) > 0 {
			expectedPattern = r.prefixes[0] + "-123"
		}

		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingJira, "JIRA reference missing").
				WithHelp("Add a JIRA reference like " + expectedPattern),
		}
	}

	// Validate projects if configured
	if len(r.prefixes) > 0 {
		for _, ref := range references {
			project := extractProjectFromReference(ref)
			if !isValidProject(project, r.prefixes) {
				return []domain.ValidationError{
					domain.New(r.Name(), domain.ErrInvalidProject,
						fmt.Sprintf("Invalid JIRA project '%s'", project)).
						WithHelp("Use one of: " + strings.Join(r.prefixes, ", ")),
				}
			}
		}
	}

	// Check reference placement in conventional commits
	if r.checkConventionalOnly && isConventionalCommit(commit.Subject) {
		conventionalType, _, _ := parseConventionalCommit(commit.Subject)

		// For merge commits, we're more lenient
		if conventionalType == "merge" {
			return nil
		}

		// Check if JIRA is in the correct position (not in description)
		if hasJiraInDescription(commit.Subject, r.pattern) {
			return []domain.ValidationError{
				domain.New(r.Name(), domain.ErrMisplacedJira, "JIRA reference should be in scope, not description").
					WithHelp("Use format: type(JIRA-123): description"),
			}
		}
	}

	return nil
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
