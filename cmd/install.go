package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "Install a go program with highlighted build errors",
	Long: `og install is a thin wrapper around go install. It accepts the same
positional arguments such as directories or a single file.

    - og install .
    - og install ./cmd/run
    - og install github.com/tanema/og

Any further go flags can be passed with a -- suffix

    og install . -- -tags=fun
`,
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cfg, false, append([]string{"go", "install"}, args...)...)
	},
}
