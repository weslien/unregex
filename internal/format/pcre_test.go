package format

import (
	"reflect"
	"strings"
	"testing"
)

func TestPcreFormat_Name(t *testing.T) {
	format := NewPcreFormat()
	expected := "Perl Compatible Regular Expressions (PCRE)"
	
	if got := format.Name(); got != expected {
		t.Errorf("PcreFormat.Name() = %v, want %v", got, expected)
	}
}

func TestPcreFormat_HasFeature(t *testing.T) {
	format := NewPcreFormat()
	
	tests := []struct {
		feature string
		want    bool
	}{
		{FeatureLookahead, true},
		{FeatureLookbehind, true},
		{FeatureNamedGroup, true},
		{FeatureAtomicGroup, true},
		{FeatureConditional, true},
		{FeaturePossessive, true},
		{FeatureUnicodeClass, true},
		{FeatureRecursion, true},
		{FeatureBackreference, true},
		{FeatureNamedBackref, true},
		{"nonexistent", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.feature, func(t *testing.T) {
			if got := format.HasFeature(tt.feature); got != tt.want {
				t.Errorf("PcreFormat.HasFeature(%q) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}
}

func TestPcreFormat_TokenizeRegex(t *testing.T) {
	format := NewPcreFormat()
	
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
			"Named group - PCRE syntax",
			"(?<name>abc)",
			[]string{"(?<name>", "abc", ")"},
		},
		{
			"Named group - Python-compatible syntax",
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
			"Negative lookahead",
			"foo(?!bar)",
			[]string{"foo", "(?!", "bar", ")"},
		},
		{
			"Positive lookbehind",
			"(?<=foo)bar",
			[]string{"(?<=", "foo", ")", "bar"},
		},
		{
			"Negative lookbehind",
			"(?<!foo)bar",
			[]string{"(?<!", "foo", ")", "bar"},
		},
		{
			"Atomic group",
			"(?>atom)",
			[]string{"(?>", "atom", ")"},
		},
		{
			"Possessive quantifiers",
			"a++b*+c?+",
			[]string{"a", "++", "b", "*+", "c", "?+"},
		},
		{
			"Curly brace quantifier",
			"a{2,3}",
			[]string{"a", "{2,3}"},
		},
		{
			"Complex pattern",
			"^(?<proto>https?)://(?:www\\.)?[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}(/.*)?$",
			[]string{
				"^", "(?<proto>", "https", "?", ")", "://", "(?:", "www", "\\.", ")", "?", 
				"[a-zA-Z0-9.-]", "+", "\\.", "[a-zA-Z]", 
				"{2,}", "(", "/", ".", "*", ")", "?", "$",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.TokenizeRegex(tt.pattern)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PcreFormat.TokenizeRegex(%q):\ngot:  %q\nwant: %q", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestPcreFormat_ExplainToken(t *testing.T) {
	format := NewPcreFormat()
	
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
		{"*+", "Possessive match of 0 or more"},
		{"++", "Possessive match of 1 or more"},
		{"?+", "Possessive match of 0 or 1"},
		{"|", "Acts as an OR operator"},
		{"(", "Start of a capturing group"},
		{")", "End of a capturing group"},
		{"(?:", "Start of a non-capturing group"},
		{"(?=", "Start of a positive lookahead"},
		{"(?!", "Start of a negative lookahead"},
		{"(?<=", "Start of a positive lookbehind"},
		{"(?<!", "Start of a negative lookbehind"},
		{"(?>", "Start of an atomic group"},
		{"(?<name>", "Start of a named capturing group called 'name'"},
		{"(?P<name>", "Start of a named capturing group called 'name'"},
		{"[a-z]", "Matches any character in the set: a-z"},
		{"[^0-9]", "Matches any character NOT in the set: 0-9"},
		{"\\d", "Matches any digit (0-9)"},
		{"\\w", "Matches any word character"},
		{"\\s", "Matches any whitespace character"},
		{"\\G", "Matches the position where the previous match ended"},
		{"\\Q", "Start of a quoted sequence"},
		{"\\E", "End of a quoted sequence"},
		{"{2,3}", "Matches between 2 and 3 occurrences"},
		{"{2,}", "Matches at least 2 occurrences"},
		{"{3}", "Matches exactly 3 occurrences"},
		{"a", "Matches the character 'a' literally"},
		{"abc", "Matches the string 'abc' literally"},
	}
	
	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got := format.ExplainToken(tt.token)
			if !strings.Contains(got, tt.want) {
				t.Errorf("PcreFormat.ExplainToken(%q) = %q, want it to contain %q", tt.token, got, tt.want)
			}
		})
	}
} 