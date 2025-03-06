# Unregex

A Go CLI application for visualizing and explaining regular expressions in human-readable form.

## Purpose

Regular expressions can be difficult to understand at a glance. Unregex breaks down a regex pattern into its components and explains what each part does, making it easier to understand, debug, or learn regular expressions.

## Installation

### Prerequisites

- Go 1.21 or higher

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/unregex.git
cd unregex

# Build using make
make build

# The binary will be in the build directory
./build/unregex -help

# Or install to your GOPATH/bin
make install
```

### Using Make

The project includes a Makefile with several useful commands:

```bash
make build           # Build the application
make run ARGS='args' # Build and run with arguments
make install         # Install to GOPATH/bin
make clean           # Remove build artifacts
make test            # Run tests
make fmt             # Format code
make help            # Show all available commands
```

For more options, run `make help`.

## Usage

You can provide a regular expression pattern in two ways:

### As a command-line argument:

```bash
./unregex "^hello(world|universe)[0-9]+$"
```

### Via stdin (pipe):

```bash
echo "^hello(world|universe)[0-9]+$" | ./unregex
```

### Specifying a Regex Format

You can specify which regex format/flavor to use with the `-format` flag:

```bash
./unregex -format pcre "(?<=look)behind"
```

Supported formats are:
- `go`: Go's regexp package (default)
- `pcre`: Perl Compatible Regular Expressions
- `posix`: POSIX Extended Regular Expressions
- `js`: JavaScript RegExp
- `python`: Python's re module

Each format supports different features and has slightly different syntax.

### Other Options

```
./unregex -help    # Display help information
./unregex -version # Display version information
```

## Example

For the regex pattern `^hello(world|universe)[0-9]+$` with Go format, the output might look like:

```
Unregex - Regex Visualizer v0.1.0

Analyzing regex pattern: ^hello(world|universe)[0-9]+$
Format: Go Regexp

Supported Features:
  ✓ Lookahead ((?=pattern) or (?!pattern))
  ✗ Lookbehind ((?<=pattern) or (?<!pattern))
  ✓ Named Groups ((?P<name>pattern))
  ✗ Atomic Groups ((?>pattern))
  ✗ Conditionals ((?(cond)then|else))
  ✗ Possessive Quantifiers (a++, a*+, a?+)
  ✓ Unicode Properties (\p{Property})
  ✗ Recursion ((?R) or (?0))
  ✓ Backreferences (\1, \2, etc.)
  ✓ Named Backreferences (\k<name>)

1. ^: Matches the start of a line
2. hello: Matches the string 'hello' literally
3. (: Start of a capturing group
4. world: Matches the string 'world' literally
5. |: Acts as an OR operator - matches the expression before or after the |
6. universe: Matches the string 'universe' literally
7. ): End of a capturing group
8. [0-9]: Matches any character in the set: 0-9
9. +: Matches 1 or more of the preceding element
10. $: Matches the end of a line

NOTE: This is a basic regex explainer. Some complex patterns might not be perfectly tokenized.
```

## Development

### Project Structure

```
unregex/
├── cmd/                  # Main applications for this project
│   └── myapp/            # Command-line client
│       └── main.go       # Command-line entry point
├── pkg/                  # Library code that can be used by other applications
│   └── utils/            # Utility functions
│       └── utils.go      # Utility functions
├── internal/             # Private application and library code
│   ├── app/              # Application logic
│   │   └── app.go        # Core application functionality
│   └── format/           # Regex format implementations 
│       ├── format.go     # Format interface and common utilities
│       ├── go.go         # Go regexp implementation
│       ├── pcre.go       # PCRE implementation
│       ├── posix.go      # POSIX ERE implementation
│       ├── js.go         # JavaScript RegExp implementation
│       └── python.go     # Python re implementation
├── go.mod                # Go module definition
├── go.sum                # Go module checksums (generated when dependencies are added)
├── README.md             # Documentation
└── LICENSE               # License file
```

### Adding dependencies

```bash
go get github.com/example/package
```

## License

[MIT](LICENSE)