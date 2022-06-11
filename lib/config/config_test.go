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
	cfg.findRoot("./")
	assert.Equal(t, "github.com/tanema/og", cfg.ModName)

	cfg.findRoot("/tmp")
	assert.Equal(t, cfg.Root, cfg.ModName)
}
