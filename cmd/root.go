package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/display"
	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

const (
	major = 0
	minor = 0
)

var (
	cfg             = &config.Config{}
	cmdMut          sync.Mutex
	version         bool
	testFuncPattern = regexp.MustCompile(`func (Test.*)\(t \*testing\.T\)`)
)

var rootCmd = &cobra.Command{
	Use:   "og [path[:[lineNum|TestName]]|TestName]",
	Short: "Run go test but make it colorful",
	Long: `Go's test output can sometimes be quit hard to parse, and harder to scan.
A common solution to this is syntax highlighting. It makes it easy to scan
and notice what exactly is wrong at a glance. og test does this.

The easy autocomplete works as follows.
    - og                           => go test ./...
    - og foldername                => go test ./foldername
    - og folder/file_test.go       => go test -run TestsInFile ./folder
    - og folder/file_test.go:20    => go test -run TestAtLine20 ./folder
    - og TestA                     => go test -run TestA ./...
    - og folder/file_test.go:TestA => go test -run TestA ./folder

Any further go flags can be passed with a -- suffix

    og -- -vet=atomic
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(cfg.Load())
		cobra.CheckErr(validateDisplay(cfg))
	},
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			printVersion(cfg)
			return
		}
		testargs, err := fmtTestArgs(cfg, args...)
		cobra.CheckErr(err)
		cobra.CheckErr(runCmd(cfg, testargs...))
		if cfg.Watch {
			cobra.CheckErr(watchTestChanges(cfg, args))
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&cfg.Dump, "dump", "D", false, "dumps the final state in json for usage")
	rootCmd.Flags().StringVarP(&cfg.Display, "display", "d", "dots", "change the display of the test outputs [dots, pdots, names, pnames]")
	rootCmd.Flags().BoolVarP(&cfg.Watch, "watch", "w", false, "watch for file changes and re-run tests")
	rootCmd.Flags().BoolVarP(&cfg.HideSkip, "hideskip", "s", false, "hide info about skipped tests")
	rootCmd.Flags().BoolVarP(&cfg.HideExcerpts, "hideexcerpts", "x", false, "hide code excerpts in build errors")
	rootCmd.Flags().BoolVarP(&cfg.HideEmpty, "hideempty", "p", false, "hide info about packages without tests")
	rootCmd.Flags().BoolVarP(&cfg.HideElapsed, "hideelapse", "e", false, "hide the elapsed time output")
	rootCmd.Flags().BoolVarP(&cfg.HideSummary, "hidesummary", "m", false, "hide the complete summary output")
	rootCmd.Flags().DurationVarP(&cfg.Threshold, "threshold", "r", 5*time.Second, "output lists of tests slower than the threshold. 0 will disable")
	rootCmd.Flags().BoolVar(&cfg.Short, "short", false, "run short tests")
	rootCmd.Flags().BoolVar(&cfg.NoCache, "nocache", false, "disable go test cache")
	rootCmd.Flags().BoolVar(&cfg.FailFast, "failfast", false, "terminate after first test failure")
	rootCmd.Flags().BoolVar(&cfg.Shuffle, "shuffle", false, "shuffle test order")
	rootCmd.Flags().StringVar(&cfg.Cover, "cover", "/tmp/cover.out", "enable coverage and output it to this path")
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "print cmd version")
}

// Execute is the main entry into the cli
func Execute() {
	rootCmd.Execute()
}

func validateDisplay(cfg *config.Config) error {
	if _, ok := display.Decorators[cfg.Display]; !ok {
		return fmt.Errorf("Unknown display type %v", cfg.Display)
	}
	if cfg.ResultsTemplate == "" {
		cfg.ResultsTemplate = display.Decorators[cfg.Display]
	}
	return nil
}

func runCmd(cfg *config.Config, args ...string) error {
	cmdMut.Lock()
	defer cmdMut.Unlock()

	stdReader, stdWriter := io.Pipe()
	defer stdReader.Close()
	errReader, errWriter := io.Pipe()
	defer errReader.Close()

	set := results.New(cfg, args[len(args)-1])
	deco := display.New(cfg)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stderr = errWriter
	cmd.Stdout = stdWriter

	var wg sync.WaitGroup
	wg.Add(2)
	go consumeTestOutput(&wg, stdReader, set, deco)
	go consumeErrOutput(&wg, errReader, set)

	cmd.Run()
	if err := stdWriter.Close(); err != nil {
		return err
	}
	if err := errWriter.Close(); err != nil {
		return err
	}
	wg.Wait()
	set.Complete()
	deco.Summary(set)
	if cfg.Dump {
		data, err := json.Marshal(set)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}

func consumeTestOutput(wg *sync.WaitGroup, r io.Reader, set *results.Set, deco *display.Renderer) error {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		set.Parse(scanner.Bytes())
		deco.Render(set)
	}
	return scanner.Err()
}

func consumeErrOutput(wg *sync.WaitGroup, r io.Reader, set *results.Set) error {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		set.ParseError(string(scanner.Bytes()))
	}
	return scanner.Err()
}

func watchTestChanges(cfg *config.Config, args []string) error {
	term.Fprintf(cfg.Out, `{{"Watching" | bold | Green}} press {{"ctr-t"| blue}} to re-run tests`, nil)
	testargs, _ := fmtTestArgs(cfg, args...)
	infoSig := make(chan os.Signal, 1)
	signal.Notify(infoSig, syscall.SIGINFO)
	go func() {
		for {
			<-infoSig
			runCmd(cfg, testargs...)
		}
	}()
	paths, _ := findPaths(args)

	fsntfy, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for _, path := range paths {
		path, err := filepath.Abs(strings.TrimSuffix(path, "/..."))
		if err != nil {
			return err
		}
		if err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if info.IsDir() || strings.HasSuffix(path, ".go") {
				if err := fsntfy.Add(path); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	for {
		select {
		case event, ok := <-fsntfy.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write && strings.HasSuffix(event.Name, ".go") {
				if err := onTestEvent(cfg, event.Name); err != nil {
					return err
				}
			}
		case err := <-fsntfy.Errors:
			return err
		}
	}
}

func onTestEvent(cfg *config.Config, path string) error {
	if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
		path = strings.ReplaceAll(path, ".go", "_test.go")
	}
	if _, err := os.Stat(path); err != nil {
		path = filepath.Dir(path)
	}
	path = strings.ReplaceAll(path, cfg.Root, ".")
	args, err := fmtTestArgs(cfg, path)
	if err != nil {
		return err
	}
	return runCmd(cfg, args...)
}

func fmtTestArgs(cfg *config.Config, args ...string) ([]string, error) {
	testArgs := []string{"go", "test", "-json", "-v"}
	if cfg.NoCache {
		testArgs = append(testArgs, "-count=1")
	}
	if cfg.Short {
		testArgs = append(testArgs, "-short")
	}
	if cfg.FailFast {
		testArgs = append(testArgs, "-failfast")
	}
	if cfg.Shuffle {
		testArgs = append(testArgs, "-shuffle", "on")
	}
	if cfg.Cover != "" {
		testArgs = append(testArgs, "-covermode", "atomic", "-coverprofile", cfg.Cover)
	}
	paths, tests := findPaths(args)
	if len(tests) > 0 {
		testArgs = append(testArgs, "-run", strings.Join(tests, "|"))
	}
	return append(testArgs, paths...), nil
}

func printVersion(cfg *config.Config) {
	term.Fprintf(
		cfg.Out,
		display.VersionTemplate,
		struct {
			Pad, OgVersion, GoVersion string
		}{
			Pad:       strings.Repeat(" ", 25),
			OgVersion: centerString(fmt.Sprintf("og%v.%v", major, minor), 25),
			GoVersion: centerString(runtime.Version(), 25),
		})
}

func centerString(str string, width int) string {
	spaces := int(float64(width-len(str)) / 2)
	return strings.Repeat(" ", width-(spaces+len(str))) + str + strings.Repeat(" ", spaces)
}

func findPaths(args []string) (paths, tests []string) {
	for _, arg := range args {
		parts := strings.Split(arg, ":")
		path := parts[0]

		if strings.HasSuffix(path, ".go") {
			if !strings.HasSuffix(path, "_test.go") {
				path = strings.ReplaceAll(path, ".go", "_test.go")
			}
			if len(parts) == 1 {
				tests = append(tests, findTestsInFile(path, -1)...)
			} else if lineNum, err := strconv.Atoi(parts[1]); err != nil {
				tests = append(tests, parts[1])
			} else {
				tests = append(tests, findTestsInFile(path, lineNum)...)
			}
			path = filepath.Dir(path)
		} else if strings.HasPrefix(arg, "Test") {
			tests = append(tests, arg)
			continue
		}

		if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "./") {
			path = "./" + path
		}
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		paths = append(paths, "./...")
	}
	return
}

func findTestsInFile(filepath string, line int) []string {
	if info, err := os.Stat(filepath); err != nil || info.IsDir() {
		return nil
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil
	}
	defer file.Close()
	names := []string{}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for currentLine := 0; scanner.Scan(); currentLine++ {
		lineText := scanner.Text()
		if testFuncPattern.MatchString(lineText) {
			names = append(names, testFuncPattern.FindStringSubmatch(lineText)[1])
		}
		if line >= 0 && currentLine >= line-1 && len(names) > 0 {
			return []string{names[len(names)-1]}
		}
	}
	if line >= 0 && len(names) > 0 {
		return []string{names[len(names)-1]}
	}
	return names
}
