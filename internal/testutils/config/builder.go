// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

// THIS PACKAGE CONTAINS TEST-ONLY CODE AND SHOULD NOT BE IMPORTED BY PRODUCTION CODE

import (
	"context"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
)

// Builder provides a functional builder for creating test configurations.
// It uses value semantics to ensure immutability during the build process.
type Builder struct {
	config types.Config
}

// NewBuilder creates a new Builder with default values.
func NewBuilder() Builder {
	return Builder{
		config: config.NewDefaultConfig(),
	}
}

// WithSubject sets the subject configuration.
func (b Builder) WithSubject(subject types.SubjectConfig) Builder {
	b.config.Subject = subject

	return b
}

// WithSubjectMaxLength sets just the subject max length.
func (b Builder) WithSubjectMaxLength(maxLength int) Builder {
	b.config.Subject.MaxLength = maxLength

	return b
}

// WithSubjectCase sets just the subject case style.
func (b Builder) WithSubjectCase(caseStyle string) Builder {
	b.config.Subject.Case = caseStyle

	return b
}

// WithSubjectImperative sets just the subject imperative requirement.
func (b Builder) WithSubjectImperative(required bool) Builder {
	b.config.Subject.RequireImperative = required

	return b
}

// WithSubjectSuffixes sets just the subject disallowed suffixes.
// NOTE: This sets "subject.invalid_suffixes" in the configuration
// to be compatible with the SubjectSuffixRule implementation.
func (b Builder) WithSubjectSuffixes(suffixes []string) Builder {
	b.config.Subject.DisallowedSuffixes = suffixes

	return b
}

// WithBody sets the body configuration.
func (b Builder) WithBody(body types.BodyConfig) Builder {
	b.config.Body = body

	return b
}

// WithBodyRequired sets just the body required flag.
func (b Builder) WithBodyRequired(required bool) Builder {
	b.config.Body.Required = required

	return b
}

// WithBodyMinLength sets just the body minimum length.
func (b Builder) WithBodyMinLength(minLength int) Builder {
	b.config.Body.MinLength = minLength

	return b
}

// WithBodyMinLines sets just the body minimum lines.
func (b Builder) WithBodyMinLines(minLines int) Builder {
	b.config.Body.MinimumLines = minLines

	return b
}

// WithBodySignOffOnly sets just the body sign-off only flag.
func (b Builder) WithBodySignOffOnly(allow bool) Builder {
	b.config.Body.AllowSignOffOnly = allow

	return b
}

// WithConventional sets the conventional configuration.
func (b Builder) WithConventional(conventional types.ConventionalConfig) Builder {
	b.config.Conventional = conventional

	return b
}

// WithConventionalRequired sets just the conventional required flag.
func (b Builder) WithConventionalRequired(required bool) Builder {
	b.config.Conventional.Required = required

	return b
}

// WithConventionalTypes sets just the conventional types.
func (b Builder) WithConventionalTypes(types []string) Builder {
	b.config.Conventional.Types = types

	return b
}

// WithConventionalScopes sets just the conventional scopes.
func (b Builder) WithConventionalScopes(scopes []string) Builder {
	b.config.Conventional.Scopes = scopes

	return b
}

// WithConventionalScopeRequired sets just the conventional scope requirement.
func (b Builder) WithConventionalScopeRequired(required bool) Builder {
	b.config.Conventional.RequireScope = required

	return b
}

// WithConventionalMaxDescLength sets just the conventional max description length.
func (b Builder) WithConventionalMaxDescLength(maxLength int) Builder {
	b.config.Conventional.MaxDescriptionLength = maxLength

	return b
}

// WithRules sets the rules configuration.
func (b Builder) WithRules(rules types.RulesConfig) Builder {
	b.config.Rules = rules

	return b
}

// WithEnabledRules sets just the enabled rules.
func (b Builder) WithEnabledRules(rules []string) Builder {
	b.config.Rules.EnabledRules = rules

	return b
}

// WithDisabledRules sets just the disabled rules.
func (b Builder) WithDisabledRules(rules []string) Builder {
	b.config.Rules.DisabledRules = rules

	return b
}

// EnableRule adds a rule to the enabled rules list.
func (b Builder) EnableRule(rule string) Builder {
	// Check if already enabled
	for _, r := range b.config.Rules.EnabledRules {
		if r == rule {
			return b
		}
	}

	// Add to enabled rules
	b.config.Rules.EnabledRules = append(b.config.Rules.EnabledRules, rule)

	// Remove from disabled rules if present
	newDisabled := make([]string, 0)

	for _, r := range b.config.Rules.DisabledRules {
		if r != rule {
			newDisabled = append(newDisabled, r)
		}
	}

	b.config.Rules.DisabledRules = newDisabled

	return b
}

