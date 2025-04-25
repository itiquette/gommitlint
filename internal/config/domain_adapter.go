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
func (c Config) SubjectMaxLength() int {
	return c.Subject.MaxLength
}

func (c Config) SubjectCase() string {
	return c.Subject.Case
}

func (c Config) SubjectRequireImperative() bool {
	return c.Subject.Imperative
}

func (c Config) SubjectInvalidSuffixes() string {
	return c.Subject.InvalidSuffixes
}

// Implementation for JiraConfigProvider.
func (c Config) JiraProjects() []string {
	return c.Subject.Jira.Projects
}

func (c Config) JiraBodyRef() bool {
	return c.Subject.Jira.BodyRef
}

func (c Config) JiraRequired() bool {
	return c.Subject.Jira.Required
}

func (c Config) JiraPattern() string {
	return c.Subject.Jira.Pattern
}

func (c Config) JiraStrict() bool {
	return c.Subject.Jira.Strict
}

// Implementation for BodyConfigProvider.
func (c Config) BodyRequired() bool {
	return c.Body.Required
}

func (c Config) BodyAllowSignOffOnly() bool {
	return c.Body.AllowSignOffOnly
}

// Implementation for ConventionalConfigProvider.
func (c Config) ConventionalTypes() []string {
	return c.Conventional.Types
}

func (c Config) ConventionalScopes() []string {
	return c.Conventional.Scopes
}

func (c Config) ConventionalMaxDescriptionLength() int {
	return c.Conventional.MaxDescriptionLength
}

func (c Config) ConventionalRequired() bool {
	return c.Conventional.Required
}

// Implementation for SecurityConfigProvider.
func (c Config) SignatureRequired() bool {
	return c.Security.SignatureRequired
}

func (c Config) AllowedSignatureTypes() []string {
	return c.Security.AllowedSignatureTypes
}

func (c Config) SignOffRequired() bool {
	return c.Security.SignOffRequired
}

func (c Config) AllowMultipleSignOffs() bool {
	return c.Security.AllowMultipleSignOffs
}

func (c Config) IdentityPublicKeyURI() string {
	return c.Security.Identity.PublicKeyURI
}

// Implementation for SpellCheckConfigProvider.
func (c Config) SpellLocale() string {
	return c.SpellCheck.Locale
}

func (c Config) SpellEnabled() bool {
	return c.SpellCheck.Enabled
}

func (c Config) SpellIgnoreWords() []string {
	return c.SpellCheck.IgnoreWords
}

func (c Config) SpellCustomWords() map[string]string {
	return c.SpellCheck.CustomWords
}

func (c Config) SpellMaxErrors() int {
	return c.SpellCheck.MaxErrors
}

// Implementation for RepositoryConfigProvider.
func (c Config) ReferenceBranch() string {
	return c.Repository.Reference
}

func (c Config) IgnoreMergeCommits() bool {
	return c.Repository.IgnoreMergeCommits
}

func (c Config) MaxCommitsAhead() int {
	return c.Repository.MaxCommitsAhead
}

func (c Config) CheckCommitsAhead() bool {
	return c.Repository.CheckCommitsAhead
}

// Implementation for RulesConfigProvider.
func (c Config) EnabledRules() []string {
	return c.Rules.EnabledRules
}

func (c Config) DisabledRules() []string {
	return c.Rules.DisabledRules
}
