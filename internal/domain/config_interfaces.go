// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// SubjectConfigProvider defines configuration for commit subject validation.
type SubjectConfigProvider interface {
	// SubjectMaxLength returns the maximum length of the commit subject.
	SubjectMaxLength() int

	// SubjectCase returns the case that the first word of the description must have.
	SubjectCase() string

	// SubjectRequireImperative returns whether imperative verbs are enforced.
	SubjectRequireImperative() bool

	// SubjectInvalidSuffixes returns characters that cannot be used at the end of the subject.
	SubjectInvalidSuffixes() string
}

// JiraConfigProvider defines configuration for Jira ticket references.
type JiraConfigProvider interface {
	// JiraProjects returns the allowed Jira project keys.
	JiraProjects() []string

	// JiraBodyRef returns whether a Jira key must be present in body ref.
	JiraBodyRef() bool

	// JiraRequired returns whether a Jira key must be present.
	JiraRequired() bool

	// JiraPattern returns the regex pattern for Jira keys.
	JiraPattern() string

	// JiraStrict returns whether strict Jira validation is enabled.
	JiraStrict() bool
}

// BodyConfigProvider defines configuration for commit body validation.
type BodyConfigProvider interface {
	// BodyRequired returns whether the current commit needs a body.
	BodyRequired() bool

	// BodyAllowSignOffOnly returns whether a body with only sign-off lines is allowed.
	BodyAllowSignOffOnly() bool
}

// ConventionalConfigProvider defines configuration for conventional commit format.
type ConventionalConfigProvider interface {
	// ConventionalTypes returns the allowed types for conventional commits.
	ConventionalTypes() []string

	// ConventionalScopes returns the allowed scopes for conventional commits.
	ConventionalScopes() []string

	// ConventionalMaxDescriptionLength returns the maximum allowed length for the description.
	ConventionalMaxDescriptionLength() int

	// ConventionalRequired returns whether Conventional Commits are required.
	ConventionalRequired() bool
}

// SecurityConfigProvider defines configuration for security-related validation.
type SecurityConfigProvider interface {
	// SignatureRequired returns whether the commit needs a valid signature.
	SignatureRequired() bool

	// AllowedSignatureTypes returns the allowed signature types (gpg, ssh).
	AllowedSignatureTypes() []string

	// SignOffRequired returns whether commits need to be signed off.
	SignOffRequired() bool

	// AllowMultipleSignOffs returns whether multiple sign-offs are allowed.
	AllowMultipleSignOffs() bool

	// IdentityPublicKeyURI returns the URI for the public key to verify identities.
	IdentityPublicKeyURI() string
}

// SpellCheckConfigProvider defines configuration for spell checking.
type SpellCheckConfigProvider interface {
	// SpellLocale returns the language/locale to use for spell checking.
	SpellLocale() string

	// SpellEnabled returns whether spell checking is enabled.
	SpellEnabled() bool

	// SpellIgnoreWords returns words to ignore during spell checking.
	SpellIgnoreWords() []string

	// SpellCustomWords returns custom word mappings for spell checking.
	SpellCustomWords() map[string]string

	// SpellMaxErrors returns the maximum number of spelling errors allowed.
	SpellMaxErrors() int
}

// RepositoryConfigProvider defines configuration for repository-related validation.
type RepositoryConfigProvider interface {
	// ReferenceBranch returns the reference branch for comparison.
	ReferenceBranch() string

	// IgnoreMergeCommits returns whether merge commits should be ignored.
	IgnoreMergeCommits() bool

	// MaxCommitsAhead returns the maximum allowed commits ahead of reference branch.
	MaxCommitsAhead() int

	// CheckCommitsAhead returns whether to check the number of commits ahead.
	CheckCommitsAhead() bool
}

// RulesConfigProvider defines configuration for rule enablement/disablement.
type RulesConfigProvider interface {
	// EnabledRules returns rules that are explicitly enabled.
	EnabledRules() []string

	// DisabledRules returns rules that are explicitly disabled.
	DisabledRules() []string
}

// ValidationConfigProvider combines all configuration providers.
// Rules can choose to depend on specific providers instead of this combined interface.
type ValidationConfigProvider interface {
	SubjectConfigProvider
	JiraConfigProvider
	BodyConfigProvider
	ConventionalConfigProvider
	SecurityConfigProvider
	SpellCheckConfigProvider
	RepositoryConfigProvider
	RulesConfigProvider
}
