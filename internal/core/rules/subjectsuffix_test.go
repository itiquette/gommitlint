// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testconfig "github.com/itiquette/gommitlint/internal/testutils/config"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

func TestSubjectSuffixRule(t *testing.T) {
	testCases := []struct {
		name            string
		subject         string
		invalidSuffixes string
		expectedValid   bool
		expectedCode    string
	}{
		{
			name:            "Valid subject without invalid suffix",
			subject:         "Add new feature",
			invalidSuffixes: ".:;",
			expectedValid:   true,
		},
		{
			name:            "Subject ending with invalid suffix period",
			subject:         "Update documentation.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Subject ending with invalid suffix colon",
			subject:         "Fix bug:",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Unicode subject with invalid suffix",
			subject:         "Fix élément.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Unicode character as invalid suffix",
			subject:         "Update description;",
			invalidSuffixes: ";",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Empty subject",
			subject:         "",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrMissingSubject),
		},
		{
			name:            "Subject with Unicode invalid suffix",
			subject:         "Add new emoji😊",
			invalidSuffixes: "😊", // Keep just one emoji to simplify the test
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Subject with space as invalid suffix",
			subject:         "Add feature ",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Subject with tab as invalid suffix",
			subject:         "Add feature\t",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
		{
			name:            "Valid Unicode subject",
			subject:         "修复问题",
			invalidSuffixes: ".:;",
			expectedValid:   true,
		},
		{
			name:            "Default invalid suffixes",
			subject:         "Update feature?",
			invalidSuffixes: ".,:;?",
			expectedValid:   false,
			expectedCode:    string(appErrors.ErrSubjectSuffix),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the rule with value semantics
			var rule rules.SubjectSuffixRule
			if testCase.invalidSuffixes != "" {
				rule = rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes(testCase.invalidSuffixes))
			} else {
				rule = rules.NewSubjectSuffixRule() // Use default suffixes
			}

			// Create commit with the test subject
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Setup context with custom config for this test
			var suffixes []string
			if testCase.invalidSuffixes != "" {
				suffixes = strings.Split(testCase.invalidSuffixes, "")
			}

			builder := testconfig.NewBuilder().
				WithSubjectSuffixes(suffixes)
			cfg := builder.Build()
			testConfig := testconfig.NewAdapter(cfg).Adapter

			ctx := testcontext.CreateTestContext()
			ctx = contextx.WithConfig(ctx, testConfig)

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Check results
			if testCase.expectedValid {
				require.Empty(t, errors, "Expected no validation errors")
			} else {
				require.NotEmpty(t, errors, "Expected validation errors")

				if len(errors) > 0 && testCase.expectedCode != "" {
					require.Equal(t, testCase.expectedCode, errors[0].Code, "Error code should match expected")
				}
			}

			// Verify name
			require.Equal(t, "SubjectSuffix", rule.Name())
		})
	}
}

func TestSubjectSuffixOptions(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		// No options provided, should use default invalid suffixes
		rule := rules.NewSubjectSuffixRule()

		// Create a valid commit
		validCommit := domain.CommitInfo{
			Subject: "This is valid",
		}

		// Create an invalid commit
		invalidCommit := domain.CommitInfo{
			Subject: "This ends with period.",
		}

		// Setup context with our test config adapter
		builder := testconfig.NewBuilder().
			WithSubjectSuffixes([]string{".", ","})
		cfg := builder.Build()
		testConfig := testconfig.NewAdapter(cfg).Adapter

		ctx := testcontext.CreateTestContext()
		ctx = contextx.WithConfig(ctx, testConfig)

		// Test valid case
		validErrors := rule.Validate(ctx, validCommit)
		require.Empty(t, validErrors, "Default config should accept valid subject")

		// Test invalid case
		invalidErrors := rule.Validate(ctx, invalidCommit)
		require.NotEmpty(t, invalidErrors, "Default config should reject subject ending with period")
	})

	t.Run("With custom invalid suffixes", func(t *testing.T) {
		// Custom invalid suffixes
		rule := rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes("!@#"))

		// Create commits for testing
		invalidCommit := domain.CommitInfo{
			Subject: "This ends with exclamation!",
		}

		validCommit := domain.CommitInfo{
			Subject: "This ends with period.",
		}

		// Setup context with custom config using our test adapter
		builder := testconfig.NewBuilder().
			WithSubjectSuffixes([]string{"!", "@", "#"})
		cfg := builder.Build()
		testConfig := testconfig.NewAdapter(cfg).Adapter

		ctx := testcontext.CreateTestContext()
		ctx = contextx.WithConfig(ctx, testConfig)

		// Test invalid case
		invalidErrors := rule.Validate(ctx, invalidCommit)
		require.NotEmpty(t, invalidErrors, "Should reject subject with configured invalid suffix")

		// Test valid case
		validErrors := rule.Validate(ctx, validCommit)
		require.Empty(t, validErrors, "Should accept subject ending with period when not in invalid set")
	})

	t.Run("Empty invalid suffixes", func(t *testing.T) {
		// Empty invalid suffixes should fall back to defaults
		rule := rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes(""))

		// Create commit with question mark (in default invalid suffixes)
		commit := domain.CommitInfo{
			Subject: "This ends with question mark?",
		}

		// Setup context with config that checks for ?
		builder := testconfig.NewBuilder().
			WithSubjectSuffixes([]string{"?"})
		cfg := builder.Build()
		testConfig := testconfig.NewAdapter(cfg).Adapter

		ctx := testcontext.CreateTestContext()
		ctx = contextx.WithConfig(ctx, testConfig)

		// Test invalid case
		errors := rule.Validate(ctx, commit)
		require.NotEmpty(t, errors, "Should reject subject with default invalid suffix")
	})
}

