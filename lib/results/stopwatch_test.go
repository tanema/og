package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopwatch(t *testing.T) {
	watch := &stopwatch{}
	assert.True(t, watch.started.IsZero())
	assert.False(t, watch.paused)
	assert.False(t, watch.running())

	watch.start()
	assert.False(t, watch.started.IsZero())
	assert.False(t, watch.paused)
	assert.True(t, watch.running())

	watch.pause()
	assert.True(t, watch.paused)
	assert.False(t, watch.running())

	watch.start()
	assert.True(t, watch.running())
	assert.False(t, watch.paused)

	watch.start()
	assert.False(t, watch.paused)
	assert.True(t, watch.running())

	watch.stop()
	assert.False(t, watch.paused)
	assert.False(t, watch.running())
}
