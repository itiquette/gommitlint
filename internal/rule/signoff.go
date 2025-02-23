// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package rule

import (
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// DCORegex is the regular expression used to validate the Developer Certificate of Origin signature.
var DCORegex = regexp.MustCompile(`^Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>$`)

// SignOff ensures that the commit message contains a
// Developer Certificate of Origin signature.
type SignOff struct {
	errors []error
}

func (d SignOff) Name() string {
	return "DCO"
}

// Message returns the check message.
func (d SignOff) Message() string {
	if len(d.errors) != 0 {
		return d.errors[0].Error()
	}

	return "Developer Certificate of Origin signature is valid"
}

// Errors returns any violations of the check.
func (d SignOff) Errors() []error {
	return d.errors
}

func ValidateSignOff(message string) interfaces.Rule { //nolint:ireturn
	rule := &SignOff{}

	for _, line := range strings.Split(message, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if DCORegex.MatchString(trimmedLine) {
			return rule
		}
	}

	rule.errors = append(rule.errors, errors.New(`Commit must be signed-off with a Developer Certificate of Origin (DCO).
Use 'git commit -s' or manually add a sign-off line.

Example - A complete commit message with sign-off:

feat: introduce rate limiting for API endpoints

Adds rate limiting to prevent API abuse:
- Implements token bucket algorithm
- Configurable limits per endpoint

Signed-off-by: Jane Smith <jane.smith@example.com>`))

	return rule
}
