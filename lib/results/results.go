package results

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/excerpt"
	"github.com/tanema/og/lib/stopwatch"
	"golang.org/x/tools/cover"
	"golang.org/x/tools/go/packages"
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
		TestSummary     Summary             `json:"test_summary"`
		PkgSummary      Summary             `json:"pkg_summary"`
		Mod             string              `json:"mod"`
		Root            string              `json:"root"`
		Cached          int                 `json:"cached"`
		Packages        map[string]*Package `json:"packages"`
		TotalTests      int                 `json:"total_tests"`
		State           Action              `json:"state"`
		BuildErrors     []*BuildError       `json:"build_errors,omitempty"`
		TimeElapsed     time.Duration       `json:"time_elapsed"`
		StatementCount  int64               `json:"statements"`
		CoveredCount    int64               `json:"covered"`
		CoveragePercent float64             `json:"percent"`
		path            string
		cfg             *config.Config
		watch           *stopwatch.Stopwatch
	}
	// BuildError captures a single build error in a package
	BuildError struct {
		Package string            `json:"package"`
		Lines   []*BuildErrorLine `json:"lines"`
	}
	// BuildErrorLine is part of the build error trace
	BuildErrorLine struct {
		Path    string   `json:"path"`
		Line    int      `json:"line"`
		Column  int      `json:"column"`
		Have    string   `json:"have"`
		Want    string   `json:"want"`
		Message string   `json:"message"`
		Excerpt []string `json:"excerpt"`
	}
	logLine struct {
		Package string
		Test    string
		Action  Action
		Output  string
	}
)

// New creates a new setult set
func New(cfg *config.Config, path string) *Set {
	return &Set{
		path:        path,
		cfg:         cfg,
		State:       Pass,
		Mod:         cfg.ModName,
		Root:        cfg.Root,
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
func (set *Set) Complete() error {
	set.TimeElapsed = set.watch.Stop()
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
	if set.cfg.Cover != "" {
		if err := set.parseCoverProfile(set.cfg.Cover, set.path); err != nil {
			return err
		}
	}
	return nil
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

func (set *Set) parseBuildError(data string) {
	if strings.HasPrefix(data, "# ") {
		pkg := strings.Split(data, " ")[1]
		err := &BuildError{
			Package: pkg,
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
	} else if !strings.HasPrefix(data, "FAIL") && len(set.BuildErrors) > 0 {
		buildErr := set.BuildErrors[len(set.BuildErrors)-1]
		parts := strings.Split(data, ":")
		var lineNum, colNum int
		lineNum, _ = strconv.Atoi(parts[1])
		var message string
		if len(parts) <= 3 {
			message = parts[2]
		} else if len(parts) > 3 {
			colNum, _ = strconv.Atoi(parts[2])
			message = parts[3]
		}
		excp := []string{}
		if file, err := os.Open(parts[0]); err == nil {
			excp = excerpt.Excerpt(file, lineNum, colNum)
		}
		buildErr.Lines = append(buildErr.Lines, &BuildErrorLine{
			Path:    parts[0],
			Line:    lineNum,
			Column:  colNum,
			Message: message,
			Excerpt: excp,
		})
	} else {
		set.BuildErrors = append(set.BuildErrors, &BuildError{
			Lines: []*BuildErrorLine{{Message: data}},
		})
	}
}

func (set *Set) parseCoverProfile(coverPath, projectdir string) error {
	projectFiles, err := filesForPath(projectdir)
	if err != nil {
		return err
	}
	profiles, err := cover.ParseProfiles(coverPath)
	if err != nil {
		return err
	}
	filePathToProfileMap := make(map[string]*cover.Profile)
	for _, prof := range profiles {
		filePathToProfileMap[prof.FileName] = prof
	}
	for _, filePath := range projectFiles {
		pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, filepath.Dir(filePath))
		if err != nil {
			return err
		} else if len(pkgs) > 1 {
			return fmt.Errorf("expected only 1 package")
		}
		profile := filePathToProfileMap[fmt.Sprintf("%v/%v", pkgs[0].PkgPath, filepath.Base(filePath))]
		if _, ok := set.Packages[pkgs[0].PkgPath]; !ok {
			continue
		}
		pkg := set.Packages[pkgs[0].PkgPath]
		if err := pkg.fileCoverage(profile, filePath); err != nil {
			return err
		}
		set.StatementCount += pkg.StatementCount
		set.CoveredCount += pkg.CoveredCount
	}
	set.CoveragePercent = calcPercent(set.StatementCount, set.CoveredCount)
	return nil
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

func filesForPath(dir string) ([]string, error) {
	base := filepath.Base(dir)
	if base == "..." {
		dir = filepath.Dir(dir)
	}
	if fi, err := os.Stat(dir); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("path must be a directory")
	}
	recursive := base == "..."
	files := make([]string, 0)
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			if path == dir {
				return nil
			} else if !recursive {
				return filepath.SkipDir
			}
		}
		if regexp.MustCompile(".go$").MatchString(path) {
			if regexp.MustCompile("_test.go$").MatchString(path) {
				return nil
			}
			path, _ = filepath.Abs(path)
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func calcPercent(statements, covered int64) float64 {
	if statements > 0 {
		return math.Floor((float64(covered)/float64(statements))*10000) / 100
	}
	return 0
}
