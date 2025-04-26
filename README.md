# **CAM**

## What is **CAM**?

Cam is a go library for file watch recursively. It offer to you events that will be triggered when a file or a directory, is created or deleted.

## Usage

Simple folder cam usage

```go
func main() {
	folder, err := cam.NewFolderCam("src", false, &cam.Events{
		FileModify: handler,
	})
	if err != nil {
		panic(err)
	}
	defer folder.Close()

	// Create a go routine is not necessary when watching only one file or folder
	folder.Watch()
}

// File cam usage

func main() {
	file, err := cam.NewFileCam("src", handler)
	if err != nil {
		panic(err)
	}

	// Create a go routine is not necessary when watching only one file or folder
	file.Watch()
}
```

Multiple folder usage

```go
func main() {
	folder1, err := cam.NewFolderCam("src", false, &cam.Events{
		FileModify: handler1,
	})
	if err != nil {
		panic(err)
	}
	defer folder1.Close()

	folder2, err := cam.NewFolderCam("logs", true, &cam.Events{
		FileModify: handler2,
	})
	if err != nil {
		panic(err)
	}
	defer folder2.Close()

	go folder1.Watch()
	go folder2.Watch()

	<-make(chan chan bool) // blocks the main goroutine
}

// Multiple file cams

func main() {
	file1, err := cam.NewFileCam("index.html", handler1)
	if err != nil {
		panic(err)
	}

	file2, err := cam.NewFileCam("log.txt", handler2)
	if err != nil {
		panic(err)
	}

	go file1.Watch()
	go file2.Watch()

	<-make(chan chan bool) // blocks the main goroutine
}

// OR

func main() {
	files := []string{"file1", "file2"}

	for _, f := range files {
		file, err := cam.NewFileCam(f, mod)
		if err != nil {
			panic(err)
		}
		go file.Watch()
		defer file.Close()
	}

	<-make(chan bool) // blocks the main goroutine
}
```

The use of `cam.Cam.Close()` in the folderCam are used to make sure the deletition events will run
The fileCam exemples don't it because their close() method are for stopping the execution

## Why CAM?

If you need high performance to watch you files, this is not the way to go. Take a look at [fsnotify](https://github.com/fsnotify/fsnotify). \
This is a personal project and a piece of other project **[LiraLR](https://github.com/RobertSDM/LiraLR)**. \
If you have some consideration on the way the project let me know.
