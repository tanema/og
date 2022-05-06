package results

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tanema/og/lib/stopwatch"
)

// Action is the states of the tests
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
	// Summary Captuset to status of tests
	Summary struct {
		Pass int `json:"pass"`
		Fail int `json:"fail"`
		Skip int `json:"skip"`
	}
	// Set is the complete test results for all packages
	Set struct {
		TestSummary Summary             `json:"test_summary"`
		PkgSummary  Summary             `json:"pkg_summary"`
		Mod         string              `json:"mod"`
		Root        string              `json:"root"`
		Cached      int                 `json:"cached"`
		Packages    map[string]*Package `json:"packages"`
		TotalTests  int                 `json:"total_tests"`
		State       Action              `json:"state"`
		BuildErrors []*BuildError       `json:"build_errors,omitempty"`
		TimeElapsed time.Duration       `json:"time_elapsed"`
		watch       *stopwatch.Stopwatch
	}
	// BuildError captures a single build error in a package
	BuildError struct {
		Package string            `json:"package"`
		Lines   []*BuildErrorLine `json:"lines"`
	}
	// BuildErrorLine is part of the build error trace
	BuildErrorLine struct {
		Path    string `json:"path"`
		Line    int    `json:"line"`
		Column  int    `json:"column"`
		Have    string `json:"have"`
		Want    string `json:"want"`
		Message string `json:"message"`
	}
	logLine struct {
		Package string
		Test    string
		Action  Action
		Output  string
	}
)

// New creates a new setult set
func New(mod, root string) *Set {
	return &Set{
		State:       Pass,
		Mod:         mod,
		Root:        root,
		watch:       stopwatch.Start(),
		Packages:    map[string]*Package{},
		BuildErrors: []*BuildError{},
	}
}

// Parse will parse a reader line by line, adding the lines to the setult set.
// decor is a callback that can be used for displaying results
func (set *Set) Parse(data []byte) {
	line := &logLine{}
	if err := json.Unmarshal(data, &line); err == nil {
		set.Add(line.Action, line.Package, line.Test, line.Output)
	} else {
		set.parseBuildError(string(data))
	}
}

// Complete will mark the set as finished
func (set *Set) Complete() {
	set.TimeElapsed = set.watch.Stop()
	for _, pkg := range set.Packages {
		for _, test := range pkg.Tests {
			if test.State == Fail {
				for _, fail := range test.Failures {
					fail.format()
				}
			}
		}
	}
}

// Add adds an event line to the setult set
func (set *Set) Add(action Action, pkgName, testName, output string) {
	packageName := strings.ReplaceAll(pkgName, set.Mod+"/", "")
	if packageName == "" {
		packageName = "<root>"
	}
	if _, ok := set.Packages[packageName]; !ok {
		set.Packages[packageName] = newPackage(packageName)
	}
	pkg := set.Packages[packageName]
	if _, ok := pkg.Tests[testName]; testName != "" && !ok {
		set.TotalTests++
		pkg.Tests[testName] = newTest(pkg.Name, testName)
	}
	if testName == "" {
		pkg.result(set, action, output)
	} else {
		pkg.Tests[testName].result(set, pkg, action, output)
	}
}

func (set *Set) parseBuildError(data string) {
	if strings.HasPrefix(data, "# ") {
		pkg := strings.Split(data, " ")[1]
		err := &BuildError{
			Package: strings.TrimPrefix(pkg, set.Mod+"/"),
			Lines:   []*BuildErrorLine{},
		}
		set.BuildErrors = append(set.BuildErrors, err)
	} else if strings.HasPrefix(strings.TrimSpace(data), "have (") {
		err := set.BuildErrors[len(set.BuildErrors)-1]
		line := err.Lines[len(err.Lines)-1]
		line.Have = strings.TrimSpace(strings.TrimPrefix(data, "have "))
	} else if strings.HasPrefix(strings.TrimSpace(data), "want (") {
		err := set.BuildErrors[len(set.BuildErrors)-1]
		line := err.Lines[len(err.Lines)-1]
		line.Want = strings.TrimSpace(strings.TrimPrefix(data, "want "))
	} else if !strings.HasPrefix(data, "FAIL") {
		err := set.BuildErrors[len(set.BuildErrors)-1]
		parts := strings.Split(data, ":")
		lineNum, _ := strconv.Atoi(parts[1])
		colNum, _ := strconv.Atoi(parts[2])
		err.Lines = append(err.Lines, &BuildErrorLine{
			Path:    parts[0],
			Line:    lineNum,
			Column:  colNum,
			Message: parts[3],
		})
	}
}

// Any will return if there are any tests at all
func (set *Set) Any() bool {
	return set.TotalTests > 0
}

// Failures will collect all the test failures across all packages into a single
// collection
func (set *Set) Failures() map[string]map[string][]*Failure {
	failures := map[string]map[string][]*Failure{}
	for _, pkg := range set.Packages {
		for _, tst := range pkg.Tests {
			if tst.State == Fail && len(tst.Failures) > 0 {
				if _, ok := failures[pkg.Name]; !ok {
					failures[pkg.Name] = map[string][]*Failure{}
				}
				failures[pkg.Name][tst.Name] = append(failures[pkg.Name][tst.Name], tst.Failures...)
			}
		}
	}
	return failures
}

// Skips will collect all the skipped test names from all the packages into a
// single collection
func (set *Set) Skips() map[string][]string {
	skips := map[string][]string{}
	for _, pkg := range set.Packages {
		for _, tst := range pkg.Tests {
			if tst.State == Skip {
				skips[pkg.Name] = append(skips[pkg.Name], tst.Name)
			}
		}
	}
	return skips
}

// FilteredPackages returns the packages that have tests.
func (set *Set) FilteredPackages(filterNone bool) map[string]*Package {
	if filterNone {
		return set.Packages
	}
	filtered := map[string]*Package{}
	for name, pkg := range set.Packages {
		if pkg.State != Skip {
			filtered[name] = pkg
		}
	}
	return filtered
}

// RankedTests returns tests that are slower than the thsethold and ranks them.
func (set *Set) RankedTests(thsethold time.Duration) []*Test {
	tests := []*Test{}
	for _, pkg := range set.Packages {
		for _, test := range pkg.Tests {
			if test.TimeElapsed >= thsethold {
				tests = append(tests, test)
			}
		}
	}
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].TimeElapsed > tests[j].TimeElapsed
	})
	return tests
}
