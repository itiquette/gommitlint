// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package defaults

import (
	"github.com/itiquette/gommitlint/internal/domain"
)

// Default configuration file names and paths.
const (
	ConfigFileName = ".gommitlint.yaml"
	XDGConfigPath  = "/gommitlint/gommitlint.yaml"
)

// Subject rule defaults.
const (
	SubjectCaseDefault            = "lower"
	SubjectImperativeDefault      = true
	SubjectInvalidSuffixesDefault = "."
	SubjectMaxLengthDefault       = 100
)

// Body rule defaults.
const (
	BodyRequiredDefault = false
)

// Conventional commit defaults.
var (
	ConventionalCommitTypesDefault = []string{
		"feat", "fix", "docs", "style", "refactor",
		"perf", "test", "build", "ci", "chore", "revert",
	}
	ConventionalCommitMaxDescLengthDefault = 72
	ConventionalCommitRequiredDefault      = true
)

// Signature defaults.
const (
	SignatureRequiredDefault = true
	SignOffRequiredDefault   = true
)

// Git branch defaults.
const (
	DefaultMainBranch   = "main"
	DefaultMasterBranch = "master"
)

// Spell checking defaults.
const (
	SpellcheckLocaleDefault  = "UK"
	SpellcheckEnabledDefault = true
)

// JIRA defaults.
const (
	JIRAPatternDefault  = `[A-Z]+-\d+`
	JIRARequiredDefault = false
)

// Miscellaneous defaults.
const (
	NCommitsAheadDefault      = true
	NCommitsAheadLimitDefault = 5
	IgnoreMergeCommitsDefault = true
)

// DefaultConfig provides a simple configuration implementation with default values.
type DefaultConfig struct {
	// rule lists
	enabledRules  []string
	disabledRules []string
}

// NewDefaultConfig creates a new default configuration.
func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{}
}

// SubjectCase returns the default subject case.
func (c *DefaultConfig) SubjectCase() string {
	return SubjectCaseDefault
}

// SubjectMaxLength returns the default subject maximum length.
func (c *DefaultConfig) SubjectMaxLength() int {
	return SubjectMaxLengthDefault
}

// SubjectImperative returns whether subjects should be imperative.
func (c *DefaultConfig) SubjectImperative() bool {
	return SubjectImperativeDefault
}

// SubjectRequireImperative returns whether imperative verbs are required.
func (c *DefaultConfig) SubjectRequireImperative() bool {
	return SubjectImperativeDefault
}

// SubjectInvalidSuffixes returns the default invalid suffixes.
func (c *DefaultConfig) SubjectInvalidSuffixes() string {
	return SubjectInvalidSuffixesDefault
}

// BodyRequired returns whether commit bodies are required.
func (c *DefaultConfig) BodyRequired() bool {
	return BodyRequiredDefault
}

// BodyAllowSignOffOnly returns whether bodies with only signoff are allowed.
func (c *DefaultConfig) BodyAllowSignOffOnly() bool {
	return true // Default to true
}

// ConventionalEnabled returns whether conventional commits are enabled.
func (c *DefaultConfig) ConventionalEnabled() bool {
	return true // Default to true
}

// ConventionalRequired returns whether conventional commits are required.
func (c *DefaultConfig) ConventionalRequired() bool {
	return ConventionalCommitRequiredDefault
}

// ConventionalTypes returns the default conventional commit types.
func (c *DefaultConfig) ConventionalTypes() []string {
	return ConventionalCommitTypesDefault
}

// ConventionalScopes returns the default conventional commit scopes.
func (c *DefaultConfig) ConventionalScopes() []string {
	return nil // Default to nil (any scope allowed)
}

// ConventionalMaxDescriptionLength returns the default max description length.
func (c *DefaultConfig) ConventionalMaxDescriptionLength() int {
	return ConventionalCommitMaxDescLengthDefault
}

// JiraRequired returns whether JIRA references are required.
func (c *DefaultConfig) JiraRequired() bool {
	return JIRARequiredDefault
}

