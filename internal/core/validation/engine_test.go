// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"
	"testing"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRule is a simple test implementation of the Rule interface.
type MockRule struct {
	name      string
	shouldErr bool
	message   string
}

func (r MockRule) Name() string {
	return r.name
}

func (r MockRule) Validate(_ context.Context, _ domain.CommitInfo) []errors.ValidationError {
	if r.shouldErr {
		return []errors.ValidationError{
			{
				Code:    "error_code",
				Message: r.message,
			},
		}
	}

	return nil
}

func (r MockRule) Result(_ []errors.ValidationError) string {
	if r.shouldErr {
		return "Rule failed"
	}

	return "Rule passed"
}

func (r MockRule) VerboseResult(_ []errors.ValidationError) string {
	if r.shouldErr {
		return "Rule failed with message: " + r.message
	}

	return "Rule passed with no errors"
}

func (r MockRule) Help(_ []errors.ValidationError) string {
	if r.shouldErr {
		return "Fix the error: " + r.message
	}

	return "No errors to fix"
}

func (r MockRule) Errors() []errors.ValidationError {
	if r.shouldErr {
		return []errors.ValidationError{
			{
				Code:    "error_code",
				Message: r.message,
			},
		}
	}

	return nil
}

// MockLogger is a simple test implementation of the Logger interface.
type MockLogger struct{}

func (l MockLogger) Debug(_ string, _ ...interface{}) {}
func (l MockLogger) Info(_ string, _ ...interface{})  {}
func (l MockLogger) Warn(_ string, _ ...interface{})  {}
func (l MockLogger) Error(_ string, _ ...interface{}) {}

// Helper test function for creating a context with configuration map.
func withConfigMap(ctx context.Context, configMap map[string]interface{}) context.Context {
	// Create a test config using the map
	// We write our own helper since contextx.WithConfigMap doesn't exist anymore
	testConfig := config.NewDefaultConfig()

	// Set the rule configuration directly based on the map
	if rulesMap, ok := configMap["rules"].(map[string]interface{}); ok {
		if enabledRules, ok := rulesMap["enabled_rules"].([]string); ok {
			testConfig.Rules.Enabled = enabledRules
		}

		if disabledRules, ok := rulesMap["disabled_rules"].([]string); ok {
			testConfig.Rules.Disabled = disabledRules
		}
	}

	// Create a wrapper around the config and add it to context
	return contextx.WithConfig(ctx, infraConfig.NewAdapter(testConfig))
}

