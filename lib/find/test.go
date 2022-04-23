package find

import (
	"bufio"
	"errors"
	"log"
	"os"
	"regexp"
)

var testFuncPattern = regexp.MustCompile(`func (Test.*)\(t \*testing\.T\)`)

// Test will find the test block at the given line number. If the number falls
// between tests it will run the last test
func Test(filepath string, line int) (string, error) {
	if info, err := os.Stat(filepath); err != nil {
		return "", err
	} else if info.IsDir() {
		return "", errors.New("filepath is directory and cannot get lines in a directory")
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	currentLine := 0
	currentTest := ""
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for ; scanner.Scan(); currentLine++ {
		lineText := scanner.Text()
		if testFuncPattern.MatchString(lineText) {
			currentTest = testFuncPattern.FindStringSubmatch(lineText)[1]
		}
		if currentLine >= line {
			break
		}
	}
	return currentTest, scanner.Err()
}
