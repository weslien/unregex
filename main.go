// Package main provides a simplified entrypoint for the unregex tool
// This allows users to install with: go install github.com/weslien/unregex@v0.1.0
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/weslien/unregex/internal/app"
	"github.com/weslien/unregex/pkg/utils"
)

func main() {
	// Define command-line flags
	formatFlag := flag.String("format", "go", "Regex format/flavor (go, pcre, posix, js, python)")
	visualizeFlag := flag.Bool("visualize", false, "Output visual annotation of the regex with numbered parts")
	helpFlag := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version information")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Unregex - %s\n\n", utils.Description())
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  unregex [options] <pattern>\n")
		fmt.Fprintf(os.Stderr, "  echo '<pattern>' | unregex [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  unregex \"^hello(world|universe)[0-9]+$\"\n")
		fmt.Fprintf(os.Stderr, "  unregex -format pcre \"(?<=look)behind\"\n")
		fmt.Fprintf(os.Stderr, "  unregex -visualize \"a{2,4}b[a-z]*\\d+\"\n")
		fmt.Fprintf(os.Stderr, "  echo \"a{2,4}b[a-z]*\\d+\" | unregex\n")
	}

	// Parse command-line flags
	flag.Parse()

	// Show help message and exit
	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	// Show version information and exit
	if *versionFlag {
		fmt.Println(utils.GetVersionInfo())
		os.Exit(0)
	}

	fmt.Printf("Unregex - Regex Visualizer v%s\n\n", utils.Version)

	// Validate regex format
	format := strings.ToLower(*formatFlag)
	if !utils.IsValidFormat(format) {
		fmt.Fprintf(os.Stderr, "Error: Unsupported regex format '%s'\n", format)
		fmt.Fprintf(os.Stderr, "Supported formats: go, pcre, posix, js, python\n")
		os.Exit(1)
	}

	// Get regex pattern from arguments or stdin
	pattern, err := getRegexPattern()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'unregex -help' for usage information")
		os.Exit(1)
	}

	// Run the regex explanation with the selected format
	if err := app.Run([]string{pattern, format, fmt.Sprintf("%v", *visualizeFlag)}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// getRegexPattern retrieves the regex pattern from command line arguments or stdin
func getRegexPattern() (string, error) {
	// Check if pattern is provided as a command line argument (after flags)
	if flag.NArg() > 0 {
		return flag.Arg(0), nil
	}

	// Check if data is being piped in through stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped in
		reader := bufio.NewReader(os.Stdin)
		input, err := io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %v", err)
		}

		// Trim whitespace and newlines
		pattern := strings.TrimSpace(string(input))
		if pattern == "" {
			return "", fmt.Errorf("empty pattern received from stdin")
		}

		return pattern, nil
	}

	// No pattern provided
	return "", fmt.Errorf("no regex pattern provided")
}
