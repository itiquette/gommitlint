// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validate

import (
	"context"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// isExplicitlyEnabled checks if a rule is explicitly enabled in the enabledRules list.
func isExplicitlyEnabled(enabledRules []string, ruleName string) bool {
	for _, rule := range enabledRules {
		// Clean rule name for comparison
		cleanRule := strings.TrimSpace(strings.Trim(rule, "\"'"))
		if cleanRule == ruleName {
			return true
		}
	}

	return false
}

// validateRuleWithContext helps validate a commit with a rule using context.
// This is extracted to a separate function to handle error handling consistently.
func validateRuleWithContext(ctx context.Context, rule domain.Rule, commit domain.CommitInfo, result *domain.CommitResult) {
	// Skip if rule is nil (shouldn't happen, but safety check)
	if rule == nil {
		return
	}

	// Validate the commit with the rule
	errors := rule.Validate(ctx, commit)

	// Create a rule result
	ruleResult := domain.RuleResult{
		RuleName: rule.Name(),
		Status:   domain.StatusPassed,
		Errors:   errors,
	}

	// Update status if there are errors
	if len(errors) > 0 {
		ruleResult.Status = domain.StatusFailed
		result.Passed = false
	}

	// Add the rule result to the commit result
	result.RuleResults = append(result.RuleResults, ruleResult)
}

// Context rule provider implementation
// =====================================

// ContextRuleProvider is a rule provider that uses context for configuration.
// It's designed to work with the enhanced configuration system.
type ContextRuleProvider struct {
	// Do not store context in struct fields
	configSnapshot  config.Config // snapshot of config from context
	analyzer        domain.CommitAnalyzer
	rules           []domain.Rule
	customFactories map[string]contextRuleFactory
}

// A factory function that creates a rule using context.
type contextRuleFactory func(context.Context) domain.Rule

// NewContextRuleProvider creates a new context-based rule provider.
func NewContextRuleProvider(ctx context.Context, analyzer domain.CommitAnalyzer) *ContextRuleProvider {
	// Create the provider with config from context
	provider := &ContextRuleProvider{
		configSnapshot:  config.GetConfig(ctx),
		analyzer:        analyzer,
		customFactories: make(map[string]contextRuleFactory),
	}

	return provider
}

// GetRules returns all available rules regardless of active status.
func (p *ContextRuleProvider) GetRules(ctx context.Context) []domain.Rule {
	// Log what we're doing
	logger := log.Logger(ctx)
	logger.Debug().Msg("Getting all rules")

	// Create the real rules directly
	standardRules := createContextBasedRules(ctx, p.analyzer)
	logger.Debug().Int("rule_count", len(standardRules)).Msg("Created standard rules")

	// Set the computed rules so they're available for future use
	p.rules = standardRules

	// Return the rules
	return standardRules
}

// GetActiveRules returns only the active rules based on configuration.
func (p *ContextRuleProvider) GetActiveRules(ctx context.Context) []domain.Rule {
	logger := log.Logger(ctx)
	logger.Debug().Msg("Getting active rules")

	// Get all rules - this will build them if needed
	allRules := p.GetRules(ctx)
	logger.Debug().Int("all_rule_count", len(allRules)).Msg("Retrieved all rules")

	// Get active rule configuration from context
	cfg := config.GetConfig(ctx)
	enabledRules := cfg.Rules.EnabledRules
	disabledRules := cfg.Rules.DisabledRules

	// Log rule configuration settings
	logger.Debug().
		Strs("enabled_rules", enabledRules).
		Strs("disabled_rules", disabledRules).
		Msg("Rule configuration settings")

	// Use the FilterRules function to get active rules
	// This handles all rules consistently without special cases
	activeRules := FilterRules(allRules, enabledRules, disabledRules)

	// Log active rules
	logActiveRules(logger, activeRules)

	// Force at least the basic rules to be active if no rules were returned
	if len(activeRules) == 0 {
		logger.Debug().Msg("No active rules found, creating default rules directly")

		// Create basic rules directly
		basicRules := []domain.Rule{
			rules.NewSubjectLengthRule(rules.WithMaxLength(cfg.Subject.MaxLength)),
			rules.NewConventionalCommitRule(),
			rules.NewImperativeVerbRule(rules.WithImperativeConventionalCommit(cfg.Conventional.Required)),
			rules.NewSubjectCaseRule(),
			rules.NewSubjectSuffixRule(),
			rules.NewSignOffRule(),
			rules.NewSignatureRule(),
			rules.NewSpellRule(),
		}

		// No special handling for any specific rule

		activeRules = basicRules

		// Log the forced rule set
		logActiveRules(logger, activeRules)
	}

	return activeRules
}

// WithActiveRules returns a new provider with only the specified rules active.
func (p *ContextRuleProvider) WithActiveRules(ruleNames []string) domain.RuleProvider {
	// Create a copy of the provider
	newProvider := &ContextRuleProvider{
		configSnapshot:  p.configSnapshot,
		analyzer:        p.analyzer,
		customFactories: make(map[string]contextRuleFactory),
	}

	// Copy the custom factories
	for name, factory := range p.customFactories {
		newProvider.customFactories[name] = factory
	}

	// Set the enabled rules in the new config
	newConfig := newProvider.configSnapshot
	newConfig.Rules.EnabledRules = ruleNames
	newConfig.Rules.DisabledRules = []string{} // Clear disabled rules
	newProvider.configSnapshot = newConfig

	return newProvider
}

// WithDisabledRules returns a new provider with the specified rules disabled.
func (p *ContextRuleProvider) WithDisabledRules(ruleNames []string) domain.RuleProvider {
	// Create a copy of the provider
	newProvider := &ContextRuleProvider{
		configSnapshot:  p.configSnapshot,
		analyzer:        p.analyzer,
		customFactories: make(map[string]contextRuleFactory),
	}

	// Copy the custom factories
	for name, factory := range p.customFactories {
		newProvider.customFactories[name] = factory
	}

	// Set the disabled rules in the new config
	newConfig := newProvider.configSnapshot
	newConfig.Rules.DisabledRules = ruleNames
	newProvider.configSnapshot = newConfig

	return newProvider
}

// WithCustomRule returns a new provider with the custom rule added.
func (p *ContextRuleProvider) WithCustomRule(rule domain.Rule) domain.RuleProvider {
	// Create a copy of the provider
	newProvider := &ContextRuleProvider{
		configSnapshot:  p.configSnapshot,
		analyzer:        p.analyzer,
		customFactories: make(map[string]contextRuleFactory),
	}

	// Copy the custom factories
	for name, factory := range p.customFactories {
		newProvider.customFactories[name] = factory
	}

	// Add the custom rule as a factory that always returns the given rule
	ruleName := rule.Name()
	newProvider.customFactories[ruleName] = func(context.Context) domain.Rule {
		return rule
	}

	return newProvider
}

// GetAvailableRuleNames returns the names of all available rules.
// This helps with discovery of supported rules.
func (p *ContextRuleProvider) GetAvailableRuleNames(ctx context.Context) []string {
	// Create all standard rules to get their names using the passed context
	standardRules := createContextBasedRules(ctx, p.analyzer)

	// Create a map to avoid duplicates
	availableRules := make(map[string]bool)

	// Add all standard rule names
	for _, rule := range standardRules {
		availableRules[rule.Name()] = true
	}

	// Add all custom rule factory names
	for name := range p.customFactories {
		availableRules[name] = true
	}

	// Convert to slice
	result := make([]string, 0, len(availableRules))
	for name := range availableRules {
		result = append(result, name)
	}

	return result
}

// BuildRules is a public method to force rebuilding rules.
// This can be used to refresh rules when configuration changes.
func (p *ContextRuleProvider) BuildRules(ctx context.Context) {
	p.buildRules(ctx)
}

// buildRules builds the rule set from the context configuration.
// This is called lazily when rules are first requested.
func (p *ContextRuleProvider) buildRules(ctx context.Context) {
	logger := log.Logger(ctx)
	logger.Debug().Msg("Building context-based rules")

	// Create the standard set of context-based rules
	standardRules := createContextBasedRules(ctx, p.analyzer)

	// Get active rule configuration from context
	enabledRules := p.configSnapshot.Rules.EnabledRules
	disabledRules := p.configSnapshot.Rules.DisabledRules

	// Filter rules based on configuration
	rules := FilterRules(standardRules, enabledRules, disabledRules)

	// Log the active rules - dereference the logger to get zerolog.Logger
	logActiveRules(logger, rules)

	// Set the computed rules
	p.rules = rules
}

// createContextBasedRules creates the standard set of context-based rules.
// This uses the actual rule implementations from the rules package.
func createContextBasedRules(ctx context.Context, analyzer domain.CommitAnalyzer) []domain.Rule {
	logger := log.Logger(ctx)
	logger.Debug().Msg("Creating context-based rule set")

	// Get configuration from context
	cfg := config.GetConfig(ctx)
	logger.Debug().
		Int("subject_max_length", cfg.Subject.MaxLength).
		Bool("body_required", cfg.Body.Required).
		Bool("conventional_required", cfg.Conventional.Required).
		Strs("enabled_rules", cfg.Rules.EnabledRules).
		Strs("disabled_rules", cfg.Rules.DisabledRules).
		Msg("Creating rules with configuration from context")

	// Create ALL rules - we will filter them later in FilterRules
	standardRules := []domain.Rule{}

	// SubjectLength rule
	subjectLengthRule := rules.NewSubjectLengthRule(
		rules.WithMaxLength(cfg.Subject.MaxLength),
	)
	standardRules = append(standardRules, subjectLengthRule)
	logger.Debug().Str("rule_name", subjectLengthRule.Name()).Msg("Created rule")

	// ConventionalCommit rule
	conventionalOptions := []rules.ConventionalCommitOption{}
	if len(cfg.Conventional.Types) > 0 {
		conventionalOptions = append(conventionalOptions, rules.WithAllowedTypes(cfg.Conventional.Types))
	}

	if cfg.Conventional.RequireScope {
		conventionalOptions = append(conventionalOptions, rules.WithRequiredScope())
	}

	conventionalRule := rules.NewConventionalCommitRule(conventionalOptions...)
	standardRules = append(standardRules, conventionalRule)
	logger.Debug().Str("rule_name", conventionalRule.Name()).Msg("Created rule")

	// ImperativeVerb rule
	imperativeRule := rules.NewImperativeVerbRule(
		rules.WithImperativeConventionalCommit(cfg.Conventional.Required),
	)
	standardRules = append(standardRules, imperativeRule)
	logger.Debug().Str("rule_name", imperativeRule.Name()).Msg("Created rule")

	// SubjectCase rule
	subjectCaseOptions := []rules.SubjectCaseOption{}
	if cfg.Subject.Case != "" {
		subjectCaseOptions = append(subjectCaseOptions, rules.WithCaseChoice(cfg.Subject.Case))
	}

	subjectCaseRule := rules.NewSubjectCaseRule(subjectCaseOptions...)
	standardRules = append(standardRules, subjectCaseRule)
	logger.Debug().Str("rule_name", subjectCaseRule.Name()).Msg("Created rule")

	// SubjectSuffix rule
	subjectSuffixRule := rules.NewSubjectSuffixRule(
		rules.WithInvalidSuffixes(strings.Join(cfg.Subject.DisallowedSuffixes, ",")),
	)
	standardRules = append(standardRules, subjectSuffixRule)
	logger.Debug().Str("rule_name", subjectSuffixRule.Name()).Msg("Created rule")

	// SignOff rule
	signOffRule := rules.NewSignOffRule(
		rules.WithRequireSignOff(cfg.Security.SignOffRequired),
		rules.WithAllowMultipleSignOffs(cfg.Security.AllowMultipleSignOffs),
	)
	standardRules = append(standardRules, signOffRule)
	logger.Debug().Str("rule_name", signOffRule.Name()).Msg("Created rule")

	// Signature rule
	signatureRule := rules.NewSignatureRule(
		rules.WithRequireSignature(cfg.Security.GPGRequired),
		rules.WithAllowedSignatureTypes(cfg.Security.AllowedSignatureTypes),
	)
	standardRules = append(standardRules, signatureRule)
	logger.Debug().Str("rule_name", signatureRule.Name()).Msg("Created rule")

	// Spell rule
	spellOptions := []rules.SpellRuleOption{}
	if cfg.SpellCheck.Language != "" {
		spellOptions = append(spellOptions, rules.WithLocale(cfg.SpellCheck.Language))
	}

	if cfg.SpellCheck.IgnoreCase {
		spellOptions = append(spellOptions, rules.WithIgnoreCase(true))
	}

	if len(cfg.SpellCheck.CustomDictionary) > 0 {
		spellOptions = append(spellOptions, rules.WithCustomWords(cfg.SpellCheck.CustomDictionary))
	}
	// Use WithMaxErrors with a default value of 3
	spellOptions = append(spellOptions, rules.WithMaxErrors(3))

	spellRule := rules.NewSpellRule(spellOptions...)
	standardRules = append(standardRules, spellRule)
	logger.Debug().Str("rule_name", spellRule.Name()).Msg("Created rule")

	// CommitsAhead rule - only add if analyzer is available
	if analyzer != nil {
		commitsAheadRule := rules.NewCommitsAheadRule(
			rules.WithMaxCommitsAhead(cfg.Repository.MaxCommitsAhead),
			rules.WithReference(cfg.Repository.ReferenceBranch),
			rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				return analyzer
			}),
		)
		standardRules = append(standardRules, commitsAheadRule)
		logger.Debug().Str("rule_name", commitsAheadRule.Name()).Msg("Created rule")
	}

	// SignedIdentity rule - only add if key directory is configured
	if cfg.Security.KeyDirectory != "" {
		signedIdentityRule := rules.NewSignedIdentityRule(
			rules.WithKeyDirectory(cfg.Security.KeyDirectory),
		)
		standardRules = append(standardRules, signedIdentityRule)
		logger.Debug().Str("rule_name", signedIdentityRule.Name()).Msg("Created rule")
	}

	// CommitBody rule
	commitBodyRule := rules.NewCommitBodyRule(
		rules.WithRequireBody(cfg.Body.Required),
		rules.WithAllowSignOffOnly(cfg.Body.AllowSignOffOnly),
	)
	standardRules = append(standardRules, commitBodyRule)
	logger.Debug().Str("rule_name", commitBodyRule.Name()).Msg("Created rule")

	// JiraReference rule
	jiraOptions := []rules.JiraReferenceOption{}
	if len(cfg.Jira.Projects) > 0 {
		jiraOptions = append(jiraOptions, rules.WithValidProjects(cfg.Jira.Projects))
	}

	if cfg.Jira.BodyRef {
		jiraOptions = append(jiraOptions, rules.WithBodyRefChecking())
	}

	jiraRule := rules.NewJiraReferenceRule(jiraOptions...)
	standardRules = append(standardRules, jiraRule)
	logger.Debug().Str("rule_name", jiraRule.Name()).Msg("Created rule")

	// Log the rules we created
	ruleNames := make([]string, 0, len(standardRules))
	for _, rule := range standardRules {
		ruleNames = append(ruleNames, rule.Name())
	}

	logger.Debug().Strs("created_rules", ruleNames).Int("created_count", len(standardRules)).Msg("Created standard rules")

	return standardRules
}
