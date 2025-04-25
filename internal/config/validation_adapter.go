// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"github.com/itiquette/gommitlint/internal/domain"
)

// ValidationConfigAdapter adapts the Config type to implement the domain.ValidationConfig interface.
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

// Subject configuration methods.

// SubjectMaxLength returns the maximum length allowed for commit subjects.
func (a *ValidationConfigAdapter) SubjectMaxLength() int {
	return a.config.Subject.MaxLength
}

// SubjectCase returns the required case style for commit subjects.
func (a *ValidationConfigAdapter) SubjectCase() string {
	return a.config.Subject.Case
}

// SubjectRequireImperative returns whether commit subjects must use imperative mood.
func (a *ValidationConfigAdapter) SubjectRequireImperative() bool {
	return a.config.Subject.Imperative
}

// SubjectInvalidSuffixes returns a string of comma-separated invalid suffixes for commit subjects.
func (a *ValidationConfigAdapter) SubjectInvalidSuffixes() string {
	return a.config.Subject.InvalidSuffixes
}

// Conventional commit configuration methods.

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

// ConventionalRequired returns whether conventional commit format is required.
func (a *ValidationConfigAdapter) ConventionalRequired() bool {
	return a.config.Conventional.Required
}

// Jira configuration methods.

// JiraProjects returns the list of JIRA project identifiers to validate against.
func (a *ValidationConfigAdapter) JiraProjects() []string {
	return a.config.Subject.Jira.Projects
}

// JiraBodyRef returns whether JIRA references in the commit body are accepted.
func (a *ValidationConfigAdapter) JiraBodyRef() bool {
	return a.config.Subject.Jira.BodyRef
}

// JiraRequired returns whether JIRA ticket references are required in commits.
func (a *ValidationConfigAdapter) JiraRequired() bool {
	return a.config.Subject.Jira.Required
}

// JiraPattern returns the regex pattern used to match JIRA ticket references.
func (a *ValidationConfigAdapter) JiraPattern() string {
	return a.config.Subject.Jira.Pattern
}

// JiraStrict returns whether strict JIRA validation is enabled.
func (a *ValidationConfigAdapter) JiraStrict() bool {
	return a.config.Subject.Jira.Strict
}

// Body configuration methods.

// BodyRequired returns whether a commit body is required.
func (a *ValidationConfigAdapter) BodyRequired() bool {
	return a.config.Body.Required
}

// BodyAllowSignOffOnly returns whether a commit body can consist of only a sign-off line.
func (a *ValidationConfigAdapter) BodyAllowSignOffOnly() bool {
	return a.config.Body.AllowSignOffOnly
}

// Security configuration methods.

// SignatureRequired returns whether commit signatures are required.
func (a *ValidationConfigAdapter) SignatureRequired() bool {
	return a.config.Security.SignatureRequired
}

// AllowedSignatureTypes returns the list of allowed signature types for commits.
func (a *ValidationConfigAdapter) AllowedSignatureTypes() []string {
	return a.config.Security.AllowedSignatureTypes
}

// SignOffRequired returns whether commit sign-offs are required.
func (a *ValidationConfigAdapter) SignOffRequired() bool {
	return a.config.Security.SignOffRequired
}

// AllowMultipleSignOffs returns whether multiple sign-offs are allowed in commits.
func (a *ValidationConfigAdapter) AllowMultipleSignOffs() bool {
	return a.config.Security.AllowMultipleSignOffs
}

// IdentityPublicKeyURI returns the URI for retrieving public keys for identity verification.
func (a *ValidationConfigAdapter) IdentityPublicKeyURI() string {
	return a.config.Security.Identity.PublicKeyURI
}

// Spell check configuration methods.

// SpellLocale returns the locale used for spell checking.
func (a *ValidationConfigAdapter) SpellLocale() string {
	return a.config.SpellCheck.Locale
}

// SpellEnabled returns whether spell checking is enabled.
func (a *ValidationConfigAdapter) SpellEnabled() bool {
	return a.config.SpellCheck.Enabled
}

// SpellIgnoreWords returns the list of words to ignore during spell checking.
func (a *ValidationConfigAdapter) SpellIgnoreWords() []string {
	return a.config.SpellCheck.IgnoreWords
}

// SpellCustomWords returns the map of custom words and their replacements for spell checking.
func (a *ValidationConfigAdapter) SpellCustomWords() map[string]string {
	return a.config.SpellCheck.CustomWords
}

// SpellMaxErrors returns the maximum number of spelling errors allowed before validation fails.
func (a *ValidationConfigAdapter) SpellMaxErrors() int {
	return a.config.SpellCheck.MaxErrors
}

// Repository configuration methods.

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

// Rule configuration methods.

// EnabledRules returns the list of explicitly enabled validation rules.
func (a *ValidationConfigAdapter) EnabledRules() []string {
	return a.config.Rules.EnabledRules
}

// DisabledRules returns the list of explicitly disabled validation rules.
func (a *ValidationConfigAdapter) DisabledRules() []string {
	return a.config.Rules.DisabledRules
}

// Make sure the adapter implements all the necessary interfaces.
var (
	_ domain.ValidationConfigProvider   = (*ValidationConfigAdapter)(nil)
	_ domain.SubjectConfigProvider      = (*ValidationConfigAdapter)(nil)
	_ domain.JiraConfigProvider         = (*ValidationConfigAdapter)(nil)
	_ domain.BodyConfigProvider         = (*ValidationConfigAdapter)(nil)
	_ domain.ConventionalConfigProvider = (*ValidationConfigAdapter)(nil)
	_ domain.SecurityConfigProvider     = (*ValidationConfigAdapter)(nil)
	_ domain.SpellCheckConfigProvider   = (*ValidationConfigAdapter)(nil)
	_ domain.RepositoryConfigProvider   = (*ValidationConfigAdapter)(nil)
	_ domain.RulesConfigProvider        = (*ValidationConfigAdapter)(nil)
)
