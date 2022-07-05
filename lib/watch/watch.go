package watch

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fs        *fsnotify.Watcher
	Changes   chan string
	Errors    chan error
	signals   chan os.Signal
	checksums map[string]string
}

func New() (*Watcher, error) {
	watcher := &Watcher{
		Changes:   make(chan string, 1),
		Errors:    make(chan error, 1),
		signals:   make(chan os.Signal, 1),
		checksums: map[string]string{},
	}
	var err error
	go watcher.watchSignals()
	watcher.fs, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return watcher, filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() || strings.HasSuffix(path, ".go") {
			if err := watcher.fs.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}

func (watcher *Watcher) Start() error {
	defer close(watcher.Changes)
	defer close(watcher.Errors)
	for {
		select {
		case event, ok := <-watcher.fs.Events:
			if !ok {
				return nil
			}
			watcher.onEvent(event)
		case err := <-watcher.fs.Errors:
			watcher.Errors <- err
		}
	}
}

func (watcher *Watcher) onEvent(event fsnotify.Event) {
	if event.Op&fsnotify.Write != fsnotify.Write || !strings.HasSuffix(event.Name, ".go") {
		return
	}
	path := event.Name
	if strings.HasSuffix(path, ".go") {
		checksum, err := sum(path)
		if err != nil {
			watcher.Errors <- err
		} else if checksum == watcher.checksums[path] {
			return
		} else {
			watcher.checksums[path] = checksum
		}
	}
	if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
		path = strings.ReplaceAll(path, ".go", "_test.go")
	}
	if _, err := os.Stat(path); err != nil {
		path = filepath.Dir(path)
	}
	watcher.Changes <- path
}

func (watcher *Watcher) watchSignals() {
	signal.Notify(watcher.signals, syscall.SIGINFO)
	go func() {
		for {
			<-watcher.signals
			watcher.Changes <- "./..."
		}
	}()
}

func sum(src string) (string, error) {
	sum := md5.New()
	info, err := os.Stat(src)
	if err != nil {
		return "", err
	} else if info.IsDir() {
		return "", fmt.Errorf("sum: file is a directory")
	}
	s, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer s.Close()
	_, err = io.Copy(sum, s)
	return fmt.Sprintf("%x", sum.Sum(nil)), err
}
