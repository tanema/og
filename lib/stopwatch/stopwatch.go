package stopwatch

import "time"

// Stopwatch is a mechanism to track the total elapsed duration of something
// that pauses and continues
type Stopwatch struct {
	start  time.Time
	paused bool
	parts  []time.Duration
	total  time.Duration
}

// Start will create a new stopwatch
func Start() *Stopwatch {
	return &Stopwatch{start: time.Now()}
}

// Pause will stop the stopwatch and record the current amount of time
func (watch *Stopwatch) Pause() time.Duration {
	if !watch.paused {
		watch.total += time.Since(watch.start)
		watch.paused = true
	}
	return watch.total
}

// Paused returns if the watch is paused
func (watch *Stopwatch) Paused() bool {
	return watch.paused
}

// Resume will restart the stopwatch if paused
func (watch *Stopwatch) Resume() {
	if !watch.paused {
		return
	}
	watch.start = time.Now()
	watch.paused = false
}

// Stop will stop the stopwatch, return the total and reset the counter
func (watch *Stopwatch) Stop() time.Duration {
	total := watch.Pause()
	watch.total = 0
	return total
}
