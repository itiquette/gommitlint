// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides a consolidated configuration adapter.
package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/config"
	configTypes "github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
)

// RuleConfig is a map of configuration values for a rule.
type RuleConfig map[string]interface{}

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
func (a *Adapter) GetConfig() configTypes.Config {
	return a.cfg
}

// Config interface implementation
// ============================================================================

// Get returns a configuration value for a given key.
// It supports nested keys with dot notation (e.g., "message.subject.max_length").
func (a *Adapter) Get(key string) interface{} {
	if key == "" {
		return nil
	}

	parts := strings.Split(key, ".")

	switch parts[0] {
	case "message":
		return a.getMessageValue(parts[1:])
	case "conventional":
		return a.getConventionalValue(parts[1:])
	case "signing":
		return a.getSigningValue(parts[1:])
	case "repo":
		return a.getRepoValue(parts[1:])
	case "jira":
		return a.getJiraValue(parts[1:])
	case "spell":
		return a.getSpellValue(parts[1:])
	case "rules":
		return a.getRulesValue(parts[1:])
	case "output":
		return a.cfg.Output
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
	if a.cfg.Message.Subject.MaxLength <= 0 {
		return 72 // Default
	}

	return a.cfg.Message.Subject.MaxLength
}

// BodyAllowSignOffOnly returns whether a sign-off alone is sufficient for the body.
func (a *Adapter) BodyAllowSignOffOnly() bool {
	return a.cfg.Message.Body.AllowSignoffOnly
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

// SubjectImperative returns whether imperative mood is required in subjects.
func (a *Adapter) SubjectImperative() bool {
	return a.cfg.Message.Subject.RequireImperative
}

// BodyMinLength returns the minimum length for commit body.
func (a *Adapter) BodyMinLength() int {
	return a.cfg.Message.Body.MinLength
}

// JiraProjects returns the list of JIRA projects.
func (a *Adapter) JiraProjects() []string {
	return a.cfg.Jira.Projects
}

// RulesConfig interface implementation

// BodyMinimumLines returns the minimum number of lines in the body.
func (a *Adapter) BodyMinimumLines() int {
	return a.cfg.Message.Body.MinLines
}

// CommitBodyRequired returns whether a commit body is required.
func (a *Adapter) CommitBodyRequired() bool {
	return a.cfg.Message.Body.MinLength > 0
}

// ReferenceBranch returns the name of the reference branch for comparison.
func (a *Adapter) ReferenceBranch() string {
	return a.cfg.Repo.Branch
}

// MaxCommitsAhead returns the maximum allowed commits ahead of reference branch.
func (a *Adapter) MaxCommitsAhead() int {
	return a.cfg.Repo.MaxCommitsAhead
}

// SpellIgnoreWords returns words to ignore during spell checking.
func (a *Adapter) SpellIgnoreWords() []string {
	return a.cfg.Spell.IgnoreWords
}

// SubjectCase returns the required case for commit subjects.
func (a *Adapter) SubjectCase() string {
	return a.cfg.Message.Subject.Case
}

// Rule state methods

// IsRuleEnabled checks if a rule is enabled by name.
// Rules can be explicitly enabled, explicitly disabled, or have default behavior.
func (a *Adapter) IsRuleEnabled(ctx context.Context, ruleName string) bool {
	// Get the centralized rule priority service
	priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())

	// Use the centralized function
	return priorityService.IsRuleEnabled(ctx, ruleName, a.cfg.Rules.Enabled, a.cfg.Rules.Disabled)
}

// IsRuleDisabled checks if a rule is disabled by name.
func (a *Adapter) IsRuleDisabled(ctx context.Context, ruleName string) bool {
	return !a.IsRuleEnabled(ctx, ruleName)
}

// RuleConfiguration interface implementation

// GetRuleConfig returns the configuration for a specific rule by name.
// It gathers all relevant configuration values for the rule.
func (a *Adapter) GetRuleConfig(ruleName string) RuleConfig {
	// Create an empty rule config
	config := make(RuleConfig)

	// Based on the rule name, populate specific configuration values
	switch ruleName {
	case "SubjectLength":
		config["max_length"] = a.cfg.Message.Subject.MaxLength
	case "SubjectCase":
		config["case"] = a.cfg.Message.Subject.Case
	case "Imperative", "ImperativeVerb":
		config["require_imperative"] = a.cfg.Message.Subject.RequireImperative
	case "SubjectSuffix":
		config["forbid_endings"] = a.cfg.Message.Subject.ForbidEndings
	case "Conventional", "ConventionalCommit":
		config["require_scope"] = a.cfg.Conventional.RequireScope
		config["types"] = a.cfg.Conventional.Types
		config["scopes"] = a.cfg.Conventional.Scopes
		config["max_description_length"] = a.cfg.Conventional.MaxDescriptionLength
	case "CommitBody":
		config["min_length"] = a.cfg.Message.Body.MinLength
		config["minimum_lines"] = a.cfg.Message.Body.MinLines
	case "JiraReference":
		config["pattern"] = a.cfg.Jira.Pattern
		config["projects"] = a.cfg.Jira.Projects
		config["body_ref"] = a.cfg.Jira.CheckBody
	case "SignOff":
		config["sign_off_required"] = a.cfg.Message.Body.RequireSignoff
		config["allow_multiple_signoffs"] = a.cfg.Signing.AllowMultipleSignoffs
	case "Signature":
		config["signature_required"] = a.cfg.Signing.RequireSignature
	case "SignedIdentity":
		config["allowed_identities"] = a.cfg.Signing.AllowedSigners
	case "Spell":
		config["language"] = a.cfg.Spell.Language
		config["custom_dictionary"] = a.cfg.Spell.IgnoreWords
	}

	return config
}

// Helper methods for Get()

func (a *Adapter) getMessageValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Message
	}

	switch parts[0] {
	case "subject":
		return a.getSubjectValue(parts[1:])
	case "body":
		return a.getBodyValue(parts[1:])
	default:
		return nil
	}
}

