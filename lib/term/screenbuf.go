//go:build !windows
// +build !windows

package term

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"unicode/utf8"
)

const (
	defaultTermWidth  = 80
	defaultTermHeight = 60

	esc              = "\033["
	color            = esc + "%vm"
	clearColor       = esc + "0m"
	cursorToStart    = esc + "0G"
	cursorUpOne      = esc + "1A"
	clearToEndOfLine = esc + "0K"
	clearLastLine    = cursorToStart + cursorUpOne + clearToEndOfLine
)

var ansiPat = regexp.MustCompile(`\033(\[\d+[mKAG])?`)

var funcMap = template.FuncMap{
	"rainbow":   rainbow,
	"bold":      styler("1"),
	"faint":     styler("2"),
	"italic":    styler("3"),
	"underline": styler("4"),
	"black":     styler("30"),
	"red":       styler("31"),
	"green":     styler("32"),
	"yellow":    styler("33"),
	"magenta":   styler("35"),
	"cyan":      styler("36"),
	"bgBlack":   styler("40"),
	"bgRed":     styler("41"),
	"bgGreen":   styler("42"),
	"bgYellow":  styler("43"),
	"bgBlue":    styler("44"),
	"bgMagenta": styler("45"),
	"bgCyan":    styler("46"),
	"bgWhite":   styler("47"),
	"blue":      styler("94"),
	"white":     styler("97"),
}

func styler(attr string) func(interface{}) string {
	return func(v interface{}) string {
		s, ok := v.(string)
		if ok && s == ">>" {
			return fmt.Sprintf(color, attr)
		}
		return fmt.Sprintf(color+"%v"+clearColor, attr, v)
	}
}

func rainbow(v interface{}) string {
	s := v.(string)
	chunks := make([]string, len(s))
	colors := []string{"31", "33", "32", "94", "36", "35"}
	for i, chr := range s {
		attr := colors[i%len(colors)]
		chunks[i] = fmt.Sprintf(color+"%c"+clearColor, attr, chr)
	}
	return strings.Join(chunks, "")
}

// Width returns the column width of the terminal
func Width() int {
	w, _ := size()
	return w
}

// Height returns the row size of the terminal
func Height() int {
	_, h := size()
	return h
}

func size() (width, height int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return defaultTermWidth, defaultTermHeight
	}
	parts := strings.Split(strings.TrimRight(string(out), "\n"), " ")
	height, err = strconv.Atoi(parts[0])
	if err != nil {
		return defaultTermWidth, defaultTermHeight
	}
	width, err = strconv.Atoi(parts[1])
	if err != nil {
		return defaultTermWidth, defaultTermHeight
	}
	return width, height
}

// Sprintf formats a string template and outputs console ready text
func Sprintf(in string, data interface{}) string {
	return string(renderStringTemplate(in, data))
}

// Fprintf will print a formatted string out to a writer
func Fprintf(w io.Writer, in string, data interface{}) {
	fmt.Fprint(w, string(renderStringTemplate(in, data)))
}

// ClearLines will move the cursor up and clear the line out for re-rendering
func ClearLines(out io.Writer, linecount int) {
	out.Write([]byte(strings.Repeat(clearLastLine, linecount)))
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
	tmpl := wrap(renderStringTemplate(in, data), Width())
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

func wrap(str []byte, width int) []byte {
	var output, currentLine []byte
	for _, s := range str {
		currentLine = append(currentLine, s)
		runes := utf8.RuneCount(ansiPat.ReplaceAll(currentLine, []byte("")))
		if s == '\n' || runes >= width-1 {
			if s != '\n' {
				currentLine = append(currentLine, '\n')
			}
			output = append(output, currentLine...)
			currentLine = []byte{}
		}
	}
	return append(output, currentLine...)
}
