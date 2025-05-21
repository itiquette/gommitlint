// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package encutils

import (
	"encoding/base64"
)

// IsBase64 checks if a string is valid base64-encoded data.
// It attempts to decode the input with standard base64 encodings.
func IsBase64(input string) bool {
	if input == "" {
		return false
	}

	// Try all standard encodings
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	}

	for _, enc := range encodings {
		if _, err := enc.DecodeString(input); err == nil {
			return true
		}
	}

	return false
}

// DecodeBase64 attempts to decode a base64 string with standard encodings.
// It returns the decoded bytes and true if successful, or nil and false if not.
func DecodeBase64(input string) ([]byte, bool) {
	if input == "" {
		return nil, false
	}

	// Try all standard encodings
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	}

	for _, enc := range encodings {
		if decoded, err := enc.DecodeString(input); err == nil {
			return decoded, true
		}
	}

	return nil, false
}
