package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const input = `
`

func TestExcerpt(t *testing.T) {
	err := BuildErrorLine{
		Path:   "./testdata/code",
		Line:   6,
		Column: 3,
	}
	excp := err.Excerpt()
	assert.NotNil(t, excp)
	assert.Equal(t, excp.Before.Line, "5")
	assert.Equal(t, excp.Highlight.Line, "6")
	assert.Equal(t, excp.After.Line, "7")
	assert.Equal(t, excp.Before.Code, "func test(param)")
	assert.Equal(t, excp.Highlight.Prefix, "  ")
	assert.Equal(t, excp.Highlight.Highlight, "p")
	assert.Equal(t, excp.Highlight.Suffix, "rint(param)")
	assert.Equal(t, excp.After.Code, "end")

	err = BuildErrorLine{
		Path:   "./testdata/code",
		Line:   1,
		Column: 4,
	}
	excp = err.Excerpt()
	assert.NotNil(t, excp)
	assert.Nil(t, excp.Before)
	assert.Equal(t, excp.Highlight.Line, "1")
	assert.Equal(t, excp.After.Line, "2")
	assert.Equal(t, excp.Highlight.Prefix, "// ")
	assert.Equal(t, excp.Highlight.Highlight, "T")
	assert.Equal(t, excp.Highlight.Suffix, "his is for the test")
	assert.Equal(t, excp.After.Code, "")

	err = BuildErrorLine{
		Path:   "./testdata/code",
		Line:   11,
		Column: 1,
	}
	excp = err.Excerpt()
	assert.NotNil(t, excp)
	assert.Nil(t, excp.After)
	assert.Equal(t, excp.Before.Line, "10")
	assert.Equal(t, excp.Highlight.Line, "11")
	assert.Equal(t, excp.Before.Code, "")
	assert.Equal(t, excp.Highlight.Prefix, "")
	assert.Equal(t, excp.Highlight.Highlight, "e")
	assert.Equal(t, excp.Highlight.Suffix, "xit()")
}

func TestDigits(t *testing.T) {
	cases := [][2]int{{1, 1}, {10, 2}, {300, 3}, {4000, 4}}
	for _, cs := range cases {
		assert.Equal(t, cs[1], digits(cs[0]))
	}
}

func TestLeftPad(t *testing.T) {
	assert.Equal(t, " 2", leftPad("2", digits(10)))
	assert.Equal(t, "  2", leftPad("2", digits(100)))
}
