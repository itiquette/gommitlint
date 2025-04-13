// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package model

// Severity levels for validation errors
//  // VerifySignatureIdentity checks if a commit is signed with a trusted key.
// func VerifySignatureIdentity(commit *object.Commit, signature string, keyDir string) SignedIdentity {
//     rule := SignedIdentity{}

//     if commit == nil {
//         err := NewValidationError(
//             "SignedIdentity",
//             "nil_commit",
//             "Commit cannot be nil",
//         )
//         rule.errors = append(rule.errors, err)
//         return rule
//     }

//     if keyDir == "" {
//         err := NewValidationError(
//             "SignedIdentity",
//             "missing_key_dir",
//             "No key directory provided",
//         ).WithContext("required", "true")

//         rule.errors = append(rule.errors, err)
//         return rule
//     }

//	    // Continue with validation...
//	    return rule
//	}
//
// Example of how a caller might handle these structured errors
// func checkCommitMessage(message string) {
// 	rule := ValidateSubjectLength(message, 100)

// 	for _, err := range rule.Errors() {
// 		if validErr, ok := err.(*ValidationError); ok {
// 			// Handle based on structured data
// 			switch validErr.Code {
// 			case "subject_too_long":
// 				actual := validErr.Context["actual"]
// 				max := validErr.Context["maximum"]
// 				fmt.Printf("Subject length (%s) exceeds maximum (%s)\n", actual, max)

// 			default:
// 				fmt.Println(validErr.Message)
// 			}
// 		} else {
// 			// Handle regular errors
// 			fmt.Println(err.Error())
// 		}
// 	}
// }

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"
)

// ValidationError represents a structured error from rule validation.
type ValidationError struct {
	Code     string            // Unique error identifier
	Message  string            // Human-readable description
	Rule     string            // Rule that generated the error
	Severity string            // "error", "warning", or "info"
	Context  map[string]string // Additional error details
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error with default severity "error".
func NewValidationError(rule, code, message string) *ValidationError {
	return &ValidationError{
		Code:     code,
		Message:  message,
		Rule:     rule,
		Severity: SeverityError,
		Context:  make(map[string]string),
	}
}

// WithContext adds context information to the error.
func (e *ValidationError) WithContext(key, value string) *ValidationError {
	e.Context[key] = value

	return e
}

// WithSeverity sets the severity level.
func (e *ValidationError) WithSeverity(severity string) *ValidationError {
	e.Severity = severity

	return e
}
