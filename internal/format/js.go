package format

import (
	"fmt"
	"strings"
)

// JsFormat implements the RegexFormat interface for JavaScript RegExp
type JsFormat struct{}

// NewJsFormat creates a new JavaScript format implementation
func NewJsFormat() RegexFormat {
	return &JsFormat{}
}

// Name returns the descriptive name of the format
func (j *JsFormat) Name() string {
	return "JavaScript RegExp"
}

// HasFeature checks if this format supports a specific regex feature
func (j *JsFormat) HasFeature(feature string) bool {
	supportedFeatures := map[string]bool{
		FeatureLookahead:     true,
		FeatureLookbehind:    true,  // Only in newer JS engines
		FeatureNamedGroup:    true,  // Only in newer JS engines
		FeatureAtomicGroup:   false,
		FeatureConditional:   false,
		FeaturePossessive:    false,
		FeatureUnicodeClass:  true,  // With /u flag
		FeatureRecursion:     false,
		FeatureBackreference: true,
		FeatureNamedBackref:  true,  // Only in newer JS engines
	}
	
	return supportedFeatures[feature]
}

// TokenizeRegex breaks a regex pattern into meaningful tokens
func (j *JsFormat) TokenizeRegex(pattern string) []string {
	var tokens []string
	var currentToken strings.Builder
	
	// Check for regex flags at the end
	flags := ""
	if len(pattern) > 2 && pattern[0] == '/' {
		lastSlashPos := strings.LastIndex(pattern, "/")
		if lastSlashPos > 0 && lastSlashPos < len(pattern)-1 {
			flags = pattern[lastSlashPos+1:]
			pattern = pattern[1:lastSlashPos]
			
			// Add flags explanation as first token
			if len(flags) > 0 {
				tokens = append(tokens, "/"+flags)
			}
		} else if pattern[0] == '/' && pattern[len(pattern)-1] == '/' {
			// No flags, but has delimiters
			pattern = pattern[1 : len(pattern)-1]
		}
	}
	
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
			
			// Check for non-greedy quantifier
			if i+1 < len(pattern) && pattern[i+1] == '?' {
				tokens = append(tokens, string(char)+"?")
				i++
			} else {
				tokens = append(tokens, string(char))
			}
			continue
		}
		
		// Handle groups
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
				case '!': // (?!pattern) - negative lookahead
					tokens = append(tokens, "(?!")
					i += 2
				case '<': // Could be lookbehind or named capture
					if i+3 < len(pattern) {
						if pattern[i+3] == '=' { // (?<=pattern) - positive lookbehind
							tokens = append(tokens, "(?<=")
							i += 3
						} else if pattern[i+3] == '!' { // (?<!pattern) - negative lookbehind
							tokens = append(tokens, "(?<!")
							i += 3
						} else { // (?<name>pattern) - named capturing group
							endName := strings.IndexByte(pattern[i+3:], '>')
							if endName >= 0 {
								endName += i + 3
								tokens = append(tokens, pattern[i:endName+1])
								i = endName
							} else {
								tokens = append(tokens, string(char))
							}
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
func (j *JsFormat) ExplainToken(token string) string {
	switch {
	case strings.HasPrefix(token, "/"):
		return explainJsFlags(token[1:])
	case token == "^":
		return "Matches the start of a line"
	case token == "$":
		return "Matches the end of a line"
	case token == ".":
		return "Matches any single character except newline"
	case token == "*":
		return "Matches 0 or more of the preceding element (greedy)"
	case token == "+":
		return "Matches 1 or more of the preceding element (greedy)"
	case token == "?":
		return "Matches 0 or 1 of the preceding element (greedy)"
	case token == "*?":
		return "Matches 0 or more of the preceding element (non-greedy)"
	case token == "+?":
		return "Matches 1 or more of the preceding element (non-greedy)"
	case token == "??":
		return "Matches 0 or 1 of the preceding element (non-greedy)"
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
	case token == "(?!":
		return "Start of a negative lookahead - matches if the pattern inside doesn't match, but doesn't consume characters"
	case token == "(?<=":
		return "Start of a positive lookbehind - matches if the pattern inside matches immediately before current position"
	case token == "(?<!":
		return "Start of a negative lookbehind - matches if the pattern inside doesn't match immediately before current position"
	case strings.HasPrefix(token, "(?<") && strings.HasSuffix(token, ">") && !strings.Contains(token, "<?") && !strings.Contains(token, "<!"):
		name := token[3 : len(token)-1]
		return fmt.Sprintf("Start of a named capturing group called '%s'", name)
	case strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]"):
		if len(token) > 2 && token[1] == '^' {
			return fmt.Sprintf("Matches any character NOT in the set: %s", token[2:len(token)-1])
		}
		return fmt.Sprintf("Matches any character in the set: %s", token[1:len(token)-1])
	case strings.HasPrefix(token, "\\"):
		return explainJsEscapeSequence(token)
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

// explainJsFlags explains JavaScript RegExp flags
func explainJsFlags(flags string) string {
	if flags == "" {
		return "No flags specified"
	}
	
	var explanations []string
	for _, flag := range flags {
		switch flag {
		case 'g':
			explanations = append(explanations, "g: Global search - find all matches rather than stopping after the first match")
		case 'i':
			explanations = append(explanations, "i: Case-insensitive search")
		case 'm':
			explanations = append(explanations, "m: Multi-line search - ^ and $ match start/end of each line")
		case 's':
			explanations = append(explanations, "s: Dot-all mode - the dot (.) matches newlines")
		case 'u':
			explanations = append(explanations, "u: Unicode mode - treat pattern as a sequence of Unicode code points")
		case 'y':
			explanations = append(explanations, "y: Sticky mode - matches only from the index indicated by the lastIndex property")
		case 'd':
			explanations = append(explanations, "d: Generate indices for substring matches")
		default:
			explanations = append(explanations, fmt.Sprintf("%c: Unknown flag", flag))
		}
	}
	
	return "Flags: " + strings.Join(explanations, ", ")
}

// explainJsEscapeSequence explains JavaScript-specific escape sequences
func explainJsEscapeSequence(sequence string) string {
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
	case 'k':
		if len(sequence) > 2 && sequence[2] == '<' {
			end := strings.IndexByte(sequence[3:], '>')
			if end >= 0 {
				name := sequence[3 : 3+end]
				return fmt.Sprintf("Backreference to the named group '%s'", name)
			}
		}
		return "Invalid named backreference"
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return fmt.Sprintf("Backreference to capturing group %c", sequence[1])
	case 'p', 'P':
		if len(sequence) > 2 && sequence[2] == '{' {
			end := strings.IndexByte(sequence[3:], '}')
			if end >= 0 {
				name := sequence[3 : 3+end]
				if sequence[1] == 'p' {
					return fmt.Sprintf("Matches a character with the unicode property '%s' (requires u flag)", name)
				} else {
					return fmt.Sprintf("Matches a character without the unicode property '%s' (requires u flag)", name)
				}
			}
		}
		return "Invalid unicode property"
	case 'u':
		if len(sequence) >= 6 && isHexDigit(sequence[2]) && isHexDigit(sequence[3]) && isHexDigit(sequence[4]) && isHexDigit(sequence[5]) {
			return fmt.Sprintf("Matches the Unicode character U+%s", sequence[2:6])
		}
		return "Invalid Unicode escape sequence"
	case 'x':
		if len(sequence) >= 4 && isHexDigit(sequence[2]) && isHexDigit(sequence[3]) {
			return fmt.Sprintf("Matches the character with hex code %s", sequence[2:4])
		}
		return "Invalid hexadecimal escape sequence"
	default:
		return fmt.Sprintf("Matches the character '%c' literally", sequence[1])
	}
}

// Helper function to check if a byte is a hex digit
func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
} 