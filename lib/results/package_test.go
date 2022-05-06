package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tanema/og/lib/stopwatch"
)

func TestPackageResult(t *testing.T) {
	setup := func() (*Set, *Package) {
		set := New("github.com/tanema/og", "/workspace/og")
		set.Packages["nope"] = &Package{
			watch: stopwatch.Start(),
			Name:  "nope",
			State: Run,
			Tests: map[string]*Test{},
		}
		return set, set.Packages["nope"]
	}

	t.Run("pass", func(t *testing.T) {
		set, pkg := setup()
		pkg.result(set, Pass, "")
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
		pkg.result(set, Pause, "")
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
		pkg.result(set, Continue, "")
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
		pkg.result(set, Run, "")
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
		pkg.result(set, Output, "this is output")
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
		pkg.result(set, Output, "ok package (cached)")
		assert.Equal(t, Pass, set.State)
		assert.Equal(t, 1, set.Cached)
		assert.Equal(t, 0, set.PkgSummary.Pass)
		assert.Equal(t, 0, set.PkgSummary.Fail)
		assert.Equal(t, 0, set.PkgSummary.Skip)
		assert.Equal(t, Run, pkg.State)
		assert.True(t, pkg.Cached)
	})
}
