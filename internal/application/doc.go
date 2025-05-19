// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package application contains the application services layer of gommitlint.

Following hexagonal architecture principles, this layer orchestrates
business operations by coordinating between the domain core and the
external adapters. It contains:

  - Service implementations that compose domain logic
  - Factories for creating domain objects
  - Report generation logic
  - Option definitions for application-level concerns

The application layer ensures that:
  - Domain logic remains pure and focused
  - Adapters stay independent of business logic
  - Use cases are properly orchestrated
  - Cross-cutting concerns are handled appropriately

Key packages:

  - factories: Creates domain objects with proper configuration
  - options: Defines application-level options and settings
  - report: Handles report generation and formatting
  - validate: Orchestrates validation workflows

All application services follow value semantics and functional patterns,
maintaining immutability and clear data flow throughout the system.
*/
package application
