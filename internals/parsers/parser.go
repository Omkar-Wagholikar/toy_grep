package parsers

import "container/list"

func ParsePatterns(pattern string) (*list.List, error) {
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
