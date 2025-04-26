package cam

import (
	"os"
	"path"
	"time"
)

type Cache struct {
	cam Cam

	path string

	isDir bool
}

type FolderCam struct {
	// A os.FileInfo for acessing informations about the file in a go,
	// not needing to load every time the info
	info os.FileInfo

	// The dir path the cam is monitoring
	path string

	// Registry of files and folders present in the folder being watched
	//
	// If some of the files or folders change, the cache will be updated
	// with the new content
	cache []*Cache

	// A flag showing if folders inside this folder will be watched
	recursion bool

	// Events
	events *Events

	// A channel that will receive a bool, when the Cam.Close() method is
	// called
	done chan bool
}

func (f *FolderCam) Watch() {
	for {
		select {
		case <-f.done:
			for _, c := range f.cache {
				c.cam.Close()
			}
			return
		case <-time.After(100 * time.Millisecond):

		}

		_, err := os.Stat(f.path)
		if err != nil {
			for _, c := range f.cache {
				c.cam.Close()
			}
			return
		}

		content, err := os.ReadDir(f.path)
		if err != nil {
			for _, c := range f.cache {
				c.cam.Close()
			}
			return
		}

		if !f.recursion {
			content = getOnlyFiles(content)
		}

		notValidyCache, notInCache := f.checkCacheValidity(content)

		f.cache = f.excludeInvalidyCache(notValidyCache)
		f.runEvents(f.events, notValidyCache, notInCache)

		cache := make([]*Cache, 0)
		for _, p := range notInCache {
			info, _ := os.Stat(p)

			if info.IsDir() {
				folder, err := NewFolderCam(p, f.recursion, f.events)
				if err != nil {
					for _, c := range f.cache {
						c.cam.Close()
					}
					return
				}
				cache = append(cache, &Cache{
					cam:   folder,
					path:  p,
					isDir: info.IsDir(),
				})
				go folder.Watch()
			} else {
				file, err := NewFileCam(p, f.events.onFileModify)
				if err != nil {
					for _, c := range f.cache {
						c.cam.Close()
					}
					return
				}
				cache = append(cache, &Cache{
					cam:   file,
					path:  p,
					isDir: info.IsDir(),
				})
				go file.Watch()
			}
		}
		f.cache = append(f.cache, cache...)
	}
}

func (f *FolderCam) Close() error {
	close(f.done)
	return nil
}

func (f *FolderCam) runEvents(events *Events, notValidyCache []*Cache, notStoredInCache []string) {
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

	for _, p := range notStoredInCache {
		stat, _ := os.Stat(p)

		if stat.IsDir() {
			events.OnDirCreate(p)
		} else {
			events.OnFileCreate(p)
		}
	}
}

func (f *FolderCam) excludeInvalidyCache(invalid []*Cache) []*Cache {
	newCache := make([]*Cache, 0)
	invalidMap := map[string]bool{}

	for _, inv := range invalid {
		invalidMap[inv.path] = true
	}

	for _, c := range f.cache {
		if !invalidMap[c.path] {
			newCache = append(newCache, c)
		}
	}

	return newCache
}

func (f *FolderCam) checkCacheValidity(content []os.DirEntry) ([]*Cache, []string) {
	notValidyInCache := make([]*Cache, 0)
	notStoredInCache := make([]string, 0)

	cacheMap := make(map[string]bool)
	contentMap := make(map[string]bool)

	maxlen := max(len(f.cache), len(content))

	for i := range maxlen {
		if i < len(f.cache) {
			cacheMap[f.cache[i].path] = true
		}
		if i < len(content) {
			contentMap[path.Join(f.path, content[i].Name())] = true
		}
	}

	for i := range maxlen {
		if i < len(f.cache) {
			if !contentMap[f.cache[i].path] {
				notValidyInCache = append(notValidyInCache, f.cache[i])
			}
		}

		if i < len(content) {
			_path := path.Join(f.path, content[i].Name())
			if !cacheMap[_path] {
				notStoredInCache = append(notStoredInCache, _path)
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

	// This callback will be executed every time a change is made in the
	// watched file
	handler func(p string, f *os.File)

	// A channel that will receive a bool, when the Cam.Close() method is
	// called
	done chan bool
}

func (f *FileCam) Watch() {
	if f.info.Size() > 0 {
		file, err := os.Open(f.path)
		if err != nil {
			return
		}
		f.handler(f.path, file)
		file.Close()
	}

	for {
		select {
		case <-f.done:
			return
		case <-time.After(100 * time.Millisecond):

		}

		stat, err := os.Stat(f.path)
		if err != nil {
			return
		}

		if stat.ModTime() != f.info.ModTime() && stat.Size() != f.info.Size() {
			f.info = stat

			file, err := os.Open(f.path)
			if err != nil {
				return
			}
			f.handler(f.path, file)
			file.Close()
		}

	}
}

func (f *FileCam) Close() error {
	close(f.done)

	return nil
}
