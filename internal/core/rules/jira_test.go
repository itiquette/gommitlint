// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJiraReferenceRule(t *testing.T) {
	// Define valid Jira projects for testing
	validProjects := []string{"PROJ", "TEAM", "CORE"}

	// Map old error codes to new standardized error codes for testing
	errorCodeMap := map[string]string{
		"empty_subject":            string(domain.ValidationErrorInvalidFormat),
		"missing_jira_key_body":    string(domain.ValidationErrorMissingJira),
		"missing_jira_key_subject": string(domain.ValidationErrorMissingJira),
		"key_not_at_end":           string(domain.ValidationErrorInvalidFormat),
		"invalid_project":          string(domain.ValidationErrorInvalidType),
		"invalid_refs_format":      string(domain.ValidationErrorInvalidFormat),
		"refs_after_signoff":       string(domain.ValidationErrorInvalidFormat),
		"invalid_key_format":       string(domain.ValidationErrorInvalidFormat),
	}

	// Test cases for subject validation (default behavior)
	subjectTestCases := []struct {
		name                 string
		subject              string
		body                 string
		isConventionalCommit bool
		validProjects        []string
		useBodyRef           bool
		expectedValid        bool
		expectedCode         string
		errorContains        string
	}{
		// Conventional Commit Positive Cases
		{
			name:                 "Valid conventional commit with jira key at end",
			subject:              "feat(auth): add user authentication [PROJ-123]",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        true,
		},
		{
			name:                 "Valid conventional commit with multiple words jira key",
			subject:              "fix(profile): resolve user profile update issue [TEAM-456]",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        true,
		},
		{
			name:                 "Valid conventional commit with multiline message",
			subject:              "refactor(api): simplify authentication middleware [CORE-789]\n\nAdditional context about the change",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        true,
		},
		{
			name:                 "Valid conventional commit with parentheses format",
			subject:              "docs(readme): update installation instructions (PROJ-123)",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        true,
		},
		{
			name:                 "Valid conventional commit with plain key at end",
			subject:              "chore(deps): update dependencies TEAM-456",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        true,
		},

		// Conventional Commit Negative Cases
		{
			name:                 "Conventional commit missing jira key",
			subject:              "feat(auth): add user authentication",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        false,
			expectedCode:         "missing_jira",
			errorContains:        "no Jira issue key found",
		},
		{
			name:                 "Conventional commit jira key not at end",
			subject:              "feat(auth): [PROJ-123] add user authentication",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        false,
			expectedCode:         "invalid_format",
			errorContains:        "must be at the end of the first line",
		},
		{
			name:                 "Conventional commit invalid jira project",
			subject:              "feat(auth): add user authentication [UNKNOWN-123]",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        false,
			expectedCode:         "invalid_type",
			errorContains:        "not a valid project",
		},
		{
			name:                 "Empty subject",
			subject:              "",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        false,
			expectedCode:         "invalid_format",
			errorContains:        "commit subject is empty",
		},
		{
			name:                 "Invalid Jira key format",
			subject:              "feat(auth): add user authentication [proj-123]", // lowercase project
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedValid:        false,
			expectedCode:         "missing_jira",
			errorContains:        "no Jira issue key found",
		},

		// Non-Conventional Commit Positive Cases
		{
			name:                 "Valid non-conventional commit anywhere",
			subject:              "PROJ-123 Implement user authentication",
			isConventionalCommit: false,
			validProjects:        validProjects,
			expectedValid:        true,
		},
		{
			name:                 "Valid non-conventional commit multiple issues",
			subject:              "Implement PROJ-123 and TEAM-456 features",
			isConventionalCommit: false,
			validProjects:        validProjects,
			expectedValid:        true,
		},
		{
			name:                 "Valid non-conventional commit with prefix format",
			subject:              "CORE-789: Implement new authentication flow",
			isConventionalCommit: false,
			validProjects:        validProjects,
			expectedValid:        true,
		},
		{
			name:                 "Valid non-conventional commit with bracket prefix",
			subject:              "[PROJ-123] Fix bug in login form",
			isConventionalCommit: false,
			validProjects:        validProjects,
			expectedValid:        true,
		},
		{
			name:                 "Valid non-conventional commit with dash format",
			subject:              "TEAM-456 - Update user interface components",
			isConventionalCommit: false,
			validProjects:        validProjects,
			expectedValid:        true,
		},

		// Non-Conventional Commit Negative Cases
		{
			name:                 "Non-conventional commit missing jira key",
			subject:              "Implement user authentication",
			isConventionalCommit: false,
			validProjects:        validProjects,
			expectedValid:        false,
			expectedCode:         "missing_jira",
			errorContains:        "no Jira issue key found",
		},
		{
			name:                 "Non-conventional commit invalid jira project",
			subject:              "Implement UNKNOWN-123 feature",
			isConventionalCommit: false,
			validProjects:        validProjects,
			expectedValid:        false,
			expectedCode:         "invalid_type",
			errorContains:        "not a valid project",
		},

		// Body validation positive cases
		{
			name:          "Valid body reference single key",
			subject:       "feat(auth): add user authentication", // No Jira key needed in subject
			body:          "This adds JWT authentication.\n\nRefs: PROJ-123",
			useBodyRef:    true,
			validProjects: validProjects,
			expectedValid: true,
		},
		{
			name:          "Valid body reference multiple keys",
			subject:       "Fix multiple bugs",
			body:          "This fixes several issues.\n\nRefs: PROJ-123, TEAM-456",
			useBodyRef:    true,
			validProjects: validProjects,
			expectedValid: true,
		},
		{
			name:          "Valid body reference with signed-off-by after",
			subject:       "Update documentation",
			body:          "Updated API docs.\n\nRefs: CORE-789\n\nSigned-off-by: Developer <dev@example.com>",
			useBodyRef:    true,
			validProjects: validProjects,
			expectedValid: true,
		},

		// Body validation negative cases
		{
			name:          "Missing Refs in body",
			subject:       "Fix bug",
			body:          "This fixes a critical bug.",
			useBodyRef:    true,
			validProjects: validProjects,
			expectedValid: false,
			expectedCode:  "missing_jira_key_body",
			errorContains: "no Jira issue key found in the commit body",
		},
		{
			name:          "Invalid Refs format",
			subject:       "Update API",
			body:          "API changes.\n\nRefs: invalid-key",
			useBodyRef:    true,
			validProjects: validProjects,
			expectedValid: false,
			expectedCode:  "invalid_refs_format",
			errorContains: "invalid Refs format",
		},
		{
			name:          "Refs line after signed-off-by",
			subject:       "Refactor code",
			body:          "Major refactoring.\n\nSigned-off-by: Developer <dev@example.com>\nRefs: PROJ-123",
			useBodyRef:    true,
			validProjects: validProjects,
			expectedValid: false,
			expectedCode:  "refs_after_signoff",
			errorContains: "Refs: line must appear before any Signed-off-by lines",
		},
		{
			name:          "Empty body",
			subject:       "Fix bug",
			body:          "",
			useBodyRef:    true,
			validProjects: validProjects,
			expectedValid: false,
			expectedCode:  "missing_jira_key_body",
			errorContains: "no Jira issue key found in the commit body",
		},
	}

	for _, testCase := range subjectTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Build options based on test case
			var options []rules.JiraReferenceOption

			if len(testCase.validProjects) > 0 {
				options = append(options, rules.WithValidProjects(testCase.validProjects))
			}

			if testCase.isConventionalCommit {
				options = append(options, rules.WithConventionalCommit())
			}

			if testCase.useBodyRef {
				options = append(options, rules.WithBodyRefChecking())
			}

			// Create the rule instance
			rule := rules.NewJiraReferenceRule(options...)

			// Create a commit for testing
			commit := &domain.CommitInfo{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			// Execute validation
			errors := rule.Validate(commit)

			// Check for expected validation result
			if testCase.expectedValid {
				assert.Empty(t, errors, "Expected no errors but got: %v", errors)
			} else {
				assert.NotEmpty(t, errors, "Expected errors but found none")

				// Check error code if specified
				if testCase.expectedCode != "" {
					// Get the mapped code for the test
					mappedCode, ok := errorCodeMap[testCase.expectedCode]
					if !ok {
						t.Logf("Warning: No mapping found for expected code %s", testCase.expectedCode)
						mappedCode = testCase.expectedCode
					}

					assert.Equal(t, mappedCode, errors[0].Code,
						"Error code should match expected mapped code")
				}

				// Check error message contains expected substring
				if testCase.errorContains != "" {
					found := false

					for _, err := range errors {
						if strings.Contains(err.Error(), testCase.errorContains) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error containing %q", testCase.errorContains)
				}

				// Verify rule name is set in ValidationError
				assert.Equal(t, "JiraReference", errors[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify Help() method provides guidance
				helpText := rule.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")

				// Test specific help messages based on expected code
				if testCase.expectedCode == "empty_subject" {
					assert.Contains(t, helpText, "Provide a non-empty", "Help should mention providing a non-empty message")
				} else if testCase.expectedCode == "missing_jira_key_subject" || testCase.expectedCode == "missing_jira_key_body" {
					assert.Contains(t, helpText, "Include a valid Jira", "Help should mention including a valid Jira key")
				} else if testCase.expectedCode == "key_not_at_end" {
					assert.Contains(t, helpText, "place the Jira issue key at the end", "Help should mention key placement")
				} else if testCase.expectedCode == "invalid_project" {
					assert.Contains(t, helpText, "not recognized as a valid project", "Help should explain invalid project")
				} else if testCase.expectedCode == "invalid_refs_format" {
					assert.Contains(t, helpText, "format", "Help should mention the correct format")
				} else if testCase.expectedCode == "refs_after_signoff" {
					assert.Contains(t, helpText, "must appear before", "Help should explain line ordering")
				}
			}

			// Verify Name() method
			assert.Equal(t, "JiraReference", rule.Name(), "Name should be 'JiraReference'")

			// Verify Result() and VerboseResult() methods return expected messages
			if testCase.expectedValid {
				assert.Equal(t, "Valid Jira reference", rule.Result(), "Expected default valid message")
				assert.Contains(t, rule.VerboseResult(), "Valid Jira reference", "Verbose result should indicate valid reference")
			} else {
				assert.Equal(t, "Missing or invalid Jira reference", rule.Result(), "Expected default error message")
				assert.NotEqual(t, rule.Result(), rule.VerboseResult(), "Verbose result should be different from regular result")
			}
		})
	}

	// Test with empty project list (should validate format only)
	t.Run("Empty project list", func(t *testing.T) {
		// Valid format should pass
		rule := rules.NewJiraReferenceRule(rules.WithConventionalCommit())
		commit := &domain.CommitInfo{
			Subject: "feat: add feature ABC-123",
			Body:    "",
		}
		errors := rule.Validate(commit)
		assert.Empty(t, errors, "Should accept any project key with valid format")

		// Invalid format should fail
		commit = &domain.CommitInfo{
			Subject: "feat: add feature abc-123",
			Body:    "",
		}
		errors = rule.Validate(commit)
		assert.NotEmpty(t, errors, "Should reject invalid format")
		assert.Equal(t, string(domain.ValidationErrorMissingJira), errors[0].Code,
			"Error code should be ValidationErrorMissingJira")
	})

	// Test context data in errors
	t.Run("Context data in errors", func(t *testing.T) {
		// Test invalid project error context
		rule := rules.NewJiraReferenceRule(
			rules.WithConventionalCommit(),
			rules.WithValidProjects(validProjects),
		)
		commit := &domain.CommitInfo{
			Subject: "feat: add feature INVALID-123",
			Body:    "",
		}
		errors := rule.Validate(commit)
		require.NotEmpty(t, errors, "Should have errors")
		assert.Equal(t, string(domain.ValidationErrorInvalidType), errors[0].Code, "Error code should be ValidationErrorInvalidType")
		assert.Equal(t, "INVALID", errors[0].Context["project"],
			"Context should contain the invalid project")
		assert.Contains(t, errors[0].Context["valid_projects"], validProjects[0],
			"Context should list valid projects")

		// Test key not at end error context
		commit = &domain.CommitInfo{
			Subject: "feat: PROJ-123 add feature",
			Body:    "",
		}
		errors = rule.Validate(commit)
		require.NotEmpty(t, errors, "Should have errors")
		assert.Equal(t, string(domain.ValidationErrorInvalidFormat), errors[0].Code, "Error code should be ValidationErrorInvalidFormat")
		assert.Equal(t, "PROJ-123", errors[0].Context["key"],
			"Context should contain the key that's not at the end")
	})
}
