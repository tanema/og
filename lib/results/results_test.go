package results

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tanema/og/lib/stopwatch"
)

func TestNew(t *testing.T) {
	set := New("github.com/tanema/og")

	assert.Equal(t, Pass, set.State)
	assert.Equal(t, "github.com/tanema/og", set.Mod)
	assert.NotNil(t, set.watch)
	assert.NotNil(t, set.Packages)
	assert.NotNil(t, set.Failures)
	assert.NotNil(t, set.Skips)
	assert.NotNil(t, set.BuildErrors)
}

func TestSetParse(t *testing.T) {
}

func TestSetAdd(t *testing.T) {
	t.Run("package event", func(t *testing.T) {
		set := New("github.com/tanema/og")
		assert.Equal(t, 0, len(set.Packages))
		set.Add(Run, "github.com/tanema/og/nope", "", "")
		assert.Equal(t, 1, len(set.Packages))
		assert.NotNil(t, set.Packages["./nope"])
		assert.Empty(t, set.Packages["./nope"].Results)
		assert.Equal(t, 0, set.TotalTests)
	})
	t.Run("test event", func(t *testing.T) {
		set := New("github.com/tanema/og")
		assert.Equal(t, 0, len(set.Packages))
		set.Add(Run, "github.com/tanema/og/nope", "TestFoo", "")
		assert.Equal(t, 1, len(set.Packages))
		assert.NotNil(t, set.Packages["./nope"])
		assert.Equal(t, 1, len(set.Packages["./nope"].Results))
		assert.Equal(t, 1, set.TotalTests)
	})
}

