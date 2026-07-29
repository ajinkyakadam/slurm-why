package diagnose

import (
	"context"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	Registry = append(Registry, Check{ID: "impossible", Applies: impossibleApplies, Run: runImpossibleCheck})
}

func impossibleApplies(j *datasource.Job) bool {
	return j.State == "PENDING" || j.State == "PD"
}

// runImpossibleCheck only investigates the "Resources" pending reason.
// Slurm deliberately leaves that code ambiguous -- it covers both "not
// enough free capacity right now" and "this request can never be granted
// on this partition/QOS" -- so it's the one case where comparing the
// request against published ceilings actually resolves something.
//
// The other reason codes that might look impossibility-adjacent
// (QOSMaxJobsPerUserLimit, AssocMaxJobsLimit, ReqNodeNotAvail) are
// concurrency throttles or node-health issues, not resource ceilings --
// Slurm already names the exact cause, and pendingCheck's generic table
// explains them correctly. There's nothing for a limit comparison to add.
func runImpossibleCheck(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error) {
	reason, err := ds.GetPendingReason(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	if reason.Code != "Resources" {
		return nil, nil
	}

	var violations []string

	if part, err := ds.GetPartitionLimits(ctx, job.Partition); err == nil {
		violations = append(violations, partitionViolations(job, part)...)
	}
	if qos, err := ds.GetQOSLimits(ctx, job.QOS); err == nil {
		violations = append(violations, qosViolations(job, qos)...)
	}

	if len(violations) == 0 {
		return nil, nil
	}

	why := "Your job requests more than this queue can ever grant, so it will pend forever -- it's not just waiting its turn:\n"
	for _, v := range violations {
		why += "  - " + v + "\n"
	}

	return &format.Diagnosis{
		Title: "Job request exceeds a hard limit and will never run as submitted",
		Why:   why,
		Fix:   "Reduce the request to fit within the limit(s) above, or ask your site admin whether a different partition/QOS can satisfy it.",
	}, nil
}

func partitionViolations(job *datasource.Job, part *datasource.PartitionLimits) []string {
	var violations []string

	if part.MaxTimeSet && job.TimeLimit > part.MaxTimeLimit {
		violations = append(violations, fmt.Sprintf(
			"you requested --time=%s but partition %q allows at most %s",
			job.TimeLimit, job.Partition, part.MaxTimeLimit))
	}
	if part.MaxNodes != datasource.LimitUnlimited && part.MaxNodes != datasource.LimitUnset && job.ReqNodes > part.MaxNodes {
		violations = append(violations, fmt.Sprintf(
			"you requested %d node(s) but partition %q allows at most %d",
			job.ReqNodes, job.Partition, part.MaxNodes))
	}
	if part.MaxMemPerNode != datasource.LimitUnlimited && part.MaxMemPerNode != datasource.LimitUnset &&
		job.ReqNodes > 0 && job.ReqMemBytes/int64(job.ReqNodes) > part.MaxMemPerNode {
		violations = append(violations, fmt.Sprintf(
			"your per-node memory request (%s) exceeds partition %q's max of %s per node",
			formatBytes(job.ReqMemBytes/int64(job.ReqNodes)), job.Partition, formatBytes(part.MaxMemPerNode)))
	}
	return violations
}

func qosViolations(job *datasource.Job, qos *datasource.QOSLimits) []string {
	var violations []string

	if qos.MaxWallSet && job.TimeLimit > qos.MaxWall {
		violations = append(violations, fmt.Sprintf(
			"you requested --time=%s but QOS %q allows at most %s",
			job.TimeLimit, job.QOS, qos.MaxWall))
	}
	return violations
}
