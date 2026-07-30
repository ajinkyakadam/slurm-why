package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

// GetPendingReason fetches the live NODELIST(REASON) code for jobID via
// squeue. This only succeeds while the job is still in the queue (pending
// or running) — a finished job must be diagnosed from sacct's State
// instead, which the OOM check already does.
func (c *Client) GetPendingReason(ctx context.Context, jobID string) (*datasource.PendingReason, error) {
	out, err := c.run(ctx, "squeue",
		"-h", "-j", jobID,
		"-o", "%i|%T|%R",
	)
	if err != nil {
		return nil, fmt.Errorf("squeue: %w", err)
	}
	return parseSqueueOutput(out, jobID)
}

func parseSqueueOutput(out []byte, jobID string) (*datasource.PendingReason, error) {
	rows := parsePipeRows(out, []string{"JobID", "State", "Reason"})
	if len(rows) == 0 {
		return nil, fmt.Errorf("squeue: job %s not found in the live queue (it may have already finished)", jobID)
	}
	reason := strings.Trim(strings.TrimSpace(rows[0]["Reason"]), "()")

	// Confirmed against a live cluster: some reasons come back as a
	// compound "Code, free-text detail" string, e.g.
	// "ReqNodeNotAvail, Reserved for maintenance" for a job blocked by
	// another reservation. Split off the code so map lookups against the
	// bare reason (e.g. "ReqNodeNotAvail") still match, and keep the rest
	// as Detail for checks that want to surface it.
	code, detail, _ := strings.Cut(reason, ",")
	return &datasource.PendingReason{
		Code:   strings.TrimSpace(code),
		Detail: strings.TrimSpace(detail),
	}, nil
}

// GetPendingJobIDs lists every job ID currently pending cluster-wide, for
// admin-mode queue triage.
func (c *Client) GetPendingJobIDs(ctx context.Context) ([]string, error) {
	out, err := c.run(ctx, "squeue", "-h", "-t", "PENDING", "-o", "%i")
	if err != nil {
		return nil, fmt.Errorf("squeue: %w", err)
	}
	return parsePendingJobIDs(out), nil
}

func parsePendingJobIDs(out []byte) []string {
	var ids []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			ids = append(ids, line)
		}
	}
	return ids
}
