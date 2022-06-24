package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestFmtArgs(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		args, err := fmtTestArgs(rootCmd, &Config{NoCover: true})
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "./..."}, args)
	})

	t.Run("test names", func(t *testing.T) {
		args, err := fmtTestArgs(rootCmd, &Config{NoCover: true}, "TestFmtArgs", "TestValidateDisplay")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestFmtArgs|TestValidateDisplay", "./..."}, args)
	})

	t.Run("filepaths", func(t *testing.T) {
		args, err := fmtTestArgs(rootCmd, &Config{NoCover: true}, "./root_test.go")
		assert.Nil(t, err)
		path, tests := findPaths([]string{"./root_test.go"})
		assert.Equal(t, append([]string{"go", "test", "-json", "-v", "-run", strings.Join(tests, "|")}, path...), args)
	})

	t.Run("filepaths with numbers", func(t *testing.T) {
		args, err := fmtTestArgs(rootCmd, &Config{NoCover: true}, "./root_test.go:30")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestFmtArgs", "./."}, args)
	})

	t.Run("filepaths with test names", func(t *testing.T) {
		args, err := fmtTestArgs(rootCmd, &Config{NoCover: true}, "./root_test.go:TestFmtArgs")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "-run", "TestFmtArgs", "./."}, args)
	})

	t.Run("package", func(t *testing.T) {
		args, err := fmtTestArgs(rootCmd, &Config{NoCover: true}, "./")
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "./"}, args)
	})

	t.Run("cfg flags", func(t *testing.T) {
		args, err := fmtTestArgs(rootCmd, &Config{NoCover: true})
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", "./..."}, args)
	})

	t.Run("extended cfg flags", func(t *testing.T) {
		args, err := fmtTestArgs(rootCmd, &Config{NoCover: false})
		assert.Nil(t, err)
		assert.Equal(t, []string{"go", "test", "-json", "-v", fmt.Sprintf("-coverprofile=%v", coverPath), "./..."}, args)
	})
}

func TestFindPaths(t *testing.T) {
	cases := []struct {
		args  []string
		paths []string
		tests []string
	}{
		{paths: []string{"./..."}},
		{args: []string{"../_testdata"}, paths: []string{"./../_testdata"}},
		{args: []string{"../_testdata"}, paths: []string{"./../_testdata"}},
		{args: []string{"../_testdata/go.go"}, paths: []string{"./../_testdata"}, tests: []string{"TestHelloWorld", "TestGoodbyeWorld"}},
		{args: []string{"../_testdata/go_test.go"}, paths: []string{"./../_testdata"}, tests: []string{"TestHelloWorld", "TestGoodbyeWorld"}},
		{args: []string{"../_testdata/go_test.go:6"}, paths: []string{"./../_testdata"}, tests: []string{"TestHelloWorld"}},
		{args: []string{"../_testdata/go_test.go:TestGoodbyeWorld"}, paths: []string{"./../_testdata"}, tests: []string{"TestGoodbyeWorld"}},
		{args: []string{"TestGoodbyeWorld"}, paths: []string{"./..."}, tests: []string{"TestGoodbyeWorld"}},
		{args: []string{"../_testdata", "TestGoodbyeWorld"}, paths: []string{"./../_testdata"}, tests: []string{"TestGoodbyeWorld"}},
		{args: []string{"../_testdata", "TestHelloWorld", "TestGoodbyeWorld"}, paths: []string{"./../_testdata"}, tests: []string{"TestHelloWorld", "TestGoodbyeWorld"}},
		{args: []string{"../_testdata", "TestFoo"}, paths: []string{"./../_testdata"}, tests: []string{"TestFoo"}},
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
		{path: "../_testdata", line: -1},
		{path: "../_testdata/not_there.go", line: -1},
		{path: "../_testdata/go.go", line: -1, tests: []string{}},
		{path: "../_testdata/go_test.go", line: -1, tests: []string{"TestHelloWorld", "TestGoodbyeWorld"}},
		{path: "../_testdata/go_test.go", line: 2, tests: []string{"TestHelloWorld"}},
		{path: "../_testdata/go_test.go", line: 3, tests: []string{"TestHelloWorld"}},
		{path: "../_testdata/go_test.go", line: 4, tests: []string{"TestHelloWorld"}},
		{path: "../_testdata/go_test.go", line: 5, tests: []string{"TestHelloWorld"}},
		{path: "../_testdata/go_test.go", line: 6, tests: []string{"TestHelloWorld"}},
		{path: "../_testdata/go_test.go", line: 8, tests: []string{"TestHelloWorld"}},
		{path: "../_testdata/go_test.go", line: 9, tests: []string{"TestGoodbyeWorld"}},
		{path: "../_testdata/go_test.go", line: 10, tests: []string{"TestGoodbyeWorld"}},
		{path: "../_testdata/go_test.go", line: 11, tests: []string{"TestGoodbyeWorld"}},
		{path: "../_testdata/go_test.go", line: 12, tests: []string{"TestGoodbyeWorld"}},
	}

	for i, testcase := range cases {
		tests := findTestsInFile(testcase.path, testcase.line)
		assert.Equal(t, testcase.tests, tests, fmt.Sprintf("testcase %v", i))
	}
}
