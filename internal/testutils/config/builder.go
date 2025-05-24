// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

// THIS PACKAGE CONTAINS TEST-ONLY CODE AND SHOULD NOT BE IMPORTED BY PRODUCTION CODE

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	internalConfig "github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
)

// Builder provides a functional builder for creating test configurations.
// It uses value semantics to ensure immutability during the build process.
type Builder struct {
	config types.Config
}

// NewBuilder creates a new Builder with default values.
func NewBuilder() Builder {
	return Builder{
		config: internalConfig.NewDefaultConfig(),
	}
}

// CreateTestContext creates a test context with configuration.
// This is the recommended approach for tests that need configuration.
//
// Parameters:
//   - t: The testing.T instance for test helper registration (can be nil for non-test usage)
//   - configModifier: A function that takes a base Config and returns a modified Config
//
// Returns:
//   - context.Context: A context with the configuration added
//
// Usage:
//
//	ctx := testconfig.CreateTestContext(t, func(c types.Config) types.Config {
//	    c.Rules.Enabled = []string{"SubjectLength"}
//	    return c
//	})
//
//nolint:thelper // This function can receive nil t, so we can't always call t.Helper()
func CreateTestContext(t *testing.T, configModifier func(types.Config) types.Config) context.Context {
	// t can be nil, so only call Helper if provided
	// This implementation to handle t.Helper() check for nil in linters
	if t != nil {
		t.Helper()
	}

	// Create base configuration
	baseConfig := internalConfig.NewDefaultConfig()

	// Apply modifier if provided
	if configModifier != nil {
		baseConfig = configModifier(baseConfig)
	}

	// Create adapter
	adapter := config.NewAdapter(baseConfig)

	// Create context
	ctx := testcontext.CreateTestContext()
	ctx = contextx.WithConfig(ctx, adapter)

	return ctx
}

// NewTestConfig provides a simpler alternative to the Builder pattern.
// It creates a config using functional options.
func NewTestConfig(options ...func(*types.Config)) types.Config {
	cfg := internalConfig.NewDefaultConfig()

	for _, option := range options {
		option(&cfg)
	}

	return cfg
}

// WithString sets a string value in the config.
func WithString(key string, value string) func(*types.Config) {
	return func(cfg *types.Config) {
		if key == "output" {
			cfg.Output = value
		} else if key == "repo.path" {
			cfg.Repo.Path = value
		} else if key == "repo.branch" {
			cfg.Repo.Branch = value
		}
	}
}

// WithMessage sets the message configuration.
func (b Builder) WithMessage(message types.MessageConfig) Builder {
	b.config.Message = message

	return b
}

// WithSubject sets the subject configuration.
func (b Builder) WithSubject(subject types.SubjectConfig) Builder {
	b.config.Message.Subject = subject

	return b
}

// WithSubjectMaxLength sets just the subject max length.
func (b Builder) WithSubjectMaxLength(maxLength int) Builder {
	b.config.Message.Subject.MaxLength = maxLength

	return b
}

// WithSubjectCase sets just the subject case style.
func (b Builder) WithSubjectCase(caseStyle string) Builder {
	b.config.Message.Subject.Case = caseStyle

	return b
}

// WithSubjectImperative sets just the subject imperative requirement.
func (b Builder) WithSubjectImperative(required bool) Builder {
	b.config.Message.Subject.RequireImperative = required

	return b
}

// WithSubjectForbidEndings sets the characters that a commit subject should not end with.
func (b Builder) WithSubjectForbidEndings(endings []string) Builder {
	b.config.Message.Subject.ForbidEndings = endings

	return b
}

// WithBody sets the body configuration.
func (b Builder) WithBody(body types.BodyConfig) Builder {
	b.config.Message.Body = body

	return b
}

// WithBodyMinLength sets just the body minimum length.
func (b Builder) WithBodyMinLength(minLength int) Builder {
	b.config.Message.Body.MinLength = minLength

	return b
}

