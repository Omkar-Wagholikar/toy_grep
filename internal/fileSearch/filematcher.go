package fileSearch

import (
	"bufio"
	"container/list"
	"fmt"
	"grep-go/internal/matcher"
	"os"
)

// FileSearch iterates over multiple files and searches for a given pattern.
// It prints all matching lines in the format "<file>:<line>".
//
// Params:
//   - filePaths: list of file paths to search
//   - pattern:   search pattern
//
// Returns:
//   - bool:  true if at least one match was found
//   - error: any error encountered while searching
func FileSearch(filePaths []string, pattern string) (bool, error) {
	foundOne := false

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "File I/O error: %v\n", err)
			// keep going instead of stopping on a single bad file
			continue
		}

		// Close each file after processing
		func() {
			defer file.Close()

			found, matches, singleFileErr := SingleFileSearch(file, pattern)
			if singleFileErr != nil {
				fmt.Fprintf(os.Stderr, "Single file search error for %s: %v\n", filePath, singleFileErr)
				return
			}

			if found {
				for lin := matches.Front(); lin != nil; lin = lin.Next() {
					lineText := lin.Value.(string)
					fmt.Printf("%s:%s\n", filePath, lineText)
				}
				foundOne = true
			}
		}()
	}

	return foundOne, nil
}

// SingleFileSearch scans a single file line-by-line and checks each line
// against the given pattern.
//
// Returns:
//   - bool:      true if at least one match found
//   - *list.List: linked list of matched lines
//   - error:     error if parsing/matching fails
func SingleFileSearch(file *os.File, pattern string) (bool, *list.List, error) {
	scanner := bufio.NewScanner(file)
	matches := list.New()

	for scanner.Scan() {
		line := scanner.Text()
		found, err := matcher.MatchPattern([]byte(line), pattern)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in file search: %v\n", err)
			return false, nil, err
		}

		if found {
			matches.PushBack(line)
		}
	}

	// Handle scanner error (I/O or bufio issue)
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		return false, nil, err
	}

	if matches.Len() > 0 {
		return true, matches, nil
	}
	return false, nil, nil
}
