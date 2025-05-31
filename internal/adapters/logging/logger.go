// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package log provides logging functionality
// using zerolog and cobra. It offers easy setup of leveled logging.
package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// Level represents the available log levels.
type Level string

// Predefined log levels.
const (
	LevelQuiet Level = "quiet" // Only error messages
	LevelBrief Level = "brief" // Info and above (default)
	LevelTrace Level = "trace" // Trace and above (most verbose)
)

// Format represents the available log output formats.
type Format string

const (
	JSON    Format = "json"
	CONSOLE Format = "console"
)

// ToZerologLevel converts a LogLevel to the corresponding zerolog.Level.
func (l Level) ToZerologLevel() zerolog.Level {
	switch l {
	case LevelQuiet:
		return zerolog.ErrorLevel
	case LevelTrace:
		return zerolog.TraceLevel
	case LevelBrief:
		return zerolog.InfoLevel
	default:
		return zerolog.InfoLevel // Default to Info for LevelBrief and unknown levels
	}
}

// InitLogger initializes and returns a context with a configured logger.
// It sets up the logger based on the command line flags for verbosity and quiet mode.
//
// Parameters:
//   - ctx: The parent context
//   - cmd: The cobra.Command instance, used to retrieve flags
//   - withCaller: Whether to include caller information in logs
//   - outputFormat: The output format (json or console)
//
// Returns:
//   - context.Context with the configured logger
//
// The logger is set up with a console writer for human-readable output unless json is specified.
// If the log level is set to trace, it includes the caller information in the log output.
func InitLogger(ctx context.Context, cmd *cobra.Command, outputFormat string) context.Context {
	level := getLogLevel(cmd)

	var writer io.Writer
	if outputFormat == "json" {
		writer = os.Stdout
		zerolog.TimeFieldFormat = time.RFC3339
	} else {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			// Improve readability
			FormatLevel: func(levelVal interface{}) string {
				if levelStr, ok := levelVal.(string); ok {
					switch levelStr {
					case "trace":
						return "\x1b[90m" + "TRC" + "\x1b[0m"
					case "info":
						return "\x1b[34m" + "INF" + "\x1b[0m"
					case "warn":
						return "\x1b[33m" + "WRN" + "\x1b[0m"
					case "error":
						return "\x1b[31m" + "ERR" + "\x1b[0m"
					default:
						return "\x1b[37m" + levelStr + "\x1b[0m"
					}
				}

				return fmt.Sprintf("%s", levelVal)
			},
		}
	}

	var logger zerolog.Logger

	// Configure logger based on settings
	loggerContext := zerolog.New(writer).Level(level).With()

	// Check if caller flag is set (handle nil cmd)
	if cmd != nil {
		if caller, _ := cmd.Flags().GetBool("caller"); caller {
			loggerContext = loggerContext.Caller()
		}
	}

	// Always include timestamp
	loggerContext = loggerContext.Timestamp()

	// Build the logger
	logger = loggerContext.Logger()

	// Add logger to context
	ctx = logger.WithContext(ctx)

	return ctx
}

// GetLogger retrieves the zerolog.Logger from the given context.
//
// Parameters:
//   - ctx: The context containing the logger
//
// Returns:
//   - *zerolog.Logger: A pointer to the logger instance
//
// Usage:
//
//	logger := log.GetLogger(ctx)
//	logger.Info().Msg("This is an info message")
func GetLogger(ctx context.Context) *zerolog.Logger {
	// Get from zerolog context
	return zerolog.Ctx(ctx)
}

// getLogLevel determines the log level based on command flags.
// It checks for the "quiet" flag first, then falls back to the "verbosity" flag.
//
// Parameters:
//   - cmd: The cobra.Command instance to retrieve flags from
//
// Returns:
//   - zerolog.Level: The determined log level
//
// Note: This function assumes that the "verbosity" flag is a string
// and the "quiet" flag is a boolean.
func getLogLevel(cmd *cobra.Command) zerolog.Level {
	// Handle nil command for early initialization
	if cmd == nil {
		return LevelBrief.ToZerologLevel()
	}

	level, _ := cmd.Flags().GetString("verbosity")
	quiet, _ := cmd.Flags().GetBool("quiet")

	if quiet {
		// For quiet mode, set to Error level only, which effectively disables all normal logs
		return zerolog.ErrorLevel
	}

	// For "log", "brief", "trace", get from Level enum
	return Level(level).ToZerologLevel()
}

// SetLogLevel sets the log level of a logger.
// This is useful for dynamically changing the log level.
//
// Parameters:
//   - logger: The logger to modify
//   - level: The new log level
//
// Returns:
//   - *zerolog.Logger: A new logger with the updated level
func SetLogLevel(logger *zerolog.Logger, level Level) *zerolog.Logger {
	newLogger := logger.Level(level.ToZerologLevel())

	return &newLogger
}
