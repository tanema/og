package find

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var testFuncPattern = regexp.MustCompile(`func (Test.*)\(t \*testing\.T\)`)

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
				fileTests, _, err := testsInFile(path)
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

func testsInFile(filepath string) ([]string, []int, error) {
	if info, err := os.Stat(filepath); err != nil {
		return nil, nil, err
	} else if info.IsDir() {
		return nil, nil, fmt.Errorf("what a weirdly named directory %v", filepath)
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	names := []string{}
	lines := []int{}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for currentLine := 0; scanner.Scan(); currentLine++ {
		lineText := scanner.Text()
		if testFuncPattern.MatchString(lineText) {
			names = append(names, testFuncPattern.FindStringSubmatch(lineText)[1])
			lines = append(lines, currentLine)
		}
	}
	return names, lines, scanner.Err()
}

func findSingleTestAtLine(filepath string, line int) (string, error) {
	names, lines, err := testsInFile(filepath)
	if err != nil {
		return "", err
	}
	i := 0
	for ; i < len(lines) && lines[i] < line; i++ {
	}
	return names[i-1], nil
}
