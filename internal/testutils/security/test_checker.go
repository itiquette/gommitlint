// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package security provides test utilities for security tests and mocks.
package security

import (
	"github.com/itiquette/gommitlint/internal/common/security"
)

// NewTestSecurityService creates a new FileSecurityService for testing.
// In tests, we should use proper file permissions instead of bypassing checks.
func NewTestSecurityService() *security.FileSecurityService {
	return &security.FileSecurityService{}
}
