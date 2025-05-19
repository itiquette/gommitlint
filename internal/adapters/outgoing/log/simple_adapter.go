// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package log

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/ports/outgoing"
	"github.com/rs/zerolog"
)

// SimpleAdapter adapts the zerolog logger to the simpler outgoing.Logger interface.
type SimpleAdapter struct {
	logger zerolog.Logger
}

// NewSimpleAdapter creates a new adapter for the simpler outgoing.Logger interface.
func NewSimpleAdapter(logger zerolog.Logger) *SimpleAdapter {
	return &SimpleAdapter{logger: logger}
}

// Ensure SimpleAdapter implements outgoing.Logger.
var _ outgoing.Logger = (*SimpleAdapter)(nil)

func (s *SimpleAdapter) Debug(msg string, args ...interface{}) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	s.addFields(s.logger.Debug(), args...).Msg(msg)
}

func (s *SimpleAdapter) Info(msg string, args ...interface{}) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	s.addFields(s.logger.Info(), args...).Msg(msg)
}

func (s *SimpleAdapter) Warn(msg string, args ...interface{}) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	s.addFields(s.logger.Warn(), args...).Msg(msg)
}

func (s *SimpleAdapter) Error(msg string, args ...interface{}) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	s.addFields(s.logger.Error(), args...).Msg(msg)
}

func (s *SimpleAdapter) addFields(event *zerolog.Event, args ...interface{}) *zerolog.Event {
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
