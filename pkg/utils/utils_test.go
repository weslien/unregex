package utils

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	// Just verify it's not empty
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestGetVersionInfo(t *testing.T) {
	info := GetVersionInfo()
	
	// Check that it contains the version
	if !strings.Contains(info, Version) {
		t.Errorf("GetVersionInfo() = %q, should contain version %q", info, Version)
	}
	
	// Check that it contains the git commit
	if !strings.Contains(info, GitCommit) {
		t.Errorf("GetVersionInfo() = %q, should contain git commit %q", info, GitCommit)
	}
	
	// Check that it contains the build date
	if !strings.Contains(info, BuildDate) {
		t.Errorf("GetVersionInfo() = %q, should contain build date %q", info, BuildDate)
	}
}

func TestDescription(t *testing.T) {
	desc := Description()
	
	// Check that it's not empty
	if desc == "" {
		t.Error("Description should not be empty")
	}
	
	// Check that it contains "regex" (case insensitive)
	lowerDesc := strings.ToLower(desc)
	if !strings.Contains(lowerDesc, "regex") && !strings.Contains(lowerDesc, "regular expression") {
		t.Errorf("Description() = %q, should contain 'regex' or 'regular expression'", desc)
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		format string
		want   bool
	}{
		{"go", true},
		{"pcre", true},
		{"posix", true},
		{"js", true},
		{"python", true},
		{"invalid", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if got := IsValidFormat(tt.format); got != tt.want {
				t.Errorf("IsValidFormat(%q) = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestGetFormatName(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"go", "Go Regexp"},
		{"pcre", "Perl Compatible Regular Expressions (PCRE)"},
		{"posix", "POSIX Extended Regular Expressions"},
		{"js", "JavaScript RegExp"},
		{"python", "Python re"},
		{"invalid", "Unknown Format"},
		{"", "Unknown Format"},
	}
	
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if got := GetFormatName(tt.format); got != tt.want {
				t.Errorf("GetFormatName(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
} 