package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/datasource/resolve"
	"github.com/ajinkyakadam/slurm-why/internal/diagnose"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

var verbose bool
var admin bool

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
	// GetJob et al. already work for any job ID regardless of who
	// submitted it -- Slurm's own permission model (AdminLevel=Operator/
	// Admin, account coordinator, or just an unrestricted site) governs
	// what a user can actually see, not this tool. --admin doesn't change
	// what gets fetched; it adds raw technical detail an operator
	// triaging someone else's job wants but a job's own owner usually
	// doesn't need.
	rootCmd.PersistentFlags().BoolVar(&admin, "admin", false, "show additional raw technical detail, for operators triaging another user's job")
}

func runDiagnose(cmd *cobra.Command, args []string) error {
	jobID := args[0]
	ctx := context.Background()

	ds := resolve.Resolve(ctx)
	if verbose {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "using %s backend\n", ds.Backend())
	}

	job, err := ds.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("could not fetch job %s: %w", jobID, err)
	}

	d, err := diagnose.Diagnose(ctx, ds, job, diagnose.Registry)
	if err != nil {
		return fmt.Errorf("diagnosing job %s: %w", jobID, err)
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), format.Render(d))
	if admin {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), renderTechnicalDetails(job))
	}
	return nil
}

// renderTechnicalDetails prints the raw Job fields the plain-English
// diagnosis doesn't surface directly -- useful for an operator fielding a
// ticket about a job they don't own, less so for the job's own user.
func renderTechnicalDetails(job *datasource.Job) string {
	return fmt.Sprintf(`
--- Technical details (--admin) ---
Job ID:      %s
State:       %s
User:        %s
Account:     %s
Partition:   %s
QOS:         %s
Requested:   cpus=%d nodes=%d gpus=%d
Elapsed:     %s / TimeLimit: %s
Exit:        code=%d signal=%d
`,
		job.ID, job.State, job.User, job.Account, job.Partition, job.QOS,
		job.ReqCPUs, job.ReqNodes, job.ReqGPUs,
		job.Elapsed, job.TimeLimit,
		job.ExitCode, job.ExitSignal)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
