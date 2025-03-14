package utils

import (
	"os"
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

func DirEntryToStrSlice(entries []os.DirEntry) []string {
	strSlice := make([]string, 0, len(entries))
	for _, e := range entries{
		strSlice = append(strSlice, e.Name())
	}
	return strSlice
}

func GetOnlyFilesInStrSlice(dirPath string, entries []os.DirEntry) []string {
	paths := make([]string, 0, len(entries))

	var newPath string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		newPath = filepath.Join(dirPath, entry.Name())
		paths = append(paths, newPath)
	}
	return paths
}
