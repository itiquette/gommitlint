// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// Logger implements domain.Logger interface using zerolog.
type Logger struct {
	logger zerolog.Logger
}

// New creates a new domain logger from zerolog.
func New(logger zerolog.Logger) domain.Logger {
	return Logger{logger: logger}
}

// Debug outputs a debug-level message with optional key-value arguments.
func (l Logger) Debug(msg string, args ...interface{}) {
	l.logWithArgs(l.logger.Debug(), msg, args...)
}

// Info outputs an info-level message with optional key-value arguments.
func (l Logger) Info(msg string, args ...interface{}) {
	l.logWithArgs(l.logger.Info(), msg, args...)
}

// Error outputs an error-level message with optional key-value arguments.
func (l Logger) Error(msg string, args ...interface{}) {
	l.logWithArgs(l.logger.Error(), msg, args...)
}

// Log outputs a message with the specified level and optional key-value arguments.
func (l Logger) Log(level string, msg string, args ...interface{}) {
	event := l.getLogEvent(level)
	l.logWithArgs(event, msg, args...)
}

// logWithArgs adds key-value arguments to log event and dispatches message.
func (l Logger) logWithArgs(event *zerolog.Event, msg string, args ...interface{}) {
	if len(args) > 0 {
		event = l.addFields(event, args...)
	}

	event.Msg(msg)
}

// addFields adds key-value pairs to zerolog event.
// Handles odd-length args by padding with empty string for consistency.
func (l Logger) addFields(event *zerolog.Event, args ...interface{}) *zerolog.Event {
	if len(args)%2 != 0 {
		args = append(args, "") // Pad odd args consistently
	}

	for i := 0; i < len(args); i += 2 {
		key := fmt.Sprint(args[i])
		event = event.Interface(key, args[i+1])
	}

	return event
}

// getLogEvent returns zerolog event for specified level.
//
//nolint:zerologlint // Returns event to caller for proper dispatch
func (l Logger) getLogEvent(level string) *zerolog.Event {
	switch level {
	case "debug":
		return l.logger.Debug()
	case "info":
		return l.logger.Info()
	case "warn", "warning":
		return l.logger.Warn()
	case "error":
		return l.logger.Error()
	default:
		return l.logger.Info() // Default to info
	}
}

// Initialization functions for CLI integration

// InitLogger creates a configured zerolog instance from CLI flags.
func InitLogger(ctx context.Context, cmd *cobra.Command, outputFormat string) context.Context {
	level := getLogLevel(cmd)
	writer := createWriter(outputFormat)
	logger := createZerologger(writer, level, cmd)

	return logger.WithContext(ctx)
}

// GetLogger retrieves zerolog from context.
func GetLogger(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// NewDomainLogger creates a domain logger from zerolog pointer - adapter for existing code.
func NewDomainLogger(logger *zerolog.Logger) domain.Logger {
	return New(*logger)
}

// createWriter creates appropriate io.Writer based on format.
func createWriter(outputFormat string) io.Writer {
	if outputFormat == "json" {
		zerolog.TimeFieldFormat = time.RFC3339

		return os.Stdout
	}

	return zerolog.ConsoleWriter{
		Out:         os.Stderr,
		TimeFormat:  time.RFC3339,
		FormatLevel: formatLevel,
	}
}

// createZerologger creates configured zerolog instance.
func createZerologger(writer io.Writer, level zerolog.Level, cmd *cobra.Command) zerolog.Logger {
	loggerContext := zerolog.New(writer).Level(level).With().Timestamp()

	// Add caller info if requested
	if cmd != nil {
		if caller, _ := cmd.Flags().GetBool("caller"); caller {
			loggerContext = loggerContext.Caller()
		}
	}

	return loggerContext.Logger()
}

// getLogLevel determines log level from CLI flags.
func getLogLevel(cmd *cobra.Command) zerolog.Level {
	if cmd == nil {
		return zerolog.InfoLevel
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	if quiet {
		return zerolog.ErrorLevel
	}

	level, _ := cmd.Flags().GetString("verbosity")
	switch level {
	case "quiet":
		return zerolog.ErrorLevel
	case "trace":
		return zerolog.TraceLevel
	case "brief":
		return zerolog.InfoLevel
	default:
		return zerolog.InfoLevel
	}
}

// formatLevel formats log level with colors.
func formatLevel(levelVal interface{}) string {
	if levelStr, ok := levelVal.(string); ok {
		switch levelStr {
		case "trace":
			return "\x1b[90mTRC\x1b[0m"
		case "info":
			return "\x1b[34mINF\x1b[0m"
		case "warn":
			return "\x1b[33mWRN\x1b[0m"
		case "error":
			return "\x1b[31mERR\x1b[0m"
		default:
			return "\x1b[37m" + levelStr + "\x1b[0m"
		}
	}

	return fmt.Sprintf("%s", levelVal)
}
