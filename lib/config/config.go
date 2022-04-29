package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/tanema/og/lib/mod"
)

const configPath = "$HOME/.config/og.json"

// Config captures running config from flags and global config
type Config struct {
	Out             io.Writer
	Root            string
	ModName         string
	Display         string        `json:"display"`
	ResultsTemplate string        `json:"results_template"`
	SummaryTemplate string        `json:"summary_template"`
	HideSkip        bool          `json:"hide_skip"`
	HideEmpty       bool          `json:"hide_empty"`
	HideElapsed     bool          `json:"hide_elapsed"`
	HideSummary     bool          `json:"hide_summary"`
	Threshold       time.Duration `json:"threshold"`
	Watch           bool
	NoCache         bool
	Parallel        bool
	Short           bool
	Vet             bool
	Race            bool
	FailFast        bool
	Shuffle         bool
	Cover           string
}

// Load will load global config from config path
func (config *Config) Load() error {
	config.Out = os.Stderr
	config.findRoot()
	path := os.ExpandEnv(configPath)
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read config file: %v", err)
	}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("malformed config: %v", err)
	}
	return nil
}

func (config *Config) findRoot() {
	config.Root, _ = filepath.Abs("./")
	module, err := mod.Get("./")
	if err != nil {
		config.ModName = config.Root
	} else {
		config.ModName = module.Mod.Path
	}
}

// Args formats go test args for use on the command
func (config *Config) Args() []string {
	testArgs := []string{}
	if config.NoCache {
		testArgs = append(testArgs, "-count=1")
	}
	if config.Parallel {
		testArgs = append(testArgs, "-parallel")
	}
	if config.Short {
		testArgs = append(testArgs, "-short")
	}
	if config.Vet {
		testArgs = append(testArgs, "-vet", "all")
	}
	if config.Race {
		testArgs = append(testArgs, "-race")
	}
	if config.FailFast {
		testArgs = append(testArgs, "-failfast")
	}
	if config.Shuffle {
		testArgs = append(testArgs, "-shuffle", "on")
	}
	if config.Cover != "" {
		testArgs = append(testArgs, "-covermode", "atomic", "-coverprofile", config.Cover)
	}
	return testArgs
}