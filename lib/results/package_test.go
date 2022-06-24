package results

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPackageResult(t *testing.T) {
	setup := func() (*Set, *Package) {
		set := New("", 10*time.Minute)
		set.Packages["nope"] = &Package{
			stopwatch: &stopwatch{},
			Name:      "nope",
			State:     Run,
			Tests:     map[string]*Test{},
		}
		return set, set.Packages["nope"]
	}

	t.Run("pass", func(t *testing.T) {
		set, pkg := setup()
		pkg.result(set, Pass, "")
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 1, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Pass, pkg.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("fail", func(t *testing.T) {
		set, pkg := setup()
		pkg.result(set, Fail, "")
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
		pkg.result(set, Skip, "")
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 1, set.PkgSummary.Skip)
		assert.Equal(t, Skip, pkg.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("paus", func(t *testing.T) {
		set, pkg := setup()
		pkg.result(set, Pause, "")
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Pause, pkg.State)
		assert.False(t, pkg.Cached)
		assert.True(t, pkg.paused)
	})
	t.Run("cont", func(t *testing.T) {
		set, pkg := setup()
		pkg.result(set, Continue, "")
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Continue, pkg.State)
		assert.False(t, pkg.Cached)
		assert.False(t, pkg.paused)
	})
	t.Run("run", func(t *testing.T) {
		set, pkg := setup()
		pkg.result(set, Run, "")
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Run, pkg.State)
		assert.False(t, pkg.Cached)
		assert.False(t, pkg.paused)
	})
	t.Run("output", func(t *testing.T) {
		set, pkg := setup()
		pkg.result(set, Output, "this is output")
		assert.Equal(t, 0, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Run, pkg.State)
		assert.False(t, pkg.Cached)
	})
	t.Run("cached", func(t *testing.T) {
		set, pkg := setup()
		pkg.result(set, Output, "ok package (cached)")
		assert.Equal(t, 1, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Run, pkg.State)
		assert.True(t, pkg.Cached)
	})
}
