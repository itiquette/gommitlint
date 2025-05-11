// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configtestutils

// GetConfigTestUtils returns the test utils instance for easy import.
// This function is not needed functionally, but helps prevent import usage errors
// in tests where the test package name conflicts with the actual package name.
func GetConfigTestUtils() *TestUtils {
	return &TestUtils{}
}

// TestUtils is a simple struct that allows functions to be called in a more
// readable way in tests.
type TestUtils struct{}
