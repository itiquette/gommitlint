// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
)

func TestColorFlagIntegration(t *testing.T) {
	tests := []struct {
		name        string
		colorFlag   string
		noColorEnv  string
		expectColor bool
		description string
	}{
		{
			name:        "color always enables color",
			colorFlag:   "always",
			noColorEnv:  "",
			expectColor: true,
			description: "--color always should enable color output",
		},
		{
			name:        "color never disables color",
			colorFlag:   "never",
			noColorEnv:  "",
			expectColor: false,
			description: "--color never should disable color output",
		},
		{
			name:        "NO_COLOR overrides color always",
			colorFlag:   "always",
			noColorEnv:  "1",
			expectColor: false,
			description: "NO_COLOR should override --color always",
		},
		{
			name:        "NO_COLOR overrides color auto",
			colorFlag:   "auto",
			noColorEnv:  "true",
			expectColor: false,
			description: "NO_COLOR should override --color auto",
		},
		{
			name:        "empty NO_COLOR respects never",
			colorFlag:   "never",
			noColorEnv:  "",
			expectColor: false,
			description: "empty NO_COLOR should respect --color never",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Set NO_COLOR environment variable if specified
			var cleanup func()

			if testCase.noColorEnv != "" {
				originalNoColor := os.Getenv("NO_COLOR")
				err := os.Setenv("NO_COLOR", testCase.noColorEnv)
				require.NoError(t, err)

				cleanup = func() {
					if originalNoColor == "" {
						os.Unsetenv("NO_COLOR")
					} else {
						os.Setenv("NO_COLOR", originalNoColor)
					}
				}
				defer cleanup()
			}

			// Test the color logic using OutputOptions
			var output bytes.Buffer
			options := cliAdapter.NewOutputOptions(&output).WithColor(testCase.colorFlag)

			// Call the shouldUseColor method directly to verify behavior
			result := options.ShouldUseColor()

			require.Equal(t, testCase.expectColor, result, testCase.description)
		})
	}
}

func TestColorFlagWithDifferentOutputFormats(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		colorFlag   string
		expectColor bool
		description string
	}{
		{
			name:        "text format with color always",
			format:      "text",
			colorFlag:   "always",
			expectColor: true,
			description: "text format should support color",
		},
		{
			name:        "json format with color always",
			format:      "json",
			colorFlag:   "always",
			expectColor: true,
			description: "json format should respect color setting",
		},
		{
			name:        "github format with color always",
			format:      "github",
			colorFlag:   "always",
			expectColor: true,
			description: "github format should respect color setting",
		},
		{
			name:        "text format with color never",
			format:      "text",
			colorFlag:   "never",
			expectColor: false,
			description: "text format should respect never setting",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Test the color logic using OutputOptions with different formats
			var output bytes.Buffer
			options := cliAdapter.NewOutputOptions(&output).
				WithColor(testCase.colorFlag).
				WithFormat(testCase.format)

			// The color setting should be consistent regardless of format
			// The format decides how to use color, not whether to use it
			result := options.ShouldUseColor()

			require.Equal(t, testCase.expectColor, result, testCase.description)
		})
	}
}
