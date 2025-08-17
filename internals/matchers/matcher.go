package matchers

import (
	"container/list"
	"fmt"
	"grep-go/internals/parsers"
)

func MatchPattern(line []byte, pattern string) (bool, error) {
	runes := []rune(string(line))

	patterns, err := parsers.ParsePatterns(pattern)
	fmt.Println("== Patterns == ")
	for ele := patterns.Front(); ele != nil; ele = ele.Next() {
		fmt.Print(ele.Value.(string), " ")
	}
	fmt.Println()
	fmt.Println("== done ==")

	if err != nil {
		return false, err
	}

	var first string = patterns.Front().Value.(string)
	var start_index int
	var status bool

	if first[0] == '^' {
		// handle string anchor for string beginning
		status, start_index, err = matchIndividualPattern(runes, first[1:], 0, nil)

		if err != nil || !status {
			return false, err
		}

		patterns.Remove(patterns.Front())
	} else {
		start_index = 0
	}

	// Try matching the entire pattern sequence starting at each position
	for startPos := start_index; startPos <= len(runes); startPos++ {
		if matchPatternsFromPosition(runes, patterns.Front(), startPos) {
			return true, nil
		}
	}

	return false, nil
}

func matchPatternsFromPosition(runes []rune, patterns *list.Element, startPos int) bool {
	currentPos := startPos
	end_detect := false
	for pat := patterns; pat != nil; pat = pat.Next() {
		pat_string := pat.Value.(string)
		end_detect = false

		if pat_string[len(pat_string)-1] == '$' {
			pat_string = pat_string[0 : len(pat_string)-1]
			end_detect = true
		}

		found, nextPos, err := matchIndividualPattern(runes, pat_string, currentPos, pat)

		if err != nil {
			fmt.Println(err)
			return false
		}

		currentPos = nextPos

		if end_detect {
			// handle string anchor for string beginning
			input_text_length := len(string(runes))
			if found && input_text_length == currentPos {
				return true
			} else {
				return false
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func matchIndividualPattern(runes []rune, pattern string, index int, pat *list.Element) (bool, int, error) {
	if index > len(runes) {
		return false, -1, nil
	}

	if index == -1 {
		return false, -1, fmt.Errorf("invalid index detected: %d", index)
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

	case pattern[0] == '+':
		// matching +
		return matchOneOrMoreBacktracking(runes, pattern, index, pat)

	case pattern[0] == '?':
		return matchOneOrNone(runes, pattern, index)

	case len(pattern) > 0 && pattern[0] == '[':
		return matchCharacterClass(runes, pattern, index)

	default:
		// Literal substring match
		return matchCompleteSubString(runes, pattern, index)
	}
}

func matchOneOrNone(runes []rune, pattern string, index int) (bool, int, error) {
	if index >= len(runes) {
		return false, -1, nil
	}

	// Remove the + to get the character/pattern to match
	basePattern := pattern[1:]

	// Handle different base patterns
	var matches func(rune) bool

	if basePattern == "\\d" {
		matches = isDigit
	} else if basePattern == "\\w" {
		matches = isAlphanumeric
	} else if len(basePattern) == 1 {
		// Single character
		char := rune(basePattern[0])
		matches = func(r rune) bool { return r == char }
	} else {
		// For more multi character pattern, custom implementation is needed
		return false, -1, fmt.Errorf("unsupported pattern with +: %s", pattern)
	}

	// case when no occourance is detected
	if !matches(runes[index]) {
		// fmt.Println("First index no match", matches(runes[index]), string(runes[index]))
		return true, index, nil
	} else {
		return true, index + 1, nil
	}
}

func matchOneOrMoreBacktracking(runes []rune, pattern string, index int, pat *list.Element) (bool, int, error) {
	if index >= len(runes) {
		return false, -1, nil
	}

	// Remove the + to get the character/pattern to match
	basePattern := pattern[1:]

	// Handle different base patterns
	var matches func(rune) bool

	if basePattern == "\\d" {
		matches = isDigit
	} else if basePattern == "\\w" {
		matches = isAlphanumeric
	} else if len(basePattern) == 1 {
		// Single character
		char := rune(basePattern[0])
		matches = func(r rune) bool { return r == char }
	} else {
		// For more multi character pattern, custom implementation is needed
		return false, -1, fmt.Errorf("unsupported pattern with +: %s", pattern)
	}

	// case when no occourance is detected
	if !matches(runes[index]) {
		fmt.Println("First index no match", matches(runes[index]), string(runes[index]))
		return false, -1, nil
	}

	// Match as many as possible while looking ahead
	i := index
	pat = pat.Next()

	for i < len(runes) && matches(runes[i]) { // this only checks for overlapping queries
		if matchPatternsFromPosition(runes, pat, i) {
			fmt.Println("total match found under overlapping + query")
			return true, i, nil
		}
		i++
	}

	// this is in case there is no overlap
	return true, index + 1, nil
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

	// fmt.Println("Checking:", pattern, index)

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
