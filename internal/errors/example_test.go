// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors_test

import (
	"fmt"
	"strconv"

	"github.com/itiquette/gommitlint/internal/errors"
)

// SubjectLengthRule represents a rule that checks subject length.
type SubjectLengthRule struct {
	MaxLength int
}

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return "SubjectLength"
}

// Help returns helpful guidance for the rule.
func (r SubjectLengthRule) Help() string {
	return fmt.Sprintf("Ensure the subject line is at most %d characters long", r.MaxLength)
}

// CommitInfo represents basic information about a commit.
type CommitInfo struct {
	Hash    string
	Subject string
	Body    string
}

// ExampleEnhancedValidationError demonstrates how to create and use enhanced validation errors.
func ExampleEnhanceValidationError() {
	// Create a sample rule
	rule := SubjectLengthRule{MaxLength: 50}

	// Create a sample commit
	commit := CommitInfo{
		Hash:    "abc123",
		Subject: "This is a very long subject line that exceeds the maximum length",
		Body:    "This is the commit body",
	}

	// Create a context with commit information
	ctx := errors.NewContext().WithCommit(
		commit.Hash,
		commit.Subject+"\n\n"+commit.Body,
		commit.Subject,
		commit.Body,
	)

	// Create an enhanced validation error
	subjectLength := len(commit.Subject)
	errorMessage := fmt.Sprintf("Subject length (%d) exceeds maximum allowed (%d)", subjectLength, rule.MaxLength)

	// Option 1: Create directly with all context information
	validationErr := errors.CreateRichError(
		rule.Name(),
		errors.ErrSubjectTooLong,
		errorMessage,
		rule.Help(), // Use the rule's help method
		ctx,
	)

	// Option 2: Create step by step
	baseErr := errors.New(rule.Name(), errors.ErrSubjectTooLong, errorMessage)
	enhancedErr := errors.EnhanceValidationError(baseErr, rule.Help())
	enhancedErr = enhancedErr.WithContext("subject_length", strconv.Itoa(subjectLength))
	enhancedErr = enhancedErr.WithContext("max_length", strconv.Itoa(rule.MaxLength))
	enhancedErr = enhancedErr.WithContext("subject", commit.Subject)
	enhancedErr = enhancedErr.WithCommitSHA(commit.Hash)

	// Use the validation error to demonstrate key features

	// 1. Error message formatting
	fmt.Printf("Basic error message: %s\n", validationErr.Error())

	// 2. Get help message
	fmt.Printf("Help: %s\n", validationErr.GetHelp())

	// 3. Access context values
	fmt.Printf("Subject length: %s\n", enhancedErr.Context["subject_length"])
	fmt.Printf("Maximum length: %s\n", enhancedErr.Context["max_length"])

	// Output:
	// Basic error message: Subject length (64) exceeds maximum allowed (50)
	// Help: Ensure the subject line is at most 50 characters long
	// Subject length: 64
	// Maximum length: 50
}
