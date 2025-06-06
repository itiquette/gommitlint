// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
)

// SignatureVerifier defines the interface for signature verification.
type SignatureVerifier interface {
	// VerifyCommit verifies a commit's signature and returns verification result.
	VerifyCommit(ctx context.Context, commit Commit, keyDir string) VerificationResult
}

// VerificationStatus represents the status of signature verification.
type VerificationStatus string

const (
	// VerificationStatusVerified indicates the signature was successfully verified.
	VerificationStatusVerified VerificationStatus = "verified"

	// VerificationStatusFailed indicates the signature verification failed.
	VerificationStatusFailed VerificationStatus = "failed"

	// VerificationStatusUnsupported indicates the signature type is not supported.
	VerificationStatusUnsupported VerificationStatus = "unsupported"

	// VerificationStatusNoKey indicates no key was found for verification.
	VerificationStatusNoKey VerificationStatus = "no_key"
)

// VerificationResult represents the result of signature verification.
type VerificationResult struct {
	status    VerificationStatus
	identity  Identity
	signature Signature
	errorCode string
	errorMsg  string
}

// NewVerificationResult creates a new verification result.
func NewVerificationResult(status VerificationStatus, identity Identity, signature Signature) VerificationResult {
	return VerificationResult{
		status:    status,
		identity:  identity,
		signature: signature,
	}
}

// Status returns the verification status.
func (r VerificationResult) Status() VerificationStatus {
	return r.status
}

// Identity returns the verified identity.
func (r VerificationResult) Identity() Identity {
	return r.identity
}

// Signature returns the signature that was verified.
func (r VerificationResult) Signature() Signature {
	return r.signature
}

// IsVerified returns true if the signature was successfully verified.
func (r VerificationResult) IsVerified() bool {
	return r.status == VerificationStatusVerified
}

// HasError returns true if there was an error during verification.
func (r VerificationResult) HasError() bool {
	return r.errorCode != ""
}

// ErrorCode returns the error code if there was an error.
func (r VerificationResult) ErrorCode() string {
	return r.errorCode
}

// ErrorMessage returns the error message if there was an error.
func (r VerificationResult) ErrorMessage() string {
	return r.errorMsg
}

// WithError returns a new VerificationResult with error information.
func (r VerificationResult) WithError(code, message string) VerificationResult {
	result := r // Copy
	result.errorCode = code
	result.errorMsg = message

	return result
}
