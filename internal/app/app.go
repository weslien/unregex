package app

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// Common character sets for sample generation
var (
	digits       = "0123456789"
	lowerLetters = "abcdefghijklmnopqrstuvwxyz"
	upperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphaNumeric = digits + lowerLetters + upperLetters
	whitespace   = " \t\n\r"
	specialChars = "!@#$%^&*()-_=+[]{}|;:,.<>?/"
)

// Initialize random number generator
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

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
		annotatedPattern := visualizePattern(pattern, tokens, colorMap)
		fmt.Println(annotatedPattern)

		// Generate and display a sample matching string
		fmt.Println(generateSampleMatch(pattern, formatName, tokens, colorMap))
	}

	fmt.Println("\nNOTE: This is a basic regex explainer. Some complex patterns might not be perfectly tokenized.")

	return nil
}

// generateSampleMatch creates an example string that matches the regex pattern
func generateSampleMatch(pattern, formatName string, tokens []string, colorMap []string) string {
	// Try to generate a deterministic sample based on the tokens
	sample, tokenMap := generateDeterministicSample(tokens)

	// Verify if the generated sample matches the pattern
	var r *regexp.Regexp
	var err error

	if formatName == "go" {
		r, err = regexp.Compile(pattern)
	} else {
		// For non-Go formats, just attempt to compile but don't rely on match checking
		r, err = regexp.Compile(pattern)
	}

	// If we couldn't compile the pattern or the sample doesn't match,
	// use a fallback approach with common examples
	matchStatus := "Verified match"
	useAlternate := false

	if err != nil || (r != nil && !r.MatchString(sample)) {
		matchStatus = "Approximate match (pattern contains advanced features)"

		// For patterns with alternation, use a special handler
		if strings.Contains(pattern, "|") {
			sample = generateAlternativeSample(pattern, formatName)
			useAlternate = true
		} else {
			sample = generateFallbackSample(pattern, formatName)
		}

		// Double-check if our alternative sample matches
		if r != nil && r.MatchString(sample) {
			matchStatus = "Verified match (using alternative)"
		}
	}

	// Build the display string with colors
	var result strings.Builder
	result.WriteString(fmt.Sprintf("%sExample matching string:%s\n", colorBold, colorReset))

	if sample == "" {
		result.WriteString("Couldn't generate a sample for this pattern (too complex or contains unsupported features)\n")
		return result.String()
	}

	// For alternation examples, we need a different approach to colorize
	if useAlternate && strings.Contains(pattern, "|") {
		// Create a second colored sample that highlights the alternate choice
		coloredSample := colorizeAlternativeExample(pattern, sample, tokens, colorMap)
		result.WriteString(coloredSample + "\n")
	} else {
		// Colorize the sample string using token positions
		var coloredSample strings.Builder
		for i, c := range sample {
			char := string(c)

			// Find the token index for this character
			tokenIndex := -1
			for idx, pos := range tokenMap {
				if i >= pos.start && i < pos.end {
					tokenIndex = idx
					break
				}
			}

			// Apply color if we found a token
			if tokenIndex >= 0 && tokenIndex < len(colorMap) {
				coloredSample.WriteString(colorMap[tokenIndex%len(colorMap)] + colorBold + char + colorReset)
			} else {
				coloredSample.WriteString(char)
			}
		}

		result.WriteString(coloredSample.String() + "\n")
	}

	result.WriteString(fmt.Sprintf("(%s)\n", matchStatus))

	return result.String()
}

