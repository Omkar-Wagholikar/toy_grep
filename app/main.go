package main

import (
	"bytes"
	"fmt"
	"grep-go/internal/matchers"
	"io"
	"os"
)

var _ = bytes.ContainsAny

// Usage: echo <input_text> | toy_grep.sh -E <pattern>
// ./toy_grep.sh -E <pattern> any_file.txt

// /path
// -E
// search_pattern
// any_file.txt

// /path
// -E
// [^abc]

func main() {

	for _, val := range os.Args {
		// fmt.println(val)
	}

	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	if len(os.Args) == 4 {
		// fmt.println("File io")

	} else {
		line, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
			os.Exit(2)
		}

		ok, err := matchers.MatchPattern(line, pattern)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}

		if !ok {
			// fmt.println("Error, exit with 1")
			os.Exit(1)
		}
	}

	// fmt.println("Successful match")
	os.Exit(0)
}
