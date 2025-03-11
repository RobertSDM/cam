package cam

import (
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type FileCam struct {
	// A os.FileInfo for acessing informations about the file in a go,
	// not needing to load every time the info
	Info os.FileInfo

	// The file path the cam is monitoring 
	Path string
}

func (f *FileCam) Watch(wg *sync.WaitGroup) {
	defer wg.Done()

	for {

		stat, err := os.Stat(f.Path)
		if err != nil {
			return
		}

		if stat.ModTime() != f.Info.ModTime() {
			f.Info = stat

			log.Printf("\"%s\" was MODIFIED\n", path.Base(f.Path))
		}

		time.Sleep(300 * time.Millisecond)
	}
}
