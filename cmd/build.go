package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Short:   "Build a go program with highlighted build errors",
	Long: `og build is a thin wrapper around go build. It accepts the same
positional arguments such as directories or a single file.

    - og build .
    - og build ./cmd/run
    - og build github.com/tanema/og

Any further go flags can be passed with a -- suffix

    og build . -- -o=bin/og
`,
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cfg, false, append([]string{"go", "build"}, args...)...)
	},
}
