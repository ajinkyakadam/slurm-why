package diagnose

import (
	"context"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	Registry = append(Registry, Check{ID: "oom", Applies: oomApplies, Run: runOOMCheck})
}

func oomApplies(j *datasource.Job) bool {
	return j.State == "OUT_OF_MEMORY" || j.ExitSignal == 9
}

func runOOMCheck(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error) {
	why := "Your job was killed for exceeding its memory allocation."
	if job.ReqMemBytes > 0 && job.MaxRSSBytes > 0 && job.MaxRSSBytes < job.ReqMemBytes {
		why += fmt.Sprintf(
			" sacct reports peak usage (%s) below what you requested (%s) -- that's not a contradiction: "+
				"the cgroup enforcing your memory limit checks continuously, while sacct only samples "+
				"roughly every 30 seconds, so a brief spike can trigger the kill without ever showing up "+
				"in the accounting record.",
			formatBytes(job.MaxRSSBytes), formatBytes(job.ReqMemBytes))
	}

	fix := "Increase your memory request (--mem or --mem-per-cpu)."
	if suggested := suggestedMemBytes(job); suggested > 0 {
		fix = fmt.Sprintf("Increase your memory request, e.g. --mem=%s.", formatBytes(suggested))
	}

	return &format.Diagnosis{
		Title: "Job was killed for using too much memory (OOM)",
		Why:   why,
		Fix:   fix,
	}, nil
}

// suggestedMemBytes proposes a new memory request comfortably above
// observed peak usage, or 1.5x the original request if no usage was
// recorded at all.
func suggestedMemBytes(job *datasource.Job) int64 {
	const headroom = 1.25
	switch {
	case job.MaxRSSBytes > 0:
		return int64(float64(job.MaxRSSBytes) * headroom)
	case job.ReqMemBytes > 0:
		return int64(float64(job.ReqMemBytes) * 1.5)
	default:
		return 0
	}
}
