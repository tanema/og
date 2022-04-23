package results

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/tanema/og/lib/stopwatch"
)

type (
	// Summary Captures to status of tests
	Summary struct {
		Pass int
		Fail int
		Skip int
	}
	// Failure is a single test failure message
	Failure struct {
		Name     string
		Package  string
		Messages []string
	}
	// Set is the complete test results for all packages
	Set struct {
		TestSummary Summary
		PkgSummary  Summary
		Mod         string
		Cached      int
		Names       []string
		Packages    map[string]*Package
		TotalTests  int
		State       string
		Failures    map[string][]Failure
		Skips       map[string][]string
		BuildErrors []string
		TimeElapsed time.Duration
		watch       *stopwatch.Stopwatch
	}
	// Package is the test results for a single package
	Package struct {
		Summary
		Name        string
		Names       []string
		Results     map[string]*Test
		State       string
		Cached      bool
		TimeElapsed time.Duration
		watch       *stopwatch.Stopwatch
	}
	// Test is the results for a single test
	Test struct {
		Name        string
		State       string
		Package     string
		Messages    []string
		TimeElapsed time.Duration
		watch       *stopwatch.Stopwatch
	}
	logLine struct {
		Package string
		Test    string
		Action  string
		Output  string
	}
)

// New creates a new result set
func New(mod string) *Set {
	return &Set{
		Mod:         mod,
		watch:       stopwatch.Start(),
		Packages:    map[string]*Package{},
		Failures:    map[string][]Failure{},
		Skips:       map[string][]string{},
		BuildErrors: []string{},
	}
}

// Parse will parse a reader line by line, adding the lines to the result set.
// decor is a callback that can be used for displaying results
func (res *Set) Parse(r io.Reader, decor func(*Set, *Package, *Test)) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := &logLine{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			res.consumeBuildError(scanner)
			if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
				continue
			}
		}
		pkg, test := res.Add(line.Package, line.Test, line.Action, line.Output)
		decor(res, pkg, test)
	}
	res.TimeElapsed = res.watch.Stop()
}

func (res *Set) consumeBuildError(scanner *bufio.Scanner) {
	if strings.HasPrefix(scanner.Text(), "FAIL") {
		scanner.Scan() // skip build fail final line
		return
	}
	for scanner.Scan() {
		out := scanner.Text()
		if strings.HasPrefix(out, "{") {
			break
		}
		res.BuildErrors = append(res.BuildErrors, out)
	}
}

// Add adds an event line to the result set
func (res *Set) Add(pkgName, testName, action, output string) (*Package, *Test) {
	packageName := strings.ReplaceAll(pkgName, res.Mod, ".")
	if _, ok := res.Packages[packageName]; !ok {
		res.Names = append(res.Names, packageName)
		res.Packages[packageName] = &Package{
			watch:   stopwatch.Start(),
			Name:    packageName,
			State:   "run",
			Results: map[string]*Test{},
		}
	}
	pkg := res.Packages[packageName]
	if _, ok := pkg.Results[testName]; testName != "" && !ok {
		res.TotalTests++
		pkg.Names = append(pkg.Names, testName)
		pkg.Results[testName] = &Test{
			watch:   stopwatch.Start(),
			Name:    testName,
			State:   "run",
			Package: pkg.Name,
		}
	}
	if testName == "" {
		return res.packageResult(pkg, action, output), nil
	}
	return nil, res.testResult(pkg, pkg.Results[testName], action, output)
}

func (res *Set) packageResult(pkg *Package, action, output string) *Package {
	switch action {
	case "pass":
		res.PkgSummary.Pass++
	case "fail":
		res.PkgSummary.Fail++
	case "skip":
		res.PkgSummary.Skip++
	case "pause":
		pkg.TimeElapsed = pkg.watch.Pause()
	case "cont":
		pkg.watch.Resume()
	case "output":
		if strings.HasPrefix(output, "ok") && strings.Contains(output, "(cached)") {
			res.Cached++
			pkg.Cached = true
		}
	}
	if action != "output" {
		pkg.State = action
	}
	if action == "pass" || action == "fail" || action == "skip" {
		pkg.TimeElapsed = pkg.watch.Stop()
		return pkg
	}
	return nil
}

func (res *Set) testResult(pkg *Package, test *Test, action, output string) *Test {
	switch action {
	case "pass":
		pkg.Pass++
		res.TestSummary.Pass++
	case "fail":
		pkg.Fail++
		res.TestSummary.Fail++
		if len(test.Messages) > 0 {
			res.Failures[pkg.Name] = append(res.Failures[pkg.Name], Failure{
				Name:     test.Name,
				Package:  pkg.Name,
				Messages: test.Messages,
			})
		}
	case "skip":
		pkg.Skip++
		res.TestSummary.Skip++
		res.Skips[pkg.Name] = append(res.Skips[pkg.Name], test.Name)
	case "pause":
		test.TimeElapsed = test.watch.Pause()
	case "cont":
		test.watch.Resume()
	case "output":
		if !strings.Contains(output, "CONT") && !strings.Contains(output, "PAUSE") {
			if msg := cleanMsg(test.Name, output); msg != "" {
				test.Messages = append(test.Messages, msg)
			}
		}
	}
	if action != "output" {
		test.State = action
	}
	if action == "pass" || action == "fail" || action == "skip" {
		test.TimeElapsed = test.watch.Stop()
		return test
	}
	return nil
}

func cleanMsg(testName, output string) string {
	msg := strings.ReplaceAll(output, testName, "")
	msg = strings.TrimSpace(msg)
	msg = regexp.MustCompile(`^[=-]{3}\s(RUN|FAIL|PASS|SKIP):?\s*`).ReplaceAllString(msg, "")
	msg = regexp.MustCompile(`\(.*\)$`).ReplaceAllString(msg, "")
	if msg == "Test:" {
		return ""
	}
	return strings.TrimSpace(msg)
}

func (res *Set) String() string {
	var sb strings.Builder
	for name, pkg := range res.Packages {
		sb.WriteString(fmt.Sprintf("Package: %v\n", name))
		for name, test := range pkg.Results {
			sb.WriteString(fmt.Sprintf("\t%v: %v\n", name, test.State))
		}
	}
	return sb.String()
}
