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

// A memory-limit kill delivers SIGKILL to the job's processes, but how
// that shows up in sacct's ExitCode ("return_code:signal") depends on how
// the batch script runs the killed process:
//   - ExitSignal == 9: the killed process *is* the job step (e.g. the
//     script uses `exec` to replace itself), so sacct reports the signal
//     directly.
//   - ExitCode == 137 (128+9): the far more common case -- the killed
//     process is a child of the bash-wrapped batch script. bash itself
//     isn't killed, it just exits normally with the propagated 128+signal
//     status, which sacct records as the *return code*, not the signal.
//   - ExitSignal == 125: Slurm's own internal OOM marker. Slurm uses a
//     dummy signal value of 253 (SIG_OOM) when its own cgroup/jobacct_gather
//     OOM detection fires; sacct reports that as signal 253-128=125 (the
//     0x80 bit signals "killed by signal"). This is the "textbook" case,
//     but far less common in practice than the two above -- confirmed by
//     testing against a live cluster, where a real cgroup memory-limit
//     kill showed up as ExitCode=137 (bash-wrapped) or ExitSignal=9
//     (exec'd directly), not this marker.
//
// None of the three patterns above are actually stable across Slurm
// versions. Multi-version testing (26.05.2, 24.11.7, 23.02.8, 21.08.8-2 --
// see test/cluster/) found a fourth, undocumented encoding: a genuine
// cgroup-enforced OOM kill on 24.11.7 reported ExitCode=9:0 in sacct (9 in
// the *return code* slot, signal 0) -- neither ExitSignal==9 nor
// ExitCode==137 -- while `scontrol show job` for that same job reported
// ExitCode=137:0 instead. sacct and scontrol disagreeing about the same
// job's own exit code, on the same version, rules out ever having a
// complete enumeration of "the" encoding to match against.
func oomApplies(j *datasource.Job) bool {
	if j.State == "OUT_OF_MEMORY" || j.ExitSignal == 9 || j.ExitCode == 137 || j.ExitSignal == 125 {
		return true
	}
	// Version-independent fallback: whatever exit code Slurm chose to
	// report, a FAILED job that used substantially more memory than it
	// requested is the memory limit having killed it -- this doesn't
	// depend on Slurm's exit-code encoding at all.
	return j.State == "FAILED" && j.ReqMemBytes > 0 && j.MaxRSSBytes > j.ReqMemBytes
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
