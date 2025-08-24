package parsers

import (
	"container/list"
	"fmt"
	"strings"
)

// Parser holds a cache of already-parsed patterns
type Parser struct {
	cache map[string]*list.List
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{
		cache: make(map[string]*list.List),
	}
}

// ParsePatterns parses a pattern string, using the cache if available
func (p *Parser) ParsePatterns(pattern string) (*list.List, error) {
	if lst, exists := p.cache[pattern]; exists {
		return lst, nil
	}

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
		} else if runes[i] == '(' {
			// Handle grouped patterns
			j := i + 1
			parenCount := 1
			for j < len(runes) && parenCount > 0 {
				switch runes[j] {
				case '(':
					parenCount++
				case ')':
					parenCount--
				}
				j++
			}
			if parenCount == 0 {
				// Found matching closing paren
				groupContent := string(runes[i+1 : j-1]) // content between ( and )

				// Check if followed by + or ?
				if j < len(runes) && runes[j] == '+' {
					// Group with + quantifier - always treat as GRP+ regardless of content
					patterns.PushBack("GRP+:" + groupContent)
					i = j + 1
				} else if j < len(runes) && runes[j] == '?' {
					// Group with ? quantifier - always treat as GRP? regardless of content
					patterns.PushBack("GRP?:" + groupContent)
					i = j + 1
				} else {
					// Group without quantifier
					// Check if this is a simple alternation at the top level
					// Only treat as ALT: if the entire content is just alternatives separated by |
					// and no other pattern elements
					isSimpleAlternation := false
					if strings.Contains(groupContent, "|") {
						// Check if it's ONLY alternation (no spaces, quantifiers, etc. outside the alternatives)
						// For now, let's be more conservative and treat most groups as GRP:
						// since the group can contain complex sub-patterns
						parts := strings.Split(groupContent, "|")
						allSimple := true
						for _, part := range parts {
							part = strings.TrimSpace(part)
							// If any part contains spaces or complex patterns, it's not a simple alternation
							if strings.Contains(part, " ") || strings.Contains(part, "?") || strings.Contains(part, "+") {
								allSimple = false
								break
							}
						}
						isSimpleAlternation = allSimple
					}

					if isSimpleAlternation {
						patterns.PushBack("ALT:" + groupContent)
					} else {
						patterns.PushBack("GRP:" + groupContent)
					}
					i = j
				}
			} else {
				// No closing ), treat ( as literal
				patterns.PushBack("(")
				i++
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
		} else if runes[i] == '.' && i+1 < len(runes) && runes[i+1] == '+' {
			// Handle .+ pattern specifically
			patterns.PushBack(".+")
			i += 2
		} else if runes[i] == '+' {
			// handle one or many +
			var last_val = patterns.Back()
			patterns.Remove(patterns.Back())

			if last_val == nil {
				return nil, fmt.Errorf("error in parsing")
			}

			last_pat := last_val.Value.(string)
			// fmt.println("-->", last_pat)

			var last_char string = string(last_pat[len(last_pat)-1])
			var rest_patt string = last_pat[0 : len(last_pat)-1]

			if len(rest_patt) > 0 {
				patterns.PushBack(rest_patt)
			}

			patterns.PushBack("+" + last_char)
			i++
		} else if runes[i] == '?' {
			// handle one or none ?
			var last_val = patterns.Back()
			patterns.Remove(patterns.Back())

			if last_val == nil {
				return nil, fmt.Errorf("error in parsing")
			}

			last_pat := last_val.Value.(string)
			var last_char string = string(last_pat[len(last_pat)-1])
			var rest_patt string = last_pat[0 : len(last_pat)-1]

			if len(rest_patt) > 0 {
				patterns.PushBack(rest_patt)
			}

			patterns.PushBack("?" + last_char)
			i++
		} else if runes[i] == '.' {
			patterns.PushBack(".")
			i += 1
		} else {
			// Collect normal characters until next special character
			j := i
			for j < len(runes) && runes[j] != '\\' && runes[j] != '[' && runes[j] != '+' && runes[j] != '?' && runes[j] != '.' && runes[j] != '(' {
				j++
			}

			if j-1 >= 0 && j < len(runes) && runes[j] == '+' {
				// "ca+ts" should accept cats and caats
				substr := string(runes[i : j-1])
				if len(substr) > 0 {
					patterns.PushBack(substr)
				}
				pos_str := "+" + string(runes[j-1:j])
				patterns.PushBack(pos_str)
				i = j + 1
				continue

			} else if j-1 >= 0 && j < len(runes) && runes[j] == '?' {
				// "ca?ts"
				substr := string(runes[i : j-1])
				if len(substr) > 0 {
					patterns.PushBack(substr)
				}
				pos_str := "?" + string(runes[j-1:j])
				patterns.PushBack(pos_str)
				i = j + 1
				continue

			} else if j-1 >= 0 && j < len(runes) && runes[j] == '.' {
				// Check if it's followed by +
				if j+1 < len(runes) && runes[j+1] == '+' {
					// "ca.+ts" pattern
					substr := string(runes[i:j])
					if len(substr) > 0 {
						patterns.PushBack(substr)
					}
					patterns.PushBack(".+")
					i = j + 2
					continue
				} else {
					// "ca.ts"
					substr := string(runes[i:j])
					if len(substr) > 0 {
						patterns.PushBack(substr)
					}
					patterns.PushBack(".")
					i = j + 1
					continue
				}
			}
			if j > i {
				substr := string(runes[i:j])
				patterns.PushBack(substr)
			}
			i = j
		}
	}

	p.cache[pattern] = patterns

	return patterns, nil
}
