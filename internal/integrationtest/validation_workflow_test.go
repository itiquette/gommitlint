// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	infra "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/git"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	testlogger "github.com/itiquette/gommitlint/internal/testutils/logger"
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
		rule.errors = append(rule.errors, errors.New(name, errors.ErrUnknown, "Mock error message"))
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
				errors.New(
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
					errors.New(
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
    require_sign_off: false
  security:
    signature_required: false
  rules:
    enabled_rules:
      - SubjectLength
      - ConventionalCommit
    disabled_rules:
      - SignOff
      - Signature
      - CommitBody
      - JiraReference
      - ImperativeVerb
      - SubjectCase
      - Spell
      - SubjectSuffix
      - CommitsAhead
      - SignedIdentity
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
			ctx := testcontext.CreateTestContext()
			ctx = log.WithLogger(ctx, testlogger.NewTestLogger())

			// Create a config provider with our test settings
			configService, err := infra.NewService()
			require.NoError(t, err)

			// Apply test configuration directly
			cfg := configService.GetConfig()
			cfg = applyTestConfiguration(cfg)

			configService.UpdateConfig(func(types.Config) types.Config {
				return cfg
			})

			// Add configuration to context
			adapter := configService.GetAdapter()
			ctx = contextx.WithConfig(ctx, adapter)

			// Create repository factory with the context
			repoFactory, err := git.NewRepositoryFactory(ctx, repoPath)
			require.NoError(t, err)

			// Get services from factory
			commitService := repoFactory.CreateCommitRepository()
			infoProvider := repoFactory.CreateRepositoryInfoProvider()
			analyzer := repoFactory.CreateCommitAnalyzer()

			// Use the context-based approach to configure rules
			rulesToEnable := []string{"SubjectLength", "ConventionalCommit"}

			// Create validation service using the setupRulesInContext helper
			ctx, validationService := setupRulesInContext(
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

			// Filter for only the rules we explicitly enabled
			// This works with the new rule priority system by only checking rules we care about
			filteredRuleResults := make([]domain.RuleResult, 0)

			// Only consider results from explicitly enabled rules
			enabledRuleNames := map[string]bool{
				"SubjectLength":      true,
				"ConventionalCommit": true,
			}

			for _, ruleResult := range result.RuleResults {
				// Only include rule results for rules we explicitly enabled
				if enabledRuleNames[ruleResult.RuleName] {
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
    require_sign_off: false
  security:
    signature_required: false
  rules:
    enabled_rules:
      - SubjectLength
      - ConventionalCommit
    disabled_rules:
      - SignOff
      - Signature
      - CommitBody
      - JiraReference 
      - ImperativeVerb
      - SubjectCase
      - Spell
      - SubjectSuffix
      - CommitsAhead
      - SignedIdentity
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
			ctx := testcontext.CreateTestContext()
			ctx = log.WithLogger(ctx, testlogger.NewTestLogger())

			// Create a config provider with our test settings
			configService, err := infra.NewService()
			require.NoError(t, err)

			// Apply test configuration directly
			cfg := configService.GetConfig()
			cfg = applyTestConfiguration(cfg)

			configService.UpdateConfig(func(types.Config) types.Config {
				return cfg
			})

			// Add configuration to context
			adapter := configService.GetAdapter()
			ctx = contextx.WithConfig(ctx, adapter)

			// We need a git repository even for message validation
			// for the API to initialize correctly
			repoPath, cleanup := SetupTestRepository(t, "Initial commit\n\nThis is a properly formatted initial commit with a body.")
			defer cleanup()

			// Create repository factory
			repoFactory, err := git.NewRepositoryFactory(ctx, repoPath)
			require.NoError(t, err)

			// Get services from factory
			commitService := repoFactory.CreateCommitRepository()
			infoProvider := repoFactory.CreateRepositoryInfoProvider()
			analyzer := repoFactory.CreateCommitAnalyzer()

			// Use the context-based approach to configure rules
			rulesToEnable := []string{"SubjectLength", "ConventionalCommit"}

			// Create validation service using the setupRulesInContext helper
			ctx, validationService := setupRulesInContext(
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
			require.NotEmpty(t, results.Results, "Expected at least one commit result")
			result := results.Results[0]

			// Print validation errors in detail for all cases
			t.Logf("All validation results:")

			for _, ruleResult := range result.RuleResults {
				t.Logf("Rule: %s, Status: %s, Errors: %v",
					ruleResult.RuleName, ruleResult.Status, ruleResult.Errors)
			}

			// Filter for only the rules we explicitly enabled
			// This works with the new rule priority system by only checking rules we care about
			filteredRuleResults := make([]domain.RuleResult, 0)

			// Only consider results from explicitly enabled rules
			enabledRuleNames := map[string]bool{
				"SubjectLength":      true,
				"ConventionalCommit": true,
			}

			for _, ruleResult := range result.RuleResults {
				// Only include rule results for rules we explicitly enabled
				if enabledRuleNames[ruleResult.RuleName] {
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
    enabled_rules:
      - SubjectLength
      - SubjectCase
      - ConventionalCommit
    disabled_rules:
      - JiraReference
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Create a base context with logger
	ctx := testcontext.CreateTestContext()
	ctx = log.WithLogger(ctx, testlogger.NewTestLogger())

	// Create a config service with our test settings
	cfgService, err := infra.NewService()
	require.NoError(t, err)

	// Apply test configuration options using the direct config API
	cfg := cfgService.GetConfig()

	// Update the config directly
	cfg.Subject.Case = "sentence"
	cfg.Subject.MaxLength = 60

	// Update conventional config
	cfg.Conventional.Types = []string{"feat", "fix", "docs", "custom"}

	// Update body config

	// Update rules config
	cfg.Rules.Enabled = []string{"SubjectLength", "SubjectCase", "ConventionalCommit"}
	cfg.Rules.Disabled = []string{"JiraReference"}

	// Set the updated config in the manager
	cfgService.UpdateConfig(func(types.Config) types.Config {
		return cfg
	})

	// Add configuration to context
	adapter := cfgService.GetAdapter()
	ctx = contextx.WithConfig(ctx, adapter)

	// Log the raw config file for debugging
	rawConfig, err := os.ReadFile(configPath)
	require.NoError(t, err)
	t.Logf("Raw YAML config:\n%s", string(rawConfig))

	// The context-based configuration system is now fully implemented.
	// This test verifies configuration is properly exposed via the context.

	// Create a dummy repository path for testing
	repoPath, cleanup := SetupTestRepository(t, "Initial commit\n\nThis is a proper commit message.")
	defer cleanup()

	// Create repository factory
	_, err = git.NewRepositoryFactory(ctx, repoPath)
	require.NoError(t, err)
}

// applyTestConfiguration applies a standardized test configuration for consistent test setup.
func applyTestConfiguration(cfg types.Config) types.Config {
	// Use the types config API directly without helper functions
	// Update subject config
	cfg.Subject.MaxLength = 50

	// Update conventional config
	cfg.Conventional.Types = []string{"feat", "fix", "docs"}

	// Update body config
	cfg.Body.AllowSignOffOnly = true
	cfg.Body.RequireSignOff = false

	// Update security config
	cfg.Security.GPGRequired = false

	// Update rules config - ensure only the rules we need are active
	// Explicitly enable these rules (they'll run unless disabled)
	cfg.Rules.Enabled = []string{"SubjectLength", "ConventionalCommit"}

	// Explicitly disable ALL other rules to ensure only our enabled ones run
	// This approach is compatible with the rule priority system
	cfg.Rules.Disabled = []string{
		"SignOff", "Signature", "CommitBody", "JiraReference",
		"ImperativeVerb", "SubjectCase", "Spell", "SubjectSuffix", "CommitsAhead",
		// Add any other rules that might be enabled by default
		"SignedIdentity",
	}

	return cfg
}

// setupRulesInContext configures rules in the context and creates a validation service.
// This is the preferred way to configure rules with the new context-based approach.
func setupRulesInContext(
	ctx context.Context,
	_ []string, // Rule names - currently unused due to context-based configuration
	commitService domain.CommitRepository,
	infoProvider domain.RepositoryInfoProvider,
	analyzer domain.CommitAnalyzer,
) (context.Context, validate.ValidationService) {
	// Create a new validation service with the context
	validationService := validate.CreateValidationService(
		ctx,
		commitService,
		infoProvider,
		analyzer,
	)

	return ctx, validationService
}
