// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// RequiredBodyThreshold is the default minimum number of line changes required
// to trigger the body check.
var RequiredBodyThreshold = 10

// Body enforces a maximum number of charcters on the commit
// header.
type Body struct {
	errors []error
}

// Status returns the name of the check.
func (h Body) Status() string {
	return "Commit Body"
}

// Message returns to check message.
func (h Body) Message() string {
	if len(h.errors) != 0 {
		return h.errors[0].Error()
	}

	return "Commit body is valid"
}

// Errors returns any violations of the check.
func (h Body) Errors() []error {
	return h.errors
}

// ValidateBody checks the header length.
func ValidateBody(message string) interfaces.Check { //nolint:ireturn
	check := &Body{}

	lines := strings.Split(strings.TrimPrefix(message, "\n"), "\n")
	valid := false

	for _, line := range lines[1:] {
		if DCORegex.MatchString(strings.TrimSpace(line)) {
			continue
		}

		if line != "" {
			valid = true

			break
		}
	}

	if !valid {
		check.errors = append(check.errors, errors.New("Commit body is empty"))
	}

	return check
}
