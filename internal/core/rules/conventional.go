// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errorx"
)

// subjectRegex Format: type(scope)!: description.
var subjectRegex = regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ](.+)$`)

// ConventionalCommitRule enforces the Conventional Commits specification format for commit messages.
//
// This rule validates that commit messages follow the structured format defined by the
// Conventional Commits specification (https://www.conventionalcommits.org/), which provides
// a standardized way to communicate the purpose and scope of changes.
type ConventionalCommitRule struct {
	allowedTypes  []string
	allowedScopes []string
	maxDescLength int
	errors        []*domain.ValidationError
	commitType    string // Store for verbose output
	scope         string // Store for verbose output
	hasBreaking   bool   // Store for verbose output
}

// NewConventionalCommitRule creates a new ConventionalCommitRule with specified configuration.
func NewConventionalCommitRule(types []string, scopes []string, maxDescLength int) *ConventionalCommitRule {
	// Default description length if not specified
	if maxDescLength <= 0 {
		maxDescLength = 72
	}

	return &ConventionalCommitRule{
		allowedTypes:  types,
		allowedScopes: scopes,
		maxDescLength: maxDescLength,
		errors:        make([]*domain.ValidationError, 0),
	}
}

// Name returns the rule identifier.
func (r *ConventionalCommitRule) Name() string {
	return "ConventionalCommit"
}

// Validate validates a commit against the conventional commit rules.
func (r *ConventionalCommitRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
	// Reset errors and state
	r.errors = make([]*domain.ValidationError, 0)
	r.commitType = ""
	r.scope = ""
	r.hasBreaking = false

	subject := commit.Subject

	// Validate the basic structure of the commit
	if !r.validateBasicFormat(subject) {
		return r.errors
	}

	// Parse and extract components
	commitType, scope, hasBreaking, description := r.extractComponents(subject)

	// Store for verbose output
	r.commitType = commitType
	r.scope = scope
	r.hasBreaking = hasBreaking

	// Validate type, scope, and description
	if !r.validateType(commitType) ||
		!r.validateScope(scope) ||
		!r.validateDescription(description) {
		return r.errors
	}

	return r.errors
}

// validateBasicFormat validates the basic format of the commit subject.
func (r *ConventionalCommitRule) validateBasicFormat(subject string) bool {
	// Handle empty subject early
	if strings.TrimSpace(subject) == "" {
		r.errors = append(r.errors, errorx.NewErrorWithContext(
			r.Name(),
			errorx.ErrInvalidFormat,
			map[string]string{"subject": subject},
			"empty message",
		))

		return false
	}

	// Simple check for ": " vs ":  " (one space vs multiple spaces)
	if strings.Contains(subject, ":  ") {
		r.errors = append(r.errors, errorx.NewErrorWithContext(
			r.Name(),
			errorx.ErrInvalidFormat,
			map[string]string{"subject": subject},
			"spacing error",
		))

		return false
	}

	// Validate basic format
	if !subjectRegex.MatchString(subject) {
		r.errors = append(r.errors, errorx.NewErrorWithContext(
			r.Name(),
			errorx.ErrInvalidFormat,
			map[string]string{"subject": subject},
			subject,
		))

		return false
	}

	return true
}

// extractComponents extracts all components from the subject line.
func (r *ConventionalCommitRule) extractComponents(subject string) (string, string, bool, string) {
	var commitType, scope, description string

	var hasBreaking bool

	matches := subjectRegex.FindStringSubmatch(subject)
	if len(matches) != 5 {
		r.errors = append(r.errors, errorx.NewErrorWithContext(
			r.Name(),
			errorx.ErrInvalidFormat,
			map[string]string{"subject": subject},
			subject,
		))

		return "", "", false, ""
	}

	commitType = matches[1]
	scope = matches[2]
	hasBreaking = matches[3] == "!"
	description = matches[4]

	return commitType, scope, hasBreaking, description
}

// validateType checks if the commit type is allowed.
func (r *ConventionalCommitRule) validateType(commitType string) bool {
	if len(r.allowedTypes) > 0 && !slices.Contains(r.allowedTypes, commitType) {
		r.errors = append(r.errors, errorx.NewErrorWithContext(
			r.Name(),
			errorx.ErrInvalidType,
			map[string]string{
				"type":          commitType,
				"allowed_types": strings.Join(r.allowedTypes, ","),
			},
			commitType, strings.Join(r.allowedTypes, ", "),
		))

		return false
	}

	return true
}

// validateScope checks if the commit scope is allowed.
func (r *ConventionalCommitRule) validateScope(scope string) bool {
	if scope != "" && len(r.allowedScopes) > 0 {
		scopesList := strings.Split(scope, ",")
		for _, scopeItem := range scopesList {
			if !slices.Contains(r.allowedScopes, scopeItem) {
				r.errors = append(r.errors, errorx.NewErrorWithContext(
					r.Name(),
					errorx.ErrInvalidScope,
					map[string]string{
						"scope":          scopeItem,
						"allowed_scopes": strings.Join(r.allowedScopes, ","),
					},
					scopeItem, strings.Join(r.allowedScopes, ", "),
				))

				return false
			}
		}
	}

	return true
}

// validateDescription checks if the commit description is valid.
func (r *ConventionalCommitRule) validateDescription(description string) bool {
	// Validate description content
	if strings.TrimSpace(description) == "" {
		r.errors = append(r.errors, errorx.NewValidationError(r.Name(), errorx.ErrEmptyDescription))

		return false
	}

	// Validate description length
	if len(description) > r.maxDescLength {
		r.errors = append(r.errors, errorx.NewErrorWithContext(
			r.Name(),
			errorx.ErrSubjectTooLong,
			map[string]string{
				"actual_length": strconv.Itoa(len(description)),
				"max_length":    strconv.Itoa(r.maxDescLength),
			},
			len(description), r.maxDescLength,
		))

		return false
	}

	return true
}

// Result returns a concise string representation of the validation result.
func (r *ConventionalCommitRule) Result() string {
	if len(r.errors) > 0 {
		return "Invalid conventional commit format"
	}

	return "Valid conventional commit format"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *ConventionalCommitRule) VerboseResult() string {
	if len(r.errors) > 0 {
		// Return a more detailed error message in verbose mode
		switch r.errors[0].Code {
		case "invalid_format":
			return "Invalid format: doesn't follow conventional format 'type(scope)!: description'"
		case "invalid_type":
			var allowedTypes string
			if val, ok := r.errors[0].Context["allowed_types"]; ok {
				allowedTypes = strings.ReplaceAll(val, ",", ", ")
			}

			return "Invalid type '" + r.commitType + "'. Must be one of: " + allowedTypes
		case "invalid_scope":
			var allowedScopes string
			if val, ok := r.errors[0].Context["allowed_scopes"]; ok {
				allowedScopes = strings.ReplaceAll(val, ",", ", ")
			}

			return "Invalid scope '" + r.scope + "'. Must be one of: " + allowedScopes
		case "empty_description":
			return "Missing description after type/scope"
		case "description_too_long":
			var actualLength, maxLength string
			if val, ok := r.errors[0].Context["actual_length"]; ok {
				actualLength = val
			}

			if val, ok := r.errors[0].Context["max_length"]; ok {
				maxLength = val
			}

			return "Description too long (" + actualLength + " chars). Maximum length is " + maxLength + " characters"
		case "spacing_error":
			return "Spacing error: Must have exactly one space after colon"
		default:
			return r.errors[0].Error()
		}
	}

	// Success verbose message with more details
	result := "Valid conventional commit with type '" + r.commitType + "'"
	if r.scope != "" {
		result += " and scope '" + r.scope + "'"
	}

	if r.hasBreaking {
		result += " (breaking change)"
	}

	return result
}

// Help returns guidance for fixing rule violations.
func (r *ConventionalCommitRule) Help() string {
	if len(r.errors) == 0 {
		return "No errors to fix"
	}

	// Check for specific error codes if available
	if len(r.errors) > 0 {
		switch r.errors[0].Code {
		case string(domain.ValidationErrorInvalidFormat):
			return errorx.FormatHelp(errorx.ErrInvalidFormat, "commit message", "type(scope)!: description")

		case string(domain.ValidationErrorInvalidType):
			// Get allowed types from context
			allowedTypes := ""
			if val, ok := r.errors[0].Context["allowed_types"]; ok {
				allowedTypes = strings.ReplaceAll(val, ",", ", ")
			}

			return errorx.FormatHelp(errorx.ErrInvalidType, allowedTypes)

		case string(domain.ValidationErrorInvalidScope):
			// Get allowed scopes from context
			allowedScopes := ""
			if val, ok := r.errors[0].Context["allowed_scopes"]; ok {
				allowedScopes = strings.ReplaceAll(val, ",", ", ")
			}

			return errorx.FormatHelp(errorx.ErrInvalidScope, allowedScopes)

		case string(domain.ValidationErrorEmptyDescription):
			return errorx.FormatHelp(errorx.ErrEmptyDescription)

		case string(domain.ValidationErrorDescriptionTooLong):
			// Get max length from context
			maxLength := r.maxDescLength

			if val, ok := r.errors[0].Context["max_length"]; ok {
				if parsedVal, err := strconv.Atoi(val); err == nil {
					maxLength = parsedVal
				}
			}

			return errorx.FormatHelp(errorx.ErrSubjectTooLong, maxLength)

		case string(domain.ValidationErrorSpacing):
			return errorx.FormatHelp(errorx.ErrInvalidFormat, "spacing in commit message", "type(scope)!: description")
		}
	}

	// Default help message
	return errorx.FormatHelp(errorx.ErrInvalidFormat, "commit message", "type(scope)!: description")
}

// Errors returns all validation errors.
func (r *ConventionalCommitRule) Errors() []*domain.ValidationError {
	return r.errors
}
