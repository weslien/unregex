package format

import (
	"fmt"
	"strings"
)

// GoFormat implements the RegexFormat interface for Go regular expressions
type GoFormat struct{}

// NewGoFormat creates a new Go format implementation
func NewGoFormat() RegexFormat {
	return &GoFormat{}
}

// Name returns the descriptive name of the format
func (g *GoFormat) Name() string {
	return "Go Regexp"
}

// HasFeature checks if this format supports a specific regex feature
func (g *GoFormat) HasFeature(feature string) bool {
	// Go regex has limited features compared to PCRE
	supportedFeatures := map[string]bool{
		FeatureLookahead:     true,  // Supports only positive lookahead
		FeatureLookbehind:    false, // No lookbehind
		FeatureNamedGroup:    true,  // Go has named groups
		FeatureAtomicGroup:   false, // No atomic groups
		FeatureConditional:   false, // No conditionals
		FeaturePossessive:    false, // No possessive quantifiers
		FeatureUnicodeClass:  true,  // Has unicode classes
		FeatureRecursion:     false, // No recursion
		FeatureBackreference: true,  // Supports backreferences
		FeatureNamedBackref:  true,  // Supports named backreferences
	}
	
	return supportedFeatures[feature]
}

// TokenizeRegex breaks a regex pattern into meaningful tokens
func (g *GoFormat) TokenizeRegex(pattern string) []string {
	var tokens []string
	var currentToken strings.Builder
	
	for i := 0; i < len(pattern); i++ {
		char := pattern[i]
		
		// Handle character classes
		if char == '[' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			
			end := FindClosingBracket(pattern, i)
			if end > i {
				tokens = append(tokens, pattern[i:end+1])
				i = end
				continue
			}
		}
		
		// Handle special escape sequences
		if char == '\\' && i+1 < len(pattern) {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, pattern[i:i+2])
			i++
			continue
		}
		
		// Handle curly brace quantifiers
		if char == '{' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			
			end := FindClosingCurlyBrace(pattern, i)
			if end > i {
				tokens = append(tokens, pattern[i:end+1])
				i = end
				continue
			}
		}
		
		// Handle simple quantifiers
		if char == '*' || char == '+' || char == '?' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
			continue
		}
		
		// Handle groups and named groups
		if char == '(' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			
			// Check for non-capturing and other special groups
			if i+2 < len(pattern) && pattern[i+1] == '?' {
				switch pattern[i+2] {
				case ':': // (?:pattern) - non-capturing group
					tokens = append(tokens, "(?:")
					i += 2
				case '=': // (?=pattern) - positive lookahead
					tokens = append(tokens, "(?=")
					i += 2
				case 'P': // (?P<name>pattern) - named capturing group
					if i+3 < len(pattern) && pattern[i+3] == '<' {
						endName := strings.IndexByte(pattern[i+4:], '>')
						if endName >= 0 {
							endName += i + 4
							tokens = append(tokens, pattern[i:endName+1])
							i = endName
						} else {
							tokens = append(tokens, string(char))
						}
					} else {
						tokens = append(tokens, string(char))
					}
				default:
					tokens = append(tokens, string(char))
				}
				continue
			} else {
				tokens = append(tokens, string(char))
				continue
			}
		}
		
		if char == ')' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
			continue
		}
		
		// Handle alternation
		if char == '|' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
			continue
		}
		
		// Handle anchors
		if char == '^' || char == '$' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
			continue
		}
		
		// Handle dot
		if char == '.' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
			continue
		}
		
		// Default case: add to current token
		currentToken.WriteByte(char)
	}
	
	// Add the last token if any
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}
	
	return tokens
}

// ExplainToken provides a human-readable explanation for a regex token
func (g *GoFormat) ExplainToken(token string) string {
	switch {
	case token == "^":
		return "Matches the start of a line"
	case token == "$":
		return "Matches the end of a line"
	case token == ".":
		return "Matches any single character except newline"
	case token == "*":
		return "Matches 0 or more of the preceding element"
	case token == "+":
		return "Matches 1 or more of the preceding element"
	case token == "?":
		return "Matches 0 or 1 of the preceding element"
	case token == "|":
		return "Acts as an OR operator - matches the expression before or after the |"
	case token == "(":
		return "Start of a capturing group"
	case token == ")":
		return "End of a capturing group"
	case token == "(?:":
		return "Start of a non-capturing group - groups the expression but doesn't create a capture group"
	case token == "(?=":
		return "Start of a positive lookahead - matches if the pattern inside matches, but doesn't consume characters"
	case strings.HasPrefix(token, "(?P<") && strings.HasSuffix(token, ">"):
		name := token[4 : len(token)-1]
		return fmt.Sprintf("Start of a named capturing group called '%s'", name)
	case strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]"):
		if len(token) > 2 && token[1] == '^' {
			return fmt.Sprintf("Matches any character NOT in the set: %s", token[2:len(token)-1])
		}
		return fmt.Sprintf("Matches any character in the set: %s", token[1:len(token)-1])
	case strings.HasPrefix(token, "\\"):
		return explainEscapeSequence(token)
	case strings.HasPrefix(token, "{") && strings.HasSuffix(token, "}"):
		content := token[1 : len(token)-1]
		if strings.Contains(content, ",") {
			parts := strings.Split(content, ",")
			if len(parts) == 2 {
				if parts[1] == "" {
					return fmt.Sprintf("Matches at least %s occurrences of the preceding element", parts[0])
				}
				return fmt.Sprintf("Matches between %s and %s occurrences of the preceding element", parts[0], parts[1])
			}
		}
		return fmt.Sprintf("Matches exactly %s occurrences of the preceding element", content)
	default:
		if len(token) == 1 {
			return fmt.Sprintf("Matches the character '%s' literally", token)
		}
		return fmt.Sprintf("Matches the string '%s' literally", token)
	}
}

// explainEscapeSequence explains common regex escape sequences
func explainEscapeSequence(sequence string) string {
	if len(sequence) < 2 {
		return "Invalid escape sequence"
	}
	
	switch sequence[1] {
	case 'd':
		return "Matches any digit (0-9)"
	case 'D':
		return "Matches any non-digit character"
	case 'w':
		return "Matches any word character (alphanumeric plus underscore)"
	case 'W':
		return "Matches any non-word character"
	case 's':
		return "Matches any whitespace character (space, tab, newline, etc.)"
	case 'S':
		return "Matches any non-whitespace character"
	case 'b':
		return "Matches a word boundary"
	case 'B':
		return "Matches a non-word boundary"
	case 'A':
		return "Matches the start of the string"
	case 'z':
		return "Matches the end of the string"
	case 'n':
		return "Matches a newline character"
	case 't':
		return "Matches a tab character"
	case 'r':
		return "Matches a carriage return character"
	case 'f':
		return "Matches a form feed character"
	case 'v':
		return "Matches a vertical tab character"
	case '0':
		return "Matches a null character"
	default:
		return fmt.Sprintf("Matches the character '%c' literally", sequence[1])
	}
} 