// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package defaults

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
