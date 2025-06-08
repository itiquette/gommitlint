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
	requireInSubject      bool
	ignoreTicketPatterns  []string
	checkConventionalOnly bool
	requiredForTypes      []string
}

// Name returns the rule name.
func (r JiraReferenceRule) Name() string {
	return "JiraReference"
}

// NewJiraReferenceRule creates a new rule for validating JIRA references from config.
func NewJiraReferenceRule(cfg config.Config) JiraReferenceRule {
	// Always use general pattern to catch all JIRA-like references
	// Project validation happens in the validation logic
	pattern := `[A-Z]+-\d+`

	// Check if conventional commit is enabled
	isConventionalEnabled := domain.IsRuleActive("conventional", cfg.Rules.Enabled, cfg.Rules.Disabled)

	return JiraReferenceRule{
		pattern:               pattern,
		prefixes:              cfg.Jira.ProjectPrefixes,
		excludedTypes:         []string{"docs", "chore", "style", "refactor", "test"},
		searchInBody:          cfg.Jira.RequireInBody,
		requireInSubject:      cfg.Jira.RequireInSubject,
		ignoreTicketPatterns:  cfg.Jira.IgnoreTicketPatterns,
		checkConventionalOnly: isConventionalEnabled,
		requiredForTypes:      []string{},
	}
}

// Validate checks a commit for Jira reference compliance.
func (r JiraReferenceRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// Check if this commit type should be excluded from JIRA validation
	if r.shouldExcludeCommitType(commit.Subject) {
		return nil
	}

	// Check if JIRA is required for this commit type
	if !r.isJiraRequiredForType(commit.Subject) && r.checkConventionalOnly {
		return nil
	}

	var errors []domain.ValidationError

	// Subject validation using inline functions
	if r.requireInSubject {
		subjectErrors := r.validateSubjectJira(commit.Subject)
		errors = append(errors, subjectErrors...)
	}

	// Body validation using inline functions
	if r.searchInBody && commit.Body != "" {
		bodyErrors := r.validateBodyJira(commit.Body)
		errors = append(errors, bodyErrors...)
	} else if r.searchInBody && commit.Body == "" && !r.requireInSubject {
		// Body validation is required but no body present
		errors = append(errors,
			domain.New(r.Name(), domain.ErrMissingJiraKeyBody, "No JIRA issue key found in commit body with 'Refs:' prefix").
				WithContextMap(map[string]string{
					"expected": "Refs: " + r.getExpectedFormat(),
				}).
				WithHelp("Add a JIRA reference in the body using format: Refs: "+r.getExpectedFormat()))
	}

	// Fallback: if no specific placement validation is configured,
	// validate basic JIRA presence and project restrictions
	if !r.requireInSubject && !r.searchInBody {
		// Extract JIRA references using basic pattern matching
		subjectReferences := r.extractJiraReferences(commit.Subject)
		bodyReferences := r.extractJiraReferences(commit.Body)

		// Filter out ignored patterns
		subjectReferences = r.filterIgnoredPatterns(subjectReferences)
		bodyReferences = r.filterIgnoredPatterns(bodyReferences)

		allReferences := append(subjectReferences, bodyReferences...)

		if len(allReferences) == 0 {
			expectedPattern := "ABC-123"
			if len(r.prefixes) > 0 {
				expectedPattern = r.prefixes[0] + "-123"
			}

			errors = append(errors,
				domain.New(r.Name(), domain.ErrMissingJira, "Missing JIRA reference").
					WithContextMap(map[string]string{
						"expected": expectedPattern,
					}).
					WithHelp("Add a JIRA reference like "+expectedPattern))
		} else {
			// Validate projects if configured
			if len(r.prefixes) > 0 {
				for _, ref := range allReferences {
					project := r.extractProjectFromReference(ref)
					if !r.isValidProject(project) {
						errors = append(errors,
							domain.New(r.Name(), domain.ErrInvalidProject,
								fmt.Sprintf("Invalid project '%s' in reference '%s'", project, ref)).
								WithContextMap(map[string]string{
									"actual":   project,
									"expected": strings.Join(r.prefixes, ", "),
								}).
								WithHelp("Use one of these projects: "+strings.Join(r.prefixes, ", ")))
					}
				}
			}
		}
	}

	return errors
}

// shouldExcludeCommitType checks if a commit type should be excluded from JIRA validation.
func (r JiraReferenceRule) shouldExcludeCommitType(subject string) bool {
	if len(r.excludedTypes) == 0 {
		return false
	}

	// Use shared conventional commit parser
	parsed := domain.ParseConventionalCommit(subject)
	if parsed.IsValid {
		for _, excluded := range r.excludedTypes {
			if parsed.Type == excluded {
				return true
			}
		}
	}

	return false
}

