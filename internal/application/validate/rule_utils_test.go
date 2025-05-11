// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validate

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

type mockRule struct {
	name string
}

func (r mockRule) Name() string {
	return r.name
}

func (r mockRule) Validate(_ context.Context, _ domain.CommitInfo) []errors.ValidationError {
	return nil
}

func (r mockRule) Help(_ []errors.ValidationError) string {
	return ""
}

func (r mockRule) Result(_ []errors.ValidationError) string {
	return ""
}

func (r mockRule) VerboseResult(_ []errors.ValidationError) string {
	return ""
}

func (r mockRule) Errors() []errors.ValidationError {
	return nil
}

func newMockRule(name string) domain.Rule {
	return &mockRule{name: name}
}

func TestFilterRules(t *testing.T) {
	// Create a set of test rules
	allRules := []domain.Rule{
		newMockRule("Rule1"),         // Pretend this is enabled by default
		newMockRule("Rule2"),         // Pretend this is enabled by default
		newMockRule("Rule3"),         // Pretend this is enabled by default
		rules.NewJiraReferenceRule(), // Disabled by default in code
	}

	t.Run("No config - returns all rules except disabled-by-default", func(t *testing.T) {
		// When no explicit enabled or disabled rules are provided
		// Should return all rules except those disabled by default
		activeRules := FilterRules(allRules, []string{}, []string{})

		// Check that we have the expected number of rules
		// JiraReference should be excluded by default
		require.Len(t, activeRules, 3, "Should return rules except disabled-by-default ones")

		// Check that the right rules are present
		found := make(map[string]bool)
		for _, rule := range activeRules {
			found[rule.Name()] = true
		}

		require.True(t, found["Rule1"], "Rule1 should be in active rules")
		require.True(t, found["Rule2"], "Rule2 should be in active rules")
		require.True(t, found["Rule3"], "Rule3 should be in active rules")
		require.False(t, found["JiraReference"], "JiraReference should NOT be in active rules by default")
	})

	t.Run("Explicitly enabled rules only", func(t *testing.T) {
		// Only Rules 1 and 3 are enabled
		enabledRules := []string{"Rule1", "Rule3"}
		activeRules := FilterRules(allRules, enabledRules, []string{})

		// Should return exactly 2 rules
		require.Len(t, activeRules, 2, "Should return only the explicitly enabled rules")

		// Check that the right rules are enabled
		names := make(map[string]bool)
		for _, rule := range activeRules {
			names[rule.Name()] = true
		}

		require.True(t, names["Rule1"], "Rule1 should be enabled")
		require.False(t, names["Rule2"], "Rule2 should not be enabled")
		require.True(t, names["Rule3"], "Rule3 should be enabled")
		require.False(t, names["JiraReference"], "JiraReference should not be enabled")
	})

	t.Run("Explicitly disabled rules only", func(t *testing.T) {
		// Only Rule2 and JiraReference are disabled
		disabledRules := []string{"Rule2", "JiraReference"}
		activeRules := FilterRules(allRules, []string{}, disabledRules)

		// Should return 2 rules (all except Rule2 and JiraReference)
		require.Len(t, activeRules, 2, "Should return all rules except the disabled ones")

		// Check that the right rules are enabled
		names := make(map[string]bool)
		for _, rule := range activeRules {
			names[rule.Name()] = true
		}

		require.True(t, names["Rule1"], "Rule1 should be enabled")
		require.False(t, names["Rule2"], "Rule2 should be disabled")
		require.True(t, names["Rule3"], "Rule3 should be enabled")
		require.False(t, names["JiraReference"], "JiraReference should be disabled")
	})

	t.Run("Enable default-disabled rule", func(t *testing.T) {
		// Explicitly enable JiraReference
		enabledRules := []string{"JiraReference"}
		activeRules := FilterRules(allRules, enabledRules, []string{})

		// Should return only JiraReference
		require.Len(t, activeRules, 1, "Should return only JiraReference")
		require.Equal(t, "JiraReference", activeRules[0].Name(), "JiraReference should be enabled")
	})

	t.Run("Enabled rules take precedence over disabled", func(t *testing.T) {
		// Rule1 is in both enabled and disabled lists
		enabledRules := []string{"Rule1", "JiraReference"}
		disabledRules := []string{"Rule1", "Rule2"}
		activeRules := FilterRules(allRules, enabledRules, disabledRules)

		// Rule1 should be enabled because it's in enabled list
		// Rule2 should be disabled because it's in disabled list
		// JiraReference should be enabled because it's in enabled list
		require.Len(t, activeRules, 2, "Should return 2 rules")

		names := make(map[string]bool)
		for _, rule := range activeRules {
			names[rule.Name()] = true
		}

		require.True(t, names["Rule1"], "Rule1 should be enabled despite being in disabled list")
		require.False(t, names["Rule2"], "Rule2 should be disabled")
		require.True(t, names["JiraReference"], "JiraReference should be enabled")
	})

	t.Run("Empty rule list returns basic rules", func(t *testing.T) {
		// When no rules are provided, should return basic set
		activeRules := FilterRules([]domain.Rule{}, []string{}, []string{})

		// Should return basic rules
		require.NotEmpty(t, activeRules, "Should return basic rules")

		// JiraReference should not be in the basic set by default
		jiraIncluded := false

		for _, rule := range activeRules {
			if rule.Name() == "JiraReference" {
				jiraIncluded = true

				break
			}
		}

		require.False(t, jiraIncluded, "JiraReference should not be in basic rules by default")

		// But if explicitly enabled, it should be included
		activeRules = FilterRules([]domain.Rule{}, []string{"JiraReference"}, []string{})
		jiraIncluded = false

		for _, rule := range activeRules {
			if rule.Name() == "JiraReference" {
				jiraIncluded = true

				break
			}
		}

		require.True(t, jiraIncluded, "JiraReference should be in basic rules when explicitly enabled")
	})
}
