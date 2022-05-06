package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tanema/og/lib/stopwatch"
)

func TestTestResult(t *testing.T) {
	setup := func() (*Set, *Package, *Test) {
		set := New("github.com/tanema/og", "/workspace/og")
		set.Packages["nope"] = &Package{
			watch: stopwatch.Start(),
			Name:  "nope",
			State: Run,
			Tests: map[string]*Test{"TestFoo": {
				watch:   stopwatch.Start(),
				Name:    "TestFoo",
				State:   Run,
				Package: "nope",
			}},
		}
		return set, set.Packages["nope"], set.Packages["nope"].Tests["TestFoo"]
	}

	t.Run("pass", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Pass, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 1, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.False(t, pkg.Cached)
	})
	t.Run("fail", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Fail, "")
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 1, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Fail, test.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("skip", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Skip, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 1, pkg.Skip)
		assert.Equal(t, Skip, test.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("paus", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Pause, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Pause, test.State)
		assert.True(t, test.watch.Paused())
	})
	t.Run("cont", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Continue, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Continue, test.State)
		assert.False(t, pkg.watch.Paused())
	})
	t.Run("run", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Run, "")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Run, test.State)
		assert.False(t, pkg.watch.Paused())
	})
	t.Run("output", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Output, "this is output")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Run, test.State)
	})
}
