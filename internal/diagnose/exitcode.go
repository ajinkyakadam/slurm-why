package diagnose

import (
	"context"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	Registry = append(Registry, Check{ID: "exitcode", Applies: exitCodeApplies, Run: runExitCodeCheck})
}

// exitCodeApplies explicitly excludes oomApplies (rather than relying on
// registry ordering) so this check can never double-fire with the OOM
// check regardless of which file's init() runs first.
func exitCodeApplies(j *datasource.Job) bool {
	return j.State == "FAILED" && !oomApplies(j)
}

type exitCodeInfo struct {
	why string
	fix string
}

// exitCodes maps common exit codes to a likely cause. Confirmed against a
// live cluster: 127 (command not found) and 139/signal 11 (segfault) both
// reproduce exactly as documented here for a bash-wrapped batch script.
// Deliberately narrow -- an unrecognized code defers (returns nil) to
// fallback.go's generic FAILED message rather than guessing.
var exitCodes = map[int]exitCodeInfo{
	126: {
		why: "Exit code 126 means the shell found your command but couldn't execute it -- usually a permissions problem (missing +x) or the file isn't actually a valid executable/script.",
		fix: "Check the file's permissions (`chmod +x`) and confirm its shebang line (`#!/bin/bash` etc.) is correct.",
	},
	127: {
		why: "Exit code 127 means \"command not found\" -- the shell couldn't find the program you tried to run.",
		fix: "Check for typos in the command/path, confirm any required `module load` ran first, and that the program is actually installed on the compute node (not just the login node).",
	},
	139: {
		why: "Exit code 139 (128+11) means your program was killed by SIGSEGV -- a segmentation fault, almost always a bug in the program itself (or an incompatible library it was linked/loaded against).",
		fix: "Check your program's core dump or debug output. If this only happens on this cluster, double check you're using the library versions/modules the program expects.",
	},
}

func runExitCodeCheck(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error) {
	// ExitSignal == 11 covers the case where the killed process runs
	// directly as the job step (e.g. the batch script execs it) --
	// sacct then reports "0:11" rather than "139:0". Same bash-wrapper-
	// vs-direct-exec distinction the OOM check's ExitCode==137/
	// ExitSignal==9 split already accounts for.
	if job.ExitSignal == 11 {
		info := exitCodes[139]
		return &format.Diagnosis{
			Title: "Job failed with a segmentation fault (SIGSEGV)",
			Why:   info.why,
			Fix:   info.fix,
		}, nil
	}

	info, ok := exitCodes[job.ExitCode]
	if !ok {
		return nil, nil
	}

	return &format.Diagnosis{
		Title: fmt.Sprintf("Job failed with exit code %d", job.ExitCode),
		Why:   info.why,
		Fix:   info.fix,
	}, nil
}
