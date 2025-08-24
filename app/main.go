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

var _ = bytes.ContainsAny

// Usage: echo <input_text> | toy_grep.sh -E <pattern>
// ./toy_grep.sh -E <pattern> any_file.txt

func main() {
	if len(os.Args) < 2 {
		os.Exit(2)
	}

	var pattern string
	var ok bool
	var err error

	switch os.Args[1] {
	case "-r":
		pattern = os.Args[3]
		// fmt.Println("Looking up dir for pattern", pattern)
		ok, err = directorywalk.DirectorySearch(os.Args[4], pattern)
	case "-E":
		pattern := os.Args[2]
		// fmt.Println("Checking files")
		if len(os.Args) == 4 {

			var matches *list.List

			var file *os.File
			file, err = os.Open(os.Args[3])
			if err != nil {
				fmt.Fprintf(os.Stderr, "File io error: %v\n", err)
			}
			ok, matches, err = fileSearch.SingleFileSearch(file, os.Args[2])
			if !ok {
				os.Exit(1)
			}
			for lin := matches.Front(); lin != nil; lin = lin.Next() {
				string_value := lin.Value.(string)
				fmt.Println(string_value)
			}

			defer file.Close()

		} else if len(os.Args) > 4 {

			ok, err = fileSearch.FileSearch(os.Args[3:], os.Args[2])

		} else {
			var line []byte
			line, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
				os.Exit(2)
			}

			ok, err = matcher.MatchPattern(line, pattern)

		}
	default:
		break
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
