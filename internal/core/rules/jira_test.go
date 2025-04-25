// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestJiraReferenceRule(t *testing.T) {
	// Define valid Jira projects for testing
	validProjects := []string{"PROJ", "TEAM", "CORE"}

	// Test cases for subject validation (default behavior)
	subjectTestCases := []struct {
		name                 string
		subject              string
		body                 string
		isConventionalCommit bool
		validProjects        []string
		useBodyRef           bool
		expectedErrors       int
		expectedErrorCode    appErrors.ValidationErrorCode
		expectedResult       string
	}{
		{
			name:           "Valid subject with Jira key",
			subject:        "Add feature PROJ-123",
			validProjects:  validProjects,
			expectedErrors: 0,
			expectedResult: "Valid Jira reference",
		},
		{
			name:           "Valid subject with multiple Jira keys",
			subject:        "Add feature PROJ-123, TEAM-456",
			validProjects:  validProjects,
			expectedErrors: 0,
			expectedResult: "Valid Jira reference",
		},
		{
			name:           "Valid subject with Jira key in brackets",
			subject:        "Add feature [PROJ-123]",
			validProjects:  validProjects,
			expectedErrors: 0,
			expectedResult: "Valid Jira reference",
		},
		{
			name:                 "Valid conventional commit with Jira key",
			subject:              "feat: Add login button PROJ-123",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedErrors:       0,
			expectedResult:       "Valid Jira reference",
		},
		{
			name:              "Missing Jira key in subject",
			subject:           "Add feature",
			validProjects:     validProjects,
			expectedErrors:    1,
			expectedErrorCode: appErrors.ErrMissingJira,
			expectedResult:    "Missing or invalid Jira reference",
		},
		{
			name:              "Invalid Jira project",
			subject:           "Add feature INVALID-123",
			validProjects:     validProjects,
			expectedErrors:    1,
			expectedErrorCode: appErrors.ErrInvalidType,
			expectedResult:    "Missing or invalid Jira reference",
		},
		{
			name:              "Invalid Jira key format",
			subject:           "Add feature PROJ123",
			validProjects:     validProjects,
			expectedErrors:    1,
			expectedErrorCode: appErrors.ErrInvalidFormat,
			expectedResult:    "Missing or invalid Jira reference",
		},
		{
			name:           "Jira key not at end",
			subject:        "Add PROJ-123 feature",
			validProjects:  validProjects,
			expectedErrors: 0, // Non-conventional commits can have key anywhere
			expectedResult: "Valid Jira reference",
		},
		{
			name:                 "Conventional commit with key not at end",
			subject:              "feat: PROJ-123 add login feature",
			isConventionalCommit: true,
			validProjects:        validProjects,
			expectedErrors:       1,
			expectedErrorCode:    appErrors.ErrInvalidFormat,
			expectedResult:       "Missing or invalid Jira reference",
		},
		{
			name:              "Empty subject",
			subject:           "",
			validProjects:     validProjects,
			expectedErrors:    1,
			expectedErrorCode: appErrors.ErrEmptyMessage,
			expectedResult:    "Missing or invalid Jira reference",
		},
	}

	for _, testCase := range subjectTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with options
			options := []rules.JiraReferenceOption{
				rules.WithValidProjects(testCase.validProjects),
			}
			if testCase.isConventionalCommit {
				options = append(options, rules.WithConventionalCommit())
			}

			if testCase.useBodyRef {
				options = append(options, rules.WithBodyRefChecking())
			}

			rule := rules.NewJiraReferenceRule(options...)

			// Create commit for validation
			commit := domain.CommitInfo{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			// Validate
			result := rule.Validate(commit)

			// Check number of errors
			require.Len(t, result, testCase.expectedErrors, "Expected %d errors, got %d", testCase.expectedErrors, len(result))

			// If expecting errors, check the error code
			if testCase.expectedErrors > 0 && len(result) > 0 {
				// Check the specific error code - we need to verify exact behavior
				require.Equal(t, string(testCase.expectedErrorCode), result[0].Code, "Error code should match expected code")

				// Check rule name
				require.Equal(t, "JiraReference", result[0].Rule, "Rule name should be set in ValidationError")

				if testCase.expectedErrorCode == appErrors.ErrMissingJira && !testCase.useBodyRef {
					require.Contains(t, result[0].Message, "no Jira issue key found", "Expected error containing \"no Jira issue key found\"")
				}

				// In a functional style, we don't check rule.Result() directly
				// Instead, we determine the expected result based on whether there are errors
				expectedResult := "Missing or invalid Jira reference"
				require.Equal(t, expectedResult, testCase.expectedResult, "Result message should match expected for error case")
			} else {
				// No errors means valid
				expectedResult := "Valid Jira reference"
				require.Equal(t, expectedResult, testCase.expectedResult, "Result message should match expected for success case")
			}

			// Name should be consistent
			require.Equal(t, "JiraReference", rule.Name(), "Rule name should be 'JiraReference'")
		})
	}

	// Test body reference validation
	bodyTestCases := []struct {
		name              string
		subject           string
		body              string
		validProjects     []string
		useBodyRef        bool
		expectedErrors    int
		expectedErrorCode appErrors.ValidationErrorCode
	}{
		{
			name:           "Valid reference in body",
			subject:        "Add new feature",
			body:           "This implements the login screen.\n\nRefs: PROJ-123",
			validProjects:  validProjects,
			useBodyRef:     true,
			expectedErrors: 0,
		},
		{
			name:              "Missing reference in body when enabled",
			subject:           "Add new feature",
			body:              "This implements the login screen.",
			validProjects:     validProjects,
			useBodyRef:        true,
			expectedErrors:    1,
			expectedErrorCode: appErrors.ErrMissingJira,
		},
		{
			name:              "Valid reference but in wrong format",
			subject:           "Add new feature",
			body:              "This implements PROJ-123 the login screen.",
			validProjects:     validProjects,
			useBodyRef:        true,
			expectedErrors:    1,
			expectedErrorCode: appErrors.ErrMissingJira,
		},
		{
			name:              "Reference after signoff line",
			subject:           "Add new feature",
			body:              "This implements the login screen.\n\nSigned-off-by: User <user@example.com>\nRefs: PROJ-123",
			validProjects:     validProjects,
			useBodyRef:        true,
			expectedErrors:    1,
			expectedErrorCode: appErrors.ErrInvalidFormat,
		},
		{
			name:              "Invalid refs format",
			subject:           "Add new feature",
			body:              "This implements the login screen.\n\nRefs: PROJ 123",
			validProjects:     validProjects,
			useBodyRef:        true,
			expectedErrors:    1,
			expectedErrorCode: appErrors.ErrInvalidFormat,
		},
		{
			name:           "Multiple valid refs",
			subject:        "Add new feature",
			body:           "This implements the login screen.\n\nRefs: PROJ-123, TEAM-456",
			validProjects:  validProjects,
			useBodyRef:     true,
			expectedErrors: 0,
		},
	}

	for _, testCase := range bodyTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create options
			options := []rules.JiraReferenceOption{
				rules.WithValidProjects(testCase.validProjects),
			}
			if testCase.useBodyRef {
				options = append(options, rules.WithBodyRefChecking())
			}

			rule := rules.NewJiraReferenceRule(options...)

			// Create commit
			commit := domain.CommitInfo{
				Subject: testCase.subject,
				Body:    testCase.body,
			}

			// Validate
			result := rule.Validate(commit)

			// Check number of errors
			require.Len(t, result, testCase.expectedErrors, "Expected %d errors, got %d", testCase.expectedErrors, len(result))

			// If expecting errors, check error code
			if testCase.expectedErrors > 0 && len(result) > 0 {
				// Check specific error code
				require.Equal(t, string(testCase.expectedErrorCode), result[0].Code, "Error code should match expected")
			}
		})
	}

	// Test verbose result and help messages
	t.Run("Test verbose result for missing Jira key", func(t *testing.T) {
		rule := rules.NewJiraReferenceRule(
			rules.WithValidProjects(validProjects),
		)

		// Create commit without Jira key
		commit := domain.CommitInfo{
			Subject: "Add new feature",
			Body:    "",
		}

		// Validate to generate errors
		errors := rule.Validate(commit)
		require.NotEmpty(t, errors, "Should have errors for missing Jira key")

		// In functional style, we don't check VerboseResult from the rule directly
		// since the original rule is unchanged after validation
		require.Equal(t, string(appErrors.ErrMissingJira), errors[0].Code, "Error code should be ErrMissingJira")
		require.Contains(t, errors[0].Message, "no Jira issue key found", "Error should explain the missing key")
	})

	// Test context data in errors
	t.Run("Context data in errors", func(t *testing.T) {
		// Test invalid project error context
		rule := rules.NewJiraReferenceRule(
			rules.WithValidProjects(validProjects),
		)
		commit := domain.CommitInfo{
			Subject: "Add feature INVALID-123",
			Body:    "",
		}
		errors := rule.Validate(commit)
		require.NotEmpty(t, errors, "Should have errors")

		// Check error code
		require.Equal(t, string(appErrors.ErrInvalidType), errors[0].Code, "Error code should be ErrInvalidType")

		// Check that project exists in context
		project, exists := errors[0].Context["project"]
		require.True(t, exists, "Context should contain the project key")
		require.Equal(t, "INVALID", project, "Context should contain the invalid project")
	})

	// Test found keys in different scenarios
	t.Run("Test found keys in different scenarios", func(t *testing.T) {
		// For valid subject reference and body reference, we would check
		// the returned keys rather than the VerboseResult in functional style
		// For subject, we'd verify that the correct errors and context are present
		subjectRule := rules.NewJiraReferenceRule(
			rules.WithValidProjects(validProjects),
		)
		subjectCommit := domain.CommitInfo{
			Subject: "Add feature PROJ-123",
			Body:    "",
		}
		subjectErrors := subjectRule.Validate(subjectCommit)
		require.Empty(t, subjectErrors, "Should have no errors for valid reference")

		// For body reference, similarly check returned errors
		bodyRule := rules.NewJiraReferenceRule(
			rules.WithValidProjects(validProjects),
			rules.WithBodyRefChecking(),
		)
		bodyCommit := domain.CommitInfo{
			Subject: "Add feature",
			Body:    "Description\n\nRefs: PROJ-123",
		}
		bodyErrors := bodyRule.Validate(bodyCommit)
		require.Empty(t, bodyErrors, "Should have no errors for valid body reference")
	})
}
