package results

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tanema/og/lib/stopwatch"
)

type Action string

const (
	Pass     Action = "pass"
	Fail     Action = "fail"
	Skip     Action = "skip"
	Continue Action = "cont"
	Pause    Action = "pause"
	Output   Action = "output"
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
		State       Action
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
		State       Action
		Cached      bool
		TimeElapsed time.Duration
		watch       *stopwatch.Stopwatch
	}
	// Test is the results for a single test
	Test struct {
		Name        string
		State       Action
		Package     string
		Messages    []string
		TimeElapsed time.Duration
		watch       *stopwatch.Stopwatch
	}
	logLine struct {
		Package string
		Test    string
		Action  Action
		Output  string
	}
)

// New creates a new result set
func New(mod string) *Set {
	return &Set{
		State:       Pass,
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
		pkg, test := res.Add(line.Action, line.Package, line.Test, line.Output)
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
func (res *Set) Add(action Action, pkgName, testName, output string) (*Package, *Test) {
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
		return res.packageResult(action, pkg, output), nil
	}
	return nil, res.testResult(action, pkg, pkg.Results[testName], output)
}

func (res *Set) packageResult(action Action, pkg *Package, output string) *Package {
	switch action {
	case Pass:
		res.PkgSummary.Pass++
	case Fail:
		res.State = Fail
		res.PkgSummary.Fail++
	case Skip:
		res.PkgSummary.Skip++
	case Pause:
		pkg.TimeElapsed = pkg.watch.Pause()
	case Continue:
		pkg.watch.Resume()
	case Output:
		if strings.HasPrefix(output, "ok") && strings.Contains(output, "(cached)") {
			res.Cached++
			pkg.Cached = true
		}
	}
	if action != Output {
		pkg.State = action
	}
	if action == Pass || action == Fail || action == Skip {
		pkg.TimeElapsed = pkg.watch.Stop()
		return pkg
	}
	return nil
}

func (res *Set) testResult(action Action, pkg *Package, test *Test, output string) *Test {
	switch action {
	case Pass:
		pkg.Pass++
		res.TestSummary.Pass++
	case Fail:
		pkg.Fail++
		res.TestSummary.Fail++
		if len(test.Messages) > 0 {
			res.Failures[pkg.Name] = append(res.Failures[pkg.Name], Failure{
				Name:     test.Name,
				Package:  pkg.Name,
				Messages: test.Messages,
			})
		}
	case Skip:
		pkg.Skip++
		res.TestSummary.Skip++
		res.Skips[pkg.Name] = append(res.Skips[pkg.Name], test.Name)
	case Pause:
		test.TimeElapsed = test.watch.Pause()
	case Continue:
		test.watch.Resume()
	case Output:
		if !strings.Contains(output, "CONT") && !strings.Contains(output, "PAUSE") {
			if msg := cleanMsg(test.Name, output); msg != "" {
				test.Messages = append(test.Messages, msg)
			}
		}
	}
	if action != Output {
		test.State = action
	}
	if action == Pass || action == Fail || action == Skip {
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

func (res *Set) RankedTests() []*Test {
	tests := []*Test{}
	for _, pkg := range res.Packages {
		for _, test := range pkg.Results {
			tests = append(tests, test)
		}
	}
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].TimeElapsed > tests[j].TimeElapsed
	})
	return tests[:10]
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
