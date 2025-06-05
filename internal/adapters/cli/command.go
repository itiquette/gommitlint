// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/adapters/git"
	format "github.com/itiquette/gommitlint/internal/adapters/output"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
)

// ValidationCommand represents a validation request as a pure value type.
// This follows hexagonal architecture with explicit dependencies.
type ValidationCommand struct {
	Target           ValidationTarget
	Output           OutputOptions
	Config           config.Config
	RepoPath         string
	SkipMergeCommits bool
}

// Execute performs validation using functional composition.
// This is the main orchestration method that coordinates hexagonal adapters.
func (cmd ValidationCommand) Execute(ctx context.Context) error {
	// Create repository adapter (hexagonal boundary)
	repo, err := git.NewRepository(ctx, cmd.RepoPath)
	if err != nil {
		return fmt.Errorf("repository creation failed: %w", err)
	}

	// Create rules using domain factory (pure function)
	validationRules := rules.CreateEnabledRules(&cmd.Config)

	// Create formatter adapter (hexagonal boundary)
	formatter := cmd.Output.CreateFormatter()
	generator := format.NewGenerator(formatter, cmd.Output.Writer)

	// Execute validation using pure domain functions
	report, err := validateTarget(ctx, cmd.Target, validationRules, repo, &cmd.Config, cmd.SkipMergeCommits)
	if err != nil {
		return err
	}

	// Generate output using adapter
	err = generator.GenerateFromReport(ctx, report)
	if err != nil {
		return err
	}

	// Return validation failure if rules failed
	if !report.Summary.AllPassed {
		return fmt.Errorf("validation failed: %d of %d commits passed", report.Summary.PassedCommits, report.Summary.TotalCommits)
	}

	return nil
}
