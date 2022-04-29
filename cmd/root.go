package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/display"
	"github.com/tanema/og/lib/find"
	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/watcher"
)

var cfg = &config.Config{}

var rootCmd = &cobra.Command{
	Use:   "og [path[:[lineNum|TestName]]|TestName]",
	Short: "A go test command wrapper to make things colorful",
	Long: `Go's test output can sometimes be quit hard to parse, and harder to scan.
A common solution to this is syntax highlighting. It makes it easy to scan
and notice what exactly is wrong at a glance. og is a tool to run go commands
with color.`,
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(validateDisplay(cfg))
		args, err := fmtArgs(cfg, args...)
		cobra.CheckErr(err)
		test(cfg, args...)
		if cfg.Watch {
			paths, _, _ := find.Paths(args)
			notifier, err := watcher.New(paths...)
			cobra.CheckErr(err)
			notifier.Watch(onEvent(cfg))
		}
	},
}

func init() {
	cobra.CheckErr(cfg.Load())

	rootCmd.Flags().StringVarP(&cfg.Display, "display", "d", "dots", "change the display of the test outputs [dots, pdots, names, pnames]")
	rootCmd.Flags().BoolVarP(&cfg.Watch, "watch", "w", false, "watch for file changes and re-run tests")
	rootCmd.Flags().BoolVarP(&cfg.HideSkip, "hideskip", "s", false, "hide info about skipped tests")
	rootCmd.Flags().BoolVarP(&cfg.HideEmpty, "hideempty", "p", false, "hide info about packages without tests")
	rootCmd.Flags().BoolVarP(&cfg.HideElapsed, "hideelapse", "e", false, "hide the elapsed time output")
	rootCmd.Flags().BoolVarP(&cfg.HideSummary, "hidesummary", "m", false, "hide the complete summary output")
	rootCmd.Flags().DurationVarP(&cfg.Threshold, "threshold", "r", time.Second, "output lists of tests slower than the threshold. 0 will disable")

	rootCmd.Flags().BoolVar(&cfg.NoCache, "nocache", false, "disable go test cache")
	rootCmd.Flags().BoolVar(&cfg.Parallel, "parallel", false, "enable parallel tests")
	rootCmd.Flags().BoolVar(&cfg.Short, "short", false, "run short tests")
	rootCmd.Flags().BoolVar(&cfg.Vet, "vet", false, "vet code alone with tests")
	rootCmd.Flags().BoolVar(&cfg.Race, "race", false, "check for race conditions as well")
	rootCmd.Flags().BoolVar(&cfg.FailFast, "failfast", false, "terminate after first test failure")
	rootCmd.Flags().BoolVar(&cfg.Shuffle, "shuffle", false, "shuffle test order")
	rootCmd.Flags().StringVar(&cfg.Cover, "cover", "", "enable coverage and output it to this path")
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
	if cfg.SummaryTemplate == "" {
		cfg.SummaryTemplate = display.SummaryTemplate
	}
	return nil
}

func onEvent(cfg *config.Config) func(string) {
	return func(path string) {
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			path = strings.ReplaceAll(path, ".go", "_test.go")
		}
		if _, err := os.Stat(path); err != nil {
			path = filepath.Dir(path)
		}
		path = strings.ReplaceAll(path, cfg.Root, ".")
		log.Printf("Running %v\n", path)
		args, _ := fmtArgs(cfg, path)
		test(cfg, args...)
	}
}

func test(cfg *config.Config, args ...string) {
	r, w := io.Pipe()
	defer r.Close()
	var wg sync.WaitGroup
	go process(cfg, &wg, r)
	runCommand(w, args...)
	w.Close()
	wg.Wait()
}

func fmtArgs(cfg *config.Config, args ...string) ([]string, error) {
	testArgs := append([]string{"go", "test", "-json", "-v"}, cfg.Args()...)
	paths, tests, err := find.Paths(args)
	if err != nil {
		return nil, err
	}
	if len(tests) > 0 {
		testArgs = append(testArgs, "-run", strings.Join(tests, "|"))
	}
	return append(testArgs, paths...), nil
}

func process(cfg *config.Config, wg *sync.WaitGroup, r io.Reader) {
	wg.Add(1)
	defer wg.Done()

	set := results.New(cfg.ModName)
	deco := display.New(cfg.Out, cfg)
	defer deco.Summary(set)
	defer set.Complete()

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		pkg, tst := set.Parse(scanner.Bytes())
		deco.Render(set, pkg, tst)
	}
}

func runCommand(w io.Writer, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Env = os.Environ()
	return cmd.Run()
}
