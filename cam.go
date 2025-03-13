package cam

import (
	"os"
)

type Cam interface {
	// Method to start monitoring the path provided to a cam.
	Watch(ctx *CamContext, fn func(os.FileInfo, *os.File))
}
