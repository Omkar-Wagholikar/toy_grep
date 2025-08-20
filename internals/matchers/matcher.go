package matchers

import (
	"container/list"
	"fmt"
	"grep-go/internals/parsers"
	"strings"
)

// c && echo -n "I see 1 cat, 2 dogs and 3 cows" | ./toy_grep.sh -E "^I see (\d (cat|dog|cow)s?(, | and )?)+$"
// c && echo -n "1 cat, 2 dogs and 3 cows" | ./toy_grep.sh -E "^(\d (cat|dog|cow)s?(, | and )?)+"
// c && echo -n ",,," | ./toy_grep.sh -E ".?(, | and )*"
// c && echo -n "cow" | ./toy_grep.sh -E "cow|dog"

func MatchPattern(line []byte, pattern string) (bool, error) {
	runes := []rune(string(line))
	// pattern = strings.ReplaceAll(pattern, " ", "")
	patterns, err := parsers.ParsePatterns(pattern)
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
			fmt.Println("Error here:", status, err)
			return false, err
		}

		patterns.Remove(patterns.Front())
	} else {
		start_index = 0
	}

	// Try matching the entire pattern sequence starting at each position
	for startPos := start_index; startPos <= len(runes); startPos++ {
		fmt.Println()
		fmt.Println("Top level check:", string(line))
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

		if len(pat_string) > 0 && pat_string[len(pat_string)-1] == '$' {
			pat_string = pat_string[0 : len(pat_string)-1]
			end_detect = true
		}

		fmt.Println("Matching individual pattern:", pat.Value, "at", startPos)

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

		found, nextPos, err := matchIndividualPattern(runes, pat_string, currentPos, pat)

		if err != nil {
			fmt.Println(err)
			return false
		}

		currentPos = nextPos

		if end_detect {
			// handle string anchor for string end
			input_text_length := len(runes)
			if found && input_text_length == currentPos {
				return true
			} else {
				fmt.Printf("End anchor failed: currentPos=%d, inputLength=%d\n", currentPos, input_text_length)
				return false
			}
		}

		if !found {
			fmt.Println("fail at:", pat.Value)
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
		return matchSingleCharacter(runes, isAlphanumeric, index)

	case pattern == "\\d":
		return matchSingleCharacter(runes, isDigit, index)

	case pattern == "\\\\d":
		// literal substring "\d"
		return matchCompleteSubString(runes, `\d`, index)

	case pattern == "\\\\w":
		// literal substring "\w"
		return matchCompleteSubString(runes, `\w`, index)

	case pattern == ".+":
		// Handle .+ pattern (one or more of any character)
		return matchDotPlusBacktracking(runes, pattern, index, pat)

	case len(pattern) > 4 && pattern[:4] == "ALT:":
		// Handle simple alternation pattern
		fmt.Println("Matching ALT:")
		return matchAlternation(runes, pattern, index)

	case len(pattern) > 5 && pattern[:4] == "ALT+":
		// Handle alternation with + quantifier
		fmt.Println("Matching ALT+")
		return matchAlternationPlus(runes, pattern, index, pat)

	case len(pattern) > 5 && pattern[:4] == "ALT?":
		// Handle alternation with ? quantifier
		fmt.Println("Matching ALT?")
		return matchAlternationOptional(runes, pattern, index)

	case len(pattern) > 4 && pattern[:4] == "GRP:":
		// Handle simple group
		fmt.Println("Matching GRP:")
		return matchGroup(runes, pattern, index)

	case len(pattern) > 5 && pattern[:4] == "GRP+":
		// Handle group with + quantifier
		fmt.Println("Matching GRP+")
		return matchGroupPlus(runes, pattern, index, pat)

	case len(pattern) > 5 && pattern[:4] == "GRP?":
		// Handle group with ? quantifier
		fmt.Println("Matching GRP?")
		return matchGroupOptional(runes, pattern, index)

	case pattern[0] == '.':
		fmt.Println("Matching .")
		return matchWildCard(runes, pattern, index)
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
func matchDotPlusBacktracking(runes []rune, pattern string, index int, pat *list.Element) (bool, int, error) {
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

func matchWildCard(runes []rune, pattern string, index int) (bool, int, error) {
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

func matchAlternation(runes []rune, pattern string, index int) (bool, int, error) {
	// Remove "ALT:" prefix

	alternativesStr := pattern[4:]
	alternatives := strings.Split(alternativesStr, "|")

	// Try each alternative
	for _, alt := range alternatives {
		alt = strings.TrimSpace(alt)
		if alt == "" {
			continue
		}

		// Try to match this alternative as a complete substring
		altRunes := []rune(alt)
		if index+len(altRunes) <= len(runes) {
			match := true
			for j := 0; j < len(altRunes); j++ {
				if runes[index+j] != altRunes[j] {
					match = false
					break
				}
			}
			if match {
				return true, index + len(altRunes), nil
			}
		}
	}

	return false, -1, nil
}

func matchAlternationPlus(runes []rune, pattern string, index int, pat *list.Element) (bool, int, error) {
	if index >= len(runes) {
		return false, -1, nil
	}

	// Remove "ALT+:" prefix
	alternativesStr := pattern[5:]
	alternatives := strings.Split(alternativesStr, "|")

	// Get the next pattern in the sequence
	nextPat := pat.Next()

	// Must match at least once
	matched := false
	currentPos := index

	// Keep trying to match alternatives until no more matches
	for {
		foundMatch := false
		bestPos := currentPos

		// Try each alternative at current position
		for _, alt := range alternatives {
			alt = strings.TrimSpace(alt)
			if alt == "" {
				continue
			}

			altRunes := []rune(alt)
			if currentPos+len(altRunes) <= len(runes) {
				match := true
				for j := 0; j < len(altRunes); j++ {
					if runes[currentPos+j] != altRunes[j] {
						match = false
						break
					}
				}
				if match {
					bestPos = currentPos + len(altRunes)
					foundMatch = true
					matched = true
					break // Use first matching alternative
				}
			}
		}

		if !foundMatch {
			break
		}

		// If we have a next pattern, try to match it from various positions
		if nextPat != nil {
			// Try to match remaining pattern from this position
			for testPos := bestPos; testPos <= len(runes); testPos++ {
				if matchPatternsFromPosition(runes, nextPat, testPos) {
					return true, testPos, nil
				}
			}
		}

		currentPos = bestPos
	}

	if !matched {
		return false, -1, nil
	}

	// If no next pattern, we're done
	return true, currentPos, nil
}

func matchAlternationOptional(runes []rune, pattern string, index int) (bool, int, error) {
	// Remove "ALT?:" prefix
	alternativesStr := pattern[5:]
	alternatives := strings.Split(alternativesStr, "|")

	fmt.Printf("ALT? matching at index %d: trying alternatives %v\n", index, alternatives)

	// Try each alternative
	for _, alt := range alternatives {
		alt = strings.TrimSpace(alt)
		if alt == "" {
			continue
		}

		fmt.Printf("  Trying alternative: '%s' at position %d\n", alt, index)

		altRunes := []rune(alt)
		if index+len(altRunes) <= len(runes) {
			match := true
			for j := 0; j < len(altRunes); j++ {
				if runes[index+j] != altRunes[j] {
					match = false
					break
				}
			}
			if match {
				fmt.Printf("  Alternative '%s' matched, advancing to %d\n", alt, index+len(altRunes))
				return true, index + len(altRunes), nil
			}
		}
		fmt.Printf("  Alternative '%s' did not match\n", alt)
	}

	// Optional means it's okay if no alternative matches
	fmt.Printf("  No alternatives matched, but optional so returning success at same position %d\n", index)
	return true, index, nil
}

func matchGroup(runes []rune, pattern string, index int) (bool, int, error) {
	// Remove "GRP:" prefix and parse the group content
	groupContent := pattern[4:]
	groupPatterns, err := parsers.ParsePatterns(groupContent)
	if err != nil {
		return false, -1, err
	}

	// Match all patterns in the group sequentially
	currentPos := index
	for groupPat := groupPatterns.Front(); groupPat != nil; groupPat = groupPat.Next() {
		patStr := groupPat.Value.(string)
		found, nextPos, err := matchIndividualPattern(runes, patStr, currentPos, groupPat)
		if err != nil || !found {
			return false, -1, err
		}
		currentPos = nextPos
	}

	return true, currentPos, nil
}

func matchGroupPlus(runes []rune, pattern string, index int, pat *list.Element) (bool, int, error) {
	if index >= len(runes) {
		return false, -1, nil
	}

	// Remove "GRP+:" prefix
	groupContent := pattern[5:]

	// Get the next pattern in the sequence
	nextPat := pat.Next()

	// Must match at least once
	matched := false
	currentPos := index

	// Keep trying to match the group until no more matches
	for {
		groupPatterns, err := parsers.ParsePatterns(groupContent)
		if err != nil {
			return false, -1, err
		}

		// Try to match all patterns in the group
		tempPos := currentPos
		groupMatched := true

		for groupPat := groupPatterns.Front(); groupPat != nil; groupPat = groupPat.Next() {
			patStr := groupPat.Value.(string)
			found, nextPos, err := matchIndividualPattern(runes, patStr, tempPos, groupPat)
			if err != nil || !found {
				groupMatched = false
				break
			}
			tempPos = nextPos
		}

		if !groupMatched {
			break
		}

		matched = true

		// If we have a next pattern, try to match it from various positions
		if nextPat != nil {
			// Try to match remaining pattern from this position onward
			for testPos := tempPos; testPos <= len(runes); testPos++ {
				if matchPatternsFromPosition(runes, nextPat, testPos) {
					return true, testPos, nil
				}
			}
		}

		currentPos = tempPos
	}

	if !matched {
		return false, -1, nil
	}

	// If no next pattern, we're done
	return true, currentPos, nil
}

func matchGroupOptional(runes []rune, pattern string, index int) (bool, int, error) {
	// Remove "GRP?:" prefix
	groupContent := pattern[5:]

	fmt.Printf("GRP? matching at index %d with content: '%s'\n", index, groupContent)

	// Parse the group content into patterns
	groupPatterns, err := parsers.ParsePatterns(groupContent)
	if err != nil {
		fmt.Printf("GRP? parsing failed: %v, treating as no match (optional)\n", err)
		return true, index, nil
	}

	// Try to match all patterns in the group sequentially
	currentPos := index
	for groupPat := groupPatterns.Front(); groupPat != nil; groupPat = groupPat.Next() {
		patStr := groupPat.Value.(string)
		fmt.Printf("  GRP? trying to match pattern: '%s' at position %d\n", patStr, currentPos)

		found, nextPos, err := matchIndividualPattern(runes, patStr, currentPos, groupPat)
		if err != nil || !found {
			// Optional group doesn't match, that's okay
			fmt.Printf("  GRP? pattern '%s' failed, optional so returning success at original position %d\n", patStr, index)
			return true, index, nil
		}
		fmt.Printf("  GRP? pattern '%s' matched, advancing from %d to %d\n", patStr, currentPos, nextPos)
		currentPos = nextPos
	}

	fmt.Printf("  GRP? all patterns matched, final position: %d\n", currentPos)
	return true, currentPos, nil
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}
