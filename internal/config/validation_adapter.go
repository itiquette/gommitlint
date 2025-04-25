// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
)

// ValidationConfigAdapter adapts the Config type to implement the domain interfaces.
// This decouples the validation logic from the specific config implementation.
type ValidationConfigAdapter struct {
	config Config
}

// NewValidationConfigAdapter creates a new adapter for the given configuration.
func NewValidationConfigAdapter(config Config) *ValidationConfigAdapter {
	return &ValidationConfigAdapter{
		config: config,
	}
}

// Subject configuration methods

// SubjectMaxLength returns the maximum length allowed for commit subjects.
func (a *ValidationConfigAdapter) SubjectMaxLength() int {
	return a.config.Subject.MaxLength
}

// SubjectCase returns the required case style for commit subjects.
func (a *ValidationConfigAdapter) SubjectCase() string {
	return a.config.Subject.Case
}

// SubjectImperativeVerb returns whether commit subjects must use imperative verb form.
func (a *ValidationConfigAdapter) SubjectImperativeVerb() bool {
	return a.config.Subject.Imperative
}

// SubjectDisallowSuffixes returns a list of disallowed suffix characters for commit subjects.
func (a *ValidationConfigAdapter) SubjectDisallowSuffixes() []string {
	if a.config.Subject.InvalidSuffixes == "" {
		return []string{}
	}

	return strings.Split(a.config.Subject.InvalidSuffixes, ",")
}

// SubjectInvalidSuffixes returns a string of comma-separated invalid suffixes for commit subjects.
func (a *ValidationConfigAdapter) SubjectInvalidSuffixes() string {
	return a.config.Subject.InvalidSuffixes
}

// SubjectRequireImperative returns whether commit subjects must use imperative mood.
func (a *ValidationConfigAdapter) SubjectRequireImperative() bool {
	return a.config.Subject.Imperative
}

// Body configuration methods

// BodyRequired returns whether a commit body is required.
func (a *ValidationConfigAdapter) BodyRequired() bool {
	return a.config.Body.Required
}

// BodyAllowSignOffOnly returns whether a commit body can consist of only a sign-off line.
func (a *ValidationConfigAdapter) BodyAllowSignOffOnly() bool {
	return a.config.Body.AllowSignOffOnly
}

// Conventional commit configuration methods

// ConventionalEnabled returns whether conventional commit validation is enabled.
func (a *ValidationConfigAdapter) ConventionalEnabled() bool {
	// Use Required as a proxy for Enabled since we don't have an Enabled field
	return a.config.Conventional.Required
}

// ConventionalRequired returns whether conventional commit format is required.
func (a *ValidationConfigAdapter) ConventionalRequired() bool {
	return a.config.Conventional.Required
}

// ConventionalTypes returns the list of allowed conventional commit types.
func (a *ValidationConfigAdapter) ConventionalTypes() []string {
	return a.config.Conventional.Types
}

// ConventionalScopes returns the list of allowed conventional commit scopes.
func (a *ValidationConfigAdapter) ConventionalScopes() []string {
	return a.config.Conventional.Scopes
}

// ConventionalMaxDescriptionLength returns the maximum length allowed for conventional commit descriptions.
func (a *ValidationConfigAdapter) ConventionalMaxDescriptionLength() int {
	return a.config.Conventional.MaxDescriptionLength
}

// Security configuration methods

// SignatureRequired returns whether commit signatures are required.
func (a *ValidationConfigAdapter) SignatureRequired() bool {
	return a.config.Security.SignatureRequired
}

// SignOffRequired returns whether commit sign-offs are required.
func (a *ValidationConfigAdapter) SignOffRequired() bool {
	return a.config.Security.SignOffRequired
}

// AllowedSignatureTypes returns the list of allowed signature types for commits.
func (a *ValidationConfigAdapter) AllowedSignatureTypes() []string {
	return a.config.Security.AllowedSignatureTypes
}

// AllowMultipleSignOffs returns whether multiple sign-offs are allowed in commits.
func (a *ValidationConfigAdapter) AllowMultipleSignOffs() bool {
	return a.config.Security.AllowMultipleSignOffs
}

// PublicKeyURI returns the URI for retrieving public keys for identity verification.
func (a *ValidationConfigAdapter) PublicKeyURI() string {
	return a.config.Security.Identity.PublicKeyURI
}

// IdentityPublicKeyURI returns the URI for retrieving public keys for identity verification.
func (a *ValidationConfigAdapter) IdentityPublicKeyURI() string {
	return a.config.Security.Identity.PublicKeyURI
}

// Spell check configuration methods

// SpellCheckEnabled returns whether spell checking is enabled.
func (a *ValidationConfigAdapter) SpellCheckEnabled() bool {
	return a.config.SpellCheck.Enabled
}

// SpellEnabled returns whether spell checking is enabled.
func (a *ValidationConfigAdapter) SpellEnabled() bool {
	return a.config.SpellCheck.Enabled
}

// SpellCheckLocale returns the language locale to use for spell checking.
func (a *ValidationConfigAdapter) SpellCheckLocale() string {
	return a.config.SpellCheck.Locale
}

// SpellLocale returns the language locale to use for spell checking.
func (a *ValidationConfigAdapter) SpellLocale() string {
	return a.config.SpellCheck.Locale
}

