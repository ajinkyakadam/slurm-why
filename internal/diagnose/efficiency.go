package diagnose

import (
	"context"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	Registry = append(Registry, Check{ID: "efficiency", Applies: efficiencyApplies, Run: runEfficiencyCheck})
}

func efficiencyApplies(j *datasource.Job) bool {
	return j.State == "COMPLETED"
}

// efficiencyThreshold is how low a resource's utilization has to be
// (used/requested) before it's worth nudging -- deliberately generous
// (25%) so this doesn't nag over routine headroom, only requests that are
// dramatically oversized relative to what the job actually used.
const efficiencyThreshold = 0.25

func runEfficiencyCheck(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error) {
	var underused []string

	if job.ReqMemBytes > 0 && job.MaxRSSBytes > 0 {
		ratio := float64(job.MaxRSSBytes) / float64(job.ReqMemBytes)
		if ratio < efficiencyThreshold {
			underused = append(underused, fmt.Sprintf(
				"memory: used %s of the %s requested (%.1f%%)",
				formatBytes(job.MaxRSSBytes), formatBytes(job.ReqMemBytes), ratio*100))
		}
	}

	if job.TimeLimit > 0 && job.Elapsed > 0 {
		ratio := float64(job.Elapsed) / float64(job.TimeLimit)
		if ratio < efficiencyThreshold {
			underused = append(underused, fmt.Sprintf(
				"time: ran for %s of the %s requested (%.1f%%)",
				job.Elapsed, job.TimeLimit, ratio*100))
		}
	}

	if len(underused) == 0 {
		return nil, nil
	}

	why := "This job completed successfully, but requested noticeably more than it used:\n"
	for _, u := range underused {
		why += "  - " + u + "\n"
	}
	why += "Over-requesting doesn't break anything, but it can hurt your queue priority and ties up cluster capacity other jobs could use."

	return &format.Diagnosis{
		Title: "Job completed, but its request looks oversized",
		Why:   why,
		Fix:   "Consider resubmitting future runs of this job with a smaller --mem/--time closer to what it actually used, with modest headroom.",
	}, nil
}