// isJiraRequiredForType checks if JIRA reference is required for a commit type.
func (r JiraReferenceRule) isJiraRequiredForType(subject string) bool {
	if len(r.requiredForTypes) == 0 {
		return true // Required for all types if not specified
	}

	// Use shared conventional commit parser
	parsed := domain.ParseConventionalCommit(subject)
	if parsed.IsValid {
		for _, required := range r.requiredForTypes {
			if parsed.Type == required {
				return true
			}
		}
	}

	return false
}

// extractJiraReferences extracts JIRA references from text.
func (r JiraReferenceRule) extractJiraReferences(text string) []string {
	var references []string

	// Use configured pattern
	regex := regexp.MustCompile(r.pattern)
	matches := regex.FindAllString(text, -1)
	references = append(references, matches...)

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
	unique := make([]string, 0, len(references))

	for _, ref := range references {
		if !seen[ref] {
			unique = append(unique, ref)
			seen[ref] = true
		}
	}

	return unique
}

// extractProjectFromReference extracts the project key from a JIRA reference.
func (r JiraReferenceRule) extractProjectFromReference(reference string) string {
	parts := strings.Split(reference, "-")
	if len(parts) >= 2 {
		return parts[0]
	}

	return ""
}

// isValidProject checks if a project key is in the valid projects list.
func (r JiraReferenceRule) isValidProject(project string) bool {
	for _, valid := range r.prefixes {
		if project == valid {
			return true
		}
	}

	return false
}

// filterIgnoredPatterns filters out JIRA references that match ignore patterns.
func (r JiraReferenceRule) filterIgnoredPatterns(references []string) []string {
	if len(r.ignoreTicketPatterns) == 0 {
		return references
	}

	var filtered []string

	for _, ref := range references {
		shouldIgnore := false

		for _, pattern := range r.ignoreTicketPatterns {
			// Use pattern matching to check if reference should be ignored
			if matched, err := regexp.MatchString(pattern, ref); err == nil && matched {
				shouldIgnore = true

				break
			}
		}

		if !shouldIgnore {
			filtered = append(filtered, ref)
		}
	}

	return filtered
}

// validateSubjectJira validates JIRA key placement in commit subjects.
func (r JiraReferenceRule) validateSubjectJira(subject string) []domain.ValidationError {
	if subject == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrEmptySubject, "Commit subject is empty").
				WithContextMap(map[string]string{
					"actual":   "empty",
					"expected": "subject with JIRA reference",
				}).
				WithHelp("Provide a non-empty commit subject with a JIRA reference"),
		}
	}

	isConventional := r.isConventionalCommit(subject)
	jiraRefs := r.extractJiraReferences(subject)

	// Filter ignored patterns
	jiraRefs = r.filterIgnoredPatterns(jiraRefs)

	if len(jiraRefs) == 0 {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingJiraKeySubject, "No JIRA issue key found in the commit subject").
				WithContextMap(map[string]string{
					"actual":   "no JIRA reference",
					"expected": r.getExpectedFormat(),
				}).
				WithHelp("Add a JIRA reference like " + r.getExpectedFormat() + " to the commit subject"),
		}
	}

	// For conventional commits, validate placement at end
	if isConventional {
		return r.validateConventionalPlacement(subject, jiraRefs)
	}

	// For non-conventional commits, validate project prefixes
	return r.validateProjectPrefixes(jiraRefs)
}

// validateBodyJira validates JIRA references in commit body using strict "Refs:" format.
func (r JiraReferenceRule) validateBodyJira(body string) []domain.ValidationError {
	if body == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingJiraKeyBody, "No JIRA issue key found in commit body with 'Refs:' prefix").
				WithContextMap(map[string]string{
					"actual":   "no body",
					"expected": "Refs: " + r.getExpectedFormat(),
				}).
				WithHelp("Add a JIRA reference in the body using format: Refs: " + r.getExpectedFormat()),
		}
	}

	// Parse Refs lines from body
	refsLines, lineNumbers := r.parseRefsLines(body)

	if len(refsLines) == 0 {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingJiraKeyBody, "No JIRA issue key found in commit body with 'Refs:' prefix").
				WithContextMap(map[string]string{
					"actual":   "no Refs line",
					"expected": "Refs: " + r.getExpectedFormat(),
				}).
				WithHelp("Add a JIRA reference in the body using format: Refs: " + r.getExpectedFormat()),
		}
	}

	var errors []domain.ValidationError

	// Validate each Refs line format
	for i, refsLine := range refsLines {
		lineNum := lineNumbers[i]

		if formatErrors := r.validateRefsLineFormat(refsLine, lineNum); len(formatErrors) > 0 {
			errors = append(errors, formatErrors...)
		}
	}

	// Validate ordering (Refs must come before Signed-off-by)
	if orderingErrors := r.validateRefsOrdering(body, lineNumbers); len(orderingErrors) > 0 {
		errors = append(errors, orderingErrors...)
	}

	return errors
}

