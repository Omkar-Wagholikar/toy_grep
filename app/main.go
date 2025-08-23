package main

import (
	"bytes"
	"fmt"
	"grep-go/internal/fileSearch"
	"grep-go/internal/matchers"
	"io"
	"os"
)

var _ = bytes.ContainsAny

// Usage: echo <input_text> | toy_grep.sh -E <pattern>
// ./toy_grep.sh -E <pattern> any_file.txt

func main() {

	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	var ok bool
	var err error

	if len(os.Args) == 4 {
		// fmt.Println("File io", os.Args[2], os.Args[3])
		var file *os.File
		file, err = os.Open(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "File io error: %v\n", err)
		}
		ok, err = fileSearch.FileSearch(file, os.Args[2])
		defer file.Close()

	} else {
		var line []byte
		line, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
			os.Exit(2)
		}

		ok, err = matchers.MatchPattern(line, pattern)

	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		// fmt.println("Error, exit with 1")
		os.Exit(1)
	}

	// fmt.println("Successful match")
	os.Exit(0)
}
