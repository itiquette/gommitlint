// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package types

// Config transformation methods
// =========================================================================

// WithSubject returns a new Config with the updated subject configuration.
func (c Config) WithSubject(subject SubjectConfig) Config {
	result := c
	result.Subject = subject

	return result
}

// WithBody returns a new Config with the updated body configuration.
func (c Config) WithBody(body BodyConfig) Config {
	result := c
	result.Body = body

	return result
}

// WithConventional returns a new Config with the updated conventional commit configuration.
func (c Config) WithConventional(conventional ConventionalConfig) Config {
	result := c
	result.Conventional = conventional

	return result
}

// WithRules returns a new Config with the updated rules configuration.
func (c Config) WithRules(rules RulesConfig) Config {
	result := c
	result.Rules = rules

	return result
}

// WithSecurity returns a new Config with the updated security configuration.
func (c Config) WithSecurity(security SecurityConfig) Config {
	result := c
	result.Security = security

	return result
}

// WithRepository returns a new Config with the updated repository configuration.
func (c Config) WithRepository(repository RepositoryConfig) Config {
	result := c
	result.Repository = repository

	return result
}

// WithOutput returns a new Config with the updated output configuration.
func (c Config) WithOutput(output OutputConfig) Config {
	result := c
	result.Output = output

	return result
}

// WithSpellCheck returns a new Config with the updated spellcheck configuration.
func (c Config) WithSpellCheck(spellcheck SpellCheckConfig) Config {
	result := c
	result.SpellCheck = spellcheck

	return result
}

