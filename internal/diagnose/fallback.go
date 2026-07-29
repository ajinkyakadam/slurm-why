package diagnose

import (
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

// fallbackDiagnosis reports a job's state plainly when no registered
// check found anything more specific to say -- covering, in particular,
// a job that already finished (successfully, or for a reason this tool
// doesn't have a dedicated check for yet). Callers must never see empty
// output just because nothing matched.
func fallbackDiagnosis(job *datasource.Job) format.Diagnosis {
	switch job.State {
	case "COMPLETED":
		return format.Diagnosis{
			Title: "Job completed successfully",
			Why: fmt.Sprintf(
				"Exit code %d, ran for %s of a %s limit. Nothing here indicates a problem.",
				job.ExitCode, job.Elapsed, job.TimeLimit),
			Fix: "No action needed.",
		}
	case "FAILED":
		return format.Diagnosis{
			Title: "Job failed",
			Why: fmt.Sprintf(
				"Exit code %d, signal %d. This tool doesn't have a specific diagnosis for this failure yet.",
				job.ExitCode, job.ExitSignal),
			Fix: "Check your job's stdout/stderr log for the actual error from your program.",
		}
	case "TIMEOUT":
		return format.Diagnosis{
			Title: "Job hit its time limit",
			Why:   fmt.Sprintf("Ran for the full %s and was killed for exceeding --time.", job.TimeLimit),
			Fix:   "Increase --time, or checkpoint/restart if the workload runs longer than any single allocation should.",
		}
	case "CANCELLED":
		return format.Diagnosis{
			Title: "Job was cancelled",
			Why:   "This job was cancelled -- by you, an admin, or a dependency/preemption policy.",
			Fix:   "If you didn't cancel it yourself, check with your site's HPC support team.",
		}
	case "RUNNING":
		return format.Diagnosis{
			Title: "Job is currently running",
			Why:   fmt.Sprintf("Elapsed %s of a %s limit.", job.Elapsed, job.TimeLimit),
			Fix:   "No action needed -- check back once it finishes.",
		}
	default:
		return format.Diagnosis{
			Title: fmt.Sprintf("Job state: %s", job.State),
			Why:   "This tool doesn't have a specific diagnosis for this state yet.",
			Fix:   fmt.Sprintf("Run `scontrol show job %s` for full details.", job.ID),
		}
	}
}
