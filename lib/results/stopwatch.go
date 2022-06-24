package results

import "time"

type stopwatch struct {
	started time.Time
	paused  bool
	parts   []time.Duration
	total   time.Duration
}

func (watch *stopwatch) start() {
	if watch.paused {
		watch.started = time.Now()
		watch.paused = false
	} else if watch.started.IsZero() {
		watch.total = 0
		watch.started = time.Now()
		watch.paused = false
	}
}

func (watch *stopwatch) running() bool {
	return !watch.paused && !watch.started.IsZero()
}

func (watch *stopwatch) Elapsed() time.Duration {
	if !watch.running() {
		return watch.total.Round(time.Millisecond)
	}
	return (watch.total + time.Since(watch.started)).Round(time.Millisecond)
}

func (watch *stopwatch) pause() time.Duration {
	if !watch.paused {
		watch.total += time.Since(watch.started)
		watch.paused = true
	}
	return watch.Elapsed()
}

func (watch *stopwatch) stop() time.Duration {
	if watch.running() {
		watch.total += time.Since(watch.started)
		watch.started = time.Time{}
	}
	return watch.Elapsed()
}
