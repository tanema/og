package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/find"
	"github.com/tanema/og/lib/term"
	"github.com/tanema/og/lib/watcher"
)

var testCmd = &cobra.Command{
	Use:     "test [path]",
	Aliases: []string{"t"},
	Short:   "Run go test but make it colorful",
	Long: `Go's test output can sometimes be quit hard to parse, and harder to scan.
A common solution to this is syntax highlighting. It makes it easy to scan
and notice what exactly is wrong at a glance. og test does this.

The easy autocomplete works as follows.
    - og test                           => go test ./...
    - og test foldername                => go test ./foldername
    - og test folder/file_test.go       => go test -run TestsInFile ./folder
    - og test folder/file_test.go:20    => go test -run TestAtLine20 ./folder
    - og test TestA                     => go test -run TestA ./...
    - og test folder/file_test.go:TestA => go test -run TestA ./folder

Any further go flags can be passed with a -- suffix

    og test -- -vet=atomic
`,
	Run: func(cmd *cobra.Command, args []string) {
		testargs, err := fmtTestArgs(cfg, args...)
		cobra.CheckErr(err)
		runCommand(cfg, true, testargs...)
		if cfg.Watch {
			paths, _, _ := find.Paths(args)
			notifier, err := watcher.New(paths...)
			cobra.CheckErr(err)
			notifier.Watch(onTestEvent(cfg))
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&cfg.Display, "display", "d", "dots", "change the display of the test outputs [dots, pdots, names, pnames]")
	testCmd.Flags().BoolVarP(&cfg.Watch, "watch", "w", false, "watch for file changes and re-run tests")
	testCmd.Flags().BoolVarP(&cfg.HideSkip, "hideskip", "s", false, "hide info about skipped tests")
	testCmd.Flags().BoolVarP(&cfg.HideExcerpts, "hideexcerpts", "x", false, "hide code excerpts in build errors")
	testCmd.Flags().BoolVarP(&cfg.HideEmpty, "hideempty", "p", false, "hide info about packages without tests")
	testCmd.Flags().BoolVarP(&cfg.HideElapsed, "hideelapse", "e", false, "hide the elapsed time output")
	testCmd.Flags().BoolVarP(&cfg.HideSummary, "hidesummary", "m", false, "hide the complete summary output")
	testCmd.Flags().DurationVarP(&cfg.Threshold, "threshold", "r", 5*time.Second, "output lists of tests slower than the threshold. 0 will disable")
	testCmd.Flags().BoolVar(&cfg.NoCache, "nocache", false, "disable go test cache")
	testCmd.Flags().BoolVar(&cfg.Short, "short", false, "run short tests")
	testCmd.Flags().BoolVar(&cfg.Vet, "vet", false, "vet code alone with tests")
	testCmd.Flags().BoolVar(&cfg.Race, "race", false, "check for race conditions as well")
	testCmd.Flags().BoolVar(&cfg.FailFast, "failfast", false, "terminate after first test failure")
	testCmd.Flags().BoolVar(&cfg.Shuffle, "shuffle", false, "shuffle test order")
	testCmd.Flags().StringVar(&cfg.Cover, "cover", "", "enable coverage and output it to this path")
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
		log.Println(term.Sprintf("running {{. | cyan}}", path))
		args, _ := fmtTestArgs(cfg, path)
		runCommand(cfg, true, args...)
	}
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
