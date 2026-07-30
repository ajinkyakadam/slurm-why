package cli

import (
	"github.com/spf13/cobra"
)

// adminCmd is the parent for admin-mode subcommands that operate across
// many jobs at once, rather than the root command's single-job diagnosis.
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Cluster-wide admin-mode commands",
}

func init() {
	rootCmd.AddCommand(adminCmd)
}