// JiraProjects returns the allowed JIRA projects.
func (c *DefaultConfig) JiraProjects() []string {
	return nil // Default to nil (any project allowed)
}

// JiraPattern returns the default JIRA pattern.
func (c *DefaultConfig) JiraPattern() string {
	return JIRAPatternDefault
}

// JiraBodyRef returns whether JIRA references can be in the body.
func (c *DefaultConfig) JiraBodyRef() bool {
	return true // Default to true
}

// JiraStrict returns whether JIRA validation is strict.
func (c *DefaultConfig) JiraStrict() bool {
	return false // Default to false
}

// SignatureRequired returns whether signatures are required.
func (c *DefaultConfig) SignatureRequired() bool {
	return SignatureRequiredDefault
}

// SignOffRequired returns whether signoffs are required.
func (c *DefaultConfig) SignOffRequired() bool {
	return SignOffRequiredDefault
}

// AllowedSignatureTypes returns the allowed signature types.
func (c *DefaultConfig) AllowedSignatureTypes() []string {
	return []string{"gpg", "ssh"} // Default to both
}

// AllowMultipleSignOffs returns whether multiple signoffs are allowed.
func (c *DefaultConfig) AllowMultipleSignOffs() bool {
	return true // Default to true
}

// IdentityPublicKeyURI returns the URI for the public key used for signature verification.
func (c *DefaultConfig) IdentityPublicKeyURI() string {
	return "" // Default to empty string (no specific URI)
}

// SpellEnabled returns whether spell checking is enabled.
func (c *DefaultConfig) SpellEnabled() bool {
	return SpellcheckEnabledDefault
}

// SpellLocale returns the default spell checking locale.
func (c *DefaultConfig) SpellLocale() string {
	return SpellcheckLocaleDefault
}

// SpellIgnoreWords returns the words to ignore in spell checking.
func (c *DefaultConfig) SpellIgnoreWords() []string {
	return nil // Default to nil
}

// SpellCustomWords returns the custom words to allow in spell checking.
func (c *DefaultConfig) SpellCustomWords() map[string]string {
	return nil // Default to nil
}

// SpellMaxErrors returns the maximum number of spelling errors allowed.
func (c *DefaultConfig) SpellMaxErrors() int {
	return 0 // Default to 0
}

// ReferenceBranch returns the default reference branch.
func (c *DefaultConfig) ReferenceBranch() string {
	return DefaultMainBranch
}

// IgnoreMergeCommits returns whether merge commits should be ignored.
func (c *DefaultConfig) IgnoreMergeCommits() bool {
	return IgnoreMergeCommitsDefault
}

// CheckCommitsAhead returns whether to check commits ahead of reference.
func (c *DefaultConfig) CheckCommitsAhead() bool {
	return NCommitsAheadDefault
}

// MaxCommitsAhead returns the maximum number of commits ahead allowed.
func (c *DefaultConfig) MaxCommitsAhead() int {
	return NCommitsAheadLimitDefault
}

// EnabledRules returns the list of explicitly enabled rules.
func (c *DefaultConfig) EnabledRules() []string {
	return c.enabledRules // Default to nil (all rules enabled)
}

// DisabledRules returns the list of disabled rules.
func (c *DefaultConfig) DisabledRules() []string {
	return c.disabledRules // Default to nil (no rules disabled)
}

// SetEnabledRules sets the list of enabled rule names.
func (c *DefaultConfig) SetEnabledRules(rules []string) {
	c.enabledRules = rules
}

// SetDisabledRules sets the list of disabled rule names.
func (c *DefaultConfig) SetDisabledRules(rules []string) {
	c.disabledRules = rules
}

// GetAvailableRules returns a list of all available rule names.
func (c *DefaultConfig) GetAvailableRules() []string {
	return []string{
		"SubjectLength",
		"SubjectCase",
		"SubjectSuffix",
		"CommitBody",
		"ConventionalCommit",
		"ImperativeVerb",
		"Signature",
		"SignOff",
		"Spell",
		"JiraReference",
		"CommitsAhead",
	}
}

