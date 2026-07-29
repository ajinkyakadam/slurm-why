package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version, commit, and date are injected at build time via -ldflags
// (see .goreleaser.yaml). They default to "dev" for local builds.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the slurm-why version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "slurm-why %s (commit %s, built %s)\n", version, commit, date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
