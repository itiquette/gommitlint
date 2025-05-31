// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package log

import (
	"fmt"

	"github.com/rs/zerolog"
)

// Logger adapts the zerolog logger to the simpler domain.Logger interface.
type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new adapter for the simpler domain.Logger interface.
func NewLogger(logger zerolog.Logger) Logger {
	return Logger{logger: logger}
}

// Logger implements the Logger interfaces defined by various consumers
// (cli.Logger, git.Logger, etc.) through structural typing.

func (l Logger) Debug(msg string, args ...any) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	l.addFields(l.logger.Debug(), args...).Msg(msg)
}

func (l Logger) Info(msg string, args ...any) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	l.addFields(l.logger.Info(), args...).Msg(msg)
}

func (l Logger) Warn(msg string, args ...any) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	l.addFields(l.logger.Warn(), args...).Msg(msg)
}

func (l Logger) Error(msg string, args ...any) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	l.addFields(l.logger.Error(), args...).Msg(msg)
}

func (l Logger) addFields(event *zerolog.Event, args ...any) *zerolog.Event {
	if len(args)%2 != 0 {
		args = append(args, "")
	}

	for i := 0; i < len(args); i += 2 {
		key := fmt.Sprint(args[i])
		value := args[i+1]
		event = event.Interface(key, value)
	}

	return event
}
