// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package testdata provides test fixtures and helpers for integration tests.
package testdata

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// AssertFileExists checks that a file exists at the given path.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	require.NoError(t, err, "file should exist: %s", path)
}

// AssertFileNotExists checks that a file does not exist at the given path.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err), "file should not exist: %s", path)
}

// AssertFileContent checks that a file contains the expected content.
func AssertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, expected, string(content))
}

// AssertFileContains checks that a file contains the expected substring.
func AssertFileContains(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(content), expected)
}

// AssertFilePermissions checks that a file has the expected permissions.
func AssertFilePermissions(t *testing.T, path string, expected os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, expected, info.Mode().Perm())
}

// CreateTestLogger creates a logger for testing.
// By default it discards output, but can be configured to write to testing.T.
func CreateTestLogger(t *testing.T, verbose bool) *zerolog.Logger {
	t.Helper()

	var writer = io.Discard
	if verbose {
		writer = TestLogWriter{t: t}
	}

	logger := zerolog.New(writer).With().Timestamp().Logger()

	return &logger
}

// TestLogWriter writes log output to testing.T.
type TestLogWriter struct {
	t *testing.T
}

// Write implements io.Writer.
func (writer TestLogWriter) Write(p []byte) (int, error) {
	writer.t.Log(strings.TrimSpace(string(p)))

	return len(p), nil
}

// LogAdapter adapts zerolog.Logger to a logger interface for testing.
type LogAdapter struct {
	logger *zerolog.Logger
}

// NewLogAdapter creates a new log adapter for testing.
// The returned adapter satisfies the Logger interfaces used by various packages.
func NewLogAdapter(logger *zerolog.Logger) *LogAdapter {
	return &LogAdapter{logger: logger}
}

// Debug logs a debug message.
func (la *LogAdapter) Debug(msg string, _ ...interface{}) {
	la.logger.Debug().Msg(msg)
}

// Info logs an info message.
func (la *LogAdapter) Info(msg string, _ ...interface{}) {
	la.logger.Info().Msg(msg)
}

// Warn logs a warning message.
func (la *LogAdapter) Warn(msg string, _ ...interface{}) {
	la.logger.Warn().Msg(msg)
}

// Error logs an error message.
func (la *LogAdapter) Error(msg string, _ ...interface{}) {
	la.logger.Error().Msg(msg)
}