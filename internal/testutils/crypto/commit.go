// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package crypto provides test utilities for crypto-related functionality.
package crypto

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
)

// GetCommitBytes extracts the byte representation of a commit for signature verification.
// This is a test utility function used for testing signature verification.
func GetCommitBytes(commit *object.Commit) ([]byte, error) {
	if commit == nil {
		return nil, errors.New("nil commit")
	}

	// This is a test utility that returns a placeholder value
	// In a real implementation, you would:
	// 1. Create a memory object
	// 2. Encode the commit without its signature
	// 3. Read the bytes back
	return []byte(fmt.Sprintf("commit %s", commit.Hash)), nil
}
