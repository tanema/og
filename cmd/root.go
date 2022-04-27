package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tanema/og/lib/display"
	"github.com/tanema/og/lib/find"
	"github.com/tanema/og/lib/mod"
	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/watcher"
)

var (
	displayFlag string
	watchFlag   bool
	displayCfg  display.Cfg
)

var rootCmd = &cobra.Command{
	Use:   "og [path[:[lineNum|TestName]]|TestName]",
	Short: "A go test command wrapper to make things colorful",
	Long: `Go's test output can sometimes be quit hard to parse, and harder to scan.
A common solution to this is syntax highlighting. It makes it easy to scan
and notice what exactly is wrong at a glance. og is a tool to run go commands
with color.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := display.Decorators[displayFlag]; !ok {
			cobra.CheckErr(fmt.Errorf("Unknown display type %v", displayFlag))
		}
		module, err := mod.Get("./")
		var modName string
		if err != nil {
			modName, _ = filepath.Abs("./")
		} else {
			modName = module.Mod.Path
		}
		cobra.CheckErr(test(modName, args...))
		if watchFlag {
			paths, _, err := find.Paths(args)
			cobra.CheckErr(err)
			notifier, err := watcher.New(paths...)
			cobra.CheckErr(err)
			notifier.Watch(func(path string) {
				test(modName, path)
			})
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&displayFlag, "display", "d", "dots", "change the display of the test outputs [dots, pdots, names, pnames]")
	rootCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "watch for file changes and re-run tests")
	rootCmd.Flags().BoolVarP(&displayCfg.HideSkip, "hideskip", "s", false, "hide info about skipped tests")
	rootCmd.Flags().BoolVarP(&displayCfg.HideEmpty, "hideempty", "p", false, "hide info about packages without tests")
	rootCmd.Flags().BoolVarP(&displayCfg.HideElapsed, "hideelapse", "e", false, "hide the elapsed time output")
	rootCmd.Flags().BoolVarP(&displayCfg.HideSummary, "hidesummary", "m", false, "hide the complete summary output")
	rootCmd.Flags().StringVarP(&displayCfg.SummaryTemplate, "summarytmpl", "u", display.SummaryTemplate, "The text/tempalte for summary")
	rootCmd.Flags().StringVarP(&displayCfg.ResultsTemplate, "resultstmpl", "t", "", "The text/tempalte for results")
	rootCmd.Flags().DurationVarP(&displayCfg.Rank, "rank", "r", 0, "output ranking of the 10 slowest tests")
}

// Execute is the main entry into the cli
func Execute() {
	rootCmd.Execute()
}

func fmtArgs(args []string) ([]string, error) {
	testArgs := []string{}
	paths, tests, err := find.Paths(args)
	if err != nil {
		return nil, err
	}
	if len(tests) > 0 {
		testArgs = append(testArgs, "-run", strings.Join(tests, "|"))
	}
	return append(testArgs, paths...), nil
}

func test(modName string, args ...string) error {
	args, err := fmtArgs(args)
	if err != nil {
		return err
	}
	res := results.New(modName)
	r, w := io.Pipe()
	var wg sync.WaitGroup
	go process(&wg, res, modName, r)
	runCommand(w, args)
	if err := r.Close(); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func process(wg *sync.WaitGroup, res *results.Set, modName string, r io.Reader) {
	if displayCfg.ResultsTemplate == "" {
		displayCfg.ResultsTemplate = display.Decorators[displayFlag]
	}
	deco := display.New(os.Stdout, displayCfg)
	wg.Add(1)
	defer wg.Done()
	res.Parse(r, deco.Render)
	deco.Summary(res)
}

func runCommand(w io.Writer, args []string) error {
	defaultArgs := []string{"test", "-json", "-v"}
	goArgs := append(defaultArgs, args...)
	cmd := exec.Command("go", goArgs...)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Env = os.Environ()
	return cmd.Run()
}
