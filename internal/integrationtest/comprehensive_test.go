// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	infra "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/log"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/composition"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/itiquette/gommitlint/internal/testutils/integrationtest"
	testlogger "github.com/itiquette/gommitlint/internal/testutils/logger"
)

// TestSimpleValidation tests basic validation rules that are expected to pass.
func TestSimpleValidation(t *testing.T) {
	// Create a temporary directory for configuration
	tempDir, err := os.MkdirTemp("", "gommitlint-simple-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create custom configuration file
	configPath := filepath.Join(tempDir, ".gommitlint.yaml")
	configContent := `
gommitlint:
  message:
    subject:
      max_length: 50
    body:
      required: false
  conventional:
    required: true
    types:
      - feat
      - fix
      - docs
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
	repoPath, cleanup := integrationtest.SetupTestRepository(t, commitMessage)
	defer cleanup()

	// Create a base context with logger
	ctx := testcontext.CreateTestContext()
	ctx = log.WithLogger(ctx, testlogger.NewTestLogger())

	// Create a config service with our test settings
	configService, err := infra.NewService()
	require.NoError(t, err)

	// Get current config and update it
	cfg := configService.GetConfig()

	// Update subject config
	cfg.Message.Subject.MaxLength = 50

	// Update conventional config
	cfg.Conventional.Types = []string{"feat", "fix", "docs"}

	// Update body config
	// Note: body.required is now handled via rules

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

	// Update the service with the modified config
	configService.UpdateConfig(func(types.Config) types.Config {
		return cfg
	})

	// Add configuration to context with proper wrapper
	adapter := configService.GetAdapter()
	ctx = contextx.WithConfig(ctx, adapter)

	// Create validation service using composition root
	logger := log.Logger(ctx)
	loggerAdapter := log.NewAdapter(*logger)
	container := composition.NewContainer(loggerAdapter, configService.GetAdapter().GetConfig())
	service, err := container.CreateValidationService(ctx, repoPath)
	require.NoError(t, err)

	validationService := service

	// Validate the HEAD commit
	result, err := validationService.ValidateCommit(ctx, "HEAD", false)
	require.NoError(t, err)

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

	// Assert the validation should pass
	require.True(t, filteredResult.Passed, "Expected validation to pass")
}
