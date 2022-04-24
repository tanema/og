package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
	filepathPattern = regexp.MustCompile(`^([.*/\.^:]*[^:]*):*(.+)*`)
	displayFlag     string
	watchFlag       bool
	rankFlag        bool
)

var rootCmd = &cobra.Command{
	Use:   "og",
	Short: "A go test command wrapper to make things colorful",
	Long: `Go's test output can sometimes be quit hard to parse, and harder to scan.
A common solution to this is syntax highlighting. It makes it easy to scan
and notice what exactly is wrong at a glance. og is a tool to run go commands
with color.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := display.Decorators[displayFlag]; !ok {
			cobra.CheckErr(fmt.Errorf("Unknown display type %v", displayFlag))
		}
		path, testArgs, err := parseArgs(args)
		cobra.CheckErr(err)
		module, err := mod.Get(path)
		var modName string
		if err != nil {
			modName, _ = filepath.Abs(path)
		} else {
			modName = module.Mod.Path
		}
		res, err := test(modName, testArgs)
		cobra.CheckErr(err)
		if watchFlag {
			notifier, err := watcher.New(path)
			cobra.CheckErr(err)
			notifier.Watch(func(path string) {
				test(modName, []string{filepath.Dir(path)})
			})
		}
		if rankFlag {
			display.Ranking(os.Stdout, res)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&displayFlag, "display", "d", "dots", "change the display of the test outputs [dots, pdots, names, pnames]")
	rootCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "watch for file changes and re-run tests")
	rootCmd.Flags().BoolVarP(&rankFlag, "rank", "r", false, "output ranking of the 10 slowest tests")
}

// Execute is the main entry into the cli
func Execute() {
	rootCmd.Execute()
}

func test(modName string, args []string) (*results.Set, error) {
	res := results.New(modName)
	r, w := io.Pipe()
	var wg sync.WaitGroup
	go process(&wg, res, modName, r)
	runCommand(w, args)
	if err := r.Close(); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	wg.Wait()
	return res, nil
}

func process(wg *sync.WaitGroup, res *results.Set, modName string, r io.Reader) {
	deco := display.Decorators[displayFlag](os.Stdout)
	wg.Add(1)
	defer wg.Done()
	res.Parse(r, deco.Render)
	deco.Summary(res)
}

func parseArgs(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "./", []string{"./..."}, nil
	}
	arg := args[0]
	if filepathPattern.MatchString(arg) {
		parts := filepathPattern.FindStringSubmatch(arg)
		path, address := parts[1], parts[2]
		if address == "" {
			return path, []string{path}, nil
		}
		if !strings.HasPrefix(path, "./") {
			path = "./" + path
		}
		lineNum, err := strconv.Atoi(address)
		if err != nil {
			return path, []string{"-run", address, path}, nil
		}
		testName, err := find.Test(path, lineNum)
		return path, []string{"-run", testName, path}, nil
	} else if strings.HasPrefix(arg, "Test") {
		return "./", []string{"-run", arg, "./..."}, nil
	}
	return "./", args, fmt.Errorf("i have no idea what you are telling me to do")
}

func runCommand(w io.Writer, args []string) error {
	cmd := exec.Command("go", append([]string{"test", "-json", "-v"}, args...)...)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Env = os.Environ()
	return cmd.Run()
}
