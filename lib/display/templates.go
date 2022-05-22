package display

const (
	BuildErrorsTemplate = `{{"Build Errors"| magenta | bold}}:{{range .Set.BuildErrors}}
{{.Package}} {{range .Lines}}
  {{.Path | cyan}}:{{.Line | bold}}:{{.Column | bold}} {{.Message | magenta}}{{if ne .Have ""}}
    Expected: {{.Want | green}}
    Actual  : {{.Have | red}}{{- end}}{{if and (gt (len .Excerpt) 0) (not $.Cfg.HideExcerpts)}}{{range .Excerpt}}
    {{.}}{{end}}{{end}}
{{- end}}{{end}}`

	FailLineTemplate = `{{define "line"}}{{.File | cyan}}:{{.Line |bold}} {{range .Messages}}{{.}}{{end}}{{end}}`

	PanicTemplate = `{{define "panic"}}{{"Panic" | bold | bgRed}} {{index .Messages 0 | red}}{{range .PanicTrace}}
      {{.Path | cyan}}:{{.Line | bold}}:{{.Fn}}{{end}}{{end}}`

	TestifyDiffTemplate = `{{define "diff"}}{{.File | cyan}}:{{.Line |bold}} {{.Diff.Error | red}}
		{{- if .Diff.Message}} "{{.Diff.Message | bold}}"{{ end}}
    {{- if ne .Diff.Expected ""}}
    Expected: {{.Diff.Expected | green}}
    Actual  : {{.Diff.Actual | red}}
    {{- end}}{{if gt (len .Diff.Comp) 0}}
    Diff: {{.Diff.Range | cyan}}
    {
    {{- range $fieldName, $field := .Diff.Comp}}{{if .Correct}}
      {{.Name | green | faint}}: {{.Val.Value | green | faint}}
    {{- else}}
      {{.Name | bold}}: {{.Expected.Value | green}} {{"!=" | bold}} {{.Actual.Value | red}}
    {{- end}}{{end}}
    }
    {{- end -}}
  {{- end}}`

	TestFailuresTemplate = `{{"Failed Tests"| red | bold}}: {{range $pkgName, $tests := .Set.Failures }}
{{- range $tstName, $failures := $tests}}
{{$pkgName}}#{{$tstName}}:
  {{- if and (and (eq (len $failures) 1) (not (index $failures 0).Diff) (not (index $failures 0).IsPanic))}} {{template "line" (index $failures 0)}}
  {{- else}}{{range $failures}}{{if .Diff}}
  {{template "diff" .}}{{else if .IsPanic}}
  {{template "panic" .}}{{else}}
  {{template "line" .}}{{end}}
  {{- end -}}{{end}}
  {{- end -}}
{{- end}}`

	TestSkipTemplate = `{{"Skipped Tests"| yellow | bold}}:
{{range $pkg, $skips := .Set.Skips }}
{{- range $skips}}  {{ $pkg | yellow }}#{{. | yellow}}
{{end}}{{end}}`

	SummaryTemplate = `{{"Tests("| bold}}{{.Set.TotalTests | bold}}{{"): " | bold}}
{{- "Pass: " | green}}{{.Set.TestSummary.Pass | green | bold}}
{{- if not .Cfg.HideSkip}}{{" Skip: " | blue}}{{.Set.TestSummary.Skip | blue | bold}}{{end}}
{{- " Fail: " | red}}{{.Set.TestSummary.Fail | red | bold}}
{{"Packages("| bold}}{{len .Set.Packages | bold}}{{"): " | bold}}
{{- "Pass: " | green}}{{.Set.PkgSummary.Pass | green | bold}}
{{- if not .Cfg.HideEmpty}}{{" NoTests: " | blue}}{{.Set.PkgSummary.Skip | blue | bold}}{{end}}
{{- " Fail: " | red}}{{.Set.PkgSummary.Fail | red | bold}}
{{- " Cached: " | green}}{{.Set.Cached | green | bold}}`

	TestRankTemplate = `{{"Slow Tests:" | bold}}
{{range (.Set.RankedTests .Cfg.Threshold)}}{{.TimeElapsed | cyan}} {{.Package}} {{.Name}}
{{end}}`

	dotsTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty)) -}}
	{{- range $testname, $test := ($pkg.FilteredTests (not $.Cfg.HideSkip)) -}}
		{{- if eq $test.State "run" "cont" "pause" -}}
		{{"%v" | faint | cyan}}
		{{- else if eq $test.State "pass" -}}
		{{"%v" | green}}
		{{- else if eq $test.State "fail" -}}
		{{"%v" | red}}
		{{- else if eq $test.State "skip" -}}
		{{"%v" | yellow}}
		{{- end -}}
	{{- end -}}
{{- end}}`

	dotsSeparateTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty))}}{{"●" | bold}} {{ $pkgname | bold }} {{range $tstname, $test := ($pkg.FilteredTests (not $.Cfg.HideSkip)) -}}
	{{- if eq $test.State "run" "cont" "pause" -}}
	{{"%v" | faint | cyan}}
	{{- else if eq $test.State "pass" -}}
	{{"%v" | green}}
	{{- else if eq $test.State "fail" -}}
	{{"%v" | red}}
	{{- else if eq $test.State "skip" -}}
	{{"%v" | yellow | bold}}
	{{- end -}}
{{end}} {{if not $.Cfg.HideElapsed}}({{$pkg.TimeElapsed | cyan}}) {{end}}{{with $pkg.CoveragePercent}}{{.}}%% {{end}}{{if $pkg.Cached}}{{"(cached)" | green}}{{end}}
{{end}}`

	namesPackageTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty)) }}
{{- if eq $pkg.State "run" "cont" "pause" -}}{{"RUN " | faint | bgCyan | white}}
{{- else if eq $pkg.State "pass" -}}{{"PASS" | bgGreen | white}}
{{- else if eq $pkg.State "fail" -}}{{"FAIL" | bgRed | white}}
{{- else if eq $pkg.State "skip" -}}{{"NONE" | bgBlue | white}}
{{- end}} {{$pkgname | bold}} {{if not $.Cfg.HideElapsed}}({{$pkg.TimeElapsed | cyan}}) {{end}}{{with $pkg.CoveragePercent}}{{.}}% {{end}}{{if $pkg.Cached}}{{"(cached)" | green}}{{end}}
{{end}}`

	namesTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty)) }}{{- range $testname, $test := ($pkg.FilteredTests (not $.Cfg.HideSkip)) -}}
{{- if eq $test.State "run" "cont" "pause" -}}{{"RUN " | faint | bgCyan | white}}
{{- else if eq $test.State "pass" -}}{{"PASS" | bgGreen | white}}
{{- else if eq $test.State "fail" -}}{{"FAIL" | bgRed | white}}
{{- else if eq $test.State "skip" -}}{{"NONE" | bgBlue | white}}
{{- end}} {{$pkgname | bold}}=>{{$testname}} {{if not $.Cfg.HideElapsed}}({{$pkg.TimeElapsed | cyan}}){{end}}
{{end}}{{end}}`
)
