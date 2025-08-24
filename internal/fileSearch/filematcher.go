package fileSearch

import (
	"bufio"
	"container/list"
	"fmt"
	"grep-go/internal/matcher"
	"os"
)

func FileSearch(file_paths []string, pattern string) (bool, error) {
	found_one := false
	for _, file_path := range file_paths {
		var file *os.File
		file, err := os.Open(file_path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "File io error: %v\n", err)
			return false, nil
		}
		found, matches, single_file_err := SingleFileSearch(file, pattern)
		defer file.Close()
		if found {

			// fmt.Println("Matched with: ", matches.Len())

			for lin := matches.Front(); lin != nil; lin = lin.Next() {
				string_value := lin.Value.(string)
				fmt.Println(file_path + ":" + string_value)
			}
			found_one = true
		}

		if single_file_err != nil {
			fmt.Fprintf(os.Stderr, "Single file search error for: %v\n", single_file_err)
			return false, nil
		}

	}
	return found_one, nil
}

func SingleFileSearch(file *os.File, pattern string) (bool, *list.List, error) {
	scanner := bufio.NewScanner(file)
	list := list.New()

	for scanner.Scan() {
		line_read := scanner.Text()
		byte_string := []byte(line_read)

		found, err := matcher.MatchPattern(byte_string, pattern)

		// fmt.Println(line_read, found)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Error in file search: %v\n", err)
			return false, nil, err
		}

		if found {
			list.PushBack(line_read)
		}
	}

	if list.Len() > 0 {
		return true, list, nil
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: Error in file read: %v\n", err)
	}
	return false, nil, nil
}
