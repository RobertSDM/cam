package cam

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Directorys that obrigatory will be monitored. It will bypass the excluded property.
	//
	// If the length is 0 (default), all the directories starting from the root directory will be watched, obeying the excluded property.
	Included []string

	// Group to add the cam created in goroutines.
	WaitGroup *sync.WaitGroup

	OnFileExclude  func(filename string)
	OnFileCreation func(stat os.FileInfo)

	OnDirExclude  func(dirname string)
	// OnDirCreation func(os.FileInfo)
}

// Create a list of FileCam from a list of paths
func (c *CamContext) FromFiles(files ...string) []*FileCam {
	cams := make([]*FileCam, 0, len(files))

	for _, file := range files {

		finfo, err := os.Stat(file)

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

		cams = append(cams, filecam)

	}

	return cams
}

// Create a dir, reading the dir directory (only the files) and passing it to the DirCam struct
func (c *CamContext) CreateDirCam(dirPath string, fn func(info os.FileInfo, file *os.File)) *DirCam {
	dinfo, err := os.Stat(dirPath)
	if err != nil {
		panic(err)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	paths := make([]string, 0, len(entries))

	var newPath string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		newPath = filepath.Join(dirPath, entry.Name())
		paths = append(paths, newPath)
	}

	for _, file := range paths {
		finfo, _ := os.Stat(file)
		filecam := &FileCam{
			Path: file,
			Info: finfo,
		}
		c.WaitGroup.Add(1)
		go filecam.Watch(c, fn)
	}

	cam := &DirCam{
		Info:  dinfo,
		Path:  dirPath,
		Files: paths,
	}

	return cam
}
