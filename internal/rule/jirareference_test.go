// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateJiraReference(t *testing.T) {
	// Define valid Jira projects for testing
	validProjects := []string{"PROJ", "TEAM", "CORE"}

	// Create Jira rule configuration
	jiraConfig := &configuration.JiraRule{
		Keys:    validProjects,
		BodyRef: false, // Default to subject checking
	}

	// Test cases for subject validation (default behavior)
	subjectTestCases := []struct {
		name                 string
		subject              string
		body                 string
		isConventionalCommit bool
		expectedErrors       bool
		errorCode            string
		errorContains        string
	}{
		// Conventional Commit Positive Cases
		{
			name:                 "Valid conventional commit with jira key at end",
			subject:              "feat(auth): add user authentication [PROJ-123]",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		{
			name:                 "Valid conventional commit with multiple words jira key",
			subject:              "fix(profile): resolve user profile update issue [TEAM-456]",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		{
			name:                 "Valid conventional commit with multiline message",
			subject:              "refactor(api): simplify authentication middleware [CORE-789]\n\nAdditional context about the change",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		{
			name:                 "Valid conventional commit with parentheses format",
			subject:              "docs(readme): update installation instructions (PROJ-123)",
			isConventionalCommit: true,
			expectedErrors:       false,
		},
		{
			name:                 "Valid conventional commit with plain key at end",
			subject:              "chore(deps): update dependencies TEAM-456",
			isConventionalCommit: true,
			expectedErrors:       false,
		},

		// Conventional Commit Negative Cases
		{
			name:                 "Conventional commit missing jira key",
			subject:              "feat(auth): add user authentication",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorCode:            "missing_jira_key_subject",
			errorContains:        "no Jira issue key found",
		},
		{
			name:                 "Conventional commit jira key not at end",
			subject:              "feat(auth): [PROJ-123] add user authentication",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorCode:            "key_not_at_end",
			errorContains:        "must be at the end of the first line",
		},
		{
			name:                 "Conventional commit invalid jira project",
			subject:              "feat(auth): add user authentication [UNKNOWN-123]",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorCode:            "invalid_project",
			errorContains:        "not a valid project",
		},
		{
			name:                 "Empty subject",
			subject:              "",
			isConventionalCommit: true,
			expectedErrors:       true,
			errorCode:            "empty_subject",
			errorContains:        "commit subject is empty",
		},
		{
			name:                 "Invalid Jira key format",
			subject:              "feat(auth): add user authentication [proj-123]", // lowercase project
			isConventionalCommit: true,
			expectedErrors:       true,
			errorCode:            "missing_jira_key_subject",
			errorContains:        "no Jira issue key found",
		},

		// Non-Conventional Commit Positive Cases
		{
			name:                 "Valid non-conventional commit anywhere",
			subject:              "PROJ-123 Implement user authentication",
			isConventionalCommit: false,
			expectedErrors:       false,
		},
		{
			name:                 "Valid non-conventional commit multiple issues",
			subject:              "Implement PROJ-123 and TEAM-456 features",
			isConventionalCommit: false,
			expectedErrors:       false,
		},
		{
			name:                 "Valid non-conventional commit with prefix format",
			subject:              "CORE-789: Implement new authentication flow",
			isConventionalCommit: false,
			expectedErrors:       false,
		},
		{
			name:                 "Valid non-conventional commit with bracket prefix",
			subject:              "[PROJ-123] Fix bug in login form",
			isConventionalCommit: false,
			expectedErrors:       false,
		},
		{
			name:                 "Valid non-conventional commit with dash format",
			subject:              "TEAM-456 - Update user interface components",
			isConventionalCommit: false,
			expectedErrors:       false,
		},

		// Non-Conventional Commit Negative Cases
		{
			name:                 "Non-conventional commit missing jira key",
			subject:              "Implement user authentication",
			isConventionalCommit: false,
			expectedErrors:       true,
			errorCode:            "missing_jira_key_subject",
			errorContains:        "no Jira issue key found",
		},
		{
			name:                 "Non-conventional commit invalid jira project",
			subject:              "Implement UNKNOWN-123 feature",
			isConventionalCommit: false,
			expectedErrors:       true,
			errorCode:            "invalid_project",
			errorContains:        "not a valid project",
		},
	}

	// Run tests for subject validation
	for _, tabletest := range subjectTestCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Execute validation with subject checking (default behavior)
			result := rule.ValidateJiraReference(tabletest.subject, tabletest.body, jiraConfig, tabletest.isConventionalCommit)

			// Check for expected errors
			if tabletest.expectedErrors {
				require.NotEmpty(t, result.Errors(), "Expected errors but found none")

				// Check error code if specified
				if tabletest.errorCode != "" {
					assert.Equal(t, tabletest.errorCode, result.Errors()[0].Code,
						"Error code should match expected")
				}

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

				// Verify rule name is set in ValidationError
				assert.Equal(t, "JiraReference", result.Errors()[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify Help() method provides guidance
				helpText := result.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")

				if strings.Contains(result.Result(), "empty") {
					assert.Contains(t, helpText, "Provide a non-empty", "Help should mention providing a non-empty message")
				} else if strings.Contains(result.Errors()[0].Error(), "no Jira issue") {
					assert.Contains(t, helpText, "Include a valid Jira", "Help should mention including a valid Jira key")
				} else if strings.Contains(result.Errors()[0].Error(), "must be at the end") {
					assert.Contains(t, helpText, "place the Jira issue key at the end", "Help should mention key placement")
				} else if strings.Contains(result.Errors()[0].Error(), "not a valid project") {
					assert.Contains(t, helpText, "not recognized as a valid project", "Help should explain invalid project")
				}
			} else {
				require.Empty(t, result.Errors(), "Unexpected errors found: %v", result.Errors())
				require.Equal(t, "Valid Jira reference", result.Result(), "Expected default valid message")

				// Test Help on valid case
				assert.Equal(t, "No errors to fix", result.Help(), "Help for valid message should indicate nothing to fix")
			}

			// Verify Name() method
			require.Equal(t, "JiraReference", result.Name(), "Name should be 'JiraReference'")
		})
	}

	// Test cases for body validation
	bodyRefJiraConfig := &configuration.JiraRule{
		Keys:    validProjects,
		BodyRef: true, // Check in body instead of subject
	}

	bodyTestCases := []struct {
		name                 string
		subject              string
		body                 string
		isConventionalCommit bool
		expectedErrors       bool
		errorCode            string
		errorContains        string
	}{
		// Body validation positive cases
		{
			name:           "Valid body reference single key",
			subject:        "feat(auth): add user authentication", // No Jira key needed in subject
			body:           "This adds JWT authentication.\n\nRefs: PROJ-123",
			expectedErrors: false,
		},
		{
			name:           "Valid body reference multiple keys",
			subject:        "Fix multiple bugs",
			body:           "This fixes several issues.\n\nRefs: PROJ-123, TEAM-456",
			expectedErrors: false,
		},
		{
			name:           "Valid body reference with signed-off-by after",
			subject:        "Update documentation",
			body:           "Updated API docs.\n\nRefs: CORE-789\n\nSigned-off-by: Developer <dev@example.com>",
			expectedErrors: false,
		},

		// Body validation negative cases
		{
			name:           "Missing Refs in body",
			subject:        "Fix bug",
			body:           "This fixes a critical bug.",
			expectedErrors: true,
			errorCode:      "missing_jira_key_body",
			errorContains:  "no Jira issue key found in the commit body",
		},
		{
			name:           "Invalid Refs format",
			subject:        "Update API",
			body:           "API changes.\n\nRefs: invalid-key",
			expectedErrors: true,
			errorCode:      "invalid_refs_format",
			errorContains:  "invalid Refs format",
		},
		{
			name:           "Refs line after signed-off-by",
			subject:        "Refactor code",
			body:           "Major refactoring.\n\nSigned-off-by: Developer <dev@example.com>\nRefs: PROJ-123",
			expectedErrors: true,
			errorCode:      "refs_after_signoff",
			errorContains:  "Refs: line must appear before any Signed-off-by lines",
		},
		{
			name:           "Invalid Refs format with colon",
			subject:        "Update API",
			body:           "API changes.\n\nRefs: PROJ-123: Added feature",
			expectedErrors: true,
			errorCode:      "invalid_refs_format",
			errorContains:  "invalid Refs format",
		},
		{
			name:           "Empty body",
			subject:        "Fix bug",
			body:           "",
			expectedErrors: true,
			errorCode:      "missing_jira_key_body",
			errorContains:  "no Jira issue key found in the commit body",
		},
	}

	// Run tests for body validation
	for _, tabletest := range bodyTestCases {
		t.Run("BodyRef: "+tabletest.name, func(t *testing.T) {
			// Execute validation with body checking (BodyRef=true)
			result := rule.ValidateJiraReference(tabletest.subject, tabletest.body, bodyRefJiraConfig, tabletest.isConventionalCommit)

			// Check for expected errors
			if tabletest.expectedErrors {
				require.NotEmpty(t, result.Errors(), "Expected errors but found none")

				// Check error code if specified
				if tabletest.errorCode != "" {
					assert.Equal(t, tabletest.errorCode, result.Errors()[0].Code,
						"Error code should match expected")
				}

				// Check error message contains expected substring
				if tabletest.errorContains != "" {
					found := false

					for _, err := range result.Errors() {
						if strings.Contains(err.Error(), tabletest.errorContains) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error containing %q but got: %v", tabletest.errorContains, result.Errors())
				}

				// Verify rule name is set in ValidationError
				assert.Equal(t, "JiraReference", result.Errors()[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify Help() method provides guidance for body errors
				helpText := result.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")

				// Verify context information is present for specific error types
				if tabletest.errorCode == "refs_after_signoff" {
					assert.Contains(t, result.Errors()[0].Context, "refs_line",
						"Context should contain refs_line information")
					assert.Contains(t, result.Errors()[0].Context, "signoff_line",
						"Context should contain signoff_line information")
				}
			} else {
				require.Empty(t, result.Errors(), "Unexpected errors found: %v", result.Errors())
				require.Equal(t, "Valid Jira reference", result.Result(), "Expected default valid message")
			}
		})
	}

	// Test with nil configuration (should default to subject checking)
	t.Run("Nil configuration", func(t *testing.T) {
		result := rule.ValidateJiraReference("feat: add feature CORE-123", "", nil, true)
		assert.Empty(t, result.Errors(), "Should validate successfully with nil config")
	})

	// Test with empty project list (should validate format only)
	t.Run("Empty project list", func(t *testing.T) {
		emptyProjectConfig := &configuration.JiraRule{
			Keys:    []string{},
			BodyRef: false,
		}

		// Valid format should pass
		result := rule.ValidateJiraReference("feat: add feature ABC-123", "", emptyProjectConfig, true)
		assert.Empty(t, result.Errors(), "Should accept any project key with valid format")

		// Invalid format should fail
		result = rule.ValidateJiraReference("feat: add feature abc-123", "", emptyProjectConfig, true)
		assert.NotEmpty(t, result.Errors(), "Should reject invalid format")
		assert.Equal(t, "missing_jira_key_subject", result.Errors()[0].Code,
			"Error code should be missing_jira_key_subject")
	})

	// Test context data in errors
	t.Run("Context data in errors", func(t *testing.T) {
		// Test invalid project error context
		result := rule.ValidateJiraReference("feat: add feature INVALID-123", "", jiraConfig, true)
		require.NotEmpty(t, result.Errors(), "Should have errors")
		assert.Equal(t, "invalid_project", result.Errors()[0].Code, "Error code should be invalid_project")
		assert.Equal(t, "INVALID", result.Errors()[0].Context["project"],
			"Context should contain the invalid project")
		assert.Contains(t, result.Errors()[0].Context["valid_projects"], validProjects[0],
			"Context should list valid projects")

		// Test key not at end error context
		result = rule.ValidateJiraReference("feat: PROJ-123 add feature", "", jiraConfig, true)
		require.NotEmpty(t, result.Errors(), "Should have errors")
		assert.Equal(t, "key_not_at_end", result.Errors()[0].Code, "Error code should be key_not_at_end")
		assert.Equal(t, "PROJ-123", result.Errors()[0].Context["key"],
			"Context should contain the key that's not at the end")
	})
}
