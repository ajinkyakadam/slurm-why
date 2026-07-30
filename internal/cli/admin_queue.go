package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ajinkyakadam/slurm-why/internal/datasource/resolve"
	"github.com/ajinkyakadam/slurm-why/internal/diagnose"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

var adminQueueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Triage every pending job cluster-wide, surfacing the ones that will never run",
	Args:  cobra.NoArgs,
	RunE:  runAdminQueue,
	// Same rationale as the root command: main.go owns error printing and
	// the exit code.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	adminCmd.AddCommand(adminQueueCmd)
}

// willNeverRunCheckIDs are the check IDs whose non-fallback diagnosis
// means a job will never run as submitted, not just waiting its turn --
// exactly the actionable subset an admin should see individually. The
// generic "pending" check always means a normal wait even though it
// applies to every pending job, so it's excluded here.
var willNeverRunCheckIDs = map[string]bool{
	"impossible": true,
	"dependency": true,
}

func runAdminQueue(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ds := resolve.Resolve(ctx)
	if verbose {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "using %s backend\n", ds.Backend())
	}

	jobIDs, err := ds.GetPendingJobIDs(ctx)
	if err != nil {
		return fmt.Errorf("listing pending jobs: %w", err)
	}

	var waitingNormally, willNeverRun, unrecognized int
	var details []string

	for _, jobID := range jobIDs {
		job, err := ds.GetJob(ctx, jobID)
		if err != nil {
			// A job can legitimately disappear between listing and
			// fetching it (e.g. it started running, or was cancelled) --
			// skip it rather than failing the whole triage over one job.
			continue
		}
		d, checkID, err := diagnose.DiagnoseWithCheckID(ctx, ds, job, diagnose.Registry)
		if err != nil {
			continue
		}

		switch {
		case willNeverRunCheckIDs[checkID]:
			willNeverRun++
			details = append(details, fmt.Sprintf("--- Job %s (user: %s) ---\n%s", job.ID, job.User, format.Render(d)))
		case checkID == "pending":
			waitingNormally++
		default:
			unrecognized++
			details = append(details, fmt.Sprintf("--- Job %s (user: %s) ---\n%s", job.ID, job.User, format.Render(d)))
		}
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Queue triage: %d pending job(s)\n\n", len(jobIDs))
	_, _ = fmt.Fprintf(out, "  Waiting normally: %d\n", waitingNormally)
	_, _ = fmt.Fprintf(out, "  Will never run as submitted: %d\n", willNeverRun)
	_, _ = fmt.Fprintf(out, "  Unrecognized: %d\n", unrecognized)

	if len(details) > 0 {
		_, _ = fmt.Fprintln(out)
		for _, d := range details {
			_, _ = fmt.Fprintln(out, d)
		}
	}

	return nil
}