func (a *Adapter) getSubjectValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Message.Subject
	}

	switch parts[0] {
	case "max_length":
		return a.cfg.Message.Subject.MaxLength
	case "case":
		return a.cfg.Message.Subject.Case
	case "require_imperative":
		return a.cfg.Message.Subject.RequireImperative
	case "forbid_endings":
		return a.cfg.Message.Subject.ForbidEndings
	default:
		return nil
	}
}

func (a *Adapter) getBodyValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Message.Body
	}

	switch parts[0] {
	case "min_length":
		return a.cfg.Message.Body.MinLength
	case "min_lines":
		return a.cfg.Message.Body.MinLines
	case "allow_signoff_only":
		return a.cfg.Message.Body.AllowSignoffOnly
	case "require_signoff":
		return a.cfg.Message.Body.RequireSignoff
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
		if len(parts) > 1 {
			idx := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &idx); err == nil {
				if idx >= 0 && idx < len(a.cfg.Conventional.Types) {
					return a.cfg.Conventional.Types[idx]
				}
			}
		}

		return a.cfg.Conventional.Types
	case "scopes":
		if len(parts) > 1 {
			idx := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &idx); err == nil {
				if idx >= 0 && idx < len(a.cfg.Conventional.Scopes) {
					return a.cfg.Conventional.Scopes[idx]
				}
			}
		}

		return a.cfg.Conventional.Scopes
	case "allow_breaking":
		return a.cfg.Conventional.AllowBreaking
	case "max_description_length":
		return a.cfg.Conventional.MaxDescriptionLength
	default:
		return nil
	}
}

func (a *Adapter) getSigningValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Signing
	}

	switch parts[0] {
	case "require_signature":
		return a.cfg.Signing.RequireSignature
	case "allow_multiple_signoffs":
		return a.cfg.Signing.AllowMultipleSignoffs
	case "allowed_signers":
		if len(parts) > 1 {
			idx := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &idx); err == nil {
				if idx >= 0 && idx < len(a.cfg.Signing.AllowedSigners) {
					return a.cfg.Signing.AllowedSigners[idx]
				}
			}
		}

		return a.cfg.Signing.AllowedSigners
	default:
		return nil
	}
}

func (a *Adapter) getRepoValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Repo
	}

	switch parts[0] {
	case "path":
		return a.cfg.Repo.Path
	case "branch":
		return a.cfg.Repo.Branch
	case "max_commits_ahead":
		return a.cfg.Repo.MaxCommitsAhead
	case "ignore_merges":
		return a.cfg.Repo.IgnoreMerges
	default:
		return nil
	}
}

// getIntegrationsValue is removed as integrations section is now flattened

func (a *Adapter) getJiraValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Jira
	}

	switch parts[0] {
	case "pattern":
		return a.cfg.Jira.Pattern
	case "projects":
		if len(parts) > 1 {
			idx := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &idx); err == nil {
				if idx >= 0 && idx < len(a.cfg.Jira.Projects) {
					return a.cfg.Jira.Projects[idx]
				}
			}
		}

		return a.cfg.Jira.Projects
	case "check_body":
		return a.cfg.Jira.CheckBody
	default:
		return nil
	}
}

func (a *Adapter) getSpellValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Spell
	}

	switch parts[0] {
	case "language":
		return a.cfg.Spell.Language
	case "ignore_words":
		if len(parts) > 1 {
			idx := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &idx); err == nil {
				if idx >= 0 && idx < len(a.cfg.Spell.IgnoreWords) {
					return a.cfg.Spell.IgnoreWords[idx]
				}
			}
		}

		return a.cfg.Spell.IgnoreWords
	default:
		return nil
	}
}

func (a *Adapter) getRulesValue(parts []string) interface{} {
	if len(parts) == 0 {
		return a.cfg.Rules
	}

	switch parts[0] {
	case "enabled":
		if len(parts) > 1 {
			idx := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &idx); err == nil {
				if idx >= 0 && idx < len(a.cfg.Rules.Enabled) {
					return a.cfg.Rules.Enabled[idx]
				}
			}
		}

		return a.cfg.Rules.Enabled
	case "disabled":
		if len(parts) > 1 {
			idx := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &idx); err == nil {
				if idx >= 0 && idx < len(a.cfg.Rules.Disabled) {
					return a.cfg.Rules.Disabled[idx]
				}
			}
		}

		return a.cfg.Rules.Disabled
	default:
		return nil
	}
}