// WithBodyMinLines sets just the body minimum lines.
func (b Builder) WithBodyMinLines(minLines int) Builder {
	b.config.Message.Body.MinLines = minLines

	return b
}

// WithBodySignOffOnly sets just the body sign-off only flag.
func (b Builder) WithBodySignOffOnly(allow bool) Builder {
	b.config.Message.Body.AllowSignoffOnly = allow

	return b
}

// WithBodyRequireSignOff sets just the body sign-off requirement.
func (b Builder) WithBodyRequireSignOff(required bool) Builder {
	b.config.Message.Body.RequireSignoff = required

	return b
}

// WithConventional sets the conventional configuration.
func (b Builder) WithConventional(conventional types.ConventionalConfig) Builder {
	b.config.Conventional = conventional

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

// WithEnable sets just the enabled rules.
func (b Builder) WithEnable(rules []string) Builder {
	b.config.Rules.Enabled = rules

	return b
}

// WithDisable sets just the disabled rules.
func (b Builder) WithDisable(rules []string) Builder {
	b.config.Rules.Disabled = rules

	return b
}

// EnableRule ads a rule to the enabled rules list.
func (b Builder) EnableRule(rule string) Builder {
	// Check if already enabled
	for _, r := range b.config.Rules.Enabled {
		if r == rule {
			return b
		}
	}

	// Ad to enabled rules
	b.config.Rules.Enabled = append(b.config.Rules.Enabled, rule)

	// Remove from disabled rules if present
	newDisabled := make([]string, 0)

	for _, r := range b.config.Rules.Disabled {
		if r != rule {
			newDisabled = append(newDisabled, r)
		}
	}

	b.config.Rules.Disabled = newDisabled

	return b
}

// DisableRule ads a rule to the disabled rules list.
func (b Builder) DisableRule(rule string) Builder {
	// Check if already disabled
	for _, r := range b.config.Rules.Disabled {
		if r == rule {
			return b
		}
	}

	// Ad to disabled rules
	b.config.Rules.Disabled = append(b.config.Rules.Disabled, rule)

	// Remove from enabled rules if present
	newEnabled := make([]string, 0)

	for _, r := range b.config.Rules.Enabled {
		if r != rule {
			newEnabled = append(newEnabled, r)
		}
	}

	b.config.Rules.Enabled = newEnabled

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

// WithJiraCheckBody configures whether to look for JIRA references in commit body.
func (b Builder) WithJiraCheckBody(checkBody bool) Builder {
	b.config.Jira.CheckBody = checkBody

	return b
}

// WithSigning sets the signing configuration.
func (b Builder) WithSigning(signing types.SigningConfig) Builder {
	b.config.Signing = signing

	return b
}

// WithSignOffRequired sets just the sign-off requirement.
func (b Builder) WithSignOffRequired(required bool) Builder {
	b.config.Message.Body.RequireSignoff = required

	return b
}

// WithRequireSignature sets the cryptographic signature requirement.
func (b Builder) WithRequireSignature(required bool) Builder {
	b.config.Signing.RequireSignature = required

	return b
}

// WithRepo sets the repository configuration.
func (b Builder) WithRepo(repo types.RepoConfig) Builder {
	b.config.Repo = repo

	return b
}

// WithRepoPath sets just the repository path.
func (b Builder) WithRepoPath(path string) Builder {
	b.config.Repo.Path = path

	return b
}

// WithRepoBranch sets just the repository branch.
func (b Builder) WithRepoBranch(branch string) Builder {
	b.config.Repo.Branch = branch

	return b
}

// WithOutput sets the output format.
func (b Builder) WithOutput(format string) Builder {
	b.config.Output = format

	return b
}

// WithSpell sets the spell check configuration.
func (b Builder) WithSpell(spell types.SpellConfig) Builder {
	b.config.Spell = spell

	return b
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
		DisableRule("CommitBody").
		DisableRule("Conventional").
		DisableRule("SubjectCase").
		DisableRule("SubjectSuffix").
		DisableRule("JiraReference").
		DisableRule("SignedIdentity")
}
