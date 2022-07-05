package cmd

import (
	"bufio"
	"embed"
	_ "embed" // to allow embedding strings
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
	"github.com/tanema/og/lib/watch"
)

const (
	major = 0
	minor = 1

	configPath = "$HOME/.config/og.json"
	coverPath  = "/tmp/cover.out"
)

type (
	renderData struct {
		Set *results.Set
		Cfg *Config
	}
	// Config captures running config from flags and global config
	Config struct {
		Display      string        `json:"display"`
		Split        bool          `json:"split"`
		HideExcerpts bool          `json:"hide_excerpts"`
		HideElapsed  bool          `json:"hide_elapsed"`
		Threshold    time.Duration `json:"threshold"`
		NoCover      bool          `json:"no_cover"`
	}
)

var (
	//go:embed templates/progress
	displays embed.FS
	//go:embed templates/summary.tmpl
	summarytmpl string
	//go:embed templates/version.tmpl
	versiontmpl     string
	version         string
	cfg             = &Config{}
	cmdMut          sync.Mutex
	testFuncPattern = regexp.MustCompile(`func (Test.*)\(t \*testing\.T\)`)
	root            string
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
	},
	Run: func(cmd *cobra.Command, args []string) {
		if version, _ := cmd.Flags().GetBool("version"); version {
			printVersion(cfg)
			return
		}
		testargs, err := fmtTestArgs(cmd, cfg, args...)
		cobra.CheckErr(err)
		cobra.CheckErr(runCmd(cmd, cfg, testargs...))
		if watch, _ := cmd.Flags().GetBool("watch"); watch {
			cobra.CheckErr(watchTestChanges(cmd, cfg, args))
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("dump", "D", false, "dumps the final state in json for usage")
	rootCmd.Flags().BoolP("watch", "w", false, "watch for file changes and re-run tests")
	rootCmd.Flags().Bool("short", false, "run short tests")
	rootCmd.Flags().Bool("nocache", false, "disable go test cache")
	rootCmd.Flags().Bool("failfast", false, "terminate after first test failure")
	rootCmd.Flags().Bool("shuffle", false, "shuffle test order")
	rootCmd.Flags().BoolP("version", "v", false, "print cmd version")

	rootCmd.Flags().StringVarP(&cfg.Display, "display", "d", "dots", "change the display of the test output [dots,names,icons,bar,spin]")
	rootCmd.Flags().BoolVarP(&cfg.Split, "split", "s", false, "show progress split up by package")
	rootCmd.Flags().BoolVarP(&cfg.HideExcerpts, "hideexcerpts", "x", false, "hide code excerpts in build errors")
	rootCmd.Flags().BoolVarP(&cfg.HideElapsed, "hideelapse", "e", false, "hide the elapsed time output")
	rootCmd.Flags().DurationVarP(&cfg.Threshold, "threshold", "r", 10*time.Second, "output lists of tests slower than the threshold. 0 will disable")
	rootCmd.Flags().BoolVarP(&cfg.NoCover, "nocover", "c", false, "disable coverage")
}

// Execute is the main entry into the cli
func Execute(ver string) {
	version = strings.TrimSpace(ver)
	rootCmd.Execute()
}

// Load will load global config from config path
func (config *Config) Load() error {
	root, _ = filepath.Abs("./")
	path := os.ExpandEnv(configPath)
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read config file: %v", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("error while reading config file [%v]: %v", configPath, err)
	}
	return nil
}

func runCmd(cmd *cobra.Command, cfg *Config, args ...string) error {
	cmdMut.Lock()
	defer cmdMut.Unlock()

	stdReader, stdWriter := io.Pipe()
	defer stdReader.Close()
	errReader, errWriter := io.Pipe()
	defer errReader.Close()

	set := results.New(args[len(args)-1], cfg.Threshold)
	tmpl, err := displays.Open(fmt.Sprintf("templates/progress/%v.tmpl", cfg.Display))
	if err != nil {
		return fmt.Errorf("undefined display %v", cfg.Display)
	}
	display, _ := ioutil.ReadAll(tmpl)
	screen := term.NewScreenBuf(os.Stderr, summarytmpl, string(display))
	gocmd := exec.Command(args[0], args[1:]...)
	gocmd.Env = os.Environ()
	gocmd.Stderr = errWriter
	gocmd.Stdout = stdWriter

	var wg sync.WaitGroup
	wg.Add(2)
	go consume(&wg, stdReader, func(data []byte) {
		set.Parse(data)
		screen.RenderTmpl("results", renderData{set, cfg})
	})
	go consume(&wg, errReader, set.ParseError)

	gocmd.Run()
	stdWriter.Close()
	errWriter.Close()
	wg.Wait()
	set.Complete(!cfg.NoCover, coverPath)
	if err := screen.RenderTmpl("summary", renderData{set, cfg}); err != nil {
		return err
	}
	if dump, _ := cmd.Flags().GetBool("dump"); dump {
		return dumpJSON(set)
	}
	return nil
}

func consume(wg *sync.WaitGroup, r io.Reader, fn func([]byte)) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fn(scanner.Bytes())
	}
}

func watchTestChanges(cmd *cobra.Command, cfg *Config, args []string) error {
	term.Println(`{{"Watching" | bold | Green}} press {{"ctr-t"| blue}} to re-run tests`, nil)
	watcher, err := watch.New()
	if err != nil {
		return err
	}
	go watcher.Start()

	for {
		select {
		case path, ok := <-watcher.Changes:
			if !ok {
				return nil
			}
			path = strings.ReplaceAll(path, root, ".")
			term.Println(`{{"Running" | bold | Magenta}} {{.Path | bold}} [{{.Time}}]`, struct{ Time, Path string }{Time: now(), Path: path})
			args, err := fmtTestArgs(cmd, cfg, path)
			if err != nil {
				return err
			}
			return runCmd(cmd, cfg, args...)
		case err := <-watcher.Errors:
			return err
		}
	}
}

func now() string {
	current := time.Now()
	return fmt.Sprintf("%02d:%02d:%02d", current.Hour(), current.Minute(), current.Second())
}

func fmtTestArgs(cmd *cobra.Command, cfg *Config, args ...string) ([]string, error) {
	testArgs := []string{"go", "test", "-json", "-v"}
	if nocache, _ := cmd.Flags().GetBool("nocache"); nocache {
		testArgs = append(testArgs, "-count=1")
	}
	if short, _ := cmd.Flags().GetBool("short"); short {
		testArgs = append(testArgs, "-short")
	}
	if failFast, _ := cmd.Flags().GetBool("failfast"); failFast {
		testArgs = append(testArgs, "-failfast")
	}
	if shuffle, _ := cmd.Flags().GetBool("shuffle"); shuffle {
		testArgs = append(testArgs, "-shuffle", "on")
	}
	if !cfg.NoCover {
		testArgs = append(testArgs, fmt.Sprintf("-coverprofile=%v", coverPath))
	}
	paths, tests := findPaths(args)
	if len(tests) > 0 {
		testArgs = append(testArgs, "-run", strings.Join(tests, "|"))
	}
	return append(testArgs, paths...), nil
}

func printVersion(cfg *Config) {
	spaces := int(float64(25-len(version)) / 2)
	str := strings.Repeat(" ", 25-(spaces+len(version))) + version + strings.Repeat(" ", spaces)
	term.Println(versiontmpl, str)
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

func dumpJSON(set *results.Set) error {
	data, err := json.Marshal(set)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
