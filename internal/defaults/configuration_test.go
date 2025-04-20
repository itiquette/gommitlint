// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package defaults

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultValues(t *testing.T) {
	// Ensure default values are as expected
	require.Equal(t, "lower", SubjectCaseDefault)
	require.True(t, SubjectImperativeDefault)
	require.Equal(t, ".", SubjectInvalidSuffixesDefault)
	require.Equal(t, 100, SubjectMaxLengthDefault)

	require.False(t, BodyRequiredDefault)

	// Array defaults
	expectedTypes := []string{
		"feat", "fix", "docs", "style", "refactor",
		"perf", "test", "build", "ci", "chore", "revert",
	}
	require.Equal(t, expectedTypes, ConventionalCommitTypesDefault)
	require.Equal(t, 72, ConventionalCommitMaxDescLengthDefault)
	require.True(t, ConventionalCommitRequiredDefault)

	// Signature defaults
	require.True(t, SignatureRequiredDefault)
	require.True(t, SignOffRequiredDefault)

	// Git branch defaults
	require.Equal(t, "main", DefaultMainBranch)
	require.Equal(t, "master", DefaultMasterBranch)

	// Spell checking defaults
	require.Equal(t, "UK", SpellcheckLocaleDefault)

	// JIRA defaults
	require.Equal(t, `[A-Z]+-\d+`, JIRAPatternDefault)
	require.False(t, JIRARequiredDefault)

	// Miscellaneous defaults
	require.True(t, NCommitsAheadDefault)
	require.Equal(t, 5, NCommitsAheadLimitDefault)
	require.True(t, IgnoreMergeCommitsDefault)
}
