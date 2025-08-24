package matcher

import (
	"container/list"
	"strings"
)

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

	// Try each alternative
	for _, alt := range alternatives {
		alt = strings.TrimSpace(alt)
		if alt == "" {
			continue
		}

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

	// Optional means it's okay if no alternative matches
	return true, index, nil
}
