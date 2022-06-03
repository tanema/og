package cmd

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/display"
	"github.com/tanema/og/lib/find"
	"github.com/tanema/og/lib/term"
)

type strWCr struct {
	*bytes.Buffer
}

func newStrWCr() *strWCr {
	return &strWCr{bytes.NewBufferString("")}
}

func (str *strWCr) Close() error {
	return nil
}

func TestValidateDisplay(t *testing.T) {
	t.Run("validate that the display is valid", func(t *testing.T) {
		cfg := &config.Config{Display: "nope"}
		assert.EqualError(t, validateDisplay(cfg), "Unknown display type nope")
	})
	t.Run("sets default templates", func(t *testing.T) {
		cfg := &config.Config{Display: "dots"}
		assert.Nil(t, validateDisplay(cfg))
		assert.Equal(t, display.Decorators["dots"], cfg.ResultsTemplate)
	})
	t.Run("does not change custom templates", func(t *testing.T) {
		cfg := &config.Config{Display: "dots", ResultsTemplate: "custom"}
		assert.Nil(t, validateDisplay(cfg))
		assert.Equal(t, "custom", cfg.ResultsTemplate)
	})
}

func TestFmtArgs(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtTestArgs(cfg)
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "./..."}, args)
	})

	t.Run("test names", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtTestArgs(cfg, "TestFmtArgs", "TestValidateDisplay")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestFmtArgs|TestValidateDisplay", "./..."}, args)
	})

	t.Run("filepaths", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtTestArgs(cfg, "./root_test.go")
		assert.Nil(t, err)
		path, tests, _ := find.Paths([]string{"./root_test.go"})
		assert.Equal(t, append([]string{"go", "test", "-json", "-v", "-run", strings.Join(tests, "|")}, path...), args)
	})

	t.Run("filepaths with numbers", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtTestArgs(cfg, "./root_test.go:30")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestValidateDisplay", "./."}, args)
	})

	t.Run("filepaths with test names", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtTestArgs(cfg, "./root_test.go:TestFmtArgs")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestFmtArgs", "./."}, args)
	})

	t.Run("package", func(t *testing.T) {
		cfg := &config.Config{}
		args, err := fmtTestArgs(cfg, "./")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "./"}, args)
	})

	t.Run("cfg flags", func(t *testing.T) {
		cfg := &config.Config{NoCache: true}
		args, err := fmtTestArgs(cfg)
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-count=1", "./..."}, args)
	})
}

func TestPrintVersion(t *testing.T) {
	buf := newStrWCr()
	printVersion(&config.Config{Out: buf})
	expected := term.Sprintf("{{. | Rainbow}}\n", fmt.Sprintf("og%v.%v\t%v", major, minor, runtime.Version()))
	assert.Equal(t, expected, buf.String())
}
