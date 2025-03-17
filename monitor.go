package cam

import (
	"os"
	"path/filepath"
	"time"

	"github.com/RobertSDM/cam/utils"
)

type Cache struct {
	Path string

	IsDir bool
}

type DirCam struct {
	// A os.FileInfo for acessing informations about the file in a go,
	// not needing to load every time the info
	Info os.FileInfo

	// The dir path the cam is monitoring
	Path string

	// Cache of the registered files in the directory.
	Cache []*Cache
}

func (d *DirCam) Watch(c *Central, fn func(info os.FileInfo, file *os.File), recursion bool) {
	defer c.Context.WG.Done()

	for {
		_, err := os.Stat(d.Path)
		if os.IsNotExist(err) {
			return
		}

		paths, _ := os.ReadDir(d.Path)

		if !recursion {
			paths = utils.GetOnlyFiles(paths)
		}

		notValidyCache, notStoredInCache := d.checkValidity(paths)

		d.updateCache(notValidyCache, notStoredInCache)
		d.runEvents(c.Context.Events, notValidyCache, notStoredInCache)

		for _, entry := range notStoredInCache {
			stat, _ := os.Stat(entry.Path)
			if stat.IsDir() {
				c.CamWatchDir(entry.Path, fn, recursion)
			} else {
				c.CamWatchFile(entry.Path, fn)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (d *DirCam) runEvents(events *CamEvent, notValidyCache []*Cache, notStoredInCache []*Cache) {
	if events == nil {
		return
	}

	for _, entry := range notValidyCache {
		if entry.IsDir {
			events.ExecEvent(E_DirDelete, entry.Path)
		} else {
			events.ExecEvent(E_FileDelete, entry.Path)
		}
	}

	for _, entry := range notStoredInCache {
		newPath := entry.Path
		stat, _ := os.Stat(newPath)

		if stat.IsDir() {
			events.ExecEvent(E_DirCreate, newPath)
		} else {
			events.ExecEvent(E_FileCreate, newPath)
		}
	}
}

func (d *DirCam) updateCache(notValidyCache []*Cache, notStoredInCache []*Cache) {
	d.Cache = d.excludeInvalidyCache(notValidyCache)

	d.Cache = append(d.Cache, notStoredInCache...)
}

func (d *DirCam) excludeInvalidyCache(invalid []*Cache) []*Cache {
	// New String Slice
	nss := make([]*Cache, 0)
	invalidMap := map[string]bool{}

	for _, inv := range invalid {
		invalidMap[inv.Path] = true
	}

	for _, c := range d.Cache {
		if !invalidMap[c.Path] {
			nss = append(nss, c)
		}
	}

	return nss
}

func (d *DirCam) checkValidity(paths []os.DirEntry) ([]*Cache, []*Cache) {
	notValidyInCache := make([]*Cache, 0)
	notStoredInCache := make([]*Cache, 0)

	filesMap := map[string]bool{}
	pathsMap := map[string]bool{}

	maxlen := max(len(d.Cache), len(paths))

	for i := range maxlen {
		if i < len(d.Cache) {
			filesMap[d.Cache[i].Path] = true
		}
		if i < len(paths) {
			pathsMap[filepath.Join(d.Path, paths[i].Name())] = true
		}
	}

	for i := range maxlen {
		if i < len(d.Cache) {
			if !pathsMap[d.Cache[i].Path] {
				notValidyInCache = append(notValidyInCache, d.Cache[i])
			}
		}

		if i < len(paths) {
			path := filepath.Join(d.Path, paths[i].Name())
			if !filesMap[path] {
				stat, _ := os.Stat(path)
				notStoredInCache = append(notStoredInCache, &Cache{
					Path:  path,
					IsDir: stat.IsDir(),
				})
			}
		}
	}

	return notValidyInCache, notStoredInCache
}

type FileCam struct {
	// A os.FileInfo for acessing informations about the file in a go,
	// not needing to load every time the info
	Info os.FileInfo

	// The file path the cam is monitoring
	Path string
}

func (f *FileCam) Watch(ctx *CamContext, fn func(info os.FileInfo, file *os.File)) {
	defer ctx.WG.Done()

	if f.Info.Size() > 0 {
		file, _ := os.Open(f.Path)
		fn(f.Info, file)
		file.Close()
	}

	for {
		stat, err := os.Stat(f.Path)
		if err != nil {
			return
		}

		if stat.ModTime() != f.Info.ModTime() {
			f.Info = stat

			file, _ := os.Open(f.Path)
			fn(f.Info, file)
			file.Close()
		}

		time.Sleep(200 * time.Millisecond)
	}
}
