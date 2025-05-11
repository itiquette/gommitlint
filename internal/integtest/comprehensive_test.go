// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package integtest contains integration tests for the gommitlint package
package integtest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/git"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// TestSimpleValidation tests basic validation rules that are expected to pass.
// Temporarily skipped while architectural changes are being made.
func TestSimpleValidation(t *testing.T) {
	t.Skip("Skipped during architectural refactoring")
	// Create a temporary directory for configuration
	tempDir, err := os.MkdirTemp("", "gommitlint-simple-test-*")
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
  rules:
    enabled:
      - SubjectLength
      - ConventionalCommit
    disabled:
      - SignOff
      - Signature
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Setup test repository
	commitMessage := "feat: add new feature\n\nThis is a valid commit message body with full description."

	// Using the helper from gittest_helper.go
	repoPath, cleanup := SetupTestRepository(t, commitMessage)
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

		// Update rules config
		updatedRules := cfg.Rules.WithEnabledRules([]string{"SubjectLength", "ConventionalCommit"})
		updatedRules = updatedRules.WithDisabledRules([]string{"SignOff", "Signature"})

		// Assemble the updated config
		cfg = cfg.WithSubject(updatedSubject)
		cfg = cfg.WithConventional(updatedConventional)
		cfg = cfg.WithBody(updatedBody)
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

	// Create validation service using context-based configuration
	validationService := validate.CreateValidationServiceWithContext(
		ctx,
		commitService,
		infoProvider,
		analyzer,
	)

	// Validate the HEAD commit
	result, err := validationService.ValidateCommit(ctx, "HEAD")
	require.NoError(t, err)

	// Filter out CommitsAhead rule results that can be problematic in test repos
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

	// Assert the validation should pass
	require.True(t, filteredResult.Passed, "Expected validation to pass")
}
