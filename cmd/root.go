// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package cmd

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func newRootCommand(_ context.Context, versionString string) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "gommitlint",
		Version: versionString,
		Short:   "Commit validator.",
		Long:    ``,
	}

	rootCmd.AddCommand(newValidateCmd())

	return rootCmd
}

func Execute(version, commitSHA, buildDate string) {
	ctx := context.Background()

	versionString := version + " (Commit SHA: " + commitSHA + ", Build date: " + buildDate + ")"
	err := newRootCommand(ctx, versionString).Execute()
	HandleError(ctx, err)
}

func HandleError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	logger := zerolog.Ctx(ctx)
	logger.Error().Err(err).Msg("An error occurred")

	os.Exit(1)
}
