// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// TestConfigurationWorkflow tests different configuration scenarios.
func TestConfigurationWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		config   config.Config
		wantPass bool
	}{
		{
			name:    "Valid commit with custom length",
			message: "feat: add feature",
			config: func() config.Config {
				cfg := DefaultConfig()
				cfg.Message.Subject.MaxLength = 50

				return cfg
			}(),
			wantPass: true,
		},
		{
			name:    "Invalid - exceeds custom length",
			message: "feat: this is a very long commit message that exceeds our custom limit",
			config: func() config.Config {
				cfg := DefaultConfig()
				cfg.Message.Subject.MaxLength = 50

				return cfg
			}(),
			wantPass: false,
		},
		{
			name:    "Valid with custom conventional types",
			message: "custom: add special feature",
			config: func() config.Config {
				cfg := DefaultConfig()
				cfg.Conventional.Types = []string{"custom", "special"}

				return cfg
			}(),
			wantPass: true,
		},
		{
			name:    "Invalid - wrong conventional type",
			message: "invalid: not allowed type",
			config: func() config.Config {
				cfg := DefaultConfig()
				cfg.Conventional.Types = []string{"feat", "fix"}

				return cfg
			}(),
			wantPass: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := TestValidateMessage(t, testCase.message, testCase.config)

			if testCase.wantPass {
				require.True(t, result.Valid, "Expected validation to pass")
			} else {
				require.False(t, result.Valid, "Expected validation to fail")
			}
		})
	}
}

// TestMessageFileWorkflow tests validation of message files.
func TestMessageFileWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantPass bool
	}{
		{
			name:     "Valid commit message",
			message:  "feat: add authentication\n\nThis adds JWT authentication.",
			wantPass: true,
		},
		{
			name:     "Invalid - empty subject",
			message:  "",
			wantPass: false,
		},
		{
			name:     "Invalid - not conventional",
			message:  "adding new feature without format",
			wantPass: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := TestValidateMessage(t, testCase.message, DefaultConfig())

			if testCase.wantPass {
				require.True(t, result.Valid, "Expected validation to pass")
			} else {
				require.False(t, result.Valid, "Expected validation to fail")
			}
		})
	}
}

// TestRepositoryWorkflow tests validation against actual git repositories.
func TestRepositoryWorkflow(t *testing.T) {
	t.Run("Valid repository commit", func(t *testing.T) {
		repoPath, cleanup := GitRepo(t, "feat: add feature\n\nDetailed description")
		defer cleanup()

		result := TestValidation(t, repoPath, DefaultConfig())
		require.True(t, result.Valid)
	})

	t.Run("Invalid repository commit", func(t *testing.T) {
		repoPath, cleanup := GitRepo(t, "bad commit message")
		defer cleanup()

		result := TestValidation(t, repoPath, DefaultConfig())
		require.False(t, result.Valid)
	})
}
