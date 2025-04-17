// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configuration_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple test for the default config manager.
func TestCreateDefaultConfigManager(t *testing.T) {
	// Create default config manager
	manager := configuration.CreateDefaultConfigManager()
	require.NotNil(t, manager)

	// Get configuration - this might fail if no config file is found
	_, err := manager.GetConfiguration()
	if err != nil {
		t.Logf("GetConfiguration returned error: %v", err)
	}

	// Test we can get rule configuration - this might fail if no config file is found
	_, err = manager.GetRuleConfiguration()
	if err != nil {
		t.Logf("GetRuleConfiguration returned error: %v", err)
	}
}

// Test that validator detects configuration issues.
func TestConfigValidator(t *testing.T) {
	// Create validator
	validator := configuration.NewConfigValidator()
	assert.NotNil(t, validator)

	// Test with nil config
	errs := validator.Validate(nil)
	assert.NotEmpty(t, errs, "Should detect nil config")

	// Test with valid config
	config := configuration.DefaultConfiguration()
	errs = validator.Validate(config)
	assert.Empty(t, errs, "Default config should be valid")

	// Test with invalid config
	invalidConfig := &configuration.AppConf{
		GommitConf: &configuration.GommitLintConfig{
			Subject: &configuration.SubjectRule{
				MaxLength: 0, // Invalid value
			},
		},
	}
	errs = validator.Validate(invalidConfig)
	assert.NotEmpty(t, errs, "Should detect invalid subject.maxLength")
}
