// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidationWorkflow tests the complete validation workflow.
func TestValidationWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantPass  bool
		wantRules []string // Rules that should trigger
	}{
		{
			name:     "Valid conventional commit",
			message:  "feat: add new feature",
			wantPass: true,
		},
		{
			name:      "Invalid - too long subject",
			message:   "feat: this is a very long commit message that exceeds the maximum length allowed by our configuration",
			wantPass:  false,
			wantRules: []string{"Subject", "ConventionalCommit"},
		},
		{
			name:      "Invalid - not conventional format",
			message:   "Add new feature",
			wantPass:  false,
			wantRules: []string{"ConventionalCommit"},
		},
		{
			name:     "Valid - short conventional",
			message:  "fix: bug",
			wantPass: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := TestValidateMessage(t, testCase.message, DefaultConfig())

			if testCase.wantPass {
				require.True(t, result.Valid, "Expected validation to pass")
				require.Empty(t, result.Errors, "Expected no validation errors")
			} else {
				require.False(t, result.Valid, "Expected validation to fail")
				require.NotEmpty(t, result.Errors, "Expected validation errors")

				// Verify specific rules triggered if specified
				if len(testCase.wantRules) > 0 {
					foundRules := make(map[string]bool)
					for _, err := range result.Errors {
						foundRules[err.Rule] = true
					}

					for _, expectedRule := range testCase.wantRules {
						require.True(t, foundRules[expectedRule],
							"Expected rule %s to trigger", expectedRule)
					}
				}
			}
		})
	}
}

// TestMessageFileValidation tests validation of commit message files.
func TestMessageFileValidation(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantPass bool
	}{
		{
			name:     "Valid commit message with body",
			message:  "feat: add authentication\n\nThis adds JWT authentication to the API.",
			wantPass: true,
		},
		{
			name:     "Valid commit message without body",
			message:  "fix: resolve parsing error",
			wantPass: true,
		},
		{
			name:     "Invalid - empty message",
			message:  "",
			wantPass: false,
		},
		{
			name:     "Invalid - non-conventional format",
			message:  "adding new feature without proper format",
			wantPass: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := TestValidateMessage(t, testCase.message, DefaultConfig())

			if testCase.wantPass {
				require.True(t, result.Valid, "Expected validation to pass")
			} else {
				require.False(t, result.Valid, "Expected validation to fail")
			}
		})
	}
}

// TestRuleSpecificValidation tests validation with specific rule configurations.
func TestRuleSpecificValidation(t *testing.T) {
	t.Run("Only subject length validation", func(t *testing.T) {
		config := WithRules("Subject")

		// Valid length
		result := TestValidateMessage(t, "feat: short message", config)
		require.True(t, result.Valid)

		// Invalid length
		longMessage := "feat: " + string(make([]byte, 100)) // Very long message
		result = TestValidateMessage(t, longMessage, config)
		require.False(t, result.Valid)
	})

	t.Run("Only conventional commit validation", func(t *testing.T) {
		config := WithRules("ConventionalCommit")

		// Valid conventional
		result := TestValidateMessage(t, "feat: add feature", config)
		require.True(t, result.Valid)

		// Invalid conventional
		result = TestValidateMessage(t, "Add feature", config)
		require.False(t, result.Valid)
	})

	t.Run("Custom subject length", func(t *testing.T) {
		config := WithSubjectMaxLength(20)

		// Valid short message
		result := TestValidateMessage(t, "feat: short", config)
		require.True(t, result.Valid)

		// Invalid long message
		result = TestValidateMessage(t, "feat: this is too long", config)
		require.False(t, result.Valid)
	})
}

// TestRepositoryValidation tests validation against git repositories.
func TestRepositoryValidation(t *testing.T) {
	t.Run("Valid repository commit", func(t *testing.T) {
		repoPath, cleanup := GitRepo(t, "feat: add authentication\n\nAdds JWT auth")
		defer cleanup()

		result := TestValidation(t, repoPath, DefaultConfig())
		require.True(t, result.Valid)
	})

	t.Run("Invalid repository commit", func(t *testing.T) {
		repoPath, cleanup := GitRepo(t, "bad commit message format")
		defer cleanup()

		result := TestValidation(t, repoPath, DefaultConfig())
		require.False(t, result.Valid)
	})

	t.Run("Repository with custom config", func(t *testing.T) {
		repoPath, cleanup := GitRepo(t, "feat: this message is longer than 30 chars")
		defer cleanup()

		config := WithSubjectMaxLength(30)
		result := TestValidation(t, repoPath, config)
		require.False(t, result.Valid)
	})
}

// TestFunctionalConfigComposition tests composition of configuration functions.
func TestFunctionalConfigComposition(t *testing.T) {
	t.Run("Combine multiple config changes", func(t *testing.T) {
		// Start with default and apply multiple changes
		config := DefaultConfig()
		config.Message.Subject.MaxLength = 30
		config.Conventional.Types = []string{"custom", "special"}

		// Valid with custom type and short length
		result := TestValidateMessage(t, "custom: short", config)
		require.True(t, result.Valid)

		// Invalid with custom type but long length
		result = TestValidateMessage(t, "custom: this is way too long for our limit", config)
		require.False(t, result.Valid)
	})
}
