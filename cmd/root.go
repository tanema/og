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
)

var (
	filepathPattern = regexp.MustCompile(`^([.*/\.^:]*[^:]*):*(.+)*`)
	displayFlag     string
)

var rootCmd = &cobra.Command{
	Use:   "og",
	Short: "A go test command wrapper to make things colorful",
	Long: `Go's test output can sometimes be quit hard to parse, and harder to scan.
A common solution to this is syntax highlighting. It makes it easy to scan
and notice what exactly is wrong at a glance. og is a tool to run go commands
with color.`,
	Run: func(cmd *cobra.Command, args []string) {
		deco, ok := display.Decorators[displayFlag]
		if !ok {
			cobra.CheckErr(fmt.Errorf("Unknown display type %v", displayFlag))
		}

		path, testArgs, err := parseArgs(args)
		cobra.CheckErr(err)

		module, err := mod.Get(path)
		cobra.CheckErr(err)
		if err != nil {
			path, _ = filepath.Abs(path)
		} else {
			path = module.Mod.Path
		}

		r, w := io.Pipe()
		defer r.Close()

		res := results.New(path)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.Parse(r, deco(os.Stdout).Render)
		}()
		runCommand(w, testArgs)

		w.Close()
		wg.Wait()
		display.Summary(os.Stdout, res)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&displayFlag, "display", "d", "dots", "change the display of the test outputs [dots, pdots, name, pname]")
}

// Execute is the main entry into the cli
func Execute() {
	rootCmd.Execute()
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

func runCommand(w io.Writer, args []string) {
	cmd := exec.Command("go", append([]string{"test", "-json", "-v"}, args...)...)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Env = os.Environ()
	cmd.Run()
}
