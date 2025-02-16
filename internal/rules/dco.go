// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// DCORegex is the regular expression used to validate the Developer Certificate of Origin signature.
var DCORegex = regexp.MustCompile(`^Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>$`)

// DCOCheck ensures that the commit message contains a
// Developer Certificate of Origin signature.
type DCOCheck struct {
	errors []error
}

func (d DCOCheck) Name() string {
	return "DCO"
}

func (d DCOCheck) Status() string {
	return "DCO"
}

// Message returns the check message.
func (d DCOCheck) Message() string {
	if len(d.errors) != 0 {
		return d.errors[0].Error()
	}

	return "Developer Certificate of Origin signature is valid"
}

// Errors returns any violations of the check.
func (d DCOCheck) Errors() []error {
	return d.errors
}

func ValidateDCO(message string) interfaces.Check { //nolint:ireturn
	check := &DCOCheck{}

	for _, line := range strings.Split(message, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if DCORegex.MatchString(trimmedLine) {
			return check
		}
	}

	check.errors = append(check.errors, errors.New(`Commit must be signed-off with a Developer Certificate of Origin (DCO).
Use 'git commit -s' or manually add a sign-off line.

Example - A complete commit message with sign-off:

feat: introduce rate limiting for API endpoints

Adds rate limiting to prevent API abuse:
- Implements token bucket algorithm
- Configurable limits per endpoint

Signed-off-by: Jane Smith <jane.smith@example.com>`))

	return check
}
