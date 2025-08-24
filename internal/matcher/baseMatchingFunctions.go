package matcher

import (
	"container/list"
	"fmt"
)

func matchPatternsFromPosition(runes []rune, patterns *list.Element, startPos int) bool {
	currentPos := startPos
	end_detect := false

	for pat := patterns; pat != nil; pat = pat.Next() {
		pat_string := pat.Value.(string)
		end_detect = false

		if len(pat_string) > 0 && pat_string[len(pat_string)-1] == '$' {
			pat_string = pat_string[0 : len(pat_string)-1]
			end_detect = true
		}

		// fmt.println("Matching individual pattern:", pat.Value, "at", startPos)

		// Handle empty pattern after removing $
		if len(pat_string) == 0 {
			if end_detect {
				// This is just a $ anchor, check if we're at end of input
				input_text_length := len(runes)
				if currentPos == input_text_length {
					return true
				} else {
					return false
				}
			} else {
				// Empty pattern without $ - this shouldn't happen, skip it
				continue
			}
		}

		// fmt.println("Calling Match Indivi Pat")
		found, nextPos, err := matchIndividualPattern(runes, pat_string, currentPos, pat)
		// fmt.println("Value returned by mip: ", found)

		if err != nil {
			// fmt.println(err)
			return false
		}

		currentPos = nextPos

		if end_detect {
			// fmt.println("Inside end_detect")
			// handle string anchor for string end
			input_text_length := len(runes)
			if found && input_text_length == currentPos {
				return true
			} else {
				// fmt.Printf("End anchor failed: currentPos=%d, inputLength=%d\n", currentPos, input_text_length)
				return false
			}
		}

		if !found {
			// fmt.println("fail at:", pat.Value)
			return false
		}
	}

	return true
}

func matchIndividualPattern(runes []rune, pattern string, index int, pat *list.Element) (bool, int, error) {
	if len(pattern) == 0 {
		return false, -1, fmt.Errorf("empty pattern encountered @index: %d", index)
	}
	if index > len(runes) {
		return false, -1, nil
	}
	if index == -1 {
		return false, -1, fmt.Errorf("invalid index detected: %d", index)
	}

	switch {
	case pattern == "\\w":
		return matchSingleCharacter(runes, IsAlphanumeric, index)

	case pattern == "\\d":
		return matchSingleCharacter(runes, IsDigit, index)

	case pattern == "\\\\d":
		// literal substring "\d"
		return matchCompleteSubString(runes, `\d`, index)

	case pattern == "\\\\w":
		// literal substring "\w"
		return matchCompleteSubString(runes, `\w`, index)

	case pattern == ".+":
		// Handle .+ pattern (one or more of any character)
		return matchDotPlusBacktracking(runes, index, pat)

	case len(pattern) > 4 && pattern[:4] == "ALT:":
		// Handle simple alternation pattern
		// fmt.println("Matching ALT:")
		return matchAlternation(runes, pattern, index)

	case len(pattern) > 5 && pattern[:4] == "ALT+":
		// Handle alternation with + quantifier
		// fmt.println("Matching ALT+")
		return matchAlternationPlus(runes, pattern, index, pat)

	case len(pattern) > 5 && pattern[:4] == "ALT?":
		// Handle alternation with ? quantifier
		// fmt.println("Matching ALT?")
		return matchAlternationOptional(runes, pattern, index)

	case len(pattern) > 4 && pattern[:4] == "GRP:":
		// Handle simple group
		// fmt.println("Matching GRP:")
		return matchGroup(runes, pattern, index)

	case len(pattern) > 5 && pattern[:4] == "GRP+":
		// Handle group with + quantifier
		// fmt.println("Matching GRP+")
		return matchGroupPlus(runes, pattern, index, pat)

	case len(pattern) > 5 && pattern[:4] == "GRP?":
		// Handle group with ? quantifier
		// fmt.println("Matching GRP?")
		return matchGroupOptional(runes, pattern, index)

	case pattern[0] == '.':
		// fmt.println("Matching .")
		return matchWildCard(runes, index)
	case pattern[0] == '+':
		// matching +
		return matchOneOrMoreBacktracking(runes, pattern, index, pat)

	case pattern[0] == '?':
		return matchOneOrNone(runes, pattern, index)

	case len(pattern) > 0 && pattern[0] == '[':
		return matchCharacterClass(runes, pattern, index)

	default:
		// Literal substring match
		// fmt.println("Matching literal string")
		var b, i, r = matchCompleteSubString(runes, pattern, index)
		// fmt.println("Value returned is: ", b, i, r)
		return b, i, r
	}
}

func matchDotPlusBacktracking(runes []rune, index int, pat *list.Element) (bool, int, error) {
	if index >= len(runes) {
		return false, -1, nil
	}

	// .+ must match at least one character
	if index >= len(runes) {
		return false, -1, nil
	}

	// Get the next pattern in the sequence
	nextPat := pat.Next()

	// If this is the last pattern, match all remaining characters
	if nextPat == nil {
		return true, len(runes), nil
	}

	// Try matching from the longest possible match down to the minimum (1 character)
	// This implements greedy matching with backtracking
	for endPos := len(runes); endPos > index; endPos-- {
		// Try to match the rest of the pattern from this position
		if matchPatternsFromPosition(runes, nextPat, endPos) {
			return true, endPos, nil
		}
	}

	// If no match found with remaining patterns, just consume one character
	return true, index + 1, nil
}

func matchWildCard(runes []rune, index int) (bool, int, error) {
	if index >= len(runes) {
		return false, -1, fmt.Errorf("index out of bounds for wildcard match")
	}
	return true, index + 1, nil
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
		matches = IsDigit
	} else if basePattern == "\\w" {
		matches = IsAlphanumeric
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
		matches = IsDigit
	} else if basePattern == "\\w" {
		matches = IsAlphanumeric
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
		// fmt.println("First index no match", matches(runes[index]), string(runes[index]))
		return false, -1, nil
	}

	// Match as many as possible while looking ahead
	i := index
	pat = pat.Next()

	for i < len(runes) && matches(runes[i]) { // this only checks for overlapping queries
		if matchPatternsFromPosition(runes, pat, i) {
			// fmt.println("total match found under overlapping + query")
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

	// fmt.println(index+len(patRunes), len(runes), string(runes), patRunes[len(patRunes)-1] == '$', index+len(patRunes)-1 == len(runes))
	if index+len(patRunes) > len(runes) && !(patRunes[len(patRunes)-1] == '$' && index+len(patRunes)-1 == len(runes)) { // sprcifically checking if the last char is not $
		// fmt.println("Caught")
		return false, -1, nil
	}

	if patRunes[len(patRunes)-1] == '$' {
		if index+len(patRunes)-1 != len(runes) {
			// fmt.println("Length mismatch")
			return false, -1, nil
		}
		for j := 0; j < len(patRunes)-1; j++ {
			// fmt.println("Checking1: ", string(runes[index+j]), string(patRunes[j]))
			if runes[index+j] != patRunes[j] {
				// fmt.println("Caught @:", string(patRunes[j]))
				return false, -1, nil
			}
		}
		return true, index + len(pattern), nil
	}

	for j := 0; j < len(patRunes); j++ {
		// fmt.println("Checking: ", string(runes[index+j]), string(patRunes[j]))
		if runes[index+j] != patRunes[j] {
			return false, -1, nil
		}
	}

	return true, index + len(patRunes), nil
}
