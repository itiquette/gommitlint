// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"regexp"
	"strings"
)

// JiraReferenceRule enforces Jira issue references in commit messages.
type JiraReferenceRule struct {
	errors []error
}

// Name returns the name of the rule.
func (j *JiraReferenceRule) Name() string {
	return "JiraReferenceRule"
}

// Result returns the rule message.
func (j *JiraReferenceRule) Result() string {
	if len(j.errors) > 0 {
		return j.errors[0].Error()
	}

	return "Jira issues are valid"
}

// Errors returns any violations of the rule.
func (j *JiraReferenceRule) Errors() []error {
	return j.errors
}

// addErrorf adds an error to the rule's errors slice.
func (j *JiraReferenceRule) addErrorf(format string, args ...interface{}) {
	j.errors = append(j.errors, fmt.Errorf(format, args...))
}

// ValidateJira validates Jira issue references in commit messages.
// Returns a JiraReferenceRule with any validation errors.
func ValidateJira(message string, validJiraProjects []string, isConventionalCommit bool) *JiraReferenceRule {
	rule := &JiraReferenceRule{}

	// Compile regex only once
	jiraKeyRegex := regexp.MustCompile(`([A-Z]+-\d+)`)

	// Normalize and trim the message
	message = strings.TrimSpace(message)
	if message == "" {
		rule.addErrorf("commit message is empty")

		return rule
	}

	lines := strings.Split(message, "\n")
	firstLine := lines[0]

	if isConventionalCommit {
		// For conventional commits, check for Jira key at the end
		matches := jiraKeyRegex.FindAllString(firstLine, -1)
		if len(matches) == 0 {
			rule.addErrorf("no Jira issue key found in the commit message")

			return rule
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
			if jiraKeyRegex.MatchString(firstLine) {
				rule.addErrorf("in conventional commit format, Jira issue key must be at the end of the first line")
			} else {
				rule.addErrorf("no Jira issue key found in the commit message")
			}

			return rule
		}

		// Validate Jira project
		projectKey := strings.Split(lastMatch, "-")[0]
		if !containsString(validJiraProjects, projectKey) {
			rule.addErrorf("Jira project %s is not a valid project", projectKey)
		}
	} else {
		// For non-conventional commits, check for common Jira reference patterns
		matches := jiraKeyRegex.FindAllString(message, -1)
		if len(matches) == 0 {
			rule.addErrorf("no Jira issue key found in the commit message")

			return rule
		}

		// Common patterns exist but we're not enforcing them
		// Just validating project keys

		// Validate all found project keys
		for _, match := range matches {
			projectKey := strings.Split(match, "-")[0]
			if !containsString(validJiraProjects, projectKey) {
				rule.addErrorf("Jira project %s is not a valid project", projectKey)

				break
			}
		}
	}

	return rule
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

//TO-Do: add body check,
