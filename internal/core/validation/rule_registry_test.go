// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// TestNewRuleRegistry tests the creation of a new rule registry.
func TestNewRuleRegistry(t *testing.T) {
	// Create a default configuration
	config := DefaultConfiguration()

	// Create a new registry
	registry := NewRuleRegistry(config)

	// Verify the registry was created
	require.NotNil(t, registry)
	require.NotNil(t, registry.rules)
	require.NotNil(t, registry.activeRules)
	require.NotNil(t, registry.configuration)

	// Check that rules were registered
	require.NotEmpty(t, registry.rules)
	require.NotEmpty(t, registry.activeRules)
}

// TestRuleRegistry_GetRules tests getting all rules from the registry.
func TestRuleRegistry_GetRules(t *testing.T) {
	registry := NewRuleRegistry(DefaultConfiguration())
	rules := registry.GetRules()

	// Check that we got rules
	require.NotEmpty(t, rules)

	// Verify that all important rules are present
	ruleNames := make(map[string]bool)

	// Print actual rule names for debugging
	t.Log("Actual rule names:")

	for _, rule := range rules {
		ruleName := rule.Name()
		t.Logf("- %s", ruleName)

		ruleNames[ruleName] = true
	}

	// Check for required rules - make sure these match the actual rule names
	requiredRules := []string{
		"SubjectLength",
		"ConventionalCommit",
		"ImperativeVerb",
		"JiraReference",
		"Signature",
		"SignOff",
		"Spell",
		"SubjectCase",
		"SubjectSuffix",
		"CommitBody", // Changed from CommitBodyRule
		"CommitsAhead",
	}

	for _, name := range requiredRules {
		require.True(t, ruleNames[name], "Missing required rule: %s", name)
	}
}

// TestRuleRegistry_GetActiveRules tests getting active rules from the registry.
func TestRuleRegistry_GetActiveRules(t *testing.T) {
	registry := NewRuleRegistry(DefaultConfiguration())

	// By default, all rules should be active
	activeRules := registry.GetActiveRules()
	require.Equal(t, len(registry.rules), len(activeRules))

	// Disable a rule
	testRuleName := "SubjectLength"
	registry.DisableRules([]string{testRuleName})

	// Check that the rule is disabled
	activeRules = registry.GetActiveRules()
	require.Len(t, activeRules, len(registry.rules)-1)

	// Check that the disabled rule is not in the active rules
	for _, rule := range activeRules {
		require.NotEqual(t, testRuleName, rule.Name())
	}
}

// TestRuleRegistry_DisableRules tests disabling rules.
func TestRuleRegistry_DisableRules(t *testing.T) {
	registry := NewRuleRegistry(DefaultConfiguration())

	// Disable multiple rules
	disabledRules := []string{"SubjectLength", "ConventionalCommit"}
	registry.DisableRules(disabledRules)

	// Check that the rules are disabled
	for _, name := range disabledRules {
		require.False(t, registry.activeRules[name])
	}

	// Verify rule count
	require.Len(t, registry.GetActiveRules(), len(registry.rules)-len(disabledRules))
}

// TestRuleRegistry_SetActiveRules tests setting active rules.
func TestRuleRegistry_SetActiveRules(t *testing.T) {
	registry := NewRuleRegistry(DefaultConfiguration())

	// Set specific active rules
	activeRules := []string{"SubjectLength", "ConventionalCommit"}
	registry.SetActiveRules(activeRules)

	// Check that only the specified rules are active
	require.Equal(t, len(activeRules), len(registry.GetActiveRules()))

	// Verify each active rule
	for _, rule := range registry.GetActiveRules() {
		require.Contains(t, activeRules, rule.Name())
	}

	// Test with empty list (should enable all rules)
	registry.SetActiveRules([]string{})
	require.Equal(t, len(registry.rules), len(registry.GetActiveRules()))
}

// TestRuleRegistry_GetRuleByName tests getting a rule by name.
func TestRuleRegistry_GetRuleByName(t *testing.T) {
	registry := NewRuleRegistry(DefaultConfiguration())

	// Get a known rule
	rule := registry.GetRuleByName("SubjectLength")
	require.NotNil(t, rule)
	require.Equal(t, "SubjectLength", rule.Name())

	// Try to get a nonexistent rule
	rule = registry.GetRuleByName("NonexistentRule")
	require.Nil(t, rule)
}

// TestRuleRegistry_ApplyConfiguration tests applying configuration.
func TestRuleRegistry_ApplyConfiguration(t *testing.T) {
	// Test with enabled rules
	config := DefaultConfiguration()
	config.EnabledRules = []string{"SubjectLength", "ConventionalCommit"}
	registry := NewRuleRegistry(config)

	// Check that only specified rules are active
	require.Equal(t, len(config.EnabledRules), len(registry.GetActiveRules()))

	// Test with disabled rules
	config = DefaultConfiguration()
	config.DisabledRules = []string{"SubjectLength"}
	registry = NewRuleRegistry(config)

	// Check that all rules except the disabled one are active
	require.Len(t, registry.GetActiveRules(), len(registry.rules)-1)

	// Check specifically that the disabled rule is not active
	for _, rule := range registry.GetActiveRules() {
		require.NotEqual(t, "SubjectLength", rule.Name())
	}
}

// TestRuleRegistry_RuleConfigurationOptions tests that rule configuration options are applied.
func TestRuleRegistry_RuleConfigurationOptions(t *testing.T) {
	// Create custom configuration
	config := DefaultConfiguration()
	config.MaxSubjectLength = 50
	config.RequireSignOff = false
	config.SpellLocale = "UK"

	registry := NewRuleRegistry(config)

	// Test that SubjectLength rule has the custom length
	subjectRule := registry.GetRuleByName("SubjectLength")
	require.NotNil(t, subjectRule)

	// Use reflection to check rule's configuration
	// This is a bit brittle but useful for testing
	validateCommit := func(_ *testing.T, rule domain.Rule, commit *domain.CommitInfo) []errors.ValidationError {
		return rule.Validate(commit)
	}

	// Create test commit with long subject
	longSubject := "a"
	for i := 0; i < 60; i++ {
		longSubject += "a"
	}

	// Should get validation error since subject > 50
	commit := &domain.CommitInfo{Subject: longSubject}
	errors := validateCommit(t, subjectRule, commit)
	require.NotEmpty(t, errors)

	// Test SignOff rule configuration
	signOffRule := registry.GetRuleByName("SignOff")
	require.NotNil(t, signOffRule)

	// Create commit without signoff
	commit = &domain.CommitInfo{
		Subject: "Valid subject",
		Body:    "Valid body without sign-off",
	}

	// Should not get validation error since require signoff is false
	errors = validateCommit(t, signOffRule, commit)
	require.Empty(t, errors)
}
