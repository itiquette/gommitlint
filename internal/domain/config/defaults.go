// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

// NewDefault creates a configuration with sensible defaults.
func NewDefault() Config {
	return Config{
		Message: MessageConfig{
			Subject: SubjectConfig{
				MaxLength:         72,
				Case:              "sentence",
				RequireImperative: false,
				ForbidEndings:     []string{".", "!", "?"},
			},
			Body: BodyConfig{
				MinLength:        0,
				MinLines:         0,
				AllowSignoffOnly: true,
				RequireSignoff:   false,
			},
		},
		Conventional: ConventionalConfig{
			RequireScope:         false,
			Types:                []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"},
			Scopes:               []string{},
			AllowBreaking:        true,
			MaxDescriptionLength: 72,
		},
		Signing: SigningConfig{
			RequireSignature:    false,
			RequireVerification: false,
			RequireMultiSignoff: false,
			KeyDirectory:        "",
			AllowedSigners:      []string{},
		},
		Repo: RepoConfig{
			MaxCommitsAhead:   0, // 0 means disabled
			ReferenceBranch:   "main",
			AllowMergeCommits: true,
		},
		Jira: JiraConfig{
			ProjectPrefixes:      []string{},
			RequireInBody:        false,
			RequireInSubject:     false,
			IgnoreTicketPatterns: []string{},
		},
		Spell: SpellConfig{
			IgnoreWords: []string{},
			Locale:      "en_US",
		},
		Rules: RulesConfig{
			Enabled:  []string{},
			Disabled: []string{},
		},
		Output: "text",
	}
}

// Validate validates the configuration and returns any validation errors.
func (c Config) Validate() []string {
	var errors []string

	// Validate subject max length
	if c.Message.Subject.MaxLength <= 0 {
		errors = append(errors, "subject max_length must be positive")
	}

	// Validate conventional types
	if len(c.Conventional.Types) == 0 {
		errors = append(errors, "conventional types cannot be empty")
	}

	// Validate output format
	validOutputs := []string{"text", "json", "github", "gitlab"}
	isValidOutput := false

	for _, valid := range validOutputs {
		if c.Output == valid {
			isValidOutput = true

			break
		}
	}

	if !isValidOutput {
		errors = append(errors, "output must be one of: text, json, github, gitlab")
	}

	return errors
}
