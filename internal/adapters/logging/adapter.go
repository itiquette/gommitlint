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

// parseLogLevel converts string log level to zerolog.Level (pure function).
func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "error":
		return zerolog.ErrorLevel // Only errors
	case "warn":
		return zerolog.WarnLevel // Warnings and above
	case "info":
		return zerolog.InfoLevel // Info and above (default)
	case "debug":
		return zerolog.DebugLevel // Debug and above
	case "trace":
		return zerolog.TraceLevel // All levels including trace
	default:
		return zerolog.InfoLevel // Safe fallback to info
	}
}

// InitLogger creates a configured zerolog instance.
func InitLogger(ctx context.Context, outputFormat string, debug bool, logLevel string) context.Context {
	level := parseLogLevel(logLevel)
	writer := createWriter(outputFormat)
	logger := createZerologger(writer, level, debug)

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
func createZerologger(writer io.Writer, level zerolog.Level, withCaller bool) zerolog.Logger {
	loggerContext := zerolog.New(writer).Level(level).With().Timestamp()

	// Add caller info if requested
	if withCaller {
		loggerContext = loggerContext.Caller()
	}

	return loggerContext.Logger()
}

// formatLevel formats log level with colors.
func formatLevel(levelVal interface{}) string {
	if levelStr, ok := levelVal.(string); ok {
		switch levelStr {
		case "debug":
			return "\x1b[90mDBG\x1b[0m"
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
