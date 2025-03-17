package cam

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/RobertSDM/cam/utils"
)

type Central struct {
	Context *CamContext
}

// Create cams from the included paths
func (c *Central) StartMonitoring(recusion bool, fn func(info os.FileInfo, file *os.File)) error {
	if len(c.Context.Included) == 0 {
		root, _ := os.Getwd()
		c.Context.Included = append(c.Context.Included, root)
	}

	for _, inc := range c.Context.Included {
		info, err := os.Stat(inc)
		if os.IsNotExist(err) {
			return err
		}

		if info.IsDir() {
			c.CamWatchDir(inc, fn, recusion)
		} else {
			c.CamWatchFile(inc, fn)
		}
	}

	return nil
}

func (c *Central) CamWatchFile(file string, fn func(info os.FileInfo, file *os.File)) error {
	finfo, err := os.Stat(file)

	if os.IsNotExist(err) {
		return err
	}

	if finfo.IsDir() {
		return errors.New("the path provided is not a file path")
	}

	isIn := utils.IsInRegexSlice(c.Context.Excluded, finfo.Name())
	if isIn {
		return errors.New("the path is in the excluded paths")
	}

	filecam := &FileCam{
		Info: finfo,
		Path: file,
	}

	c.Context.WG.Add(1)
	go filecam.Watch(c.Context, fn)

	return nil
}

// Create and initializes the monitoring of a files
func (c *Central) CamsWatchFiles(files []string, fn func(info os.FileInfo, file *os.File)) {
	for _, file := range files {
		c.CamWatchFile(file, fn)
	}
}

// Create initializes the monitoriment of a directory and all his files
func (c *Central) CamWatchDir(dirPath string, fn func(info os.FileInfo, file *os.File), recursion bool) error {
	dinfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	isInExcluded := utils.IsInRegexSlice(c.Context.Excluded, dinfo.Name())
	if isInExcluded {
		return errors.New("the file is not valid")
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
	go dircam.Watch(c, fn, recursion)

	for _, path := range paths {
		fullpath := filepath.Join(dirPath, path.Name())
		stat, _ := os.Stat(fullpath)

		if !stat.IsDir() {
			files = append(files, fullpath)
		} else if recursion {
			c.CamWatchDir(fullpath, fn, recursion)
		}
	}

	c.CamsWatchFiles(files, fn)

	return nil
}
