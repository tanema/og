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

	bold             = "1"
	faint            = "2"
	italic           = "3"
	underline        = "4"
	invert           = "7"
	fg               = "3"
	bg               = "4"
	brfg             = "9"
	brbg             = "10"
	black            = "0"
	red              = "1"
	green            = "2"
	yellow           = "3"
	blue             = "4"
	magenta          = "5"
	cyan             = "6"
	white            = "7"
	esc              = "\033["
	rgbfgcolor       = esc + fg + "8;5;%vm"
	rgbbgcolor       = esc + bg + "8;5;%vm"
	truefgcolor      = esc + fg + "8;2;%v;%v;%vm"
	truebgcolor      = esc + bg + "8;2;%v;%v;%vm"
	cursorToStart    = esc + "0G"
	cursorUpOne      = esc + "1A"
	clearToEndOfLine = esc + "0K"
	clearLastLine    = cursorToStart + cursorUpOne + clearToEndOfLine
)

var rainbowColors = []string{brfg + red, brfg + yellow, brfg + green, brfg + blue, brfg + cyan, brfg + magenta}
var ansiPat = regexp.MustCompile(`\033\[(\d*[mKAG])?`)
var funcMap = template.FuncMap{
	"rainbow":   rainbow,
	"Rainbow":   bgRainbow,
	"bright":    bright,
	"Bright":    bgBright,
	"bold":      styler(bold),
	"faint":     styler(faint),
	"italic":    styler(italic),
	"underline": styler(underline),
	"invert":    styler(invert),
	"black":     styler(fg + black),
	"red":       styler(fg + red),
	"green":     styler(fg + green),
	"yellow":    styler(fg + yellow),
	"blue":      styler(fg + blue),
	"magenta":   styler(fg + magenta),
	"cyan":      styler(fg + cyan),
	"white":     styler(fg + white),
	"Black":     styler(bg + black),
	"Red":       styler(bg + red),
	"Green":     styler(bg + green),
	"Yellow":    styler(bg + yellow),
	"Blue":      styler(bg + blue),
	"Magenta":   styler(bg + magenta),
	"Cyan":      styler(bg + cyan),
	"White":     styler(bg + white),
}

type ansiStr struct {
	str  string
	vals []string
}

func parseAnsiString(str string) ansiStr {
	ansipat := regexp.MustCompile(`^\\033\[(\d*;?)+m`)
	points := []string{}
	if ansipat.MatchString(str) {
		ansiCmd := ansipat.FindString(str)
		str = strings.TrimRight(ansipat.ReplaceAllString(str, ""), "\033[m")
		points = append(points, strings.Split(strings.TrimRight(strings.TrimLeft(ansiCmd, `\033[`), "m"), ";")...)
	}
	return ansiStr{str: str, vals: points}
}

func (ansi *ansiStr) add(a string) {
	ansi.vals = append(ansi.vals, a)
}

func (ansi *ansiStr) replace(before, after string) {
	for i, val := range ansi.vals {
		if strings.HasPrefix(val, before) && len(val) > len(before) {
			ansi.vals[i] = strings.Replace(val, before, after, 1)
			break
		}
	}
}

func (ansi ansiStr) String() string {
	return fmt.Sprintf("\033[%vm%v\033[m", strings.Join(ansi.vals, ";"), ansi.str)
}

func styler(attr string) func(interface{}) string {
	return func(v interface{}) string {
		ansistr := parseAnsiString(fmt.Sprintf("%v", v))
		ansistr.add(attr)
		return ansistr.String()
	}
}

func bright(v interface{}) string {
	ansistr := parseAnsiString(fmt.Sprintf("%v", v))
	ansistr.replace(fg, brfg)
	return ansistr.String()
}

func bgBright(v interface{}) string {
	ansistr := parseAnsiString(fmt.Sprintf("%v", v))
	ansistr.replace(bg, brbg)
	return ansistr.String()
}

func rainbow(v string) string {
	chunks := make([]string, len(v))
	for i, chr := range v {
		chunks[i] = fmt.Sprintf("\033[%vm%c\033[m", rainbowColors[i%len(rainbowColors)], chr)
	}
	return strings.Join(chunks, "")
}

func bgRainbow(v string) string {
	chunks := make([]string, len(v))
	for i, chr := range v {
		chunks[i] = fmt.Sprintf("\033[%v;7m%c\033[m", rainbowColors[i%len(rainbowColors)], chr)
	}
	return strings.Join(chunks, "")
}

func Size() (width, height int) {
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
	width, _ := Size()
	tmpl := wrap(renderStringTemplate(in, data), width)
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
