package display

const (
	DumpTemplate = `{{define "line"}}{{range .Dump}}{{.}}{{end}}{{end}}`

	BuildErrorsTemplate = `{{"Build Errors"| magenta | bold}}:{{range .Set.BuildErrors}}
{{.Package}} {{range .Lines}}
{{if ne .Path ""}}{{.Path | cyan}}:{{.Line | bold}}:{{.Column | bold}}{{end}} {{.Message | magenta}}{{if ne .Have ""}}
    Expected: {{.Want | green}}
    Actual  : {{.Have | red}}{{- end}}{{if not $.Cfg.HideExcerpts}}{{with .Excerpt}}
    {{with .Before}}{{.Line}}  {{.Code | faint}}{{end}}
    {{with .Highlight}}{{.Line}}  {{.Prefix | bold}}{{.Highlight | bold | Red}}{{.Suffix | bold}}{{end}}
    {{with .After}}{{.Line}}  {{.Code | faint}}{{end}}{{end}}{{end}}{{end}}{{end}}`

	FailLineTemplate = `{{define "line"}}{{if ne .File ""}}{{.File | cyan}}:{{.Line |bold}}{{end}} {{if eq (len .Messages) 1 -}}
    {{index .Messages 0}}
{{- else -}}
{{range .Messages}}
    {{.}}
{{- end -}}
{{- end -}}
{{- end}}`

	PanicTemplate = `{{define "panic"}}{{"Panic" | bold | Red}} {{index .Messages 0 | red}}{{range .PanicTrace}}
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
  {{- if and (and (eq (len $failures) 1) (not (index $failures 0).Diff) (not (index $failures 0).IsPanic))}}{{template "line" (index $failures 0)}}
  {{- else}}{{range $failures}}{{if .Diff}}
  {{template "diff" .}}{{else if .IsPanic}}
  {{template "panic" .}}{{else}}
  {{template "line" .}}{{end}}
  {{- end -}}
  {{- end -}}
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
{{- if eq $pkg.State "run" "cont" "pause" -}}{{"RUN " | faint | Cyan | white}}
{{- else if eq $pkg.State "pass" -}}{{"PASS" | Green | white | bold}}
{{- else if eq $pkg.State "fail" -}}{{"FAIL" | Red | white | bold}}
{{- else if eq $pkg.State "skip" -}}{{"NONE" | Blue | white | bold}}
{{- end}} {{$pkgname | bold}} {{if not $.Cfg.HideElapsed}}({{$pkg.TimeElapsed | cyan}}) {{end}}{{with $pkg.CoveragePercent}}{{.}}% {{end}}{{if $pkg.Cached}}{{"(cached)" | green}}{{end}}
{{end}}`

	namesTemplate = `{{range $pkgname, $pkg := (.Set.FilteredPackages (not $.Cfg.HideEmpty)) }}{{- range $testname, $test := ($pkg.FilteredTests (not $.Cfg.HideSkip)) -}}
{{- if eq $test.State "run" "cont" "pause" -}}{{"RUN " | faint | Cyan | white}}
{{- else if eq $test.State "pass" -}}{{"PASS" | Green | white | bold}}
{{- else if eq $test.State "fail" -}}{{"FAIL" | Red | white | bold}}
{{- else if eq $test.State "skip" -}}{{"NONE" | Blue | white | bold}}
{{- end}} {{$pkgname | bold}}=>{{$testname}} {{if not $.Cfg.HideElapsed}}({{$pkg.TimeElapsed | cyan}}){{end}}
{{end}}{{end}}`

	VersionTemplate = `{{.Pad | Red}}
{{.Pad | Yellow}}
{{.OgVersion | black | Green}}
{{.GoVersion | white | Blue}}
{{.Pad | Cyan}}
{{"This tool is actually gay" | magenta | Magenta}}
`
)