// WithJira returns a new Config with the updated Jira configuration.
func (c Config) WithJira(jira JiraConfig) Config {
	result := c
	result.Jira = jira

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

// WithDisallowedSuffixes returns a new SubjectConfig with the updated disallowed suffixes.
func (c SubjectConfig) WithDisallowedSuffixes(suffixes []string) SubjectConfig {
	result := c
	result.DisallowedSuffixes = make([]string, len(suffixes))
	copy(result.DisallowedSuffixes, suffixes)

	return result
}

// BodyConfig transformation methods
// =========================================================================

// WithRequired returns a new BodyConfig with the updated required setting.
func (c BodyConfig) WithRequired(required bool) BodyConfig {
	result := c
	result.Required = required

	return result
}

// WithMinLength returns a new BodyConfig with the updated minimum length setting.
func (c BodyConfig) WithMinLength(minLength int) BodyConfig {
	result := c
	result.MinLength = minLength

	return result
}

// WithMinimumLines returns a new BodyConfig with the updated minimum lines setting.
func (c BodyConfig) WithMinimumLines(minLines int) BodyConfig {
	result := c
	result.MinimumLines = minLines

	return result
}

// WithAllowSignOffOnly returns a new BodyConfig with the updated sign-off only setting.
func (c BodyConfig) WithAllowSignOffOnly(allow bool) BodyConfig {
	result := c
	result.AllowSignOffOnly = allow

	return result
}

// ConventionalConfig transformation methods
// =========================================================================

// WithRequired returns a new ConventionalConfig with the updated required setting.
func (c ConventionalConfig) WithRequired(required bool) ConventionalConfig {
	result := c
	result.Required = required

	return result
}

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

// WithAllowBreakingChanges returns a new ConventionalConfig with the updated breaking changes setting.
func (c ConventionalConfig) WithAllowBreakingChanges(allow bool) ConventionalConfig {
	result := c
	result.AllowBreakingChanges = allow

	return result
}

// WithMaxDescriptionLength returns a new ConventionalConfig with the updated max description length.
func (c ConventionalConfig) WithMaxDescriptionLength(maxLength int) ConventionalConfig {
	result := c
	result.MaxDescriptionLength = maxLength

	return result
}

// RulesConfig transformation methods
// =========================================================================

// WithEnabledRules returns a new RulesConfig with the updated enabled rules.
func (c RulesConfig) WithEnabledRules(rules []string) RulesConfig {
	result := c
	result.EnabledRules = make([]string, len(rules))
	copy(result.EnabledRules, rules)

	return result
}

// WithDisabledRules returns a new RulesConfig with the updated disabled rules.
func (c RulesConfig) WithDisabledRules(rules []string) RulesConfig {
	result := c
	result.DisabledRules = make([]string, len(rules))
	copy(result.DisabledRules, rules)

	return result
}

// DisabledRules returns the list of disabled rule names from the Config.
func (c Config) DisabledRules() []string {
	return c.Rules.DisabledRules
}

// WithDisabledRules returns a new Config with the updated disabled rules.
// This is a convenience method that updates the Rules component of the Config.
func (c Config) WithDisabledRules(rules []string) Config {
	result := c
	result.Rules = result.Rules.WithDisabledRules(rules)

	return result
}

// EnableRule adds a rule to the enabled rules list and removes it from disabled rules if present.
func (c RulesConfig) EnableRule(rule string) RulesConfig {
	result := c

	// Check if rule is in the enabled list
	for _, r := range result.EnabledRules {
		if r == rule {
			return result // Already enabled
		}
	}

	// Add to enabled rules
	result.EnabledRules = append(result.EnabledRules, rule)

	// Remove from disabled rules if present
	var newDisabled []string

	for _, r := range result.DisabledRules {
		if r != rule {
			newDisabled = append(newDisabled, r)
		}
	}

	result.DisabledRules = newDisabled

	return result
}

// DisableRule adds a rule to the disabled rules list and removes it from enabled rules if present.
func (c RulesConfig) DisableRule(rule string) RulesConfig {
	result := c

	// Check if rule is in the disabled list
	for _, r := range result.DisabledRules {
		if r == rule {
			return result // Already disabled
		}
	}

	// Add to disabled rules
	result.DisabledRules = append(result.DisabledRules, rule)

	// Remove from enabled rules if present
	var newEnabled []string

	for _, r := range result.EnabledRules {
		if r != rule {
			newEnabled = append(newEnabled, r)
		}
	}

	result.EnabledRules = newEnabled

	return result
}

// SecurityConfig transformation methods
// =========================================================================

// WithSignOffRequired returns a new SecurityConfig with the updated sign-off required flag.
func (c SecurityConfig) WithSignOffRequired(required bool) SecurityConfig {
	result := c
	result.SignOffRequired = required

	return result
}

// WithGPGRequired returns a new SecurityConfig with the updated GPG required flag.
func (c SecurityConfig) WithGPGRequired(required bool) SecurityConfig {
	result := c
	result.GPGRequired = required

	return result
}

// WithKeyDirectory returns a new SecurityConfig with the updated key directory.
func (c SecurityConfig) WithKeyDirectory(dir string) SecurityConfig {
	result := c
	result.KeyDirectory = dir

	return result
}

// WithAllowedSignatureTypes returns a new SecurityConfig with the updated allowed signature types.
func (c SecurityConfig) WithAllowedSignatureTypes(types []string) SecurityConfig {
	result := c
	result.AllowedSignatureTypes = make([]string, len(types))
	copy(result.AllowedSignatureTypes, types)

	return result
}

// WithAllowedKeyrings returns a new SecurityConfig with the updated allowed keyrings.
func (c SecurityConfig) WithAllowedKeyrings(keyrings []string) SecurityConfig {
	result := c
	result.AllowedKeyrings = make([]string, len(keyrings))
	copy(result.AllowedKeyrings, keyrings)

	return result
}

// WithAllowedIdentities returns a new SecurityConfig with the updated allowed identities.
func (c SecurityConfig) WithAllowedIdentities(identities []string) SecurityConfig {
	result := c
	result.AllowedIdentities = make([]string, len(identities))
	copy(result.AllowedIdentities, identities)

	return result
}

// WithAllowMultipleSignOffs returns a new SecurityConfig with the updated multiple sign-offs setting.
func (c SecurityConfig) WithAllowMultipleSignOffs(allow bool) SecurityConfig {
	result := c
	result.AllowMultipleSignOffs = allow

	return result
}

// RepositoryConfig transformation methods
// =========================================================================

// WithPath returns a new RepositoryConfig with the updated path.
func (c RepositoryConfig) WithPath(path string) RepositoryConfig {
	result := c
	result.Path = path

	return result
}

// WithReferenceBranch returns a new RepositoryConfig with the updated reference branch.
func (c RepositoryConfig) WithReferenceBranch(branch string) RepositoryConfig {
	result := c
	result.ReferenceBranch = branch

	return result
}

// WithMaxCommitsAhead returns a new RepositoryConfig with the updated max commits ahead.
func (c RepositoryConfig) WithMaxCommitsAhead(maxCommits int) RepositoryConfig {
	result := c
	result.MaxCommitsAhead = maxCommits

	return result
}

// WithMaxHistoryDays returns a new RepositoryConfig with the updated max history days.
func (c RepositoryConfig) WithMaxHistoryDays(days int) RepositoryConfig {
	result := c
	result.MaxHistoryDays = days

	return result
}

// WithOutputFormat returns a new RepositoryConfig with the updated output format.
func (c RepositoryConfig) WithOutputFormat(format string) RepositoryConfig {
	result := c
	result.OutputFormat = format

	return result
}

// WithIgnoreMergeCommits returns a new RepositoryConfig with the updated ignore merge commits setting.
func (c RepositoryConfig) WithIgnoreMergeCommits(ignore bool) RepositoryConfig {
	result := c
	result.IgnoreMergeCommits = ignore

	return result
}

// OutputConfig transformation methods
// =========================================================================

// WithFormat returns a new OutputConfig with the updated format.
func (c OutputConfig) WithFormat(format string) OutputConfig {
	result := c
	result.Format = format

	return result
}

// WithVerbose returns a new OutputConfig with the updated verbose setting.
func (c OutputConfig) WithVerbose(verbose bool) OutputConfig {
	result := c
	result.Verbose = verbose

	return result
}

// WithQuiet returns a new OutputConfig with the updated quiet setting.
func (c OutputConfig) WithQuiet(quiet bool) OutputConfig {
	result := c
	result.Quiet = quiet

	return result
}

// WithColor returns a new OutputConfig with the updated color setting.
func (c OutputConfig) WithColor(color bool) OutputConfig {
	result := c
	result.Color = color

	return result
}

// SpellCheckConfig transformation methods
// =========================================================================

// WithEnabled returns a new SpellCheckConfig with the updated enabled setting.
func (c SpellCheckConfig) WithEnabled(enabled bool) SpellCheckConfig {
	result := c
	result.Enabled = enabled

	return result
}

// WithLanguage returns a new SpellCheckConfig with the updated language.
func (c SpellCheckConfig) WithLanguage(language string) SpellCheckConfig {
	result := c
	result.Language = language

	return result
}

// WithIgnoreCase returns a new SpellCheckConfig with the updated ignore case setting.
func (c SpellCheckConfig) WithIgnoreCase(ignore bool) SpellCheckConfig {
	result := c
	result.IgnoreCase = ignore

	return result
}

// WithCustomDictionary returns a new SpellCheckConfig with the updated custom dictionary.
func (c SpellCheckConfig) WithCustomDictionary(dictionary []string) SpellCheckConfig {
	result := c
	result.CustomDictionary = make([]string, len(dictionary))
	copy(result.CustomDictionary, dictionary)

	return result
}

// JiraConfig transformation methods
// =========================================================================

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

// WithBodyRef returns a new JiraConfig with the updated body reference setting.
func (c JiraConfig) WithBodyRef(bodyRef bool) JiraConfig {
	result := c
	result.BodyRef = bodyRef

	return result
}
