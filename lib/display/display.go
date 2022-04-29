package display

import (
	"io"
	"strings"

	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

// Decorators are all of the decorator that are available
var Decorators = map[string]string{
	"dots":   dotsTemplate,
	"pdots":  dotsSeparateTemplate,
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
func New(w io.Writer, cfg *config.Config) *Renderer {
	return &Renderer{
		sb:  term.NewScreenBuf(w),
		cfg: cfg,
	}
}

// Render will render the display of the tests
func (render *Renderer) Render(set *results.Set, pkg *results.Package, test *results.Test) {
	render.sb.Render(render.cfg.ResultsTemplate, renderData{set, render.cfg})
}

// Summary will re-render and add on the summary
func (render *Renderer) Summary(set *results.Set) {
	formatBuildErrors(set)
	formatFailures(set)
	render.sb.Render(render.cfg.ResultsTemplate+render.cfg.SummaryTemplate, renderData{set, render.cfg})
}

func formatBuildErrors(set *results.Set) {
	for i, msg := range set.BuildErrors {
		if strings.Contains(msg, "have (") {
			set.BuildErrors[i] = term.Sprintf("{{. | red}}", msg)
		} else if strings.Contains(msg, "want (") {
			set.BuildErrors[i] = term.Sprintf("{{. | green}}", msg)
		}
	}
}

func formatFailures(set *results.Set) {
	for pkg, fails := range set.Failures {
		for i, fail := range fails {
			finalMessages := []string{}
			for j := 0; j < len(fail.Messages); j++ {
				msg := set.Failures[pkg][i].Messages[j]
				if strings.Contains(msg, "--- Expected") || strings.Contains(msg, "+++ Actual") {
					continue
				} else if strings.Contains(msg, "expected:") || strings.Contains(msg, "Want:") {
					finalMessages = append(finalMessages, term.Sprintf(`{{. | green}}`, msg))
				} else if strings.HasPrefix(msg, "(testify.compStruct) {") {
					finalMessages = append(finalMessages, term.Sprintf(`{{"{" | green}}`, msg))
					j++
					for ; ; j++ {
						msg = set.Failures[pkg][i].Messages[j]
						if strings.HasPrefix(msg, "+ ") {
							msg = strings.TrimPrefix(msg, "+ ")
							finalMessages = append(finalMessages, term.Sprintf("  {{. | red}}", msg))
						} else if msg == "}" {
							finalMessages = append(finalMessages, term.Sprintf("{{. | green}}", msg))
							break
						} else {
							msg = strings.TrimPrefix(msg, "- ")
							finalMessages = append(finalMessages, term.Sprintf("  {{. | green}}", msg))
						}
					}
				} else if strings.Contains(msg, "@@ ") {
					finalMessages = append(finalMessages, term.Sprintf("{{. | cyan}}", msg))
				} else {
					finalMessages = append(finalMessages, term.Sprintf("{{. | red}}", msg))
				}
			}
			set.Failures[pkg][i].Messages = finalMessages
		}
	}
}
