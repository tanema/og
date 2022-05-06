package watcher

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	events *fsnotify.Watcher
}

func New(paths ...string) (*Watcher, error) {
	fsntfy, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	watcher := &Watcher{events: fsntfy}
	return watcher, walkAndFind(paths, watcher.events.Add)
}

func walkAndFind(paths []string, add func(string) error) error {
	for _, path := range paths {
		path, err := filepath.Abs(strings.TrimSuffix(path, "/..."))
		if err != nil {
			return err
		}
		if err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if info.IsDir() || strings.HasSuffix(path, ".go") {
				if err := add(path); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
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
