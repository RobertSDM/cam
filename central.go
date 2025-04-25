package cam

import (
	"errors"
	"os"
	"path"
	"sync"

	"github.com/RobertSDM/cam/utils"
)

type Central struct {
	context *context

	// events are functions that will be executed in a specific moment
	events *events

	// WaitGroup to add the cam created in goroutines.
	WG *sync.WaitGroup
}

// Create a new cam wathing a file or a folder
func (c *Central) NewCams(_paths []string, recursion bool, handle func(filepath string, file *os.File)) error {
	if c.WG == nil {
		return errors.New("a WaitGroup must be set")
	}

	if c.events == nil {
		c.events = &events{}
	}

	if c.context == nil {
		c.context = &context{}
	}

	if c.events.fileModify == nil {
		c.events.fileModify = handle
	}

	for _, _path := range _paths {
		stat, err := os.Stat(_path)
		if err != nil {
			return err
		}

		c.context.Included = append(c.context.Included, _path)

		if stat.IsDir() {
			c.newFolderCam(_path, recursion)
		} else {
			c.newFileCam(_path)
		}
	}

	return nil
}

// Starts the goroutine to watch the file
func (c *Central) newFileCam(filepath string) error {
	finfo, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return err
	}

	if finfo.IsDir() {
		return errors.New("the path provided is not a file path")
	}

	allowed := utils.VerifyConditions(c.context.Included, c.context.Excluded, filepath)
	if !allowed {
		return errors.New("the path is not allowed")
	}

	filecam := &FileCam{
		info: finfo,
		path: filepath,
	}

	c.WG.Add(1)
	go filecam.Watch(c)

	return nil
}

// The same as [camWatchFile] but iterates through a slice of file paths
func (c *Central) newFileCams(files []string) {
	for _, file := range files {
		c.newFileCam(file)
	}
}

// Starts the goroutine to monitor the folder
func (c *Central) newFolderCam(dirPath string, recursion bool) error {
	dinfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	allowed := utils.VerifyConditions(c.context.Included, c.context.Excluded, dirPath)
	if !allowed {
		return errors.New("the path is not allowed")
	}

	paths, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	files := make([]string, 0)
	cache := make([]*Cache, 0)

	strCache := utils.DirEntriesToStrSlice(paths, dirPath)

	for _, str := range strCache {
		stat, _ := os.Stat(str)
		cache = append(cache, &Cache{
			path:  str,
			isDir: stat.IsDir(),
		})
	}

	dircam := &FolderCam{
		info:      dinfo,
		path:      dirPath,
		cache:     cache,
		recursion: recursion,
	}

	c.WG.Add(1)
	go dircam.Watch(c)

	for _, _path := range paths {
		fullpath := path.Join(dirPath, _path.Name())
		stat, _ := os.Stat(fullpath)

		if !stat.IsDir() {
			files = append(files, fullpath)
		} else if recursion {
			c.newFolderCam(fullpath, recursion)
		}
	}

	c.newFileCams(files)

	return nil
}
