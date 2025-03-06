package app

import (
	"fmt"

	"github.com/weslien/unregex/internal/format"
)

// Run executes the main application logic
func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no regex pattern provided")
	}

	pattern := args[0]

	// Get the format from args or default to "go"
	formatName := "go"
	if len(args) > 1 {
		formatName = args[1]
	}

	return ExplainRegex(pattern, formatName)
}

// ExplainRegex parses and explains a regex pattern
func ExplainRegex(pattern, formatName string) error {
	// Get the appropriate regex format implementation
	regexFormat := format.GetFormat(formatName)

	fmt.Printf("Analyzing regex pattern: %s\n", pattern)
	fmt.Printf("Format: %s\n\n", regexFormat.Name())

	// Get features supported by this format
	printSupportedFeatures(regexFormat)

	// Tokenize and explain the pattern
	tokens := regexFormat.TokenizeRegex(pattern)

	for i, token := range tokens {
		explanation := regexFormat.ExplainToken(token)
		fmt.Printf("%d. %s: %s\n", i+1, token, explanation)
	}

	fmt.Println("\nNOTE: This is a basic regex explainer. Some complex patterns might not be perfectly tokenized.")

	return nil
}

// printSupportedFeatures prints a summary of features supported by the format
func printSupportedFeatures(regexFormat format.RegexFormat) {
	features := []struct {
		name        string
		code        string
		description string
	}{
		{name: "Lookahead", code: format.FeatureLookahead, description: "(?=pattern) or (?!pattern)"},
		{name: "Lookbehind", code: format.FeatureLookbehind, description: "(?<=pattern) or (?<!pattern)"},
		{name: "Named Groups", code: format.FeatureNamedGroup, description: "(?P<name>pattern)"},
		{name: "Atomic Groups", code: format.FeatureAtomicGroup, description: "(?>pattern)"},
		{name: "Conditionals", code: format.FeatureConditional, description: "(?(cond)then|else)"},
		{name: "Possessive Quantifiers", code: format.FeaturePossessive, description: "a++, a*+, a?+"},
		{name: "Unicode Properties", code: format.FeatureUnicodeClass, description: "\\p{Property}"},
		{name: "Recursion", code: format.FeatureRecursion, description: "(?R) or (?0)"},
		{name: "Backreferences", code: format.FeatureBackreference, description: "\\1, \\2, etc."},
		{name: "Named Backreferences", code: format.FeatureNamedBackref, description: "\\k<name>"},
	}

	fmt.Println("Supported Features:")

	for _, feature := range features {
		supported := "✗"
		if regexFormat.HasFeature(feature.code) {
			supported = "✓"
		}
		fmt.Printf("  %s %s (%s)\n", supported, feature.name, feature.description)
	}

	fmt.Println()
}
