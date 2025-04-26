package cam

import (
	"errors"
	"io"
	"os"
	"path"
)

type Cam interface {
	io.Closer

	// Create a goroutine to start watching a path
	Watch()
}

// type context struct {
// 	// Excluded paths will not be monitored.
// 	// Regex can also be informed.
// 	Excluded []string
// }

type Events struct {
	// Event triggered when a file is created in a watched directory
	FileDelete func(path string)

	// Event triggered when a file is deleted from a watched directory
	FileCreate func(path string)

	// Event triggered when a watched directory is excluded
	FolderDelete func(path string)

	// Event triggered when a directory is created in a watched directory
	FolderCreate func(path string)

	// Event triggered when a file is modified
	FileModify func(path string, file *os.File)
}

func (e *Events) onFileModify(path string, file *os.File) {
	if e.FileModify != nil {
		e.FileModify(path, file)
	}
}

func (e *Events) OnFileCreate(path string) {
	if e.FileCreate != nil {
		e.FileCreate(path)
	}
}

func (e *Events) OnFileDelete(path string) {
	if e.FileDelete != nil {
		e.FileDelete(path)
	}
}

func (e *Events) OnDirCreate(path string) {
	if e.FolderCreate != nil {
		e.FolderCreate(path)
	}
}

func (e *Events) OnDirDelete(path string) {
	if e.FolderDelete != nil {
		e.FolderDelete(path)
	}
}

// Starts the goroutine to watch the file
func NewFileCam(p string, handler func(p string, f *os.File)) (cam Cam, err error) {
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return nil, errors.New("the path provided doesn't exist")
	}

	if info.IsDir() {
		return nil, errors.New("the path provided is not a file path")
	}

	filecam := &FileCam{
		info:    info,
		path:    p,
		handler: handler,
		done:    make(chan bool),
	}

	return filecam, nil
}

func NewFileCams(paths []string, handler func(p string, f *os.File)) {
	for _, p := range paths {
		NewFileCam(p, handler)
	}
}

// Starts the goroutine to monitor the folder
func NewFolderCam(p string, recursion bool, events *Events) (cam Cam, err error) {
	info, err := os.Stat(p)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}

	cache := make([]*Cache, 0)

	if !recursion {
		content = getOnlyFiles(content)
	}

	var pathCam Cam
	for _, c := range content {
		if c.IsDir() {
			folder, err := NewFolderCam(path.Join(p, c.Name()), recursion, events)
			if err != nil {
				return nil, err
			}
			pathCam = folder
			go folder.Watch()
		} else {
			file, err := NewFileCam(path.Join(p, c.Name()), events.onFileModify)
			if err != nil {
				return nil, err
			}
			pathCam = file
			go file.Watch()
		}

		cache = append(cache, &Cache{
			cam:   pathCam,
			path:  path.Join(p, c.Name()),
			isDir: c.IsDir(),
		})
	}

	foldercam := &FolderCam{
		info:      info,
		path:      p,
		cache:     cache,
		recursion: recursion,
		events:    events,
		done:      make(chan bool),
	}

	return foldercam, nil
}
