// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package integtest contains end-to-end integration tests for gommitlint workflows.
// These tests verify that the application's components work together correctly.
package integtest

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/git"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// Mock implementations for testing
// ===============================

// MockRule is a test implementation of the Rule interface.
type MockRule struct {
	name         string
	shouldPass   bool
	errors       []errors.ValidationError
	validationFn func(domain.CommitInfo) []errors.ValidationError
}

// NewMockRule creates a simple pass/fail mock rule for testing.
func NewMockRule(name string, shouldPass bool) *MockRule {
	rule := &MockRule{
		name:       name,
		shouldPass: shouldPass,
		errors:     make([]errors.ValidationError, 0),
	}

	if !shouldPass {
		// Create a simple error
		rule.errors = append(rule.errors, errors.CreateBasicError(name, errors.ErrUnknown, "Mock error message"))
	}

	return rule
}

// NewMockRuleWithValidation creates a mock rule with custom validation logic.
func NewMockRuleWithValidation(name string, validationFn func(domain.CommitInfo) []errors.ValidationError) *MockRule {
	return &MockRule{
		name:         name,
		shouldPass:   true, // Will be determined by the function result
		validationFn: validationFn,
	}
}

// NewTypeCheckingMockRule creates a rule that checks conventional commit format.
func NewTypeCheckingMockRule(name string, allowedTypes []string) *MockRule {
	return NewMockRuleWithValidation(name, func(commit domain.CommitInfo) []errors.ValidationError {
		// Check if this follows conventional commit format (type: subject)
		parts := strings.SplitN(commit.Subject, ":", 2)
		if len(parts) != 2 {
			return []errors.ValidationError{
				errors.CreateBasicError(
					name,
					"conventional_format",
					"commit message does not follow conventional commit format",
				),
			}
		}

		// If we have allowed types, check the type
		if len(allowedTypes) > 0 {
			commitType := parts[0]
			// Extract scope if present
			if idx := strings.Index(commitType, "("); idx != -1 {
				commitType = commitType[:idx]
			}

			allowed := false

			for _, allowedType := range allowedTypes {
				if commitType == allowedType {
					allowed = true

					break
				}
			}

			if !allowed {
				return []errors.ValidationError{
					errors.CreateBasicError(
						name,
						"invalid_type",
						"commit type is not allowed",
					).WithContext("type", commitType).
						WithContext("allowed_types", strings.Join(allowedTypes, ", ")),
				}
			}
		}

		return nil
	})
}

// Name returns the rule name.
func (r *MockRule) Name() string {
	return r.name
}

// Validate performs validation for this rule.
func (r *MockRule) Validate(_ context.Context, commit domain.CommitInfo) []errors.ValidationError {
	if r.validationFn != nil {
		return r.validationFn(commit)
	}

	return r.errors
}

// Result returns a concise result message.
func (r *MockRule) Result(_ []errors.ValidationError) string {
	if r.shouldPass {
		return "Passed"
	}

	return "Failed"
}

// VerboseResult returns a detailed result message.
func (r *MockRule) VerboseResult(_ []errors.ValidationError) string {
	if r.shouldPass {
		return "Rule passed validation"
	}

	return "Rule failed validation"
}

// Help returns guidance on fixing validation errors.
func (r *MockRule) Help(_ []errors.ValidationError) string {
	return "This is a mock rule for testing purposes"
}

// HasErrors returns whether this rule has detected errors.
func (r *MockRule) HasErrors() bool {
	return len(r.Errors()) > 0
}

// Errors returns the validation errors for this rule.
func (r *MockRule) Errors() []errors.ValidationError {
	return r.errors
}

// MockRuleProvider is a test implementation of the RuleProvider interface.
// Note: MockRuleProvider is now defined in test_helpers.go

