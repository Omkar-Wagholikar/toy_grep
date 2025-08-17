package matchers

import (
	"container/list"
	"grep-go/internals/parsers"
)

func MatchPattern(line []byte, pattern string) (bool, error) {
	runes := []rune(string(line))

	patterns, err := parsers.ParsePatterns(pattern)
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
