package results

import (
	"strings"
)

// Package is the test results for a single package
type Package struct {
	*stopwatch
	Summary
	Name            string           `json:"name"`
	Tests           map[string]*Test `json:"tests,omitempty"`
	State           Action           `json:"state"`
	Cached          bool             `json:"cached,omitempty"`
	StatementCount  int64            `json:"statements,omitempty"`
	CoveredCount    int64            `json:"covered,omitempty"`
	CoveragePercent float64          `json:"percent,omitempty"`
}

func newPackage(name string) *Package {
	pkg := &Package{
		stopwatch: &stopwatch{},
		Name:      name,
		State:     Run,
		Tests:     map[string]*Test{},
	}
	pkg.start()
	return pkg
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
	case Run, Continue:
		pkg.start()
	case Pause:
		pkg.pause()
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
		pkg.stop()
	}
}
