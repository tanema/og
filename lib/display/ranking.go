package display

import (
	"io"

	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

const rankingTemplate = `{{"Ranked Tests:" | bold}}
{{range .}}{{.TimeElapsed}} {{.Package}} {{.Name}}
{{end}}`

func Ranking(w io.Writer, set *results.Set) {
	term.Fprintf(w, rankingTemplate, set.RankedTests())
}
