// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// RuleDependencies holds external dependencies needed by some rules.
type RuleDependencies struct {
	CryptoVerifier   CryptoVerifier
	CryptoRepository CryptoKeyRepository
	Repository       Repository
}
