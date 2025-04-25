package cam

import (
	"os"
)

type Cam interface {
	// Create a goroutine to start watching a path
	Watch(ctx *Central)
}

type context struct {
	// Excluded paths will not be monitored.
	// Regex can also be informed.
	Excluded []string

	// If the length is 0 (default), when calling the NewCams method, all the directories starting from the root directory will be watched, obeying the excluded property.
	Included []string
}

type events struct {
	// Event triggered when a file is created in a watched directory
	FileDelete func(path string)

	// Event triggered when a file is deleted from a watched directory
	FieCreate func(path string)

	// Event triggered when a watched directory is excluded
	FolderDelete func(path string)

	// Event triggered when a directory is created in a watched directory
	FolderCreate func(path string)

	// Event triggered when a file is modified
	fileModify func(path string, file *os.File)
}

func (e *events) onFileModify(path string, file *os.File) {
	if e.fileModify != nil {
		e.fileModify(path, file)
	}
}

func (e *events) OnFileCreate(path string) {
	if e.FieCreate != nil {
		e.FieCreate(path)
	}
}

func (e *events) OnFileDelete(path string) {
	if e.FileDelete != nil {
		e.FileDelete(path)
	}
}

func (e *events) OnDirCreate(path string) {
	if e.FolderCreate != nil {
		e.FolderCreate(path)
	}
}

func (e *events) OnDirDelete(path string) {
	if e.FolderDelete != nil {
		e.FolderDelete(path)
	}
}
