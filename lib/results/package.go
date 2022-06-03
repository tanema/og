package results

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
	"time"

	"github.com/tanema/og/lib/stopwatch"
	"golang.org/x/tools/cover"
)

// Package is the test results for a single package
type Package struct {
	Summary
	Name            string           `json:"name"`
	Tests           map[string]*Test `json:"tests"`
	State           Action           `json:"state"`
	Cached          bool             `json:"cached"`
	TimeElapsed     time.Duration    `json:"time_elapsed"`
	StatementCount  int64            `json:"statements"`
	CoveredCount    int64            `json:"covered"`
	CoveragePercent float64          `json:"percent"`
	watch           *stopwatch.Stopwatch
}

func newPackage(name string) *Package {
	return &Package{
		watch: stopwatch.Start(),
		Name:  name,
		State: Run,
		Tests: map[string]*Test{},
	}
}

func (pkg *Package) result(set *Set, action Action, output string) {
	switch action {
	case Pass:
		set.PkgSummary.Pass++
	case Fail:
		set.State = Fail
		set.PkgSummary.Fail++
	case Skip:
		set.PkgSummary.Skip++
	case Pause:
		pkg.TimeElapsed = pkg.watch.Pause()
	case Continue:
		pkg.watch.Resume()
	case Output:
		pkg.addLogOutput(set, output)
	}
	if action != Output {
		pkg.State = action
	}
	if action == Pass || action == Fail || action == Skip {
		pkg.TimeElapsed = pkg.watch.Stop()
	}
}

func (pkg *Package) addLogOutput(set *Set, msg string) {
	if strings.HasPrefix(msg, "ok") && strings.Contains(msg, "(cached)") {
		set.Cached++
		pkg.Cached = true
	}
}

// FilteredTests filters out skipped tests.
func (pkg *Package) FilteredTests(filterSkip bool) map[string]*Test {
	if filterSkip {
		return pkg.Tests
	}
	filtered := map[string]*Test{}
	for name, test := range pkg.Tests {
		if test.State != Skip {
			filtered[name] = test
		}
	}
	return filtered
}

func (pkg *Package) fileCoverage(profile *cover.Profile, filePath string) error {
	src, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, src, 0)
	if err != nil {
		return err
	}
	for i := range node.Decls {
		switch x := node.Decls[i].(type) {
		case *ast.FuncDecl:
			if profile != nil {
				start := fset.Position(x.Pos())
				end := fset.Position(x.End())
				for _, block := range profile.Blocks {
					if block.StartLine > end.Line || (block.StartLine == end.Line && block.StartCol >= end.Column) {
						// Block starts after the function statement ends
						continue
					} else if block.EndLine < start.Line || (block.EndLine == start.Line && block.EndCol <= start.Column) {
						// Block ends before the function statement starts
						continue
					}
					pkg.StatementCount += int64(block.NumStmt)
					if block.Count > 0 {
						pkg.CoveredCount += int64(block.NumStmt)
					}
				}
			}
			pkg.CoveragePercent = calcPercent(pkg.StatementCount, pkg.CoveredCount)
		}
	}
	return nil
}
