package parsers

import (
	"container/list"
	"fmt"
	"os"
)

func ParsePatterns(pattern string) (*list.List, error) {
	patterns := list.New()
	runes := []rune(pattern)
	i := 0
	count := 0
	for i < len(runes) {
		count += 1
		if count == 20 {
			fmt.Println("=>", i)

			fmt.Println("== Patterns == ")
			for ele := patterns.Front(); ele != nil; ele = ele.Next() {
				fmt.Println("-\t", ele.Value.(string), " ")
			}
			fmt.Println()
			fmt.Println("== done ==")

			os.Exit(3)
		}

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
			fmt.Println("-->", last_pat)

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
			for j < len(runes) && runes[j] != '\\' && runes[j] != '[' && runes[j] != '+' && runes[j] != '?' && runes[j] != '.' {
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
	fmt.Println("Parse complete")
	return patterns, nil
}
