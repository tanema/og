package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tanema/og/lib/config"
	"github.com/tanema/og/lib/display"
	"github.com/tanema/og/lib/results"
	"github.com/tanema/og/lib/term"
)

var cfg = &config.Config{}

var rootCmd = &cobra.Command{
	Use:   "og [path[:[lineNum|TestName]]|TestName]",
	Short: "A go command wrapper to make things colorful",
	Long: `og wraps around common go commands to highlight and decorate the output
for better scannability and simply for vanity.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(cfg.Load())
		cobra.CheckErr(validateDisplay(cfg))
	},
}

func init() {
	log.SetFlags(log.Ltime)
	rootCmd.PersistentFlags().BoolVarP(&cfg.Raw, "raw", "R", false, "just run the go command with og easy autocomplete and watch")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Dump, "dump", "D", false, "dumps the final state in json for usage")
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

func runCommand(cfg *config.Config, summary bool, args ...string) {
	var out io.WriteCloser = cfg.Out
	var wg sync.WaitGroup
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()

	if cfg.Raw {
		fmt.Println(term.Sprintf("{{. | cyan}}", strings.Join(args, " ")))
	} else {
		r, w := io.Pipe()
		defer r.Close()
		go processOutput(cfg, &wg, r, summary)
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

func processOutput(cfg *config.Config, wg *sync.WaitGroup, r io.Reader, summary bool) {
	wg.Add(1)
	defer wg.Done()
	set := results.New(cfg.ModName, cfg.Root)
	deco := display.New(cfg)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		set.Parse(scanner.Bytes())
		deco.Render(set)
	}
	set.Complete()
	if summary {
		deco.Summary(set)
	} else {
		deco.BuildErrors(set)
	}
	if cfg.Dump {
		data, _ := json.Marshal(set)
		fmt.Println(string(data))
	}
}
