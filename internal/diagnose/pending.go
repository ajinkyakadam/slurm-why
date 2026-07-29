package diagnose

import (
	"context"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	Registry = append(Registry, Check{ID: "pending", Applies: pendingApplies, Run: runPendingCheck})
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
	"InvalidQoS": {
		why: "The QOS you requested doesn't exist or isn't available to your account.",
		fix: "Run `sacctmgr show qos` to see valid QOS names, then resubmit with a correct --qos.",
	},
	"InvalidAccount": {
		why: "The account you requested doesn't exist or you're not a member of it.",
		fix: "Run `sacctmgr show assoc user=$USER` to see accounts you belong to, then resubmit with a correct --account.",
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
