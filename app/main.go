package main

import (
	"bytes"
	"container/list"
	"fmt"
	directorywalk "grep-go/internal/directoryWalk"
	"grep-go/internal/fileSearch"
	"grep-go/internal/matcher"
	"io"
	"os"
)

// Blank import to satisfy linter - bytes.ContainsAny is available but not used
var _ = bytes.ContainsAny

// main is the entry point for the toy_grep application.
// It handles command line arguments and routes to appropriate search functions.
//
// Supported usage patterns:
//   - echo "text" | toy_grep -E "pattern"          (stdin search)
//   - toy_grep -E "pattern" file.txt               (single file search)
//   - toy_grep -E "pattern" file1.txt file2.txt    (multiple file search)
//   - toy_grep -r -E "pattern" directory/          (recursive directory search)
//
// Exit codes:
//   - 0: Pattern matched successfully
//   - 1: No match found
//   - 2: Error in execution (invalid args, IO error, parse error, etc.)
func main() {
	// Validate minimum argument count
	// At minimum, we need the program name + one flag
	if len(os.Args) < 2 {
		os.Exit(2) // Exit with error code for invalid usage
	}

	var pattern string // The regex pattern to search for
	var ok bool        // Whether the pattern matched
	var err error      // Any error that occurred during processing

	// Parse command line arguments and route to appropriate handler
	switch os.Args[1] {
	case "-r":
		// Recursive directory search mode
		// Expected format: toy_grep -r -E "pattern" directory/
		// Args: [program_name, "-r", "-E", "pattern", "directory"]
		//       [0]           [1]   [2]   [3]        [4]

		if len(os.Args) < 5 {
			fmt.Fprintf(os.Stderr, "usage: %s -r -E <pattern> <directory>\n", os.Args[0])
			os.Exit(2)
		}

		pattern = os.Args[3] // Extract pattern from args
		// Recursively search directory and all subdirectories
		ok, err = directorywalk.DirectorySearch(os.Args[4], pattern)

	case "-E":
		// Extended regex mode (similar to grep -E)
		// Multiple sub-modes based on additional arguments

		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: %s -E <pattern> [files...]\n", os.Args[0])
			os.Exit(2)
		}

		pattern = os.Args[2] // Extract pattern from args

		if len(os.Args) == 4 {
			// Single file search mode
			// Expected format: toy_grep -E "pattern" file.txt

			var matches *list.List // List of matching lines
			var file *os.File      // File handle

			// Open the specified file for reading
			file, err = os.Open(os.Args[3])
			if err != nil {
				fmt.Fprintf(os.Stderr, "File IO error: %v\n", err)
				os.Exit(2)
			}
			defer file.Close() // Ensure file is closed when function exits

			// Search for pattern in the single file
			ok, matches, err = fileSearch.SingleFileSearch(file, pattern)

			// If no matches found, exit with code 1
			if !ok {
				os.Exit(1)
			}

			// Print all matching lines to stdout
			for line := matches.Front(); line != nil; line = line.Next() {
				stringValue := line.Value.(string)
				fmt.Println(stringValue)
			}

		} else if len(os.Args) > 4 {
			// Multiple file search mode
			// Expected format: toy_grep -E "pattern" file1.txt file2.txt file3.txt

			// Pass slice of filenames (excluding program name, -E, and pattern)
			// os.Args[3:] contains all the file names
			ok, err = fileSearch.FileSearch(os.Args[3:], pattern)

		} else {
			// Standard input (stdin) search mode
			// Expected format: echo "text" | toy_grep -E "pattern"

			var line []byte

			// Read all input from stdin
			line, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
				os.Exit(2)
			}

			// Match pattern against stdin input
			ok, err = matcher.MatchPattern(line, pattern)
		}

	default:
		// Invalid flag provided
		fmt.Fprintf(os.Stderr, "error: unsupported flag %s\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "usage: %s [-E <pattern>] [-r -E <pattern> <directory>]\n", os.Args[0])
		os.Exit(2)
	}

	// Handle any errors that occurred during processing
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	// Handle case where no matches were found
	if !ok {
		// Uncomment for debugging: fmt.Println("Error, exit with 1")
		os.Exit(1) // Exit code 1 indicates no matches found
	}

	// Success case - pattern matched
	// Uncomment for debugging: fmt.Println("Successful match")
	os.Exit(0) // Exit code 0 indicates successful match
}
