// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package registry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/application/registry"
	"github.com/itiquette/gommitlint/internal/common/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
)

// MockRule is a simple rule implementation for testing.
type MockRule struct {
	name           string
	alwaysValidate bool
}

func (r MockRule) Name() string {
	return r.name
}

func (r MockRule) Validate(_ context.Context, _ domain.CommitInfo) []errors.ValidationError {
	if r.alwaysValidate {
		return nil
	}

	return []errors.ValidationError{
		errors.NewValidationError("Mock", "mock_error", "Mock error", ""),
	}
}

// MockConfigurableRule implements ConfigurableRule for testing.
type MockConfigurableRule struct {
	MockRule
	configured bool
}

func (r MockConfigurableRule) WithContext(ctx context.Context) domain.Rule {
	// Create a copy
	result := r

	// Get config from context
	cfg := config.GetConfig(ctx)
	if cfg == nil {
		return result
	}

	// Apply configuration - in this mock, we check a specific key
	if cfg.GetBool("mock.configured") {
		result.configured = true
		result.alwaysValidate = true // Make it pass validation when configured
	}

	return result
}

// TestRuleRegistry_GetActiveRules tests the rule filtering logic.
func TestRuleRegistry_GetActiveRules(t *testing.T) {
	// Create test rules
	rule1 := MockRule{name: "Rule1", alwaysValidate: true}
	rule2 := MockRule{name: "Rule2", alwaysValidate: false}
	configRule := MockConfigurableRule{
		MockRule: MockRule{name: "ConfigRule", alwaysValidate: false},
	}

	type testCase struct {
		name          string
		rules         []domain.Rule
		testCtxGetter func() context.Context // Use function instead of storing context directly
		expectedCount int
		expectedNames []string
	}

	tests := []testCase{
		{
			name:          "Nil context returns all rules",
			rules:         []domain.Rule{rule1, rule2, configRule},
			testCtxGetter: func() context.Context { return nil },
			expectedCount: 3,
			expectedNames: []string{"Rule1", "Rule2", "ConfigRule"},
		},
		{
			name:          "Empty context returns all rules",
			rules:         []domain.Rule{rule1, rule2, configRule},
			testCtxGetter: func() context.Context { return context.Background() },
			expectedCount: 3,
			expectedNames: []string{"Rule1", "Rule2", "ConfigRule"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create registry with test rules
			reg := registry.NewRuleRegistry(testCase.rules...)

			// Get active rules
			activeRules := reg.GetActiveRules(testCase.testCtxGetter())

			// Verify count
			require.Len(t, activeRules, testCase.expectedCount, "Wrong number of active rules")

			// Verify rule names
			actualNames := make([]string, len(activeRules))
			for i, rule := range activeRules {
				actualNames[i] = rule.Name()
			}

			// Check that all expected names are present
			for _, expectedName := range testCase.expectedNames {
				found := false

				for _, actualName := range actualNames {
					if actualName == expectedName {
						found = true

						break
					}
				}

				require.True(t, found, "Expected rule %q not found in active rules", expectedName)
			}
		})
	}
}

// TestRuleRegistry_ConfigurableRules tests that configurable rules get configured.
func TestRuleRegistry_ConfigurableRules(t *testing.T) {
	// Create a configurable rule for testing
	configRule := MockConfigurableRule{
		MockRule: MockRule{name: "ConfigRule", alwaysValidate: false},
	}

	// Create a mock config
	mockConfig := &testConfig{
		values: map[string]interface{}{
			"mock.configured": true,
		},
	}

	// Create context with config
	ctx := config.WithConfig(context.Background(), mockConfig)

	// Create registry with test rule
	reg := registry.NewRuleRegistry(configRule)

	// Get active rules
	activeRules := reg.GetActiveRules(ctx)

	// Should still have 1 rule
	require.Len(t, activeRules, 1, "Should have 1 active rule")

	// Validate the rule with an empty commit
	var commit domain.CommitInfo
	errors := activeRules[0].Validate(ctx, commit)

	// The rule should pass validation (no errors) because it was configured
	require.Empty(t, errors, "Configured rule should pass validation")
}

// testConfig is a simple Config implementation for testing.
type testConfig struct {
	values map[string]interface{}
}

func (c *testConfig) Get(key string) interface{} {
	return c.values[key]
}

func (c *testConfig) GetString(key string) string {
	if v, ok := c.values[key].(string); ok {
		return v
	}

	return ""
}

func (c *testConfig) GetBool(key string) bool {
	if v, ok := c.values[key].(bool); ok {
		return v
	}

	return false
}

func (c *testConfig) GetInt(key string) int {
	if v, ok := c.values[key].(int); ok {
		return v
	}

	return 0
}

func (c *testConfig) GetStringSlice(key string) []string {
	if v, ok := c.values[key].([]string); ok {
		return v
	}

	return []string{}
}
