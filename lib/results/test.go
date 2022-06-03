package results

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tanema/og/lib/stopwatch"
)

type (
	// Test is the results for a single test
	Test struct {
		Name        string        `json:"name"`
		State       Action        `json:"state"`
		Package     string        `json:"package"`
		TimeElapsed time.Duration `json:"time_elapsed"`
		Failures    []*Failure    `json:"failures,omitempty"`
		watch       *stopwatch.Stopwatch
	}
	// Failure is a single test failure message
	Failure struct {
		Name          string           `json:"failure"`
		Package       string           `json:"package"`
		File          string           `json:"file"`
		Line          int              `json:"line"`
		Messages      []string         `json:"messages"`
		Diff          *TestifyDiff     `json:"diff,omitempty"`
		IsPanic       bool             `json:"IsPanic"`
		PanicTrace    []PanicTraceLine `json:"panic_trace"`
		likelyTestify int
	}
	// PanicTraceLine captues a single line in a panic trace that is within the project
	PanicTraceLine struct {
		Fn   string `json:"fn"`
		Line int    `json:"line"`
		Path string `json:"path"`
	}
	// TestifyDiff captures testifys output after assert.Equal failts
	TestifyDiff struct {
		Error    string                             `json:"error"`
		Range    string                             `json:"range"`
		Expected string                             `json:"expected"`
		Actual   string                             `json:"actual"`
		Message  string                             `json:"message"`
		Comp     map[string]*TestifyCompStructField `json:"comp"`
	}
	// TestifyCompStructField captures the state of a testify.compStruct
	TestifyCompStructField struct {
		Name     string        `json:"name"`
		Correct  bool          `json:"correct"`
		Val      *TestifyField `json:"val,omitempty"`
		Expected *TestifyField `json:"expected,omitempty"`
		Actual   *TestifyField `json:"actual,omitempty"`
	}
	// TestifyField is one of the fields in a testify.compStruct expected or actual
	TestifyField struct {
		Type   string `json:"type"`
		Value  string `json:"value"`
		Length string `json:"length"`
	}
)

var (
	testifyTracePattern = regexp.MustCompile(`^\s*Error Trace:\s*(.*_test\.go):(.*)\s*`)
	testifyErrorPattern = regexp.MustCompile(`^\s*Error:\s*`)
	filepathPattern     = regexp.MustCompile(`^\s*(.*_test\.go):(\d+):\s*(.*)`)
	miscMessagePattern  = regexp.MustCompile(`[=-]{3}\s(RUN|FAIL|PASS|SKIP|CONT|PAUSE):?\s*Test[a-zA-Z\/^\s]*\s*(\(\d+\.\d+s\))*`)
	testifyTestPattern  = regexp.MustCompile(`^\s*Test:\s*Test[a-zA-Z\/^\s]*`)
)

func newTest(pkgName, testName string) *Test {
	return &Test{
		watch:   stopwatch.Start(),
		Name:    testName,
		State:   Run,
		Package: pkgName,
	}
}

func (test *Test) addLogOutput(mod, root, msg string) {
	if strings.TrimSpace(miscMessagePattern.ReplaceAllString(msg, "")) == "" {
		return
	}

	if strings.HasPrefix(msg, "panic: ") {
		msg = strings.TrimPrefix(msg, "panic: ")
		msg = strings.TrimSpace(msg)
		msg = strings.TrimSuffix(msg, " [recovered]")
		test.Failures = append(test.Failures, &Failure{
			Name:     test.Name,
			Package:  test.Package,
			Messages: []string{msg},
			IsPanic:  true,
		})
	} else if filepathMatches := filepathPattern.FindStringSubmatch(msg); len(filepathMatches) > 0 {
		lineNum, _ := strconv.Atoi(filepathMatches[2])
		msg := []string{}
		if strings.TrimSpace(filepathMatches[3]) != "" {
			msg = append(msg, filepathMatches[3])
		}
		test.Failures = append(test.Failures, &Failure{
			Name:     test.Name,
			Package:  test.Package,
			File:     filepathMatches[1],
			Line:     lineNum,
			Messages: msg,
		})
	} else if len(test.Failures) > 0 {
		failure := test.Failures[len(test.Failures)-1]
		if failure.IsPanic {
			if strings.HasPrefix(msg, mod) {
				fn := strings.Split(strings.TrimPrefix(strings.TrimSpace(msg), mod+"/"), ".")[1]
				failure.PanicTrace = append(failure.PanicTrace, PanicTraceLine{Fn: fn})
			} else if strings.HasPrefix(strings.TrimSpace(msg), root) {
				path := strings.Split(strings.TrimPrefix(strings.TrimSpace(msg), root+"/"), " ")[0]
				parts := strings.Split(path, ":")
				lineNum, _ := strconv.Atoi(parts[1])
				failure.PanicTrace[len(failure.PanicTrace)-1].Path = parts[0]
				failure.PanicTrace[len(failure.PanicTrace)-1].Line = lineNum
				if failure.File == "" {
					failure.File = parts[0]
					failure.Line = lineNum
				}
			}
		} else {
			if testifyTracePattern.MatchString(msg) || testifyTestPattern.MatchString(msg) {
				failure.likelyTestify++
				return
			}
			if testifyErrorPattern.MatchString(msg) {
				failure.likelyTestify++
			}
			failure.Messages = append(failure.Messages, strings.TrimSpace(msg))
		}
	} else {
		test.Failures = append(test.Failures, &Failure{
			Name:     test.Name,
			Package:  test.Package,
			Messages: []string{msg},
		})
	}
}

