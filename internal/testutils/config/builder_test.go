// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/testutils/config"
	"github.com/stretchr/testify/require"
)

func TestBuilder_EnableRule(t *testing.T) {
	t.Run("add rule to enabled list", func(t *testing.T) {
		cfg := config.NewBuilder().
			EnableRule("SubjectCase").
			Build()

		require.Contains(t, cfg.Rules.Enabled, "SubjectCase")
	})

	t.Run("remove from disabled when enabling", func(t *testing.T) {
		cfg := config.NewBuilder().
			DisableRule("SubjectCase").
			EnableRule("SubjectCase").
			Build()

		require.Contains(t, cfg.Rules.Enabled, "SubjectCase")
		require.NotContains(t, cfg.Rules.Disabled, "SubjectCase")
	})
}

func TestBuilder_WithSubjectMaxLength(t *testing.T) {
	cfg := config.NewBuilder().
		WithSubjectMaxLength(100).
		Build()

	require.Equal(t, 100, cfg.Subject.MaxLength)
}

// TestCompatibility_WithJira removed since compatibility layer was deleted

func TestBuilder_Minimal(t *testing.T) {
	cfg := config.Minimal().Build()

	// Check disabled rules
	require.Contains(t, cfg.Rules.Disabled, "SubjectCase")
	require.Contains(t, cfg.Rules.Disabled, "SubjectSuffix")
	require.Contains(t, cfg.Rules.Disabled, "CommitBody")
	require.Contains(t, cfg.Rules.Disabled, "JiraReference")
	require.Contains(t, cfg.Rules.Disabled, "SignedIdentity")
}
