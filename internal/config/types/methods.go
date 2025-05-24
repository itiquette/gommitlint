// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package types

// Config transformation methods
// =========================================================================

// WithMessage returns a new Config with the updated message configuration.
func (c Config) WithMessage(message MessageConfig) Config {
	result := c
	result.Message = message

	return result
}

// WithConventional returns a new Config with the updated conventional commit configuration.
func (c Config) WithConventional(conventional ConventionalConfig) Config {
	result := c
	result.Conventional = conventional

	return result
}

// WithSigning returns a new Config with the updated signing configuration.
func (c Config) WithSigning(signing SigningConfig) Config {
	result := c
	result.Signing = signing

	return result
}

// WithRepo returns a new Config with the updated repository configuration.
func (c Config) WithRepo(repo RepoConfig) Config {
	result := c
	result.Repo = repo

	return result
}

// WithJira returns a new Config with the updated Jira configuration.
func (c Config) WithJira(jira JiraConfig) Config {
	result := c
	result.Jira = jira

	return result
}

// WithSpell returns a new Config with the updated spell checking configuration.
func (c Config) WithSpell(spell SpellConfig) Config {
	result := c
	result.Spell = spell

	return result
}

// WithRules returns a new Config with the updated rules configuration.
func (c Config) WithRules(rules RulesConfig) Config {
	result := c
	result.Rules = rules

	return result
}

// WithOutput returns a new Config with the updated output format.
func (c Config) WithOutput(output string) Config {
	result := c
	result.Output = output

	return result
}

// MessageConfig transformation methods
// =========================================================================

// WithSubject returns a new MessageConfig with the updated subject configuration.
func (c MessageConfig) WithSubject(subject SubjectConfig) MessageConfig {
	result := c
	result.Subject = subject

	return result
}

// WithBody returns a new MessageConfig with the updated body configuration.
func (c MessageConfig) WithBody(body BodyConfig) MessageConfig {
	result := c
	result.Body = body

	return result
}

// SubjectConfig transformation methods
// =========================================================================

// WithMaxLength returns a new SubjectConfig with the updated maximum length.
func (c SubjectConfig) WithMaxLength(maxLength int) SubjectConfig {
	result := c
	result.MaxLength = maxLength

	return result
}

// WithCase returns a new SubjectConfig with the updated case style.
func (c SubjectConfig) WithCase(caseStyle string) SubjectConfig {
	result := c
	result.Case = caseStyle

	return result
}

// WithRequireImperative returns a new SubjectConfig with the updated imperative requirement.
func (c SubjectConfig) WithRequireImperative(required bool) SubjectConfig {
	result := c
	result.RequireImperative = required

	return result
}

// WithForbidEndings returns a new SubjectConfig with the updated forbidden endings.
func (c SubjectConfig) WithForbidEndings(endings []string) SubjectConfig {
	result := c
	result.ForbidEndings = make([]string, len(endings))
	copy(result.ForbidEndings, endings)

	return result
}

// BodyConfig transformation methods
// =========================================================================

// WithMinLength returns a new BodyConfig with the updated minimum length setting.
func (c BodyConfig) WithMinLength(minLength int) BodyConfig {
	result := c
	result.MinLength = minLength

	return result
}

// WithMinLines returns a new BodyConfig with the updated minimum lines setting.
func (c BodyConfig) WithMinLines(minLines int) BodyConfig {
	result := c
	result.MinLines = minLines

	return result
}

// WithAllowSignoffOnly returns a new BodyConfig with the updated sign-off only setting.
func (c BodyConfig) WithAllowSignoffOnly(allow bool) BodyConfig {
	result := c
	result.AllowSignoffOnly = allow

	return result
}

// WithRequireSignoff returns a new BodyConfig with the updated require sign-off setting.
func (c BodyConfig) WithRequireSignoff(require bool) BodyConfig {
	result := c
	result.RequireSignoff = require

	return result
}

// ConventionalConfig transformation methods
// =========================================================================

// WithRequireScope returns a new ConventionalConfig with the updated scope requirement.
func (c ConventionalConfig) WithRequireScope(require bool) ConventionalConfig {
	result := c
	result.RequireScope = require

	return result
}

// WithTypes returns a new ConventionalConfig with the updated allowed types.
func (c ConventionalConfig) WithTypes(types []string) ConventionalConfig {
	result := c
	result.Types = make([]string, len(types))
	copy(result.Types, types)

	return result
}

// WithScopes returns a new ConventionalConfig with the updated allowed scopes.
func (c ConventionalConfig) WithScopes(scopes []string) ConventionalConfig {
	result := c
	result.Scopes = make([]string, len(scopes))
	copy(result.Scopes, scopes)

	return result
}

// WithAllowBreaking returns a new ConventionalConfig with the updated breaking changes setting.
func (c ConventionalConfig) WithAllowBreaking(allow bool) ConventionalConfig {
	result := c
	result.AllowBreaking = allow

	return result
}

// WithMaxDescriptionLength returns a new ConventionalConfig with the updated max description length.
func (c ConventionalConfig) WithMaxDescriptionLength(maxLength int) ConventionalConfig {
	result := c
	result.MaxDescriptionLength = maxLength

	return result
}

