package cam

import (
	"errors"
	"os"
	"path"

	"github.com/RobertSDM/cam/utils"
)

type Central struct {
	Context *CamContext
}

func (c *Central) NewCam(_path string, recursion bool, handle func(filepath string, file *os.File)) error {
	stat, err := os.Stat(_path)
	if err != nil {
		return err
	}

	if c.Context.Events == nil {
		c.Context.Events = &CamEvent{}
	}

	c.Context.Events.onFModify = handle
	c.Context.Included = append(c.Context.Included, _path)

	if stat.IsDir() {
		c.camWatchDir(_path, recursion)
	} else {
		c.camWatchFile(_path)
	}

	return nil
}

func (c *Central) camWatchFile(file string) error {
	finfo, err := os.Stat(file)

	if os.IsNotExist(err) {
		return err
	}

	if finfo.IsDir() {
		return errors.New("the path provided is not a file path")
	}

	allowed := utils.VerifyConditions(c.Context.Included, c.Context.Excluded, file)
	if !allowed {
		return errors.New("the path is not allowed")
	}

	filecam := &FileCam{
		Info: finfo,
		Path: file,
	}

	c.Context.WG.Add(1)
	go filecam.Watch(c.Context)

	return nil
}

// Create and initializes the monitoring of a files
func (c *Central) camsWatchFiles(files []string) {
	for _, file := range files {
		c.camWatchFile(file)
	}
}

// Create initializes the monitoriment of a directory and all his files
func (c *Central) camWatchDir(dirPath string, recursion bool) error {
	dinfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	allowed := utils.VerifyConditions(c.Context.Included, c.Context.Excluded, dirPath)
	if !allowed {
		return errors.New("the path is not allowed")
	}

	paths, _ := os.ReadDir(dirPath)
	files := []string{}

	var cache []*Cache
	strCache := utils.DirEntryToStrSlice(paths, dirPath)

	for _, str := range strCache {
		stat, _ := os.Stat(str)
		cache = append(cache, &Cache{
			Path:  str,
			IsDir: stat.IsDir(),
		})
	}

	dircam := &DirCam{
		Info:  dinfo,
		Path:  dirPath,
		Cache: cache,
	}

	c.Context.WG.Add(1)
	go dircam.Watch(c, recursion)

	for _, _path := range paths {
		fullpath := path.Join(dirPath, _path.Name())
		stat, _ := os.Stat(fullpath)

		if !stat.IsDir() {
			files = append(files, fullpath)
		} else if recursion {
			c.camWatchDir(fullpath, recursion)
		}
	}

	c.camsWatchFiles(files)

	return nil
}
