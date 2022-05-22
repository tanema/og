package excerpt

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/tanema/og/lib/term"
)

const linePad = 1

// Excerpt will extract and highlight a part of the code
func Excerpt(r io.Reader, line, col int) []string {
	out := []string{}
	start, end := max(1, line-linePad), line+linePad-1
	digitCount := digits(line + linePad)
	curline := 0
	scanner := bufio.NewScanner(r)
	success := scanner.Scan()
	for success && curline <= end {
		curline++
		if curline >= start {
			text := scanner.Text()
			lineNum := leftPad(strconv.Itoa(curline), digitCount)
			if curline == line {
				out = append(out, term.Sprintf(
					"{{index . 0 | bold}}  {{index . 1 | bold}}{{index . 2 | bold | bgRed}}{{index . 3 | bold}}",
					[]string{lineNum, text[:max(0, col-1)], text[max(0, col-1):col], text[col:]},
				))
			} else {
				out = append(out, term.Sprintf("{{index . 0}}  {{index . 1 | faint}}", []string{lineNum, text}))
			}
		}
		success = scanner.Scan()
	}
	return out
}

func leftPad(str string, desiredLen int) string {
	left := desiredLen - len(str)
	if left <= 0 {
		return str
	}
	return repeat(" ", left) + str
}

func digits(i int) int {
	count := 0
	for i > 0 {
		i = i / 10
		count++
	}
	return count
}

func repeat(str string, rep int) string {
	if rep <= 0 {
		return ""
	}
	return strings.Repeat(str, rep)
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
