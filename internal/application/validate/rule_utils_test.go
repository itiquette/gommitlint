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
		rules.NewCommitBodyRule(),    // Disabled by default in code
	}

	t.Run("No config - returns all rules except disabled-by-default", func(t *testing.T) {
		// When no explicit enabled or disabled rules are provided
		// Should return all rules except those disabled by default
		activeRules := FilterRules(allRules, []string{}, []string{})

		// Check that we have the expected number of rules
		// JiraReference and CommitBody should be excluded by default
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
		require.False(t, found["CommitBody"], "CommitBody should NOT be in active rules by default")
	})

	t.Run("Explicitly enabled rules only", func(t *testing.T) {
		// Only Rules 1 and 3 are enabled
		enabledRules := []string{"Rule1", "Rule3"}
		activeRules := FilterRules(allRules, enabledRules, []string{})

		// Should return exactly the explicitly enabled rules
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
		require.False(t, names["CommitBody"], "CommitBody should not be enabled")
	})

	t.Run("Explicitly disabled rules only", func(t *testing.T) {
		// Only Rule2 and JiraReference are disabled
		disabledRules := []string{"Rule2", "JiraReference"}
		activeRules := FilterRules(allRules, []string{}, disabledRules)

		// Should return all default-enabled rules except the explicitly disabled ones
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
		require.False(t, names["CommitBody"], "CommitBody should be disabled by default")
	})

	t.Run("Enable default-disabled rule (JiraReference)", func(t *testing.T) {
		// Explicitly enable JiraReference
		enabledRules := []string{"JiraReference"}
		activeRules := FilterRules(allRules, enabledRules, []string{})

		// Should return JiraReference and all default-enabled rules
		require.Equal(t, 4, len(activeRules), "Should return JiraReference and all default-enabled rules")

		// Check that the right rules are enabled
		names := make(map[string]bool)
		for _, rule := range activeRules {
			names[rule.Name()] = true
		}

		require.True(t, names["Rule1"], "Rule1 should still be enabled")
		require.True(t, names["Rule2"], "Rule2 should still be enabled")
		require.True(t, names["Rule3"], "Rule3 should still be enabled")
		require.True(t, names["JiraReference"], "JiraReference should be enabled explicitly")
		require.False(t, names["CommitBody"], "CommitBody should still be disabled by default")
	})

	t.Run("Enable both default-disabled rules", func(t *testing.T) {
		// Explicitly enable both JiraReference and CommitBody
		enabledRules := []string{"JiraReference", "CommitBody"}
		activeRules := FilterRules(allRules, enabledRules, []string{})

		// Should return JiraReference, CommitBody, and all default-enabled rules
		require.Equal(t, 5, len(activeRules), "Should return all rules when default-disabled are explicitly enabled")

		// Check that the right rules are enabled
		names := make(map[string]bool)
		for _, rule := range activeRules {
			names[rule.Name()] = true
		}

		require.True(t, names["Rule1"], "Rule1 should be enabled")
		require.True(t, names["Rule2"], "Rule2 should be enabled")
		require.True(t, names["Rule3"], "Rule3 should be enabled")
		require.True(t, names["JiraReference"], "JiraReference should be explicitly enabled")
		require.True(t, names["CommitBody"], "CommitBody should be explicitly enabled")
	})

	t.Run("Enabled rules take precedence over disabled", func(t *testing.T) {
		// Rule1 and JiraReference are in both enabled and disabled lists
		enabledRules := []string{"Rule1", "JiraReference"}
		disabledRules := []string{"Rule1", "Rule2", "JiraReference"}
		activeRules := FilterRules(allRules, enabledRules, disabledRules)

		// Rule1 and JiraReference should be enabled because they're in enabled list (takes priority)
		// Rule2 should be disabled because it's in disabled list but not in enabled list
		require.Len(t, activeRules, 3, "Should return 3 rules")

		names := make(map[string]bool)
		for _, rule := range activeRules {
			names[rule.Name()] = true
		}

		require.True(t, names["Rule1"], "Rule1 should be enabled despite being in disabled list")
		require.False(t, names["Rule2"], "Rule2 should be disabled")
		require.True(t, names["Rule3"], "Rule3 should be enabled by default")
		require.True(t, names["JiraReference"], "JiraReference should be enabled despite being in disabled list")
		require.False(t, names["CommitBody"], "CommitBody should be disabled by default")
	})

	t.Run("Empty rule list returns basic rules", func(t *testing.T) {
		// When no rules are provided, should return basic set
		activeRules := FilterRules([]domain.Rule{}, []string{}, []string{})

		// Should return basic rules
		require.NotEmpty(t, activeRules, "Should return basic rules")

		// Default-disabled rules should not be in the basic set by default
		containsJira := false
		containsCommitBody := false

		for _, rule := range activeRules {
			if rule.Name() == "JiraReference" {
				containsJira = true
			}

			if rule.Name() == "CommitBody" {
				containsCommitBody = true
			}
		}

		require.False(t, containsJira, "JiraReference should not be in basic rules by default")
		require.False(t, containsCommitBody, "CommitBody should not be in basic rules by default")
	})

	t.Run("Empty rule list with explicitly enabled default-disabled rules", func(t *testing.T) {
		// When no rules are provided but default-disabled rules are explicitly enabled
		activeRules := FilterRules([]domain.Rule{}, []string{"JiraReference", "CommitBody"}, []string{})

		// Should create and include the explicitly enabled rules
		containsJira := false
		containsCommitBody := false

		for _, rule := range activeRules {
			if rule.Name() == "JiraReference" {
				containsJira = true
			}

			if rule.Name() == "CommitBody" {
				containsCommitBody = true
			}
		}

		require.True(t, containsJira, "JiraReference should be in rules when explicitly enabled")
		require.True(t, containsCommitBody, "CommitBody should be in rules when explicitly enabled")
	})

	t.Run("Rules with commented names", func(t *testing.T) {
		// Rules with commented names in YAML should be ignored
		enabledRules := []string{"Rule1", "#JiraReference", "CommitBody"}
		disabledRules := []string{"#Rule2", "Rule3"}
		activeRules := FilterRules(allRules, enabledRules, disabledRules)

		// Rule1 and CommitBody should be enabled
		// JiraReference should be disabled (comment ignores it)
		// Rule2 should be enabled (comment in disabled list is ignored)
		// Rule3 should be disabled
		names := make(map[string]bool)
		for _, rule := range activeRules {
			names[rule.Name()] = true
		}

		require.True(t, names["Rule1"], "Rule1 should be enabled")
		require.True(t, names["Rule2"], "Rule2 should be enabled (comment in disabled list)")
		require.False(t, names["Rule3"], "Rule3 should be disabled")
		require.False(t, names["JiraReference"], "JiraReference should be disabled (commented in enabled list)")
		require.True(t, names["CommitBody"], "CommitBody should be enabled")
	})
}

func TestCleanRuleNames(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Clean normal rule names",
			input:    []string{"SubjectLength", "ConventionalCommit", "JiraReference"},
			expected: []string{"SubjectLength", "ConventionalCommit", "JiraReference"},
		},
		{
			name:     "Remove quotes",
			input:    []string{"\"SubjectLength\"", "'ConventionalCommit'", "JiraReference"},
			expected: []string{"SubjectLength", "ConventionalCommit", "JiraReference"},
		},
		{
			name:     "Remove whitespace",
			input:    []string{" SubjectLength ", "  ConventionalCommit  ", "JiraReference"},
			expected: []string{"SubjectLength", "ConventionalCommit", "JiraReference"},
		},
		{
			name:     "Skip commented lines",
			input:    []string{"SubjectLength", "#ConventionalCommit", "JiraReference"},
			expected: []string{"SubjectLength", "JiraReference"},
		},
		{
			name:     "Handle complex cases",
			input:    []string{"\"SubjectLength\"", "#'ConventionalCommit'", " JiraReference ", "#  CommitBody  "},
			expected: []string{"SubjectLength", "JiraReference"},
		},
		{
			name:     "Handle empty list",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := cleanRuleNames(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}
