package display

const (
	SummaryTemplate = `{{if not .Cfg.HideSummary}}
{{if gt (len .Set.BuildErrors) 0}}{{"Build Errors"| magenta | bold}}:
{{range .Set.BuildErrors}} {{. | magenta}}
{{end -}}{{- end -}}
{{- if gt (len .Set.Failures) 0}}{{"Failed Tests"| red | bold}}:
{{range $pkg, $fails := .Set.Failures }}{{"●" | bold}} {{ $pkg | bold }}
{{range $fails}}  {{.Name | bold}}:
{{range .Messages}}    {{.}}
{{end}}{{- end}}{{end}}{{- end -}}
{{- if gt (len .Set.Skips) 0}}{{- "Skipped Tests"| yellow | bold}}:
{{range $pkg, $skips := .Set.Skips }}{{"●" | bold}} {{ $pkg | bold }}
{{range $skips}}  {{. | yellow}}
{{end -}}
{{- end}}{{- end -}}
{{- "Tests" | bold}}: Pass: {{.Set.TestSummary.Pass | green}}{{if not .Cfg.HideSkip}} Skip: {{.Set.TestSummary.Skip | blue}}{{end}} Fail: {{.Set.TestSummary.Fail | red}} Total: {{.Set.TotalTests | bold}}
{{"Packages"| bold}}: Pass: {{.Set.PkgSummary.Pass | green}}{{if not .Cfg.HideEmpty}} NoTests: {{.Set.PkgSummary.Skip | blue}}{{end}} Fail: {{.Set.PkgSummary.Fail | red}} Cached: {{.Set.Cached | green}} Total: {{len .Set.Packages | bold}}{{end}}{{if not .Cfg.HideElapsed}}
{{"Elapsed" | bold}}: {{.Set.TimeElapsed | cyan}}{{end}}
{{if gt .Cfg.Rank 0}}{{"Ranked Tests:" | bold}}
{{range (.Set.RankedTests .Cfg.Rank)}}{{.TimeElapsed | cyan}} {{.Package}} {{.Name}}
{{end}}{{end}}`

	dotsTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty)) -}}
	{{- range $testname, $test := ($pkg.FilteredTests (not $.Cfg.HideSkip)) -}}
		{{- if eq $test.State "run" "cont" "pause" -}}
		{{"●" | faint | cyan}}
		{{- else if eq $test.State "pass" -}}
		{{"●" | green}}
		{{- else if eq $test.State "fail" -}}
		{{"●" | red}}
		{{- else if eq $test.State "skip" -}}
		{{"●" | yellow}}
		{{- end -}}
	{{- end -}}
{{- end}}`

	dotsSeparateTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty))}}{{"●" | bold}} {{ $pkgname | bold }} {{range $tstname, $test := ($pkg.FilteredTests (not $.Cfg.HideSkip)) -}}
	{{- if eq $test.State "run" "cont" "pause" -}}
	{{"●" | faint | cyan}}
	{{- else if eq $test.State "pass" -}}
	{{"●" | green}}
	{{- else if eq $test.State "fail" -}}
	{{"●" | red}}
	{{- else if eq $test.State "skip" -}}
	{{"●" | yellow | bold}}
	{{- end -}}
{{end}} {{if not $.Cfg.HideElapsed}}({{$pkg.TimeElapsed | cyan}}){{end}}{{if $pkg.Cached}}{{"(cached)" | green}}{{end}}
{{end}}`

	namesPackageTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty)) }}
{{- if eq $pkg.State "run" "cont" "pause" -}}{{"RUN " | faint | bgCyan | white}}
{{- else if eq $pkg.State "pass" -}}{{"PASS" | bgGreen | white}}
{{- else if eq $pkg.State "fail" -}}{{"FAIL" | bgRed | white}}
{{- else if eq $pkg.State "skip" -}}{{"NONE" | bgBlue | white}}
{{- end}} {{$pkgname | bold}} {{if not $.Cfg.HideElapsed}}({{$pkg.TimeElapsed | cyan}}){{end}} {{if $pkg.Cached}}{{"(cached)" | green}}{{end}}
{{end}}`

	namesTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty)) }}{{- range $testname, $test := ($pkg.FilteredTests (not $.Cfg.HideSkip)) -}}
{{- if eq $test.State "run" "cont" "pause" -}}{{"RUN " | faint | bgCyan | white}}
{{- else if eq $test.State "pass" -}}{{"PASS" | bgGreen | white}}
{{- else if eq $test.State "fail" -}}{{"FAIL" | bgRed | white}}
{{- else if eq $test.State "skip" -}}{{"NONE" | bgBlue | white}}
{{- end}} {{$pkgname | bold}}=>{{$testname}} {{if not $.Cfg.HideElapsed}}({{$pkg.TimeElapsed | cyan}}){{end}}
{{end}}{{end}}`
)
