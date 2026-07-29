package cli

import (
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "slurm-why",
	Short: "Diagnose confusing Slurm job states in plain English",
	Long: `slurm-why explains why a Slurm job did what it did — OOM kills,
pending-reason codes, and requests that can never be granted — as a
plain-English Diagnosis -> Why -> Fix, instead of raw sacct/squeue/scontrol
output.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show which data backend (REST or CLI) was used")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
