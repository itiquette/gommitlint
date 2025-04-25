// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import "github.com/itiquette/gommitlint/internal/domain"

// Ensure Config implements all the config provider interfaces.
var (
	_ domain.ValidationConfigProvider   = (*Config)(nil)
	_ domain.SubjectConfigProvider      = (*Config)(nil)
	_ domain.JiraConfigProvider         = (*Config)(nil)
	_ domain.BodyConfigProvider         = (*Config)(nil)
	_ domain.ConventionalConfigProvider = (*Config)(nil)
	_ domain.SecurityConfigProvider     = (*Config)(nil)
	_ domain.SpellCheckConfigProvider   = (*Config)(nil)
	_ domain.RepositoryConfigProvider   = (*Config)(nil)
	_ domain.RulesConfigProvider        = (*Config)(nil)
)

// Implementation for SubjectConfigProvider.

// SubjectMaxLength returns the maximum length allowed for commit subjects.
func (c Config) SubjectMaxLength() int {
	return c.Subject.MaxLength
}

// SubjectCase returns the required case style for commit subjects.
func (c Config) SubjectCase() string {
	return c.Subject.Case
}

// SubjectRequireImperative returns whether commit subjects must use imperative mood.
func (c Config) SubjectRequireImperative() bool {
	return c.Subject.Imperative
}

// SubjectInvalidSuffixes returns a string of comma-separated invalid suffixes for commit subjects.
func (c Config) SubjectInvalidSuffixes() string {
	return c.Subject.InvalidSuffixes
}

// Implementation for JiraConfigProvider.

// JiraProjects returns the list of JIRA project identifiers to validate against.
func (c Config) JiraProjects() []string {
	return c.Subject.Jira.Projects
}

// JiraBodyRef returns whether JIRA references in the commit body are accepted.
func (c Config) JiraBodyRef() bool {
	return c.Subject.Jira.BodyRef
}

// JiraRequired returns whether JIRA ticket references are required in commits.
func (c Config) JiraRequired() bool {
	return c.Subject.Jira.Required
}

// JiraPattern returns the regex pattern used to match JIRA ticket references.
func (c Config) JiraPattern() string {
	return c.Subject.Jira.Pattern
}

// JiraStrict returns whether strict JIRA validation is enabled.
func (c Config) JiraStrict() bool {
	// Not part of the Validation package's JiraConfig, so return a default
	return false
}

// Implementation for BodyConfigProvider.

// BodyRequired returns whether a commit body is required.
func (c Config) BodyRequired() bool {
	return c.Body.Required
}

// BodyAllowSignOffOnly returns whether a commit body can consist of only a sign-off line.
func (c Config) BodyAllowSignOffOnly() bool {
	return c.Body.AllowSignOffOnly
}

// Implementation for ConventionalConfigProvider.

// ConventionalTypes returns the list of allowed conventional commit types.
func (c Config) ConventionalTypes() []string {
	return c.Conventional.Types
}

// ConventionalScopes returns the list of allowed conventional commit scopes.
func (c Config) ConventionalScopes() []string {
	return c.Conventional.Scopes
}

// ConventionalMaxDescriptionLength returns the maximum length allowed for conventional commit descriptions.
func (c Config) ConventionalMaxDescriptionLength() int {
	return c.Conventional.MaxDescriptionLength
}

// ConventionalRequired returns whether conventional commit format is required.
func (c Config) ConventionalRequired() bool {
	return c.Conventional.Required
}

// Implementation for SecurityConfigProvider.

// SignatureRequired returns whether commit signatures are required.
func (c Config) SignatureRequired() bool {
	return c.Security.SignatureRequired
}

// AllowedSignatureTypes returns the list of allowed signature types for commits.
func (c Config) AllowedSignatureTypes() []string {
	return c.Security.AllowedSignatureTypes
}

// SignOffRequired returns whether commit sign-offs are required.
func (c Config) SignOffRequired() bool {
	return c.Security.SignOffRequired
}

// AllowMultipleSignOffs returns whether multiple sign-offs are allowed in commits.
func (c Config) AllowMultipleSignOffs() bool {
	return c.Security.AllowMultipleSignOffs
}

// IdentityPublicKeyURI returns the URI for retrieving public keys for identity verification.
func (c Config) IdentityPublicKeyURI() string {
	// Not directly exposed in our config, use a default
	return ""
}

// Implementation for SpellCheckConfigProvider.

// SpellLocale returns the locale used for spell checking.
func (c Config) SpellLocale() string {
	return c.SpellCheck.Locale
}

// SpellEnabled returns whether spell checking is enabled.
func (c Config) SpellEnabled() bool {
	return c.SpellCheck.Enabled
}

// SpellIgnoreWords returns the list of words to ignore during spell checking.
func (c Config) SpellIgnoreWords() []string {
	return c.SpellCheck.IgnoreWords
}

// SpellCustomWords returns the map of custom words and their replacements for spell checking.
func (c Config) SpellCustomWords() map[string]string {
	return c.SpellCheck.CustomWords
}

// SpellMaxErrors returns the maximum number of spelling errors allowed before validation fails.
func (c Config) SpellMaxErrors() int {
	return c.SpellCheck.MaxErrors
}

// Implementation for RepositoryConfigProvider.

// ReferenceBranch returns the name of the reference branch used for comparison.
func (c Config) ReferenceBranch() string {
	return c.Repository.Reference
}

// IgnoreMergeCommits returns whether merge commits should be ignored during validation.
func (c Config) IgnoreMergeCommits() bool {
	// Match the behavior in the service where it filters merge commits if requested
	return true
}

// MaxCommitsAhead returns the maximum number of commits allowed ahead of the reference branch.
func (c Config) MaxCommitsAhead() int {
	return c.Repository.MaxCommitsAhead
}

// CheckCommitsAhead returns whether to validate the count of commits ahead of the reference branch.
func (c Config) CheckCommitsAhead() bool {
	// Assume that if MaxCommitsAhead > 0, we want to check
	return c.Repository.MaxCommitsAhead > 0
}

// Implementation for RulesConfigProvider.

// EnabledRules returns the list of explicitly enabled validation rules.
func (c Config) EnabledRules() []string {
	return c.Rules.EnabledRules
}

// DisabledRules returns the list of explicitly disabled validation rules.
func (c Config) DisabledRules() []string {
	return c.Rules.DisabledRules
}
