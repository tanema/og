package excerpt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const input = `// This is for the test

a = 1

func test(param)
  print(param)
end

a = 42

exit()
`

func TestExcerpt(t *testing.T) {
	expected := []string{
		"5  \x1b[2mfunc test(param)\x1b[0m",
		"\x1b[1m6\x1b[0m  \x1b[1m \x1b[0m\x1b[41m\x1b[1m \x1b[0m\x1b[0m\x1b[1mprint(param)\x1b[0m",
		"7  \x1b[2mend\x1b[0m",
	}
	assert.Equal(t, expected, Excerpt(strings.NewReader(input), 6, 2))
}

func TestDigits(t *testing.T) {
	cases := [][2]int{{1, 1}, {10, 2}, {300, 3}, {4000, 4}}
	for _, cs := range cases {
		assert.Equal(t, cs[1], digits(cs[0]))
	}

	assert.Equal(t, " 2", leftPad("2", digits(10)))
}
