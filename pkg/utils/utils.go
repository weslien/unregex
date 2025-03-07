package utils

// Version information set during build by the Makefile
var (
	// Version is the semantic version of the application
	Version = "0.2.2"
	
	// GitCommit is the git commit hash of the build
	GitCommit = "unknown"
	
	// BuildDate is the date when the application was built
	BuildDate = "unknown"
)

// GetVersionInfo returns a formatted string with the version information
func GetVersionInfo() string {
	return "Unregex " + Version + " (" + GitCommit + ") built on " + BuildDate
}

// Description returns a short description of the application
func Description() string {
	return "A tool to visualize and explain regular expressions"
}

// FormatPattern formats a regex pattern for display
func FormatPattern(pattern string) string {
	return pattern
}

// IsValidFormat checks if the specified regex format is supported
func IsValidFormat(format string) bool {
	validFormats := map[string]bool{
		"go":     true,
		"pcre":   true,
		"posix":  true,
		"js":     true,
		"python": true,
	}
	
	return validFormats[format]
}

// GetFormatName returns a readable name for the format
func GetFormatName(format string) string {
	formatNames := map[string]string{
		"go":     "Go Regexp",
		"pcre":   "Perl Compatible Regular Expressions (PCRE)",
		"posix":  "POSIX Extended Regular Expressions",
		"js":     "JavaScript RegExp",
		"python": "Python re",
	}
	
	if name, ok := formatNames[format]; ok {
		return name
	}
	return "Unknown Format"
} 