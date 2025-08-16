package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
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

	fmt.Fprintf(os.Stdout, "Successful exec")
	os.Exit(0)
}

func matchPattern(line []byte, pattern string) (bool, error) {
	m := make(map[rune]bool)
	single_match_allowed := false
	switch pattern {
	case "\\w":
		// tests for presence of atleast 1 alpha numeric char
		m = createUniversalChars(m)
		m['_'] = true
		single_match_allowed = true

	case "\\d":
		// tests for presence of atleast 1 number
		m = generatePatternFromRange(m, '0', '9')
		single_match_allowed = true

	default:
		// tests for presence or absence of a set of given numbers
		m = generatePatternFromChars(m, pattern)
		single_match_allowed = true
	}

	val, err := matchPat(line, m, single_match_allowed)
	return val, err
}

func generatePatternFromRange(m map[rune]bool, start rune, end rune) map[rune]bool {
	// fmt.Println("Generate from range", string(start), string(end))
	for i := start; i <= end; i++ {
		m[i] = true
	}
	return m
}

func generatePatternFromChars(m map[rune]bool, line string) map[rune]bool {
	// fmt.Println("Generate from chars", line)
	var inside string

	if len(line) > 1 {
		inside = line[1 : len(line)-1] // when [abc] is given
		if inside[0] == '^' {          // when [^abc] is given
			universal_chars := make(map[rune]bool)

			universal_chars = createUniversalChars(universal_chars)

			for _, val := range inside[1:] {
				delete(universal_chars, val)
			}

			m = universal_chars
			return universal_chars
		}
	} else {
		inside = line // when "a" is given
	}

	for _, val := range inside {
		m[val] = true // Add all characters in the map
	}
	return m
}

func createUniversalChars(m map[rune]bool) map[rune]bool {

	m = generatePatternFromRange(m, 'a', 'z')
	m = generatePatternFromRange(m, 'A', 'Z')
	m = generatePatternFromRange(m, '0', '9')

	return m
}

func matchPat(line []byte, pattern map[rune]bool, single_match_allowed bool) (bool, error) {
	fmt.Println("===MatchPat===")
	fmt.Println(string(line), single_match_allowed)
	fmt.Print("Map: ")
	for k := range pattern {
		fmt.Print(string(k))
	}
	fmt.Println()
	fmt.Println("==============")

	single_not_match := false

	for _, val := range string(line) {
		_, ok := pattern[val]
		if !ok {
			single_not_match = true
		} else if single_match_allowed {
			return true, nil
		}
	}

	return !single_not_match, nil
}