func (fail *Failure) format() {
	if fail.likelyTestify > 1 {
		fail.formatTestifyDiff()
	}
}

func (fail *Failure) formatTestifyDiff() {
	diff := &TestifyDiff{Comp: map[string]*TestifyCompStructField{}}
	for i := 0; i < len(fail.Messages); i++ {
		line := fail.Messages[i]
		if strings.HasPrefix(line, "Error:") {
			diff.Error = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "Error:"), ":"))
		} else if strings.HasPrefix(line, "expected: ") {
			diff.Expected = strings.TrimPrefix(strings.TrimPrefix(line, "expected: "), "testify.compStruct")
		} else if strings.HasPrefix(line, "actual  : ") {
			diff.Actual = strings.TrimPrefix(strings.TrimPrefix(line, "actual  : "), "testify.compStruct")
		} else if strings.HasPrefix(line, "Diff:") || strings.HasPrefix(line, "--- Expected") || strings.HasPrefix(line, "+++ Actual") {
			continue
		} else if strings.HasPrefix(line, "Messages:") {
			diff.Message = strings.TrimSpace(strings.TrimPrefix(line, "Messages:"))
		} else if strings.HasPrefix(line, "@@") && strings.HasSuffix(line, "@@") {
			diff.Range = strings.Trim(line, "@ ")
		} else if strings.HasPrefix(line, "(testify.compStruct) {") {
			for i++; i < len(fail.Messages); i++ {
				line := fail.Messages[i]
				if line == "}" {
					break
				}
				expected := strings.HasPrefix(line, "- ")
				actual := strings.HasPrefix(line, "+ ")
				parts := strings.Split(strings.Trim(line, "+- "), " ")
				fieldName := strings.TrimSuffix(parts[0], ":")
				if _, ok := diff.Comp[fieldName]; !ok {
					diff.Comp[fieldName] = &TestifyCompStructField{
						Name:    fieldName,
						Correct: !expected && !actual,
					}
				}
				fieldDiff := diff.Comp[fieldName]
				field := &TestifyField{
					Type:  parts[1],
					Value: strings.TrimSuffix(parts[len(parts)-1], ","),
				}
				if len(parts) > 3 {
					if strings.HasPrefix(parts[2], "(len=") {
						field.Length = strings.Trim(parts[2], "(len=)")
					}
				}
				if expected {
					fieldDiff.Expected = field
				} else if actual {
					fieldDiff.Actual = field
				} else {
					fieldDiff.Val = field
				}
			}
		}
	}
	fail.Diff = diff
}

func (test *Test) result(set *Set, pkg *Package, action Action, output string) {
	switch action {
	case Pass:
		pkg.Pass++
		set.TestSummary.Pass++
	case Fail:
		pkg.Fail++
		set.TestSummary.Fail++
	case Skip:
		pkg.Skip++
		set.TestSummary.Skip++
	case Pause:
		test.TimeElapsed = test.watch.Pause()
	case Continue:
		test.watch.Resume()
	case Output:
		test.addLogOutput(set.Mod, set.Root, output)
	}
	if action != Output {
		test.State = action
	}
	if action == Pass || action == Fail || action == Skip {
		test.TimeElapsed = test.watch.Stop()
	}
}
