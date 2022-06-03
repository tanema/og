package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/tanema/og/lib/find"
)

const configPath = "$HOME/.config/og.json"

// Config captures running config from flags and global config
type Config struct {
	Out             io.WriteCloser
	Root            string
	ModName         string
	Display         string        `json:"display"`
	ResultsTemplate string        `json:"results_template"`
	HideExcerpts    bool          `json:"hide_excerpts"`
	HideSkip        bool          `json:"hide_skip"`
	HideEmpty       bool          `json:"hide_empty"`
	HideElapsed     bool          `json:"hide_elapsed"`
	HideSummary     bool          `json:"hide_summary"`
	Threshold       time.Duration `json:"threshold"`
	Watch           bool
	NoCache         bool
	Short           bool
	FailFast        bool
	Shuffle         bool
	Raw             bool
	Dump            bool
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
	module, err := find.Mod("./")
	if err != nil {
		config.ModName = config.Root
	} else {
		config.ModName = module.Mod.Path
	}
}

// TestArgs formats go test args for use on the command
func (config *Config) TestArgs() []string {
	testArgs := []string{"go", "test"}
	if !config.Raw {
		testArgs = append(testArgs, "-json", "-v")
	}
	if config.NoCache {
		testArgs = append(testArgs, "-count=1")
	}
	if config.Short {
		testArgs = append(testArgs, "-short")
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
