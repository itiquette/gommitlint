// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package testdata provides test fixtures and helpers for crypto adapter tests.
package testdata

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CryptoCommit represents a test commit with crypto information.
type CryptoCommit struct {
	Hash      string
	Message   string
	Author    string
	Email     string
	Signature string
	Key       string
}

// NewCryptoCommit creates a new crypto commit for testing.
func NewCryptoCommit(hash, message, author, email, signature, key string) CryptoCommit {
	return CryptoCommit{
		Hash:      hash,
		Message:   message,
		Author:    author,
		Email:     email,
		Signature: signature,
		Key:       key,
	}
}


// CreateTempKeyDir creates a temporary directory for test keys.
func CreateTempKeyDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "gommitlint-keys-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// WriteTestKey writes a test key file.
func WriteTestKey(t *testing.T, dir, filename, content string) string {
	t.Helper()

	keyPath := filepath.Join(dir, filename)
	err := os.WriteFile(keyPath, []byte(content), 0600)
	require.NoError(t, err)

	return keyPath
}

// IsBase64 checks if a string is valid base64 (standard or URL-safe).
func IsBase64(str string) bool {
	if str == "" {
		return false
	}

	// Try standard base64
	_, err := base64.StdEncoding.DecodeString(str)
	if err == nil {
		return true
	}

	// Try URL-safe base64
	_, err = base64.URLEncoding.DecodeString(str)
	if err == nil {
		return true
	}

	// Try without padding
	_, err = base64.RawStdEncoding.DecodeString(str)
	if err == nil {
		return true
	}

	_, err = base64.RawURLEncoding.DecodeString(str)

	return err == nil
}

// DecodeBase64 attempts to decode a base64 string.
func DecodeBase64(str string) ([]byte, bool) {
	if str == "" {
		return nil, false
	}

	// Try standard base64
	data, err := base64.StdEncoding.DecodeString(str)
	if err == nil {
		return data, true
	}

	// Try URL-safe base64
	data, err = base64.URLEncoding.DecodeString(str)
	if err == nil {
		return data, true
	}

	// Try without padding
	data, err = base64.RawStdEncoding.DecodeString(str)
	if err == nil {
		return data, true
	}

	data, err = base64.RawURLEncoding.DecodeString(str)
	if err == nil {
		return data, true
	}

	return nil, false
}