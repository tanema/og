package results

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tanema/og/lib/stopwatch"
)

type Action string

const (
	Run      Action = "run"
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
func (res *Set) Parse(data []byte) (*Package, *Test) {
	line := &logLine{}
	if err := json.Unmarshal(data, &line); err == nil {
		return res.Add(line.Action, line.Package, line.Test, line.Output)
	}
	res.BuildErrors = append(res.BuildErrors, string(data))
	return nil, nil
}

// Complete will mark the set as finished
func (res *Set) Complete() {
	res.TimeElapsed = res.watch.Stop()
}

// Add adds an event line to the result set
func (res *Set) Add(action Action, pkgName, testName, output string) (*Package, *Test) {
	packageName := strings.ReplaceAll(pkgName, res.Mod, ".")
	if _, ok := res.Packages[packageName]; !ok {
		res.Packages[packageName] = &Package{
			watch:   stopwatch.Start(),
			Name:    packageName,
			State:   Run,
			Results: map[string]*Test{},
		}
	}
	pkg := res.Packages[packageName]
	if _, ok := pkg.Results[testName]; testName != "" && !ok {
		res.TotalTests++
		pkg.Results[testName] = &Test{
			watch:   stopwatch.Start(),
			Name:    testName,
			State:   Run,
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
		test.Messages = append(test.Messages, output)
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

func (test *Test) AddMessage(output string) {
	if strings.Contains(output, "CONT") || strings.Contains(output, "PAUSE") {
		return
	}
	msg := strings.ReplaceAll(output, test.Name, "")
	msg = strings.TrimRightFunc(msg, func(r rune) bool { return r == ' ' || r == '\n' })
	msg = regexp.MustCompile(`[=-]{3}\s(RUN|FAIL|PASS|SKIP):?\s*`).ReplaceAllString(msg, "")
	msg = strings.TrimPrefix(msg, "FAIL")
	msg = regexp.MustCompile(`\(.*\)$`).ReplaceAllString(msg, "")
	if strings.TrimSpace(msg) == "Test:" || strings.TrimSpace(msg) == "" {
		return
	}
	test.Messages = append(test.Messages, msg)
}

// FilteredPackages returns the packages that have tests.
func (res *Set) FilteredPackages(filterNone bool) map[string]*Package {
	if filterNone {
		return res.Packages
	}
	filtered := map[string]*Package{}
	for name, pkg := range res.Packages {
		if pkg.State != Skip {
			filtered[name] = pkg
		}
	}
	return filtered
}

// FilteredTests filters out skipped tests.
func (pkg *Package) FilteredTests(filterSkip bool) map[string]*Test {
	if filterSkip {
		return pkg.Results
	}
	filtered := map[string]*Test{}
	for name, test := range pkg.Results {
		if test.State != Skip {
			filtered[name] = test
		}
	}
	return filtered
}

// RankedTests returns tests that are slower than the threshold and ranks them.
func (res *Set) RankedTests(threshold time.Duration) []*Test {
	tests := []*Test{}
	for _, pkg := range res.Packages {
		for _, test := range pkg.Results {
			if test.TimeElapsed >= threshold {
				tests = append(tests, test)
			}
		}
	}
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].TimeElapsed > tests[j].TimeElapsed
	})
	return tests
}
