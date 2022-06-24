package results

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	set := New("", 10*time.Minute)

	assert.Equal(t, Run, set.State)
	assert.NotNil(t, set.stopwatch)
	assert.NotNil(t, set.Packages)
	assert.NotNil(t, set.BuildErrors)
}

func TestSetParse(t *testing.T) {
	t.Run("json package event", func(t *testing.T) {
		set := New("", 10*time.Minute)
		set.Parse([]byte(`{"Package": "github.com/tanema/og/nope", "Action": "pass"}`))
		assert.Equal(t, set.Packages["github.com/tanema/og/nope"].State, Pass)
	})
	t.Run("json test event", func(t *testing.T) {
		set := New("", 10*time.Minute)
		set.Parse([]byte(`{"Package": "github.com/tanema/og/nope", "Test": "TestA", "Action": "pass"}`))
		assert.Equal(t, set.Packages["github.com/tanema/og/nope"].Tests["TestA"].State, Pass)
		assert.Empty(t, set.BuildErrors)
	})
}

func TestSetAdd(t *testing.T) {
	t.Run("package event", func(t *testing.T) {
		set := New("", 10*time.Minute)
		assert.Equal(t, 0, len(set.Packages))
		set.Add(Run, "github.com/tanema/og/nope", "", "")
		assert.Equal(t, 1, len(set.Packages))
		assert.NotNil(t, set.Packages["github.com/tanema/og/nope"])
		assert.Empty(t, set.Packages["github.com/tanema/og/nope"].Tests)
		assert.Equal(t, 0, set.TotalTests)
	})
	t.Run("test event", func(t *testing.T) {
		set := New("", 10*time.Minute)
		assert.Equal(t, 0, len(set.Packages))
		set.Add(Run, "github.com/tanema/og/nope", "TestFoo", "")
		assert.Equal(t, 1, len(set.Packages))
		assert.NotNil(t, set.Packages["github.com/tanema/og/nope"])
		assert.Equal(t, 1, len(set.Packages["github.com/tanema/og/nope"].Tests))
		assert.Equal(t, 1, set.TotalTests)
	})
}
