package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/display"
	"github.com/tanema/og/lib/find"
)

func TestValidateDisplay(t *testing.T) {
	t.Run("validate that the display is valid", func(t *testing.T) {
		cfg := &config.Config{Display: "nope"}
		assert.EqualError(t, validateDisplay(cfg), "Unknown display type nope")
	})
	t.Run("sets default templates", func(t *testing.T) {
		cfg := &config.Config{Display: "dots"}
		assert.Nil(t, validateDisplay(cfg))
		assert.Equal(t, display.Decorators["dots"], cfg.ResultsTemplate)
		assert.Equal(t, display.SummaryTemplate, cfg.SummaryTemplate)
	})
	t.Run("does not change custom templates", func(t *testing.T) {
		cfg := &config.Config{Display: "dots", ResultsTemplate: "custom", SummaryTemplate: "template"}
		assert.Nil(t, validateDisplay(cfg))
		assert.Equal(t, "custom", cfg.ResultsTemplate)
		assert.Equal(t, "template", cfg.SummaryTemplate)
	})
}

func TestFmtArgs(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtArgs(cfg)
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "./..."}, args)
	})

	t.Run("test names", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtArgs(cfg, "TestFmtArgs", "TestValidateDisplay")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestFmtArgs|TestValidateDisplay", "./..."}, args)
	})

	t.Run("filepaths", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtArgs(cfg, "./root_test.go")
		assert.Nil(t, err)
		path, tests, _ := find.Paths([]string{"./root_test.go"})
		assert.Equal(t, append([]string{"go", "test", "-json", "-v", "-run", strings.Join(tests, "|")}, path...), args)
	})

	t.Run("filepaths with numbers", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtArgs(cfg, "./root_test.go:20")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestValidateDisplay", "./."}, args)
	})

	t.Run("filepaths with test names", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtArgs(cfg, "./root_test.go:TestFmtArgs")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestFmtArgs", "./."}, args)
	})

	t.Run("package", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtArgs(cfg, "./")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "./"}, args)
	})

	t.Run("cfg flags", func(t *testing.T) {
		cfg := &config.Config{NoCache: true}
		args, err := fmtArgs(cfg)
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-count=1", "./..."}, args)
	})
}

func TestRunCommand(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	runCommand(buf, "echo", "hello world")
	assert.Equal(t, "hello world\n", buf.String())
}
