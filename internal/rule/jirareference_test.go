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

func TestValidateJiraCheck(t *testing.T) {
	// Define valid Jira projects for testing
	validProjects := []string{"PROJ", "TEAM", "CORE"}

	// Test cases covering various scenarios
	testCases := []struct {
		name                 string
		message              string
		isConventionalCommit bool
		expectedErrors       bool
		errorContains        string
	}{
		// Conventional Commit Positive Cases
		{
			name:                 "Valid Conventional Commit with Jira Key at End",
			message:              "feat(auth): add user authentication [PROJ-123]",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		{
			name:                 "Valid Conventional Commit with Multiple Words Jira Key",
			message:              "fix(profile): resolve user profile update issue [TEAM-456]",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		{
			name:                 "Valid Conventional Commit with Multiline Message",
			message:              "refactor(api): simplify authentication middleware [CORE-789]\n\nAdditional context about the change",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		// Conventional Commit Negative Cases
		{
			name:                 "Conventional Commit Missing Jira Key",
			message:              "feat(auth): add user authentication",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorContains:        "no Jira issue key found",
		},
		{
			name:                 "Conventional Commit Jira Key Not at End",
			message:              "feat(auth): [PROJ-123] add user authentication",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorContains:        "Jira issue key must be at the end",
		},
		{
			name:                 "Conventional Commit Invalid Jira Project",
			message:              "feat(auth): add user authentication [UNKNOWN-123]",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorContains:        "not a valid project",
		},
		// Non-Conventional Commit Positive Cases
		{
			name:                 "Valid Non-Conventional Commit Anywhere",
			message:              "PROJ-123 Implement user authentication",
			isConventionalCommit: false,
			expectedErrors:       false,
		},
		{
			name:                 "Valid Non-Conventional Commit Multiple Issues",
			message:              "Implement PROJ-123 and TEAM-456 features",
			isConventionalCommit: false,
			expectedErrors:       false,
		},
		// Non-Conventional Commit Negative Cases
		{
			name:                 "Non-Conventional Commit Missing Jira Key",
			message:              "Implement user authentication",
			isConventionalCommit: false,
			expectedErrors:       true,
			errorContains:        "no Jira issue key found",
		},
		{
			name:                 "Non-Conventional Commit Invalid Jira Project",
			message:              "Implement UNKNOWN-123 feature",
			isConventionalCommit: false,
			expectedErrors:       true,
			errorContains:        "not a valid project",
		},
	}

	// Run test cases
	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Execute Jira result
			result := rule.ValidateJira(tabletest.message, validProjects, tabletest.isConventionalCommit)

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
			require.NotEmpty(t, result.Result(), "Message should not be empty")
		})
	}
}

// Benchmark the Jira check validation.
func BenchmarkValidateJiraCheck(b *testing.B) {
	validProjects := []string{"PROJ", "TEAM", "CORE"}

	benchCases := []struct {
		name                 string
		message              string
		isConventionalCommit bool
	}{
		{
			name:                 "Conventional Commit",
			message:              "feat(auth): add user authentication [PROJ-123]",
			isConventionalCommit: true,
		},
		{
			name:                 "Non-Conventional Commit",
			message:              "PROJ-123 Implement user authentication",
			isConventionalCommit: false,
		},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				rule.ValidateJira(bc.message, validProjects, bc.isConventionalCommit)
			}
		})
	}
}
