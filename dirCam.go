package cam

import (
	"os"
	"path/filepath"
	"time"

	"github.com/RobertSDM/cam/utils"
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

func (d *DirCam) Watch(ctx *CamContext, fn func(info os.FileInfo, file *os.File), recursion bool) {
	defer ctx.WG.Done()

	for {
		_, err := os.Stat(d.Path)
		if os.IsNotExist(err) {
			return
		}

		paths, _ := os.ReadDir(d.Path)
		if !recursion {
			paths = utils.GetOnlyFiles(paths)
		}

		notValidyCache, notStoredInCache := checkValidity(d.Cache, paths, d.Path)

		for _, m := range notValidyCache {
			if ctx.OnFileExclude != nil {
				ctx.OnFileExclude(filepath.Base(m))
			}
		}
		d.Cache = excludeInvalidyCache(d.Cache, notValidyCache)

		for _, m := range notStoredInCache {
			newPath := filepath.Join(d.Path, m.Name())

			stat, _ := os.Stat(newPath)

			if !stat.IsDir() {
				err := ctx.NewCamFromFile(newPath, fn)
				if err != nil {
					continue
				}
				if ctx.OnFileCreation != nil {
					ctx.OnFileCreation(stat)
				}
			} else {
				err := ctx.NewCamFromDir(newPath, fn, recursion)
				if err != nil {
					continue
				}
				if ctx.OnDirCreation != nil {
					ctx.OnDirCreation(d.Path)
				}
			}

			d.Cache = append(d.Cache, newPath)
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

func checkValidity(cache []string, paths []os.DirEntry, dirPath string) ([]string, []os.DirEntry) {
	fms := make([]string, 0)
	pms := make([]os.DirEntry, 0)

	filesMap := map[string]bool{}
	pathsMap := map[string]bool{}

	for i := range max(len(cache), len(paths)) {
		if i < len(cache) {
			filesMap[cache[i]] = true
		}
		if i < len(paths) {
			pathsMap[paths[i].Name()] = true
		}
	}

	for i := range max(len(cache), len(paths)) {
		if i < len(cache) {
			if !pathsMap[filepath.Base(cache[i])] {
				fms = append(fms, cache[i])
			}
		}

		if i < len(paths) {
			if !filesMap[filepath.Join(dirPath, paths[i].Name())] {
				pms = append(pms, paths[i])
			}
		}
	}

	return fms, pms
}
