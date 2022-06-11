package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/display"
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
		assert.Nil(t, validateDisplay(cfg, "test"))
		assert.Equal(t, display.Decorators["dots"], cfg.ResultsTemplate)
	})
	t.Run("does not change custom templates", func(t *testing.T) {
		cfg := &config.Config{Display: "dots", ResultsTemplate: "custom"}
		assert.Nil(t, validateDisplay(cfg, "test"))
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
		path, tests := findPaths([]string{"./root_test.go"})
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

	t.Run("extended cfg flags", func(t *testing.T) {
		cfg := &config.Config{
			NoCache:  true,
			Short:    true,
			FailFast: true,
			Shuffle:  true,
			Cover:    "file.out",
		}
		args, err := fmtTestArgs(cfg)
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v",
			"-count=1", "-short", "-failfast", "-shuffle",
			"on", "-covermode", "atomic", "-coverprofile", "file.out", "./...",
		}, args)
	})
}

func TestPrintVersion(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.Config{Out: &buf}
	printVersion(cfg)
}

func TestCenterString(t *testing.T) {
	assert.Equal(t, "         foo        ", centerString("foo", 20))
	assert.Equal(t, " foo ", centerString("foo", 5))
	assert.Equal(t, "  foo ", centerString("foo", 6))
	assert.Equal(t, "   foo  ", centerString("foo", 8))
}

func TestFindPaths(t *testing.T) {
	cases := []struct {
		args  []string
		paths []string
		tests []string
	}{
		{paths: []string{"./..."}},
		{args: []string{"testdata"}, paths: []string{"./testdata"}},
		{args: []string{"./testdata"}, paths: []string{"./testdata"}},
		{args: []string{"testdata/go.go"}, paths: []string{"./testdata"}, tests: []string{"TestHelloWorld", "TestGoodbyeWorld"}},
		{args: []string{"testdata/go_test.go"}, paths: []string{"./testdata"}, tests: []string{"TestHelloWorld", "TestGoodbyeWorld"}},
		{args: []string{"testdata/go_test.go:6"}, paths: []string{"./testdata"}, tests: []string{"TestHelloWorld"}},
		{args: []string{"testdata/go_test.go:TestGoodbyeWorld"}, paths: []string{"./testdata"}, tests: []string{"TestGoodbyeWorld"}},
		{args: []string{"TestGoodbyeWorld"}, paths: []string{"./..."}, tests: []string{"TestGoodbyeWorld"}},
		{args: []string{"testdata", "TestGoodbyeWorld"}, paths: []string{"./testdata"}, tests: []string{"TestGoodbyeWorld"}},
		{args: []string{"testdata", "TestHelloWorld", "TestGoodbyeWorld"}, paths: []string{"./testdata"}, tests: []string{"TestHelloWorld", "TestGoodbyeWorld"}},
		{args: []string{"testdata", "TestFoo"}, paths: []string{"./testdata"}, tests: []string{"TestFoo"}},
	}

	for _, testcase := range cases {
		paths, tests := findPaths(testcase.args)
		assert.Equal(t, testcase.paths, paths)
		assert.Equal(t, testcase.tests, tests)
	}

}

func TestFindTestsInFile(t *testing.T) {
	cases := []struct {
		path  string
		line  int
		tests []string
	}{
		{path: "./testdata", line: -1},
		{path: "./testdata/not_there.go", line: -1},
		{path: "./testdata/go.go", line: -1, tests: []string{}},
		{path: "./testdata/go_test.go", line: -1, tests: []string{"TestHelloWorld", "TestGoodbyeWorld"}},
		{path: "./testdata/go_test.go", line: 2, tests: []string{"TestHelloWorld"}},
		{path: "./testdata/go_test.go", line: 3, tests: []string{"TestHelloWorld"}},
		{path: "./testdata/go_test.go", line: 4, tests: []string{"TestHelloWorld"}},
		{path: "./testdata/go_test.go", line: 5, tests: []string{"TestHelloWorld"}},
		{path: "./testdata/go_test.go", line: 6, tests: []string{"TestHelloWorld"}},
		{path: "./testdata/go_test.go", line: 8, tests: []string{"TestHelloWorld"}},
		{path: "./testdata/go_test.go", line: 9, tests: []string{"TestGoodbyeWorld"}},
		{path: "./testdata/go_test.go", line: 10, tests: []string{"TestGoodbyeWorld"}},
		{path: "./testdata/go_test.go", line: 11, tests: []string{"TestGoodbyeWorld"}},
		{path: "./testdata/go_test.go", line: 12, tests: []string{"TestGoodbyeWorld"}},
	}

	for i, testcase := range cases {
		tests := findTestsInFile(testcase.path, testcase.line)
		assert.Equal(t, testcase.tests, tests, fmt.Sprintf("testcase %v", i))
	}
}
