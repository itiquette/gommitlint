// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/config"
	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
)

func TestConfigWithDefaults_Creation(t *testing.T) {
	// Test creating config with application defaults
	cfg := config.NewConfigWithDefaults()
	require.NotNil(t, cfg)
	require.Equal(t, "sentence", cfg.Message.Subject.Case)
	require.Equal(t, 72, cfg.Message.Subject.MaxLength)

	// Verify application-specific defaults
	expectedDisabled := []string{"jirareference", "commitbody", "spell"}
	require.Equal(t, expectedDisabled, cfg.Rules.Disabled)
}

func TestLoadConfig_Integration(t *testing.T) {
	// Test loading configuration using pure function
	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify default values are loaded
	require.Equal(t, "sentence", cfg.Message.Subject.Case)
	require.Equal(t, 72, cfg.Message.Subject.MaxLength)
	require.Equal(t, "text", cfg.Output)
}

func TestMergeConfigs_Functionality(t *testing.T) {
	// Test configuration merging
	base := config.NewConfigWithDefaults()

	// Create overlay with different values
	overlay := configTypes.Config{
		Message: configTypes.MessageConfig{
			Subject: configTypes.SubjectConfig{
				MaxLength: 100,
			},
		},
		Output: "json",
	}

	// Merge configurations
	result, err := config.MergeConfigs(base, overlay)
	require.NoError(t, err)

	// Verify overlay values took precedence
	require.Equal(t, 100, result.Message.Subject.MaxLength)
	require.Equal(t, "json", result.Output)

	// Verify base values remain for unspecified fields
	require.Equal(t, "sentence", result.Message.Subject.Case)

	// Verify immutability - base config unchanged
	require.Equal(t, 72, base.Message.Subject.MaxLength)
}
