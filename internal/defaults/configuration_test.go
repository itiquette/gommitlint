// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package defaults

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultValues(t *testing.T) {
	// Ensure default values are as expected
	assert.Equal(t, "lower", SubjectCaseDefault)
	assert.True(t, SubjectImperativeDefault)
	assert.Equal(t, ".", SubjectInvalidSuffixesDefault)
	assert.Equal(t, 100, SubjectMaxLengthDefault)

	assert.False(t, BodyRequiredDefault)

	// Array defaults
	expectedTypes := []string{
		"feat", "fix", "docs", "style", "refactor",
		"perf", "test", "build", "ci", "chore", "revert",
	}
	assert.Equal(t, expectedTypes, ConventionalCommitTypesDefault)
	assert.Equal(t, 72, ConventionalCommitMaxDescLengthDefault)
	assert.True(t, ConventionalCommitRequiredDefault)

	// Signature defaults
	assert.True(t, SignatureRequiredDefault)
	assert.True(t, SignOffRequiredDefault)

	// Git branch defaults
	assert.Equal(t, "main", DefaultMainBranch)
	assert.Equal(t, "master", DefaultMasterBranch)

	// Spell checking defaults
	assert.Equal(t, "UK", SpellcheckLocaleDefault)

	// JIRA defaults
	assert.Equal(t, `[A-Z]+-\d+`, JIRAPatternDefault)
	assert.False(t, JIRARequiredDefault)

	// Miscellaneous defaults
	assert.True(t, NCommitsAheadDefault)
	assert.Equal(t, 5, NCommitsAheadLimitDefault)
	assert.True(t, IgnoreMergeCommitsDefault)
}
