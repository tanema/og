package results

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type (
	// ExcerptLine are lines around an excerpt
	ExcerptLine struct {
		Line, Code string
	}
	// ExcerptHighlightLine is the target of the excerpt
	ExcerptHighlightLine struct {
		Line, Prefix, Highlight, Suffix string
	}
	// Excerpt is a generated structure that captures where an error happened
	Excerpt struct {
		Before    *ExcerptLine
		Highlight *ExcerptHighlightLine
		After     *ExcerptLine
	}
)

// Excerpt generates the data needed to display a code snippet of where a build
// error occurred
func (line *BuildErrorLine) Excerpt() *Excerpt {
	file, err := os.Open(line.Path)
	if err != nil {
		return nil
	}
	start, end := max(1, line.Line-1), line.Line+1
	digitCount := digits(line.Line + 1)
	scanner := bufio.NewScanner(file)
	curline := 1
	for ; curline < line.Line-1; curline++ {
		scanner.Scan()
	}
	excpt := &Excerpt{}
	if curline < line.Line && scanner.Scan() {
		excpt.Before = &ExcerptLine{Line: leftPad(strconv.Itoa(start), digitCount), Code: strings.ReplaceAll(scanner.Text(), "\t", "  ")}
	}
	if scanner.Scan() {
		text := scanner.Text()
		excpt.Highlight = &ExcerptHighlightLine{
			Line:      leftPad(strconv.Itoa(line.Line), digitCount),
			Prefix:    strings.ReplaceAll(text[:max(0, line.Column-1)], "\t", "  "),
			Highlight: strings.ReplaceAll(text[max(0, line.Column-1):line.Column], "\t", "  "),
			Suffix:    strings.ReplaceAll(text[line.Column:], "\t", "  "),
		}
	} else {
		return nil
	}
	if scanner.Scan() {
		excpt.After = &ExcerptLine{Line: leftPad(strconv.Itoa(end), digitCount), Code: strings.ReplaceAll(scanner.Text(), "\t", "  ")}
	}
	return excpt
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
