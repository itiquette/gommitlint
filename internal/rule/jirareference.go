// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/configuration"
)

// Common regex patterns compiled once at package level.
var (
	jiraKeyRegex  = regexp.MustCompile(`([A-Z]+-\d+)`)
	refsLineRegex = regexp.MustCompile(`^Refs:\s*([A-Z]+-\d+(?:\s*,\s*[A-Z]+-\d+)*)$`)
)

// JiraReference enforces proper Jira issue references in commit messages.
//
// This rule ensures that commit messages include valid Jira issue keys (e.g., PROJECT-123)
// according to the project's conventions for placement and format. It helps maintain
// traceability between code changes and issue tracking systems, making it easier to
// understand the purpose and context of each commit.
//
// The rule supports two validation modes:
//  1. Subject-based validation - Checks for Jira keys in the commit subject line
//  2. Body-based validation - Checks for Jira keys in dedicated "Refs:" lines in the commit body
//
// Examples:
//
//   - For conventional commits with subject validation:
//     "feat(auth): add login feature PROJ-123" would pass
//     "fix: resolve timeout issue [PROJ-123]" would pass
//     "feat(PROJ-123): add login feature" would fail (key not at end)
//
//   - For non-conventional commits with subject validation:
//     "Add login feature PROJ-123" would pass
//     "PROJ-123: Fix memory leak" would pass (key can be anywhere)
//     "Add new feature" would fail (missing key)
//
//   - For body reference validation:
//     "feat: add feature
//
//     Implements the login functionality
//     as described in the spec.
//
//     Refs: PROJ-123" would pass
//
//     "feat: add feature
//
//     Implements login.
//
//     Signed-off-by: Dev <dev@example.com>
//     Refs: PROJ-123" would fail (Refs after Sign-off)
//
// If a list of valid Jira project keys is provided in the configuration, the
// rule also validates that all referenced projects exist in that list.
type JiraReference struct {
	errors []error
}

// Name returns the name of the rule.
func (j *JiraReference) Name() string {
	return "JiraReference"
}

// Result returns the rule message.
func (j *JiraReference) Result() string {
	if len(j.errors) > 0 {
		return j.errors[0].Error()
	}

	return "Jira issues are valid"
}

// Errors returns any violations of the rule.
func (j *JiraReference) Errors() []error {
	return j.errors
}

// Help returns a description of how to fix the rule violation.
func (j *JiraReference) Help() string {
	if len(j.errors) == 0 {
		return "No errors to fix"
	}

	errMsg := j.errors[0].Error()

	if strings.Contains(errMsg, "commit subject is empty") {
		return "Provide a non-empty commit message with a Jira issue reference"
	}

	if strings.Contains(errMsg, "no Jira issue key found") {
		if strings.Contains(errMsg, "in the body") {
			return `Include a valid Jira issue key (e.g., PROJECT-123) in your commit body with the "Refs:" prefix.

Examples:
- Refs: PROJECT-123
- Refs: PROJECT-123, PROJECT-456
- Refs: PROJECT-123, PROJECT-456, PROJECT-789

The Refs: line should appear at the end of the commit body, before any Signed-off-by lines.`
		}

		return `Include a valid Jira issue key (e.g., PROJECT-123) in your commit subject.

For conventional commits, place the Jira key at the end of the first line:
- feat(auth): add login feature PROJ-123
- fix: resolve timeout issue [PROJ-123]
- docs(readme): update installation steps (PROJ-123)

For other commit formats, include the Jira key anywhere in the subject.
`
	}

	if strings.Contains(errMsg, "must be at the end") {
		return `In conventional commit format, place the Jira issue key at the end of the first line.

Examples:
- feat(auth): add login feature PROJ-123
- fix: resolve timeout issue [PROJ-123]
- docs(readme): update installation steps (PROJ-123)

Avoid putting the Jira key in the middle of the line:
- INCORRECT: feat(PROJ-123): add login feature
- INCORRECT: fix: PROJ-123 resolve timeout issue
`
	}

	if strings.Contains(errMsg, "not a valid project") {
		projectKeys := getValidProjectsFromConfig()
		if len(projectKeys) > 0 {
			return fmt.Sprintf(`The Jira project reference is not recognized as a valid project.

Valid projects: %s

Please use one of these project keys in your Jira reference.
`, strings.Join(projectKeys, ", "))
		}

		return `The Jira project reference is not valid.
Jira project keys should be uppercase letters followed by a hyphen and numbers (e.g., PROJECT-123).`
	}

	if strings.Contains(errMsg, "invalid Refs format") {
		return `The "Refs:" line in your commit body has an invalid format.

The correct format is:
Refs: PROJECT-123
or for multiple references:
Refs: PROJECT-123, PROJECT-456

Make sure:
- "Refs:" is at the beginning of the line
- Project keys follow the format PROJECT-123
- Multiple references are separated by commas
- The Refs line appears before any Signed-off-by lines`
	}

	if strings.Contains(errMsg, "Refs: line must appear before") {
		return `The "Refs:" line must appear before any "Signed-off-by" lines in your commit message.

The correct order is:
1. Commit subject
2. Blank line
3. Commit body (if any)
4. Refs: line(s)
5. Signed-off-by line(s)`
	}

	if strings.Contains(errMsg, "invalid Jira issue key format") {
		return `The Jira issue key has an invalid format.

Jira keys must follow the format PROJECT-123, where:
- PROJECT is one or more uppercase letters
- Followed by a hyphen (-)
- Followed by one or more digits (123)

Examples of valid keys: PROJ-123, DEV-456, TEAM-7890`
	}

	// Default help
	return `Ensure your commit message contains a valid Jira issue reference.
The Jira issue key should follow the format PROJECT-123.`
}

