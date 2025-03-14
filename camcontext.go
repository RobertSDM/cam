package cam

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/RobertSDM/cam/utils"
)

type CamContext struct {
	// Excluded paths will not be monitored.
	// Regex can also be informed.
	Excluded []string

	// Directorys that obrigatory will be monitored. It will bypass the excluded property.
	//
	// If the length is 0 (default), all the directories starting from the root directory will be watched, obeying the excluded property.
	Included []string

	// WaitGroup to add the cam created in goroutines.
	WG *sync.WaitGroup

	// Event triggered when a file is created in a watched directory
	OnFileExclude func(filename string)
	// Event triggered when a file is deleted from a watched directory
	OnFileCreation func(stat os.FileInfo)

	// Event triggered when a watched directory is excluded
	OnDirExclude  func(dirname string)
	// Event triggered when a directory is created in a watched directory
	OnDirCreation func(dirname string)
}

// Create cams from the included paths
func (c *CamContext) NewCams(recusion bool, fn func(info os.FileInfo, file *os.File)) error {
	for _, inc := range c.Included {
		info, err := os.Stat(inc)
		if os.IsNotExist(err) {
			return err
		}

		if info.IsDir() {
			c.NewCamFromDir(inc, fn, recusion)
		} else {
			c.NewCamFromFile(inc, fn)
		}
	}

	return nil
}

// Create and initializes the monitoring of a slice of files
func (c *CamContext) NewCamsFromFiles(files []string, fn func(info os.FileInfo, file *os.File)) {
	for _, file := range files {
		c.NewCamFromFile(file, fn)
	}
}

// Create and initializes the monitoring of a files
func (c *CamContext) NewCamFromFile(file string, fn func(info os.FileInfo, file *os.File)) error {
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

	return nil
}

// Create initializes the monitoriment of a directory and all his files
func (c *CamContext) NewCamFromDir(dirPath string, fn func(info os.FileInfo, file *os.File), recursion bool) error {
	dinfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	var paths []string
	paths = utils.GetOnlyFilesInStrSlice(dirPath, entries)
	c.NewCamsFromFiles(paths, fn)

	if recursion {
		paths = utils.DirEntryToStrSlice(entries)
	}

	dircam := &DirCam{
		Info:  dinfo,
		Path:  dirPath,
		Cache: paths,
	}

	c.WG.Add(1)
	go dircam.Watch(c, fn, recursion)

	return nil
}
