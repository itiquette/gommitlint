// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package format

import (
	"encoding/json"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
)

// JSONFailure represents a rule failure in JSON format.
type JSONFailure struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Help    string `json:"help,omitempty"`
}

// JSONResult represents a commit validation result in JSON format.
type JSONResult struct {
	Hash     string        `json:"hash"`
	Subject  string        `json:"subject"`
	Passed   bool          `json:"passed"`
	Failures []JSONFailure `json:"failures,omitempty"`
}

// ConvertValidationResultsToJSON converts ValidationResult slice to JSON string.
// This is a pure function that handles JSON serialization with proper error handling.
func ConvertValidationResultsToJSON(results []domain.ValidationResult) string {
	jsonResults := make([]JSONResult, len(results))

	for index, result := range results {
		jsonResults[index] = convertSingleResult(result)
	}

	return marshalToJSON(jsonResults)
}

// convertSingleResult converts a single ValidationResult to JSONResult.
// Pure function with no side effects.
func convertSingleResult(result domain.ValidationResult) JSONResult {
	jsonResult := JSONResult{
		Hash:    result.Commit.Hash,
		Subject: result.Commit.Subject,
		Passed:  result.Passed(),
	}

	// Convert failures
	jsonResult.Failures = make([]JSONFailure, len(result.Failures))
	for i, failure := range result.Failures {
		jsonResult.Failures[i] = JSONFailure{
			Rule:    failure.Rule,
			Message: failure.Message,
			Help:    failure.Help,
		}
	}

	return jsonResult
}

// marshalToJSON marshals data to JSON with error handling.
// Pure function that returns formatted JSON or error message.
func marshalToJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}

	return string(jsonBytes)
}
