package display

import (
	"strings"

	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

const summaryTemplate = `
{{if gt (len .BuildErrors) 0}}{{"Build Errors"| magenta | bold}}:
{{range .BuildErrors}} {{. | magenta}}
{{end -}}{{- end -}}
{{- if gt (len .Failures) 0}}{{"Failed Tests"| red | bold}}:
{{range $pkg, $fails := .Failures }}{{"●" | bold}} {{ $pkg | bold }}
{{range $fails}}  {{.Name | bold}}:
{{range .Messages}}    {{.}}
{{end}}{{- end}}{{end}}{{- end -}}
{{- if gt (len .Skips) 0}}{{- "Skipped Tests"| yellow | bold}}:
{{range $pkg, $skips := .Skips }}{{"●" | bold}} {{ $pkg | bold }}
{{range $skips}}  {{. | yellow}}
{{end -}}
{{- end}}{{- end -}}
{{- "Tests" | bold}}: Pass: {{.TestSummary.Pass | green}} Skip: {{.TestSummary.Skip | blue}} Fail: {{.TestSummary.Fail | red}} Total: {{.TotalTests | bold}}
{{"Packages"| bold}}: Pass: {{.PkgSummary.Pass | green}} NoTests: {{.PkgSummary.Skip | blue}} Fail: {{.PkgSummary.Fail | red}} Cached: {{.Cached | green}} Total: {{len .Packages | bold}}
{{"Elapsed" | bold}}: {{.TimeElapsed}}
`

func formatForSummary(set *results.Set) {
	formatBuildErrors(set)
	formatFailures(set)
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
				if msg == "--- Expected" || msg == "+++ Actual" {
					continue
				} else if strings.HasPrefix(msg, "expected:") {
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
				} else if strings.HasPrefix(msg, "@@") {
					finalMessages = append(finalMessages, term.Sprintf("{{. | cyan}}", msg))
				} else {
					finalMessages = append(finalMessages, term.Sprintf("{{. | red}}", msg))
				}
			}
			set.Failures[pkg][i].Messages = finalMessages
		}
	}
}