// SpellCheckIgnoreWords returns the list of words to ignore during spell checking.
func (a *ValidationConfigAdapter) SpellCheckIgnoreWords() []string {
	return a.config.SpellCheck.IgnoreWords
}

// SpellIgnoreWords returns the list of words to ignore during spell checking.
func (a *ValidationConfigAdapter) SpellIgnoreWords() []string {
	return a.config.SpellCheck.IgnoreWords
}

// SpellCheckCustomWords returns the map of custom words and their replacements for spell checking.
func (a *ValidationConfigAdapter) SpellCheckCustomWords() map[string]string {
	return a.config.SpellCheck.CustomWords
}

// SpellCustomWords returns the map of custom words and their replacements for spell checking.
func (a *ValidationConfigAdapter) SpellCustomWords() map[string]string {
	return a.config.SpellCheck.CustomWords
}

// SpellCheckMaxErrors returns the maximum number of spelling errors to report.
func (a *ValidationConfigAdapter) SpellCheckMaxErrors() int {
	return a.config.SpellCheck.MaxErrors
}

// SpellMaxErrors returns the maximum number of spelling errors to report.
func (a *ValidationConfigAdapter) SpellMaxErrors() int {
	return a.config.SpellCheck.MaxErrors
}

// Jira configuration methods

// JiraRequired returns whether JIRA ticket references are required in commits.
func (a *ValidationConfigAdapter) JiraRequired() bool {
	return a.config.Subject.Jira.Required
}

// JiraProjects returns the list of JIRA project identifiers to validate against.
func (a *ValidationConfigAdapter) JiraProjects() []string {
	return a.config.Subject.Jira.Projects
}

// JiraPattern returns the regex pattern used to match JIRA ticket references.
func (a *ValidationConfigAdapter) JiraPattern() string {
	return a.config.Subject.Jira.Pattern
}

// JiraBodyRef returns whether JIRA references in the commit body are accepted.
func (a *ValidationConfigAdapter) JiraBodyRef() bool {
	return a.config.Subject.Jira.BodyRef
}

// JiraStrict returns whether strict JIRA validation is enabled.
func (a *ValidationConfigAdapter) JiraStrict() bool {
	return a.config.Subject.Jira.Strict
}

// Repository configuration methods

// Reference returns the name of the reference branch used for comparison.
func (a *ValidationConfigAdapter) Reference() string {
	return a.config.Repository.Reference
}

// ReferenceBranch returns the name of the reference branch used for comparison.
func (a *ValidationConfigAdapter) ReferenceBranch() string {
	return a.config.Repository.Reference
}

// IgnoreMergeCommits returns whether merge commits should be ignored during validation.
func (a *ValidationConfigAdapter) IgnoreMergeCommits() bool {
	return a.config.Repository.IgnoreMergeCommits
}

// MaxCommitsAhead returns the maximum number of commits allowed ahead of the reference branch.
func (a *ValidationConfigAdapter) MaxCommitsAhead() int {
	return a.config.Repository.MaxCommitsAhead
}

// CheckCommitsAhead returns whether to validate the count of commits ahead of the reference branch.
func (a *ValidationConfigAdapter) CheckCommitsAhead() bool {
	return a.config.Repository.CheckCommitsAhead
}

// Rule configuration methods

// EnabledRules returns the list of explicitly enabled validation rules.
func (a *ValidationConfigAdapter) EnabledRules() []string {
	return a.config.Rules.EnabledRules
}

// DisabledRules returns the list of explicitly disabled validation rules.
func (a *ValidationConfigAdapter) DisabledRules() []string {
	return a.config.Rules.DisabledRules
}

// GetAvailableRules returns a list of all available rule names.
func (a *ValidationConfigAdapter) GetAvailableRules() []string {
	// This is a placeholder - actual implementation depends on the rule registry
	return []string{}
}

// GetActiveRules returns a list of currently active rule names.
func (a *ValidationConfigAdapter) GetActiveRules() []string {
	// This is a placeholder - actual implementation depends on the rule registry
	return []string{}
}

// SetEnabledRules sets the list of explicitly enabled validation rules.
// This is used by the rule engine to activate specific rules.
func (a *ValidationConfigAdapter) SetEnabledRules(ruleNames []string) {
	a.config.Rules.EnabledRules = ruleNames
}

// SetDisabledRules sets the list of explicitly disabled validation rules.
// This is used by the rule engine to deactivate specific rules.
func (a *ValidationConfigAdapter) SetDisabledRules(ruleNames []string) {
	a.config.Rules.DisabledRules = ruleNames
}

// Make sure the adapter implements all the necessary interfaces.
var (
	// Specific interfaces.
	_ domain.SubjectConfigProvider      = (*ValidationConfigAdapter)(nil)
	_ domain.BodyConfigProvider         = (*ValidationConfigAdapter)(nil)
	_ domain.ConventionalConfigProvider = (*ValidationConfigAdapter)(nil)
	_ domain.SecurityConfigProvider     = (*ValidationConfigAdapter)(nil)
	_ domain.SpellCheckConfigProvider   = (*ValidationConfigAdapter)(nil)
	_ domain.JiraConfigProvider         = (*ValidationConfigAdapter)(nil)
	_ domain.RepositoryConfigProvider   = (*ValidationConfigAdapter)(nil)
	_ domain.RuleConfigProvider         = (*ValidationConfigAdapter)(nil)
)
