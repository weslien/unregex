package format

// RegexFormat defines the interface for different regex format implementations
type RegexFormat interface {
	// Name returns the descriptive name of the format
	Name() string
	
	// TokenizeRegex breaks a regex pattern into meaningful tokens
	TokenizeRegex(pattern string) []string
	
	// ExplainToken provides a human-readable explanation for a regex token
	ExplainToken(token string) string
	
	// HasFeature checks if this format supports a specific regex feature
	HasFeature(feature string) bool
}

// Feature constants for different regex capabilities
const (
	FeatureLookahead      = "lookahead"
	FeatureLookbehind     = "lookbehind"
	FeatureNamedGroup     = "named_group"
	FeatureAtomicGroup    = "atomic_group"
	FeatureConditional    = "conditional"
	FeaturePossessive     = "possessive"
	FeatureUnicodeClass   = "unicode_class"
	FeatureRecursion      = "recursion"
	FeatureBackreference  = "backreference"
	FeatureNamedBackref   = "named_backref"
)

// GetFormat returns the appropriate RegexFormat implementation for the specified format
func GetFormat(formatName string) RegexFormat {
	switch formatName {
	case "go":
		return NewGoFormat()
	case "pcre":
		return NewPcreFormat()
	case "posix":
		return NewPosixFormat()
	case "js":
		return NewJsFormat()
	case "python":
		return NewPythonFormat()
	default:
		// Default to Go format
		return NewGoFormat()
	}
}

// findClosingBracket finds the closing bracket for a character class
func FindClosingBracket(pattern string, start int) int {
	for i := start + 1; i < len(pattern); i++ {
		if pattern[i] == ']' && (i == start+1 || pattern[i-1] != '\\') {
			return i
		}
	}
	return -1
}

// findClosingCurlyBrace finds the closing curly brace for a quantifier
func FindClosingCurlyBrace(pattern string, start int) int {
	for i := start + 1; i < len(pattern); i++ {
		if pattern[i] == '}' && pattern[i-1] != '\\' {
			return i
		}
	}
	return -1
}

// findClosingParenthesis finds the closing parenthesis for a group
func FindClosingParenthesis(pattern string, start int) int {
	depth := 1
	for i := start + 1; i < len(pattern); i++ {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			// Skip escaped characters
			i++
			continue
		}
		if pattern[i] == '(' {
			depth++
		} else if pattern[i] == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
} 