package display

import (
	"io"

	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

var Decorators = map[string]func(io.Writer) Renderer{
	"dots":  Dot,
	"pdots": PackageDot,
	"name":  Name,
	"pname": PackageName,
}

// Renderer describes a object that can display result data
type Renderer interface {
	Render(set *results.Set, pkg *results.Package, test *results.Test)
}

// ScreenDisplay is the mechanism for displaying dot outputs with a screen buffer
// This is good for updating shorter outputs however outputs that are likely to
// go beyond the screen should print instead
type ScreenDisplay struct {
	sb   *term.ScreenBuf
	tmpl string
}

// PrintDisplay is the mechanism for displaying output line by line rather
// than a display that refreshes
type PrintDisplay struct {
	out         io.Writer
	packageTmpl string
	testTmpl    string
}

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

// Render will render the display of the tests
func (prin *PrintDisplay) Render(set *results.Set, pkg *results.Package, test *results.Test) {
	if pkg != nil && len(pkg.Results) > 0 {
		term.Fprintf(prin.out, prin.packageTmpl, pkg)
	} else if test != nil {
		term.Fprintf(prin.out, prin.testTmpl, test)
	}
}
