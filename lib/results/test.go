package results

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type (
	// Test is the results for a single test
	Test struct {
		*stopwatch
		Name     string     `json:"name"`
		State    Action     `json:"state"`
		Package  string     `json:"package"`
		Failures []*Failure `json:"failures,omitempty"`
	}
	// Failure is a single test failure message
	Failure struct {
		Name          string           `json:"failure,omitempty"`
		Package       string           `json:"package,omitempty"`
		File          string           `json:"file,omitempty"`
		Line          int              `json:"line,omitempty"`
		Messages      []string         `json:"messages,omitempty"`
		Diff          *TestifyDiff     `json:"diff,omitempty"`
		IsPanic       bool             `json:"IsPanic,omitempty"`
		PanicTrace    []PanicTraceLine `json:"panic_trace,omitempty"`
		likelyTestify int
	}
	// PanicTraceLine captues a single line in a panic trace that is within the project
	PanicTraceLine struct {
		Fn   string `json:"fn,omitempty"`
		Line int    `json:"line,omitempty"`
		Path string `json:"path,omitempty"`
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
		Type   string `json:"type,omitempty"`
		Value  string `json:"value,omitempty"`
		Length string `json:"length,omitempty"`
	}
)

var (
	testifyTracePattern = regexp.MustCompile(`^\s*Error Trace:\s*(.*_test\.go):(.*)\s*`)
	testifyErrorPattern = regexp.MustCompile(`^\s*Error:\s*`)
	filepathPattern     = regexp.MustCompile(`^\s*(.*_test\.go):(\d+):\s*(.*)`)
	miscMessagePattern  = regexp.MustCompile(`[=-]{3}\s(RUN|FAIL|PASS|SKIP|CONT|PAUSE):?\s*Test.*`)
	testifyTestPattern  = regexp.MustCompile(`^\s*Test:\s*Test[a-zA-Z\/^\s]*`)
)

func newTest(pkgName, testName string) *Test {
	test := &Test{
		stopwatch: &stopwatch{},
		Name:      testName,
		State:     Run,
		Package:   pkgName,
	}
	test.start()
	return test
}

func (test *Test) result(set *Set, pkg *Package, action Action, output string) {
	switch action {
	case Pass:
		pkg.Pass++
		set.TestSummary.Pass++
	case Fail:
		pkg.Fail++
		set.TestSummary.Fail++
		set.FailedTests = append(set.FailedTests, test)
	case Skip:
		pkg.Skip++
		set.TestSummary.Skip++
		set.SkippedTests = append(set.SkippedTests, test)
	case Run, Continue:
		test.start()
	case Pause:
		test.pause()
	case Output:
		test.addLogOutput(output)
	}
	if action != Output {
		test.State = action
	}
	if action == Pass || action == Fail || action == Skip {
		if elapsed := test.stop(); elapsed > set.threshold {
			set.SlowTests = append(set.SlowTests, test)
		}
	}
}

func (test *Test) addLogOutput(msg string) {
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
			rootpath, _ := filepath.Abs("./")
			if strings.HasPrefix(msg, test.Package) {
				fn := strings.Split(strings.TrimPrefix(strings.TrimSpace(msg), test.Package+"."), ".")[0]
				failure.PanicTrace = append(failure.PanicTrace, PanicTraceLine{Fn: fn})
			} else if strings.HasPrefix(strings.TrimSpace(msg), rootpath) {
				parts := strings.Split(strings.ReplaceAll(strings.TrimSpace(msg), rootpath, "."), ":")
				lineNum, _ := strconv.Atoi(strings.Split(parts[1], " ")[0])
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
		} else if strings.HasPrefix(line, "Messages:") {
			diff.Message = strings.TrimSpace(strings.TrimPrefix(line, "Messages:"))
		} else if strings.HasPrefix(line, "Diff:") {
			i += 3
			diff.Range = strings.Trim(fail.Messages[i], "@ ")
			i++
			if !strings.Contains(fail.Messages[i], ":") {
				i++
			}
			for ; i < len(fail.Messages); i++ {
				line := fail.Messages[i]
				if !strings.Contains(line, ":") {
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
