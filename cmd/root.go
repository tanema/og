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

	"github.com/spf13/cobra"
	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/display"
	"github.com/tanema/og/lib/find"
	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
	"github.com/tanema/og/lib/watcher"
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
			printVersion()
			return
		}
		testargs, err := fmtTestArgs(cfg, args...)
		cobra.CheckErr(err)
		runCommand(cfg, testargs...)
		if cfg.Watch {
			watchTestChanges(cfg, args)
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
	rootCmd.Flags().BoolVar(&cfg.NoCache, "nocache", false, "disable go test cache")
	rootCmd.Flags().BoolVar(&cfg.Short, "short", false, "run short tests")
	rootCmd.Flags().BoolVar(&cfg.Vet, "vet", false, "vet code alone with tests")
	rootCmd.Flags().BoolVar(&cfg.Race, "race", false, "check for race conditions as well")
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

func runCommand(cfg *config.Config, args ...string) {
	cmdMut.Lock()
	defer cmdMut.Unlock()
	var out io.WriteCloser = cfg.Out
	var wg sync.WaitGroup
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	if cfg.Raw {
		fmt.Println(term.Sprintf("{{. | cyan}}", strings.Join(args, " ")))
	} else {
		r, w := io.Pipe()
		defer r.Close()
		go processOutput(cfg, &wg, r, args[len(args)-1])
		out = w
	}
	cmd.Stderr = out
	cmd.Stdout = out
	cmd.Run()
	if !cfg.Raw {
		out.Close()
	}
	wg.Wait()
}

func processOutput(cfg *config.Config, wg *sync.WaitGroup, r io.Reader, path string) {
	wg.Add(1)
	defer wg.Done()
	set := results.New(cfg, path)
	deco := display.New(cfg)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		set.Parse(scanner.Bytes())
		deco.Render(set)
	}
	if err := set.Complete(); err != nil {
		panic(err)
	}
	deco.Summary(set)
	if cfg.Dump {
		data, err := json.Marshal(set)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
	}
}

func onTestEvent(cfg *config.Config) func(string) {
	return func(path string) {
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			path = strings.ReplaceAll(path, ".go", "_test.go")
		}
		if _, err := os.Stat(path); err != nil {
			path = filepath.Dir(path)
		}
		path = strings.ReplaceAll(path, cfg.Root, ".")
		args, _ := fmtTestArgs(cfg, path)
		runCommand(cfg, args...)
	}
}

func watchTestChanges(cfg *config.Config, args []string) {
	fmt.Println(term.Sprintf(`{{"Watching" | bold | bgGreen}} press {{"ctr-t"| blue}} to re-run tests`, nil))
	testargs, _ := fmtTestArgs(cfg, args...)
	infoSig := make(chan os.Signal, 1)
	signal.Notify(infoSig, syscall.SIGINFO)
	go func() {
		for {
			<-infoSig
			term.ClearLines(os.Stdout, 1)
			runCommand(cfg, testargs...)
		}
	}()
	paths, _, _ := find.Paths(args)
	notifier, err := watcher.New(paths...)
	cobra.CheckErr(err)
	notifier.Watch(onTestEvent(cfg))
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

func printVersion() {
	fmt.Println(term.Sprintf("{{. | rainbow}}", fmt.Sprintf("og%v.%v\t%v", major, minor, runtime.Version())))
}
