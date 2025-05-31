// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package crypto

import (
	"strings"
)

// IsArmored checks if a string appears to be an armored (PEM-like) block
// containing base64 data between header and footer lines.
// This is commonly used for GPG and SSH keys and signatures.
func IsArmored(input string) bool {
	input = strings.ToLower(input)

	// Check for BEGIN/END markers characteristic of PEM blocks
	hasBegin := strings.Contains(input, "-----begin")
	hasEnd := strings.Contains(input, "-----end")

	if hasBegin && hasEnd {
		return true
	}

	// Check for other common armor indicators
	markers := []string{"pgp", "signature", "public key", "private key", "ssh"}
	markerCount := 0

	for _, marker := range markers {
		if strings.Contains(input, marker) {
			markerCount++
			if markerCount >= 1 && (hasBegin || hasEnd) {
				return true
			}
		}
	}

	return false
}