// validateConventionalPlacement validates JIRA key placement in conventional commits.
func (r JiraReferenceRule) validateConventionalPlacement(subject string, jiraRefs []string) []domain.ValidationError {
	// For conventional commits, JIRA key can be:
	// 1. In the scope: feat(PROJ-123): description
	// 2. At the end: feat: description PROJ-123
	if !r.isJiraInValidConventionalPosition(subject, jiraRefs) {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrJiraKeyNotAtEnd, "JIRA key must be at the end of conventional commit subject line").
				WithContextMap(map[string]string{
					"actual":   "JIRA key misplaced",
					"expected": "JIRA key at end of subject",
				}).
				WithHelp("Move JIRA key to the end: 'feat(scope): description " + r.getExpectedFormat() + "'"),
		}
	}

	// Validate project prefixes
	return r.validateProjectPrefixes(jiraRefs)
}

// validateProjectPrefixes validates that JIRA keys use allowed project prefixes.
func (r JiraReferenceRule) validateProjectPrefixes(jiraRefs []string) []domain.ValidationError {
	if len(r.prefixes) == 0 {
		return nil // No project validation configured
	}

	var errors []domain.ValidationError

	for _, ref := range jiraRefs {
		project := r.extractProjectFromReference(ref)
		if !r.isValidProject(project) {
			errors = append(errors,
				domain.New(r.Name(), domain.ErrInvalidProject,
					fmt.Sprintf("JIRA project '%s' is not a valid project", project)).
					WithContextMap(map[string]string{
						"actual":   project,
						"expected": strings.Join(r.prefixes, ", "),
					}).
					WithHelp("Use one of these projects: "+strings.Join(r.prefixes, ", ")))
		}
	}

	return errors
}

// validateRefsLineFormat validates the format of a single "Refs:" line.
func (r JiraReferenceRule) validateRefsLineFormat(refsLine string, _ int) []domain.ValidationError {
	// Expected format: "Refs: PROJ-123" or "Refs: PROJ-123, TEAM-456"
	// Allow any case for initial parsing, individual key validation will check case
	// Allow trailing spaces for flexibility
	refsPattern := `^Refs:\s+([a-zA-Z]+-\d+(?:\s*,\s*[a-zA-Z]+-\d+)*)\s*$`
	matched, _ := regexp.MatchString(refsPattern, refsLine)

	if !matched {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrInvalidRefsFormat, "Invalid Refs format, should be 'Refs: PROJ-123'").
				WithContextMap(map[string]string{
					"actual":   refsLine,
					"expected": "Refs: " + r.getExpectedFormat(),
				}).
				WithHelp("Use format: Refs: " + r.getExpectedFormat() + " or Refs: " + r.getExpectedFormat() + ", OTHER-456"),
		}
	}

	// Extract and validate individual JIRA keys
	jiraKeys := r.extractJiraFromRefsLine(refsLine)

	var errors []domain.ValidationError

	var validKeys []string

	// First validate JIRA key format and collect valid keys
	for _, key := range jiraKeys {
		if !r.isValidJiraFormat(key) {
			errors = append(errors,
				domain.New(r.Name(), domain.ErrInvalidKeyFormat, "Invalid JIRA issue key format (should be PROJECT-123)").
					WithContextMap(map[string]string{
						"actual":   key,
						"expected": "PROJECT-123",
					}).
					WithHelp("Use format: PROJECT-123 (uppercase project, dash, number)"))
		} else {
			validKeys = append(validKeys, key)
		}
	}

	// Only validate project prefixes for keys that passed format validation
	if len(r.prefixes) > 0 {
		for _, key := range validKeys {
			project := r.extractProjectFromReference(key)
			if !r.isValidProject(project) {
				errors = append(errors,
					domain.New(r.Name(), domain.ErrInvalidProject,
						fmt.Sprintf("JIRA project '%s' is not a valid project", project)).
						WithContextMap(map[string]string{
							"actual":   project,
							"expected": strings.Join(r.prefixes, ", "),
						}).
						WithHelp("Use one of these projects: "+strings.Join(r.prefixes, ", ")))
			}
		}
	}

	return errors
}

