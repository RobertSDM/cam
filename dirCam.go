package cam

import (
	"os"
	"path/filepath"
	"slices"
	"time"
)

type DirCam struct {
	// A os.FileInfo for acessing informations about the file in a go,
	// not needing to load every time the info
	Info os.FileInfo

	// The dir path the cam is monitoring
	Path string

	// Cache of the registered files in the directory.
	Files []string
}

func (d *DirCam) Watch(ctx *CamContext, fn func(info os.FileInfo, file *os.File)) {
	defer ctx.WaitGroup.Done()

	for {

		_, err := os.Stat(d.Path)
		if os.IsNotExist(err) {

			if ctx.OnDirExclude != nil {
				ctx.OnDirExclude(filepath.Base(d.Path))
			}

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
				if ctx.OnFileExclude != nil {
					ctx.OnFileExclude(filepath.Base(file))
				}
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

				finfo, _ := os.Stat(newPath)
				filecam := FileCam{
					Path: newPath,
					Info: finfo,
				}

				ctx.WaitGroup.Add(1)
				go filecam.Watch(ctx, fn)

				if ctx.OnFileCreation != nil {
					ctx.OnFileCreation(finfo)
				}
			}

		}

		time.Sleep(500 * time.Millisecond)
	}
}