// addErrorf adds an error to the rule's errors slice.
func (j *JiraReference) addErrorf(format string, args ...interface{}) {
	j.errors = append(j.errors, fmt.Errorf(format, args...))
}

// ValidateJiraReference checks if the commit message contains valid Jira issue references
// according to the configured validation rules.
//
// Parameters:
//   - subject: The commit subject line
//   - body: The commit message body
//   - jira: Configuration for Jira validation rules
//   - isConventionalCommit: Whether the commit follows conventional commit format
//
// The function checks for valid Jira issue references based on the configured mode:
//   - When BodyRef is enabled, it looks for "Refs: PROJ-123" lines in the commit body
//   - Otherwise, it validates Jira references in the commit subject
//
// For conventional commits in subject validation mode, the Jira key must appear at the end
// of the first line. For non-conventional commits, the key can appear anywhere in the subject.
//
// If a list of valid project keys is provided in the configuration, the function also
// validates that all referenced projects exist in that list.
//
// Returns:
//   - A JiraReference instance with validation results
func ValidateJiraReference(subject string, body string, jira *configuration.JiraRule, isConventionalCommit bool) *JiraReference {
	rule := &JiraReference{}

	// Determine validation strategy based on configuration
	checkBodyRefs := jira != nil && jira.BodyRef

	var validJiraProjects []string
	if jira != nil {
		validJiraProjects = jira.Keys
	}

	// Normalize and trim the subject
	subject = strings.TrimSpace(subject)
	if subject == "" {
		rule.addErrorf("commit subject is empty")

		return rule
	}

	// Validate based on the configured strategy
	if checkBodyRefs {
		validateBodyReferences(rule, body, validJiraProjects)
	} else {
		validateSubjectReferences(rule, subject, validJiraProjects, isConventionalCommit)
	}

	return rule
}

// validateSubjectReferences validates Jira references in the commit subject.
//
// Parameters:
//   - rule: The JiraReference rule being populated
//   - subject: The commit subject to validate
//   - validJiraProjects: List of valid Jira project keys, if any
//   - isConventionalCommit: Whether to validate as a conventional commit
//
// The function dispatches to the appropriate validation function based on whether
// the commit follows conventional commit format or not.
func validateSubjectReferences(rule *JiraReference, subject string, validJiraProjects []string, isConventionalCommit bool) {
	lines := strings.Split(subject, "\n")
	firstLine := lines[0]

	if isConventionalCommit {
		validateConventionalCommitSubject(rule, firstLine, validJiraProjects)
	} else {
		validateNonConventionalCommitSubject(rule, subject, validJiraProjects)
	}
}

// validateConventionalCommitSubject validates a conventional commit subject line.
//
// Parameters:
//   - rule: The JiraReference rule being populated
//   - firstLine: The first line of the commit message
//   - validJiraProjects: List of valid Jira project keys, if any
//
// The function checks that:
//  1. A Jira issue key exists in the subject
//  2. The key appears at the end of the subject line
//  3. The key references a valid project (if validJiraProjects is provided)
//
// For conventional commits, the Jira key must be at the end of the subject line,
// optionally enclosed in brackets or parentheses.
func validateConventionalCommitSubject(rule *JiraReference, firstLine string, validJiraProjects []string) {
	matches := jiraKeyRegex.FindAllString(firstLine, -1)
	if len(matches) == 0 {
		rule.addErrorf("no Jira issue key found in the commit subject")

		return
	}

	// Get the last match
	lastMatch := matches[len(matches)-1]

	// Check if the last match is at the end of the first line
	// Supporting common formats: PROJ-123, [PROJ-123], (PROJ-123)
	validSuffixes := []string{
		lastMatch,
		fmt.Sprintf("[%s]", lastMatch),
		fmt.Sprintf("(%s)", lastMatch),
	}

	found := false

	for _, suffix := range validSuffixes {
		if strings.HasSuffix(firstLine, suffix) {
			found = true

			break
		}
	}

	if !found {
		rule.addErrorf("in conventional commit format, Jira issue key must be at the end of the first line")

		return
	}

	// Validate Jira project if found
	validateJiraProject(rule, lastMatch, validJiraProjects)
}

