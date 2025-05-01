// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

import (
	"github.com/itiquette/gommitlint/internal/domain"
)

// ValidationConfigAdapter adapts the Config to implement all the domain interfaces.
// This is a bridge to ensure backward compatibility during the migration.
type ValidationConfigAdapter struct {
	config Config
}

// NewValidationConfigAdapter creates a new adapter for the given configuration.
func NewValidationConfigAdapter(config Config) *ValidationConfigAdapter {
	return &ValidationConfigAdapter{
		config: config,
	}
}

// Make sure ValidationConfigAdapter implements all required interfaces.
var (
	_ domain.SubjectConfigProvider      = (*ValidationConfigAdapter)(nil)
	_ domain.JiraConfigProvider         = (*ValidationConfigAdapter)(nil)
	_ domain.BodyConfigProvider         = (*ValidationConfigAdapter)(nil)
	_ domain.ConventionalConfigProvider = (*ValidationConfigAdapter)(nil)
	_ domain.SecurityConfigProvider     = (*ValidationConfigAdapter)(nil)
	_ domain.SpellCheckConfigProvider   = (*ValidationConfigAdapter)(nil)
	_ domain.RepositoryConfigProvider   = (*ValidationConfigAdapter)(nil)
	_ domain.RuleConfigProvider         = (*ValidationConfigAdapter)(nil)
)

// Subject configuration methods

// SubjectMaxLength returns the maximum length allowed for commit subjects.
func (a *ValidationConfigAdapter) SubjectMaxLength() int {
	return a.config.SubjectMaxLength()
}

// SubjectCase returns the required case style for commit subjects.
func (a *ValidationConfigAdapter) SubjectCase() string {
	return a.config.SubjectCase()
}

// SubjectRequireImperative returns whether commit subjects must use imperative mood.
func (a *ValidationConfigAdapter) SubjectRequireImperative() bool {
	return a.config.SubjectRequireImperative()
}

// SubjectInvalidSuffixes returns characters that cannot be used at the end of the subject.
func (a *ValidationConfigAdapter) SubjectInvalidSuffixes() string {
	return a.config.SubjectInvalidSuffixes()
}

// Body configuration methods

// BodyRequired returns whether a commit body is required.
func (a *ValidationConfigAdapter) BodyRequired() bool {
	return a.config.BodyRequired()
}

// BodyAllowSignOffOnly returns whether a commit body can consist of only a sign-off line.
func (a *ValidationConfigAdapter) BodyAllowSignOffOnly() bool {
	return a.config.BodyAllowSignOffOnly()
}

// Conventional commit configuration methods

// ConventionalTypes returns the allowed types for conventional commits.
func (a *ValidationConfigAdapter) ConventionalTypes() []string {
	return a.config.ConventionalTypes()
}

// ConventionalScopes returns the allowed scopes for conventional commits.
func (a *ValidationConfigAdapter) ConventionalScopes() []string {
	return a.config.ConventionalScopes()
}

// ConventionalMaxDescriptionLength returns the maximum length allowed for conventional commit descriptions.
func (a *ValidationConfigAdapter) ConventionalMaxDescriptionLength() int {
	return a.config.ConventionalMaxDescriptionLength()
}

// ConventionalRequired returns whether conventional commit format is required.
func (a *ValidationConfigAdapter) ConventionalRequired() bool {
	return a.config.ConventionalRequired()
}

// JIRA configuration methods

// JiraProjects returns the list of JIRA project identifiers to validate against.
func (a *ValidationConfigAdapter) JiraProjects() []string {
	return a.config.JiraProjects()
}

// JiraBodyRef returns whether JIRA references in the commit body are accepted.
func (a *ValidationConfigAdapter) JiraBodyRef() bool {
	return a.config.JiraBodyRef()
}

// JiraRequired returns whether JIRA ticket references are required in commits.
func (a *ValidationConfigAdapter) JiraRequired() bool {
	return a.config.JiraRequired()
}

// JiraPattern returns the regex pattern used to match JIRA ticket references.
func (a *ValidationConfigAdapter) JiraPattern() string {
	return a.config.JiraPattern()
}

// JiraStrict returns whether strict JIRA validation is enabled.
func (a *ValidationConfigAdapter) JiraStrict() bool {
	return a.config.JiraStrict()
}

