package matcher

import (
	"grep-go/internal/parsers"
)

func MatchPattern(line []byte, pattern string) (bool, error) {
	runes := []rune(string(line))
	// pattern = strings.ReplaceAll(pattern, " ", "")
	parser := parsers.NewParser()
	patterns, err := parser.ParsePatterns(pattern)

	if err != nil {
		return false, err
	}

	var first string = patterns.Front().Value.(string)
	var start_index int
	var status bool

	if first[0] == '^' {
		// handle string anchor for string beginning
		status, start_index, err = matchIndividualPattern(runes, first[1:], 0, nil)
		// fmt.println("Valuse received under mp: ", status)
		if err != nil || !status {
			// fmt.println("Error here:", status, err)
			return false, err
		}
		if start_index == 1+len(runes) {
			return true, nil
		}
		patterns.Remove(patterns.Front())
	} else {
		start_index = 0
	}

	// Try matching the entire pattern sequence starting at each position
	for startPos := start_index; startPos <= len(runes); startPos++ {
		// fmt.println()
		// fmt.println("Top level check:", string(line))
		if matchPatternsFromPosition(runes, patterns.Front(), startPos) {
			return true, nil
		}
	}

	return false, nil
}
