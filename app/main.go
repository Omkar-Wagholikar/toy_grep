package main

import (
	"bytes"
	"container/list"
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

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchPattern(line, pattern)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

	os.Exit(0)
}

func matchPattern(line []byte, pattern string) (bool, error) {
	runes := []rune(string(line))

	patterns, err := parsePatterns(pattern)
	if err != nil {
		return false, err
	}

	// Try matching the entire pattern sequence starting at each position
	for startPos := 0; startPos <= len(runes); startPos++ {
		if matchPatternsFromPosition(runes, patterns, startPos) {
			return true, nil
		}
	}

	return false, nil
}

func matchPatternsFromPosition(runes []rune, patterns *list.List, startPos int) bool {
	currentPos := startPos

	for pat := patterns.Front(); pat != nil; pat = pat.Next() {
		found, nextPos, err := matchIndividualPattern(runes, pat.Value.(string), currentPos)
		if err != nil || !found {
			return false
		}
		currentPos = nextPos
	}

	return true
}

func parsePatterns(pattern string) (*list.List, error) {
	patterns := list.New()
	runes := []rune(pattern)
	i := 0

	for i < len(runes) {
		if runes[i] == '\\' && i+1 < len(runes) {
			if i+1 < len(runes) && runes[i+1] == '\\' {
				// This is \\something - literal backslash followed by something
				if i+2 < len(runes) {
					esc := string(runes[i : i+3]) // "\\d" or "\\w"
					patterns.PushBack(esc)
					i += 3
				} else {
					// Just "\\" at end
					patterns.PushBack("\\")
					i += 2
				}
			} else {
				// This is \something - escape sequence
				esc := string(runes[i : i+2]) // "\d" or "\w"
				patterns.PushBack(esc)
				i += 2
			}
		} else if runes[i] == '[' {
			// Handle character class [abc] or [^abc]
			j := i + 1
			for j < len(runes) && runes[j] != ']' {
				j++
			}
			if j < len(runes) {
				charClass := string(runes[i : j+1]) // include the ]
				patterns.PushBack(charClass)
				i = j + 1
			} else {
				// No closing ], treat [ as literal
				patterns.PushBack("[")
				i++
			}
		} else {
			// Collect normal characters until next special character
			j := i
			for j < len(runes) && runes[j] != '\\' && runes[j] != '[' {
				j++
			}
			if j > i {
				substr := string(runes[i:j])
				patterns.PushBack(substr)
			}
			i = j
		}
	}
	return patterns, nil
}

func matchIndividualPattern(runes []rune, pattern string, index int) (bool, int, error) {
	if index > len(runes) {
		return false, -1, nil
	}

	switch {
	case pattern == "\\w":
		return matchSingleCharacter(runes, isAlphanumeric, index)

	case pattern == "\\d":
		return matchSingleCharacter(runes, isDigit, index)

	case pattern == "\\\\d":
		// literal substring "\d"
		return matchCompleteSubString(runes, `\d`, index)

	case pattern == "\\\\w":
		// literal substring "\w"
		return matchCompleteSubString(runes, `\w`, index)

	case len(pattern) > 0 && pattern[0] == '[':
		return matchCharacterClass(runes, pattern, index)

	default:
		// Literal substring match
		return matchCompleteSubString(runes, pattern, index)
	}
}

func matchCharacterClass(runes []rune, pattern string, index int) (bool, int, error) {
	if index >= len(runes) {
		return false, -1, nil
	}

	// Parse character class [abc] or [^abc]
	negated := false
	chars := pattern[1 : len(pattern)-1] // remove [ and ]

	if len(chars) > 0 && chars[0] == '^' {
		negated = true
		chars = chars[1:]
	}

	targetChar := runes[index]
	found := false

	// Check if character is in the class
	for _, char := range chars {
		if targetChar == char {
			found = true
			break
		}
	}

	if negated {
		found = !found
	}

	if found {
		return true, index + 1, nil
	}

	return false, -1, nil
}

func matchSingleCharacter(runes []rune, predicate func(rune) bool, index int) (bool, int, error) {
	if index >= len(runes) {
		return false, -1, nil
	}

	if predicate(runes[index]) {
		return true, index + 1, nil
	}

	return false, -1, nil
}

func matchCompleteSubString(runes []rune, pattern string, index int) (bool, int, error) {
	patRunes := []rune(pattern)

	if index+len(patRunes) > len(runes) {
		return false, -1, nil
	}

	for j := 0; j < len(patRunes); j++ {
		if runes[index+j] != patRunes[j] {
			return false, -1, nil
		}
	}

	return true, index + len(patRunes), nil
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}
