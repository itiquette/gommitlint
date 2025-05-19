// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package logger contains test utilities for logging in tests.
// This package is intended for testing purposes only.
package logger

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// PrintLogger is a simple Logger implementation that prints to stdout/stderr.
// It's intended for testing only and should not be used in production code.
type PrintLogger struct {
	Level int // 0=error, 1=warn, 2=info, 3=debug
}

// Ensure PrintLogger implements outgoing.Logger.
var _ outgoing.Logger = (*PrintLogger)(nil)

// NewPrintLogger creates a new PrintLogger with the given level.
func NewPrintLogger(level int) *PrintLogger {
	return &PrintLogger{
		Level: level,
	}
}

// Debug implements outgoing.Logger.Debug.
func (l *PrintLogger) Debug(msg string, kvs ...interface{}) {
	if l.Level < 3 {
		return
	}

	l.log("DEBUG", msg, kvs...)
}

// Info implements outgoing.Logger.Info.
func (l *PrintLogger) Info(msg string, kvs ...interface{}) {
	if l.Level < 2 {
		return
	}

	l.log("INFO", msg, kvs...)
}

// Warn implements outgoing.Logger.Warn.
func (l *PrintLogger) Warn(msg string, kvs ...interface{}) {
	if l.Level < 1 {
		return
	}

	l.log("WARN", msg, kvs...)
}

// Error implements outgoing.Logger.Error.
func (l *PrintLogger) Error(msg string, kvs ...interface{}) {
	l.log("ERROR", msg, kvs...)
}

// log is the internal logging implementation.
func (l *PrintLogger) log(level, msg string, kvs ...interface{}) {
	parts := []string{fmt.Sprintf("[%s] %s", level, msg)}

	// Format key-value pairs
	for i := 0; i < len(kvs); i += 2 {
		if i+1 < len(kvs) {
			key := fmt.Sprintf("%v", kvs[i])
			value := fmt.Sprintf("%v", kvs[i+1])
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		}
	}

	fmt.Println(strings.Join(parts, " "))
}