// colorizeAlternativeExample creates a colored version of the sample string
// that properly highlights the alternative choice in the pattern
func colorizeAlternativeExample(pattern, sample string, tokens []string, colorMap []string) string {
	var result strings.Builder

	// For a pattern like ^hello(world|universe)[0-9]+$
	// We want to visually show hellouniverse123 with proper coloring

	// Try to extract key parts - for our example pattern:
	// - prefix: hello
	// - alt1: world
	// - alt2: universe
	// - suffix: followed by digits

	// Identify used alternative from the pattern
	altPattern := regexp.MustCompile(`\(([^|]+)\|([^)]+)\)`)
	matches := altPattern.FindStringSubmatch(pattern)

	if len(matches) >= 3 {
		alt1 := matches[1]
		alt2 := matches[2]

		// For this example, we'll color:
		// "hello" as one color
		// "universe" as another (highlighting this is the alternative)
		// "123" as another

		// Find these portions in the sample
		prefixPattern := regexp.MustCompile(`^[^(]*`)
		prefixMatch := prefixPattern.FindString(pattern)
		prefixMatch = strings.ReplaceAll(prefixMatch, "^", "") // Remove ^ anchor

		suffixPattern := regexp.MustCompile(`\)[^)]*$`)
		suffixMatch := suffixPattern.FindString(pattern)
		suffixMatch = strings.ReplaceAll(suffixMatch, "$", "") // Remove $ anchor
		suffixMatch = strings.TrimPrefix(suffixMatch, ")")     // Remove closing paren

		// Get the actual digits portion of the sample
		digitMatches := regexp.MustCompile(`\d+`).FindAllString(sample, -1)
		digits := ""
		if len(digitMatches) > 0 {
			digits = digitMatches[len(digitMatches)-1]
		}

		// Determine which alternative is in the sample
		usingAlt2 := strings.Contains(sample, alt2)

		// Color the prefix
		if prefixMatch != "" {
			result.WriteString(colorMap[0] + colorBold + prefixMatch + colorReset)
		}

		// Color the chosen alternative
		if usingAlt2 {
			// Use a special color for the alternate choice
			result.WriteString(colorMap[2] + colorBold + alt2 + colorReset)
		} else {
			result.WriteString(colorMap[1] + colorBold + alt1 + colorReset)
		}

		// Color the suffix (digits)
		if digits != "" {
			result.WriteString(colorMap[3] + colorBold + digits + colorReset)
		}
	} else {
		// Fallback if we can't parse the alternation properly
		result.WriteString(sample)
	}

	return result.String()
}

// Position represents a start and end position
type Position struct {
	start, end int
}

// generateDeterministicSample tries to create a sample string based on the tokens
func generateDeterministicSample(tokens []string) (string, []Position) {
	var sample strings.Builder
	tokenMap := make([]Position, len(tokens))

	// Stack to track active groups - for handling alternations properly
	type Group struct {
		openIndex  int    // Index of the opening parenthesis
		content    string // Content built so far
		altIndices []int  // Indices of alternation operators
	}
	var groups []Group

	// Pass 1: Process special structures like alternation
	// First identify groups and their alternations
	groupMap := make(map[int]int)    // Maps opening to closing parenthesis indices
	altGroupMap := make(map[int]int) // Maps alternation operators to their group

	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "(" {
			groups = append(groups, Group{openIndex: i, altIndices: []int{}})
		} else if tokens[i] == "|" && len(groups) > 0 {
			// Add this alternation to the current group
			currentGroup := &groups[len(groups)-1]
			currentGroup.altIndices = append(currentGroup.altIndices, i)
			altGroupMap[i] = len(groups) - 1
		} else if tokens[i] == ")" && len(groups) > 0 {
			// Map this closing parenthesis to its opening one
			openIndex := groups[len(groups)-1].openIndex
			groupMap[openIndex] = i
			groups = groups[:len(groups)-1] // Pop the group
		}
	}

	// Pass a flag to determine if we've used an alternation's right side
	usedAltRight := make(map[int]bool)

	// Go through tokens and build the sample
	for i, token := range tokens {
		startPos := sample.Len()

		// Handle different token types
		switch token {
		case "^", "$", "\\b", "\\B":
			// Zero-width assertions don't contribute to the sample
		case ".":
			sample.WriteString("x")
		case "\\d":
			sample.WriteString("5")
		case "\\w":
			sample.WriteString("a")
		case "\\s":
			sample.WriteString(" ")
		case "+":
			// Repeat the preceding character once more (for +)
			if sample.Len() > 0 {
				lastChar := sample.String()[sample.Len()-1:]
				sample.WriteString(lastChar)
			}
		case "*", "?", "{", "}":
			// Other quantifiers don't contribute directly
		case "(":
			// Opening of a group - no contribution
		case ")":
			// Closing of a group - no contribution
		case "|":
			// Handle alternation
			if groupIdx, exists := altGroupMap[i]; exists {
				// This is a tracked alternation within a group
				// We'll randomly pick one side of the alternation
				// For predictability in examples, we'll favor the right side
				if !usedAltRight[groupIdx] {
					// Use the right side of the alternation (clear what we've built for the left side)
					// Find the right expression in the next tokens
					rightStart := i + 1
					rightEnd := -1

					// Find the end of the alternation (next | or ) at this level)
					depth := 0
					for j := rightStart; j < len(tokens); j++ {
						if tokens[j] == "(" {
							depth++
						} else if tokens[j] == ")" {
							if depth == 0 {
								rightEnd = j
								break
							}
							depth--
						} else if tokens[j] == "|" && depth == 0 {
							// Another alternation at this level
							rightEnd = j
							break
						}
					}

					if rightEnd > rightStart {
						// Skip to after the right expression
						// We'll handle the right side when we naturally get to those tokens
						usedAltRight[groupIdx] = true
					}
				}
			}
		case "[0-9]":
			sample.WriteString("7")
		case "[a-z]":
			sample.WriteString("m")
		case "[A-Z]":
			sample.WriteString("M")
		case "[a-zA-Z]":
			sample.WriteString("k")
		case "[a-zA-Z0-9]":
			sample.WriteString("k")
		default:
			// If token contains character ranges or special sequences
			if strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]") {
				// For character classes, pick something in the range
				sample.WriteString("x")
			} else if strings.HasPrefix(token, "\\") {
				// Handle escape sequences
				if len(token) > 1 {
					switch token[1] {
					case 'd':
						sample.WriteString("5")
					case 'w':
						sample.WriteString("a")
					case 's':
						sample.WriteString(" ")
					default:
						// For other escape sequences, just add a placeholder
						sample.WriteString("x")
					}
				}
			} else {
				// For literal text, include it directly
				sample.WriteString(token)
			}
		}

		// Record the position of this token in the sample
		tokenMap[i] = Position{startPos, sample.Len()}
	}

	return sample.String(), tokenMap
}

