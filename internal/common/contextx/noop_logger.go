// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package contextx

import (
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// noOpLogger is a logger implementation that discards all log messages.
// It's used as a fallback when no logger is found in the context.
type noOpLogger struct{}

// Ensure noOpLogger implements the Logger interface.
var _ outgoing.Logger = (*noOpLogger)(nil)

// NewNoOpLogger creates a new logger that discards all messages.
func NewNoOpLogger() outgoing.Logger {
	return &noOpLogger{}
}

// Debug implements the Logger interface but does nothing.
func (n *noOpLogger) Debug(_ string, _ ...interface{}) {}

// Info implements the Logger interface but does nothing.
func (n *noOpLogger) Info(_ string, _ ...interface{}) {}

// Warn implements the Logger interface but does nothing.
func (n *noOpLogger) Warn(_ string, _ ...interface{}) {}

// Error implements the Logger interface but does nothing.
func (n *noOpLogger) Error(_ string, _ ...interface{}) {}
