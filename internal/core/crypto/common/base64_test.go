// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package common

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/common/encutils"
	"github.com/stretchr/testify/require"
)

func TestIsBase64(t *testing.T) {
	// Note: This test uses encutils.IsBase64 directly since the functions
	// have been moved to encutils package
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Standard base64",
			input:    "SGVsbG8gV29ybGQ=",
			expected: true,
		},
		{
			name:     "URL-safe base64",
			input:    "SGVsbG8_V29ybGQ=",
			expected: true,
		},
		{
			name:     "Base64 without padding",
			input:    "SGVsbG8gV29ybGQ",
			expected: true,
		},
		{
			name:     "Invalid base64",
			input:    "Not base64!",
			expected: false,
		},
		{
			name:     "Almost valid base64",
			input:    "SGVsbG8gV29ybGQ=!",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := encutils.IsBase64(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestDecodeBase64(t *testing.T) {
	// Note: This test uses encutils.DecodeBase64 directly since the functions
	// have been moved to encutils package
	tests := []struct {
		name         string
		input        string
		expected     []byte
		shouldDecode bool
	}{
		{
			name:         "Empty string",
			input:        "",
			expected:     nil,
			shouldDecode: false,
		},
		{
			name:         "Standard base64",
			input:        "SGVsbG8gV29ybGQ=",
			expected:     []byte("Hello World"),
			shouldDecode: true,
		},
		{
			name:         "URL-safe base64",
			input:        "SGVsbG8_V29ybGQ=",
			expected:     []byte("Hello?World"),
			shouldDecode: true,
		},
		{
			name:         "Base64 without padding",
			input:        "SGVsbG8gV29ybGQ",
			expected:     []byte("Hello World"),
			shouldDecode: true,
		},
		{
			name:         "Invalid base64",
			input:        "Not base64!",
			expected:     nil,
			shouldDecode: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, success := encutils.DecodeBase64(testCase.input)
			require.Equal(t, testCase.shouldDecode, success)

			if testCase.shouldDecode {
				require.Equal(t, testCase.expected, result)
			}
		})
	}
}

func TestIsArmored(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name: "GPG signature block",
			input: `-----BEGIN PGP SIGNATURE-----
Version: GnuPG v2.0.22 (GNU/Linux)

iQIcBAABAgAGBQJY3+HhAAoJEJvV+YPmBGUQ2dIQAJM3xrgMr1QVH6MttR0vaMaS
v9CurJ6xfbX7vn7BdcZY5ShQZH5CJAqK1JJnvVKjLQyB9S9JYdcWK3D8G5JfWJaF
Q==
-----END PGP SIGNATURE-----`,
			expected: true,
		},
		{
			name:     "SSH public key",
			input:    `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDE4R8arP user@example.com`,
			expected: false,
		},
		{
			name: "SSH signature block",
			input: `-----BEGIN SSH SIGNATURE-----
U1NIU0lHAAAAAQAAADMAAAALc3NoLWVkMjU1MTkAAAAg0+/s2JSj4+kZxJKkZIFtAEp
N3n4HR8xiknrpOKpgJHMAAAAHZ2l0QDIuMzQBDQ==
-----END SSH SIGNATURE-----`,
			expected: true,
		},
		{
			name:     "Plain text",
			input:    "This is just a regular text without any signature markers.",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := IsArmored(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}
