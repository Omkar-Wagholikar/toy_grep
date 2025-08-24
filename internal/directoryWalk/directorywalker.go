package directorywalk

import (
	"fmt"
	"grep-go/internal/fileSearch"
	"io/fs"
	"path/filepath"
)

func DirectorySearch(rootPath string, pattern string) (bool, error) {
	var filePaths []string

	// Collect all file paths
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// fmt.Printf("Error accessing %s: %v\n", path, err)
			return err
		}
		if !d.IsDir() {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		return false, fmt.Errorf("error walking the directory: %w", err)
	}

	// Call FileSearch on all collected files
	// for i, s := range filePaths {
	// 	fmt.Println(i, s)
	// }

	return fileSearch.FileSearch(filePaths, pattern)
}
