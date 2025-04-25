// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

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

// SubjectCase returns the case style required for commit subjects.
func (c Config) SubjectCase() string {
	return c.Subject.Case
}

// SubjectRequireImperative returns whether commit subjects must use imperative mood.
func (c Config) SubjectRequireImperative() bool {
	return c.Subject.Imperative
}

// SubjectInvalidSuffixes returns the suffixes that are not allowed in commit subjects.
func (c Config) SubjectInvalidSuffixes() string {
	return c.Subject.InvalidSuffixes
}

// Implementation for JiraConfigProvider.

// JiraProjects returns the list of JIRA project keys that can be referenced in commits.
func (c Config) JiraProjects() []string {
	return c.Subject.Jira.Projects
}

// JiraBodyRef returns whether JIRA references are allowed in the commit body.
func (c Config) JiraBodyRef() bool {
	return c.Subject.Jira.BodyRef
}

// JiraRequired returns whether JIRA references are required in commits.
func (c Config) JiraRequired() bool {
	return c.Subject.Jira.Required
}

// JiraPattern returns the regex pattern for matching JIRA references.
func (c Config) JiraPattern() string {
	return c.Subject.Jira.Pattern
}

// JiraStrict returns whether strict JIRA reference validation is enabled.
func (c Config) JiraStrict() bool {
	return c.Subject.Jira.Strict
}

// Implementation for BodyConfigProvider.

// BodyRequired returns whether a commit body is required.
func (c Config) BodyRequired() bool {
	return c.Body.Required
}

// BodyAllowSignOffOnly returns whether a commit body with only a sign-off is acceptable.
func (c Config) BodyAllowSignOffOnly() bool {
	return c.Body.AllowSignOffOnly
}

// Implementation for ConventionalConfigProvider.

// ConventionalTypes returns the allowed types for conventional commits.
func (c Config) ConventionalTypes() []string {
	return c.Conventional.Types
}

// ConventionalScopes returns the allowed scopes for conventional commits.
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

// AllowedSignatureTypes returns the signature types that are allowed for commit verification.
func (c Config) AllowedSignatureTypes() []string {
	return c.Security.AllowedSignatureTypes
}

// SignOffRequired returns whether commit sign-offs are required.
func (c Config) SignOffRequired() bool {
	return c.Security.SignOffRequired
}

// AllowMultipleSignOffs returns whether multiple sign-offs are allowed in a commit.
func (c Config) AllowMultipleSignOffs() bool {
	return c.Security.AllowMultipleSignOffs
}

// IdentityPublicKeyURI returns the URI for public keys used in identity verification.
func (c Config) IdentityPublicKeyURI() string {
	return c.Security.Identity.PublicKeyURI
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

// SpellCustomWords returns the custom word replacements for spell checking.
func (c Config) SpellCustomWords() map[string]string {
	return c.SpellCheck.CustomWords
}

// SpellMaxErrors returns the maximum number of spelling errors allowed.
func (c Config) SpellMaxErrors() int {
	return c.SpellCheck.MaxErrors
}

// Implementation for RepositoryConfigProvider.

// ReferenceBranch returns the branch to use as reference for commit validation.
func (c Config) ReferenceBranch() string {
	return c.Repository.Reference
}

// IgnoreMergeCommits returns whether merge commits should be ignored during validation.
func (c Config) IgnoreMergeCommits() bool {
	return c.Repository.IgnoreMergeCommits
}

// MaxCommitsAhead returns the maximum number of commits allowed ahead of the reference branch.
func (c Config) MaxCommitsAhead() int {
	return c.Repository.MaxCommitsAhead
}

// CheckCommitsAhead returns whether to check the number of commits ahead of the reference branch.
func (c Config) CheckCommitsAhead() bool {
	return c.Repository.CheckCommitsAhead
}

// Implementation for RulesConfigProvider.

// EnabledRules returns the list of enabled validation rules.
func (c Config) EnabledRules() []string {
	return c.Rules.EnabledRules
}

// DisabledRules returns the list of disabled validation rules.
func (c Config) DisabledRules() []string {
	return c.Rules.DisabledRules
}
