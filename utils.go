package cam

import (
	"os"
	"path/filepath"
)

func getOnlyFiles(paths []os.DirEntry) []os.DirEntry {
	files := make([]os.DirEntry, 0, len(paths))

	for _, path := range paths {
		if path.IsDir() {
			continue
		}

		files = append(files, path)
	}

	return files
}

func isInRegexSlice(regexSlice []string, target string) bool {
	for _, ex := range regexSlice {
		match, _ := filepath.Match(ex, target)

		if match {
			return true
		}
	}

	return false
}
