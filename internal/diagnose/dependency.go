package diagnose

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	Registry = append(Registry, Check{ID: "dependency", Applies: dependencyApplies, Run: runDependencyCheck})
}

func dependencyApplies(j *datasource.Job) bool {
	return j.State == "PENDING" || j.State == "PD"
}

// depStatusPattern matches "jobid(status)" pairs within a --dependency
// spec, e.g. "afterok:123(failed)" or "afterok:456(unfulfilled)". Slurm
// annotates every referenced job this way; "unfulfilled" means the parent
// simply hasn't finished yet (not broken), anything else is a resolved
// outcome that didn't satisfy the dependency type.
var depStatusPattern = regexp.MustCompile(`(\d+)\(([a-zA-Z]+)\)`)

func runDependencyCheck(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error) {
	reason, err := ds.GetPendingReason(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	if reason.Code != "DependencyNeverSatisfied" {
		return nil, nil
	}

	dep, err := ds.GetDependency(ctx, job.ID)
	if err != nil {
		return nil, err
	}

	var broken []string
	for _, m := range depStatusPattern.FindAllStringSubmatch(dep.Raw, -1) {
		jobID, status := m[1], m[2]
		if !strings.EqualFold(status, "unfulfilled") {
			broken = append(broken, fmt.Sprintf("job %s (%s)", jobID, status))
		}
	}

	if len(broken) == 0 {
		return &format.Diagnosis{
			Title: "Job is waiting on a dependency that can never be satisfied",
			Why: fmt.Sprintf(
				"Slurm reports this job's dependency (%s) can never be satisfied, but the specific failed job couldn't be identified from the dependency spec.",
				dep.Raw),
			Fix: fmt.Sprintf("Run `scontrol show job %s` and check the Dependency field for the specific job ID(s) and their outcome.", job.ID),
		}, nil
	}

	return &format.Diagnosis{
		Title: "Job is waiting on a dependency that can never be satisfied",
		Why: fmt.Sprintf(
			"This job depends on %s -- since that outcome can't change, the dependency condition (e.g. --dependency=afterok requiring a successful COMPLETED) will never be satisfied, so this job will pend forever as submitted.",
			strings.Join(broken, ", ")),
		Fix: "Cancel this job and resubmit once the dependency is fixed, or resubmit without the --dependency requirement if it's no longer needed.",
	}, nil
}
