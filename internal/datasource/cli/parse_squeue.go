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
	return &datasource.PendingReason{Code: reason}, nil
}