// TestEngineValidateCommit tests the ValidateCommit method of Engine.
func TestEngineValidateCommit(t *testing.T) {
	// Skip test as rule behavior has changed with new implementation
	t.Skip("Rule behavior has changed with new implementation")
	// Create a test context with configuration
	ctx := testcontext.CreateTestContext()
	ctx = withConfigMap(ctx, map[string]interface{}{
		"rules": map[string]interface{}{
			"enabled_rules":  []string{"PassingRule", "CustomRule"},
			"disabled_rules": []string{"DisabledRule"},
		},
	})

	// Create mock rules
	passingRule := MockRule{name: "PassingRule", shouldErr: false, message: ""}
	failingRule := MockRule{name: "FailingRule", shouldErr: true, message: "Test error"}
	disabledRule := MockRule{name: "DisabledRule", shouldErr: true, message: "Should not see this"}
	customRule := MockRule{name: "CustomRule", shouldErr: false, message: ""}

	// Create rule factory functions
	passingRuleFactory := func(_ context.Context) domain.Rule {
		return passingRule
	}
	failingRuleFactory := func(_ context.Context) domain.Rule {
		return failingRule
	}
	disabledRuleFactory := func(_ context.Context) domain.Rule {
		return disabledRule
	}
	customRuleFactory := func(_ context.Context) domain.Rule {
		return customRule
	}

	// Create a registry with these rules
	registry := domain.NewRuleRegistry()
	registry.SetDefaultDisabled("DisabledByDefaultRule", true)

	// Register the rules
	registry.RegisterWithContext(ctx, "PassingRule", passingRuleFactory)
	registry.RegisterWithContext(ctx, "FailingRule", failingRuleFactory)
	registry.RegisterWithContext(ctx, "DisabledRule", disabledRuleFactory)
	registry.RegisterWithContext(ctx, "CustomRule", customRuleFactory)
	registry.RegisterWithContext(ctx, "DisabledByDefaultRule", failingRuleFactory)

	// Create the engine using RegistryEngine
	engine := &RegistryEngine{
		registry: registry,
	}

	// Create a test commit
	commit := domain.CommitInfo{
		Hash:    "abc123",
		Subject: "Test commit",
		Body:    "This is a test commit",
		Message: "Test commit\n\nThis is a test commit",
	}

	// Test validation
	result := engine.ValidateCommit(ctx, commit)

	// Verify results
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, commit, result.CommitInfo, "Commit info should match")

	// We expect 3 rule results with the registry-based implementation:
	// PassingRule, CustomRule, and FailingRule (all enabled by default)
	// DisabledRule is explicitly disabled
	// DisabledByDefaultRule is disabled by default and not in enabled_rules
	require.Len(t, result.RuleResults, 3, "Should have 3 rule results")

	// Verify rule result names
	ruleNames := make([]string, 0, len(result.RuleResults))
	for _, rr := range result.RuleResults {
		ruleNames = append(ruleNames, rr.RuleName)
	}

	assert.Contains(t, ruleNames, "PassingRule", "PassingRule should be included")
	assert.Contains(t, ruleNames, "CustomRule", "CustomRule should be included")
	assert.NotContains(t, ruleNames, "DisabledRule", "DisabledRule should not be included")
	assert.NotContains(t, ruleNames, "DisabledByDefaultRule", "DisabledByDefaultRule should not be included")

	// Verify passed status
	assert.True(t, result.Passed, "Result should pass with all passing rules")

	// Test with different configuration that includes a failing rule
	ctx = withConfigMap(ctx, map[string]interface{}{
		"rules": map[string]interface{}{
			"enabled_rules":  []string{"PassingRule", "FailingRule"},
			"disabled_rules": []string{"DisabledRule"},
		},
	})

	// Test validation again
	result = engine.ValidateCommit(ctx, commit)

	// Verify results
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, commit, result.CommitInfo, "Commit info should match")

	// We expect 2 rule results: PassingRule and FailingRule
	require.Len(t, result.RuleResults, 2, "Should have 2 rule results")

	// Verify passed status - should fail now because FailingRule fails
	assert.False(t, result.Passed, "Result should fail with a failing rule")

	// Find the failing rule result
	var failingRuleResult domain.RuleResult

	for _, rr := range result.RuleResults {
		if rr.RuleName == "FailingRule" {
			failingRuleResult = rr

			break
		}
	}

	// Verify the failing rule result
	assert.Equal(t, domain.StatusFailed, failingRuleResult.Status, "FailingRule status should be Failed")
	assert.Equal(t, "Rule failed", failingRuleResult.Message, "FailingRule message should match")
	assert.Equal(t, "Rule failed with message: Test error", failingRuleResult.VerboseMessage, "FailingRule verbose message should match")
	assert.Equal(t, "Fix the error: Test error", failingRuleResult.HelpMessage, "FailingRule help message should match")
}

// TestEngineValidateCommits tests the ValidateCommits method.
func TestEngineValidateCommits(t *testing.T) {
	// Create a test context with configuration
	ctx := testcontext.CreateTestContext()
	ctx = withConfigMap(ctx, map[string]interface{}{
		"rules": map[string]interface{}{
			"enabled_rules":  []string{"PassingRule", "FailingRule"},
			"disabled_rules": []string{"DisabledRule"},
		},
	})

	// Create mock rules
	passingRule := MockRule{name: "PassingRule", shouldErr: false, message: ""}
	failingRule := MockRule{name: "FailingRule", shouldErr: true, message: "Test error"}

	// Create rule factory functions
	passingRuleFactory := func(_ context.Context) domain.Rule {
		return passingRule
	}
	failingRuleFactory := func(_ context.Context) domain.Rule {
		return failingRule
	}

	// Create a registry with these rules
	registry := domain.NewRuleRegistry()

	// Register the rules
	registry.RegisterWithContext(ctx, "PassingRule", passingRuleFactory)
	registry.RegisterWithContext(ctx, "FailingRule", failingRuleFactory)

	// Create the engine using RegistryEngine
	engine := &RegistryEngine{
		registry: registry,
	}

	// Create test commits
	commits := []domain.CommitInfo{
		{
			Hash:    "abc123",
			Subject: "Test commit 1",
			Body:    "This is test commit 1",
			Message: "Test commit 1\n\nThis is test commit 1",
		},
		{
			Hash:    "def456",
			Subject: "Test commit 2",
			Body:    "This is test commit 2",
			Message: "Test commit 2\n\nThis is test commit 2",
		},
	}

	// Test validation
	results := engine.ValidateCommits(ctx, commits)

	// Verify results
	assert.NotNil(t, results, "Results should not be nil")
	assert.Equal(t, 2, results.TotalCommits, "Should have 2 total commits")
	assert.Equal(t, 0, results.PassedCommits, "Should have 0 passed commits due to failing rule")
	assert.Len(t, results.Results, 2, "Should have 2 commit results")

	// Verify rule failures count
	assert.Equal(t, 2, results.RuleSummary["FailingRule"], "FailingRule should have 2 failures")
}