// Security configuration methods

// SignatureRequired returns whether commit signatures are required.
func (a *ValidationConfigAdapter) SignatureRequired() bool {
	return a.config.SignatureRequired()
}

// AllowedSignatureTypes returns the list of allowed signature types for commits.
func (a *ValidationConfigAdapter) AllowedSignatureTypes() []string {
	return a.config.AllowedSignatureTypes()
}

// SignOffRequired returns whether commit sign-offs are required.
func (a *ValidationConfigAdapter) SignOffRequired() bool {
	return a.config.SignOffRequired()
}

// AllowMultipleSignOffs returns whether multiple sign-offs are allowed in commits.
func (a *ValidationConfigAdapter) AllowMultipleSignOffs() bool {
	return a.config.AllowMultipleSignOffs()
}

// IdentityPublicKeyURI returns the URI for retrieving public keys for identity verification.
func (a *ValidationConfigAdapter) IdentityPublicKeyURI() string {
	return a.config.IdentityPublicKeyURI()
}

// SpellCheck configuration methods

// SpellLocale returns the language locale to use for spell checking.
func (a *ValidationConfigAdapter) SpellLocale() string {
	return a.config.SpellLocale()
}

// SpellEnabled returns whether spell checking is enabled.
func (a *ValidationConfigAdapter) SpellEnabled() bool {
	return a.config.SpellEnabled()
}

// SpellIgnoreWords returns the list of words to ignore during spell checking.
func (a *ValidationConfigAdapter) SpellIgnoreWords() []string {
	return a.config.SpellIgnoreWords()
}

// SpellCustomWords returns the map of custom words and their replacements for spell checking.
func (a *ValidationConfigAdapter) SpellCustomWords() map[string]string {
	return a.config.SpellCustomWords()
}

// SpellMaxErrors returns the maximum number of spelling errors to report.
func (a *ValidationConfigAdapter) SpellMaxErrors() int {
	return a.config.SpellMaxErrors()
}

// Repository configuration methods

// ReferenceBranch returns the name of the reference branch used for comparison.
func (a *ValidationConfigAdapter) ReferenceBranch() string {
	return a.config.ReferenceBranch()
}

// IgnoreMergeCommits returns whether merge commits should be ignored during validation.
func (a *ValidationConfigAdapter) IgnoreMergeCommits() bool {
	return a.config.IgnoreMergeCommits()
}

// MaxCommitsAhead returns the maximum number of commits allowed ahead of the reference branch.
func (a *ValidationConfigAdapter) MaxCommitsAhead() int {
	return a.config.MaxCommitsAhead()
}

// CheckCommitsAhead returns whether to validate the count of commits ahead of the reference branch.
func (a *ValidationConfigAdapter) CheckCommitsAhead() bool {
	return a.config.CheckCommitsAhead()
}

// Rule configuration methods

// EnabledRules returns the list of explicitly enabled validation rules.
func (a *ValidationConfigAdapter) EnabledRules() []string {
	return a.config.EnabledRules()
}

// DisabledRules returns the list of explicitly disabled validation rules.
func (a *ValidationConfigAdapter) DisabledRules() []string {
	return a.config.DisabledRules()
}

// GetAvailableRules returns a list of all available rule names.
func (a *ValidationConfigAdapter) GetAvailableRules() []string {
	return a.config.GetAvailableRules()
}

// GetActiveRules returns a list of currently active rule names.
func (a *ValidationConfigAdapter) GetActiveRules() []string {
	return a.config.GetActiveRules()
}

// SetEnabledRules sets the list of explicitly enabled validation rules.
// This is used by the rule engine to activate specific rules.
// Since we're using value semantics, this creates a new config with the updated rules
// and replaces the adapter's config.
func (a *ValidationConfigAdapter) SetEnabledRules(ruleNames []string) {
	a.config = a.config.WithEnabledRules(ruleNames)
}

// SetDisabledRules sets the list of explicitly disabled validation rules.
// This is used by the rule engine to deactivate specific rules.
// Since we're using value semantics, this creates a new config with the updated rules
// and replaces the adapter's config.
func (a *ValidationConfigAdapter) SetDisabledRules(ruleNames []string) {
	a.config = a.config.WithDisabledRules(ruleNames)
}
