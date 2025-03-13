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
	Cache []string
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

		notValidyCache, notStoredInCache := checkValidity(d.Cache, paths, d.Path)

		for _, m := range notValidyCache {
			if ctx.OnFileExclude != nil {
				ctx.OnFileExclude(filepath.Base(m))
			}
		}
		d.Cache = excludeInvalidyCache(d.Cache, notValidyCache)

		for _, m := range notStoredInCache {
			if m.IsDir() {
				continue
			}

			exists = false

			newPath := filepath.Join(d.Path, m.Name())
			if slices.Contains(d.Cache, newPath) {
				exists = true
			}

			if !exists {
				d.Cache = append(d.Cache, newPath)

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

func excludeInvalidyCache(cache []string, invalid []string) []string {
	// New String Slice
	nss := make([]string, 0)
	invalidMap := map[string]bool{}

	for _, inv := range invalid {
		invalidMap[inv] = true
	}

	for _, c := range cache {
		if !invalidMap[c] {
			nss = append(nss, c)
		}
	}

	return nss
}

func checkValidity(files []string, paths []os.DirEntry, dirPath string) ([]string, []os.DirEntry) {
	fms := make([]string, 0)
	pms := make([]os.DirEntry, 0)

	filesMap := map[string]bool{}
	pathsMap := map[string]bool{}

	for i := range max(len(files), len(paths)) {
		if i < len(files) {
			filesMap[files[i]] = true
		}
		if i < len(paths) {
			pathsMap[paths[i].Name()] = true
		}
	}

	for i := range max(len(files), len(paths)) {
		if i < len(files) {
			if !pathsMap[filepath.Base(files[i])] {
				fms = append(fms, files[i])
			}
		}

		if i < len(paths) {
			if paths[i].IsDir() {
				continue
			}
			if !filesMap[filepath.Join(dirPath, paths[i].Name())] {
				pms = append(pms, paths[i])
			}
		}
	}

	return fms, pms
}
