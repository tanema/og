package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/tanema/og/lib/term"
)

const (
	major = 0
	minor = 0
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print the version number of og",
	Long:    `All software has versions. This is og's`,
	Run: func(cmd *cobra.Command, args []string) {
		versionStr := fmt.Sprintf("og%v.%v\n%v", major, minor, runtime.Version())
		fmt.Println(term.Sprintf("{{. | rainbow}}", versionStr))
	},
}
