// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides a consolidated configuration adapter.
package config

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/config"
	configTypes "github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
)

// Adapter is a unified configuration adapter that implements all config interfaces.
type Adapter struct {
	cfg configTypes.Config
}

// NewAdapter creates a new unified configuration adapter.
func NewAdapter(cfg configTypes.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Ensure Adapter implements all required interfaces.
var _ config.Config = (*Adapter)(nil)

// GetConfig returns the concrete config structure.
// This is the preferred method for accessing config in the application.
func (a *Adapter) GetConfig() configTypes.Config {
	return a.cfg
}

// SetConfig allows updating the internal config.
// This maintains the current pattern where config can be replaced.
func (a *Adapter) SetConfig(cfg configTypes.Config) {
	a.cfg = cfg
}

// Get returns a configuration value for a given key.
func (a *Adapter) Get(key string) interface{} {
	parts := strings.Split(key, ".")
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "subject":
		return a.getSubjectValue(parts[1:])
	case "body":
		return a.getBodyValue(parts[1:])
	case "conventional":
		return a.getConventionalValue(parts[1:])
	case "rules":
		return a.getRulesValue(parts[1:])
	case "security":
		return a.getSecurityValue(parts[1:])
	case "repository":
		return a.getRepositoryValue(parts[1:])
	case "output":
		return a.getOutputValue(parts[1:])
	case "spell":
		return a.getSpellValue(parts[1:])
	case "jira":
		return a.getJiraValue(parts[1:])
	default:
		return nil
	}
}

// GetString returns a string configuration value for a given key.
func (a *Adapter) GetString(key string) string {
	val := a.Get(key)
	if s, ok := val.(string); ok {
		return s
	}

	return ""
}

// GetBool returns a boolean configuration value for a given key.
func (a *Adapter) GetBool(key string) bool {
	val := a.Get(key)
	if b, ok := val.(bool); ok {
		return b
	}

	return false
}

// GetInt returns an integer configuration value for a given key.
func (a *Adapter) GetInt(key string) int {
	val := a.Get(key)
	if i, ok := val.(int); ok {
		return i
	}

	return 0
}

// GetStringSlice returns a string slice configuration value for a given key.
func (a *Adapter) GetStringSlice(key string) []string {
	val := a.Get(key)
	if ss, ok := val.([]string); ok {
		return ss
	}

	return nil
}

// ValidationConfig interface implementation

// EnabledRules returns the list of explicitly enabled rules.
func (a *Adapter) EnabledRules() []string {
	return a.cfg.Rules.Enabled
}

// DisabledRules returns the list of explicitly disabled rules.
func (a *Adapter) DisabledRules() []string {
	return a.cfg.Rules.Disabled
}

// SubjectMaxLength returns the maximum length for commit subjects.
func (a *Adapter) SubjectMaxLength() int {
	if a.cfg.Subject.MaxLength <= 0 {
		return 72 // Default
	}

	return a.cfg.Subject.MaxLength
}

// BodyAllowSignOffOnly returns whether a sign-off alone is sufficient for the body.
func (a *Adapter) BodyAllowSignOffOnly() bool {
	return a.cfg.Body.AllowSignOffOnly
}

// ConventionalTypes returns the allowed types for conventional commits.
func (a *Adapter) ConventionalTypes() []string {
	return a.cfg.Conventional.Types
}

// ConventionalScopes returns the allowed scopes for conventional commits.
func (a *Adapter) ConventionalScopes() []string {
	return a.cfg.Conventional.Scopes
}

// ConventionalScopeRequired returns whether a scope is required in conventional commits.
func (a *Adapter) ConventionalScopeRequired() bool {
	return a.cfg.Conventional.RequireScope
}

// SubjectCase returns the required case for commit subjects.
func (a *Adapter) SubjectCase() string {
	return a.cfg.Subject.Case
}

// SubjectImperative returns whether commit subjects must use imperative mood.
func (a *Adapter) SubjectImperative() bool {
	return a.cfg.Subject.Imperative
}

// SubjectInvalidSuffixes returns a list of suffixes that are not allowed in commit subjects.
func (a *Adapter) SubjectInvalidSuffixes() []string {
	return a.cfg.Subject.DisallowedSuffixes
}

// JiraBodyRef returns whether a JIRA reference is required in the commit body.
func (a *Adapter) JiraBodyRef() bool {
	return a.cfg.Jira.BodyRef
}

// JiraProjects returns the list of valid JIRA project prefixes.
func (a *Adapter) JiraProjects() []string {
	return a.cfg.Jira.Projects
}

