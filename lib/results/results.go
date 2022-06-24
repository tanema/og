package results

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

var coverLinePat = regexp.MustCompile(`^(.+):([0-9]+)\.([0-9]+),([0-9]+)\.([0-9]+) ([0-9]+) ([0-9]+)$`)

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
		Pass int `json:"pass,omitempty"`
		Fail int `json:"fail,omitempty"`
		Skip int `json:"skip,omitempty"`
	}
	// Set is the complete test results for all packages
	Set struct {
		*stopwatch
		Name            string
		TestSummary     Summary             `json:"test_summary"`
		PkgSummary      Summary             `json:"pkg_summary"`
		Cached          int                 `json:"cached"`
		Packages        map[string]*Package `json:"packages"`
		TotalTests      int                 `json:"total_tests"`
		State           Action              `json:"state"`
		BuildErrors     []*BuildError       `json:"build_errors,omitempty"`
		StatementCount  int64               `json:"statements,omitempty"`
		CoveredCount    int64               `json:"covered,omitempty"`
		CoveragePercent float64             `json:"percent,omitempty"`
		FailedTests     []*Test             `json:"failed_tests,omitempty"`
		SkippedTests    []*Test             `json:"skipped_tests,omitempty"`
		SlowTests       []*Test             `json:"slow_tests,omitempty"`
		threshold       time.Duration
		path            string
	}
	// BuildError captures a single build error in a package
	BuildError struct {
		Package string `json:"package,omitempty"`
		Path    string `json:"path,omitempty"`
		Line    int64  `json:"line,omitempty"`
		Column  int64  `json:"column,omitempty"`
		Have    string `json:"have,omitempty"`
		Want    string `json:"want,omitempty"`
		Message string `json:"message"`
		Raw     string `json:"raw"`
	}
	logLine struct {
		Package string
		Test    string
		Action  Action
		Output  string
	}
)

// New creates a new setult set
func New(path string, threshold time.Duration) *Set {
	name := path
	pkgs, err := packages.Load(&packages.Config{}, ".")
	if err == nil && len(pkgs) > 0 {
		name = strings.ReplaceAll(pkgs[0].PkgPath, "/"+pkgs[0].Name, "")
	}
	set := &Set{
		Name:        name,
		stopwatch:   &stopwatch{},
		path:        path,
		State:       Run,
		Packages:    map[string]*Package{},
		BuildErrors: []*BuildError{},
		threshold:   threshold,
	}
	set.start()
	return set
}

// Parse will parse a reader line by line, adding the lines to the setult set.
// decor is a callback that can be used for displaying results
func (set *Set) Parse(data []byte) {
	line := &logLine{}
	if err := json.Unmarshal(data, &line); err == nil {
		set.Add(line.Action, line.Package, line.Test, line.Output)
	}
}

// Complete will mark the set as finished
func (set *Set) Complete(shouldCover bool, coverProfile string) {
	defer set.stop()
	if set.State != Fail {
		set.State = Pass
	}
	for _, pkg := range set.Packages {
		for _, test := range pkg.Tests {
			if test.State == Continue || test.State == Pause || test.State == Run {
				test.State = Fail
			}
			if test.State == Fail {
				for _, fail := range test.Failures {
					fail.format()
				}
			}
		}
	}
	sort.Slice(set.SlowTests, func(i, j int) bool {
		return set.SlowTests[i].Elapsed() > set.SlowTests[j].Elapsed()
	})
	if shouldCover {
		set.parseCoverProfile(coverProfile)
	}
}

// Add adds an event line to the setult set
func (set *Set) Add(action Action, pkgName, testName, output string) {
	if _, ok := set.Packages[pkgName]; !ok {
		set.Packages[pkgName] = newPackage(pkgName)
	}
	pkg := set.Packages[pkgName]
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

// ParseError will try its best to parse an error message for formatting
// to the output, adding excerpts and diffs where possible
func (set *Set) ParseError(bdata []byte) {
	data := string(bdata)
	if strings.HasPrefix(strings.TrimSpace(data), "have (") {
		builderr := set.BuildErrors[len(set.BuildErrors)-1]
		builderr.Have = strings.TrimSpace(strings.TrimPrefix(data, "have "))
		return
	} else if strings.HasPrefix(strings.TrimSpace(data), "want (") {
		builderr := set.BuildErrors[len(set.BuildErrors)-1]
		builderr.Want = strings.TrimSpace(strings.TrimPrefix(data, "want "))
		return
	}
	builderr := &BuildError{Raw: data}
	data = strings.TrimLeft(data, "# ")
	parts := strings.Split(data, ":")
	switch len(parts) {
	case 1:
		builderr.Message = parts[0]
	case 2:
		builderr.Path = parts[0]
		builderr.Message = parts[1]
	case 3:
		builderr.Path = parts[0]
		builderr.Line = atoi(parts[1])
		builderr.Message = parts[2]
	default:
		builderr.Path = parts[0]
		builderr.Line = atoi(parts[1])
		builderr.Message = parts[2]
		builderr.Column = atoi(parts[2])
		builderr.Message = strings.Join(parts[3:], " ")
	}
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, filepath.Dir(builderr.Path))
	if err == nil && len(pkgs) > 0 {
		builderr.Package = pkgs[0].PkgPath
	}
	set.BuildErrors = append(set.BuildErrors, builderr)
}

func (set *Set) parseCoverProfile(coverPath string) {
	rd, err := os.Open(coverPath)
	if err != nil {
		return // just bail out if there is no file
	}
	defer rd.Close()
	s := bufio.NewScanner(rd)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "mode: ") {
			continue
		}
		matches := coverLinePat.FindStringSubmatch(line)
		pkg := set.Packages[filepath.Dir(matches[1])]
		stmts := atoi(matches[6])
		pkg.StatementCount += stmts
		set.StatementCount += stmts
		if atoi(matches[7]) > 0 {
			pkg.CoveredCount += stmts
			set.CoveredCount += stmts
		}
		pkg.CoveragePercent = calcPercent(pkg.StatementCount, pkg.CoveredCount)
	}
	set.CoveragePercent = calcPercent(set.StatementCount, set.CoveredCount)
}

func calcPercent(statements, covered int64) float64 {
	if statements > 0 {
		return math.Floor((float64(covered)/float64(statements))*10000) / 100
	}
	return 0
}
