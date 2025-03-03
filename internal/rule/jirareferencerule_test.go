// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateJiraRule(t *testing.T) {
	// Define valid Jira projects for testing
	validProjects := []string{"PROJ", "TEAM", "CORE"}

	// Test cases covering various scenarios
	testCases := []struct {
		name                 string
		subject              string
		isConventionalCommit bool
		expectedErrors       bool
		errorContains        string
	}{
		// Conventional Commit Positive Cases
		{
			name:                 "Valid Conventional Commit with Jira Key at End",
			subject:              "feat(auth): add user authentication [PROJ-123]",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		{
			name:                 "Valid Conventional Commit with Multiple Words Jira Key",
			subject:              "fix(profile): resolve user profile update issue [TEAM-456]",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		{
			name:                 "Valid Conventional Commit with Multiline Message",
			subject:              "refactor(api): simplify authentication middleware [CORE-789]\n\nAdditional context about the change",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		// Conventional Commit Negative Cases
		{
			name:                 "Conventional Commit Missing Jira Key",
			subject:              "feat(auth): add user authentication",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorContains:        "no Jira issue key found",
		},
		{
			name:                 "Conventional Commit Jira Key Not at End",
			subject:              "feat(auth): [PROJ-123] add user authentication",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorContains:        "Jira issue key must be at the end",
		},
		{
			name:                 "Conventional Commit Invalid Jira Project",
			subject:              "feat(auth): add user authentication [UNKNOWN-123]",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorContains:        "not a valid project",
		},
		// Non-Conventional Commit Positive Cases
		{
			name:                 "Valid Non-Conventional Commit Anywhere",
			subject:              "PROJ-123 Implement user authentication",
			isConventionalCommit: false,
			expectedErrors:       false,
		},
		{
			name:                 "Valid Non-Conventional Commit Multiple Issues",
			subject:              "Implement PROJ-123 and TEAM-456 features",
			isConventionalCommit: false,
			expectedErrors:       false,
		},
		// Non-Conventional Commit Negative Cases
		{
			name:                 "Non-Conventional Commit Missing Jira Key",
			subject:              "Implement user authentication",
			isConventionalCommit: false,
			expectedErrors:       true,
			errorContains:        "no Jira issue key found",
		},
		{
			name:                 "Non-Conventional Commit Invalid Jira Project",
			subject:              "Implement UNKNOWN-123 feature",
			isConventionalCommit: false,
			expectedErrors:       true,
			errorContains:        "not a valid project",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Execute Jira result
			result := rule.ValidateJira(tabletest.subject, validProjects, tabletest.isConventionalCommit)

			// Check for expected errors
			if tabletest.expectedErrors {
				require.NotEmpty(t, result.Errors(), "Expected errors but found none")

				// Check error message contains expected substring
				if tabletest.errorContains != "" {
					found := false

					for _, err := range result.Errors() {
						if strings.Contains(err.Error(), tabletest.errorContains) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error containing %q", tabletest.errorContains)
				}
			} else {
				require.Empty(t, result.Errors(), "Unexpected errors found")
			}

			// Verify Status and Message methods work
			require.NotEmpty(t, result.Name(), "Status should not be empty")
			require.NotEmpty(t, result.Result(), "Result should not be empty")
		})
	}
}
