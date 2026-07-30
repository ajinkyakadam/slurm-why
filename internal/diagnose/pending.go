package diagnose

import (
	"context"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	// Priority: pendingRunLast -- this is the generic catch-all for a PD
	// job, so it must run after any more specific PD check (impossible,
	// dependency, ...) has had a chance to defer to it by returning nil.
	Registry = append(Registry, Check{ID: "pending", Applies: pendingApplies, Run: runPendingCheck, Priority: pendingRunLast})
}

func pendingApplies(j *datasource.Job) bool {
	return j.State == "PENDING" || j.State == "PD"
}

type reasonInfo struct {
	why string
	fix string
}

// pendingReasons maps the NODELIST(REASON) codes squeue/scontrol report
// for a pending job to a plain-English explanation and fix.
var pendingReasons = map[string]reasonInfo{
	"Priority": {
		why: "Other jobs in the queue currently have higher scheduling priority than yours.",
		fix: "This is normal -- no action needed. Run `sprio -j <jobid>` to see the priority factors, or request fewer resources/less walltime to improve your queue position.",
	},
	"Resources": {
		why: "There currently aren't enough free resources in the partition to start your job.",
		fix: "This is normal if your request fits the partition -- it should start once resources free up. If it never starts, double check your request actually fits some node in the partition.",
	},
	"QOSMaxJobsPerUserLimit": {
		why: "You've hit the maximum number of concurrently running jobs allowed under your QOS.",
		fix: "Wait for one of your running jobs to finish, or ask your site admin about a QOS with a higher per-user job limit.",
	},
	"AssocMaxJobsLimit": {
		why: "You've hit the maximum number of concurrently running jobs allowed for your account/association.",
		fix: "Wait for one of your running jobs to finish, or ask your site admin about raising this association's job limit.",
	},
	"ReqNodeNotAvail": {
		why: "A node you specifically requested is down, reserved, or otherwise unavailable.",
		fix: "Drop any specific --nodelist constraint if you don't strictly need it, or check `sinfo` for the node's state.",
	},
	"BeginTime": {
		why: "Your job has a --begin time that hasn't arrived yet.",
		fix: "This is expected -- the job becomes eligible to run once the requested begin time arrives.",
	},
	// Confirmed against the Slurm source (job_state_reason.c): the code is
	// "InvalidQOS" (all-caps QOS), not "InvalidQoS" -- the latter never
	// matched any real squeue output.
	"InvalidQOS": {
		why: "The QOS you requested doesn't exist or isn't available to your account.",
		fix: "Run `sacctmgr show qos` to see valid QOS names, then resubmit with a correct --qos.",
	},
	"InvalidAccount": {
		why: "The account you requested doesn't exist or you're not a member of it.",
		fix: "Run `sacctmgr show assoc user=$USER` to see accounts you belong to, then resubmit with a correct --account.",
	},
	// Confirmed against a live cluster: a job with an unresolved
	// --dependency shows this reason (Dependency=afterok:N(unfulfilled)
	// in scontrol show job) while the parent is still running/pending --
	// a normal wait, distinct from DependencyNeverSatisfied below, which
	// the dedicated dependency check owns.
	"Dependency": {
		why: "This job is waiting for another job it depends on (--dependency) to reach the required state first.",
		fix: "This is normal -- no action needed. It'll become eligible once the dependency job finishes.",
	},
	// The three codes below are confirmed against a live Slurm cluster
	// (21.08): a single job that exceeds a partition's per-job ceiling
	// gets one of these specific codes, not the generic "Resources" --
	// Slurm already tells you exactly which limit you hit. The
	// impossible-request check enriches these with the actual numbers;
	// this is the fallback if that lookup fails for any reason.
	"PartitionNodeLimit": {
		why: "Your job requests more nodes than this partition allows for a single job (its configured MaxNodes).",
		fix: "Run `scontrol show partition <partition>` to see the MaxNodes ceiling, and reduce --nodes to fit within it.",
	},
	"MaxMemPerLimit": {
		why: "Your job requests more memory per node than this partition allows for a single job (its configured MaxMemPerNode).",
		fix: "Run `scontrol show partition <partition>` to see the MaxMemPerNode ceiling, and reduce --mem to fit within it.",
	},
	"PartitionTimeLimit": {
		why: "Your job's --time exceeds this partition's maximum allowed walltime for a single job (its configured MaxTime).",
		fix: "Run `scontrol show partition <partition>` to see the MaxTime ceiling, and reduce --time to fit within it.",
	},
}

func runPendingCheck(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error) {
	reason, err := ds.GetPendingReason(ctx, job.ID)
	if err != nil {
		return nil, err
	}

	info, ok := pendingReasons[reason.Code]
	if !ok {
		return &format.Diagnosis{
			Title: fmt.Sprintf("Job is waiting to run (reason: %s)", reason.Code),
			Why:   fmt.Sprintf("Slurm reports the pending reason as %q, which isn't one of the common codes this tool recognizes yet.", reason.Code),
			Fix:   "Check `man squeue` (search for NODELIST(REASON)) for what this code means, or ask your site's HPC support team.",
		}, nil
	}

	return &format.Diagnosis{
		Title: fmt.Sprintf("Job is waiting to run (reason: %s)", reason.Code),
		Why:   info.why,
		Fix:   info.fix,
	}, nil
}
