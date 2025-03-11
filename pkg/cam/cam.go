package cam

import (
	"os"
	"path/filepath"
	"sync"
)

type Cam interface {
	// Method to start monitoring the path provided to a cam.
	Watch(wg *sync.WaitGroup)
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
}

func (c *CamContext) StartMonitoring() {
}

func (c *CamContext) FromDir(dirPath string) Cam {
	dinfo, err := os.Stat(dirPath)
	if err != nil {
		panic(err)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	paths := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		newPath := filepath.Join(dirPath, entry.Name())
		paths = append(paths, newPath)
	}

	cam := &DirCam{
		Info:  dinfo,
		Path:  dirPath,
		Files: paths,
	}

	return cam
}