func TestSubjectSuffixRuleWithCustomOptions(t *testing.T) {
	// Create rule with options
	rule := rules.NewSubjectSuffixRule(
		rules.WithInvalidSuffixes("!@#"),
	)

	// Create test commit
	commit := domain.CommitInfo{
		Subject: "Test with exclamation!",
	}

	// Setup context with config that matches rule options using our test adapter
	builder := testconfig.NewBuilder().
		WithSubjectSuffixes([]string{"!", "@", "#"})
	cfg := builder.Build()
	testConfig := testconfig.NewAdapter(cfg).Adapter

	ctx := testcontext.CreateTestContext()
	ctx = contextx.WithConfig(ctx, testConfig)

	// Test validation
	errors := rule.Validate(ctx, commit)
	require.NotEmpty(t, errors, "Should return errors for invalid subject")

	// Simple validation of rule name
	require.Equal(t, "SubjectSuffix", rule.Name())
}

func TestSubjectSuffixRuleWithConfig(t *testing.T) {
	tests := []struct {
		name            string
		invalidSuffixes []string
		subject         string
		expectErrors    bool
		description     string
	}{
		{
			name:            "Default invalid suffixes - valid subject",
			invalidSuffixes: []string{".", ","},
			subject:         "Add new feature",
			expectErrors:    false,
			description:     "Should pass with default suffixes and valid subject",
		},
		{
			name:            "Default invalid suffixes - invalid subject",
			invalidSuffixes: []string{".", ","},
			subject:         "Add new feature.",
			expectErrors:    true,
			description:     "Should fail with default suffixes and subject ending with period",
		},
		{
			name:            "Custom invalid suffixes - valid with default invalid suffix",
			invalidSuffixes: []string{"!", "?"},
			subject:         "Add new feature.", // Period is allowed with custom config
			expectErrors:    false,
			description:     "Should pass when period is not in custom invalid suffixes",
		},
		{
			name:            "Custom invalid suffixes - invalid with custom suffix",
			invalidSuffixes: []string{"!", "?"},
			subject:         "Add new feature!", // Exclamation mark is not allowed
			expectErrors:    true,
			description:     "Should fail when ending with a character in custom invalid suffixes",
		},
		{
			name:            "Empty subject",
			invalidSuffixes: []string{".", ","},
			subject:         "",
			expectErrors:    true,
			description:     "Should fail with empty subject",
		},
		{
			name:            "Unicode invalid suffixes",
			invalidSuffixes: []string{"😊"},
			subject:         "Add new emoji😊",
			expectErrors:    true,
			description:     "Should fail with Unicode invalid suffixes",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with our test adapter
			builder := testconfig.NewBuilder().
				WithSubjectSuffixes(testCase.invalidSuffixes)
			cfg := builder.Build()
			testConfig := testconfig.NewAdapter(cfg).Adapter

			// Add config to context
			ctx := testcontext.CreateTestContext()
			ctx = contextx.WithConfig(ctx, testConfig)

			// Create rule
			rule := rules.NewSubjectSuffixRule()

			// Create test commit
			commit := domain.CommitInfo{
				Hash:    "abc123",
				Subject: testCase.subject,
				Message: testCase.subject,
			}

			// Execute validation
			errors := rule.Validate(ctx, commit)

			// Check results
			if testCase.expectErrors {
				require.NotEmpty(t, errors, "Expected validation errors")
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}

			// Check rule name
			require.Equal(t, "SubjectSuffix", rule.Name())
		})
	}
}