// GetActiveRules returns a list of currently active rule names.
func (c *DefaultConfig) GetActiveRules() []string {
	enabled := c.EnabledRules()
	if len(enabled) > 0 {
		return enabled
	}

	// If no rules are explicitly enabled, return all available rules
	// except those explicitly disabled
	available := c.GetAvailableRules()
	disabled := c.DisabledRules()

	// If there are no disabled rules, return all available rules
	if len(disabled) == 0 {
		return available
	}

	// Otherwise, filter out disabled rules
	disabledMap := make(map[string]bool)
	for _, name := range disabled {
		disabledMap[name] = true
	}

	active := make([]string, 0, len(available))

	for _, name := range available {
		if !disabledMap[name] {
			active = append(active, name)
		}
	}

	return active
}

// GetSubjectConfig returns configuration for subject validation.
func (c *DefaultConfig) GetSubjectConfig() domain.SubjectValidationConfig {
	return domain.SubjectValidationConfig{
		Case:            c.SubjectCase(),
		Imperative:      c.SubjectImperative(),
		InvalidSuffixes: c.SubjectInvalidSuffixes(),
		MaxLength:       c.SubjectMaxLength(),
		Jira: domain.JiraValidationConfig{
			Pattern:  c.JiraPattern(),
			Required: c.JiraRequired(),
			Projects: c.JiraProjects(),
			BodyRef:  c.JiraBodyRef(),
			Strict:   c.JiraStrict(),
		},
	}
}

// GetBodyConfig returns configuration for body validation.
func (c *DefaultConfig) GetBodyConfig() domain.BodyValidationConfig {
	return domain.BodyValidationConfig{
		Required:         c.BodyRequired(),
		AllowSignOffOnly: c.BodyAllowSignOffOnly(),
	}
}

// GetConventionalConfig returns configuration for conventional commit validation.
func (c *DefaultConfig) GetConventionalConfig() domain.ConventionalValidationConfig {
	return domain.ConventionalValidationConfig{
		Required:             c.ConventionalRequired(),
		Types:                c.ConventionalTypes(),
		Scopes:               c.ConventionalScopes(),
		MaxDescriptionLength: c.ConventionalMaxDescriptionLength(),
	}
}

// GetSpellCheckConfig returns configuration for spell checking.
func (c *DefaultConfig) GetSpellCheckConfig() domain.SpellCheckValidationConfig {
	return domain.SpellCheckValidationConfig{
		Enabled:     c.SpellEnabled(),
		Locale:      c.SpellLocale(),
		IgnoreWords: c.SpellIgnoreWords(),
		CustomWords: c.SpellCustomWords(),
		MaxErrors:   c.SpellMaxErrors(),
	}
}

// GetSecurityConfig returns configuration for security validation.
func (c *DefaultConfig) GetSecurityConfig() domain.SecurityValidationConfig {
	return domain.SecurityValidationConfig{
		SignatureRequired:     c.SignatureRequired(),
		SignOffRequired:       c.SignOffRequired(),
		AllowedSignatureTypes: c.AllowedSignatureTypes(),
		AllowMultipleSignOffs: c.AllowMultipleSignOffs(),
		Identity: domain.IdentityValidationConfig{
			PublicKeyURI: c.IdentityPublicKeyURI(),
		},
	}
}

// GetRepositoryConfig returns configuration for repository validation.
func (c *DefaultConfig) GetRepositoryConfig() domain.RepositoryValidationConfig {
	return domain.RepositoryValidationConfig{
		Reference:          c.ReferenceBranch(),
		IgnoreMergeCommits: c.IgnoreMergeCommits(),
		CheckCommitsAhead:  c.CheckCommitsAhead(),
		MaxCommitsAhead:    c.MaxCommitsAhead(),
	}
}

// GetRulesConfig returns configuration for rule enablement.
func (c *DefaultConfig) GetRulesConfig() domain.RulesValidationConfig {
	return domain.RulesValidationConfig{
		EnabledRules:  c.EnabledRules(),
		DisabledRules: c.DisabledRules(),
	}
}
