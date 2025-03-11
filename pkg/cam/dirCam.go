package cam

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type DirCam struct {
	Info os.FileInfo

	Path string

	Files []string
}

func (d *DirCam) Watch(wg *sync.WaitGroup) {
	defer wg.Done()

	for _, file := range d.Files {
		finfo, _ := os.Stat(file)
		filecam := FileCam{
			Path: file,
			Info: finfo,
		}
		wg.Add(1)
		go filecam.Watch(wg)
	}

	for {

		_, err := os.Stat(d.Path)
		if os.IsNotExist(err) {
			fmt.Printf("\"%s\" was DELETED", filepath.Base(d.Path))
			return
		}

		paths, _ := os.ReadDir(d.Path)
		var exists bool

		for i, file := range d.Files {
			exists = false

			var newPath string
			for _, path := range paths {

				if path.IsDir() {
					continue
				}

				newPath = filepath.Join(d.Path, path.Name())

				if newPath == file {
					exists = true
					break
				}

			}

			if !exists {
				fmt.Printf("\"%s\" was DELETED\n", filepath.Base(file))
				d.Files = slices.Delete(d.Files, i, i+1)
			}

		}

		for _, path := range paths {
			if path.IsDir() {
				continue
			}

			exists = false

			newPath := filepath.Join(d.Path, path.Name())
			if slices.Contains(d.Files, newPath) {
				exists = true
			}

			if !exists {
				d.Files = append(d.Files, newPath)
				fmt.Printf("\"%s\" was CREATED\n", path.Name())

				finfo, _ := os.Stat(newPath)
				filecam := FileCam{
					Path: newPath,
					Info: finfo,
				}

				wg.Add(1)
				go filecam.Watch(wg)
			}

		}

		time.Sleep(500 * time.Millisecond)
	}
}
