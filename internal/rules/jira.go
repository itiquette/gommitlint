// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"regexp"

	"github.com/janderssonse/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// JiraCheck enforces that a Jira issue is mentioned in the header.
type JiraCheck struct {
	errors []error
}

// Status returns the name of the check.
func (j *JiraCheck) Status() string {
	return "Jira issues"
}

// Message returns to check message.
func (j *JiraCheck) Message() string {
	if len(j.errors) != 0 {
		return j.errors[0].Error()
	}

	return "Jira issues are valid"
}

// Errors returns any violations of the check.
func (j *JiraCheck) Errors() []error {
	return j.errors
}

// ValidateJiraCheck validates if a Jira issue is mentioned in the header.
func ValidateJiraCheck(message string, jirakeys []string) interfaces.Check { //nolint:ireturn
	check := &JiraCheck{}

	reg := regexp.MustCompile(`.* \[?([A-Z]*)-[1-9]{1}\d*\]?.*`)

	if reg.MatchString(message) {
		submatch := reg.FindStringSubmatch(message)
		jiraProject := submatch[1]

		if !find(jirakeys, jiraProject) {
			check.errors = append(check.errors, errors.Errorf("Jira project %s is not a valid jira project", jiraProject))
		}
	} else {
		check.errors = append(check.errors, errors.Errorf("No Jira issue tag found in %q", message))
	}

	return check
}

func find(slice []string, value string) bool {
	for _, elem := range slice {
		if elem == value {
			return true
		}
	}

	return false
}
