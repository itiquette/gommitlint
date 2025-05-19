// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validate provides the application service for commit validation.
// It orchestrates the validation process by coordinating between the domain
// layer (validation engine, rules) and the infrastructure adapters (git repository).
// This service implements the primary use case of validating commit messages
// against configured rules, handling error reporting, and managing the validation
// workflow. The package serves as the application layer in the hexagonal
// architecture, translating between external requests and domain logic.
package validate