// validateRefsOrdering validates that Refs lines appear before Signed-off-by lines.
func (r JiraReferenceRule) validateRefsOrdering(body string, refsLineNumbers []int) []domain.ValidationError {
	signoffLineNumbers := r.findSignoffLines(body)

	if len(signoffLineNumbers) == 0 {
		return nil // No Signed-off-by lines to check against
	}

	var errors []domain.ValidationError

	for _, refsLineNum := range refsLineNumbers {
		for _, signoffLineNum := range signoffLineNumbers {
			if refsLineNum > signoffLineNum {
				errors = append(errors,
					domain.New(r.Name(), domain.ErrRefsAfterSignoff, "Refs: line must appear before Signed-off-by lines").
						WithContextMap(map[string]string{
							"actual":   "Refs after Signed-off-by",
							"expected": "Refs before Signed-off-by",
						}).
						WithHelp("Move Refs: lines before Signed-off-by lines in the commit body"))

				return errors // Only report first violation
			}
		}
	}

	return errors
}

// Helper methods for parsing and validation

// isConventionalCommit checks if subject follows conventional commit format.
func (r JiraReferenceRule) isConventionalCommit(subject string) bool {
	// Use strict lowercase pattern for proper conventional commit detection
	pattern := `^[a-z]+(?:\([^)]*\))?!?:`
	matched, _ := regexp.MatchString(pattern, subject)

	return matched
}

// isJiraAtEnd checks if JIRA references are at the end of conventional commit subject.
func (r JiraReferenceRule) isJiraAtEnd(subject string, jiraRefs []string) bool {
	if len(jiraRefs) == 0 {
		return false
	}

	// Find the position of the last JIRA reference
	lastJiraPos := -1

	for _, ref := range jiraRefs {
		pos := strings.LastIndex(subject, ref)
		if pos > lastJiraPos {
			lastJiraPos = pos + len(ref)
		}
	}

	// Check if there's only whitespace after the last JIRA reference
	textAfterJira := strings.TrimSpace(subject[lastJiraPos:])

	return textAfterJira == ""
}

// isJiraInValidConventionalPosition checks if JIRA references are in valid positions for conventional commits.
func (r JiraReferenceRule) isJiraInValidConventionalPosition(subject string, jiraRefs []string) bool {
	if len(jiraRefs) == 0 {
		return false
	}

	// Check if JIRA is at the end
	if r.isJiraAtEnd(subject, jiraRefs) {
		return true
	}

	// Check if JIRA is in the scope part of a conventional commit
	scopePattern := `^[a-z]+\([^)]*\)!?:`
	if matched, _ := regexp.MatchString(scopePattern, subject); matched {
		// Extract the scope part
		re := regexp.MustCompile(`^[a-z]+\(([^)]*)\)!?:`)

		matches := re.FindStringSubmatch(subject)
		if len(matches) > 1 {
			scope := matches[1]
			// Check if any JIRA reference is in the scope
			for _, ref := range jiraRefs {
				if strings.Contains(scope, ref) {
					return true
				}
			}
		}
	}

	return false
}

// parseRefsLines extracts "Refs:" lines from commit body and returns them with line numbers.
func (r JiraReferenceRule) parseRefsLines(body string) ([]string, []int) {
	lines := strings.Split(body, "\n")

	var refsLines []string

	var lineNumbers []int

	for lineIndex, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Match lines that look like Refs line attempts
		if strings.HasPrefix(trimmed, "Refs:") ||
			regexp.MustCompile(`^Refs\s+[A-Z]+-\d+`).MatchString(trimmed) {
			refsLines = append(refsLines, trimmed)
			lineNumbers = append(lineNumbers, lineIndex+1) // 1-based line numbering
		}
	}

	return refsLines, lineNumbers
}

// extractJiraFromRefsLine extracts JIRA keys from a "Refs:" line.
func (r JiraReferenceRule) extractJiraFromRefsLine(refsLine string) []string {
	// Remove "Refs:" prefix and extract keys
	content := strings.TrimPrefix(refsLine, "Refs:")
	content = strings.TrimSpace(content)

	// Split by comma and clean up
	parts := strings.Split(content, ",")

	var keys []string

	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key != "" {
			keys = append(keys, key)
		}
	}

	return keys
}

// findSignoffLines finds line numbers of "Signed-off-by:" lines in body.
func (r JiraReferenceRule) findSignoffLines(body string) []int {
	lines := strings.Split(body, "\n")

	var lineNumbers []int

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "Signed-off-by:") {
			lineNumbers = append(lineNumbers, i+1) // 1-based line numbering
		}
	}

	return lineNumbers
}

// isValidJiraFormat validates JIRA key format (PROJECT-123).
func (r JiraReferenceRule) isValidJiraFormat(key string) bool {
	pattern := `^[A-Z]+-\d+$`
	matched, _ := regexp.MatchString(pattern, key)

	return matched
}

// getExpectedFormat returns example JIRA format based on configuration.
func (r JiraReferenceRule) getExpectedFormat() string {
	if len(r.prefixes) > 0 {
		return r.prefixes[0] + "-123"
	}

	return "PROJ-123"
}
