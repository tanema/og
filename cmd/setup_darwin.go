//go:build darwin && amd64
// +build darwin,amd64

package cmd

import (
	"math"
	"syscall"
)

const minFileDescriptors float64 = 2048

// MacOSX sets a very low default file descriptor limit per process.
// This function sets file descriptor limits to a more sane value.
func init() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic("Could not read max file limit")
	}
	rLimit.Cur = uint64(math.Max(minFileDescriptors, float64(rLimit.Cur)))
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic("Could not set file descriptor limits. you might encounter issues if your project holds many files. You can set the limits manually using ulimit -n 2048")
	}
}
