package format

import (
	"fmt"
	"strings"
)

// PosixFormat implements the RegexFormat interface for POSIX Extended Regular Expressions
type PosixFormat struct{}

// NewPosixFormat creates a new POSIX format implementation
func NewPosixFormat() RegexFormat {
	return &PosixFormat{}
}

// Name returns the descriptive name of the format
func (p *PosixFormat) Name() string {
	return "POSIX Extended Regular Expressions"
}

// HasFeature checks if this format supports a specific regex feature
func (p *PosixFormat) HasFeature(feature string) bool {
	// POSIX ERE has limited features
	supportedFeatures := map[string]bool{
		FeatureLookahead:     false,
		FeatureLookbehind:    false,
		FeatureNamedGroup:    false,
		FeatureAtomicGroup:   false,
		FeatureConditional:   false,
		FeaturePossessive:    false,
		FeatureUnicodeClass:  false,
		FeatureRecursion:     false,
		FeatureBackreference: true,
		FeatureNamedBackref:  false,
	}
	
	return supportedFeatures[feature]
}

// TokenizeRegex breaks a regex pattern into meaningful tokens
func (p *PosixFormat) TokenizeRegex(pattern string) []string {
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
			
			// Check for POSIX character classes
			if i+2 < len(pattern) && pattern[i+1] == '[' && pattern[i+2] == ':' {
				end := strings.Index(pattern[i:], ":]")
				if end > 3 { // [[:class:]]
					endBracket := FindClosingBracket(pattern, i)
					if endBracket > i+end+2 { // Make sure the bracket closes after the POSIX class
						tokens = append(tokens, pattern[i:endBracket+1])
						i = endBracket
						continue
					}
				}
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
		
		// Handle groups
		if char == '(' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
			continue
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
func (p *PosixFormat) ExplainToken(token string) string {
	switch {
	case token == "^":
		return "Matches the start of a line"
	case token == "$":
		return "Matches the end of a line"
	case token == ".":
		return "Matches any single character"
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
	case strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]"):
		if strings.Contains(token, "[[:") && strings.Contains(token, ":]]") {
			// Extract POSIX character class name
			start := strings.Index(token, "[[:")
			end := strings.Index(token, ":]]")
			if start >= 0 && end > start+3 {
				className := token[start+3 : end]
				return explainPosixCharClass(className)
			}
		}
		
		if len(token) > 2 && token[1] == '^' {
			return fmt.Sprintf("Matches any character NOT in the set: %s", token[2:len(token)-1])
		}
		return fmt.Sprintf("Matches any character in the set: %s", token[1:len(token)-1])
	case strings.HasPrefix(token, "\\"):
		return explainPosixEscapeSequence(token)
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

// explainPosixCharClass explains POSIX character classes
func explainPosixCharClass(className string) string {
	switch className {
	case "alnum":
		return "Matches any alphanumeric character (a-z, A-Z, 0-9)"
	case "alpha":
		return "Matches any alphabetic character (a-z, A-Z)"
	case "ascii":
		return "Matches any ASCII character (0-127)"
	case "blank":
		return "Matches space and tab characters"
	case "cntrl":
		return "Matches control characters"
	case "digit":
		return "Matches decimal digits (0-9)"
	case "graph":
		return "Matches visible characters (not including space)"
	case "lower":
		return "Matches lowercase letters (a-z)"
	case "print":
		return "Matches visible characters (including space)"
	case "punct":
		return "Matches punctuation characters"
	case "space":
		return "Matches whitespace characters (space, tab, newline, etc.)"
	case "upper":
		return "Matches uppercase letters (A-Z)"
	case "word":
		return "Matches word characters (alphanumeric plus underscore)"
	case "xdigit":
		return "Matches hexadecimal digits (0-9, a-f, A-F)"
	default:
		return fmt.Sprintf("Unknown POSIX character class '[:%s:]'", className)
	}
}

// explainPosixEscapeSequence explains POSIX-specific escape sequences
func explainPosixEscapeSequence(sequence string) string {
	if len(sequence) < 2 {
		return "Invalid escape sequence"
	}
	
	// Most POSIX regex implementations support these common escape sequences
	switch sequence[1] {
	case 'n':
		return "Matches a newline character"
	case 't':
		return "Matches a tab character"
	case 'r':
		return "Matches a carriage return character"
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return fmt.Sprintf("Backreference to capturing group %c", sequence[1])
	default:
		return fmt.Sprintf("Matches the character '%c' literally", sequence[1])
	}
} 