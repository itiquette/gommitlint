// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package integrationtest

import (
	"testing"

	"github.com/stretchr/testify/require"

	gitTestdata "github.com/itiquette/gommitlint/internal/adapters/git/testdata"
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
		repoPath, cleanup := gitTestdata.GitRepo(t, "feat: add feature\n\nDetailed description")
		defer cleanup()

		result := TestValidation(t, repoPath, DefaultConfig())
		require.True(t, result.Valid)
	})

	t.Run("Invalid repository commit", func(t *testing.T) {
		repoPath, cleanup := gitTestdata.GitRepo(t, "bad commit message")
		defer cleanup()

		result := TestValidation(t, repoPath, DefaultConfig())
		require.False(t, result.Valid)
	})
}

// TestConfigImmutability verifies that configuration helpers maintain immutability.
func TestConfigImmutability(t *testing.T) {
	original := DefaultConfig()
	originalLength := original.Message.Subject.MaxLength

	// Create modified config
	modified := WithSubjectMaxLength(100)

	// Original should be unchanged
	require.Equal(t, originalLength, original.Message.Subject.MaxLength)
	require.Equal(t, 100, modified.Message.Subject.MaxLength)

	// Multiple modifications should not affect each other
	config1 := WithSubjectMaxLength(50)
	config2 := WithSubjectMaxLength(80)

	require.Equal(t, 50, config1.Message.Subject.MaxLength)
	require.Equal(t, 80, config2.Message.Subject.MaxLength)
}
