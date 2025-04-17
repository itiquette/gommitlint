// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configuration_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidator_Validate(t *testing.T) {
	// Create validator
	validator := configuration.NewConfigValidator()

	// Test cases
	tests := []struct {
		name       string
		config     *configuration.AppConf
		expectErrs bool
	}{
		{
			name:       "nil configuration",
			config:     nil,
			expectErrs: true,
		},
		{
			name: "Invalid subject length",
			config: &configuration.AppConf{
				GommitConf: &configuration.GommitLintConfig{
					Subject: &configuration.SubjectRule{
						MaxLength: 0, // Invalid
						Case:      "lower",
					},
					Body: &configuration.BodyRule{},
				},
			},
			expectErrs: true,
		},
		{
			name: "Invalid subject case",
			config: &configuration.AppConf{
				GommitConf: &configuration.GommitLintConfig{
					Subject: &configuration.SubjectRule{
						MaxLength: 50,
						Case:      "invalid_case", // Invalid
					},
					Body: &configuration.BodyRule{},
				},
			},
			expectErrs: true,
		},
		{
			name: "Missing body configuration",
			config: &configuration.AppConf{
				GommitConf: &configuration.GommitLintConfig{
					Subject: &configuration.SubjectRule{
						MaxLength: 50,
						Case:      "lower",
					},
					// No body config
				},
			},
			expectErrs: true,
		},
		{
			name: "Invalid conventional commit config",
			config: &configuration.AppConf{
				GommitConf: &configuration.GommitLintConfig{
					Subject: &configuration.SubjectRule{
						MaxLength: 50,
						Case:      "lower",
					},
					Body: &configuration.BodyRule{},
					ConventionalCommit: &configuration.ConventionalRule{
						Required:             true,
						Types:                []string{}, // Empty types when required
						MaxDescriptionLength: 0,          // Also invalid
					},
				},
			},
			expectErrs: true,
		},
		{
			name: "Valid configuration",
			config: &configuration.AppConf{
				GommitConf: &configuration.GommitLintConfig{
					Subject: &configuration.SubjectRule{
						MaxLength: 50,
						Case:      "lower",
					},
					Body: &configuration.BodyRule{},
					ConventionalCommit: &configuration.ConventionalRule{
						Required: false,
					},
				},
			},
			expectErrs: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Validate configuration
			errs := validator.Validate(testCase.config)

			// Verify errors
			if testCase.expectErrs {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
