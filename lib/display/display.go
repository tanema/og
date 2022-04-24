package display

import (
	"io"

	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

// Decorators are all of the decorator that are available
var Decorators = map[string]func(io.Writer) Renderer{
	"dots":   Dot,
	"pdots":  PackageDot,
	"names":  Name,
	"pnames": PackageName,
}

type (
	// Renderer describes a object that can display result data
	Renderer interface {
		Render(set *results.Set, pkg *results.Package, test *results.Test)
		Summary(set *results.Set)
	}
	// ScreenDisplay is the mechanism for displaying dot outputs with a screen buffer
	// This is good for updating shorter outputs however outputs that are likely to
	// go beyond the screen should print instead
	ScreenDisplay struct {
		out  io.Writer
		sb   *term.ScreenBuf
		tmpl string
	}
	// PrintDisplay is the mechanism for displaying output line by line rather
	// than a display that refreshes
	PrintDisplay struct {
		out         io.Writer
		packageTmpl string
		testTmpl    string
	}
)

// Dot will create a dot display for all tests as 1 line
func Dot(w io.Writer) Renderer {
	return &ScreenDisplay{
		sb:   term.NewScreenBuf(w),
		tmpl: dotsTemplate,
	}
}

// PackageDot will display as dots with each of the packages
func PackageDot(w io.Writer) Renderer {
	return &ScreenDisplay{
		out:  w,
		sb:   term.NewScreenBuf(w),
		tmpl: dotsSeparateTemplate,
	}
}

// Name will display package names
func Name(w io.Writer) Renderer {
	return &PrintDisplay{
		out:         w,
		packageTmpl: namesSinglePackageTemplate,
		testTmpl:    namesSingleTestTemplate,
	}
}

// PackageName will display package names
func PackageName(w io.Writer) Renderer {
	return &ScreenDisplay{
		sb:   term.NewScreenBuf(w),
		tmpl: namesPackageTemplate,
	}
}

// Render will render the display of the tests
func (screen *ScreenDisplay) Render(set *results.Set, pkg *results.Package, test *results.Test) {
	screen.sb.Render(screen.tmpl, set)
}

// Summary will re-render and add on the summary
func (screen *ScreenDisplay) Summary(set *results.Set) {
	formatForSummary(set)
	screen.sb.Render(screen.tmpl+summaryTemplate, set)
}

// Render will render the display of the tests
func (prin *PrintDisplay) Render(set *results.Set, pkg *results.Package, test *results.Test) {
	if pkg != nil {
		term.Fprintf(prin.out, prin.packageTmpl, pkg)
	} else if test != nil {
		term.Fprintf(prin.out, prin.testTmpl, test)
	}
}

// Summary will add on the summary
func (prin *PrintDisplay) Summary(set *results.Set) {
	formatForSummary(set)
	term.Fprintf(prin.out, summaryTemplate, set)
}