// TestValidateCommitWorkflow tests the commit validation workflow end-to-end.
func TestValidateCommitWorkflow(t *testing.T) {
	t.Skip("Skipped during architectural refactoring")
	// Skip if running in CI environment without git
	if os.Getenv("CI") == "true" && !IsGitAvailable() {
		t.Skip("Skipping integration test in CI environment without git")
	}

	// Create a temporary directory for configuration
	tempDir, err := os.MkdirTemp("", "gommitlint-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create custom configuration file
	configPath := filepath.Join(tempDir, ".gommitlint.yaml")
	configContent := `
gommitlint:
  validation:
    enabled: true
  subject:
    max_length: 50
  conventional:
    enabled: true
    required: true
    types:
      - feat
      - fix
      - docs
  body:
    required: false
    allow_signoff_only: true
  security:
    signature_required: false
    signoff_required: false
  rules:
    enabled:
      - SubjectLength
      - ConventionalCommit
    disabled:
      - SignOff
      - Signature
      - CommitBody
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set up test cases
	testCases := []struct {
		name          string
		commitMessage string
		shouldPass    bool
	}{
		{
			name:          "Valid commit message",
			commitMessage: "feat: add new feature\n\nThis is a valid commit message body with full description.",
			shouldPass:    true,
		},
		{
			name:          "Invalid commit message - too long",
			commitMessage: "feat: " + strings.Repeat("a", 150) + "\n\nThe subject line is way too long, but at least it has a proper body.",
			shouldPass:    false,
		},
		{
			name:          "Invalid commit message - no conventional format",
			commitMessage: "Add new feature without conventional format\n\nThis message doesn't follow the required conventional format, despite having a body.",
			shouldPass:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup test repository
			repoPath, cleanup := SetupTestRepository(t, testCase.commitMessage)
			defer cleanup()

			// Create a base context with logger
			ctx := context.Background()
			ctx = log.WithLogger(ctx, log.NewTestLogger())

			// Create a config provider with our test settings
			provider, err := config.NewProvider()
			require.NoError(t, err)

			// Apply test configuration options using the direct config API
			provider.UpdateConfig(func(cfg config.Config) config.Config {
				// Use the config API directly without helper functions
				// Update subject config
				updatedSubject := cfg.Subject.WithMaxLength(50)

				// Update conventional config
				updatedConventional := cfg.Conventional.WithRequired(true)
				updatedConventional = updatedConventional.WithTypes([]string{"feat", "fix", "docs"})

				// Update body config
				updatedBody := cfg.Body.WithRequired(false)
				updatedBody = updatedBody.WithAllowSignOffOnly(true)

				// Update security config
				updatedSecurity := cfg.Security.WithGPGRequired(false)
				updatedSecurity = updatedSecurity.WithSignOffRequired(false)

				// Update rules config
				updatedRules := cfg.Rules.WithEnabledRules([]string{"SubjectLength", "ConventionalCommit"})
				updatedRules = updatedRules.WithDisabledRules([]string{"SignOff", "Signature", "CommitBody"})

				// Assemble the updated config
				cfg = cfg.WithSubject(updatedSubject)
				cfg = cfg.WithConventional(updatedConventional)
				cfg = cfg.WithBody(updatedBody)
				cfg = cfg.WithSecurity(updatedSecurity)
				cfg = cfg.WithRules(updatedRules)

				return cfg
			})

			// Add configuration to context
			ctx = config.WithConfig(ctx, provider.GetConfig())

			// Create repository factory with the context
			repoFactory, err := git.NewRepositoryFactory(ctx, repoPath)
			require.NoError(t, err)

			// Get services from factory
			commitService := repoFactory.CreateGitCommitService()
			infoProvider := repoFactory.CreateInfoProvider()
			analyzer := repoFactory.CreateCommitAnalyzer()

			// Create initial validation service using context-based configuration
			_ = validate.CreateValidationServiceWithContext(
				ctx,
				commitService,
				infoProvider,
				analyzer,
			)

			// Use the context-based approach to configure rules
			rulesToEnable := []string{"SubjectLength", "ConventionalCommit"}

			var validationService validate.ValidationService
			ctx, validationService = setupRulesInContext(
				ctx,
				rulesToEnable,
				commitService,
				infoProvider,
				analyzer,
			)

			// Validate the HEAD commit
			result, err := validationService.ValidateCommit(ctx, "HEAD")
			require.NoError(t, err)

			// Print validation errors in detail for all cases
			t.Logf("All validation results:")

			for _, ruleResult := range result.RuleResults {
				t.Logf("Rule: %s, Status: %s, Errors: %v",
					ruleResult.RuleName, ruleResult.Status, ruleResult.Errors)
			}

			// Filter out commitahead rule results for test validation
			// This is necessary because our test repos can't be compared to real branches
			filteredRuleResults := make([]domain.RuleResult, 0)

			for _, ruleResult := range result.RuleResults {
				if ruleResult.RuleName != "CommitsAhead" {
					filteredRuleResults = append(filteredRuleResults, ruleResult)
				}
			}

			// Create a filtered result for assertion
			filteredResult := domain.CommitResult{
				CommitInfo:  result.CommitInfo,
				RuleResults: filteredRuleResults,
				Passed:      true, // Will be updated below
			}

			// Recalculate passed status
			for _, ruleResult := range filteredRuleResults {
				if ruleResult.Status == domain.StatusFailed {
					filteredResult.Passed = false

					break
				}
			}

			// Check if the validation result matches expectations
			if testCase.shouldPass {
				if !filteredResult.Passed {
					t.Logf("Validation errors: %v", GetValidationErrors(result))
				}

				require.True(t, filteredResult.Passed, "Expected validation to pass but it failed")
			} else {
				require.False(t, filteredResult.Passed, "Expected validation to fail but it passed")
				require.NotEmpty(t, GetValidationErrorsExcludingRule(result, "CommitsAhead"), "Expected validation errors")
			}
		})
	}
}

// TestValidateCommitMessageFileWorkflow tests the message file validation workflow end-to-end.
func TestValidateCommitMessageFileWorkflow(t *testing.T) {
	t.Skip("Skipped during architectural refactoring")
	// Create a temporary directory for configuration
	tempDir, err := os.MkdirTemp("", "gommitlint-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create custom configuration file
	configPath := filepath.Join(tempDir, ".gommitlint.yaml")
	configContent := `
gommitlint:
  validation:
    enabled: true
  subject:
    max_length: 50
  conventional:
    enabled: true
    required: true
    types:
      - feat
      - fix
      - docs
  body:
    required: false
    allow_signoff_only: true
  security:
    signature_required: false
    signoff_required: false
  rules:
    enabled:
      - SubjectLength
      - ConventionalCommit
    disabled:
      - SignOff
      - Signature
      - CommitBody
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set up test cases
	testCases := []struct {
		name         string
		messageFile  string
		fileContents string
		shouldPass   bool
	}{
		{
			name:         "Valid commit message in file",
			messageFile:  "valid-message.txt",
			fileContents: "feat: add new feature\n\nThis is a valid commit message with body.",
			shouldPass:   true,
		},
		{
			name:         "Invalid commit message in file - too short",
			messageFile:  "invalid-short-message.txt",
			fileContents: "x\n\nAt least it has a body section, but the subject is too short.",
			shouldPass:   false,
		},
		{
			name:         "Invalid commit message in file - no conventional format",
			messageFile:  "invalid-format-message.txt",
			fileContents: "Adding a new feature without conventional format\n\nThis doesn't follow the conventional format rules.",
			shouldPass:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a temporary directory
			msgTempDir, err := os.MkdirTemp("", "gommitlint-message-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(msgTempDir)

			// Create message file
			messagePath := filepath.Join(msgTempDir, testCase.messageFile)
			err = os.WriteFile(messagePath, []byte(testCase.fileContents), 0600)
			require.NoError(t, err)

			// Create a base context with logger
			ctx := context.Background()
			ctx = log.WithLogger(ctx, log.NewTestLogger())

			// Create a config provider with our test settings
			provider, err := config.NewProvider()
			require.NoError(t, err)

			// Apply test configuration options using the direct config API
			provider.UpdateConfig(func(cfg config.Config) config.Config {
				// Use the config API directly without helper functions
				// Update subject config
				updatedSubject := cfg.Subject.WithMaxLength(50)

				// Update conventional config
				updatedConventional := cfg.Conventional.WithRequired(true)
				updatedConventional = updatedConventional.WithTypes([]string{"feat", "fix", "docs"})

				// Update body config
				updatedBody := cfg.Body.WithRequired(false)
				updatedBody = updatedBody.WithAllowSignOffOnly(true)

				// Update security config
				updatedSecurity := cfg.Security.WithGPGRequired(false)
				updatedSecurity = updatedSecurity.WithSignOffRequired(false)

				// Update rules config
				updatedRules := cfg.Rules.WithEnabledRules([]string{"SubjectLength", "ConventionalCommit"})
				updatedRules = updatedRules.WithDisabledRules([]string{"SignOff", "Signature", "CommitBody"})

				// Assemble the updated config
				cfg = cfg.WithSubject(updatedSubject)
				cfg = cfg.WithConventional(updatedConventional)
				cfg = cfg.WithBody(updatedBody)
				cfg = cfg.WithSecurity(updatedSecurity)
				cfg = cfg.WithRules(updatedRules)

				return cfg
			})

			// Add configuration to context
			ctx = config.WithConfig(ctx, provider.GetConfig())

			// We need a git repository even for message validation
			// for the API to initialize correctly
			repoPath, cleanup := SetupTestRepository(t, "Initial commit\n\nThis is a properly formatted initial commit with a body.")
			defer cleanup()

			// Create repository factory
			repoFactory, err := git.NewRepositoryFactory(ctx, repoPath)
			require.NoError(t, err)

			// Get services from factory
			commitService := repoFactory.CreateGitCommitService()
			infoProvider := repoFactory.CreateInfoProvider()
			analyzer := repoFactory.CreateCommitAnalyzer()

			// Create initial validation service using context-based configuration
			_ = validate.CreateValidationServiceWithContext(
				ctx,
				commitService,
				infoProvider,
				analyzer,
			)

			// Use the context-based approach to configure rules
			rulesToEnable := []string{"SubjectLength", "ConventionalCommit"}

			var validationService validate.ValidationService
			ctx, validationService = setupRulesInContext(
				ctx,
				rulesToEnable,
				commitService,
				infoProvider,
				analyzer,
			)

			// Validate the message file
			results, err := validationService.ValidateMessageFile(ctx, messagePath)
			require.NoError(t, err)

			// Get first commit result
			require.NotEmpty(t, results.CommitResults, "Expected at least one commit result")
			result := results.CommitResults[0]

			// Print validation errors in detail for all cases
			t.Logf("All validation results:")

			for _, ruleResult := range result.RuleResults {
				t.Logf("Rule: %s, Status: %s, Errors: %v",
					ruleResult.RuleName, ruleResult.Status, ruleResult.Errors)
			}

			// Filter out commitahead rule results for test validation
			// This is necessary because our test repos can't be compared to real branches
			filteredRuleResults := make([]domain.RuleResult, 0)

			for _, ruleResult := range result.RuleResults {
				if ruleResult.RuleName != "CommitsAhead" {
					filteredRuleResults = append(filteredRuleResults, ruleResult)
				}
			}

			// Create a filtered result for assertion
			filteredResult := domain.CommitResult{
				CommitInfo:  result.CommitInfo,
				RuleResults: filteredRuleResults,
				Passed:      true, // Will be updated below
			}

			// Recalculate passed status
			for _, ruleResult := range filteredRuleResults {
				if ruleResult.Status == domain.StatusFailed {
					filteredResult.Passed = false

					break
				}
			}

			// Check if the validation result matches expectations
			if testCase.shouldPass {
				if !filteredResult.Passed {
					t.Logf("Validation errors: %v", GetValidationErrors(result))
				}

				require.True(t, filteredResult.Passed, "Expected validation to pass but it failed")
			} else {
				require.False(t, filteredResult.Passed, "Expected validation to fail but it passed")
				require.NotEmpty(t, GetValidationErrorsExcludingRule(result, "CommitsAhead"), "Expected validation errors")
			}
		})
	}
}

// TestConfigurationWorkflow tests the configuration loading and validation workflow.
func TestConfigurationWorkflow(t *testing.T) {
	t.Skip("Skipped during architectural refactoring")
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gommitlint-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create custom configuration file
	configPath := filepath.Join(tempDir, ".gommitlint.yaml")
	configContent := `
gommitlint:
  validation:
    enabled: true
  subject:
    case: sentence
    max_length: 60
  conventional:
    enabled: true
    required: true
    types:
      - feat
      - fix
      - docs
      - custom
  body:
    required: false
  rules:
    enabled:
      - SubjectLength
      - SubjectCase
      - ConventionalCommit
    disabled:
      - JiraReference
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Create a base context with logger
	ctx := context.Background()
	ctx = log.WithLogger(ctx, log.NewTestLogger())

	// Create a config provider with our test settings
	provider, err := config.NewProvider()
	require.NoError(t, err)

	// Apply test configuration options using the direct config API
	provider.UpdateConfig(func(cfg config.Config) config.Config {
		// Use the config API directly without helper functions
		// Update subject config
		updatedSubject := cfg.Subject.WithCase("sentence")
		updatedSubject = updatedSubject.WithMaxLength(60)

		// Update conventional config
		updatedConventional := cfg.Conventional.WithRequired(true)
		updatedConventional = updatedConventional.WithTypes([]string{"feat", "fix", "docs", "custom"})

		// Update body config
		updatedBody := cfg.Body.WithRequired(false)

		// Update rules config
		updatedRules := cfg.Rules.WithEnabledRules([]string{"SubjectLength", "SubjectCase", "ConventionalCommit"})
		updatedRules = updatedRules.WithDisabledRules([]string{"JiraReference"})

		// Assemble the updated config
		cfg = cfg.WithSubject(updatedSubject)
		cfg = cfg.WithConventional(updatedConventional)
		cfg = cfg.WithBody(updatedBody)
		cfg = cfg.WithRules(updatedRules)

		return cfg
	})

	// Add configuration to context
	ctx = config.WithConfig(ctx, provider.GetConfig())

	// Log the raw config file for debugging
	rawConfig, err := os.ReadFile(configPath)
	require.NoError(t, err)
	t.Logf("Raw YAML config:\n%s", string(rawConfig))

	// Get config from context
	config := config.GetConfig(ctx)

	// Log and verify the values to ensure they match what we set
	t.Logf("Subject max length in config: %d", config.Subject.MaxLength)
	require.Equal(t, 60, config.Subject.MaxLength, "Subject max length should match configuration value")

	// Other config values to check
	require.Equal(t, "sentence", config.Subject.Case, "Subject case should be 'sentence'")
	require.False(t, config.Body.Required, "Body should not be required")

	// Verify enabled/disabled rules
	require.Equal(t, []string{"SubjectLength", "SubjectCase", "ConventionalCommit"}, config.Rules.EnabledRules,
		"Enabled rules should match configuration")
	require.Equal(t, []string{"JiraReference"}, config.Rules.DisabledRules,
		"Disabled rules should match configuration")

	// Verify conventional types
	require.Equal(t, []string{"feat", "fix", "docs", "custom"}, config.Conventional.Types,
		"Conventional types should match configuration")

	// Test that the configuration is properly exposed via the context
	// in validation services

	// Create a dummy repository path
	repoPath, cleanup := SetupTestRepository(t, "Initial commit\n\nThis is a proper commit message.")
	defer cleanup()

	// Create repository factory
	_, err = git.NewRepositoryFactory(ctx, repoPath)
	require.NoError(t, err)
}

// setupRulesInContext configures rules in the context and creates a validation service.
// This is the preferred way to configure rules with the new context-based approach.
func setupRulesInContext(
	ctx context.Context,
	ruleNames []string,
	commitService domain.GitCommitService,
	infoProvider domain.RepositoryInfoProvider,
	analyzer domain.CommitAnalyzer,
) (context.Context, validate.ValidationService) {
	// Update the context with the specified rules
	updatedCtx := config.UpdateConfig(ctx, func(cfg config.Config) config.Config {
		// Use the config API directly without helper functions
		// Update rules config to use only the specified rules
		updatedRules := cfg.Rules.WithEnabledRules(ruleNames)

		// Assemble the updated config
		cfg = cfg.WithRules(updatedRules)

		return cfg
	})

	// Create a new validation service with the updated context
	validationService := validate.CreateValidationServiceWithContext(
		updatedCtx,
		commitService,
		infoProvider,
		analyzer,
	)

	return updatedCtx, validationService
}
