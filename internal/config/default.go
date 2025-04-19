// Package config provides configuration management for gommitlint.
package config

import (
	"os"
	"path/filepath"

	"github.com/itiquette/gommitlint/internal/defaults"
)

// DefaultConfiguration returns a default configuration.
func DefaultConfiguration() *AppConf {
	// Initialize with defaults
	imperativeVal := defaults.SubjectImperativeDefault
	signOff := defaults.SignOffRequiredDefault
	jiraRequired := defaults.JIRARequiredDefault
	conventional := defaults.ConventionalCommitRequiredDefault
	ignoreCommits := defaults.IgnoreMergeCommitsDefault
	nCommitsAhead := defaults.NCommitsAheadDefault

	return &AppConf{
		GommitConf: &GommitLintConfig{
			Subject: &SubjectRule{
				Case:            defaults.SubjectCaseDefault,
				Imperative:      &imperativeVal,
				InvalidSuffixes: defaults.SubjectInvalidSuffixesDefault,
				MaxLength:       defaults.SubjectMaxLengthDefault,
				Jira: &JiraRule{
					Required: jiraRequired,
					Pattern:  defaults.JIRAPatternDefault,
				},
			},
			Body: &BodyRule{
				Required: defaults.BodyRequiredDefault,
			},
			ConventionalCommit: &ConventionalRule{
				Types:                defaults.ConventionalCommitTypesDefault,
				MaxDescriptionLength: defaults.ConventionalCommitMaxDescLengthDefault,
				Required:             conventional,
			},
			SpellCheck: &SpellingRule{
				Locale:  defaults.SpellcheckLocaleDefault,
				Enabled: defaults.SpellcheckEnabledDefault,
			},
			Signature: &SignatureRule{
				Required: defaults.SignatureRequiredDefault,
			},
			SignOffRequired:    &signOff,
			NCommitsAhead:      &nCommitsAhead,
			IgnoreMergeCommits: &ignoreCommits,
			Reference:          "main",
		},
	}
}

// getDefaultConfigPaths returns the default configuration search paths.
// Paths are ordered by precedence (highest precedence first):
// 1. Project-level .gommitlint.yaml in the current directory
// 2. XDG config home
func getDefaultConfigPaths() []string {
	paths := []string{}

	// Add project-level configuration (highest precedence)
	paths = append(paths, ".gommitlint.yaml")

	// Add XDG config home
	var xdgConfigPath string
	if xdgHome, exists := os.LookupEnv("XDG_CONFIG_HOME"); exists && xdgHome != "" {
		xdgConfigPath = filepath.Join(xdgHome, "gommitlint", "gommitlint.yaml")
		paths = append(paths, xdgConfigPath)
	} else {
		// Default XDG config path
		home, err := os.UserHomeDir()
		if err == nil {
			xdgConfigPath = filepath.Join(home, ".config", "gommitlint", "gommitlint.yaml")
			paths = append(paths, xdgConfigPath)
		}
	}

	return paths
}
