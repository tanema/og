package term

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"text/template"
	"unicode"
	"unicode/utf8"
)

type (
	attribute string
	icon      struct {
		color attribute
		char  string
	}
)

const (
	fGBold      attribute = "1"
	fGFaint     attribute = "2"
	fGItalic    attribute = "3"
	fGUnderline attribute = "4"
	fGBlack     attribute = "30"
	fGRed       attribute = "31"
	fGGreen     attribute = "32"
	fGYellow    attribute = "33"
	fGBlue      attribute = "94"
	fGMagenta   attribute = "35"
	fGCyan      attribute = "36"
	fGWhite     attribute = "97"
	bGBlack     attribute = "40"
	bGRed       attribute = "41"
	bGGreen     attribute = "42"
	bGYellow    attribute = "43"
	bGBlue      attribute = "44"
	bGMagenta   attribute = "45"
	bGCyan      attribute = "46"
	bGWhite     attribute = "47"

	saveCurPos = "\033[s"
	restCurPos = "\033[u\033[J"
)

var funcMap = template.FuncMap{
	"black":     styler(fGBlack),
	"red":       styler(fGRed),
	"green":     styler(fGGreen),
	"yellow":    styler(fGYellow),
	"blue":      styler(fGBlue),
	"magenta":   styler(fGMagenta),
	"cyan":      styler(fGCyan),
	"white":     styler(fGWhite),
	"bgBlack":   styler(bGBlack),
	"bgRed":     styler(bGRed),
	"bgGreen":   styler(bGGreen),
	"bgYellow":  styler(bGYellow),
	"bgBlue":    styler(bGBlue),
	"bgMagenta": styler(bGMagenta),
	"bgCyan":    styler(bGCyan),
	"bgWhite":   styler(bGWhite),
	"bold":      styler(fGBold),
	"faint":     styler(fGFaint),
	"italic":    styler(fGItalic),
	"underline": styler(fGUnderline),
	"iconQ":     iconer(iconInitial),
	"iconGood":  iconer(iconGood),
	"iconWarn":  iconer(iconWarn),
	"iconBad":   iconer(iconBad),
	"iconSel":   iconer(iconSelect),
	"iconChk":   iconer(iconCheckboxCheck),
	"iconBox":   iconer(iconCheckbox),
}

func styler(attr attribute) func(interface{}) string {
	return func(v interface{}) string {
		s, ok := v.(string)
		if ok && s == ">>" {
			return fmt.Sprintf("\033[%sm", attr)
		}
		return fmt.Sprintf("\033[%sm%v\033[0m", attr, v)
	}
}

func iconer(ic icon) func() string {
	return func() string { return styler(ic.color)(ic.char) }
}

// Sprintf formats a string template and outputs console ready text
func Sprintf(in string, data interface{}) string {
	return string(renderStringTemplate(in, data))
}

// Fprintf will print a formatted string out to a writer
func Fprintf(w io.Writer, in string, data interface{}) {
	fmt.Fprint(w, string(renderStringTemplate(in, data)))
}

func renderStringTemplate(in string, data interface{}) []byte {
	tpl, err := template.New("").Funcs(funcMap).Parse(in)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// ScreenBuf is a convenient way to write to terminal screens. It creates,
// clears and, moves up or down lines as needed to write the output to the
// terminal using ANSI escape codes.
type ScreenBuf struct {
	w   io.Writer
	buf *bytes.Buffer
	mut sync.Mutex
}

// NewScreenBuf creates and initializes a new ScreenBuf.
func NewScreenBuf(w io.Writer) *ScreenBuf {
	return &ScreenBuf{buf: &bytes.Buffer{}, w: w}
}

// Reset will empty the buffer and refill it with control characters that will
// clear the previous data on the next flush call.
func (s *ScreenBuf) Reset() {
	s.mut.Lock()
	defer s.mut.Unlock()
	linecount := bytes.Count(s.buf.Bytes(), []byte("\n"))
	s.buf.Reset()
	ClearLines(s.buf, linecount)
}

// Render will write a text/template out to the console, using a mutex so that
// only a single writer at a time can write. This prevents the buffer from losing
// sync with the newlines
func (s *ScreenBuf) Render(in string, data interface{}) {
	s.Reset()
	defer s.Flush()
	s.Write(in, data)
}

// Write will write to the buffer, this will not render to the screen without calling
// Flush. It will also not reset the screen, this is append only. Call reset first.
func (s *ScreenBuf) Write(in string, data interface{}) {
	s.mut.Lock()
	defer s.mut.Unlock()
	tmpl := ansiwrap(renderStringTemplate(in, data), Width())
	if len(tmpl) > 0 && tmpl[len(tmpl)-1] != '\n' {
		tmpl = append(tmpl, '\n')
	}
	s.buf.Write(tmpl)
}

// Flush will flush the render buffer to the screen, this should be called after
// sever calls to Write
func (s *ScreenBuf) Flush() {
	s.mut.Lock()
	defer s.mut.Unlock()
	io.Copy(s.w, bytes.NewBuffer(s.buf.Bytes()))
}

// ansiwrap will wrap a byte array (add linebreak) with awareness of
// ansi character widths
func ansiwrap(str []byte, width int) []byte {
	var output, currentLine []byte
	for _, s := range str {
		currentLine = append(currentLine, s)
		if s == '\n' || runeCount(currentLine) >= width-1 {
			if s != '\n' {
				currentLine = append(currentLine, '\n')
			}
			output = append(output, currentLine...)
			currentLine = []byte{}
		}
	}
	return append(output, currentLine...)
}

// copied from ansiwrap.
// https://github.com/manifoldco/ansiwrap/blob/master/ansiwrap.go#L193
// ansiwrap worked well but I needed a version the preserved
// spacing so I just copied this method over for acurate space counting.
// There is a major problem with this though. It is not able to count
// tab spaces
func runeCount(b []byte) int {
	l, inSequence := 0, false
	for len(b) > 0 {
		if b[0] == '\033' {
			inSequence = true
			b = b[1:]
			continue
		}
		r, rl := utf8.DecodeRune(b)
		b = b[rl:]
		if inSequence {
			if r == 'm' {
				inSequence = false
			}
			continue
		}
		if !unicode.IsPrint(r) {
			continue
		}
		l++
	}
	return l
}
