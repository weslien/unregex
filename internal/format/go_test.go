package format

import (
	"reflect"
	"strings"
	"testing"
)

func TestGoFormat_Name(t *testing.T) {
	format := NewGoFormat()
	expected := "Go Regexp"
	
	if got := format.Name(); got != expected {
		t.Errorf("GoFormat.Name() = %v, want %v", got, expected)
	}
}

func TestGoFormat_HasFeature(t *testing.T) {
	format := NewGoFormat()
	
	tests := []struct {
		feature string
		want    bool
	}{
		{FeatureLookahead, true},
		{FeatureLookbehind, false},
		{FeatureNamedGroup, true},
		{FeatureAtomicGroup, false},
		{FeatureConditional, false},
		{FeaturePossessive, false},
		{FeatureUnicodeClass, true},
		{FeatureRecursion, false},
		{FeatureBackreference, true},
		{FeatureNamedBackref, true},
		{"nonexistent", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.feature, func(t *testing.T) {
			if got := format.HasFeature(tt.feature); got != tt.want {
				t.Errorf("GoFormat.HasFeature(%q) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}
}

func TestGoFormat_TokenizeRegex(t *testing.T) {
	format := NewGoFormat()
	
	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			"Simple pattern",
			"abc",
			[]string{"abc"},
		},
		{
			"Character class",
			"[a-z]",
			[]string{"[a-z]"},
		},
		{
			"Anchors and quantifiers",
			"^abc+$",
			[]string{"^", "abc", "+", "$"},
		},
		{
			"Groups and alternation",
			"(foo|bar)",
			[]string{"(", "foo", "|", "bar", ")"},
		},
		{
			"Escape sequences",
			"\\d\\w\\s",
			[]string{"\\d", "\\w", "\\s"},
		},
		{
			"Named group",
			"(?P<name>abc)",
			[]string{"(?P<name>", "abc", ")"},
		},
		{
			"Non-capturing group",
			"(?:abc)",
			[]string{"(?:", "abc", ")"},
		},
		{
			"Positive lookahead",
			"foo(?=bar)",
			[]string{"foo", "(?=", "bar", ")"},
		},
		{
			"Curly brace quantifier",
			"a{2,3}",
			[]string{"a", "{2,3}"},
		},
		{
			"Complex pattern",
			"^(https?://)?[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}(/.*)?$",
			[]string{
				"^", "(", "https", "?", "://", ")", "?", 
				"[a-zA-Z0-9.-]", "+", "\\.", "[a-zA-Z]", 
				"{2,}", "(", "/", ".", "*", ")", "?", "$",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.TokenizeRegex(tt.pattern)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoFormat.TokenizeRegex(%q):\ngot:  %q\nwant: %q", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestGoFormat_ExplainToken(t *testing.T) {
	format := NewGoFormat()
	
	tests := []struct {
		token string
		want  string
	}{
		{"^", "Matches the start of a line"},
		{"$", "Matches the end of a line"},
		{".", "Matches any single character except newline"},
		{"*", "Matches 0 or more of the preceding element"},
		{"+", "Matches 1 or more of the preceding element"},
		{"?", "Matches 0 or 1 of the preceding element"},
		{"|", "Acts as an OR operator - matches the expression before or after the |"},
		{"(", "Start of a capturing group"},
		{")", "End of a capturing group"},
		{"(?:", "Start of a non-capturing group - groups the expression but doesn't create a capture group"},
		{"(?=", "Start of a positive lookahead - matches if the pattern inside matches, but doesn't consume characters"},
		{"(?P<name>", "Start of a named capturing group called 'name'"},
		{"[a-z]", "Matches any character in the set: a-z"},
		{"[^0-9]", "Matches any character NOT in the set: 0-9"},
		{"\\d", "Matches any digit (0-9)"},
		{"\\w", "Matches any word character (alphanumeric plus underscore)"},
		{"\\s", "Matches any whitespace character (space, tab, newline, etc.)"},
		{"{2,3}", "Matches between 2 and 3 occurrences of the preceding element"},
		{"{2,}", "Matches at least 2 occurrences of the preceding element"},
		{"{3}", "Matches exactly 3 occurrences of the preceding element"},
		{"a", "Matches the character 'a' literally"},
		{"abc", "Matches the string 'abc' literally"},
	}
	
	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got := format.ExplainToken(tt.token)
			if !strings.Contains(got, tt.want) {
				t.Errorf("GoFormat.ExplainToken(%q) = %q, want it to contain %q", tt.token, got, tt.want)
			}
		})
	}
} 