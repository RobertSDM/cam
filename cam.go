package cam

import (
	"os"
	"sync"
)

type Cam interface {
	// Method to start monitoring the path provided to a cam.
	Watch(ctx *CamContext, fn func(os.FileInfo, *os.File))
}

type CamContext struct {
	// Excluded paths will not be monitored.
	// Regex can also be informed.
	Excluded []string

	// If the length is 0 (default), when calling the NewCams method, all the directories starting from the root directory will be watched, obeying the excluded property.
	Included []string

	// WaitGroup to add the cam created in goroutines.
	WG *sync.WaitGroup

	// Events are functions that will be executed in a specific moment
	Events *CamEvent
}

type CamEvent struct {
	// Event triggered when a file is created in a watched directory
	OnFDelete func(path string)
	// Event triggered when a file is deleted from a watched directory
	OnFCreate func(path string)
	// Event triggered when a watched directory is excluded
	OnDDelete func(path string)
	// Event triggered when a directory is created in a watched directory
	OnDCreate func(path string)
	onFModify func(path string, file *os.File)
}

func (e *CamEvent) onFileModify(path string, file *os.File) {
	if e.onFModify != nil {
		e.onFModify(path, file)
	}
}

func (e *CamEvent) OnFileCreate(path string) {
	if e.OnFCreate != nil {
		e.OnFCreate(path)
	}
}

func (e *CamEvent) OnFileDelete(path string) {
	if e.OnFDelete != nil {
		e.OnFDelete(path)
	}
}

func (e *CamEvent) OnDirCreate(path string) {
	if e.OnDCreate != nil {
		e.OnDCreate(path)
	}
}

func (e *CamEvent) OnDirDelete(path string) {
	if e.OnDDelete != nil {
		e.OnDDelete(path)
	}
}
