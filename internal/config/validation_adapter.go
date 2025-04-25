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
func (a *ValidationConfigAdapter) SubjectMaxLength() int {
	return a.config.Subject.MaxLength
}

func (a *ValidationConfigAdapter) SubjectCase() string {
	return a.config.Subject.Case
}

func (a *ValidationConfigAdapter) SubjectRequireImperative() bool {
	return a.config.Subject.Imperative
}

func (a *ValidationConfigAdapter) SubjectInvalidSuffixes() string {
	return a.config.Subject.InvalidSuffixes
}

// Conventional commit configuration methods.
func (a *ValidationConfigAdapter) ConventionalTypes() []string {
	return a.config.Conventional.Types
}

func (a *ValidationConfigAdapter) ConventionalScopes() []string {
	return a.config.Conventional.Scopes
}

func (a *ValidationConfigAdapter) ConventionalMaxDescriptionLength() int {
	return a.config.Conventional.MaxDescriptionLength
}

func (a *ValidationConfigAdapter) ConventionalRequired() bool {
	return a.config.Conventional.Required
}

// Jira configuration methods.
func (a *ValidationConfigAdapter) JiraProjects() []string {
	return a.config.Subject.Jira.Projects
}

func (a *ValidationConfigAdapter) JiraBodyRef() bool {
	return a.config.Subject.Jira.BodyRef
}

func (a *ValidationConfigAdapter) JiraRequired() bool {
	return a.config.Subject.Jira.Required
}

func (a *ValidationConfigAdapter) JiraPattern() string {
	return a.config.Subject.Jira.Pattern
}

func (a *ValidationConfigAdapter) JiraStrict() bool {
	return a.config.Subject.Jira.Strict
}

// Body configuration methods.
func (a *ValidationConfigAdapter) BodyRequired() bool {
	return a.config.Body.Required
}

func (a *ValidationConfigAdapter) BodyAllowSignOffOnly() bool {
	return a.config.Body.AllowSignOffOnly
}

// Security configuration methods.
func (a *ValidationConfigAdapter) SignatureRequired() bool {
	return a.config.Security.SignatureRequired
}

func (a *ValidationConfigAdapter) AllowedSignatureTypes() []string {
	return a.config.Security.AllowedSignatureTypes
}

func (a *ValidationConfigAdapter) SignOffRequired() bool {
	return a.config.Security.SignOffRequired
}

func (a *ValidationConfigAdapter) AllowMultipleSignOffs() bool {
	return a.config.Security.AllowMultipleSignOffs
}

func (a *ValidationConfigAdapter) IdentityPublicKeyURI() string {
	return a.config.Security.Identity.PublicKeyURI
}

// Spell check configuration methods.
func (a *ValidationConfigAdapter) SpellLocale() string {
	return a.config.SpellCheck.Locale
}

func (a *ValidationConfigAdapter) SpellEnabled() bool {
	return a.config.SpellCheck.Enabled
}

func (a *ValidationConfigAdapter) SpellIgnoreWords() []string {
	return a.config.SpellCheck.IgnoreWords
}

func (a *ValidationConfigAdapter) SpellCustomWords() map[string]string {
	return a.config.SpellCheck.CustomWords
}

func (a *ValidationConfigAdapter) SpellMaxErrors() int {
	return a.config.SpellCheck.MaxErrors
}

// Repository configuration methods.
func (a *ValidationConfigAdapter) ReferenceBranch() string {
	return a.config.Repository.Reference
}

func (a *ValidationConfigAdapter) IgnoreMergeCommits() bool {
	return a.config.Repository.IgnoreMergeCommits
}

func (a *ValidationConfigAdapter) MaxCommitsAhead() int {
	return a.config.Repository.MaxCommitsAhead
}

func (a *ValidationConfigAdapter) CheckCommitsAhead() bool {
	return a.config.Repository.CheckCommitsAhead
}

// Rule configuration methods.
func (a *ValidationConfigAdapter) EnabledRules() []string {
	return a.config.Rules.EnabledRules
}

func (a *ValidationConfigAdapter) DisabledRules() []string {
	return a.config.Rules.DisabledRules
}

// Make sure the adapter implements the interface.
var _ domain.RuleValidationConfig = (*ValidationConfigAdapter)(nil)
