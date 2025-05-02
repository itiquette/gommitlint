// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import "github.com/itiquette/gommitlint/internal/config"

// ConfigAdapter adapts Config to the ValidationConfig interface needed by the validation service.
type ConfigAdapter struct {
	config config.Config
}

// NewConfigAdapter creates a new adapter for Config.
func NewConfigAdapter(cfg config.Config) *ConfigAdapter {
	return &ConfigAdapter{
		config: cfg,
	}
}

// This adapter implements all the ValidationConfig methods by delegating to the underlying Config

// SubjectRequireImperative delegates to the underlying config.
func (a *ConfigAdapter) SubjectRequireImperative() bool {
	return a.config.SubjectRequireImperative()
}

// SubjectMaxLength delegates to the underlying config.
func (a *ConfigAdapter) SubjectMaxLength() int {
	return a.config.SubjectMaxLength()
}

// SubjectCase delegates to the underlying config.
func (a *ConfigAdapter) SubjectCase() string {
	return a.config.SubjectCase()
}

// SubjectInvalidSuffixes delegates to the underlying config.
func (a *ConfigAdapter) SubjectInvalidSuffixes() string {
	// The domain interface expects a string
	return a.config.SubjectInvalidSuffixes()
}

// BodyRequired delegates to the underlying config.
func (a *ConfigAdapter) BodyRequired() bool {
	return a.config.BodyRequired()
}

// BodyAllowSignOffOnly delegates to the underlying config.
func (a *ConfigAdapter) BodyAllowSignOffOnly() bool {
	return a.config.BodyAllowSignOffOnly()
}

// ConventionalRequired delegates to the underlying config.
func (a *ConfigAdapter) ConventionalRequired() bool {
	return a.config.ConventionalRequired()
}

// ConventionalTypes delegates to the underlying config.
func (a *ConfigAdapter) ConventionalTypes() []string {
	return a.config.ConventionalTypes()
}

// ConventionalScopes delegates to the underlying config.
func (a *ConfigAdapter) ConventionalScopes() []string {
	return a.config.ConventionalScopes()
}

// ConventionalMaxDescriptionLength delegates to the underlying config.
func (a *ConfigAdapter) ConventionalMaxDescriptionLength() int {
	return a.config.ConventionalMaxDescriptionLength()
}

// JiraRequired delegates to the underlying config.
func (a *ConfigAdapter) JiraRequired() bool {
	return a.config.JiraRequired()
}

// JiraProjects delegates to the underlying config.
func (a *ConfigAdapter) JiraProjects() []string {
	return a.config.JiraProjects()
}

// JiraPattern delegates to the underlying config.
func (a *ConfigAdapter) JiraPattern() string {
	return a.config.JiraPattern()
}

// JiraBodyRef delegates to the underlying config.
func (a *ConfigAdapter) JiraBodyRef() bool {
	return a.config.JiraBodyRef()
}

// JiraStrict delegates to the underlying config.
func (a *ConfigAdapter) JiraStrict() bool {
	return a.config.JiraStrict()
}

// SignatureRequired delegates to the underlying config.
func (a *ConfigAdapter) SignatureRequired() bool {
	return a.config.SignatureRequired()
}

// SignOffRequired delegates to the underlying config.
func (a *ConfigAdapter) SignOffRequired() bool {
	return a.config.SignOffRequired()
}

// AllowMultipleSignOffs delegates to the underlying config.
func (a *ConfigAdapter) AllowMultipleSignOffs() bool {
	return a.config.AllowMultipleSignOffs()
}

// AllowedSignatureTypes delegates to the underlying config.
func (a *ConfigAdapter) AllowedSignatureTypes() []string {
	return a.config.AllowedSignatureTypes()
}

// IdentityPublicKeyURI delegates to the underlying config.
func (a *ConfigAdapter) IdentityPublicKeyURI() string {
	return a.config.IdentityPublicKeyURI()
}

// SpellEnabled delegates to the underlying config.
func (a *ConfigAdapter) SpellEnabled() bool {
	return a.config.SpellEnabled()
}

// SpellLocale delegates to the underlying config.
func (a *ConfigAdapter) SpellLocale() string {
	return a.config.SpellLocale()
}

// SpellMaxErrors delegates to the underlying config.
func (a *ConfigAdapter) SpellMaxErrors() int {
	return a.config.SpellMaxErrors()
}

// SpellIgnoreWords delegates to the underlying config.
func (a *ConfigAdapter) SpellIgnoreWords() []string {
	return a.config.SpellIgnoreWords()
}

// SpellCustomWords delegates to the underlying config.
func (a *ConfigAdapter) SpellCustomWords() map[string]string {
	// Create a copy of the custom words map to maintain immutability
	if a.config.SpellCheck.CustomWords == nil {
		return make(map[string]string)
	}

	// Return a copy of the map
	result := make(map[string]string, len(a.config.SpellCheck.CustomWords))
	for word, replacement := range a.config.SpellCheck.CustomWords {
		result[word] = replacement
	}

	return result
}

// ReferenceBranch delegates to the underlying config.
func (a *ConfigAdapter) ReferenceBranch() string {
	return a.config.ReferenceBranch()
}

// IgnoreMergeCommits delegates to the underlying config.
func (a *ConfigAdapter) IgnoreMergeCommits() bool {
	return a.config.IgnoreMergeCommits()
}

// MaxCommitsAhead delegates to the underlying config.
func (a *ConfigAdapter) MaxCommitsAhead() int {
	return a.config.MaxCommitsAhead()
}

// CheckCommitsAhead delegates to the underlying config.
func (a *ConfigAdapter) CheckCommitsAhead() bool {
	return a.config.CheckCommitsAhead()
}

// EnabledRules delegates to the underlying config.
func (a *ConfigAdapter) EnabledRules() []string {
	return a.config.EnabledRules()
}

// DisabledRules delegates to the underlying config.
func (a *ConfigAdapter) DisabledRules() []string {
	return a.config.DisabledRules()
}

// GetActiveRules implements the domain.RuleConfigProvider interface.
// We don't need active rules here as the validation service will calculate them as needed.
func (a *ConfigAdapter) GetActiveRules() []string {
	// Return an empty slice or enabled rules if available
	if len(a.config.EnabledRules()) > 0 {
		return a.config.EnabledRules()
	}

	return []string{}
}

// GetAvailableRules implements the domain.RuleConfigProvider interface.
func (a *ConfigAdapter) GetAvailableRules() []string {
	// Return an empty slice - the rule provider will determine available rules
	return []string{}
}

// These methods are used by the validation service to set rules
// We'll update the underlying config to maintain immutability

// SetEnabledRules sets the enabled rules.
func (a *ConfigAdapter) SetEnabledRules(ruleNames []string) {
	a.config = a.config.WithEnabledRules(ruleNames)
}

// SetDisabledRules sets the disabled rules.
func (a *ConfigAdapter) SetDisabledRules(ruleNames []string) {
	a.config = a.config.WithDisabledRules(ruleNames)
}
