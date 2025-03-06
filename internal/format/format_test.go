package format

import (
	"testing"
)

// TestFormatImplementations tests that all format implementations satisfy the interface requirements
func TestFormatImplementations(t *testing.T) {
	// Assert that all format implementations satisfy the RegexFormat interface
	var _ RegexFormat = &GoFormat{}
	var _ RegexFormat = &PcreFormat{}
	var _ RegexFormat = &PosixFormat{}
	var _ RegexFormat = &JsFormat{}
	var _ RegexFormat = &PythonFormat{}
}

// TestGetFormat tests the GetFormat function with various formats
func TestGetFormat(t *testing.T) {
	tests := []struct {
		name       string
		formatName string
		wantType   string
	}{
		{"Go format", "go", "*format.GoFormat"},
		{"PCRE format", "pcre", "*format.PcreFormat"},
		{"POSIX format", "posix", "*format.PosixFormat"},
		{"JavaScript format", "js", "*format.JsFormat"},
		{"Python format", "python", "*format.PythonFormat"},
		{"Unknown format defaults to Go", "unknown", "*format.GoFormat"},
		{"Empty format defaults to Go", "", "*format.GoFormat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFormat(tt.formatName)
			gotType := getFormatType(got)
			if gotType != tt.wantType {
				t.Errorf("GetFormat(%q) = %v, want type %v", tt.formatName, gotType, tt.wantType)
			}
		})
	}
}

// TestHelperFunctions tests the helper functions like FindClosingBracket
func TestHelperFunctions(t *testing.T) {
	// Test FindClosingBracket
	t.Run("FindClosingBracket", func(t *testing.T) {
		tests := []struct {
			pattern string
			start   int
			want    int
		}{
			{"[abc]", 0, 4},
			{"[a-z]", 0, 4},
			{"[^0-9]", 0, 5},
			{"[\\]]", 0, 3},  // Escaped closing bracket
			{"[]abc]", 0, 1},  // Closing bracket at beginning is literal
			{"[invalid", 0, -1}, // No closing bracket
		}

		for _, tt := range tests {
			got := FindClosingBracket(tt.pattern, tt.start)
			if got != tt.want {
				t.Errorf("FindClosingBracket(%q, %d) = %d, want %d", tt.pattern, tt.start, got, tt.want)
			}
		}
	})

	// Test FindClosingCurlyBrace
	t.Run("FindClosingCurlyBrace", func(t *testing.T) {
		tests := []struct {
			pattern string
			start   int
			want    int
		}{
			{"{3}", 0, 2},
			{"{1,3}", 0, 4},
			{"{2,}", 0, 3},
			{"{\\}}", 0, 3},  // Escaped closing brace
			{"{invalid", 0, -1}, // No closing brace
		}

		for _, tt := range tests {
			got := FindClosingCurlyBrace(tt.pattern, tt.start)
			if got != tt.want {
				t.Errorf("FindClosingCurlyBrace(%q, %d) = %d, want %d", tt.pattern, tt.start, got, tt.want)
			}
		}
	})

	// Test FindClosingParenthesis
	t.Run("FindClosingParenthesis", func(t *testing.T) {
		tests := []struct {
			pattern string
			start   int
			want    int
		}{
			{"(abc)", 0, 4},
			{"(a(b)c)", 0, 6},
			{"(a(b)(c)d)", 0, 9},
			{"(\\))", 0, 3},  // Escaped closing parenthesis
			{"(invalid", 0, -1}, // No closing parenthesis
		}

		for _, tt := range tests {
			got := FindClosingParenthesis(tt.pattern, tt.start)
			if got != tt.want {
				t.Errorf("FindClosingParenthesis(%q, %d) = %d, want %d", tt.pattern, tt.start, got, tt.want)
			}
		}
	})
}

// Helper function to get format type name for testing
func getFormatType(f RegexFormat) string {
	switch f.(type) {
	case *GoFormat:
		return "*format.GoFormat"
	case *PcreFormat:
		return "*format.PcreFormat"
	case *PosixFormat:
		return "*format.PosixFormat"
	case *JsFormat:
		return "*format.JsFormat"
	case *PythonFormat:
		return "*format.PythonFormat"
	default:
		return "unknown"
	}
} 