package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

var _ = bytes.ContainsAny

// Usage: echo <input_text> | toy_grep.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	// flag := os.Args[1]
	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assuming we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchPattern(line, pattern)

	if err != nil {
		fmt.Println("err != nil")
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		fmt.Println("!ok")
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// fmt.Fprintf(os.Stdout, "Successful exec")
	// default exit code is 0 which means success
	os.Exit(0)
}

func matchPattern(line []byte, pattern string) (bool, error) {
	// fmt.Println("this is the match pattern function", pattern)
	// fmt.Println(pattern)

	if pattern == "\\d" {
		return matchDigit(line)
	} else if pattern == "\\w" {
		return matchWord(line)
	}
	{
		return matchLine(line, pattern)
	}
}

func matchWord(line []byte) (bool, error) {
	// fmt.Println("this is the match word function")

	for _, char := range string(line) {
		num := char - '0'
		cap := char - 'A'
		sml := char - 'a'

		if !((num >= 0 && num <= 9) ||
			(cap >= 0 && cap <= 25) ||
			(sml >= 0 && sml <= 25) ||
			char == '_') {
			return false, nil
		}
	}
	// fmt.Println("matchWord completed successfully")
	return true, nil
}

func matchDigit(line []byte) (bool, error) {

	for _, char := range string(line) {
		num := char - '0'
		if num >= 0 && num <= 9 {
			return true, nil
		}
	}

	return false, nil
}

func matchLine(line []byte, pattern string) (bool, error) {
	// fmt.Println("this is the match line function")
	if utf8.RuneCountInString(pattern) != 1 {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	var ok bool = bytes.ContainsAny(line, pattern)

	return ok, nil
}
