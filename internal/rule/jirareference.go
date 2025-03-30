// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
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
// This rule ensures that commit messages include valid Jira issue keys (e.g., PROJECT-123)
// and that they follow the project's conventions for placement and format.
//
// For conventional commits, the rule checks that a Jira issue key appears at the end
// of the first line, in formats like "feat(scope): description PROJ-123",
// "fix: bug fix [PROJ-123]", or "docs: update readme (PROJ-123)".
//
// For non-conventional commits, the rule verifies that a Jira issue key appears
// anywhere in the message and that it refers to a valid project.
//
// If BodyRef is enabled, the rule will also look for Jira references in the commit body
// using the "Refs:" prefix, such as "Refs: PROJECT-123" or "Refs: PROJECT-123, PROJECT-456".
//
// All referenced Jira projects must be in the list of valid projects provided
// to the validator, if any projects are specified.
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

// validateBodyReferences validates Jira references in the commit body.
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

// Returns true if valid, false if invalid.
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
func containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}

	return false
}

// getValidProjectsFromConfig extracts valid project keys
// for use in the Help method.
func getValidProjectsFromConfig() []string {
	// This is a dummy implementation that would be replaced with actual logic
	// in a real implementation that has access to the validJiraProjects
	return []string{"PROJ", "CORE", "TEAM", "etc."}
}
