package diagnose

import (
	"context"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	Registry = append(Registry, Check{ID: "nodefail", Applies: nodeFailApplies, Run: runNodeFailCheck})
}

// nodeFailApplies fires on the terminal NODE_FAIL state. Confirmed against
// a live cluster: Slurm's default behavior on a node failure is to
// transparently requeue the job (back to PENDING with reason BeginTime,
// self-clearing once the node returns), not mark it NODE_FAIL -- that
// terminal state only shows up when the job was submitted with
// --no-requeue, or after requeue attempts are exhausted.
func nodeFailApplies(j *datasource.Job) bool {
	return j.State == "NODE_FAIL"
}

func runNodeFailCheck(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error) {
	return &format.Diagnosis{
		Title: "Job failed because its compute node went down",
		Why:   "This job was running when its node failed -- this is a hardware/infrastructure problem, not something wrong with your job or script.",
		Fix:   "Resubmit the job. If you didn't use --no-requeue, Slurm normally retries automatically on node failure instead of marking the job NODE_FAIL, so if this keeps happening check `sinfo -R` for nodes stuck in a down/drained state.",
	}, nil
}
