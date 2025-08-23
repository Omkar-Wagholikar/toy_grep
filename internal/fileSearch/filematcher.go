package fileSearch

import (
	"bufio"
	"container/list"
	"fmt"
	"grep-go/internal/matchers"
	"os"
)

func FileSearch(file *os.File, pattern string) (bool, error) {
	scanner := bufio.NewScanner(file)
	list := list.New()

	for scanner.Scan() {
		line_read := scanner.Text()
		byte_string := []byte(line_read)
		found, err := matchers.MatchPattern(byte_string, pattern)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Error in file search: %v\n", err)
			return false, err
		}

		if found {
			list.PushBack(line_read)
		}
	}

	if list.Len() > 0 {
		for lin := list.Front(); lin != nil; lin = lin.Next() {
			fmt.Println(lin.Value.(string))
		}
		return true, nil
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: Error in file read: %v\n", err)
	}
	return false, nil
}
