package cam

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/RobertSDM/cam/utils"
)

type CamEvent struct {
	// Event triggered when a file is created in a watched directory
	OnFExclude func(filename string)
	// Event triggered when a file is deleted from a watched directory
	OnFCreation func(stat os.FileInfo)

	// Event triggered when a watched directory is excluded
	OnDExclude func(dirname string)
	// Event triggered when a directory is created in a watched directory
	OnDCreation func(dirname string)
}

type CamContext struct {
	// Excluded paths will not be monitored.
	// Regex can also be informed.
	Excluded []string

	// If the length is 0 (default), when calling the NewCams method, all the directories starting from the root directory will be watched, obeying the excluded property.
	Included []string

	// WaitGroup to add the cam created in goroutines.
	WG *sync.WaitGroup

	Events *CamEvent
}

// Create cams from the included paths
func (c *CamContext) Monitor(recusion bool, fn func(info os.FileInfo, file *os.File)) error {
	if len(c.Included) == 0 {
		root, _ := os.Getwd()
		c.Included = append(c.Included, root)
	}

	for _, inc := range c.Included {
		info, err := os.Stat(inc)
		if os.IsNotExist(err) {
			return err
		}

		if info.IsDir() {
			c.NewCamFromDir(inc, fn, recusion)
		} else {
			c.NewCamsFromFiles([]string{inc}, fn)
		}
	}

	return nil
}

// Create and initializes the monitoring of a files
func (c *CamContext) NewCamsFromFiles(files []string, fn func(info os.FileInfo, file *os.File)) error {
	for _, file := range files {
		finfo, err := os.Stat(file)

		if os.IsNotExist(err) {
			return err
		}
		if finfo.IsDir() {
			return errors.New("the path provided is not a file path")
		}

		for _, ex := range c.Excluded {
			match, _ := filepath.Match(ex, finfo.Name())

			if match {
				return errors.New("the file is not valid")
			}
		}

		filecam := &FileCam{
			Info: finfo,
			Path: file,
		}

		c.WG.Add(1)
		go filecam.Watch(c, fn)
	}

	return nil
}

// Create initializes the monitoriment of a directory and all his files
func (c *CamContext) NewCamFromDir(dirPath string, fn func(info os.FileInfo, file *os.File), recursion bool) error {
	dinfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	for _, ex := range c.Excluded {
		match, _ := filepath.Match(ex, dinfo.Name())

		if match {
			return errors.New("the file is not valid")
		}
	}

	paths, _ := os.ReadDir(dirPath)
	files := []string{}

	dircam := &DirCam{
		Info:  dinfo,
		Path:  dirPath,
		Cache: utils.DirEntryToStrSlice(paths, dirPath),
	}
	c.WG.Add(1)
	go dircam.Watch(c, fn, recursion)

	for _, path := range paths {
		fullpath := filepath.Join(dirPath, path.Name())
		stat, _ := os.Stat(fullpath)
		if !stat.IsDir() {
			files = append(files, fullpath)
		} else if recursion {
			c.NewCamFromDir(fullpath, fn, recursion)
		}
	}
	c.NewCamsFromFiles(files, fn)

	return nil
}
