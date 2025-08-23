package fileSearch

import (
	"bufio"
	"fmt"
	"grep-go/internal/matchers"
	"os"
)

func FileSearch(file *os.File, pattern string) (bool, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line_read := scanner.Text()
		byte_string := []byte(line_read)
		found, err := matchers.MatchPattern(byte_string, pattern)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Error in file search: %v\n", err)
			return false, err
		}

		if found {
			fmt.Println(line_read)
			return true, err
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: Error in file read: %v\n", err)
	}
	return false, nil
}
