// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package security

// Checker represents a service for checking file security.
// This interface provides methods for validating permissions and ownership.
type Checker interface {
	IsSecurePermissions(path string, level Level) (bool, error)
	VerifyOwnership(path string) (bool, error)
}

// SecurityChecker is an alias for Checker for backward compatibility.
// This is kept for compatibility but new code should use Checker directly.
//
// Deprecated: Use Checker instead to avoid stuttering in the type name.
//
//nolint:revive // Kept for backward compatibility
type SecurityChecker = Checker

// TestSecurityChecker is a security checker that always returns true for testing.
type TestSecurityChecker struct{}

// NewTestSecurityChecker creates a new TestSecurityChecker for testing.
// This should only be used in tests as it bypasses security checks.
func NewTestSecurityChecker() *TestSecurityChecker {
	return &TestSecurityChecker{}
}

// IsSecurePermissions always returns true for the test security checker.
// It ignores the path and level parameters and always passes the security check.
func (s *TestSecurityChecker) IsSecurePermissions(_ string, _ Level) (bool, error) {
	return true, nil
}

// VerifyOwnership always returns true for the test security checker.
// It ignores the path parameter and always passes the ownership check.
func (s *TestSecurityChecker) VerifyOwnership(_ string) (bool, error) {
	return true, nil
}
