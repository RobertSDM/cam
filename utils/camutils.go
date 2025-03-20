package utils

import (
	"os"
	"path"
	"path/filepath"
)

func GetOnlyFiles(paths []os.DirEntry) []os.DirEntry {
	files := make([]os.DirEntry, 0, len(paths))

	for _, path := range paths {
		if path.IsDir() {
			continue
		}

		files = append(files, path)
	}

	return files
}

func DirEntryToStrSlice(entries []os.DirEntry, dirPath string) []string {
	strSlice := make([]string, 0, len(entries))
	for _, e := range entries {
		strSlice = append(strSlice, path.Join(dirPath, e.Name()))
	}
	return strSlice
}

func GetOnlyFilesInStrSlice(entries []os.DirEntry, dirPath string) []string {
	paths := make([]string, 0, len(entries))

	var newPath string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		newPath = path.Join(dirPath, entry.Name())
		paths = append(paths, newPath)
	}
	return paths
}

func IsInRegexSlice(regexSlice []string, target string) bool {
	for _, ex := range regexSlice {
		match, _ := filepath.Match(ex, target)

		if match {
			return true
		}
	}

	return false
}

func VerifyConditions(included []string, excluded []string, _path string) bool {
	if len(included) > 0 {
		isInIncluded := IsInRegexSlice(included, _path)
		if isInIncluded {
			return true
		}
	}

	return !IsInRegexSlice(excluded, _path)
}