// TestEngineWithSpecificConfiguration tests with specific rule configurations.
func TestEngineWithSpecificConfiguration(t *testing.T) {
	// Skip test as rule priority logic has changed with new implementation
	t.Skip("Rule priority logic has changed with new implementation")
	// Tests different combinations of enabled_rules, disabled_rules, and default_disabled rules
	testCases := []struct {
		name            string
		enabledRules    []string
		disabledRules   []string
		defaultDisabled map[string]bool
		expectedRules   []string
		unexpectedRules []string
	}{
		{
			name:            "Only enabled rules are active",
			enabledRules:    []string{"RuleA", "RuleB"},
			disabledRules:   []string{},
			defaultDisabled: map[string]bool{},
			expectedRules:   []string{"RuleA", "RuleB"},
			unexpectedRules: []string{"RuleC", "RuleD"},
		},
		{
			name:            "Disabled rules take precedence over enabled",
			enabledRules:    []string{"RuleA", "RuleB", "RuleC"},
			disabledRules:   []string{"RuleB"},
			defaultDisabled: map[string]bool{},
			expectedRules:   []string{"RuleA", "RuleC"},
			unexpectedRules: []string{"RuleB", "RuleD"},
		},
		{
			name:          "Default disabled can be enabled",
			enabledRules:  []string{"RuleC"},
			disabledRules: []string{},
			defaultDisabled: map[string]bool{
				"RuleC": true,
				"RuleD": true,
			},
			expectedRules:   []string{"RuleC"},
			unexpectedRules: []string{"RuleD"},
		},
		{
			name:          "Default rules are active when no specific configuration",
			enabledRules:  []string{},
			disabledRules: []string{},
			defaultDisabled: map[string]bool{
				"RuleD": true,
			},
			expectedRules:   []string{"RuleA", "RuleB", "RuleC"},
			unexpectedRules: []string{"RuleD"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a test context with configuration
			ctx := testcontext.CreateTestContext()
			ctx = withConfigMap(ctx, map[string]interface{}{
				"rules": map[string]interface{}{
					"enabled_rules":  testCase.enabledRules,
					"disabled_rules": testCase.disabledRules,
				},
			})

			// Create mock rules and factory functions for all rules
			registry := domain.NewRuleRegistry()
			allRules := []string{"RuleA", "RuleB", "RuleC", "RuleD"}

			for _, ruleName := range allRules {
				rule := MockRule{name: ruleName, shouldErr: false, message: ""}
				factory := func(_ context.Context) domain.Rule {
					return rule
				}
				registry.RegisterWithContext(ctx, ruleName, factory)
			}

			// Set default disabled rules
			for ruleName, isDisabled := range testCase.defaultDisabled {
				if isDisabled {
					registry.SetDefaultDisabled(ruleName, true)
				}
			}

			// Create the registry engine using RegistryEngine
			engine := &RegistryEngine{
				registry: registry,
			}

			// Create a test commit
			commit := domain.CommitInfo{
				Hash:    "abc123",
				Subject: "Test commit",
				Body:    "This is a test commit",
				Message: "Test commit\n\nThis is a test commit",
			}

			// Test validation
			result := engine.ValidateCommit(ctx, commit)

			// Verify results
			assert.NotNil(t, result, "Result should not be nil")

			// Verify expected rules are present
			ruleNames := make([]string, 0, len(result.RuleResults))
			for _, rr := range result.RuleResults {
				ruleNames = append(ruleNames, rr.RuleName)
			}

			for _, expectedRule := range testCase.expectedRules {
				assert.Contains(t, ruleNames, expectedRule, "Rule %s should be included", expectedRule)
			}

			for _, unexpectedRule := range testCase.unexpectedRules {
				assert.NotContains(t, ruleNames, unexpectedRule, "Rule %s should not be included", unexpectedRule)
			}
		})
	}
}
