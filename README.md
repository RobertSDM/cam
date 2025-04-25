# **CAM**

## What is **CAM**?

Cam is a go library for file watch recursively. It offer to you events that will be triggered when a file or a directory, is created or deleted.

## Usage

Simple usage

```go
func main() {
	wg := &sync.WaitGroup{}

	central := &cam.Central{
		WG: wg,
	}

	err := central.NewCams([]string{"src"}, true, modhandler)
	if err != nil {
		panic(err)
	}

	wg.Wait()
}
```

## Why CAM?

If you need high performance to watch you files, this is not the way to go. Take a look at [fsnotify](https://github.com/fsnotify/fsnotify). \
This is a personal project and a piece of other project **[LiraLR](https://github.com/RobertSDM/LiraLR)**. \
If you have some consideration on the way the project let me know.
