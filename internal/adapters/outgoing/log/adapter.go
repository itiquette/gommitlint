// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package log

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/ports/outgoing"
	"github.com/rs/zerolog"
)

// Adapter adapts the zerolog logger to the simpler outgoing.Logger interface.
type Adapter struct {
	logger zerolog.Logger
}

// NewAdapter creates a new adapter for the simpler outgoing.Logger interface.
func NewAdapter(logger zerolog.Logger) *Adapter {
	return &Adapter{logger: logger}
}

// Ensure Adapter implements outgoing.Logger.
var _ outgoing.Logger = Adapter{}

func (s Adapter) Debug(msg string, args ...any) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	s.addFields(s.logger.Debug(), args...).Msg(msg)
}

func (s Adapter) Info(msg string, args ...any) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	s.addFields(s.logger.Info(), args...).Msg(msg)
}

func (s Adapter) Warn(msg string, args ...any) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	s.addFields(s.logger.Warn(), args...).Msg(msg)
}

func (s Adapter) Error(msg string, args ...any) {
	//nolint:zerologlint // false positive - addFields returns event that is then dispatched
	s.addFields(s.logger.Error(), args...).Msg(msg)
}

func (s Adapter) addFields(event *zerolog.Event, args ...any) *zerolog.Event {
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
