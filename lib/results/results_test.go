package results

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tanema/og/lib/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{ModName: "github.com/tanema/og", Root: "/workspace/og"}
	set := New(cfg, "")

	assert.Equal(t, Pass, set.State)
	assert.Equal(t, "github.com/tanema/og", set.Mod)
	assert.NotNil(t, set.watch)
	assert.NotNil(t, set.Packages)
	assert.NotNil(t, set.Failures)
	assert.NotNil(t, set.Skips)
	assert.NotNil(t, set.BuildErrors)
}

func TestSetParse(t *testing.T) {
	t.Run("json package event", func(t *testing.T) {
		cfg := &config.Config{ModName: "github.com/tanema/og", Root: "/workspace/og"}
		set := New(cfg, "")
		set.Parse([]byte(`{"Package": "github.com/tanema/og/nope", "Action": "pass"}`))
		assert.Equal(t, set.Packages["github.com/tanema/og/nope"].State, Pass)
	})
	t.Run("json test event", func(t *testing.T) {
		cfg := &config.Config{ModName: "github.com/tanema/og", Root: "/workspace/og"}
		set := New(cfg, "")
		set.Parse([]byte(`{"Package": "github.com/tanema/og/nope", "Test": "TestA", "Action": "pass"}`))
		assert.Equal(t, set.Packages["github.com/tanema/og/nope"].Tests["TestA"].State, Pass)
		assert.Empty(t, set.BuildErrors)
	})
}

func TestSetAdd(t *testing.T) {
	t.Run("package event", func(t *testing.T) {
		cfg := &config.Config{ModName: "github.com/tanema/og", Root: "/workspace/og"}
		set := New(cfg, "")
		assert.Equal(t, 0, len(set.Packages))
		set.Add(Run, "github.com/tanema/og/nope", "", "")
		assert.Equal(t, 1, len(set.Packages))
		assert.NotNil(t, set.Packages["github.com/tanema/og/nope"])
		assert.Empty(t, set.Packages["github.com/tanema/og/nope"].Tests)
		assert.Equal(t, 0, set.TotalTests)
	})
	t.Run("test event", func(t *testing.T) {
		cfg := &config.Config{ModName: "github.com/tanema/og", Root: "/workspace/og"}
		set := New(cfg, "")
		assert.Equal(t, 0, len(set.Packages))
		set.Add(Run, "github.com/tanema/og/nope", "TestFoo", "")
		assert.Equal(t, 1, len(set.Packages))
		assert.NotNil(t, set.Packages["github.com/tanema/og/nope"])
		assert.Equal(t, 1, len(set.Packages["github.com/tanema/og/nope"].Tests))
		assert.Equal(t, 1, set.TotalTests)
	})
}

func TestSetFilteredPackages(t *testing.T) {
	cfg := &config.Config{ModName: "github.com/tanema/og", Root: "/workspace/og"}
	set := New(cfg, "")
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
	cfg := &config.Config{ModName: "github.com/tanema/og", Root: "/workspace/og"}
	set := New(cfg, "")
	good1 := &Test{Name: "TestA", TimeElapsed: time.Millisecond}
	good2 := &Test{Name: "TestB", TimeElapsed: time.Nanosecond}
	bad1 := &Test{Name: "TestC", TimeElapsed: time.Second}
	bad2 := &Test{Name: "TestD", TimeElapsed: time.Minute}
	set.Packages = map[string]*Package{
		"./one":   {Tests: map[string]*Test{"TestA": good1}},
		"./two":   {Tests: map[string]*Test{"TestB": good2, "TestC": bad1}},
		"./three": {Tests: map[string]*Test{"TestD": bad2}},
	}

	assert.Equal(t, []*Test{bad2, bad1}, set.RankedTests(time.Second))
}
