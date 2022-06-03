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
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/display"
	"github.com/tanema/og/lib/find"
	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

const (
	major = 0
	minor = 0
)

var (
	cfg     = &config.Config{}
	cmdMut  sync.Mutex
	version bool
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
		testcmd := runner(cfg)
		cobra.CheckErr(testcmd(testargs...))
		if cfg.Watch {
			cobra.CheckErr(watchTestChanges(cfg, testcmd, args))
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&cfg.Raw, "raw", "R", false, "just run the go command with og easy autocomplete and watch")
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

func runner(cfg *config.Config) func(...string) error {
	if cfg.Raw {
		return raw
	}
	return runCmd
}

func raw(args ...string) error {
	cmdMut.Lock()
	defer cmdMut.Unlock()
	fmt.Println(term.Sprintf("{{. | cyan}}", strings.Join(args, " ")))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr
	return cmd.Run()
}

func runCmd(args ...string) error {
	cmdMut.Lock()
	defer cmdMut.Unlock()
	var wg sync.WaitGroup
	wg.Add(1)
	r, w := io.Pipe()
	defer r.Close()

	set := results.New(cfg, args[len(args)-1])
	deco := display.New(cfg)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stderr = w
	cmd.Stdout = w

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	go func() {
		for scanner.Scan() {
			set.Parse(scanner.Bytes())
			deco.Render(set)
		}
		set.Complete()
		deco.Summary(set)
		wg.Done()
	}()
	if err := scanner.Err(); err != nil {
		return err
	}
	cmd.Run()
	if err := w.Close(); err != nil {
		return err
	}
	wg.Wait()
	if cfg.Dump {
		data, err := json.Marshal(set)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	}
	return nil
}

func watchTestChanges(cfg *config.Config, tstcmd func(...string) error, args []string) error {
	fmt.Println(term.Sprintf(`{{"Watching" | bold | bgGreen}} press {{"ctr-t"| blue}} to re-run tests`, nil))
	testargs, _ := fmtTestArgs(cfg, args...)
	infoSig := make(chan os.Signal, 1)
	signal.Notify(infoSig, syscall.SIGINFO)
	go func() {
		for {
			<-infoSig
			tstcmd(testargs...)
		}
	}()
	paths, _, _ := find.Paths(args)

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
				if err := onTestEvent(cfg, tstcmd, event.Name); err != nil {
					return err
				}
			}
		case err := <-fsntfy.Errors:
			return err
		}
	}
}

func onTestEvent(cfg *config.Config, tstcmd func(...string) error, path string) error {
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
	return tstcmd(args...)
}

func fmtTestArgs(cfg *config.Config, args ...string) ([]string, error) {
	testArgs := cfg.TestArgs()
	paths, tests, err := find.Paths(args)
	if err != nil {
		return nil, err
	}
	if len(tests) > 0 {
		testArgs = append(testArgs, "-run", strings.Join(tests, "|"))
	}
	return append(testArgs, paths...), nil
}

func printVersion(cfg *config.Config) {
	fmt.Fprintln(cfg.Out, term.Sprintf("{{. | Rainbow}}", fmt.Sprintf("og%v.%v\t%v", major, minor, runtime.Version())))
}
