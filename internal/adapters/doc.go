// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package adapters contains the infrastructure adapters that implement the ports defined by the core domain.

Following hexagonal architecture principles, this package contains:

  - cli: Command-line interface adapter (driving adapter)
  - git: Git repository adapter (driven adapter)
  - loader: Configuration loading adapter (driven adapter)
  - logging: Logging adapter (driven adapter)
  - output: Output formatting adapter (driven adapter)
  - signing: Cryptographic signing adapter (driven adapter)

Adapters translate between the external world and the core domain, ensuring that
the domain remains independent of infrastructure concerns. Each adapter implements
one or more port interfaces, providing concrete implementations while keeping
the core domain pure and testable.

The adapter pattern allows the application to be driven by different interfaces
(CLI, API) and to drive different infrastructure components (Git, file system,
crypto libraries) without coupling the business logic to these external concerns.
*/
package adapters
