// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// VerificationStatus represents the outcome of a signature verification.
type VerificationStatus string

// Verification status constants.
const (
	VerificationStatusVerified    VerificationStatus = "verified"
	VerificationStatusFailed      VerificationStatus = "failed"
	VerificationStatusNoKey       VerificationStatus = "no_key"
	VerificationStatusUnsupported VerificationStatus = "unsupported"
)

// VerificationResult represents the result of a signature verification.
// This is an immutable value type.
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

// WithError adds error information to a verification result and returns a new instance.
func (v VerificationResult) WithError(code, message string) VerificationResult {
	result := v
	result.errorCode = code
	result.errorMsg = message

	return result
}

// Status returns the verification status.
func (v VerificationResult) Status() VerificationStatus {
	return v.status
}

// Identity returns the verified identity.
func (v VerificationResult) Identity() Identity {
	return v.identity
}

// Signature returns the signature that was verified.
func (v VerificationResult) Signature() Signature {
	return v.signature
}

// ErrorCode returns the error code, if any.
func (v VerificationResult) ErrorCode() string {
	return v.errorCode
}

// ErrorMessage returns the error message, if any.
func (v VerificationResult) ErrorMessage() string {
	return v.errorMsg
}

// IsVerified returns true if the signature was successfully verified.
func (v VerificationResult) IsVerified() bool {
	return v.status == VerificationStatusVerified
}

// HasError returns true if there is an error code or message.
func (v VerificationResult) HasError() bool {
	return v.errorCode != "" || v.errorMsg != ""
}

// Verifier defines the interface for signature verification.
type Verifier interface {
	// CanVerify checks if this verifier can handle the given signature.
	CanVerify(signature Signature) bool

	// Verify checks if the signature is valid for the given data and returns
	// the verification result with the associated identity.
	Verify(signature Signature, data []byte, keyDir string) VerificationResult
}
