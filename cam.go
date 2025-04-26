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

type Events struct {
	// Event triggered when a file is created in a watched directory
	FileDelete func(path string)

	// Event triggered when a file is deleted from a watched directory
	FileCreate func(path string)

	// Event triggered when a watched directory is excluded
	FolderDelete func(path string)

	// Event triggered when a directory is created in a watched directory
	FolderCreate func(path string)

	// This callback will be executed every time a change is made in the
	// watched file
	//
	// The file must be closed by the callback, or it will remain open
	// consuming memory
	FileModify func(path string, file *os.File)
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

// Starts the goroutine to monitor the folder
func NewFolderCam(p string, recursion bool, events *Events, excluded []string) (cam Cam, err error) {
	if events == nil || events.FileModify == nil {
		return nil, errors.New("the FileModify event needs to be defined")
	}

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
		_path := path.Join(p, c.Name())

		if isInRegexSlice(excluded, _path) {
			continue
		}

		if c.IsDir() {
			folder, err := NewFolderCam(_path, recursion, events, excluded)
			if err != nil {
				return nil, err
			}
			pathCam = folder
			go folder.Watch()
		} else {
			file, err := NewFileCam(_path, events.FileModify)
			if err != nil {
				return nil, err
			}
			pathCam = file
			go file.Watch()
		}

		cache = append(cache, &Cache{
			cam:   pathCam,
			path:  _path,
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
		excluded:  excluded,
	}

	return foldercam, nil
}
