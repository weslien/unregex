package format

import (
	"reflect"
	"strings"
	"testing"
)

func TestJsFormat_Name(t *testing.T) {
	format := NewJsFormat()
	expected := "JavaScript RegExp"
	
	if got := format.Name(); got != expected {
		t.Errorf("JsFormat.Name() = %v, want %v", got, expected)
	}
}

func TestJsFormat_HasFeature(t *testing.T) {
	format := NewJsFormat()
	
	tests := []struct {
		feature string
		want    bool
	}{
		{FeatureLookahead, true},
		{FeatureLookbehind, true}, // Newer JS engines support this
		{FeatureNamedGroup, true}, // Newer JS engines support this
		{FeatureAtomicGroup, false},
		{FeatureConditional, false},
		{FeaturePossessive, false},
		{FeatureUnicodeClass, true}, // With /u flag
		{FeatureRecursion, false},
		{FeatureBackreference, true},
		{FeatureNamedBackref, true}, // Newer JS engines support this
		{"nonexistent", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.feature, func(t *testing.T) {
			if got := format.HasFeature(tt.feature); got != tt.want {
				t.Errorf("JsFormat.HasFeature(%q) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}
}

func TestJsFormat_TokenizeRegex(t *testing.T) {
	format := NewJsFormat()
	
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
			"Pattern with flags",
			"/abc/gi",
			[]string{"/gi", "abc"},
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
			"Named group - JS syntax",
			"(?<name>abc)",
			[]string{"(?<name>", "abc", ")"},
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
			"Positive lookbehind (newer JS)",
			"(?<=foo)bar",
			[]string{"(?<=", "foo", ")", "bar"},
		},
		{
			"Negative lookbehind (newer JS)",
			"(?<!foo)bar",
			[]string{"(?<!", "foo", ")", "bar"},
		},
		{
			"Non-greedy quantifiers",
			"a*?b+?c??",
			[]string{"a", "*?", "b", "+?", "c", "??"},
		},
		{
			"Curly brace quantifier",
			"a{2,3}",
			[]string{"a", "{2,3}"},
		},
		{
			"Complex pattern with flags",
			"/^(?<proto>https?):\\/\\/(?:www\\.)?[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}(\\/.*)?$/gimsu",
			[]string{
				"/gimsu", "^", "(?<proto>", "https", "?", ")", ":", "\\/", "\\/", "(?:", "www", "\\.", ")", "?", 
				"[a-zA-Z0-9.-]", "+", "\\.", "[a-zA-Z]", 
				"{2,}", "(", "\\/", ".", "*", ")", "?", "$",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := format.TokenizeRegex(tt.pattern)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsFormat.TokenizeRegex(%q):\ngot:  %q\nwant: %q", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestJsFormat_ExplainToken(t *testing.T) {
	format := NewJsFormat()
	
	tests := []struct {
		token string
		want  string
	}{
		{"/g", "Global search"},
		{"/i", "Case-insensitive"},
		{"/m", "Multi-line search"},
		{"/s", "Dot-all mode"},
		{"/u", "Unicode mode"},
		{"/y", "Sticky mode"},
		{"/gimuy", "Global search"},
		{"^", "Matches the start of a line"},
		{"$", "Matches the end of a line"},
		{".", "Matches any single character except newline"},
		{"*", "Matches 0 or more of the preceding element (greedy)"},
		{"+", "Matches 1 or more of the preceding element (greedy)"},
		{"?", "Matches 0 or 1 of the preceding element (greedy)"},
		{"*?", "Matches 0 or more of the preceding element (non-greedy)"},
		{"+?", "Matches 1 or more of the preceding element (non-greedy)"},
		{"??", "Matches 0 or 1 of the preceding element (non-greedy)"},
		{"|", "Acts as an OR operator"},
		{"(", "Start of a capturing group"},
		{")", "End of a capturing group"},
		{"(?:", "Start of a non-capturing group"},
		{"(?=", "Start of a positive lookahead"},
		{"(?!", "Start of a negative lookahead"},
		{"(?<=", "Start of a positive lookbehind"},
		{"(?<!", "Start of a negative lookbehind"},
		{"(?<name>", "Start of a named capturing group called 'name'"},
		{"[a-z]", "Matches any character in the set: a-z"},
		{"[^0-9]", "Matches any character NOT in the set: 0-9"},
		{"\\d", "Matches any digit (0-9)"},
		{"\\w", "Matches any word character"},
		{"\\s", "Matches any whitespace character"},
		{"\\u0061", "Matches the Unicode character U+0061"},
		{"\\x41", "Matches the character with hex code 41"},
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
				t.Errorf("JsFormat.ExplainToken(%q) = %q, want it to contain %q", tt.token, got, tt.want)
			}
		})
	}
} 