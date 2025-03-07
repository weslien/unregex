package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/weslien/unregex/internal/format"
)

// ANSI color codes
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorBold    = "\033[1m"
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

	// Check if visualization is enabled
	visualize := false
	if len(args) > 2 && args[2] == "true" {
		visualize = true
	}

	return ExplainRegex(pattern, formatName, visualize)
}

// ExplainRegex parses and explains a regex pattern
func ExplainRegex(pattern, formatName string, visualize bool) error {
	// Get the appropriate regex format implementation
	regexFormat := format.GetFormat(formatName)

	fmt.Printf("%sAnalyzing regex pattern:%s %s\n", colorBold, colorReset, pattern)
	fmt.Printf("Format: %s\n\n", regexFormat.Name())

	// Get features supported by this format
	printSupportedFeatures(regexFormat)

	// Tokenize and explain the pattern
	tokens := regexFormat.TokenizeRegex(pattern)

	// Create a map to rotate through colors for each token
	colorMap := []string{colorRed, colorGreen, colorBlue, colorYellow, colorMagenta, colorCyan}

	// Print the explanations
	fmt.Printf("%sToken explanations:%s\n", colorBold, colorReset)
	explanations := make([]string, len(tokens))
	for i, token := range tokens {
		color := colorMap[i%len(colorMap)]
		explanation := regexFormat.ExplainToken(token)
		explanations[i] = explanation
		fmt.Printf("%s%s%d.%s %s%s%s%s: %s\n",
			color, colorBold, i+1, colorReset,
			color, colorBold, token, colorReset,
			explanation)
	}

	// If visualization is enabled, print the annotated pattern
	if visualize {
		fmt.Println()
		annotatedPattern := visualizePattern(pattern, tokens)
		fmt.Println(annotatedPattern)
	}

	fmt.Println("\nNOTE: This is a basic regex explainer. Some complex patterns might not be perfectly tokenized.")

	return nil
}

// visualizePattern creates an annotated representation of the regex with numbers
func visualizePattern(pattern string, tokens []string) string {
	// First, generate a colored version of the pattern with token boundaries
	var coloredPattern strings.Builder
	var annotationLine strings.Builder
	var legendLine strings.Builder

	// Keep track of position in the pattern
	pos := 0

	// Create a map to rotate through colors for each token
	colorMap := []string{colorRed, colorGreen, colorBlue, colorYellow, colorMagenta, colorCyan}

	// Process each token
	for i, token := range tokens {
		// Find the token in the pattern starting from current position
		tokenPos := strings.Index(pattern[pos:], token)
		if tokenPos != -1 {
			tokenPos += pos // Adjust for the slice start

			// Add any text before this token (should be empty in most cases)
			if tokenPos > pos {
				coloredPattern.WriteString(pattern[pos:tokenPos])
				for j := pos; j < tokenPos; j++ {
					annotationLine.WriteString(" ")
				}
			}

			// Add the colored token
			color := colorMap[i%len(colorMap)]
			coloredPattern.WriteString(color + colorBold + token + colorReset)

			// Add the token number in the annotation line
			marker := strconv.Itoa(i + 1)
			padding := strings.Repeat(" ", (len(token)-len(marker))/2)
			annotationLine.WriteString(color + padding + marker)

			// Add spaces to align with the token length
			if len(token) > len(marker) {
				extraPadding := len(token) - len(marker) - len(padding)
				annotationLine.WriteString(strings.Repeat(" ", extraPadding))
			}
			annotationLine.WriteString(colorReset)

			// Add to the legend
			if i%3 == 0 && i > 0 {
				legendLine.WriteString("\n")
			} else if i > 0 {
				legendLine.WriteString("  ")
			}
			legendLine.WriteString(fmt.Sprintf("%s%s%d%s: %s", color, colorBold, i+1, colorReset, token))

			// Update position for next token
			pos = tokenPos + len(token)
		}
	}

	// Add any remaining part of the pattern
	if pos < len(pattern) {
		coloredPattern.WriteString(pattern[pos:])
	}

	// Build the final result
	var result strings.Builder
	result.WriteString("Colored pattern:\n")
	result.WriteString(coloredPattern.String() + "\n")
	result.WriteString(annotationLine.String() + "\n\n")
	result.WriteString("Legend:\n")
	result.WriteString(legendLine.String() + "\n")

	return result.String()
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
		{name: "Named Groups", code: format.FeatureNamedGroup, description: "(?P<n>pattern)"},
		{name: "Atomic Groups", code: format.FeatureAtomicGroup, description: "(?>pattern)"},
		{name: "Conditionals", code: format.FeatureConditional, description: "(?(cond)then|else)"},
		{name: "Possessive Quantifiers", code: format.FeaturePossessive, description: "a++, a*+, a?+"},
		{name: "Unicode Properties", code: format.FeatureUnicodeClass, description: "\\p{Property}"},
		{name: "Recursion", code: format.FeatureRecursion, description: "(?R) or (?0)"},
		{name: "Backreferences", code: format.FeatureBackreference, description: "\\1, \\2, etc."},
		{name: "Named Backreferences", code: format.FeatureNamedBackref, description: "\\k<n>"},
	}

	fmt.Printf("%sSupported Features:%s\n", colorBold, colorReset)

	for _, feature := range features {
		supported := colorRed + "✗" + colorReset
		if regexFormat.HasFeature(feature.code) {
			supported = colorGreen + "✓" + colorReset
		}
		fmt.Printf("  %s %s (%s)\n", supported, feature.name, feature.description)
	}

	fmt.Println()
}
