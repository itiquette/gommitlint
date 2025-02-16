// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// JiraCheck enforces Jira issue references in commit messages.
type JiraCheck struct {
	errors []error
}

// Status returns the name of the check.
func (j *JiraCheck) Status() string {
	return "Jira issues"
}

// Message returns the check message.
func (j *JiraCheck) Message() string {
	if len(j.errors) > 0 {
		return j.errors[0].Error()
	}

	return "Jira issues are valid"
}

// Errors returns any violations of the check.
func (j *JiraCheck) Errors() []error {
	return j.errors
}

// ValidateJiraCheck validates Jira issue references in commit messages.
func ValidateJiraCheck(message string, validJiraProjects []string, isConventionalCommit bool) *JiraCheck {
	check := &JiraCheck{}

	// Regex for matching Jira issue key
	jiraKeyRegex := regexp.MustCompile(`([A-Z]+-\d+)`)

	// For conventional commits, look for Jira key at the end
	if isConventionalCommit {
		// Split the first line to handle conventional commit format
		parts := strings.Split(message, "\n")
		firstLine := parts[0]

		// Find all Jira issue keys
		matches := jiraKeyRegex.FindAllString(firstLine, -1)
		if len(matches) == 0 {
			check.errors = append(check.errors,
				errors.Errorf("no Jira issue key found at the end of the commit message"))

			return check
		}

		// Get the last match
		lastMatch := matches[len(matches)-1]

		// Check if the last match is truly at the end
		variants := []string{
			lastMatch,
			fmt.Sprintf("[%s]", lastMatch),
		}

		found := false

		for _, variant := range variants {
			if strings.HasSuffix(firstLine, variant) {
				found = true

				break
			}
		}

		if !found {
			// Specifically check for Jira keys appearing mid-message
			hasJiraKey := jiraKeyRegex.MatchString(firstLine)
			if hasJiraKey {
				check.errors = append(check.errors,
					errors.Errorf("Jira issue key must be at the end of the commit message"))

				return check
			}
		}

		// Check for Jira project validity if a key was found
		projectKey := strings.Split(lastMatch, "-")[0]
		if !containsString(validJiraProjects, projectKey) {
			check.errors = append(check.errors,
				errors.Errorf("Jira project %s is not a valid project", projectKey))
		}
	} else {
		// For non-conventional commits, check for Jira key anywhere
		matches := jiraKeyRegex.FindAllString(message, -1)
		if len(matches) == 0 {
			check.errors = append(check.errors,
				errors.Errorf("no Jira issue key found in the commit message"))

			return check
		}

		// Check each found project key
		for _, match := range matches {
			projectKey := strings.Split(match, "-")[0]
			if !containsString(validJiraProjects, projectKey) {
				check.errors = append(check.errors,
					errors.Errorf("Jira project %s is not a valid project", projectKey))
			}
		}
	}

	return check
}

// containsString checks if a slice contains a specific string.
func containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}

	return false
}
