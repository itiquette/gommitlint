// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package format

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
)

// Result formats a rule result into a simple message.
func Result(_ string, errors []domain.ValidationError) string {
	if len(errors) == 0 {
		return "No errors"
	}

	messages := make([]string, 0, len(errors))
	for _, err := range errors {
		messages = append(messages, err.Message)
	}

	return strings.Join(messages, "; ")
}

// VerboseResult formats a rule result with more detail.
func VerboseResult(_ string, errors []domain.ValidationError) string {
	if len(errors) == 0 {
		return "No errors"
	}

	parts := make([]string, 0, len(errors))

	for _, err := range errors {
		msg := err.Message
		if err.Code != "" {
			msg = fmt.Sprintf("[%s] %s", err.Code, msg)
		}

		if len(err.Context) > 0 {
			var contextParts []string
			for k, v := range err.Context {
				contextParts = append(contextParts, fmt.Sprintf("%s=%s", k, v))
			}

			msg = fmt.Sprintf("%s (%s)", msg, strings.Join(contextParts, ", "))
		}

		parts = append(parts, msg)
	}

	return strings.Join(parts, "; ")
}

// Help extracts help messages from validation errors.
func Help(_ string, errors []domain.ValidationError) string {
	var helps []string

	seen := make(map[string]bool)

	for _, err := range errors {
		if err.Help != "" && !seen[err.Help] {
			helps = append(helps, err.Help)
			seen[err.Help] = true
		}
	}

	return strings.Join(helps, "; ")
}
