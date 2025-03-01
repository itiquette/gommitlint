// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// SignOffRegex is the regular expression used to validate the Developer Certificate of Origin signature.
var SignOffRegex = regexp.MustCompile(`^Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>$`)

// SignOff ensures that the commit message contains a
// Developer Certificate of Origin signature.
type SignOff struct {
	errors []error
}

func (d SignOff) Name() string {
	return "SignOff"
}

// Result returns the check message.
func (d SignOff) Result() string {
	if len(d.errors) != 0 {
		return d.errors[0].Error()
	}

	return "Sign-off exists"
}

// Errors returns any violations of the check.
func (d SignOff) Errors() []error {
	return d.errors
}

func ValidateSignOff(body string) *SignOff {
	rule := &SignOff{}

	for _, line := range strings.Split(body, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if SignOffRegex.MatchString(trimmedLine) {
			return rule
		}
	}

	rule.errors = append(rule.errors, errors.New(`Commit must be signed-off.
Use 'git commit -s' or manually add a sign-off line.

Example - A complete commit message with sign-off:

feat: introduce rate limiting for API endpoints

Adds rate limiting to prevent API abuse:
- Implements token bucket algorithm
- Configurable limits per endpoint

Signed-off-by: Laval Lajon <laval.lajon@cavora.exampleorg>`))

	return rule
}
