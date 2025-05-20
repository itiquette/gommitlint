// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package encutils

import (
	"encoding/base64"
)

// IsBase64 checks if a string is valid base64-encoded data by attempting
// to decode it using various base64 encoding schemes.
func IsBase64(input string) bool {
	if input == "" {
		return false
	}

	// Try standard base64 encoding
	_, err := base64.StdEncoding.DecodeString(input)
	if err == nil {
		return true
	}

	// Try URL encoding
	_, err = base64.URLEncoding.DecodeString(input)
	if err == nil {
		return true
	}

	// Try RawStdEncoding (no padding)
	_, err = base64.RawStdEncoding.DecodeString(input)
	if err == nil {
		return true
	}

	// Try RawURLEncoding (no padding)
	_, err = base64.RawURLEncoding.DecodeString(input)

	return err == nil
}

// DecodeBase64 attempts to decode a base64 string using multiple encoding variants.
// It returns the decoded bytes and true if successful, or nil and false if not.
func DecodeBase64(input string) ([]byte, bool) {
	if input == "" {
		return nil, false
	}

	// Try standard base64 encoding
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err == nil {
		return decoded, true
	}

	// Try URL encoding
	decoded, err = base64.URLEncoding.DecodeString(input)
	if err == nil {
		return decoded, true
	}

	// Try RawStdEncoding (no padding)
	decoded, err = base64.RawStdEncoding.DecodeString(input)
	if err == nil {
		return decoded, true
	}

	// Try RawURLEncoding (no padding)
	decoded, err = base64.RawURLEncoding.DecodeString(input)
	if err == nil {
		return decoded, true
	}

	return nil, false
}
