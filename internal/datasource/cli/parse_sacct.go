package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

// sacctFields is the explicit --format field list used for sacct queries.
// Order must match the parsePipeRows call site below.
var sacctFields = []string{
	"JobID", "State", "ExitCode", "Account", "Partition", "QOS", "User",
	"ReqMem", "MaxRSS", "Elapsed", "Timelimit", "ReqTRES",
}

// GetJob fetches accounting data for jobID via sacct. sacct is queried
// without -X so that step lines (".batch", ".extern", ".0", ...) are
// included — MaxRSS is frequently only populated on a step line, not on
// the parent job line, so all rows are scanned and the maximum is kept.
func (c *Client) GetJob(ctx context.Context, jobID string) (*datasource.Job, error) {
	out, err := c.run(ctx, "sacct",
		"-j", jobID,
		"--parsable2", "--noheader",
		"-o", strings.Join(sacctFields, ","),
	)
	if err != nil {
		return nil, fmt.Errorf("sacct: %w", err)
	}
	return parseSacctOutput(out, jobID)
}

func parseSacctOutput(out []byte, jobID string) (*datasource.Job, error) {
	rows := parsePipeRows(out, sacctFields)
	if len(rows) == 0 {
		return nil, fmt.Errorf("sacct: no accounting record found for job %s", jobID)
	}

	var main map[string]string
	var maxRSS int64
	for _, row := range rows {
		if row["JobID"] == jobID {
			main = row
		}
		rss, err := parseSlurmMemory(row["MaxRSS"])
		if err == nil && rss > maxRSS {
			maxRSS = rss
		}
	}
	if main == nil {
		// Fall back to the first row (e.g. an array-task ID like "123_4"
		// that sacct echoes back slightly differently than requested).
		main = rows[0]
	}

	exitCode, exitSignal := parseSacctExitCode(main["ExitCode"])

	reqMem, err := parseSlurmMemory(main["ReqMem"])
	if err != nil {
		return nil, fmt.Errorf("sacct: ReqMem: %w", err)
	}

	elapsed, _, err := parseSlurmDuration(main["Elapsed"])
	if err != nil {
		return nil, fmt.Errorf("sacct: Elapsed: %w", err)
	}
	timeLimit, _, err := parseSlurmDuration(main["Timelimit"])
	if err != nil {
		return nil, fmt.Errorf("sacct: Timelimit: %w", err)
	}

	tres := parseTRES(main["ReqTRES"])
	reqGPUs := 0
	for k, v := range tres {
		if strings.HasPrefix(k, "gres/gpu") {
			reqGPUs += atoiOr0(v)
		}
	}

	return &datasource.Job{
		ID:          jobID,
		State:       main["State"],
		ExitCode:    exitCode,
		ExitSignal:  exitSignal,
		Account:     main["Account"],
		Partition:   main["Partition"],
		QOS:         main["QOS"],
		User:        main["User"],
		ReqMemBytes: reqMem,
		MaxRSSBytes: maxRSS,
		ReqCPUs:     atoiOr0(tres["cpu"]),
		ReqNodes:    atoiOr0(tres["node"]),
		ReqGPUs:     reqGPUs,
		Elapsed:     elapsed,
		TimeLimit:   timeLimit,
	}, nil
}

// parseSacctExitCode splits sacct's "return_code:signal" ExitCode field
// (e.g. "0:0", "137:9") into its two components.
func parseSacctExitCode(s string) (code, signal int) {
	before, after, ok := strings.Cut(s, ":")
	if !ok {
		n, _ := strconv.Atoi(strings.TrimSpace(s))
		return n, 0
	}
	c, _ := strconv.Atoi(strings.TrimSpace(before))
	sig, _ := strconv.Atoi(strings.TrimSpace(after))
	return c, sig
}
