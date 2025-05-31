// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package format

import (
	"context"
)

// Formatter converts validation results to different output formats.
// This is an adapter interface for presentation concerns.
type Formatter interface {
	// Format converts validation results to a formatted string
	Format(ctx context.Context, results interface{}) string

	// ContentType returns the MIME type of the formatted output
	ContentType() string
}
