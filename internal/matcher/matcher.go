package matcher

import (
	"grep-go/internal/parsers"
)

// MatchPattern tries to match a pattern string against a given line of text.
// It supports:
//   - '^' anchor (pattern must match at beginning of line)
//   - Sequential matching of parsed sub-patterns
//
// Params:
//   - line:   The line (as []byte) to match against
//   - pattern The search pattern string
//
// Returns:
//   - bool:  true if the pattern matches, false otherwise
//   - error: if parsing or matching fails
func MatchPattern(line []byte, pattern string) (bool, error) {
	// Convert input to runes (to handle Unicode properly)
	runes := []rune(string(line))

	// Create a parser instance and parse the pattern
	parser := parsers.NewParser()
	patterns, err := parser.ParsePatterns(pattern)
	if err != nil {
		return false, err
	}

	// Extract the first parsed sub-pattern (e.g. "^foo" or "bar")
	first := patterns.Front().Value.(string)

	// Keep track of starting index for matching
	var startIndex int
	var status bool

	// Case 1: Pattern starts with '^' anchor -> must match at start of line
	if first[0] == '^' {
		status, startIndex, err = matchIndividualPattern(runes, first[1:], 0, nil)
		if err != nil || !status {
			// If the first part fails, no match is possible
			return false, err
		}

		// If the entire string matched and no characters remain, it's a match
		if startIndex == 1+len(runes) {
			return true, nil
		}

		// Otherwise, remove the first pattern (already matched) and continue
		patterns.Remove(patterns.Front())
	} else {
		// No '^' anchor, so we try starting from any position
		startIndex = 0
	}

	// Case 2: Try matching the remaining patterns starting at every position
	for pos := startIndex; pos <= len(runes); pos++ {
		if matchPatternsFromPosition(runes, patterns.Front(), pos) {
			return true, nil
		}
	}

	// No match found
	return false, nil
}
