// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package testdata

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

// AssertValidationError checks that a validation error has expected properties.
func AssertValidationError(t *testing.T, err domain.ValidationError, expectedCode, expectedRule string) {
	t.Helper()
	require.Equal(t, expectedCode, err.Code, "error code mismatch")
	require.Equal(t, expectedRule, err.Rule, "error rule mismatch")
	require.NotEmpty(t, err.Message, "error message should not be empty")
}

// AssertRuleFailure checks that a rule failure has expected properties.
func AssertRuleFailure(t *testing.T, failure domain.ValidationError, expectedRule string) {
	t.Helper()
	require.Equal(t, expectedRule, failure.Rule, "rule mismatch")
	require.NotEmpty(t, failure.Message, "failure message should not be empty")
}

// AssertErrorContains checks that an error contains the expected substring.
func AssertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()
	require.Error(t, err)
	require.Contains(t, err.Error(), expected)
}