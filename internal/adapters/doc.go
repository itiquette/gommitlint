// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package adapters contains the infrastructure adapters that implement domain interfaces.

Following functional hexagonal architecture principles, this package contains:

  - cli: Command-line interface adapter (primary/driving adapter)
  - config: Configuration loading adapter (secondary/driven adapter)
  - git: Git repository adapter (secondary/driven adapter)
  - logging: Logging adapter (secondary/driven adapter)
  - output: Output formatting adapter (secondary/driven adapter)
  - signing: Cryptographic verification adapter (secondary/driven adapter)

Adapters use value semantics and pure functions to translate between the external
world and the core domain. Each adapter implements domain interfaces while keeping
the domain logic completely independent of infrastructure concerns.

All adapters follow these principles:
  - Value receivers and immutable data structures
  - Explicit dependencies passed as parameters
  - No hidden global state or service locators
  - Pure functional composition where possible
*/
package adapters
