package find

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// Mod will find the module for a directory by searching upward for a go.mod file
func Mod(curDir string) (*modfile.Module, error) {
	curDir, _ = filepath.Abs(strings.TrimSuffix(curDir, "/..."))
	fileParts := strings.Split(curDir, string(filepath.Separator))
	for i := len(fileParts) - 1; i >= 1; i-- {
		if fileParts[i] == "" {
			continue
		}
		path := strings.Join(fileParts[:i+1], string(filepath.Separator))
		if path != "" {
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("could not file path %v", path)
			}
			if !info.IsDir() {
				continue
			}
		}
		path = strings.Join(append(fileParts[:i+1], "go.mod"), string(filepath.Separator))
		if _, err := os.Stat(path); err != nil {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("there was a problem opening the go.mod file at %v", path)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("there was a problem reading the go.mod file at %v", path)
		}
		mfile, err := modfile.Parse(path, data, nil)
		if err != nil {
			return nil, fmt.Errorf("there was a problem parsing the go.mod file at %v", path)
		}
		return mfile.Module, nil
	}
	return nil, errors.New("could not find go.mod, are you in a project?")
}
