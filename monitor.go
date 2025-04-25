package cam

import (
	"os"
	"path"
	"time"

	"github.com/RobertSDM/cam/utils"
)

type Cache struct {
	isDir bool

	// The file or folder path from a directory that is being watched
	path string
}

type FolderCam struct {
	// A os.FileInfo for acessing informations about the file in a go,
	// not needing to load every time the info
	info os.FileInfo

	// The dir path the cam is monitoring
	path string

	// cache of the registered files in the directory.
	cache []*Cache

	recursion bool
}

func (d *FolderCam) Watch(c *Central) {
	defer c.WG.Done()

	for {
		_, err := os.Stat(d.path)
		if os.IsNotExist(err) {
			return
		}

		paths, _ := os.ReadDir(d.path)

		if !d.recursion {
			paths = utils.GetOnlyFiles(paths)
		}

		notValidyCache, notStoredInCache := d.checkValidity(paths)

		d.updateCache(notValidyCache, notStoredInCache)
		d.runEvents(c.Events, notValidyCache, notStoredInCache)

		for _, entry := range notStoredInCache {
			stat, _ := os.Stat(entry.path)
			if stat.IsDir() {
				c.newFolderCam(entry.path, d.recursion)
			} else {
				c.newFileCam(entry.path)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (d *FolderCam) runEvents(events *Events, notValidyCache []*Cache, notStoredInCache []*Cache) {
	if events == nil {
		return
	}

	for _, entry := range notValidyCache {
		if entry.isDir {
			events.OnDirDelete(entry.path)
		} else {
			events.OnFileDelete(entry.path)
		}
	}

	for _, entry := range notStoredInCache {
		newPath := entry.path
		stat, _ := os.Stat(newPath)

		if stat.IsDir() {
			events.OnDirCreate(newPath)
		} else {
			events.OnFileCreate(newPath)
		}
	}
}

func (d *FolderCam) updateCache(notValidyCache []*Cache, notStoredInCache []*Cache) {
	d.cache = d.excludeInvalidyCache(notValidyCache)
	d.cache = append(d.cache, notStoredInCache...)
}

func (d *FolderCam) excludeInvalidyCache(invalid []*Cache) []*Cache {
	// New String Slice
	nss := make([]*Cache, 0)
	invalidMap := map[string]bool{}

	for _, inv := range invalid {
		invalidMap[inv.path] = true
	}

	for _, c := range d.cache {
		if !invalidMap[c.path] {
			nss = append(nss, c)
		}
	}

	return nss
}

func (d *FolderCam) checkValidity(paths []os.DirEntry) ([]*Cache, []*Cache) {
	notValidyInCache := make([]*Cache, 0)
	notStoredInCache := make([]*Cache, 0)

	filesMap := map[string]bool{}
	pathsMap := map[string]bool{}

	maxlen := max(len(d.cache), len(paths))

	for i := range maxlen {
		if i < len(d.cache) {
			filesMap[d.cache[i].path] = true
		}
		if i < len(paths) {
			pathsMap[path.Join(d.path, paths[i].Name())] = true
		}
	}

	for i := range maxlen {
		if i < len(d.cache) {
			if !pathsMap[d.cache[i].path] {
				notValidyInCache = append(notValidyInCache, d.cache[i])
			}
		}

		if i < len(paths) {
			_path := path.Join(d.path, paths[i].Name())
			if !filesMap[_path] {
				stat, _ := os.Stat(_path)
				notStoredInCache = append(notStoredInCache, &Cache{
					path:  _path,
					isDir: stat.IsDir(),
				})
			}
		}
	}

	return notValidyInCache, notStoredInCache
}

type FileCam struct {
	// A os.FileInfo for acessing informations about the file in a go,
	// not needing to load every time the info
	info os.FileInfo

	// The file path the cam is monitoring
	path string
}

func (f *FileCam) Watch(c *Central) {
	defer c.WG.Done()

	if f.info.Size() > 0 {
		file, _ := os.Open(f.path)
		c.Events.onFileModify(f.path, file)
		file.Close()
	}

	for {
		stat, err := os.Stat(f.path)
		if err != nil {
			return
		}

		if stat.ModTime() != f.info.ModTime() {
			f.info = stat

			file, _ := os.Open(f.path)
			c.Events.onFileModify(f.path, file)
			file.Close()
		}

		time.Sleep(200 * time.Millisecond)
	}
}
