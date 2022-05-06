package results

import (
	"strings"
	"time"

	"github.com/tanema/og/lib/stopwatch"
)

// Package is the test results for a single package
type Package struct {
	Summary
	Name        string           `json:"name"`
	Tests       map[string]*Test `json:"tests"`
	State       Action           `json:"state"`
	Cached      bool             `json:"cached"`
	TimeElapsed time.Duration    `json:"time_elapsed"`
	watch       *stopwatch.Stopwatch
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
		if strings.HasPrefix(output, "ok") && strings.Contains(output, "(cached)") {
			set.Cached++
			pkg.Cached = true
		}
	}
	if action != Output {
		pkg.State = action
	}
	if action == Pass || action == Fail || action == Skip {
		pkg.TimeElapsed = pkg.watch.Stop()
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
