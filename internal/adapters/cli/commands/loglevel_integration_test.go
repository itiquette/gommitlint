// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	logadapter "github.com/itiquette/gommitlint/internal/adapters/logging"
)

func TestLogLevelFlagIntegration(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		expectError bool
		description string
	}{
		{
			name:        "error log level",
			logLevel:    "error",
			expectError: false,
			description: "should accept error log level",
		},
		{
			name:        "warn log level",
			logLevel:    "warn",
			expectError: false,
			description: "should accept warn log level",
		},
		{
			name:        "info log level",
			logLevel:    "info",
			expectError: false,
			description: "should accept info log level (default)",
		},
		{
			name:        "debug log level",
			logLevel:    "debug",
			expectError: false,
			description: "should accept debug log level",
		},
		{
			name:        "trace log level",
			logLevel:    "trace",
			expectError: false,
			description: "should accept trace log level",
		},
		{
			name:        "invalid log level",
			logLevel:    "invalid",
			expectError: false,
			description: "should accept invalid log level (fallback to info)",
		},
		{
			name:        "empty log level",
			logLevel:    "",
			expectError: false,
			description: "should accept empty log level (fallback to info)",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app and set flags (simulating Before hook behavior)
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "log-level", Value: "info"},
					&cli.StringFlag{Name: "format", Value: "text"},
					&cli.BoolFlag{Name: "debug"},
				},
			}

			err := app.Set("log-level", testCase.logLevel)
			require.NoError(t, err)

			// Test InitLogger with the flag value (simulating main.go Before hook)
			ctx := context.Background()
			output := app.String("format")
			debug := app.Bool("debug")
			logLevel := app.String("log-level")

			// This should not panic or error for any log level
			resultCtx := logadapter.InitLogger(ctx, output, debug, logLevel)

			// Verify context is returned properly
			require.NotNil(t, resultCtx, testCase.description)

			// Verify logger can be retrieved from context
			logger := logadapter.GetLogger(resultCtx)
			require.NotNil(t, logger, "should be able to retrieve logger from context")
		})
	}
}

func TestLogLevelFlagDefault(t *testing.T) {
	// Test that the default log level works correctly
	app := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "log-level", Value: "info"}, // Default from main.go
			&cli.StringFlag{Name: "format", Value: "text"},
			&cli.BoolFlag{Name: "debug"},
		},
	}

	// Don't set the flag explicitly - should use default
	ctx := context.Background()
	output := app.String("format")
	debug := app.Bool("debug")
	logLevel := app.String("log-level") // Should be "info" by default

	require.Equal(t, "info", logLevel, "default log level should be info")

	// Should work with default value
	resultCtx := logadapter.InitLogger(ctx, output, debug, logLevel)
	require.NotNil(t, resultCtx, "should work with default log level")

	logger := logadapter.GetLogger(resultCtx)
	require.NotNil(t, logger, "should be able to retrieve logger from context")
}

func TestLogLevelFlagWithOtherFlags(t *testing.T) {
	// Test log level flag works in combination with other flags
	tests := []struct {
		name        string
		logLevel    string
		format      string
		debug       bool
		description string
	}{
		{
			name:        "error with json format",
			logLevel:    "error",
			format:      "json",
			debug:       false,
			description: "should work with json format",
		},
		{
			name:        "trace with debug enabled",
			logLevel:    "trace",
			format:      "text",
			debug:       true,
			description: "should work with debug flag",
		},
		{
			name:        "info with github format",
			logLevel:    "info",
			format:      "github",
			debug:       false,
			description: "should work with github format",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "log-level", Value: "info"},
					&cli.StringFlag{Name: "format", Value: "text"},
					&cli.BoolFlag{Name: "debug"},
				},
			}

			err := app.Set("log-level", testCase.logLevel)
			require.NoError(t, err)
			err = app.Set("format", testCase.format)
			require.NoError(t, err)

			err = app.Set("debug", "true")
			if testCase.debug {
				require.NoError(t, err)
			}

			ctx := context.Background()
			output := app.String("format")
			debug := app.Bool("debug")
			logLevel := app.String("log-level")

			// Should work with all flag combinations
			resultCtx := logadapter.InitLogger(ctx, output, debug, logLevel)
			require.NotNil(t, resultCtx, testCase.description)

			logger := logadapter.GetLogger(resultCtx)
			require.NotNil(t, logger, "should retrieve logger regardless of flag combination")
		})
	}
}
