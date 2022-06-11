package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/modfile"
)

const configPath = "$HOME/.config/og.json"

// Config captures running config from flags and global config
type Config struct {
	Out             io.Writer
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
	Dump            bool
	Cover           string
}

// Load will load global config from config path
func (config *Config) Load() error {
	config.Out = os.Stderr
	config.findRoot("./")
	path := os.ExpandEnv(configPath)
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read config file: %v", err)
	}
	return json.Unmarshal(data, config)
}

func (config *Config) findRoot(curPath string) {
	config.Root, _ = filepath.Abs(curPath)
	fileParts := strings.Split(config.Root, string(filepath.Separator))
	for i := len(fileParts) - 1; i >= 1; i-- {
		path := strings.Join(append(fileParts[:i+1], "go.mod"), string(filepath.Separator))
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		mfile, err := modfile.ParseLax(path, data, nil)
		if err != nil {
			break
		}
		config.ModName = mfile.Module.Mod.Path
		return
	}
	config.ModName = config.Root
}
