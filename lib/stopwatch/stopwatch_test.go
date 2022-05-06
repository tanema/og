package stopwatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopwatch(t *testing.T) {
	watch := Start()
	assert.NotNil(t, watch.start)
	watch.Pause()
	assert.True(t, watch.paused)
	assert.True(t, watch.Paused())
	watch.Resume()
	assert.False(t, watch.paused)
	assert.False(t, watch.Paused())
	dur := watch.Stop()
	assert.True(t, watch.paused)
	assert.True(t, dur > 0)
}
