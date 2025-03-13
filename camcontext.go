package cam

import (
	"errors"
	"fmt"
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
	// Included []string

	// WaitGroup to add the cam created in goroutines.
	WG *sync.WaitGroup

	// Event triggered when a file is created in a watched directory
	OnFileExclude func(filename string)
	// Event triggered when a file is deleted from a watched directory
	OnFileCreation func(stat os.FileInfo)

	// Event triggered When a watched directory is excluded
	OnDirExclude func(dirname string)
	// OnDirCreation func(os.FileInfo)
}

// Create and initializes the monitoring of a slice of files
func (c *CamContext) NewCamsFromFiles(files []string, fn func(info os.FileInfo, file *os.File)) {
	for _, file := range files {
		finfo, err := os.Stat(file)

		var valid bool
		for _, ex := range c.Excluded {
			match, _ := filepath.Match(ex, finfo.Name())

			valid = match
			if match {
				break
			}
		}

		if valid {
			fmt.Println("no no")
			continue
		}

		if os.IsNotExist(err) {
			fmt.Printf("\"%s\" does not exist\n", filepath.Base(file))
			continue
		}

		if finfo.IsDir() {
			fmt.Println("the path provided is a dir path")
			continue
		}

		filecam := &FileCam{
			Info: finfo,
			Path: file,
		}

		c.WG.Add(1)
		go filecam.Watch(c, fn)
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

	if c.OnFileCreation != nil {
		c.OnFileCreation(finfo)
	}

	c.WG.Add(1)
	go filecam.Watch(c, fn)

	return nil
}

// Create initializes the monitoriment of a directory and all his files
func (c *CamContext) NewCamFromDir(dirPath string, fn func(info os.FileInfo, file *os.File)) error {
	dinfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	paths := utils.GetOnlyFilesInSlice(dirPath, entries)
	c.NewCamsFromFiles(paths, fn)

	dircam := &DirCam{
		Info:  dinfo,
		Path:  dirPath,
		Cache: paths,
	}

	c.WG.Add(1)
	go dircam.Watch(c, fn)

	return nil
}