// validateNonConventionalCommitSubject validates a non-conventional commit subject.
//
// Parameters:
//   - rule: The JiraReference rule being populated
//   - subject: The commit subject to validate
//   - validJiraProjects: List of valid Jira project keys, if any
//
// The function checks that:
//  1. At least one Jira issue key exists in the subject
//  2. All keys reference valid projects (if validJiraProjects is provided)
//
// For non-conventional commits, Jira keys can appear anywhere in the subject.
func validateNonConventionalCommitSubject(rule *JiraReference, subject string, validJiraProjects []string) {
	matches := jiraKeyRegex.FindAllString(subject, -1)
	if len(matches) == 0 {
		rule.addErrorf("no Jira issue key found in the commit subject")

		return
	}

	// Validate all found project keys
	for _, match := range matches {
		if !validateJiraProject(rule, match, validJiraProjects) {
			return // Stop on first invalid project
		}
	}
}

// A valid "Refs:" line follows the format: "Refs: PROJ-123" or "Refs: PROJ-123, PROJ-456".
func validateBodyReferences(rule *JiraReference, body string, validJiraProjects []string) {
	body = strings.TrimSpace(body)
	if body == "" {
		rule.addErrorf("no Jira issue key found in the commit body")

		return
	}

	// Look for "Refs:" lines
	bodyLines := strings.Split(body, "\n")

	foundRefs := false
	signOffFound := false
	signOffLineNum := -1
	refsLineNum := -1

	// First pass: find Refs: and Signed-off-by lines
	for ind, line := range bodyLines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Signed-off-by:") {
			signOffFound = true
			signOffLineNum = ind
		}

		if refsLineRegex.MatchString(line) {
			foundRefs = true
			refsLineNum = ind
		} else if strings.HasPrefix(line, "Refs:") {
			// Line starts with Refs: but doesn't match the expected format
			rule.addErrorf("invalid Refs format in commit body, should be 'Refs: PROJ-123' or 'Refs: PROJ-123, PROJ-456'")

			return
		}
	}

	// Validate that Refs: exists
	if !foundRefs {
		rule.addErrorf("no Jira issue key found in the commit body with Refs: prefix")

		return
	}

	// Validate that Refs: appears before any Signed-off-by lines
	if signOffFound && refsLineNum > signOffLineNum {
		rule.addErrorf("Refs: line must appear before any Signed-off-by lines")

		return
	}

	// Validate the Jira keys in the Refs: line
	for _, line := range bodyLines {
		line = strings.TrimSpace(line)
		if refsLineRegex.MatchString(line) {
			// Extract and validate all Jira keys
			matches := jiraKeyRegex.FindAllString(line, -1)
			for _, match := range matches {
				if !validateJiraProject(rule, match, validJiraProjects) {
					return // Stop on first invalid project
				}
			}

			break // Process only the first Refs: line
		}
	}
}

// validateJiraProject checks if a Jira issue key is valid.
//
// Parameters:
//   - rule: The JiraReference rule being populated
//   - jiraKey: The Jira issue key to validate
//   - validJiraProjects: List of valid Jira project keys, if any
//
// The function checks that:
//  1. The key follows the correct format (PROJECT-123)
//  2. The project part of the key exists in the validJiraProjects list, if provided
//
// Returns:
//   - true if the key is valid, false otherwise
func validateJiraProject(rule *JiraReference, jiraKey string, validJiraProjects []string) bool {
	// First, verify the key has the correct format
	if !jiraKeyRegex.MatchString(jiraKey) {
		rule.addErrorf("invalid Jira issue key format: %s (should be PROJECT-123)", jiraKey)

		return false
	}

	// If no project list is provided, just validate the format
	if len(validJiraProjects) == 0 {
		return true
	}

	// When project list is provided, validate against it
	projectKey := strings.Split(jiraKey, "-")[0]
	if !containsString(validJiraProjects, projectKey) {
		rule.addErrorf("Jira project %s is not a valid project", projectKey)

		return false
	}

	return true
}

// containsString checks if a string is present in a slice of strings.
//
// Parameters:
//   - slice: The slice of strings to search
//   - value: The string to look for
//
// Returns:
//   - true if the value is found in the slice, false otherwise
func containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}

	return false
}

// getValidProjectsFromConfig extracts valid project keys for use in the Help method.
//
// Returns:
//   - A slice of string containing valid Jira project keys
func getValidProjectsFromConfig() []string {
	// This is a dummy implementation that would be replaced with actual logic
	// in a real implementation that has access to the validJiraProjects
	return []string{"PROJ", "CORE", "TEAM", "etc."}
}
