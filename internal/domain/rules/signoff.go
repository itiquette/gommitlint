// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// signOffRegex returns a regex for strict DCO compliance.
// Enforces proper "Name <email@domain>" format as required by DCO specification.
func signOffRegex() *regexp.Regexp {
	return regexp.MustCompile(`^Signed-off-by:\s+.+\s+<.+@.+>$`)
}

// SignOffRule validates that commit messages include a sign-off line.
type SignOffRule struct {
	minSignoffCount int
}

// NewSignOffRule creates a new rule for validating commit sign-offs from config.
func NewSignOffRule(cfg config.Config) SignOffRule {
	return SignOffRule{
		minSignoffCount: cfg.Message.Body.MinSignoffCount,
	}
}

// Validate checks for the presence and format of a Developer Certificate of Origin sign-off.
func (r SignOffRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// Skip validation if no sign-offs required
	if r.minSignoffCount == 0 {
		return nil
	}

	var errors []domain.ValidationError

	// Extract and validate sign-offs using rule methods
	signoffs := r.extractSignoffs(commit.Body)

	// Validate sign-off count
	if countErrors := r.validateSignoffCount(signoffs); len(countErrors) > 0 {
		errors = append(errors, countErrors...)
	}

	// Validate sign-off format (only if we have sign-offs)
	if len(signoffs) > 0 {
		if formatErrors := r.validateSignoffFormat(signoffs); len(formatErrors) > 0 {
			errors = append(errors, formatErrors...)
		}

		// Validate sign-off placement
		if placementErrors := r.validateSignoffPlacement(commit.Body); len(placementErrors) > 0 {
			errors = append(errors, placementErrors...)
		}

		// Validate unique signers (only if multiple required)
		if r.minSignoffCount > 1 {
			if uniqueErrors := r.validateUniqueSigners(signoffs); len(uniqueErrors) > 0 {
				errors = append(errors, uniqueErrors...)
			}
		}
	}

	return errors
}

// Name returns the rule name.
func (r SignOffRule) Name() string {
	return "SignOff"
}

// extractSignoffs extracts all valid sign-off lines from the commit body.
// Returns a slice of sign-off strings found in the body.
func (r SignOffRule) extractSignoffs(body string) []string {
	if strings.TrimSpace(body) == "" {
		return nil
	}

	lines := strings.Split(body, "\n")
	regex := signOffRegex()

	var signoffs []string

	// Extract all valid sign-off lines
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if regex.MatchString(trimmedLine) {
			signoffs = append(signoffs, trimmedLine)
		}
	}

	return signoffs
}

// validateSignoffCount validates that the required number of sign-offs are present.
func (r SignOffRule) validateSignoffCount(signoffs []string) []domain.ValidationError {
	if r.minSignoffCount == 0 {
		return nil // No sign-offs required
	}

	actualCount := len(signoffs)
	if actualCount < r.minSignoffCount {
		err := domain.New(r.Name(), domain.ErrMissingSignoff, "Missing required sign-off")
		err = err.WithContextMap(map[string]string{
			"actual":   strconv.Itoa(actualCount),
			"expected": strconv.Itoa(r.minSignoffCount),
		})
		err = err.WithHelp("Add DCO sign-off line: 'Signed-off-by: Your Real Name <your.email@domain.com>'")

		return []domain.ValidationError{err}
	}

	return nil
}

// validateSignoffFormat validates that all sign-offs have strict DCO format.
// Enforces "Signed-off-by: Name <email@domain>" format for DCO compliance.
func (r SignOffRule) validateSignoffFormat(signoffs []string) []domain.ValidationError {
	regex := signOffRegex()

	var errors []domain.ValidationError

	// Strict DCO compliance: require proper "Name <email@domain>" format
	for _, signoff := range signoffs {
		if !regex.MatchString(signoff) {
			err := domain.New(r.Name(), domain.ErrInvalidSignoffFormat, "Invalid DCO sign-off format")
			err = err.WithContextMap(map[string]string{
				"actual":   signoff,
				"expected": "Signed-off-by: Real Name <email@domain.com>",
			})
			err = err.WithHelp("Use strict DCO format: 'Signed-off-by: Your Real Name <your.email@domain.com>'")
			errors = append(errors, err)
		}
	}

	return errors
}

// validateSignoffPlacement validates that sign-offs are placed at the end of the body.
func (r SignOffRule) validateSignoffPlacement(body string) []domain.ValidationError {
	if strings.TrimSpace(body) == "" {
		return nil
	}

	lines := strings.Split(body, "\n")
	regex := signOffRegex()

	// Find the last sign-off line
	lastSignoffIndex := -1

	for idx := len(lines) - 1; idx >= 0; idx-- {
		trimmedLine := strings.TrimSpace(lines[idx])
		if trimmedLine == "" {
			continue // Skip empty lines
		}

		if regex.MatchString(trimmedLine) {
			lastSignoffIndex = idx

			break
		}
	}

	if lastSignoffIndex == -1 {
		return nil // No sign-offs found, placement validation not applicable
	}

	// Check if there's any non-empty, non-sign-off content after the last sign-off
	for lineIdx := lastSignoffIndex + 1; lineIdx < len(lines); lineIdx++ {
		trimmedLine := strings.TrimSpace(lines[lineIdx])
		if trimmedLine != "" && !regex.MatchString(trimmedLine) {
			err := domain.New(r.Name(), domain.ErrMisplacedSignoff, "Sign-offs must be at the end of the commit message")
			err = err.WithContextMap(map[string]string{
				"actual":   "Content found after sign-off",
				"expected": "Sign-offs at end",
			})
			err = err.WithHelp("Move all sign-off lines to the end")

			return []domain.ValidationError{err}
		}
	}

	return nil
}

// validateUniqueSigners validates that when multiple sign-offs are required, they are from different signers.
func (r SignOffRule) validateUniqueSigners(signoffs []string) []domain.ValidationError {
	if r.minSignoffCount <= 1 || len(signoffs) <= 1 {
		return nil // Uniqueness validation not applicable
	}

	seenSigners := make(map[string]bool)
	duplicates := []string{}

	for _, signoff := range signoffs {
		// Extract email from sign-off line
		email := r.extractEmailFromSignoff(signoff)
		if email == "" {
			continue // Skip malformed sign-offs (format validation handles this)
		}

		if seenSigners[email] {
			duplicates = append(duplicates, email)
		} else {
			seenSigners[email] = true
		}
	}

	if len(duplicates) > 0 {
		err := domain.New(r.Name(), domain.ErrInsufficientSignoffs, fmt.Sprintf("Duplicate signers found: %v", duplicates))
		err = err.WithContextMap(map[string]string{
			"actual":   strings.Join(duplicates, ", "),
			"expected": strconv.Itoa(r.minSignoffCount) + " unique signers",
		})
		err = err.WithHelp("Each sign-off must be from a different person")

		return []domain.ValidationError{err}
	}

	return nil
}

// extractEmailFromSignoff extracts the email address from a sign-off line.
func (r SignOffRule) extractEmailFromSignoff(signoff string) string {
	// Match email in angle brackets
	emailRegex := regexp.MustCompile(`<([^>]+)>`)

	matches := emailRegex.FindStringSubmatch(signoff)
	if len(matches) >= 2 {
		return matches[1]
	}

	return ""
}

// CountBodyLines counts the number of non-empty lines in a body.
func CountBodyLines(body string) int {
	// If body is empty, return 0
	if strings.TrimSpace(body) == "" {
		return 0
	}

	lines := strings.Split(body, "\n")
	count := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			count++
		}
	}

	return count
}