// SpellLocale returns the locale to use for spell checking.
func (a *Adapter) SpellLocale() string {
	return a.cfg.SpellCheck.Language
}

// SpellMaxErrors returns the maximum number of spelling errors allowed.
func (a *Adapter) SpellMaxErrors() int {
	// This field doesn't exist in the config, returning a default
	return 0
}

// SpellIgnoreWords returns a list of words to ignore during spell checking.
func (a *Adapter) SpellIgnoreWords() []string {
	// The spell ignore words are stored in CustomDictionary
	return a.cfg.SpellCheck.CustomDictionary
}

// SpellCustomWords returns a list of custom words to consider valid during spell checking.
func (a *Adapter) SpellCustomWords() []string {
	return a.cfg.SpellCheck.CustomDictionary
}

// MaxCommitsAhead returns the maximum number of commits allowed ahead of the main branch.
func (a *Adapter) MaxCommitsAhead() int {
	return a.cfg.Repository.MaxCommitsAhead
}

// RulesConfig interface implementation

// BodyMinLength returns the minimum length for commit body.
func (a *Adapter) BodyMinLength() int {
	return a.cfg.Body.MinLength
}

// ReferenceBranch returns the reference branch for comparison.
func (a *Adapter) ReferenceBranch() string {
	return a.cfg.Repository.ReferenceBranch
}

// IgnoreMergeCommits returns whether to ignore merge commits.
func (a *Adapter) IgnoreMergeCommits() bool {
	return a.cfg.Repository.IgnoreMergeCommits
}

// ConventionalMaxDescriptionLength returns the maximum description length for conventional commits.
func (a *Adapter) ConventionalMaxDescriptionLength() int {
	return a.cfg.Conventional.MaxDescriptionLength
}

// RuleConfiguration interface implementation

// IsRuleEnabled checks if a specific rule is enabled.
func (a *Adapter) IsRuleEnabled(ruleName string) bool {
	// First, check if it's explicitly enabled
	for _, rule := range a.cfg.Rules.Enabled {
		if rule == ruleName {
			return true
		}
	}

	// Then check if it's explicitly disabled
	for _, rule := range a.cfg.Rules.Disabled {
		if rule == ruleName {
			return false
		}
	}

	// Finally, check default disabled rules
	defaultDisabled := domain.GetDefaultDisabledRules()
	if defaultDisabled[ruleName] {
		return false
	}

	// If not in any list, it's enabled by default
	return true
}

// IsRuleDisabled checks if a specific rule is disabled.
func (a *Adapter) IsRuleDisabled(ruleName string) bool {
	// If explicitly enabled, it's not disabled
	for _, rule := range a.cfg.Rules.Enabled {
		if rule == ruleName {
			return false
		}
	}

	// If explicitly disabled, it's disabled
	for _, rule := range a.cfg.Rules.Disabled {
		if rule == ruleName {
			return true
		}
	}

	// Check default disabled rules
	defaultDisabled := domain.GetDefaultDisabledRules()

	return defaultDisabled[ruleName]
}

// GetEnabledRules returns the list of explicitly enabled rules.
func (a *Adapter) GetEnabledRules() []string {
	return a.cfg.Rules.Enabled
}

// GetDisabledRules returns the list of explicitly disabled rules.
func (a *Adapter) GetDisabledRules() []string {
	return a.cfg.Rules.Disabled
}

// GetRuleConfig returns configuration for a specific rule.
func (a *Adapter) GetRuleConfig(ruleName string) map[string]interface{} {
	config := make(map[string]interface{})

	switch ruleName {
	case "SubjectLength":
		config["max_length"] = a.cfg.Subject.MaxLength
	case "SubjectCase":
		config["case"] = a.cfg.Subject.Case
	case "SubjectSuffix":
		config["disallowed_suffixes"] = a.cfg.Subject.DisallowedSuffixes
	case "ConventionalCommit":
		config["require_scope"] = a.cfg.Conventional.RequireScope
		config["types"] = a.cfg.Conventional.Types
		config["scopes"] = a.cfg.Conventional.Scopes
		config["max_description_length"] = a.cfg.Conventional.MaxDescriptionLength
	case "CommitBody":
		config["min_length"] = a.cfg.Body.MinLength
		config["minimum_lines"] = a.cfg.Body.MinLines
	case "JiraReference":
		config["pattern"] = a.cfg.Jira.Pattern
		config["projects"] = a.cfg.Jira.Projects
		config["body_ref"] = a.cfg.Jira.BodyRef
	case "SignOff":
		config["sign_off_required"] = a.cfg.Body.RequireSignOff
	case "Signature":
		config["gpg_required"] = a.cfg.Security.GPGRequired
	case "SignedIdentity":
		config["allowed_identities"] = a.cfg.Security.AllowedIdentities
	case "Spell":
		config["language"] = a.cfg.SpellCheck.Language
		config["custom_dictionary"] = a.cfg.SpellCheck.CustomDictionary
	}

	return config
}

