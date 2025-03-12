package cam

import (
	"log"
	"os"
	"path"
	"time"
)

type FileCam struct {
	// A os.FileInfo for acessing informations about the file in a go,
	// not needing to load every time the info
	Info os.FileInfo

	// The file path the cam is monitoring
	Path string
}

func (f *FileCam) Watch(ctx *CamContext, fn func(info os.FileInfo, file *os.File)) {
	defer ctx.WaitGroup.Done()

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

			log.Printf("\"%s\" was MODIFIED\n", path.Base(f.Path))
		}

		time.Sleep(500 * time.Millisecond)
	}
}