// DisableRule adds a rule to the disabled rules list.
func (b Builder) DisableRule(rule string) Builder {
	// Check if already disabled
	for _, r := range b.config.Rules.DisabledRules {
		if r == rule {
			return b
		}
	}

	// Add to disabled rules
	b.config.Rules.DisabledRules = append(b.config.Rules.DisabledRules, rule)

	// Remove from enabled rules if present
	newEnabled := make([]string, 0)

	for _, r := range b.config.Rules.EnabledRules {
		if r != rule {
			newEnabled = append(newEnabled, r)
		}
	}

	b.config.Rules.EnabledRules = newEnabled

	return b
}

// WithJira sets the JIRA configuration.
func (b Builder) WithJira(jira types.JiraConfig) Builder {
	b.config.Jira = jira

	return b
}

// WithJiraPattern sets just the JIRA pattern.
func (b Builder) WithJiraPattern(pattern string) Builder {
	b.config.Jira.Pattern = pattern

	return b
}

// WithJiraProjects sets just the JIRA projects.
func (b Builder) WithJiraProjects(projects []string) Builder {
	b.config.Jira.Projects = projects

	return b
}

// WithJiraBodyRef sets just the JIRA body reference flag.
func (b Builder) WithJiraBodyRef(bodyRef bool) Builder {
	b.config.Jira.BodyRef = bodyRef

	return b
}

// WithSecurity sets the security configuration.
func (b Builder) WithSecurity(security types.SecurityConfig) Builder {
	b.config.Security = security

	return b
}

// WithSignOffRequired sets just the sign-off requirement.
func (b Builder) WithSignOffRequired(required bool) Builder {
	b.config.Security.SignOffRequired = required

	return b
}

// WithGPGRequired sets just the GPG requirement.
func (b Builder) WithGPGRequired(required bool) Builder {
	b.config.Security.GPGRequired = required

	return b
}

// WithOutput sets the output configuration.
func (b Builder) WithOutput(output types.OutputConfig) Builder {
	b.config.Output = output

	return b
}

// WithOutputFormat sets just the output format.
func (b Builder) WithOutputFormat(format string) Builder {
	b.config.Output.Format = format

	return b
}

// WithSpellCheck sets the spell check configuration.
func (b Builder) WithSpellCheck(spellcheck types.SpellCheckConfig) Builder {
	b.config.SpellCheck = spellcheck

	return b
}

// WithJiraConfig returns a new Builder with the updated JIRA configuration (alias for WithJira).
func (b Builder) WithJiraConfig(jira types.JiraConfig) Builder {
	return b.WithJira(jira)
}

// WithSubjectConfig returns a new Builder with the updated subject configuration (alias for WithSubject).
func (b Builder) WithSubjectConfig(subject types.SubjectConfig) Builder {
	return b.WithSubject(subject)
}

// WithBodyConfig returns a new Builder with the updated body configuration (alias for WithBody).
func (b Builder) WithBodyConfig(body types.BodyConfig) Builder {
	return b.WithBody(body)
}

// WithConventionalConfig returns a new Builder with the updated conventional configuration (alias for WithConventional).
func (b Builder) WithConventionalConfig(conventional types.ConventionalConfig) Builder {
	return b.WithConventional(conventional)
}

// WithSecurityConfig returns a new Builder with the updated security configuration (alias for WithSecurity).
func (b Builder) WithSecurityConfig(security types.SecurityConfig) Builder {
	return b.WithSecurity(security)
}

// WithOutputConfig returns a new Builder with the updated output configuration (alias for WithOutput).
func (b Builder) WithOutputConfig(output types.OutputConfig) Builder {
	return b.WithOutput(output)
}

// WithSpellCheckConfig returns a new Builder with the updated spell check configuration (alias for WithSpellCheck).
func (b Builder) WithSpellCheckConfig(spellcheck types.SpellCheckConfig) Builder {
	return b.WithSpellCheck(spellcheck)
}

// WithRulesConfig returns a new Builder with the updated rules configuration (alias for WithRules).
func (b Builder) WithRulesConfig(rules types.RulesConfig) Builder {
	return b.WithRules(rules)
}

// Build returns the constructed config.
func (b Builder) Build() types.Config {
	return b.config
}

// BuildContext creates a new context with the config.
func (b Builder) BuildContext(ctx context.Context) context.Context {
	return contextx.WithConfig(ctx, NewAdapter(b.config).Adapter)
}

// Default returns a new builder with default test configuration.
func Default() Builder {
	return NewBuilder()
}

// Minimal returns a new builder with minimal test configuration (most rules disabled).
func Minimal() Builder {
	return NewBuilder().
		WithBodyRequired(false).
		WithConventionalRequired(false).
		DisableRule("SubjectCase").
		DisableRule("SubjectSuffix").
		DisableRule("CommitBody").
		DisableRule("JiraReference").
		DisableRule("SignedIdentity")
}