// Helper methods for Get()

func (a *Adapter) getSubjectValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Subject
	}

	switch parts[0] {
	case "case":
		return a.cfg.Subject.Case
	case "max_length":
		return a.cfg.Subject.MaxLength
	case "require_imperative":
		return a.cfg.Subject.Imperative
	case "disallowed_suffixes":
		return a.cfg.Subject.DisallowedSuffixes
	default:
		return nil
	}
}

func (a *Adapter) getBodyValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Body
	}

	switch parts[0] {
	case "min_length":
		return a.cfg.Body.MinLength
	case "minimum_lines", "min_lines":
		return a.cfg.Body.MinLines
	case "allow_sign_off_only":
		return a.cfg.Body.AllowSignOffOnly
	case "require_sign_off":
		return a.cfg.Body.RequireSignOff
	default:
		return nil
	}
}

func (a *Adapter) getConventionalValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Conventional
	}

	switch parts[0] {
	case "require_scope":
		return a.cfg.Conventional.RequireScope
	case "types":
		return a.cfg.Conventional.Types
	case "scopes":
		return a.cfg.Conventional.Scopes
	case "allow_breaking_changes":
		return a.cfg.Conventional.AllowBreakingChanges
	case "max_description_length":
		return a.cfg.Conventional.MaxDescriptionLength
	default:
		return nil
	}
}

func (a *Adapter) getRulesValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Rules
	}

	switch parts[0] {
	case "enabled_rules", "enabled":
		return a.cfg.Rules.Enabled
	case "disabled_rules", "disabled":
		return a.cfg.Rules.Disabled
	default:
		return nil
	}
}

func (a *Adapter) getSecurityValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Security
	}

	switch parts[0] {
	case "gpg_required":
		return a.cfg.Security.GPGRequired
	case "key_directory":
		return a.cfg.Security.KeyDirectory
	case "allowed_signature_types":
		return a.cfg.Security.AllowedSignatureTypes
	case "allowed_keyrings":
		return a.cfg.Security.AllowedKeyrings
	case "allowed_identities":
		return a.cfg.Security.AllowedIdentities
	case "allow_multiple_sign_offs":
		return a.cfg.Security.MultipleSignoffs
	default:
		return nil
	}
}

func (a *Adapter) getRepositoryValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Repository
	}

	switch parts[0] {
	case "path":
		return a.cfg.Repository.Path
	case "reference_branch":
		return a.cfg.Repository.ReferenceBranch
	case "max_commits_ahead":
		return a.cfg.Repository.MaxCommitsAhead
	case "max_history_days":
		return a.cfg.Repository.MaxHistoryDays
	case "ignore_merge_commits":
		return a.cfg.Repository.IgnoreMergeCommits
	default:
		return nil
	}
}

func (a *Adapter) getOutputValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Output
	}

	switch parts[0] {
	case "format":
		return a.cfg.Output.Format
	case "verbose":
		return a.cfg.Output.Verbose
	case "quiet":
		return a.cfg.Output.Quiet
	default:
		return nil
	}
}

func (a *Adapter) getSpellValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.SpellCheck
	}

	switch parts[0] {
	case "language":
		return a.cfg.SpellCheck.Language
	case "ignore_case":
		return a.cfg.SpellCheck.IgnoreCase
	case "custom_dictionary":
		return a.cfg.SpellCheck.CustomDictionary
	default:
		return nil
	}
}

func (a *Adapter) getJiraValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Jira
	}

	switch parts[0] {
	case "pattern":
		return a.cfg.Jira.Pattern
	case "projects":
		if len(parts) == 1 {
			return a.cfg.Jira.Projects
		}
		// Handle array indexing like "projects.0"
		if len(parts) > 1 {
			// Parse the index
			var idx int
			if _, err := fmt.Sscanf(parts[1], "%d", &idx); err == nil {
				if idx >= 0 && idx < len(a.cfg.Jira.Projects) {
					return a.cfg.Jira.Projects[idx]
				}
			}
		}

		return nil
	case "body_ref":
		return a.cfg.Jira.BodyRef
	default:
		return nil
	}
}
