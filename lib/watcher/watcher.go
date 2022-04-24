package watcher

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	path   string
	events *fsnotify.Watcher
}

func New(path string) (*Watcher, error) {
	path, _ = filepath.Abs(strings.TrimSuffix(path, "/..."))
	fsntfy, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watcher := &Watcher{
		path:   path,
		events: fsntfy,
	}

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasSuffix(path, ".go") {
			if err := watcher.events.Add(path); err != nil {
				return err
			}
		}
		return nil
	})

	return watcher, err
}

func (watcher *Watcher) Stop() {
	defer watcher.events.Close()
}

func (watcher *Watcher) Watch(fn func(string)) error {
	for {
		select {
		case event, ok := <-watcher.events.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write && strings.HasSuffix(event.Name, ".go") {
				fn(event.Name)
			}
		case err := <-watcher.events.Errors:
			return err
		}
	}
}
