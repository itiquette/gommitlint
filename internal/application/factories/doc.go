// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package factories provides factory implementations for creating domain objects.

This package contains factories that instantiate domain rules and other
objects with the appropriate configuration. It abstracts the complexity
of object creation and dependency injection from the rest of the application.

The factories handle:
  - Rule instantiation with proper configuration
  - Dependency injection for rules requiring external services
  - Grouping of related rules by their requirements

All factories follow the factory pattern and return domain interfaces,
ensuring that the application layer remains decoupled from concrete
implementations.
*/
package factories
