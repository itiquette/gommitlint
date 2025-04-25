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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/infrastructure/git"
)

// TestSimpleValidation tests basic validation rules that are expected to pass.
func TestSimpleValidation(t *testing.T) {
	// Skip if running in CI environment without git
	if os.Getenv("CI") == "true" && !isGitAvailable() {
		t.Skip("Skipping integration test in CI environment without git")
	}

	// Create a temporary directory for configuration
	tempDir, err := os.MkdirTemp("", "gommitlint-comprehensive-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create simplified test cases that should pass
	testCases := []struct {
		name          string
		commitMessage string
		configContent string
		ruleToCheck   string
	}{
		// Basic validation tests that should pass
		{
			name:          "Subject length - valid",
			commitMessage: "feat: add new feature",
			configContent: `validation: { subject: { max_length: 100 } }`,
			ruleToCheck:   "SubjectLength",
		},
		{
			name:          "Conventional format - valid",
			commitMessage: "feat: add new feature",
			configContent: `validation: { conventional: { enabled: true } }`,
			ruleToCheck:   "ConventionalCommit",
		},
		{
			name:          "Subject case - valid lower case",
			commitMessage: "feat: add new feature with lower case",
			configContent: `validation: { subject: { case: lower } }`,
			ruleToCheck:   "SubjectCase",
		},
		{
			name:          "Commit body - valid when optional",
			commitMessage: "feat: add new feature",
			configContent: `validation: { body: { required: false } }`,
			ruleToCheck:   "CommitBody",
		},
		{
			name:          "Imperative mood - valid",
			commitMessage: "feat: add new feature",
			configContent: `validation: { imperative: { enabled: true } }`,
			ruleToCheck:   "ImperativeVerb",
		},
		{
			name:          "Subject suffix - valid",
			commitMessage: "feat: add new feature",
			configContent: `validation: { subject: { disallow_suffixes: [.] } }`,
			ruleToCheck:   "SubjectSuffix",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config file
			configPath := filepath.Join(tempDir, "config.yaml")
			err = os.WriteFile(configPath, []byte(testCase.configContent), 0600)
			require.NoError(t, err)

			// Set environment variable to use this config
			oldConfigPath := os.Getenv("GOMMITLINT_CONFIG")
			os.Setenv("GOMMITLINT_CONFIG", configPath)
			defer os.Setenv("GOMMITLINT_CONFIG", oldConfigPath)

			// Setup test repository
			repoPath, cleanup := setupTestRepository(t, testCase.commitMessage)
			defer cleanup()

			// Load configuration
			configManager, err := config.New()
			require.NoError(t, err)

			// Create repository factory
			repoFactory, err := git.NewRepositoryFactory(repoPath)
			require.NoError(t, err)

			// Get services from factory
			commitService := repoFactory.CreateGitCommitService()
			infoProvider := repoFactory.CreateInfoProvider()
			analyzer := repoFactory.CreateCommitAnalyzer()

			// Create validation service with dependencies
			validationService := validate.CreateValidationServiceWithDependencies(
				configManager.GetValidationConfig(),
				commitService,
				infoProvider,
				analyzer,
			)

			// Enable only the rule we want to test
			err = validationService.SetActiveRules([]string{testCase.ruleToCheck})
			require.NoError(t, err)

			// Create context
			ctx := context.Background()

			// Validate the HEAD commit
			result, err := validationService.ValidateCommit(ctx, "HEAD")
			require.NoError(t, err)

			// All test cases should pass
			require.True(t, result.Passed, "Expected validation to pass but it failed for rule %s: %v",
				testCase.ruleToCheck, getValidationErrors(result))
		})
	}
}

// TestRuleActivation tests enabling and disabling rules.
func TestRuleActivation(t *testing.T) {
	// Skip if running in CI environment without git
	if os.Getenv("CI") == "true" && !isGitAvailable() {
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
  rules: 
    enabled: []
    disabled: []
  security:
    signature_required: false
    signoff_required: false
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Setup test repository
	repoPath, cleanup := setupTestRepository(t, "feat: initial commit")
	defer cleanup()

	// Create a config manager that only loads our test config
	configManager := createTestConfigManager(t, configPath)

	// Create repository factory
	repoFactory, err := git.NewRepositoryFactory(repoPath)
	require.NoError(t, err)

	// Get services from factory
	commitService := repoFactory.CreateGitCommitService()
	infoProvider := repoFactory.CreateInfoProvider()
	analyzer := repoFactory.CreateCommitAnalyzer()

	// Create validation service with dependencies
	validationService := validate.CreateValidationServiceWithDependencies(
		configManager.GetValidationConfig(),
		commitService,
		infoProvider,
		analyzer,
	)

	// First disable all existing rules
	allRules := validationService.GetAvailableRuleNames()
	t.Logf("Available rules: %v", allRules)
	err = validationService.DisableRules(allRules)
	require.NoError(t, err)

	// Verify no active rules
	activeRules := validationService.GetActiveRules()
	t.Logf("Active rules: %v", activeRules)
	require.Empty(t, activeRules, "Should have no active rules after disabling all")

	// Instead of a custom rule, we'll use an existing rule
	// Set just one existing rule as active
	standardRuleName := "SubjectLength"
	err = validationService.SetActiveRules([]string{standardRuleName})
	require.NoError(t, err)

	// Re-check active rules
	activeRules = validationService.GetActiveRules()
	t.Logf("Active rules after setting %s: %v", standardRuleName, activeRules)
	require.Contains(t, activeRules, standardRuleName, "Standard rule should be active")

	// Create context
	ctx := context.Background()

	// Validate with standard rule
	result, err := validationService.ValidateCommit(ctx, "HEAD")
	require.NoError(t, err)

	// Log all results for debugging
	t.Logf("All validation results:")

	for _, ruleResult := range result.RuleResults {
		t.Logf("Rule: %s, Status: %s, Errors: %v",
			ruleResult.RuleName, ruleResult.Status, ruleResult.Errors)
	}

	// Verify that the commit passes the SubjectLength validation
	require.True(t, result.Passed, "Test commit should pass standard rule validation")
}

// This unused test rule has been removed
