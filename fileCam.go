package cam

import (
	"os"
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
	defer ctx.WG.Done()

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

		time.Sleep(500 * time.Millisecond)
	}
}
