package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	cfg := &Config{}
	assert.Nil(t, cfg.Load())
}

func TestFindRoot(t *testing.T) {
	cfg := &Config{}
	cfg.findRoot()
	assert.Equal(t, "github.com/tanema/og", cfg.ModName)
}

func TestArgs(t *testing.T) {
	cfg := &Config{
		NoCache:  true,
		Short:    true,
		FailFast: true,
		Shuffle:  true,
		Cover:    "file.out",
	}
	assert.Equal(t, []string{"go", "test", "-json", "-v",
		"-count=1", "-short", "-failfast", "-shuffle",
		"on", "-covermode", "atomic", "-coverprofile", "file.out",
	}, cfg.TestArgs())
}