func TestSetPackageResult(t *testing.T) {
	setup := func() (*Set, *Package) {
		set := New("github.com/tanema/og")
		set.Packages["./nope"] = &Package{
			watch:   stopwatch.Start(),
			Name:    "./nope",
			State:   Run,
			Results: map[string]*Test{},
		}
		return set, set.Packages["./nope"]
	}

	t.Run("pass", func(t *testing.T) {
		set, pkg := setup()
		set.packageResult(Pass, pkg, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 1, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Pass, pkg.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("fail", func(t *testing.T) {
		set, pkg := setup()
		set.packageResult(Fail, pkg, "")
		assert.Equal(t, Fail, set.State)
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 1, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Fail, pkg.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("skip", func(t *testing.T) {
		set, pkg := setup()
		set.packageResult(Skip, pkg, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 1, set.PkgSummary.Skip)
		assert.Equal(t, Skip, pkg.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("paus", func(t *testing.T) {
		set, pkg := setup()
		set.packageResult(Pause, pkg, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Pause, pkg.State)
		assert.False(t, pkg.Cached)
		assert.True(t, pkg.watch.Paused())
	})
	t.Run("cont", func(t *testing.T) {
		set, pkg := setup()
		set.packageResult(Continue, pkg, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Continue, pkg.State)
		assert.False(t, pkg.Cached)
		assert.False(t, pkg.watch.Paused())
	})
	t.Run("run", func(t *testing.T) {
		set, pkg := setup()
		set.packageResult(Run, pkg, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Run, pkg.State)
		assert.False(t, pkg.Cached)
		assert.False(t, pkg.watch.Paused())
	})
	t.Run("output", func(t *testing.T) {
		set, pkg := setup()
		set.packageResult(Output, pkg, "this is output")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Run, pkg.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("cached", func(t *testing.T) {
		set, pkg := setup()
		set.packageResult(Output, pkg, "ok package (cached)")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 1, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Run, pkg.State)
		assert.True(t, pkg.Cached)
	})
}

func TestSetTestResult(t *testing.T) {
	setup := func() (*Set, *Package, *Test) {
		set := New("github.com/tanema/og")
		set.Packages["./nope"] = &Package{
			watch: stopwatch.Start(),
			Name:  "./nope",
			State: Run,
			Results: map[string]*Test{"TestFoo": {
				watch:   stopwatch.Start(),
				Name:    "TestFoo",
				State:   Run,
				Package: "./nope",
			}},
		}
		return set, set.Packages["./nope"], set.Packages["./nope"].Results["TestFoo"]
	}

	t.Run("pass", func(t *testing.T) {
		set, pkg, test := setup()
		set.testResult(Pass, pkg, test, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 1, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.False(t, pkg.Cached)
	})
	t.Run("fail", func(t *testing.T) {
		set, pkg, test := setup()
		set.testResult(Fail, pkg, test, "")
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 1, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Fail, test.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("skip", func(t *testing.T) {
		set, pkg, test := setup()
		set.testResult(Skip, pkg, test, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 1, pkg.Skip)
		assert.Equal(t, Skip, test.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("paus", func(t *testing.T) {
		set, pkg, test := setup()
		set.testResult(Pause, pkg, test, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Pause, test.State)
		assert.True(t, test.watch.Paused())
	})
	t.Run("cont", func(t *testing.T) {
		set, pkg, test := setup()
		set.testResult(Continue, pkg, test, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Continue, test.State)
		assert.False(t, pkg.watch.Paused())
	})
	t.Run("run", func(t *testing.T) {
		set, pkg, test := setup()
		set.testResult(Run, pkg, test, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Run, test.State)
		assert.False(t, pkg.watch.Paused())
	})
	t.Run("output", func(t *testing.T) {
		set, pkg, test := setup()
		set.testResult(Output, pkg, test, "this is output")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Run, test.State)
		assert.Equal(t, []string{"this is output"}, test.Messages)
	})
}

func TestCleanMessage(t *testing.T) {
	tests := []struct {
		name, in, out string
	}{
		{"TestA", "=== RUN   TestA", ""},
		{"TestC", "=== CONT  TestC", ""},
		{"TestC", "=== PAUSE  TestC", ""},
		{"TestA", "--- PASS: TestA (0.00s)", ""},
		{"TestF/testg", "=== RUN   TestF/testg", ""},
		{"TestB", "mess_test.go:13: failed", "mess_test.go:13: failed"},
		{"TestB", "--- FAIL: TestB (0.00s)", ""},
		{"TestE", "--- SKIP: TestE (0.00s)", ""},
		{"TestF/testh", "mess_test.go:40: Skipped with val: clippy", "mess_test.go:40: Skipped with val: clippy"},
		{"TestF/testg", "    --- PASS: TestF/testg (0.00s)", ""},
		{"TestF/testg", "FAIL", ""},
		{"TestA", "					Test:					TestA", ""},
		{"TestF/testg", "					Error:				Expected nil, but got: true", "					Error:				Expected nil, but got: true"},
	}
	for _, testcase := range tests {
		test := &Test{Name: testcase.name, Messages: []string{}}
		test.AddMessage(testcase.in)
		if testcase.out == "" {
			assert.Empty(t, test.Messages)
		} else {
			assert.Equal(t, testcase.out, test.Messages[0])
		}
	}
}

func TestSetFilteredPackages(t *testing.T) {
	set := New("github.com/tanema/og")
	set.Packages = map[string]*Package{
		"./one":   {Name: "./one", State: Pass},
		"./two":   {Name: "./two", State: Skip},
		"./three": {Name: "./three", State: Fail},
	}

	t.Run("filter skipped", func(t *testing.T) {
		assert.Equal(t, map[string]*Package{
			"./one":   {Name: "./one", State: Pass},
			"./three": {Name: "./three", State: Fail},
		}, set.FilteredPackages(false))
	})

	t.Run("no filter", func(t *testing.T) {
		assert.Equal(t, set.Packages, set.FilteredPackages(true))
	})
}

func TestSetFilteredTests(t *testing.T) {
	set := New("github.com/tanema/og")
	good1 := &Test{Name: "TestA", TimeElapsed: time.Millisecond}
	good2 := &Test{Name: "TestB", TimeElapsed: time.Nanosecond}
	bad1 := &Test{Name: "TestC", TimeElapsed: time.Second}
	bad2 := &Test{Name: "TestD", TimeElapsed: time.Minute}
	set.Packages = map[string]*Package{
		"./one":   {Results: map[string]*Test{"TestA": good1}},
		"./two":   {Results: map[string]*Test{"TestB": good2, "TestC": bad1}},
		"./three": {Results: map[string]*Test{"TestD": bad2}},
	}

	assert.Equal(t, []*Test{bad2, bad1}, set.RankedTests(time.Second))
}
