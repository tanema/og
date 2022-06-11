package display

import (
	"fmt"

	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

// Decorators are all of the decorator that are available
var Decorators = map[string]string{
	"dots":   fmt.Sprintf(dotsTemplate, "●", "●", "●", "●"),
	"icons":  fmt.Sprintf(dotsTemplate, "?", "✔", "✗", "⚠"),
	"pdots":  fmt.Sprintf(dotsSeparateTemplate, "●", "●", "●", "●"),
	"picons": fmt.Sprintf(dotsSeparateTemplate, "?", "✔", "✗", "⚠"),
	"names":  namesTemplate,
	"pnames": namesPackageTemplate,
}

type (
	// Renderer is the struct that does the display
	Renderer struct {
		sb  *term.ScreenBuf
		cfg *config.Config
	}
	renderData struct {
		Set *results.Set
		Cfg *config.Config
	}
)

// New will create a new display
func New(cfg *config.Config) *Renderer {
	return &Renderer{
		sb:  term.NewScreenBuf(cfg.Out),
		cfg: cfg,
	}
}

// Render will render the display of the tests
func (render *Renderer) Render(set *results.Set) {
	if !set.Any() {
		return
	}
	render.sb.Render(render.cfg.ResultsTemplate, renderData{set, render.cfg})
}

// Summary will re-render and add on the summary
func (render *Renderer) Summary(set *results.Set) {
	data := renderData{set, render.cfg}
	render.sb.Reset()
	defer render.sb.Flush()
	if set.Any() {
		render.sb.Write(render.cfg.ResultsTemplate, data)
	}
	if len(set.BuildErrors) > 0 {
		render.sb.Write(BuildErrorsTemplate, data)
	}
	if !set.Any() {
		render.sb.Write(`{{"No Tests"| bold | Blue}} {{.Mod | bold}}/({{len .Packages | cyan}})`, set)
		return
	}
	if len(set.Failures()) > 0 {
		render.sb.Write(FailLineTemplate+PanicTemplate+TestifyDiffTemplate+TestFailuresTemplate, data)
	}
	if !render.cfg.HideSkip && len(set.Skips()) > 0 {
		render.sb.Write(TestSkipTemplate, data)
	}
	if !render.cfg.HideSummary {
		render.sb.Write(SummaryTemplate, data)
	}
	if render.cfg.Cover != "" {
		render.sb.Write(`{{"Coverage" | bold}}: {{.Set.CoveragePercent}}%`, data)
	}
	if !render.cfg.HideElapsed {
		render.sb.Write(`{{"Elapsed" | bold}}: {{.Set.TimeElapsed | cyan | bold}}`, data)
	}
	if render.cfg.Threshold > 0 && len(set.RankedTests(render.cfg.Threshold)) > 0 {
		render.sb.Write(TestRankTemplate, data)
	}
}
