package format

import (
	"fmt"
	"strings"
)

// PythonFormat implements the RegexFormat interface for Python regular expressions
type PythonFormat struct{}

// NewPythonFormat creates a new Python format implementation
func NewPythonFormat() RegexFormat {
	return &PythonFormat{}
}

// Name returns the descriptive name of the format
func (p *PythonFormat) Name() string {
	return "Python re"
}

// HasFeature checks if this format supports a specific regex feature
func (p *PythonFormat) HasFeature(feature string) bool {
	supportedFeatures := map[string]bool{
		FeatureLookahead:     true,
		FeatureLookbehind:    true,
		FeatureNamedGroup:    true,
		FeatureAtomicGroup:   false,
		FeatureConditional:   false,
		FeaturePossessive:    false,
		FeatureUnicodeClass:  true,
		FeatureRecursion:     false,
		FeatureBackreference: true,
		FeatureNamedBackref:  true,
	}
	
	return supportedFeatures[feature]
}

// TokenizeRegex breaks a regex pattern into meaningful tokens
func (p *PythonFormat) TokenizeRegex(pattern string) []string {
	var tokens []string
	var currentToken strings.Builder
	
	// Check for raw string marker and flags
	if len(pattern) > 0 && (pattern[0] == 'r' || pattern[0] == 'R') {
		if len(pattern) > 1 && (pattern[1] == '"' || pattern[1] == '\'') {
			tokens = append(tokens, pattern[0:2])
			pattern = pattern[2:]
		}
	}
	
	// Handle inline flags at the beginning
	if len(pattern) > 2 && pattern[0] == '(' && pattern[1] == '?' {
		flagEnd := strings.IndexByte(pattern, ')')
		if flagEnd > 2 {
			isFlag := true
			for i := 2; i < flagEnd; i++ {
				if pattern[i] != 'a' && pattern[i] != 'i' && pattern[i] != 'L' && 
				   pattern[i] != 'm' && pattern[i] != 's' && pattern[i] != 'u' && 
				   pattern[i] != 'x' {
					isFlag = false
					break
				}
			}
			if isFlag {
				tokens = append(tokens, pattern[0:flagEnd+1])
				pattern = pattern[flagEnd+1:]
			}
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
			
			// Python has some multi-character escape sequences
			if i+2 < len(pattern) && pattern[i+1] == 'x' {
				// \xhh - up to 2 hex digits
				hexEnd := i + 4
				if hexEnd > len(pattern) {
					hexEnd = len(pattern)
				}
				for j := i + 2; j < hexEnd; j++ {
					if !isHexDigit(pattern[j]) {
						hexEnd = j
						break
					}
				}
				tokens = append(tokens, pattern[i:hexEnd])
				i = hexEnd - 1
				continue
			} else if i+2 < len(pattern) && pattern[i+1] == 'u' {
				// \uxxxx - exactly 4 hex digits
				if i+6 <= len(pattern) && isHexDigit(pattern[i+2]) && isHexDigit(pattern[i+3]) && 
				   isHexDigit(pattern[i+4]) && isHexDigit(pattern[i+5]) {
					tokens = append(tokens, pattern[i:i+6])
					i += 5
					continue
				}
			} else if i+2 < len(pattern) && pattern[i+1] == 'U' {
				// \Uxxxxxxxx - exactly 8 hex digits
				if i+10 <= len(pattern) && isHexDigit(pattern[i+2]) && isHexDigit(pattern[i+3]) && 
				   isHexDigit(pattern[i+4]) && isHexDigit(pattern[i+5]) && isHexDigit(pattern[i+6]) && 
				   isHexDigit(pattern[i+7]) && isHexDigit(pattern[i+8]) && isHexDigit(pattern[i+9]) {
					tokens = append(tokens, pattern[i:i+10])
					i += 9
					continue
				}
			} else if i+2 < len(pattern) && pattern[i+1] == 'N' && pattern[i+2] == '{' {
				// \N{name} - Unicode character by name
				end := strings.IndexByte(pattern[i+3:], '}')
				if end >= 0 {
					tokens = append(tokens, pattern[i:i+end+4])
					i += end + 3
					continue
				}
			} else {
				tokens = append(tokens, pattern[i:i+2])
				i++
				continue
			}
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
				case 'P': // Python specific named group syntaxes
					if i+3 < len(pattern) {
						if pattern[i+3] == '<' { // (?P<name>pattern) - Named group
							endName := strings.IndexByte(pattern[i+4:], '>')
							if endName >= 0 {
								endName += i + 4
								tokens = append(tokens, pattern[i:endName+1])
								i = endName
								continue
							}
						} else if pattern[i+3] == '=' { // (?P=name) - Named backreference
							// Find the end of the name
							j := i + 4
							for j < len(pattern) && pattern[j] != ')' {
								j++
							}
							if j < len(pattern) {
								tokens = append(tokens, pattern[i:j+1])
								i = j
								continue
							}
						}
					}
					tokens = append(tokens, string(char))
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
func (p *PythonFormat) ExplainToken(token string) string {
	switch {
	case strings.HasPrefix(token, "r'") || strings.HasPrefix(token, "r\"") || 
	     strings.HasPrefix(token, "R'") || strings.HasPrefix(token, "R\""):
		return "Raw string marker - backslashes are treated literally"
	case strings.HasPrefix(token, "(?") && strings.HasSuffix(token, ")") && len(token) > 3:
		// Check for inline flags
		isFlag := true
		for i := 2; i < len(token)-1; i++ {
			if token[i] != 'a' && token[i] != 'i' && token[i] != 'L' && 
			   token[i] != 'm' && token[i] != 's' && token[i] != 'u' && 
			   token[i] != 'x' {
				isFlag = false
				break
			}
		}
		if isFlag {
			return explainPythonFlags(token[2 : len(token)-1])
		}
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
	case strings.HasPrefix(token, "(?P<") && strings.HasSuffix(token, ">"):
		name := token[4 : len(token)-1]
		return fmt.Sprintf("Start of a named capturing group called '%s'", name)
	case strings.HasPrefix(token, "(?P=") && strings.HasSuffix(token, ")"):
		name := token[4 : len(token)-1]
		return fmt.Sprintf("Backreference to the named group '%s'", name)
	case strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]"):
		if len(token) > 2 && token[1] == '^' {
			return fmt.Sprintf("Matches any character NOT in the set: %s", token[2:len(token)-1])
		}
		return fmt.Sprintf("Matches any character in the set: %s", token[1:len(token)-1])
	case strings.HasPrefix(token, "\\"):
		return explainPythonEscapeSequence(token)
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
	
	return fmt.Sprintf("Unknown token: %s", token)
}

// explainPythonFlags explains Python regex flags
func explainPythonFlags(flags string) string {
	if flags == "" {
		return "No flags specified"
	}
	
	var explanations []string
	for _, flag := range flags {
		switch flag {
		case 'a':
			explanations = append(explanations, "a: ASCII-only matching")
		case 'i':
			explanations = append(explanations, "i: Case-insensitive matching")
		case 'L':
			explanations = append(explanations, "L: Locale-dependent matching")
		case 'm':
			explanations = append(explanations, "m: Multi-line matching - ^ and $ match at line breaks")
		case 's':
			explanations = append(explanations, "s: Dot matches all - the dot (.) matches any character including newline")
		case 'u':
			explanations = append(explanations, "u: Unicode matching")
		case 'x':
			explanations = append(explanations, "x: Verbose - whitespace and comments in pattern are ignored")
		default:
			explanations = append(explanations, fmt.Sprintf("%c: Unknown flag", flag))
		}
	}
	
	return "Flags: " + strings.Join(explanations, ", ")
}

// explainPythonEscapeSequence explains Python-specific escape sequences
func explainPythonEscapeSequence(sequence string) string {
	if len(sequence) < 2 {
		return "Invalid escape sequence"
	}
	
	switch sequence[1] {
	case 'A':
		return "Matches only at the start of the string"
	case 'Z':
		return "Matches only at the end of the string"
	case 'd':
		return "Matches any decimal digit (0-9)"
	case 'D':
		return "Matches any non-digit character"
	case 's':
		return "Matches any whitespace character (space, tab, newline, etc.)"
	case 'S':
		return "Matches any non-whitespace character"
	case 'w':
		return "Matches any alphanumeric character (including underscore)"
	case 'W':
		return "Matches any non-alphanumeric character"
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
	case 'a':
		return "Matches a bell (BEL) character"
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return fmt.Sprintf("Backreference to capturing group %c", sequence[1])
	case 'g':
		if len(sequence) > 3 && sequence[2] == '<' {
			end := strings.IndexByte(sequence[3:], '>')
			if end >= 0 {
				name := sequence[3 : 3+end]
				return fmt.Sprintf("Backreference to the named group '%s'", name)
			}
		}
		return "Invalid named backreference"
	case 'x':
		if len(sequence) >= 4 && isHexDigit(sequence[2]) && isHexDigit(sequence[3]) {
			return fmt.Sprintf("Matches the character with hex code %s", sequence[2:4])
		}
		return "Invalid hexadecimal escape sequence"
	case 'u':
		if len(sequence) >= 6 && isHexDigit(sequence[2]) && isHexDigit(sequence[3]) && 
		   isHexDigit(sequence[4]) && isHexDigit(sequence[5]) {
			return fmt.Sprintf("Matches the Unicode character U+%s", sequence[2:6])
		}
		return "Invalid Unicode escape sequence"
	case 'U':
		if len(sequence) >= 10 && isHexDigit(sequence[2]) && isHexDigit(sequence[3]) && 
		   isHexDigit(sequence[4]) && isHexDigit(sequence[5]) && isHexDigit(sequence[6]) && 
		   isHexDigit(sequence[7]) && isHexDigit(sequence[8]) && isHexDigit(sequence[9]) {
			return fmt.Sprintf("Matches the Unicode character U+%s", sequence[2:10])
		}
		return "Invalid extended Unicode escape sequence"
	case 'N':
		if len(sequence) > 3 && sequence[2] == '{' {
			end := strings.IndexByte(sequence[2:], '}')
			if end > 0 {
				name := sequence[3 : 2+end]
				return fmt.Sprintf("Matches the Unicode character named '%s'", name)
			}
		}
		return "Invalid Unicode name escape sequence"
	default:
		return fmt.Sprintf("Matches the character '%c' literally", sequence[1])
	}
} 