// Simplified version to handle alternation patterns better
func generateAlternativeSample(pattern, formatName string) string {
	// Try a simpler approach for pattern with alternation
	if strings.Contains(pattern, "|") {
		// For a pattern like ^hello(world|universe)[0-9]+$
		// we'll generate hellouniverse123 to show the alternative

		// Identify common alternation patterns and replace with the right side
		pattern = regexp.MustCompile(`\([^|]+\|([^)]+)\)`).ReplaceAllString(pattern, "$1")

		// Then use the fallback to clean up other syntax
		return generateFallbackSample(pattern, formatName)
	}

	return generateFallbackSample(pattern, formatName)
}

// generateFallbackSample creates a sample string using common patterns
func generateFallbackSample(pattern, formatName string) string {
	// Some common replacements for regex patterns
	replacements := map[string]string{
		"^":        "",
		"$":        "",
		"\\d":      "5",
		"\\d+":     "123",
		"\\d*":     "456",
		"\\d{2}":   "42",
		"\\d{1,3}": "789",
		"\\w":      "a",
		"\\w+":     "word",
		"\\w*":     "text",
		"\\s":      " ",
		"\\s+":     "   ",
		".":        "x",
		".+":       "some text",
		".*":       "any text",
		"[0-9]":    "5",
		"[a-z]":    "k",
		"[A-Z]":    "K",
	}

	// Start with the pattern and make replacements
	sample := pattern

	// Replace all occurrences of each regex element with its sample value
	for regex, replacement := range replacements {
		sample = strings.ReplaceAll(sample, regex, replacement)
	}

	// Remove common regex syntax elements
	syntaxToRemove := []string{"(", ")", "[", "]", "{", "}", "?", "+", "*", "|", "\\"}
	for _, s := range syntaxToRemove {
		sample = strings.ReplaceAll(sample, s, "")
	}

	// If the sample is empty, provide a generic example
	if len(strings.TrimSpace(sample)) == 0 {
		return "example123"
	}

	return sample
}

// visualizePattern creates an annotated representation of the regex with numbers
func visualizePattern(pattern string, tokens []string, colorMap []string) string {
	// First, generate a colored version of the pattern with token boundaries
	var coloredPattern strings.Builder
	var annotationLine strings.Builder
	var legendLine strings.Builder

	// Keep track of position in the pattern
	pos := 0

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
