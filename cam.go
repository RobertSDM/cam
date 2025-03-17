package cam

import (
	"os"
	"sync"
)

type Cam interface {
	// Method to start monitoring the path provided to a cam.
	Watch(ctx *CamContext, fn func(os.FileInfo, *os.File))
}

const (
	E_FileCreate = iota
	E_FileDelete
	E_DirCreate
	E_DirDelete
)

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
}

func (e *CamEvent) ExecEvent(event int, path string) {
	switch event {
	case E_DirCreate:
		e.onDirCreate(path)
	case E_FileCreate:
		e.onFileCreate(path)
	case E_DirDelete:
		e.onDirDelete(path)
	case E_FileDelete:
		e.onFileDelete(path)
	}
}

func (e *CamEvent) onFileCreate(path string) {
	if e.OnFCreate != nil {
		e.OnFCreate(path)
	}
}

func (e *CamEvent) onFileDelete(path string) {
	if e.OnFDelete != nil {
		e.OnFDelete(path)
	}
}

func (e *CamEvent) onDirCreate(path string) {
	if e.OnDCreate != nil {
		e.OnDCreate(path)
	}
}

func (e *CamEvent) onDirDelete(path string) {
	if e.OnDDelete != nil {
		e.OnDDelete(path)
	}
}
