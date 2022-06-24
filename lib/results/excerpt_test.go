package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const input = `
`

func TestExcerpt(t *testing.T) {
	err := BuildError{
		Path:   "../../_testdata/go.go",
		Line:   8,
		Column: 2,
	}
	expected := Excerpt{
		Before:    &ExcerptLine{Line: "7", Code: "func subtract(a, b int) int {"},
		Highlight: &ExcerptHighlightLine{Line: "8", Prefix: "  ", Highlight: "r", Suffix: "eturn a - b"},
		After:     &ExcerptLine{Line: "9", Code: "}"},
	}
	actual := err.Excerpt()
	assert.NotNil(t, actual)
	assert.Equal(t, *expected.Before, *actual.Before)
	assert.Equal(t, *expected.Highlight, *actual.Highlight)
	assert.Equal(t, *expected.After, *actual.After)

	err = BuildError{
		Path:   "../../_testdata/go.go",
		Line:   1,
		Column: 1,
	}
	expected = Excerpt{
		Highlight: &ExcerptHighlightLine{Line: "1", Prefix: "", Highlight: "p", Suffix: "ackage main"},
		After:     &ExcerptLine{Line: "2", Code: ""},
	}
	actual = err.Excerpt()
	assert.NotNil(t, actual)
	assert.Nil(t, actual.Before)
	assert.Equal(t, *expected.Highlight, *actual.Highlight)
	assert.Equal(t, *expected.After, *actual.After)

	err = BuildError{
		Path:   "../../_testdata/go.go",
		Line:   17,
		Column: 1,
	}
	expected = Excerpt{
		Before:    &ExcerptLine{Line: "16", Code: "  return a * b"},
		Highlight: &ExcerptHighlightLine{Line: "17", Prefix: "", Highlight: "}", Suffix: ""},
	}
	actual = err.Excerpt()
	assert.NotNil(t, actual)
	assert.Equal(t, *expected.Before, *actual.Before)
	assert.Equal(t, *expected.Highlight, *actual.Highlight)
	assert.Nil(t, actual.After)
}

func TestDigits(t *testing.T) {
	cases := [][2]int64{{1, 1}, {10, 2}, {300, 3}, {4000, 4}}
	for _, cs := range cases {
		assert.Equal(t, int(cs[1]), digits(cs[0]))
	}
}

func TestLeftPad(t *testing.T) {
	assert.Equal(t, " 2", leftPad(2, digits(10)))
	assert.Equal(t, "  2", leftPad(2, digits(100)))
}
