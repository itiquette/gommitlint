// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package main

import (
	"github.com/itiquette/gommitlint/internal/core/rules"
)

import (
	"fmt"
	"log"

	"github.com/golangci/misspell"
)

// This file is to help test the misspell library directly.
// Use "go run spell_test_main.go" to check what misspellings the library detects.

func main() {
	// Initialize the spell checker
	replacer := misspell.New()
	replacer.Compile()

	// Test various misspellings
	testWords := []string{
		"receive",  // should be "receive"
		"feture",   // should be "feature"
		"the",      // should be "the"
		"their",    // should be "their"
		"February", // should be "February"
	}

	for _, word := range testWords {
		corrected, diffs := replacer.Replace(word)
		if corrected != word {
			fmt.Printf("Found misspelling: %q should be %q\n", word, corrected)

			for _, diff := range diffs {
				fmt.Printf("  Diff: %q -> %q\n", diff.Original, diff.Corrected)
			}
		} else {
			fmt.Printf("No misspelling found for: %q\n", word)
		}
	}

	// Let's try with a full sentence
	sentence := "I need to receive the feture list from the team."
	corrected, diffs := replacer.Replace(sentence)

	if corrected != sentence {
		fmt.Printf("\nFound misspellings in sentence:\n")
		fmt.Printf("Original: %q\n", sentence)
		fmt.Printf("Corrected: %q\n", corrected)

		for _, diff := range diffs {
			fmt.Printf("  Diff: %q -> %q\n", diff.Original, diff.Corrected)
		}
	} else {
		fmt.Printf("\nNo misspellings found in sentence: %q\n", sentence)
	}

	// Try directly with NewSpellRule
	_ = rules.NewSpellRule()

	log.Printf("Created spell rule")
}
