package display

const (
	dotsTemplate = `{{$pkgs := .Packages}}
{{- range .Names -}}
	{{$pkg := index $pkgs .}}
	{{- range $pkg.Names -}}
		{{$test := index $pkg.Results .}}
		{{- if eq $test.State "run" "cont" -}}
		{{"●" | faint | cyan}}
		{{- else if eq $test.State "pass" -}}
		{{"●" | green}}
		{{- else if eq $test.State "fail" -}}
		{{"●" | red}}
		{{- else if eq $test.State "skip" -}}
		{{"●" | blue}}
		{{- else -}}
		{{"●" | faint}}
		{{- end -}}
	{{- end -}}
{{- end}}`

	dotsSeparateTemplate = `{{range $name, $pkg := .Packages }}{{"●" | bold}} {{ $name | bold }} {{range $_name, $test := $pkg.Results -}}
	{{- if eq $test.State "run" "cont" -}}
	{{"●" | faint | cyan}}
	{{- else if eq $test.State "pass" -}}
	{{"●" | green}}
	{{- else if eq $test.State "fail" -}}
	{{"●" | red}}
	{{- else if eq $test.State "skip" -}}
	{{"●" | blue}}
	{{- else -}}
	{{"●" | faint}}
	{{- end -}}
{{end}} {{if $pkg.Cached}}(cached){{end}}
{{end -}}`

	namesPackageTemplate = `{{range $name, $pkg := .Packages }}
{{- if eq $pkg.State "run" "cont" -}}{{"RUN " | faint | bgCyan | white}}
{{- else if eq $pkg.State "pass" -}}{{"PASS" | bgGreen | white}}
{{- else if eq $pkg.State "fail" -}}{{"FAIL" | bgRed | white}}
{{- else if eq $pkg.State "skip" -}}{{"NONE" | bgBlue | white}}
{{- else -}}{{"PAUS" | faint | white}}
{{- end}} {{$name | bold}} ({{$pkg.TimeElapsed}}) {{if $pkg.Cached}}(cached){{end}}
{{end -}}`

	namesSingleTestTemplate = `  {{ if eq .State "pass" -}}{{"PASS" | bgGreen | white}}
{{- else if eq .State "fail" -}}{{"FAIL" | bgRed | white}}
{{- else if eq .State "skip" -}}{{"SKIP" | bgBlue | white}}
{{- end}} {{.Package}} {{.Name | bold}} ({{.TimeElapsed}})
`

	namesSinglePackageTemplate = `=={{ if eq .State "pass" -}}{{"PASS" | bgGreen | white}}
{{- else if eq .State "fail" -}}{{"FAIL" | bgRed | white}}
{{- else if eq .State "skip" -}}{{"NONE" | bgBlue | white}}
{{- end}} {{.Name | bold}} ({{.TimeElapsed}})
`
)
