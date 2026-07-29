package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ajinkyakadam/slurm-why/internal/datasource/resolve"
	"github.com/ajinkyakadam/slurm-why/internal/diagnose"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "slurm-why <job-id>",
	Short: "Diagnose confusing Slurm job states in plain English",
	Long: `slurm-why explains why a Slurm job did what it did — OOM kills,
pending-reason codes, and requests that can never be granted — as a
plain-English Diagnosis -> Why -> Fix, instead of raw sacct/squeue/scontrol
output.`,
	Args: cobra.ExactArgs(1),
	RunE: runDiagnose,
	// main.go is the single place errors get printed (and sets the exit
	// code); without these, Cobra would also print "Error: ..." plus a
	// full usage block on every runtime failure, not just argument-parsing
	// mistakes, duplicating output and burying the actual error.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show which data backend (REST or CLI) was used")
}

func runDiagnose(cmd *cobra.Command, args []string) error {
	jobID := args[0]
	ctx := context.Background()

	ds := resolve.Resolve(ctx)
	if verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "using %s backend\n", ds.Backend())
	}

	job, err := ds.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("could not fetch job %s: %w", jobID, err)
	}

	d, err := diagnose.Diagnose(ctx, ds, job, diagnose.Registry)
	if err != nil {
		return fmt.Errorf("diagnosing job %s: %w", jobID, err)
	}

	fmt.Fprint(cmd.OutOrStdout(), format.Render(d))
	return nil
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
