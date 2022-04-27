package find

import (
	"bufio"
	"errors"
	"log"
	"os"
	"regexp"
)

var testFuncPattern = regexp.MustCompile(`func (Test.*)\(t \*testing\.T\)`)

func findSingleTestAtLine(filepath string, line int) (string, error) {
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

func findAllTestsInFile(filepath string) ([]string, error) {
	if info, err := os.Stat(filepath); err != nil {
		return nil, err
	} else if info.IsDir() {
		return nil, errors.New("just pass directory, no need to scrape it")
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tests := []string{}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lineText := scanner.Text()
		if testFuncPattern.MatchString(lineText) {
			tests = append(tests, testFuncPattern.FindStringSubmatch(lineText)[1])
		}
	}
	return tests, scanner.Err()
}
