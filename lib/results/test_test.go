package results

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestResult(t *testing.T) {
	setup := func() (*Set, *Package, *Test) {
		set := New("", 10*time.Minute)
		set.Packages["nope"] = &Package{
			stopwatch: &stopwatch{},
			Name:      "nope",
			State:     Run,
			Tests: map[string]*Test{"TestFoo": {
				stopwatch: &stopwatch{},
				Name:      "TestFoo",
				State:     Run,
				Package:   "nope",
			}},
		}
		return set, set.Packages["nope"], set.Packages["nope"].Tests["TestFoo"]
	}

	t.Run("pass", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Pass, "")
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
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 1, pkg.Skip)
		assert.Equal(t, Skip, test.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("paus", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Pause, "")
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Pause, test.State)
		assert.True(t, test.paused)
	})
	t.Run("cont", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Continue, "")
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Continue, test.State)
		assert.False(t, pkg.paused)
	})
	t.Run("run", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Run, "")
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Run, test.State)
		assert.False(t, pkg.paused)
	})
	t.Run("output", func(t *testing.T) {
		set, pkg, test := setup()
		test.result(set, pkg, Output, "this is output")
		assert.Equal(t, 0, pkg.Pass)
		assert.Equal(t, 0, pkg.Fail)
		assert.Equal(t, 0, pkg.Skip)
		assert.Equal(t, Run, test.State)
	})
}
