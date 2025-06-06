// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package integrationtest contains end-to-end integration tests for gommitlint workflows.

These tests verify that the application's components work together correctly
in realistic scenarios. They test complete workflows including:
  - CLI command execution
  - Configuration loading
  - Git repository operations
  - Validation rule execution
  - Output formatting

# Test Organization

  - branch_comparison_test.go: Tests branch comparison functionality
  - cli_workflow_test.go: Tests the CLI validate command end-to-end
  - comprehensive_test.go: Tests comprehensive validation scenarios
  - context_config_test.go: Tests context-based configuration workflow
  - crypto_workflow_test.go: Tests cryptographic verification workflows
  - rulepriority_integration_test.go: Tests rule priority system
  - validation_workflow_test.go: Tests core validation workflow

# Important Notes

THIS PACKAGE CONTAINS TEST-ONLY CODE AND SHOULD NOT BE IMPORTED BY PRODUCTION CODE.

Integration tests are automatically skipped in environments without Git.
The tests check for git availability using the git command.

These tests create temporary Git repositories and configuration files to test
real-world scenarios.
*/
package integrationtest
