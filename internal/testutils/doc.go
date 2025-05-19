// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package testutils contains test utilities and helpers for gommitlint tests.
// This package provides essential test utilities for creating test contexts,
// configurations, and logging capabilities.
//
// # THIS PACKAGE CONTAINS TEST-ONLY CODE AND SHOULD NOT BE IMPORTED BY PRODUCTION CODE
//
// # Structure
//
// The package is organized into focused sub-packages:
//
// - config: Configuration builders and adapters for test scenarios
// - context: Context creation utilities with preloaded test configurations
// - logger: Test logger implementations including print and zerolog adapters
// - git: Git repository test helpers for integration testing
// - domain: Domain test utilities including commit builders, context keys, and validation helpers
//
// # Usage Patterns
//
// The test utilities follow consistent patterns:
//
//  1. Context-based configuration:
//     ctx := testcontext.CreateTestContext()
//     ctx = contextx.WithConfig(ctx, testConfig)
//
//  2. Configuration building:
//     cfg := testconfig.NewBuilder().
//     WithSubjectMaxLength(50).
//     Build()
//
//  3. Test logging:
//     logger := testlogger.NewTestLogger()
//
// # Design Principles
//
// - Simplicity: Utilities are focused and easy to use
// - Consistency: Similar patterns across all test packages
// - Integration: Works seamlessly with testify and standard testing
// - Type Safety: Leverages Go's type system for compile-time safety
package testutils