// SigningConfig transformation methods
// =========================================================================

// WithRequireSignature returns a new SigningConfig with the updated signature required flag.
func (c SigningConfig) WithRequireSignature(required bool) SigningConfig {
	result := c
	result.RequireSignature = required

	return result
}

// WithAllowMultipleSignoffs returns a new SigningConfig with the updated multiple sign-offs setting.
func (c SigningConfig) WithAllowMultipleSignoffs(allow bool) SigningConfig {
	result := c
	result.AllowMultipleSignoffs = allow

	return result
}

// WithAllowedSigners returns a new SigningConfig with the updated allowed signers.
func (c SigningConfig) WithAllowedSigners(signers []string) SigningConfig {
	result := c
	result.AllowedSigners = make([]string, len(signers))
	copy(result.AllowedSigners, signers)

	return result
}

// RepoConfig transformation methods
// =========================================================================

// WithPath returns a new RepoConfig with the updated path.
func (c RepoConfig) WithPath(path string) RepoConfig {
	result := c
	result.Path = path

	return result
}

// WithBranch returns a new RepoConfig with the updated branch.
func (c RepoConfig) WithBranch(branch string) RepoConfig {
	result := c
	result.Branch = branch

	return result
}

// WithMaxCommitsAhead returns a new RepoConfig with the updated max commits ahead.
func (c RepoConfig) WithMaxCommitsAhead(maxCommits int) RepoConfig {
	result := c
	result.MaxCommitsAhead = maxCommits

	return result
}

// WithIgnoreMerges returns a new RepoConfig with the updated ignore merges setting.
func (c RepoConfig) WithIgnoreMerges(ignore bool) RepoConfig {
	result := c
	result.IgnoreMerges = ignore

	return result
}

// JiraConfig transformation methods
// =========================================================================

// WithPattern returns a new JiraConfig with the updated pattern.
func (c JiraConfig) WithPattern(pattern string) JiraConfig {
	result := c
	result.Pattern = pattern

	return result
}

// WithProjects returns a new JiraConfig with the updated projects.
func (c JiraConfig) WithProjects(projects []string) JiraConfig {
	result := c
	result.Projects = make([]string, len(projects))
	copy(result.Projects, projects)

	return result
}

// WithCheckBody returns a new JiraConfig with the updated check body setting.
func (c JiraConfig) WithCheckBody(checkBody bool) JiraConfig {
	result := c
	result.CheckBody = checkBody

	return result
}

// SpellConfig transformation methods
// =========================================================================

// WithLanguage returns a new SpellConfig with the updated language.
func (c SpellConfig) WithLanguage(language string) SpellConfig {
	result := c
	result.Language = language

	return result
}

// WithIgnoreWords returns a new SpellConfig with the updated ignore words.
func (c SpellConfig) WithIgnoreWords(words []string) SpellConfig {
	result := c
	result.IgnoreWords = make([]string, len(words))
	copy(result.IgnoreWords, words)

	return result
}

// RulesConfig transformation methods
// =========================================================================

// WithEnable returns a new RulesConfig with the updated enabled rules.
func (c RulesConfig) WithEnable(rules []string) RulesConfig {
	result := c
	result.Enabled = make([]string, len(rules))
	copy(result.Enabled, rules)

	return result
}

// WithDisable returns a new RulesConfig with the updated disabled rules.
func (c RulesConfig) WithDisable(rules []string) RulesConfig {
	result := c
	result.Disabled = make([]string, len(rules))
	copy(result.Disabled, rules)

	return result
}

// EnableRule ads a rule to the enabled rules list and removes it from disabled rules if present.
func (c RulesConfig) EnableRule(rule string) RulesConfig {
	result := c

	// Check if rule is in the enabled list
	for _, r := range result.Enabled {
		if r == rule {
			return result // Already enabled
		}
	}

	// Ad to enabled rules
	result.Enabled = append(result.Enabled, rule)

	// Remove from disabled rules if present
	var newDisabled []string

	for _, r := range result.Disabled {
		if r != rule {
			newDisabled = append(newDisabled, r)
		}
	}

	result.Disabled = newDisabled

	return result
}

// DisableRule ads a rule to the disabled rules list and removes it from enabled rules if present.
func (c RulesConfig) DisableRule(rule string) RulesConfig {
	result := c

	// Check if rule is in the disabled list
	for _, r := range result.Disabled {
		if r == rule {
			return result // Already disabled
		}
	}

	// Ad to disabled rules
	result.Disabled = append(result.Disabled, rule)

	// Remove from enabled rules if present
	var newEnabled []string

	for _, r := range result.Enabled {
		if r != rule {
			newEnabled = append(newEnabled, r)
		}
	}

	result.Enabled = newEnabled

	return result
}

// Convenience methods for commonly accessed fields
// =========================================================================

// Disable returns the list of disabled rule names from the Config.
func (c Config) Disable() []string {
	return c.Rules.Disabled
}

// WithDisable returns a new Config with the updated disabled rules.
// This is a convenience method that updates the Rules component of the Config.
func (c Config) WithDisable(rules []string) Config {
	result := c
	result.Rules = result.Rules.WithDisable(rules)

	return result
}
