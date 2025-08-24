package matcher

import (
	"container/list"
	"grep-go/internal/parsers"
)

func matchGroup(runes []rune, pattern string, index int) (bool, int, error) {
	// Remove "GRP:" prefix and parse the group content
	groupContent := pattern[4:]
	parser := parsers.NewParser()
	groupPatterns, err := parser.ParsePatterns(groupContent)

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

	// Collect all possible match positions for backtracking
	var matchPositions []int
	currentPos := index

	// Keep trying to match the group and collect all valid positions
	for {
		parser := parsers.NewParser()
		groupPatterns, err := parser.ParsePatterns(groupContent)

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

		// Record this as a valid match position
		matchPositions = append(matchPositions, tempPos)
		currentPos = tempPos
	}

	// Must match at least once
	if len(matchPositions) == 0 {
		return false, -1, nil
	}

	// If no next pattern, return the last (greedy) match
	if nextPat == nil {
		return true, matchPositions[len(matchPositions)-1], nil
	}

	// Try backtracking: start from the longest match and work backwards
	for i := len(matchPositions) - 1; i >= 0; i-- {
		pos := matchPositions[i]

		if matchPatternsFromPosition(runes, nextPat, pos) {
			return true, pos, nil
		}
	}

	// No backtracking position worked
	return false, -1, nil
}

func matchGroupOptional(runes []rune, pattern string, index int) (bool, int, error) {
	// Remove "GRP?:" prefix
	groupContent := pattern[5:]

	// Parse the group content into patterns
	parser := parsers.NewParser()
	groupPatterns, err := parser.ParsePatterns(groupContent)

	if err != nil {
		return true, index, nil
	}

	// Try to match all patterns in the group sequentially
	currentPos := index
	for groupPat := groupPatterns.Front(); groupPat != nil; groupPat = groupPat.Next() {
		patStr := groupPat.Value.(string)

		found, nextPos, err := matchIndividualPattern(runes, patStr, currentPos, groupPat)
		if err != nil || !found {
			// Optional group doesn't match, that's okay
			return true, index, nil
		}
		currentPos = nextPos
	}

	return true, currentPos, nil
}
