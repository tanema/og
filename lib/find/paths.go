package find

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Paths(args []string) (paths, tests []string, err error) {
	for _, arg := range args {
		parts := strings.Split(arg, ":")
		path := parts[0]

		if strings.HasSuffix(path, ".go") {
			dir := filepath.Dir(path)
			if !strings.HasPrefix(dir, "/") && !strings.HasPrefix(dir, "./") {
				dir = "./" + dir
			}
			paths = append(paths, dir)
		} else if info, err := os.Stat(strings.TrimRight(path, "/...")); err == nil && info.IsDir() {
			if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "./") {
				path = "./" + path
			}
			paths = append(paths, path)
		} else if strings.HasPrefix(arg, "Test") {
			tests = append(tests, arg)
		} else {
			paths = append(paths, path)
		}
		if len(parts) == 1 {
			if strings.HasSuffix(path, ".go") {
				if !strings.HasSuffix(path, "_test.go") {
					path = strings.ReplaceAll(path, ".go", "_test.go")
				}
				fileTests, err := findAllTestsInFile(path)
				if err != nil {
					return nil, nil, err
				}
				tests = append(tests, fileTests...)
			}
		} else if lineNum, err := strconv.Atoi(parts[1]); err != nil {
			tests = append(tests, parts[1])
		} else if testName, err := findSingleTestAtLine(path, lineNum); err == nil {
			tests = append(tests, testName)
		} else {
			return nil, nil, err
		}
	}
	if len(paths) == 0 {
		paths = append(paths, "./...")
	}
	return
}
