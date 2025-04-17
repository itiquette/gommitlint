// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signedidentityrule

import (
	"encoding/base64"
)

// isBase64 checks if a string is valid base64-encoded data.
//
// Parameters:
//   - str: The string to check for base64 validity
//
// The function attempts to decode the string using multiple base64 encoding variants:
//  1. Standard base64 encoding (with padding)
//  2. URL-safe base64 encoding (with padding)
//  3. Raw standard base64 encoding (without padding)
//  4. Raw URL-safe base64 encoding (without padding)
//
// This comprehensive approach handles various base64 formats that might be used in
// different contexts, such as web applications (URL-safe) or space-optimized formats
// (raw/no padding).
//
// Returns:
//   - bool: true if the string can be decoded as base64 using any of the supported
//     encodings, false otherwise (including for empty strings)
func isBase64(str string) bool {
	if str == "" {
		return false
	}

	// Try standard base64 encoding
	_, err := base64.StdEncoding.DecodeString(str)
	if err == nil {
		return true
	}

	// Try URL encoding
	_, err = base64.URLEncoding.DecodeString(str)
	if err == nil {
		return true
	}

	// Try RawStdEncoding (no padding)
	_, err = base64.RawStdEncoding.DecodeString(str)
	if err == nil {
		return true
	}

	// Try RawURLEncoding (no padding)
	_, err = base64.RawURLEncoding.DecodeString(str)

	return err == nil
}
